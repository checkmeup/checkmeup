package slack

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func unrestrictedTestClient() *Client {
	return NewClientWithHTTPClient(&http.Client{Timeout: 10 * time.Second})
}

func TestSend_PostsJSONBody(t *testing.T) {
	var gotBody []byte
	var gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := unrestrictedTestClient()
	msg := DownMessage("My API", "uptime", "HTTP 500")
	code, err := c.Send(srv.URL, msg)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if code != http.StatusOK {
		t.Fatalf("code = %d, want 200", code)
	}
	if gotContentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", gotContentType)
	}
	var sent Message
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if sent.Text == "" {
		t.Fatal("want non-empty text field in Slack message")
	}
	if len(sent.Blocks) == 0 {
		t.Fatal("want at least one Block Kit block")
	}
}

func TestSend_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := unrestrictedTestClient()
	code, err := c.Send(srv.URL, TestMessage())
	if err == nil {
		t.Fatal("want error for a non-2xx response")
	}
	if code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", code)
	}
}

func TestSend_ConnectionFailureReturnsZeroStatus(t *testing.T) {
	c := unrestrictedTestClient()
	code, err := c.Send("http://127.0.0.1:0", TestMessage())
	if err == nil {
		t.Fatal("want error when the request never reaches a server")
	}
	if code != 0 {
		t.Fatalf("code = %d, want 0 for a connection failure", code)
	}
}

func TestDownMessage_ContainsMonitorDetails(t *testing.T) {
	msg := DownMessage("payments", "uptime", "HTTP 503")
	if msg.Text == "" {
		t.Fatal("want non-empty text")
	}
	body := msg.Blocks[len(msg.Blocks)-1].Text.Text
	for _, want := range []string{"payments", "uptime", "HTTP 503"} {
		found := false
		for _, b := range msg.Blocks {
			if b.Text != nil && contains(b.Text.Text, want) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("want %q to appear in Block Kit blocks; body = %q", want, body)
		}
	}
}

func TestRecoveryMessage_ContainsDowntime(t *testing.T) {
	msg := RecoveryMessage("payments", "uptime", "5m 30s")
	found := false
	for _, b := range msg.Blocks {
		if b.Text != nil && contains(b.Text.Text, "5m 30s") {
			found = true
			break
		}
	}
	if !found {
		t.Error("want downtime duration in recovery message blocks")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
