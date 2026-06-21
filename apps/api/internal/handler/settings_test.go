package handler

// Integration test for the Telegram bot's incoming webhook. The rest of the
// old settings handlers (save/test Telegram/email) moved to
// notification_channels_test.go — see EP-28 / ADR-023.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

const testTelegramWebhookSecret = "test-telegram-webhook-secret"

func testSettingsHandler(t *testing.T) *SettingsHandler {
	t.Helper()
	cfg := &config.Config{
		Env:                   "development",
		JWTSecret:             testJWTSecret,
		JWTAccessTTL:          15 * time.Minute,
		JWTRefreshTTL:         7 * 24 * time.Hour,
		AppURL:                "http://localhost:5173",
		TelegramWebhookSecret: testTelegramWebhookSecret,
	}
	return NewSettingsHandler(cfg, telegram.NewClient(""))
}

func TestHandleTelegramWebhook(t *testing.T) {
	settingsH := testSettingsHandler(t)

	doWebhookPost := func(body []byte, secretHeader string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/webhook/telegram", bytes.NewReader(body))
		if secretHeader != "" {
			req.Header.Set("X-Telegram-Bot-Api-Secret-Token", secretHeader)
		}
		w := httptest.NewRecorder()
		settingsH.HandleTelegramWebhook(w, req)
		return w
	}

	t.Run("missing secret header still returns 200 (scanner-proofing, per comment)", func(t *testing.T) {
		w := doWebhookPost([]byte(`{}`), "")
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})

	t.Run("wrong secret header returns 200", func(t *testing.T) {
		w := doWebhookPost([]byte(`{}`), "wrong-secret")
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body still returns 200 (Telegram must never see a retry-worthy status)", func(t *testing.T) {
		w := doWebhookPost([]byte("not json"), testTelegramWebhookSecret)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})

	t.Run("correct secret with a valid /start update returns 200", func(t *testing.T) {
		body, err := json.Marshal(map[string]any{
			"update_id": 1,
			"message": map[string]any{
				"text": "/start",
				"chat": map[string]any{"id": 555},
			},
		})
		if err != nil {
			t.Fatalf("marshal update: %v", err)
		}
		w := doWebhookPost(body, testTelegramWebhookSecret)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
}
