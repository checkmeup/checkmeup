package telegram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// rewriteTransport redirects every request to target, regardless of the
// host/scheme baked into the request URL. This lets us point Client's
// hardcoded https://api.telegram.org URLs at a local httptest.Server.
type rewriteTransport struct {
	target *url.URL
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.target.Scheme
	req.URL.Host = t.target.Host
	return http.DefaultTransport.RoundTrip(req)
}

// newTestClient builds a Client whose requests are routed to a local
// httptest.Server instead of the real Telegram Bot API.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *[]*http.Request, *[][]byte) {
	t.Helper()

	var requests []*http.Request
	var bodies [][]byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		requests = append(requests, r)
		bodies = append(bodies, body)
		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}

	c := NewClient("test-token")
	c.httpClient.Transport = &rewriteTransport{target: target}
	return c, &requests, &bodies
}

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiResponse{OK: true})
}

// ─── NewClient ───────────────────────────────────────────────────────────

func TestNewClient(t *testing.T) {
	c := NewClient("abc123")
	if c.botToken != "abc123" {
		t.Fatalf("botToken = %q, want %q", c.botToken, "abc123")
	}
	if c.httpClient == nil {
		t.Fatal("want non-nil httpClient")
	}
}

// ─── SetWebhook ──────────────────────────────────────────────────────────

func TestSetWebhook_NoBotToken(t *testing.T) {
	c := NewClient("")
	if err := c.SetWebhook("https://example.com/hook", "secret"); err == nil {
		t.Fatal("want error when bot token is not configured")
	}
}

func TestSetWebhook_SendsURLAndSecret(t *testing.T) {
	c, requests, bodies := newTestClient(t, okHandler)

	if err := c.SetWebhook("https://example.com/hook", "shh"); err != nil {
		t.Fatalf("SetWebhook: %v", err)
	}

	if len(*requests) != 1 {
		t.Fatalf("want 1 request, got %d", len(*requests))
	}
	req := (*requests)[0]
	if !strings.Contains(req.URL.Path, "/bottest-token/setWebhook") {
		t.Fatalf("unexpected request path: %q", req.URL.Path)
	}

	var payload map[string]string
	if err := json.Unmarshal((*bodies)[0], &payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if payload["url"] != "https://example.com/hook" {
		t.Fatalf("url = %q, want %q", payload["url"], "https://example.com/hook")
	}
	if payload["secret_token"] != "shh" {
		t.Fatalf("secret_token = %q, want %q", payload["secret_token"], "shh")
	}
}

func TestSetWebhook_OmitsSecretWhenEmpty(t *testing.T) {
	c, _, bodies := newTestClient(t, okHandler)

	if err := c.SetWebhook("https://example.com/hook", ""); err != nil {
		t.Fatalf("SetWebhook: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal((*bodies)[0], &payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if _, ok := payload["secret_token"]; ok {
		t.Fatalf("want no secret_token key when secret is empty, got %v", payload)
	}
}

func TestSetWebhook_PropagatesAPIError(t *testing.T) {
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(apiResponse{OK: false, Description: "invalid url"})
	})

	err := c.SetWebhook("not-a-url", "")
	if err == nil {
		t.Fatal("want error when the API reports ok=false")
	}
	if !strings.Contains(err.Error(), "invalid url") {
		t.Fatalf("want error to surface the API description, got %v", err)
	}
}

func TestSetWebhook_RequestFailure(t *testing.T) {
	c := NewClient("test-token")
	c.httpClient.Transport = &rewriteTransport{target: &url.URL{Scheme: "http", Host: "127.0.0.1:0"}}

	if err := c.SetWebhook("https://example.com/hook", ""); err == nil {
		t.Fatal("want error when the HTTP request itself fails")
	}
}

// ─── SendMessage ─────────────────────────────────────────────────────────

func TestSendMessage_NoBotToken(t *testing.T) {
	c := NewClient("")
	if err := c.SendMessage("123", "hello"); err == nil {
		t.Fatal("want error when bot token is not configured")
	}
}

func TestSendMessage_SendsChatIDAndText(t *testing.T) {
	c, requests, bodies := newTestClient(t, okHandler)

	if err := c.SendMessage("42", "<b>hi</b>"); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}

	req := (*requests)[0]
	if !strings.Contains(req.URL.Path, "/bottest-token/sendMessage") {
		t.Fatalf("unexpected request path: %q", req.URL.Path)
	}

	var payload sendMessageRequest
	if err := json.Unmarshal((*bodies)[0], &payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if payload.ChatID != "42" {
		t.Fatalf("chat_id = %q, want %q", payload.ChatID, "42")
	}
	if payload.Text != "<b>hi</b>" {
		t.Fatalf("text = %q, want %q", payload.Text, "<b>hi</b>")
	}
	if payload.ParseMode != "HTML" {
		t.Fatalf("parse_mode = %q, want %q", payload.ParseMode, "HTML")
	}
}

func TestSendMessage_PropagatesAPIError(t *testing.T) {
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(apiResponse{OK: false, Description: "chat not found"})
	})

	err := c.SendMessage("42", "hi")
	if err == nil {
		t.Fatal("want error when the API reports ok=false")
	}
	if !strings.Contains(err.Error(), "chat not found") {
		t.Fatalf("want error to surface the API description, got %v", err)
	}
}

func TestSendMessage_DecodeFailure(t *testing.T) {
	c, _, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	})

	if err := c.SendMessage("42", "hi"); err == nil {
		t.Fatal("want error when the response body cannot be decoded")
	}
}

// ─── HandleUpdate ────────────────────────────────────────────────────────

func TestHandleUpdate_NilMessageIsNoop(t *testing.T) {
	c, requests, _ := newTestClient(t, okHandler)

	c.HandleUpdate(WebhookUpdate{UpdateID: 1, Message: nil})

	if len(*requests) != 0 {
		t.Fatalf("want no requests for a nil message, got %d", len(*requests))
	}
}

func TestHandleUpdate_IgnoresNonStartCommands(t *testing.T) {
	c, requests, _ := newTestClient(t, okHandler)

	update := WebhookUpdate{Message: &WebhookMessage{Text: "hello there"}}
	c.HandleUpdate(update)

	if len(*requests) != 0 {
		t.Fatalf("want no requests for a non-/start message, got %d", len(*requests))
	}
}

func TestHandleUpdate_RepliesToStartWithChatID(t *testing.T) {
	c, requests, bodies := newTestClient(t, okHandler)

	update := WebhookUpdate{Message: &WebhookMessage{Text: "/start"}}
	update.Message.Chat.ID = 987654321
	c.HandleUpdate(update)

	if len(*requests) != 1 {
		t.Fatalf("want 1 reply request, got %d", len(*requests))
	}

	var payload sendMessageRequest
	if err := json.Unmarshal((*bodies)[0], &payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if payload.ChatID != "987654321" {
		t.Fatalf("chat_id = %q, want %q", payload.ChatID, "987654321")
	}
	if !strings.Contains(payload.Text, "987654321") {
		t.Fatalf("want reply text to include the chat id, got %q", payload.Text)
	}
}
