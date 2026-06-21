package email

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/resend/resend-go/v2"
)

// newTestSender builds a Sender whose resend client talks to a local
// httptest.Server instead of the real Resend API.
func newTestSender(t *testing.T, handler http.HandlerFunc) (*Sender, *[]resend.SendEmailRequest) {
	t.Helper()

	var requests []resend.SendEmailRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req resend.SendEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		requests = append(requests, req)
		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	base, err := url.Parse(srv.URL + "/")
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}

	s := NewSender("test-key")
	s.client.BaseURL = base
	return s, &requests
}

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resend.SendEmailResponse{Id: "test-id"})
}

// ─── dev mode (no RESEND_API_KEY): every Send* is a silent no-op ───────────

func TestSender_DevModeSkipsSending(t *testing.T) {
	s := NewSender("")
	if s.client != nil {
		t.Fatalf("want nil client when apiKey is empty")
	}

	if err := s.SendPasswordReset("user@example.com", "https://checkmeup.net/reset"); err != nil {
		t.Fatalf("SendPasswordReset: want nil error in dev mode, got %v", err)
	}
	if err := s.SendAlertEmail("user@example.com", "subject", "<p>html</p>"); err != nil {
		t.Fatalf("SendAlertEmail: want nil error in dev mode, got %v", err)
	}
	if err := s.SendTestAlertEmail("user@example.com"); err != nil {
		t.Fatalf("SendTestAlertEmail: want nil error in dev mode, got %v", err)
	}
	if err := s.SendFeatureSuggestion("user@example.com", "an idea"); err != nil {
		t.Fatalf("SendFeatureSuggestion: want nil error in dev mode, got %v", err)
	}
}

// ─── SendPasswordReset (local httptest.Server, no live network) ────────────

func TestSendPasswordReset(t *testing.T) {
	s, requests := newTestSender(t, okHandler)

	if err := s.SendPasswordReset("user@example.com", "https://checkmeup.net/reset/abc123"); err != nil {
		t.Fatalf("SendPasswordReset: %v", err)
	}

	if len(*requests) != 1 {
		t.Fatalf("want 1 request, got %d", len(*requests))
	}
	req := (*requests)[0]
	if req.From != fromAddress {
		t.Fatalf("want From %q, got %q", fromAddress, req.From)
	}
	if len(req.To) != 1 || req.To[0] != "user@example.com" {
		t.Fatalf("want To [user@example.com], got %v", req.To)
	}
	if req.Subject != "Reset your checkmeup password" {
		t.Fatalf("unexpected subject: %q", req.Subject)
	}
	if !strings.Contains(req.Html, "https://checkmeup.net/reset/abc123") {
		t.Fatalf("want html to contain the reset URL, got %q", req.Html)
	}
}

// ─── SendAlertEmail / SendTestAlertEmail ────────────────────────────────────

func TestSendAlertEmail(t *testing.T) {
	s, requests := newTestSender(t, okHandler)

	if err := s.SendAlertEmail("alerts@example.com", "DOWN: my-monitor", "<p>it's down</p>"); err != nil {
		t.Fatalf("SendAlertEmail: %v", err)
	}

	req := (*requests)[0]
	if req.To[0] != "alerts@example.com" || req.Subject != "DOWN: my-monitor" || req.Html != "<p>it's down</p>" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestSendTestAlertEmail(t *testing.T) {
	s, requests := newTestSender(t, okHandler)

	if err := s.SendTestAlertEmail("alerts@example.com"); err != nil {
		t.Fatalf("SendTestAlertEmail: %v", err)
	}

	req := (*requests)[0]
	if req.Subject != "checkmeup: test alert" {
		t.Fatalf("unexpected subject: %q", req.Subject)
	}
	if !strings.Contains(req.Html, "connected") {
		t.Fatalf("want canned 'connected' copy, got %q", req.Html)
	}
}

// ─── SendFeatureSuggestion (escapes user-supplied text) ─────────────────────

func TestSendFeatureSuggestion(t *testing.T) {
	s, requests := newTestSender(t, okHandler)

	if err := s.SendFeatureSuggestion("user@example.com", "needs <b>bold</b>\nsecond line"); err != nil {
		t.Fatalf("SendFeatureSuggestion: %v", err)
	}

	req := (*requests)[0]
	if len(req.To) != 1 || req.To[0] != founderAddress {
		t.Fatalf("want To [%s], got %v", founderAddress, req.To)
	}
	if req.Subject != "checkmeup: new feature suggestion" {
		t.Fatalf("unexpected subject: %q", req.Subject)
	}
	if strings.Contains(req.Html, "<b>bold</b>") {
		t.Fatalf("want suggestion text HTML-escaped, got %q", req.Html)
	}
	if !strings.Contains(req.Html, "&lt;b&gt;bold&lt;/b&gt;") {
		t.Fatalf("want escaped tags in html, got %q", req.Html)
	}
	if !strings.Contains(req.Html, "needs &lt;b&gt;bold&lt;/b&gt;<br>second line") {
		t.Fatalf("want newline converted to <br>, got %q", req.Html)
	}
}

// ─── error propagation from the Resend API ──────────────────────────────────

func TestSend_PropagatesAPIError(t *testing.T) {
	s, _ := newTestSender(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"resend is down"}`))
	})

	err := s.SendAlertEmail("alerts@example.com", "subject", "<p>html</p>")
	if err == nil {
		t.Fatal("want an error when the API returns a non-2xx response")
	}
	if !strings.Contains(err.Error(), "resend is down") {
		t.Fatalf("want error to surface the API message, got %v", err)
	}
}
