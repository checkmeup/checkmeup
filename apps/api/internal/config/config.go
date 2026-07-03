package config

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Config struct {
	Env                           string
	Port                          string
	DatabaseURL                   string
	JWTSecret                     string
	JWTAccessTTL                  time.Duration
	JWTRefreshTTL                 time.Duration
	CORSOrigins                   []string
	MigrationsDir                 string
	StaticDir                     string
	ResendAPIKey                  string
	AppURL                        string
	BaseURL                       string
	TelegramBotToken              string
	TelegramWebhookSecret         string // sha256(TelegramBotToken) hex — derived, never stored separately
	PaddleEnvironment             string // "production" (default) or "sandbox" — selects api.paddle.com vs sandbox-api.paddle.com
	PaddleAPIKey                  string // Paddle API key (server-side, secret)
	PaddleWebhookSecret           string // Paddle webhook signing secret
	PaddleSoloPriceID             string // Paddle price ID for Solo plan (monthly)
	PaddleStartupPriceID          string // Paddle price ID for Startup plan (monthly)
	PaddleEnterprisePriceID       string // Paddle price ID for Enterprise plan (monthly)
	PaddleSoloAnnualPriceID       string // Paddle price ID for Solo plan (annual)
	PaddleStartupAnnualPriceID    string // Paddle price ID for Startup plan (annual)
	PaddleEnterpriseAnnualPriceID string // Paddle price ID for Enterprise plan (annual)
}

func Load() *Config {
	loadDotEnv(".env")

	cfg := &Config{
		Env:                           getEnv("ENV", "development"),
		Port:                          getEnv("PORT", "8080"),
		DatabaseURL:                   mustEnv("DATABASE_URL"),
		JWTSecret:                     mustJWTSecret(),
		JWTAccessTTL:                  parseDuration(getEnv("JWT_ACCESS_TTL", "15m")),
		JWTRefreshTTL:                 parseDuration(getEnv("JWT_REFRESH_TTL", "168h")),
		CORSOrigins:                   parseOrigins(getEnv("CORS_ORIGINS", "http://localhost:5173")),
		MigrationsDir:                 getEnv("MIGRATIONS_DIR", "migrations"),
		StaticDir:                     getEnv("STATIC_DIR", ""),
		ResendAPIKey:                  getEnv("RESEND_API_KEY", ""),
		AppURL:                        getEnv("APP_URL", "http://localhost:5173"),
		BaseURL:                       getEnv("BASE_URL", "http://localhost:8080"),
		TelegramBotToken:              getEnv("TELEGRAM_BOT_TOKEN", ""),
		PaddleEnvironment:             getEnv("PADDLE_ENVIRONMENT", "production"),
		PaddleAPIKey:                  getEnv("PADDLE_API_KEY", ""),
		PaddleWebhookSecret:           getEnv("PADDLE_WEBHOOK_SECRET", ""),
		PaddleSoloPriceID:             getEnv("PADDLE_SOLO_PRICE_ID", ""),
		PaddleStartupPriceID:          getEnv("PADDLE_STARTUP_PRICE_ID", ""),
		PaddleEnterprisePriceID:       getEnv("PADDLE_ENTERPRISE_PRICE_ID", ""),
		PaddleSoloAnnualPriceID:       getEnv("PADDLE_SOLO_ANNUAL_PRICE_ID", ""),
		PaddleStartupAnnualPriceID:    getEnv("PADDLE_STARTUP_ANNUAL_PRICE_ID", ""),
		PaddleEnterpriseAnnualPriceID: getEnv("PADDLE_ENTERPRISE_ANNUAL_PRICE_ID", ""),
	}
	if cfg.TelegramBotToken != "" {
		h := sha256.Sum256([]byte(cfg.TelegramBotToken))
		cfg.TelegramWebhookSecret = hex.EncodeToString(h[:])
	}
	return cfg
}

func (c *Config) IsDev() bool {
	return c.Env == "development"
}

func loadDotEnv(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || os.Getenv(strings.TrimSpace(key)) != "" {
			continue
		}
		_ = os.Setenv(strings.TrimSpace(key), strings.TrimSpace(value))
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required environment variable not set", "key", key)
		os.Exit(1)
	}
	return v
}

// mustJWTSecret requires JWT_SECRET to be at least 32 bytes (256 bits) —
// short enough to matter and the docs already tell operators to generate
// "32+ random bytes" (see docs/deploy.md), so this just enforces what was
// already documented instead of trusting it silently.
func mustJWTSecret() string {
	v := mustEnv("JWT_SECRET")
	if len(v) < 32 {
		slog.Error("JWT_SECRET too short — must be at least 32 characters", "length", len(v))
		os.Exit(1)
	}
	return v
}

func parseOrigins(s string) []string {
	var origins []string
	for _, o := range strings.Split(s, ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		slog.Error("invalid duration", "value", s)
		os.Exit(1)
	}
	return d
}
