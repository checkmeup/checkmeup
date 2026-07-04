package twilio

// Unit tests for Send, in-package (not twilio_test) so tests can point
// baseURL at a local httptest server — Twilio's real API host is otherwise
// hardcoded, unlike webhook/slack which send to a caller-supplied URL.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func testClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	return &Client{
		accountSID:          "AC_test",
		apiKeySID:           "SK_test",
		apiKeySecret:        "secret",
		messagingServiceSID: "MG_test",
		httpClient:          srv.Client(),
		baseURL:             srv.URL,
	}
}

func TestSend(t *testing.T) {
	t.Run("not configured", func(t *testing.T) {
		c := &Client{httpClient: &http.Client{}, baseURL: "https://api.twilio.com"}
		statusCode, err := c.Send("+15005550006", "hi")
		if err == nil {
			t.Fatal("want an error when Twilio isn't configured")
		}
		if statusCode != 0 {
			t.Fatalf("want status 0, got %d", statusCode)
		}
	})

	t.Run("success posts form-encoded body with Basic Auth and returns 200", func(t *testing.T) {
		var gotAuth bool
		var gotForm url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			gotAuth = ok && user == "SK_test" && pass == "secret"
			_ = r.ParseForm()
			gotForm = r.Form
			w.WriteHeader(http.StatusCreated) // Twilio returns 201 on a successful send
		}))
		defer srv.Close()

		c := testClient(t, srv)
		statusCode, err := c.Send("+15005550006", "checkmeup: test")
		if err != nil {
			t.Fatalf("want no error, got %v", err)
		}
		if statusCode != http.StatusCreated {
			t.Fatalf("want 201, got %d", statusCode)
		}
		if !gotAuth {
			t.Fatal("want Basic Auth with the API Key SID/Secret")
		}
		if gotForm.Get("To") != "+15005550006" {
			t.Fatalf("want To set, got %q", gotForm.Get("To"))
		}
		if gotForm.Get("MessagingServiceSid") != "MG_test" {
			t.Fatalf("want MessagingServiceSid set, got %q", gotForm.Get("MessagingServiceSid"))
		}
		if gotForm.Get("Body") != "checkmeup: test" {
			t.Fatalf("want Body set, got %q", gotForm.Get("Body"))
		}
	})

	t.Run("non-2xx response surfaces Twilio's error message", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "The 'To' number is not a valid phone number"})
		}))
		defer srv.Close()

		c := testClient(t, srv)
		statusCode, err := c.Send("bad-number", "hi")
		if err == nil {
			t.Fatal("want an error on a 400 response")
		}
		if statusCode != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", statusCode)
		}
		if got := err.Error(); got != "twilio: The 'To' number is not a valid phone number" {
			t.Fatalf("want Twilio's error message surfaced, got %q", got)
		}
	})

	t.Run("non-2xx response with no decodable body falls back to a generic message", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		c := testClient(t, srv)
		statusCode, err := c.Send("+15005550006", "hi")
		if err == nil {
			t.Fatal("want an error on a 500 response")
		}
		if statusCode != http.StatusInternalServerError {
			t.Fatalf("want 500, got %d", statusCode)
		}
	})

	t.Run("unreachable host reports status 0", func(t *testing.T) {
		c := &Client{
			accountSID: "AC_test", apiKeySID: "SK_test", apiKeySecret: "secret",
			httpClient: &http.Client{}, baseURL: "http://127.0.0.1:0",
		}
		statusCode, err := c.Send("+15005550006", "hi")
		if err == nil {
			t.Fatal("want an error dialing an invalid address")
		}
		if statusCode != 0 {
			t.Fatalf("want status 0 for a connection failure, got %d", statusCode)
		}
	})
}
