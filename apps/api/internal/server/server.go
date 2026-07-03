package server

import (
	"context"
	"errors"
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
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/handler"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
	"github.com/checkmeup/checkmeup/internal/respond"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/webhook"
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
	r.Use(s.securityHeaders())
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
	mailer := email.NewSender(s.cfg.ResendAPIKey)
	wh := webhook.NewClient()
	sl := slack.NewClient()
	auth := handler.NewAuthHandler(s.cfg, s.db)
	monitors := handler.NewMonitorHandler(s.cfg, s.db, tg)
	settings := handler.NewSettingsHandler(s.cfg, tg)
	notifChannels := handler.NewNotificationChannelHandler(s.db, tg, mailer, wh, sl)
	ping := handler.NewPingHandler(s.db, tg, mailer, wh, sl)
	statusPages := handler.NewStatusPageHandler(s.db)
	statusPublic := handler.NewStatusPublicHandler(s.db)
	billing := handler.NewBillingHandler(s.cfg, s.db)
	maintenance := handler.NewMaintenanceHandler(s.db)
	suggestions := handler.NewSuggestionHandler(s.cfg, s.db)

	// Public status page — registered before SPA catch-all so Go handles it
	r.With(httprate.LimitByIP(300, time.Minute)).Get("/status/{slug}", statusPublic.ServeHTTP)

	// Badges (EP-30): embeddable SVGs, rate-limited per ADR-013 (not exempted
	// just because they're images — README/CDN embeds can hit these often,
	// so the limit is generous relative to auth routes; Cache-Control on the
	// response is the primary defense against repeat hits).
	r.With(httprate.LimitByIP(300, time.Minute)).Get("/status/{slug}/badge.svg", statusPublic.ServePageBadge)
	r.With(httprate.LimitByIP(300, time.Minute)).Get("/status/{slug}/badge/{monitor_id}.svg", statusPublic.ServeMonitorBadge)

	if s.cfg.StaticDir != "" {
		r.Get("/*", s.handleSPA)
	}

	// No-auth public endpoints
	r.With(httprate.Limit(60, time.Minute, httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
		return chi.URLParam(r, "token"), nil
	}))).Get("/ping/{token}", ping.ReceivePing)
	r.With(httprate.LimitByIP(60, time.Minute)).Post("/webhook/telegram", settings.HandleTelegramWebhook)
	r.With(httprate.LimitByIP(60, time.Minute)).Post("/webhook/paddle", billing.Webhook)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", s.handleHealth)

		r.Route("/auth", func(r chi.Router) {
			r.With(httprate.LimitByIP(5, time.Hour)).Post("/sign-up", auth.SignUp)
			r.With(httprate.LimitByIP(10, 10*time.Minute)).Post("/sign-in", auth.SignIn)
			r.With(httprate.LimitByIP(30, time.Minute)).Post("/sign-out", auth.SignOut)
			r.With(httprate.LimitByIP(30, time.Minute)).Post("/refresh", auth.Refresh)
			r.With(httprate.LimitByIP(3, 10*time.Minute)).Post("/forgot-password", auth.ForgotPassword)
			r.With(httprate.LimitByIP(10, time.Hour)).Post("/reset-password", auth.ResetPassword)
		})

		r.Group(func(r chi.Router) {
			r.Use(apimiddleware.RequireAuth(s.cfg.JWTSecret))
			// Blanket per-org ceiling on every authenticated route, on top
			// of any tighter per-route limit below — bounds how hard a
			// leaked/compromised access token can hammer the API even
			// though it already passed auth. Keyed by org (not IP) since
			// the whole point is limiting a token, not an address.
			r.Use(httprate.Limit(300, time.Minute, httprate.WithKeyFuncs(authOrgKey)))

			r.Get("/me", auth.Me)
			r.Post("/auth/accept-terms", auth.AcceptTerms)
			r.With(
				httprate.Limit(5, time.Hour, httprate.WithKeyByIP(), httprate.WithLimitHandler(suggestionRateLimited)),
				httprate.Limit(20, time.Hour, httprate.WithKeyFuncs(authOrgKey), httprate.WithLimitHandler(suggestionRateLimited)),
			).Post("/suggestions", suggestions.SubmitSuggestion)

			r.Route("/notification-channels", func(r chi.Router) {
				r.Get("/", notifChannels.ListNotificationChannels)
				r.Post("/", notifChannels.CreateNotificationChannel)
				r.With(httprate.LimitByIP(5, time.Minute)).Post("/test", notifChannels.TestNotificationChannel)
				r.Route("/{id}", func(r chi.Router) {
					r.Patch("/", notifChannels.UpdateNotificationChannel)
					r.Delete("/", notifChannels.DeleteNotificationChannel)
					r.Post("/regenerate-secret", notifChannels.RegenerateWebhookSecret)
				})
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
				// Real Paddle API calls, not just DB writes — tighter than
				// the blanket per-org limit above since a normal org
				// changes plans a handful of times a year, not per minute.
				r.With(httprate.Limit(20, time.Hour, httprate.WithKeyFuncs(authOrgKey))).Post("/checkout", billing.CreateCheckout)
				r.With(httprate.Limit(20, time.Hour, httprate.WithKeyFuncs(authOrgKey))).Post("/change-plan", billing.ChangePlan)
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

			r.Route("/monitors/domain", func(r chi.Router) {
				r.Get("/", monitors.ListDomainMonitors)
				r.Post("/", monitors.CreateDomainMonitor)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", monitors.GetDomainMonitor)
					r.Patch("/", monitors.UpdateDomainMonitor)
					r.Delete("/", monitors.DeleteDomainMonitor)
					r.Post("/pause", monitors.PauseDomainMonitor)
					r.Post("/resume", monitors.ResumeDomainMonitor)
				})
			})

			r.Route("/monitors/port", func(r chi.Router) {
				r.Get("/", monitors.ListPortMonitors)
				r.Post("/", monitors.CreatePortMonitor)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", monitors.GetPortMonitor)
					r.Patch("/", monitors.UpdatePortMonitor)
					r.Delete("/", monitors.DeletePortMonitor)
					r.Post("/pause", monitors.PausePortMonitor)
					r.Post("/resume", monitors.ResumePortMonitor)
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

// authOrgKey rate-limits per org rather than per IP, so a single abusive
// org can't dodge an IP-keyed limit by rotating addresses. Only usable
// behind RequireAuth, since it reads claims out of the request context.
func authOrgKey(r *http.Request) (string, error) {
	claims := apimiddleware.ClaimsFrom(r.Context())
	if claims == nil {
		return "", errors.New("no claims in context")
	}
	return claims.OrgID, nil
}

func suggestionRateLimited(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusTooManyRequests, "Too many suggestions submitted — try again later.", "rate_limited")
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

// contentSecurityPolicy allow-lists the third parties actually loaded by the
// SPA: Google Tag Manager/Analytics (ADR-027) and Paddle's checkout overlay
// (ADR-026). Paddle's own JS dynamically pulls in more of *.paddle.com than
// its cdn.paddle.com entry point, so those directives are host-wildcarded
// per Paddle's own CSP guidance rather than enumerated.
//
// img-src is wide open (any https origin) rather than allow-listed: status
// pages render an org-supplied LogoURL (see StatusPageCreateView.vue) that
// can point at any external host, so a tight img-src would silently break
// every custom-logo status page instead of failing loudly.
const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self' https://www.googletagmanager.com https://cdn.paddle.com; " +
	"style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data: https:; " +
	"connect-src 'self' https://www.googletagmanager.com https://www.google-analytics.com https://*.google-analytics.com https://*.analytics.google.com https://*.paddle.com; " +
	"frame-src https://*.paddle.com; " +
	"font-src 'self'; " +
	"object-src 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'; " +
	"frame-ancestors 'none'"

// securityHeaders sets response headers with no legitimate reason to be
// absent — they cost nothing per-request and the app has no product need
// (embeds, third-party framing) that would justify leaving them off.
func (s *Server) securityHeaders() func(http.Handler) http.Handler {
	secure := !s.cfg.IsDev()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Content-Security-Policy", contentSecurityPolicy)
			if secure {
				// Only sent over HTTPS (prod) — sending it over plain HTTP
				// dev has no effect but is a confusing thing to see in
				// devtools while debugging.
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
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
