package handler

// Integration tests for the notification-channel handlers (EP-28 / ADR-023),
// which replaced the old single telegram/email settings fields. Same
// conventions as the rest of this package: real Postgres (ADR-010), package
// handler so the unexported request/response types are reused directly, and
// withURLParam (maintenance_test.go) injects the chi "id" route param these
// handlers read.
//
// TestNotificationChannel's actual-send paths mirror the old
// TestTelegram/TestEmail boundary: telegram.Client with no bot token errors
// before any network call, so only the resulting 502 is testable; email.Sender
// with no Resend API key no-ops with a nil error (ADR-012), so a real 204 is
// exercised for the email type.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

func testNotificationChannelHandler(t *testing.T) (*AuthHandler, *NotificationChannelHandler, *pgxpool.Pool) {
	t.Helper()
	pool := testPool(t)
	cfg := &config.Config{
		Env:           "development",
		JWTSecret:     testJWTSecret,
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 7 * 24 * time.Hour,
		AppURL:        "http://localhost:5173",
	}
	authH := NewAuthHandler(cfg, pool)
	channelsH := NewNotificationChannelHandler(pool, telegram.NewClient(""), email.NewSender(""), webhook.NewClient(), slack.NewClient())
	return authH, channelsH, pool
}

func doChannelRequest(t *testing.T, method string, handler http.HandlerFunc, access *http.Cookie, id string, body any) *httptest.ResponseRecorder {
	t.Helper()
	return doMaintenanceRequest(t, method, handler, access, id, body)
}

func TestListNotificationChannels(t *testing.T) {
	authH, channelsH, pool := testNotificationChannelHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, channelsH.ListNotificationChannels, http.MethodGet, "/api/v1/notification-channels", nil)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("empty for a fresh org", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodGet, channelsH.ListNotificationChannels, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		channels := decodeBody[[]notificationChannelResponse](t, w)
		if len(channels) != 0 {
			t.Fatalf("want no channels for a fresh org, got %d", len(channels))
		}
	})

	t.Run("lists channels newest first", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		createChannel(t, channelsH, u.access, "telegram", map[string]any{"chatId": "111"}, "First")
		createChannel(t, channelsH, u.access, "email", map[string]any{"email": "a@b.com"}, "Second")

		w := doAuthed(t, http.MethodGet, channelsH.ListNotificationChannels, u.access, nil)
		channels := decodeBody[[]notificationChannelResponse](t, w)
		if len(channels) != 2 {
			t.Fatalf("want 2 channels, got %d", len(channels))
		}
		if channels[0].Name != "Second" {
			t.Fatalf("want newest first (Second), got %q", channels[0].Name)
		}
	})
}

func createChannel(t *testing.T, h *NotificationChannelHandler, access *http.Cookie, channelType string, config map[string]any, name string) notificationChannelResponse {
	t.Helper()
	w := doAuthed(t, http.MethodPost, h.CreateNotificationChannel, access, notificationChannelRequest{
		Type: channelType, Name: name, Config: config,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create channel: want 201, got %d: %s", w.Code, w.Body.String())
	}
	return decodeBody[notificationChannelResponse](t, w)
}

func TestCreateNotificationChannel(t *testing.T) {
	authH, channelsH, pool := testNotificationChannelHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, channelsH.CreateNotificationChannel, http.MethodPost, "/api/v1/notification-channels", notificationChannelRequest{
			Type: "telegram", Name: "X", Config: map[string]any{"chatId": "1"},
		})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, channelsH.CreateNotificationChannel, u.access, notificationChannelRequest{
			Type: "telegram", Config: map[string]any{"chatId": "1"},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, channelsH.CreateNotificationChannel, u.access, notificationChannelRequest{
			Type: "slack", Name: "X", Config: map[string]any{"webhookUrl": "https://example.com"},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400 (slack not supported until its own epic ships), got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("telegram missing chatId", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, channelsH.CreateNotificationChannel, u.access, notificationChannelRequest{
			Type: "telegram", Name: "X", Config: map[string]any{},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("email missing email", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, channelsH.CreateNotificationChannel, u.access, notificationChannelRequest{
			Type: "email", Name: "X", Config: map[string]any{},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("creates a telegram channel, enabled by default", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u.access, "telegram", map[string]any{"chatId": "98765"}, "Ops Telegram")
		if c.Type != "telegram" || c.Name != "Ops Telegram" || !c.Enabled {
			t.Fatalf("unexpected channel: %+v", c)
		}
		if c.Config["chatId"] != "98765" {
			t.Fatalf("want chatId 98765 round-tripped, got %v", c.Config["chatId"])
		}
	})

	t.Run("creates an email channel", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u.access, "email", map[string]any{"email": "ops@example.com"}, "Ops Email")
		if c.Type != "email" || c.Config["email"] != "ops@example.com" {
			t.Fatalf("unexpected channel: %+v", c)
		}
	})

	t.Run("webhook missing url", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, channelsH.CreateNotificationChannel, u.access, notificationChannelRequest{
			Type: "webhook", Name: "X", Config: map[string]any{},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("webhook url must be https", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, channelsH.CreateNotificationChannel, u.access, notificationChannelRequest{
			Type: "webhook", Name: "X", Config: map[string]any{"url": "http://example.com/hook"},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400 for a non-https URL, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("creates a webhook channel with a server-generated signing secret", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u.access, "webhook", map[string]any{"url": "https://example.com/hook"}, "Ops Webhook")
		if c.Type != "webhook" || c.Config["url"] != "https://example.com/hook" {
			t.Fatalf("unexpected channel: %+v", c)
		}
		secret, _ := c.Config["secret"].(string)
		if secret == "" {
			t.Fatal("want a non-empty signing secret generated automatically (US-1401)")
		}
	})

	t.Run("a client-supplied secret is ignored — the server always generates its own", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u.access, "webhook", map[string]any{"url": "https://example.com/hook", "secret": "client-chosen"}, "Ops Webhook")
		if c.Config["secret"] == "client-chosen" {
			t.Fatal("want the server to ignore a client-supplied secret")
		}
	})
}

func TestUpdateNotificationChannel(t *testing.T) {
	authH, channelsH, pool := testNotificationChannelHandler(t)

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doChannelRequest(t, http.MethodPatch, channelsH.UpdateNotificationChannel, u.access, "00000000-0000-0000-0000-000000000000", notificationChannelRequest{
			Name: "X", Config: map[string]any{"chatId": "1"},
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("another org's channel is not found (multi-tenancy)", func(t *testing.T) {
		u1 := signUpTestUser(t, authH, pool)
		u2 := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u1.access, "telegram", map[string]any{"chatId": "1"}, "X")

		w := doChannelRequest(t, http.MethodPatch, channelsH.UpdateNotificationChannel, u2.access, c.ID, notificationChannelRequest{
			Name: "Hijacked", Config: map[string]any{"chatId": "2"},
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("renames, updates config, and can disable", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u.access, "telegram", map[string]any{"chatId": "111"}, "Old name")

		enabled := false
		w := doChannelRequest(t, http.MethodPatch, channelsH.UpdateNotificationChannel, u.access, c.ID, notificationChannelRequest{
			Name: "New name", Config: map[string]any{"chatId": "222"}, Enabled: &enabled,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		updated := decodeBody[notificationChannelResponse](t, w)
		if updated.Name != "New name" || updated.Config["chatId"] != "222" || updated.Enabled {
			t.Fatalf("unexpected channel after update: %+v", updated)
		}
	})

	t.Run("config validated against the channel's existing type, which never changes", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u.access, "email", map[string]any{"email": "a@b.com"}, "X")

		w := doChannelRequest(t, http.MethodPatch, channelsH.UpdateNotificationChannel, u.access, c.ID, notificationChannelRequest{
			Name: "X", Config: map[string]any{}, // missing "email"
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("updating a webhook's URL keeps the original signing secret, ignoring any client-supplied value", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u.access, "webhook", map[string]any{"url": "https://example.com/hook"}, "Hook")
		originalSecret := c.Config["secret"]

		w := doChannelRequest(t, http.MethodPatch, channelsH.UpdateNotificationChannel, u.access, c.ID, notificationChannelRequest{
			Name: "Hook", Config: map[string]any{"url": "https://example.com/new-hook", "secret": "attacker-supplied"},
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		updated := decodeBody[notificationChannelResponse](t, w)
		if updated.Config["url"] != "https://example.com/new-hook" {
			t.Fatalf("want the URL updated, got %v", updated.Config["url"])
		}
		if updated.Config["secret"] != originalSecret {
			t.Fatalf("want the secret preserved (%v), got %v", originalSecret, updated.Config["secret"])
		}
	})
}

func TestRegenerateWebhookSecret(t *testing.T) {
	authH, channelsH, pool := testNotificationChannelHandler(t)

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doChannelRequest(t, http.MethodPost, channelsH.RegenerateWebhookSecret, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("rejects non-webhook channels", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u.access, "telegram", map[string]any{"chatId": "1"}, "X")

		w := doChannelRequest(t, http.MethodPost, channelsH.RegenerateWebhookSecret, u.access, c.ID, nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("another org's channel is not found (multi-tenancy)", func(t *testing.T) {
		u1 := signUpTestUser(t, authH, pool)
		u2 := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u1.access, "webhook", map[string]any{"url": "https://example.com/hook"}, "X")

		w := doChannelRequest(t, http.MethodPost, channelsH.RegenerateWebhookSecret, u2.access, c.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("rotates the secret without touching the URL", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u.access, "webhook", map[string]any{"url": "https://example.com/hook"}, "X")
		originalSecret := c.Config["secret"]

		w := doChannelRequest(t, http.MethodPost, channelsH.RegenerateWebhookSecret, u.access, c.ID, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		updated := decodeBody[notificationChannelResponse](t, w)
		if updated.Config["url"] != "https://example.com/hook" {
			t.Fatalf("want the URL unchanged, got %v", updated.Config["url"])
		}
		if updated.Config["secret"] == originalSecret {
			t.Fatal("want a new secret after regenerating")
		}
	})
}

func TestDeleteNotificationChannel(t *testing.T) {
	authH, channelsH, pool := testNotificationChannelHandler(t)

	t.Run("deletes a channel", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u.access, "telegram", map[string]any{"chatId": "1"}, "X")

		w := doChannelRequest(t, http.MethodDelete, channelsH.DeleteNotificationChannel, u.access, c.ID, nil)
		if w.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", w.Code, w.Body.String())
		}

		listW := doAuthed(t, http.MethodGet, channelsH.ListNotificationChannels, u.access, nil)
		channels := decodeBody[[]notificationChannelResponse](t, listW)
		if len(channels) != 0 {
			t.Fatalf("want channel gone after delete, got %d remaining", len(channels))
		}
	})

	t.Run("another org's channel is left alone (multi-tenancy)", func(t *testing.T) {
		u1 := signUpTestUser(t, authH, pool)
		u2 := signUpTestUser(t, authH, pool)
		c := createChannel(t, channelsH, u1.access, "telegram", map[string]any{"chatId": "1"}, "X")

		doChannelRequest(t, http.MethodDelete, channelsH.DeleteNotificationChannel, u2.access, c.ID, nil)

		listW := doAuthed(t, http.MethodGet, channelsH.ListNotificationChannels, u1.access, nil)
		channels := decodeBody[[]notificationChannelResponse](t, listW)
		if len(channels) != 1 {
			t.Fatalf("want the original owner's channel untouched, got %d", len(channels))
		}
	})
}

// TestTestNotificationChannel exercises the draft-config test endpoint, which
// doesn't require auth/org context — it tests a (possibly unsaved) type+config
// directly, same as the old TestTelegram/TestEmail design.
func TestTestNotificationChannel(t *testing.T) {
	_, channelsH, pool := testNotificationChannelHandler(t)

	t.Run("unsupported type", func(t *testing.T) {
		w := doJSON(t, channelsH.TestNotificationChannel, http.MethodPost, "/api/v1/notification-channels/test", testNotificationChannelRequest{
			Type: "slack", Config: map[string]any{},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("telegram missing chatId", func(t *testing.T) {
		w := doJSON(t, channelsH.TestNotificationChannel, http.MethodPost, "/api/v1/notification-channels/test", testNotificationChannelRequest{
			Type: "telegram", Config: map[string]any{},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("telegram with no bot token configured surfaces as a 502", func(t *testing.T) {
		w := doJSON(t, channelsH.TestNotificationChannel, http.MethodPost, "/api/v1/notification-channels/test", testNotificationChannelRequest{
			Type: "telegram", Config: map[string]any{"chatId": "123"},
		})
		if w.Code != http.StatusBadGateway {
			t.Fatalf("want 502, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "telegram_error" {
			t.Fatalf("want code telegram_error, got %q", body["code"])
		}
	})

	t.Run("email with no Resend API key configured no-ops successfully (ADR-012 dev behavior)", func(t *testing.T) {
		w := doJSON(t, channelsH.TestNotificationChannel, http.MethodPost, "/api/v1/notification-channels/test", testNotificationChannelRequest{
			Type: "email", Config: map[string]any{"email": "ops@example.com"},
		})
		if w.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("webhook missing url", func(t *testing.T) {
		w := doJSON(t, channelsH.TestNotificationChannel, http.MethodPost, "/api/v1/notification-channels/test", testNotificationChannelRequest{
			Type: "webhook", Config: map[string]any{},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("webhook posts a signed sample payload to a reachable endpoint", func(t *testing.T) {
		// validateChannelConfig requires https:// (US-1401), so this needs a
		// TLS test server — and a webhook.Client built around that server's
		// own *http.Client (which trusts its self-signed cert), since the
		// shared channelsH's client uses the default transport.
		var gotSig string
		var gotEvent map[string]any
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotSig = r.Header.Get(webhook.SignatureHeader)
			_ = json.NewDecoder(r.Body).Decode(&gotEvent)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		tlsTrustingH := NewNotificationChannelHandler(pool, telegram.NewClient(""), email.NewSender(""), webhook.NewClientWithHTTPClient(srv.Client()), slack.NewClient())
		w := doJSON(t, tlsTrustingH.TestNotificationChannel, http.MethodPost, "/api/v1/notification-channels/test", testNotificationChannelRequest{
			Type: "webhook", Config: map[string]any{"url": srv.URL},
		})
		if w.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", w.Code, w.Body.String())
		}
		if gotSig == "" {
			t.Fatal("want a signature header on the test request")
		}
		if gotEvent["eventType"] != "test" {
			t.Fatalf("want eventType \"test\", got %v", gotEvent["eventType"])
		}
	})

	t.Run("webhook with an unreachable endpoint surfaces as a 502", func(t *testing.T) {
		w := doJSON(t, channelsH.TestNotificationChannel, http.MethodPost, "/api/v1/notification-channels/test", testNotificationChannelRequest{
			Type: "webhook", Config: map[string]any{"url": "https://127.0.0.1:0"},
		})
		if w.Code != http.StatusBadGateway {
			t.Fatalf("want 502, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "webhook_error" {
			t.Fatalf("want code webhook_error, got %q", body["code"])
		}
	})
}
