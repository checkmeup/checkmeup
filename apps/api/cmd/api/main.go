package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/rdap"
	"github.com/checkmeup/checkmeup/internal/server"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/twilio"
	"github.com/checkmeup/checkmeup/internal/webhook"
	"github.com/checkmeup/checkmeup/internal/whois"
	"github.com/checkmeup/checkmeup/internal/worker"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

// dbMaxConns caps the pool explicitly rather than falling back to pgx's
// default (max(4, NumCPU), likely 4-8 on the current Hetzner CX23) — every
// monitor check writes through this same pool alongside all live
// dashboard/API/status-page traffic, and the 2026-07-04 capacity-planning
// discussion flagged the default as the most likely actual ceiling on
// monitor/customer capacity, ahead of the worker's own semaphore/timeout math.
const dbMaxConns = 25

// workerShutdownTimeout bounds how long a stopping process waits for the
// worker's tick and the alerts still being sent, after the HTTP server's own
// shutdownTimeout (10 s): both fit in the 30 s kamal gives a container between
// SIGTERM and SIGKILL.
const workerShutdownTimeout = 15 * time.Second

func main() {
	cfg := config.Load()
	logger := newLogger(cfg)
	slog.SetDefault(logger)

	runMigrations(cfg, logger)

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to parse database url", "err", err)
		os.Exit(1)
	}
	poolConfig.MaxConns = dbMaxConns

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		logger.Error("database ping failed", "err", err)
		os.Exit(1)
	}
	logger.Info("database connected")

	tg := telegram.NewClient(cfg.TelegramBotToken)
	mailer := email.NewSender(cfg.ResendAPIKey)
	wh := webhook.NewClient()
	sl := slack.NewClient()
	rd := rdap.NewClient()
	wi := whois.NewClient()
	sm := twilio.NewClient(cfg.TwilioAccountSID, cfg.TwilioAPIKeySID, cfg.TwilioAPIKeySecret, cfg.TwilioMessagingServiceSID,
		twilio.StatusCallbackURL(cfg.BaseURL, cfg.JWTSecret))
	registerTelegramWebhook(cfg, tg, logger)

	// A deploy stops the old container with SIGTERM: stop serving, let the
	// worker finish its tick, and send the alerts already started.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	workerDone := make(chan struct{})
	go func() {
		worker.Run(ctx, worker.Notifiers{Queries: db.New(pool), Telegram: tg, Mailer: mailer, Webhook: wh, Slack: sl, SMS: sm, RDAP: rd, WHOIS: wi, Logger: logger})
		close(workerDone)
	}()

	srv := server.New(cfg, logger, pool, version)
	if err := srv.Run(ctx); err != nil {
		if ctx.Err() == nil {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
		logger.Error("server shutdown", "err", err)
	}
	if !waitForWorker(workerDone, workerShutdownTimeout) {
		logger.Error("worker shutdown timed out; checks or alerts still running are lost")
	}
	logger.Info("stopped")
}

// waitForWorker reports whether the worker loop stopped and every alert it
// started was sent within timeout.
func waitForWorker(workerDone <-chan struct{}, timeout time.Duration) bool {
	finished := make(chan struct{})
	go func() {
		<-workerDone
		worker.WaitForAlerts()
		close(finished)
	}()
	select {
	case <-finished:
		return true
	case <-time.After(timeout):
		return false
	}
}

func newLogger(cfg *config.Config) *slog.Logger {
	if cfg.IsDev() {
		return slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func runMigrations(cfg *config.Config, logger *slog.Logger) {
	sqlDB, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := goose.SetDialect("postgres"); err != nil {
		logger.Error("goose set dialect", "err", err)
		os.Exit(1)
	}
	if err := goose.Up(sqlDB, cfg.MigrationsDir); err != nil {
		logger.Error("migrations failed", "err", err)
		os.Exit(1)
	}
	logger.Info("migrations applied", "dir", cfg.MigrationsDir)
}

func registerTelegramWebhook(cfg *config.Config, tg *telegram.Client, logger *slog.Logger) {
	if cfg.TelegramBotToken == "" || cfg.IsDev() {
		return
	}
	webhookURL := cfg.BaseURL + "/webhook/telegram"
	if err := tg.SetWebhook(webhookURL, cfg.TelegramWebhookSecret); err != nil {
		logger.Error("telegram webhook registration failed", "err", err)
		return
	}
	logger.Info("telegram webhook registered", "url", webhookURL)
}
