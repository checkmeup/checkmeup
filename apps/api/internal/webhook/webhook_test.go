package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSign(t *testing.T) {
	sig := Sign([]byte(`{"a":1}`), "secret")

	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte(`{"a":1}`))
	want := hex.EncodeToString(mac.Sum(nil))

	if sig != want {
		t.Fatalf("Sign() = %q, want %q", sig, want)
	}
}

func TestGenerateSecret(t *testing.T) {
	a, err := GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret: %v", err)
	}
	b, err := GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret: %v", err)
	}
	if a == b {
		t.Fatal("want two distinct secrets")
	}
	if len(a) != 64 { // 32 bytes hex-encoded
		t.Fatalf("len(secret) = %d, want 64", len(a))
	}
}

func TestSend_PostsSignedJSONBody(t *testing.T) {
	var gotBody []byte
	var gotSig string
	var gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotSig = r.Header.Get(SignatureHeader)
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient()
	event := Event{EventType: "down", MonitorName: "API", MonitorType: "uptime", Reason: "HTTP 500", Timestamp: "2026-06-22T00:00:00Z"}
	code, err := c.Send(srv.URL, "shh", event)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if code != http.StatusOK {
		t.Fatalf("code = %d, want 200", code)
	}
	if gotContentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", gotContentType)
	}
	if gotSig != Sign(gotBody, "shh") {
		t.Fatalf("signature header does not match HMAC of the actual body sent")
	}
}

func TestSend_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClient()
	code, err := c.Send(srv.URL, "shh", Event{EventType: "down"})
	if err == nil {
		t.Fatal("want error for a non-2xx response")
	}
	if code != http.StatusInternalServerError {
		t.Fatalf("code = %d, want 500", code)
	}
}

func TestSend_ConnectionFailureReturnsZeroStatus(t *testing.T) {
	c := NewClient()
	code, err := c.Send("http://127.0.0.1:0", "shh", Event{EventType: "down"})
	if err == nil {
		t.Fatal("want error when the request never reaches a server")
	}
	if code != 0 {
		t.Fatalf("code = %d, want 0 for a connection failure", code)
	}
}
