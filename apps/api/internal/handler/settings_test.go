package handler

// Integration tests for the org settings handlers. Same conventions as the
// other *_test.go files in this package: real Postgres (ADR-010), package
// handler so unexported request types are reused directly.
//
// TestTelegram/TestEmail don't check auth internally (no orgIDFrom call —
// they just relay the chatId/email from the request body to the Telegram/
// email client) but are routed behind RequireAuth in server.go, so their
// "unauthenticated" cases go through that middleware explicitly rather than
// calling the handler directly, unlike every other handler in this package.
//
// TestTelegram's actual-send success path isn't covered: telegram.Client
// with an empty bot token errors out before any network call (by design,
// see telegram.go), so the only telegram-backed behavior testable without a
// real bot token is the resulting 502. TestEmail's success path *is*
// testable — email.Sender with an empty Resend API key no-ops with a nil
// error (ADR-012: "missing key in dev = skip, don't crash"), so a real 204
// is exercised.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/email"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

const testTelegramWebhookSecret = "test-telegram-webhook-secret"

func testSettingsHandler(t *testing.T) (*AuthHandler, *SettingsHandler, *pgxpool.Pool) {
	t.Helper()
	pool := testPool(t)
	cfg := &config.Config{
		Env:                   "development",
		JWTSecret:             testJWTSecret,
		JWTAccessTTL:          15 * time.Minute,
		JWTRefreshTTL:         7 * 24 * time.Hour,
		AppURL:                "http://localhost:5173",
		TelegramWebhookSecret: testTelegramWebhookSecret,
	}
	authH := NewAuthHandler(cfg, pool)
	settingsH := NewSettingsHandler(cfg, pool, telegram.NewClient(""), email.NewSender(""))
	return authH, settingsH, pool
}

func TestGetSettings(t *testing.T) {
	authH, settingsH, pool := testSettingsHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
		w := httptest.NewRecorder()
		settingsH.GetSettings(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("fresh org defaults", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodGet, settingsH.GetSettings, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[settingsResponse](t, w)
		if resp.TelegramChatID != nil {
			t.Fatalf("want no telegram chat id, got %v", *resp.TelegramChatID)
		}
		if resp.AlertEmail == nil || *resp.AlertEmail != u.email {
			t.Fatalf("want alert email defaulted to the signup email %q, got %v", u.email, resp.AlertEmail)
		}
		if resp.EmailAlertsEnabled {
			t.Fatal("want email alerts off by default")
		}
	})
}

func TestSaveTelegram(t *testing.T) {
	authH, settingsH, pool := testSettingsHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, settingsH.SaveTelegram, http.MethodPut, "/api/v1/settings/telegram", saveTelegramRequest{ChatID: "123"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/settings/telegram", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(settingsH.SaveTelegram)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("sets and trims the chat id", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPut, settingsH.SaveTelegram, u.access, saveTelegramRequest{ChatID: "  98765  "})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[settingsResponse](t, w)
		if resp.TelegramChatID == nil || *resp.TelegramChatID != "98765" {
			t.Fatalf("want trimmed chat id 98765, got %v", resp.TelegramChatID)
		}
	})

	t.Run("an empty chat id clears it", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		setW := doAuthed(t, http.MethodPut, settingsH.SaveTelegram, u.access, saveTelegramRequest{ChatID: "111"})
		if setW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d", setW.Code)
		}
		clearW := doAuthed(t, http.MethodPut, settingsH.SaveTelegram, u.access, saveTelegramRequest{ChatID: ""})
		if clearW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", clearW.Code, clearW.Body.String())
		}
		resp := decodeBody[settingsResponse](t, clearW)
		if resp.TelegramChatID != nil {
			t.Fatalf("want chat id cleared, got %v", *resp.TelegramChatID)
		}
	})
}

func TestSaveEmail(t *testing.T) {
	authH, settingsH, pool := testSettingsHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, settingsH.SaveEmail, http.MethodPut, "/api/v1/settings/email", saveEmailRequest{Email: "a@b.com"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/settings/email", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(settingsH.SaveEmail)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("sets and trims the alert email, overriding the signup default", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPut, settingsH.SaveEmail, u.access, saveEmailRequest{Email: "  ops@example.com  "})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[settingsResponse](t, w)
		if resp.AlertEmail == nil || *resp.AlertEmail != "ops@example.com" {
			t.Fatalf("want trimmed alert email ops@example.com, got %v", resp.AlertEmail)
		}
	})

	t.Run("an empty email clears it", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPut, settingsH.SaveEmail, u.access, saveEmailRequest{Email: ""})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[settingsResponse](t, w)
		if resp.AlertEmail != nil {
			t.Fatalf("want alert email cleared, got %v", *resp.AlertEmail)
		}
	})
}

func TestSetEmailAlertsEnabled(t *testing.T) {
	authH, settingsH, pool := testSettingsHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, settingsH.SetEmailAlertsEnabled, http.MethodPut, "/api/v1/settings/email/enabled", setEmailAlertsEnabledRequest{Enabled: true})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/settings/email/enabled", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(settingsH.SetEmailAlertsEnabled)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("toggles on and off", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)

		onW := doAuthed(t, http.MethodPut, settingsH.SetEmailAlertsEnabled, u.access, setEmailAlertsEnabledRequest{Enabled: true})
		if onW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", onW.Code, onW.Body.String())
		}
		if on := decodeBody[settingsResponse](t, onW); !on.EmailAlertsEnabled {
			t.Fatal("want email alerts enabled true")
		}

		offW := doAuthed(t, http.MethodPut, settingsH.SetEmailAlertsEnabled, u.access, setEmailAlertsEnabledRequest{Enabled: false})
		if offW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", offW.Code, offW.Body.String())
		}
		if off := decodeBody[settingsResponse](t, offW); off.EmailAlertsEnabled {
			t.Fatal("want email alerts enabled false")
		}
	})
}

func TestHandleTelegramWebhook(t *testing.T) {
	_, settingsH, _ := testSettingsHandler(t)

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

func TestTestTelegram(t *testing.T) {
	authH, settingsH, pool := testSettingsHandler(t)

	t.Run("unauthenticated (routed behind RequireAuth, unlike other settings handlers)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/settings/telegram/test", bytes.NewReader([]byte(`{"chatId":"123"}`)))
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(settingsH.TestTelegram)).ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/settings/telegram/test", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(settingsH.TestTelegram)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing chatId", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, settingsH.TestTelegram, u.access, testTelegramRequest{ChatID: "  "})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("no bot token configured surfaces as a 502", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, settingsH.TestTelegram, u.access, testTelegramRequest{ChatID: "123"})
		if w.Code != http.StatusBadGateway {
			t.Fatalf("want 502, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "telegram_error" {
			t.Fatalf("want code telegram_error, got %q", body["code"])
		}
	})
}

func TestTestEmail(t *testing.T) {
	authH, settingsH, pool := testSettingsHandler(t)

	t.Run("unauthenticated (routed behind RequireAuth, unlike other settings handlers)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/settings/email/test", bytes.NewReader([]byte(`{"email":"a@b.com"}`)))
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(settingsH.TestEmail)).ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/settings/email/test", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(settingsH.TestEmail)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing email", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, settingsH.TestEmail, u.access, testEmailRequest{Email: "  "})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("no Resend API key configured no-ops successfully (ADR-012 dev behavior)", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, settingsH.TestEmail, u.access, testEmailRequest{Email: "ops@example.com"})
		if w.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", w.Code, w.Body.String())
		}
	})
}
