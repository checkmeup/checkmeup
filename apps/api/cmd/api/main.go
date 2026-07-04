package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

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
	"github.com/checkmeup/checkmeup/internal/worker"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	cfg := config.Load()
	logger := newLogger(cfg)
	slog.SetDefault(logger)

	runMigrations(cfg, logger)

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
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
	sm := twilio.NewClient(cfg.TwilioAccountSID, cfg.TwilioAPIKeySID, cfg.TwilioAPIKeySecret, cfg.TwilioMessagingServiceSID)
	registerTelegramWebhook(cfg, tg, logger)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go worker.Run(workerCtx, worker.Notifiers{Queries: db.New(pool), Telegram: tg, Mailer: mailer, Webhook: wh, Slack: sl, SMS: sm, RDAP: rd, Logger: logger})

	srv := server.New(cfg, logger, pool, version)
	if err := srv.Start(); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
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

