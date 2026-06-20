package main

// Tests for the cmd/api wiring (newLogger, runMigrations,
// registerTelegramWebhook). main() itself isn't tested directly — it blocks
// forever on srv.Start() and os.Exit(1)s on most failure paths, neither of
// which is safely callable in-process from a test.
//
// runMigrations also os.Exit(1)s on every error path (bad DSN, dialect, or
// migration failure), so only its success path is covered here — there's no
// way to exercise the error branches without killing the test binary, short
// of refactoring main.go to accept an injectable exit function, which this
// task doesn't do.
//
// registerTelegramWebhook's "successfully registers" branch isn't covered
// either — that requires a real network call to api.telegram.org. Its other
// three branches (no token, dev mode, and a failed SetWebhook call) are
// covered without any network I/O: the "failed SetWebhook" case is exercised
// by giving registerTelegramWebhook a cfg with a token (so it doesn't skip)
// but a telegram.Client constructed with an *empty* token (so SetWebhook's
// own internal guard fails fast — no HTTP request attempted).

import (
	"bytes"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

func testLoggerToBuf() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewTextHandler(&buf, nil)), &buf
}

func TestNewLogger(t *testing.T) {
	t.Run("development uses a text handler", func(t *testing.T) {
		logger := newLogger(&config.Config{Env: "development"})
		if got := fmt.Sprintf("%T", logger.Handler()); !strings.Contains(got, "TextHandler") {
			t.Fatalf("want a TextHandler, got %s", got)
		}
	})

	t.Run("non-development uses a JSON handler", func(t *testing.T) {
		logger := newLogger(&config.Config{Env: "production"})
		if got := fmt.Sprintf("%T", logger.Handler()); !strings.Contains(got, "JSONHandler") {
			t.Fatalf("want a JSONHandler, got %s", got)
		}
	})
}

func TestRunMigrations(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://checkmeup:checkmeup@db:5432/checkmeup?sslmode=disable"
	}
	// Sanity-check the DB is reachable before handing it to runMigrations,
	// which os.Exit(1)s on a connection failure instead of returning an
	// error — a bad fallback DSN here should fail loudly via t.Fatalf, not
	// silently kill the test binary.
	probe, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := probe.Ping(); err != nil {
		_ = probe.Close()
		t.Fatalf("ping test db: %v", err)
	}
	_ = probe.Close()

	logger, buf := testLoggerToBuf()
	cfg := &config.Config{DatabaseURL: url, MigrationsDir: "../../migrations"}

	// Goose migrations are idempotent — running against an already-migrated
	// DB (every other test file in this module runs against the same one)
	// just reports "no migrations to run" instead of erroring.
	runMigrations(cfg, logger)

	if !strings.Contains(buf.String(), "migrations applied") {
		t.Fatalf("want a 'migrations applied' log line, got %q", buf.String())
	}
}

func TestRegisterTelegramWebhook(t *testing.T) {
	t.Run("no token configured is a no-op", func(t *testing.T) {
		logger, buf := testLoggerToBuf()
		cfg := &config.Config{Env: "production", BaseURL: "http://localhost:8080"}
		registerTelegramWebhook(cfg, telegram.NewClient(""), logger)
		if buf.Len() != 0 {
			t.Fatalf("want no log output, got %q", buf.String())
		}
	})

	t.Run("dev mode is a no-op even with a token configured", func(t *testing.T) {
		logger, buf := testLoggerToBuf()
		cfg := &config.Config{Env: "development", BaseURL: "http://localhost:8080", TelegramBotToken: "fake-token"}
		registerTelegramWebhook(cfg, telegram.NewClient("fake-token"), logger)
		if buf.Len() != 0 {
			t.Fatalf("want no log output, got %q", buf.String())
		}
	})

	t.Run("a failed SetWebhook call is logged as an error, not fatal", func(t *testing.T) {
		logger, buf := testLoggerToBuf()
		cfg := &config.Config{Env: "production", BaseURL: "http://localhost:8080", TelegramBotToken: "fake-token"}
		// tg is deliberately built with an empty token so SetWebhook's own
		// guard rejects it before any HTTP request is attempted, exercising
		// the error-handling branch with zero network I/O.
		registerTelegramWebhook(cfg, telegram.NewClient(""), logger)
		if !strings.Contains(buf.String(), "telegram webhook registration failed") {
			t.Fatalf("want a registration-failed log line, got %q", buf.String())
		}
	})
}
