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
	Env              string
	Port             string
	DatabaseURL      string
	JWTSecret        string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration
	CORSOrigins      []string
	MigrationsDir    string
	StaticDir        string
	ResendAPIKey     string
	AppURL           string
	BaseURL          string
	TelegramBotToken      string
	TelegramWebhookSecret string // sha256(TelegramBotToken) hex — derived, never stored separately
	LSAPIKey              string // LemonSqueezy API key
	LSStoreID             string // LemonSqueezy store ID
	LSWebhookSecret       string // LemonSqueezy webhook signing secret
	LSSoloVariantID       string // LemonSqueezy variant ID for Solo plan
	LSStartupVariantID    string // LemonSqueezy variant ID for Startup plan
	LSEnterpriseVariantID string // LemonSqueezy variant ID for Enterprise plan
}

func Load() *Config {
	loadDotEnv(".env")

	cfg := &Config{
		Env:           getEnv("ENV", "development"),
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   mustEnv("DATABASE_URL"),
		JWTSecret:     mustEnv("JWT_SECRET"),
		JWTAccessTTL:  parseDuration(getEnv("JWT_ACCESS_TTL", "15m")),
		JWTRefreshTTL: parseDuration(getEnv("JWT_REFRESH_TTL", "168h")),
		CORSOrigins:   parseOrigins(getEnv("CORS_ORIGINS", "http://localhost:5173")),
		MigrationsDir: getEnv("MIGRATIONS_DIR", "migrations"),
		StaticDir:     getEnv("STATIC_DIR", ""),
		ResendAPIKey:  getEnv("RESEND_API_KEY", ""),
		AppURL:        getEnv("APP_URL", "http://localhost:5173"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
		TelegramBotToken:  getEnv("TELEGRAM_BOT_TOKEN", ""),
		LSAPIKey:          getEnv("LS_API_KEY", ""),
		LSStoreID:         getEnv("LS_STORE_ID", ""),
		LSWebhookSecret:   getEnv("LS_WEBHOOK_SECRET", ""),
		LSSoloVariantID:       getEnv("LS_SOLO_VARIANT_ID", ""),
		LSStartupVariantID:    getEnv("LS_STARTUP_VARIANT_ID", ""),
		LSEnterpriseVariantID: getEnv("LS_ENTERPRISE_VARIANT_ID", ""),
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
