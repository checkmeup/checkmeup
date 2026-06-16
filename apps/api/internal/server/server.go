package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/handler"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
	"github.com/checkmeup/checkmeup/internal/respond"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

type Server struct {
	cfg     *config.Config
	logger  *slog.Logger
	db      *pgxpool.Pool
	router  *chi.Mux
	version string
}

func New(cfg *config.Config, logger *slog.Logger, db *pgxpool.Pool, version string) *Server {
	s := &Server{cfg: cfg, logger: logger, db: db, version: version}
	s.router = s.buildRouter()
	return s
}

func (s *Server) buildRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(s.requestLogger())
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 64*1024) // 64 KB — no legitimate payload exceeds this
			next.ServeHTTP(w, r)
		})
	})

	tg := telegram.NewClient(s.cfg.TelegramBotToken)
	auth := handler.NewAuthHandler(s.cfg, s.db)
	monitors := handler.NewMonitorHandler(s.cfg, s.db, tg)
	settings := handler.NewSettingsHandler(s.cfg, s.db, tg)
	ping := handler.NewPingHandler(s.db, tg)
	statusPages := handler.NewStatusPageHandler(s.db)
	statusPublic := handler.NewStatusPublicHandler(s.db)
	billing := handler.NewBillingHandler(s.cfg, s.db)
	maintenance := handler.NewMaintenanceHandler(s.db)

	// Public status page — registered before SPA catch-all so Go handles it
	r.Get("/status/{slug}", statusPublic.ServeHTTP)

	if s.cfg.StaticDir != "" {
		r.Get("/*", s.handleSPA)
	}

	// No-auth public endpoints
	r.With(httprate.Limit(60, time.Minute, httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
		return chi.URLParam(r, "token"), nil
	}))).Get("/ping/{token}", ping.ReceivePing)
	r.With(httprate.LimitByIP(60, time.Minute)).Post("/webhook/telegram", settings.HandleTelegramWebhook)
	r.With(httprate.LimitByIP(60, time.Minute)).Post("/webhook/lemonsqueezy", billing.Webhook)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", s.handleHealth)

		r.Route("/auth", func(r chi.Router) {
			r.With(httprate.LimitByIP(5, time.Hour)).Post("/sign-up", auth.SignUp)
			r.With(httprate.LimitByIP(10, 10*time.Minute)).Post("/sign-in", auth.SignIn)
			r.Post("/sign-out", auth.SignOut)
			r.Post("/refresh", auth.Refresh)
			r.With(httprate.LimitByIP(3, 10*time.Minute)).Post("/forgot-password", auth.ForgotPassword)
			r.Post("/reset-password", auth.ResetPassword)
		})

		r.Group(func(r chi.Router) {
			r.Use(apimiddleware.RequireAuth(s.cfg.JWTSecret))

			r.Get("/me", auth.Me)

			r.Route("/settings", func(r chi.Router) {
				r.Get("/", settings.GetSettings)
				r.Put("/telegram", settings.SaveTelegram)
				r.With(httprate.LimitByIP(5, time.Minute)).Post("/telegram/test", settings.TestTelegram)
			})

			r.Route("/monitors/cron", func(r chi.Router) {
				r.Get("/", monitors.ListCronMonitors)
				r.Post("/", monitors.CreateCronMonitor)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", monitors.GetCronMonitor)
					r.Patch("/", monitors.UpdateCronMonitor)
					r.Delete("/", monitors.DeleteCronMonitor)
					r.Post("/pause", monitors.PauseCronMonitor)
					r.Post("/resume", monitors.ResumeCronMonitor)
					r.Get("/pings", monitors.GetCronPings)
				})
			})

			r.Route("/monitors/uptime", func(r chi.Router) {
				r.Get("/", monitors.ListUptimeMonitors)
				r.Post("/", monitors.CreateUptimeMonitor)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", monitors.GetUptimeMonitor)
					r.Patch("/", monitors.UpdateUptimeMonitor)
					r.Delete("/", monitors.DeleteUptimeMonitor)
					r.Post("/pause", monitors.PauseUptimeMonitor)
					r.Post("/resume", monitors.ResumeUptimeMonitor)
				})
			})

			r.Route("/billing", func(r chi.Router) {
				r.Get("/", billing.GetBillingInfo)
				r.Post("/checkout", billing.CreateCheckout)
			})

			r.Route("/status-pages", func(r chi.Router) {
				r.Get("/check-slug", statusPages.CheckSlug)
				r.Get("/", statusPages.ListStatusPages)
				r.Post("/", statusPages.CreateStatusPage)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", statusPages.GetStatusPage)
					r.Patch("/", statusPages.UpdateStatusPage)
					r.Delete("/", statusPages.DeleteStatusPage)
					r.Put("/monitors", statusPages.SetStatusPageMonitors)
				})
			})

			r.Route("/monitors/ssl", func(r chi.Router) {
				r.Get("/", monitors.ListSSLMonitors)
				r.Post("/", monitors.CreateSSLMonitor)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", monitors.GetSSLMonitor)
					r.Patch("/", monitors.UpdateSSLMonitor)
					r.Delete("/", monitors.DeleteSSLMonitor)
					r.Post("/pause", monitors.PauseSSLMonitor)
					r.Post("/resume", monitors.ResumeSSLMonitor)
				})
			})

			r.Route("/maintenance-windows", func(r chi.Router) {
				r.Get("/", maintenance.ListMaintenanceWindows)
				r.Post("/", maintenance.CreateMaintenanceWindow)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", maintenance.GetMaintenanceWindow)
					r.Patch("/", maintenance.UpdateMaintenanceWindow)
					r.Delete("/", maintenance.DeleteMaintenanceWindow)
					r.Post("/end", maintenance.EndMaintenanceWindowNow)
				})
			})
		})
	})

	return r
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.cfg.Port)
	s.logger.Info("server starting", "addr", addr, "env", s.cfg.Env)

	srv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return srv.ListenAndServe()
}

func (s *Server) handleSPA(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.cfg.StaticDir, filepath.Clean("/"+r.URL.Path))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(s.cfg.StaticDir, "index.html"))
		return
	}
	http.ServeFile(w, r, path)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		s.logger.Error("health check: db ping failed", "err", err)
		respond.Error(w, http.StatusServiceUnavailable, "database unavailable", "db_unavailable")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"status": "ok", "version": s.version})
}

func (s *Server) requestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			s.logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration", time.Since(start),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}
