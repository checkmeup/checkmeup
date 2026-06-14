package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/server"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

func main() {
	cfg := config.Load()

	var logger *slog.Logger
	if cfg.IsDev() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	slog.SetDefault(logger)

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

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		logger.Error("database ping failed", "err", err)
		os.Exit(1)
	}
	logger.Info("database connected")

	if cfg.TelegramBotToken != "" && !cfg.IsDev() {
		tg := telegram.NewClient(cfg.TelegramBotToken)
		webhookURL := cfg.BaseURL + "/webhook/telegram"
		if err := tg.SetWebhook(webhookURL); err != nil {
			logger.Error("telegram webhook registration failed", "err", err)
		} else {
			logger.Info("telegram webhook registered", "url", webhookURL)
		}
	}

	srv := server.New(cfg, logger, db)
	if err := srv.Start(); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}
