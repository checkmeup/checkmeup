package config

// Unit tests for config.go. No DB/network — these only touch the process
// environment (via t.Setenv, which auto-restores after each test) and
// temp files.
//
// mustEnv and parseDuration both os.Exit(1) on failure, which would kill
// the test binary if called in-process. Rather than skip those branches
// (as done elsewhere in this codebase where os.Exit shows up — see
// cmd/api/main_test.go), they're tested here via the standard Go
// subprocess re-exec idiom: re-invoke this same test binary with a single
// test selected and an env flag set, and assert the *child* process exits
// non-zero. No production code changes needed for this — it's a pure test
// technique.

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestIsDev(t *testing.T) {
	cases := []struct {
		env  string
		want bool
	}{
		{"development", true},
		{"production", false},
		{"staging", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.env, func(t *testing.T) {
			c := &Config{Env: tc.env}
			if got := c.IsDev(); got != tc.want {
				t.Fatalf("IsDev() with Env=%q = %v, want %v", tc.env, got, tc.want)
			}
		})
	}
}

func TestParseOrigins(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"single origin", "http://localhost:5173", []string{"http://localhost:5173"}},
		{"multiple origins", "http://a.com,http://b.com", []string{"http://a.com", "http://b.com"}},
		{"whitespace around commas is trimmed", "http://a.com, http://b.com , http://c.com", []string{"http://a.com", "http://b.com", "http://c.com"}},
		{"trailing comma produces no empty entry", "http://a.com,", []string{"http://a.com"}},
		{"empty string yields no origins", "", nil},
		{"all-empty entries yield no origins", " , , ", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseOrigins(tc.in)
			if !equalStrings(got, tc.want) {
				t.Fatalf("parseOrigins(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"15m", 15 * time.Minute},
		{"168h", 168 * time.Hour},
		{"1h30m", 90 * time.Minute},
		{"0s", 0},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := parseDuration(tc.in); got != tc.want {
				t.Fatalf("parseDuration(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	t.Run("returns the set value", func(t *testing.T) {
		t.Setenv("CONFIG_TEST_GETENV", "configured")
		if got := getEnv("CONFIG_TEST_GETENV", "fallback"); got != "configured" {
			t.Fatalf("want configured, got %q", got)
		}
	})

	t.Run("returns the fallback when unset", func(t *testing.T) {
		t.Setenv("CONFIG_TEST_GETENV_UNSET", "") // ensure a clean slate, then unset
		_ = os.Unsetenv("CONFIG_TEST_GETENV_UNSET")
		if got := getEnv("CONFIG_TEST_GETENV_UNSET", "fallback"); got != "fallback" {
			t.Fatalf("want fallback, got %q", got)
		}
	})

	t.Run("returns the fallback when set to an empty string", func(t *testing.T) {
		t.Setenv("CONFIG_TEST_GETENV_EMPTY", "")
		if got := getEnv("CONFIG_TEST_GETENV_EMPTY", "fallback"); got != "fallback" {
			t.Fatalf("want fallback for an empty value, got %q", got)
		}
	})
}

func TestMustEnv(t *testing.T) {
	t.Run("returns the set value", func(t *testing.T) {
		t.Setenv("CONFIG_TEST_MUSTENV", "required-value")
		if got := mustEnv("CONFIG_TEST_MUSTENV"); got != "required-value" {
			t.Fatalf("want required-value, got %q", got)
		}
	})
}

// TestMustEnv_MissingVarExits and TestParseDuration_InvalidExits use the
// subprocess re-exec idiom: when BE_CRASHER=1 is set, the test calls the
// exit-triggering code directly and returns (the child process then exits
// via os.Exit(1) from inside mustEnv/parseDuration); otherwise it re-execs
// itself with that flag set and asserts the child exited non-zero.

func TestMustEnv_MissingVarExits(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		mustEnv("CONFIG_TEST_MUSTENV_DOES_NOT_EXIST")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMustEnv_MissingVarExits")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok && !exitErr.Success() {
		return
	}
	t.Fatalf("want mustEnv to os.Exit(1) on a missing var, got err=%v", err)
}

func TestParseDuration_InvalidExits(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		parseDuration("not-a-valid-duration")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestParseDuration_InvalidExits")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok && !exitErr.Success() {
		return
	}
	t.Fatalf("want parseDuration to os.Exit(1) on an invalid duration, got err=%v", err)
}

func writeTempEnvFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp env file: %v", err)
	}
	return path
}

func TestLoadDotEnv(t *testing.T) {
	t.Run("sets a var that isn't already set", func(t *testing.T) {
		path := writeTempEnvFile(t, "CONFIG_TEST_NEW_VAR=hello\n")
		t.Cleanup(func() { _ = os.Unsetenv("CONFIG_TEST_NEW_VAR") })

		loadDotEnv(path)

		if got := os.Getenv("CONFIG_TEST_NEW_VAR"); got != "hello" {
			t.Fatalf("want CONFIG_TEST_NEW_VAR=hello, got %q", got)
		}
	})

	t.Run("does not override an already-set var", func(t *testing.T) {
		t.Setenv("CONFIG_TEST_EXISTING_VAR", "original")
		path := writeTempEnvFile(t, "CONFIG_TEST_EXISTING_VAR=from-file\n")

		loadDotEnv(path)

		if got := os.Getenv("CONFIG_TEST_EXISTING_VAR"); got != "original" {
			t.Fatalf("want the pre-existing value preserved, got %q", got)
		}
	})

	t.Run("skips blank lines and comments, trims key and value", func(t *testing.T) {
		path := writeTempEnvFile(t, "\n# a comment\n  CONFIG_TEST_TRIMMED  =  value with spaces  \n\n# another comment\n")
		t.Cleanup(func() { _ = os.Unsetenv("CONFIG_TEST_TRIMMED") })

		loadDotEnv(path)

		if got := os.Getenv("CONFIG_TEST_TRIMMED"); got != "value with spaces" {
			t.Fatalf("want trimmed key/value %q, got %q", "value with spaces", got)
		}
	})

	t.Run("a line with no '=' is ignored", func(t *testing.T) {
		path := writeTempEnvFile(t, "NOT_A_VALID_LINE\nCONFIG_TEST_AFTER_BAD_LINE=ok\n")
		t.Cleanup(func() { _ = os.Unsetenv("CONFIG_TEST_AFTER_BAD_LINE") })

		loadDotEnv(path) // must not panic on the malformed line

		if got := os.Getenv("CONFIG_TEST_AFTER_BAD_LINE"); got != "ok" {
			t.Fatalf("want parsing to continue past the malformed line, got %q", got)
		}
	})

	t.Run("a missing file is a silent no-op", func(t *testing.T) {
		loadDotEnv(filepath.Join(t.TempDir(), "does-not-exist.env")) // must not panic
	})
}

func TestLoad(t *testing.T) {
	// Every optional var Load() reads, neutralized so default-fallback
	// behavior is deterministic regardless of the ambient environment.
	optionalVars := []string{
		"ENV", "PORT", "JWT_ACCESS_TTL", "JWT_REFRESH_TTL", "CORS_ORIGINS",
		"MIGRATIONS_DIR", "STATIC_DIR", "RESEND_API_KEY", "APP_URL", "BASE_URL",
		"TELEGRAM_BOT_TOKEN", "PADDLE_ENVIRONMENT", "PADDLE_API_KEY", "PADDLE_WEBHOOK_SECRET",
		"PADDLE_SOLO_PRICE_ID", "PADDLE_STARTUP_PRICE_ID", "PADDLE_ENTERPRISE_PRICE_ID",
		"PADDLE_SOLO_ANNUAL_PRICE_ID", "PADDLE_STARTUP_ANNUAL_PRICE_ID", "PADDLE_ENTERPRISE_ANNUAL_PRICE_ID",
	}

	t.Run("required fields and computed defaults", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://test")
		t.Setenv("JWT_SECRET", "test-secret-that-is-at-least-32-chars-long")
		for _, k := range optionalVars {
			t.Setenv(k, "")
		}

		cfg := Load()

		if cfg.DatabaseURL != "postgres://test" {
			t.Fatalf("want DatabaseURL postgres://test, got %q", cfg.DatabaseURL)
		}
		if cfg.JWTSecret != "test-secret-that-is-at-least-32-chars-long" {
			t.Fatalf("want JWTSecret test-secret, got %q", cfg.JWTSecret)
		}
		if cfg.Env != "development" {
			t.Fatalf("want default Env development, got %q", cfg.Env)
		}
		if !cfg.IsDev() {
			t.Fatal("want IsDev() true with the default Env")
		}
		if cfg.Port != "8080" {
			t.Fatalf("want default Port 8080, got %q", cfg.Port)
		}
		if cfg.JWTAccessTTL != 15*time.Minute {
			t.Fatalf("want default JWTAccessTTL 15m, got %v", cfg.JWTAccessTTL)
		}
		if cfg.JWTRefreshTTL != 168*time.Hour {
			t.Fatalf("want default JWTRefreshTTL 168h, got %v", cfg.JWTRefreshTTL)
		}
		if !equalStrings(cfg.CORSOrigins, []string{"http://localhost:5173"}) {
			t.Fatalf("want default CORSOrigins [http://localhost:5173], got %v", cfg.CORSOrigins)
		}
		if cfg.MigrationsDir != "migrations" {
			t.Fatalf("want default MigrationsDir migrations, got %q", cfg.MigrationsDir)
		}
		if cfg.AppURL != "http://localhost:5173" {
			t.Fatalf("want default AppURL, got %q", cfg.AppURL)
		}
		if cfg.PaddleEnvironment != "production" {
			t.Fatalf("want default PaddleEnvironment production, got %q", cfg.PaddleEnvironment)
		}
		if cfg.BaseURL != "http://localhost:8080" {
			t.Fatalf("want default BaseURL, got %q", cfg.BaseURL)
		}
		if cfg.TelegramBotToken != "" {
			t.Fatalf("want empty TelegramBotToken, got %q", cfg.TelegramBotToken)
		}
		if cfg.TelegramWebhookSecret != "" {
			t.Fatalf("want no derived webhook secret without a bot token, got %q", cfg.TelegramWebhookSecret)
		}
	})

	t.Run("custom values override defaults", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://custom")
		t.Setenv("JWT_SECRET", "custom-secret-that-is-at-least-32-chars-long")
		t.Setenv("ENV", "production")
		t.Setenv("PORT", "9090")
		t.Setenv("JWT_ACCESS_TTL", "5m")
		t.Setenv("JWT_REFRESH_TTL", "24h")
		t.Setenv("CORS_ORIGINS", "https://a.example.com, https://b.example.com")
		t.Setenv("MIGRATIONS_DIR", "custom-migrations")

		cfg := Load()

		if cfg.Env != "production" || cfg.IsDev() {
			t.Fatalf("want Env production / IsDev false, got Env=%q IsDev=%v", cfg.Env, cfg.IsDev())
		}
		if cfg.Port != "9090" {
			t.Fatalf("want Port 9090, got %q", cfg.Port)
		}
		if cfg.JWTAccessTTL != 5*time.Minute {
			t.Fatalf("want JWTAccessTTL 5m, got %v", cfg.JWTAccessTTL)
		}
		if cfg.JWTRefreshTTL != 24*time.Hour {
			t.Fatalf("want JWTRefreshTTL 24h, got %v", cfg.JWTRefreshTTL)
		}
		if !equalStrings(cfg.CORSOrigins, []string{"https://a.example.com", "https://b.example.com"}) {
			t.Fatalf("want parsed CORSOrigins, got %v", cfg.CORSOrigins)
		}
		if cfg.MigrationsDir != "custom-migrations" {
			t.Fatalf("want MigrationsDir custom-migrations, got %q", cfg.MigrationsDir)
		}
	})

	t.Run("derives the Telegram webhook secret from the bot token", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://test")
		t.Setenv("JWT_SECRET", "test-secret-that-is-at-least-32-chars-long")
		t.Setenv("TELEGRAM_BOT_TOKEN", "my-bot-token")

		cfg := Load()

		sum := sha256.Sum256([]byte("my-bot-token"))
		want := hex.EncodeToString(sum[:])
		if cfg.TelegramWebhookSecret != want {
			t.Fatalf("want derived webhook secret %q, got %q", want, cfg.TelegramWebhookSecret)
		}
		if cfg.TelegramBotToken != "my-bot-token" {
			t.Fatalf("want TelegramBotToken my-bot-token, got %q", cfg.TelegramBotToken)
		}
	})
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
