package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// unrestrictedTestClient builds a Client with no SSRF protections, for tests
// that need to reach a local httptest server and aren't themselves testing
// httpsafe.BlockPrivateDial/httpsafe.RefuseRedirects.
func unrestrictedTestClient() *Client {
	return NewClientWithHTTPClient(&http.Client{Timeout: 10 * time.Second})
}

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

	c := unrestrictedTestClient()
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

	c := unrestrictedTestClient()
	code, err := c.Send(srv.URL, "shh", Event{EventType: "down"})
	if err == nil {
		t.Fatal("want error for a non-2xx response")
	}
	if code != http.StatusInternalServerError {
		t.Fatalf("code = %d, want 500", code)
	}
}

func TestSend_ConnectionFailureReturnsZeroStatus(t *testing.T) {
	c := unrestrictedTestClient()
	code, err := c.Send("http://127.0.0.1:0", "shh", Event{EventType: "down"})
	if err == nil {
		t.Fatal("want error when the request never reaches a server")
	}
	if code != 0 {
		t.Fatalf("code = %d, want 0 for a connection failure", code)
	}
}

// ─── SSRF protections (NewClient only — unrestrictedTestClient bypasses them) ─

func TestSend_BlocksRestrictedAddresses(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"loopback", "https://127.0.0.1:1/hook"},
		{"loopback IPv6", "https://[::1]:1/hook"},
		{"private RFC1918", "https://10.0.0.1:1/hook"},
		{"link-local / cloud metadata", "https://169.254.169.254/hook"},
		{"unspecified", "https://0.0.0.0:1/hook"},
	}
	c := NewClient()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, err := c.Send(tc.url, "shh", Event{EventType: "down"})
			if err == nil {
				t.Fatalf("want NewClient to refuse to dial %s", tc.url)
			}
			if code != 0 {
				t.Fatalf("code = %d, want 0 for a blocked address", code)
			}
		})
	}
}

// RefuseRedirects and BlockPrivateDial themselves are tested in
// package httpsafe, where they now live.
