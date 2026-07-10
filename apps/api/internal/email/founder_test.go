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

// newTestFounderNotifier builds a FounderNotifier whose resend client talks
// to a local httptest.Server instead of the real Resend API.
func newTestFounderNotifier(t *testing.T, handler http.HandlerFunc) (*FounderNotifier, *[]resend.SendEmailRequest) {
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

	f := NewFounderNotifier("test-key")
	f.client.BaseURL = base
	return f, &requests
}

func TestFounderNotifier_DevModeSkipsSending(t *testing.T) {
	f := NewFounderNotifier("")
	if f.client != nil {
		t.Fatalf("want nil client when apiKey is empty")
	}

	if err := f.SendFeatureSuggestion("user@example.com", "an idea"); err != nil {
		t.Fatalf("SendFeatureSuggestion: want nil error in dev mode, got %v", err)
	}
}

// ─── SendFeatureSuggestion (escapes user-supplied text) ─────────────────────

func TestSendFeatureSuggestion(t *testing.T) {
	f, requests := newTestFounderNotifier(t, okHandler)

	if err := f.SendFeatureSuggestion("user@example.com", "needs <b>bold</b>\nsecond line"); err != nil {
		t.Fatalf("SendFeatureSuggestion: %v", err)
	}

	req := (*requests)[0]
	if len(req.To) != 1 || req.To[0] != founderAddress {
		t.Fatalf("want To [%s], got %v", founderAddress, req.To)
	}
	if req.Subject != "Checkmeup: new feature suggestion" {
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

func TestFounderNotifier_PropagatesAPIError(t *testing.T) {
	f, _ := newTestFounderNotifier(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"resend is down"}`))
	})

	err := f.SendFeatureSuggestion("user@example.com", "an idea")
	if err == nil {
		t.Fatal("want an error when the API returns a non-2xx response")
	}
	if !strings.Contains(err.Error(), "resend is down") {
		t.Fatalf("want error to surface the API message, got %v", err)
	}
}
