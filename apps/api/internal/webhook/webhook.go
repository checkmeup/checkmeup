// Package webhook sends signed event notifications to a user-configured URL
// (PRD webhook-alerts) — the generic alert channel alongside telegram and email, for
// users who want to wire monitor events into their own automation.
package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/checkmeup/checkmeup/internal/deliver"
	"github.com/checkmeup/checkmeup/internal/httpsafe"
)

// SignatureHeader carries the hex-encoded HMAC-SHA256 signature of the raw
// request body (PRD webhook-alerts), so the receiver can verify a request really came
// from checkmeup before acting on it.
const SignatureHeader = "X-Checkmeup-Signature"

type Client struct {
	httpClient *http.Client
}

// NewClient builds a Client hardened against SSRF: a webhook URL is
// arbitrary, user-supplied data (PRD webhook-alerts), so the destination has to be
// restricted to the public internet, not just "https" — otherwise a
// customer could point a webhook at internal infrastructure (e.g. a cloud
// metadata endpoint) and have checkmeup's own server make the request for
// them. See package httpsafe for the dial-time IP check and redirect
// handling shared with the Slack channel.
func NewClient() *Client {
	dialer := httpsafe.Dialer(deliver.Timeout)
	return &Client{httpClient: &http.Client{
		Timeout:       deliver.Timeout,
		Transport:     &http.Transport{DialContext: dialer.DialContext},
		CheckRedirect: httpsafe.RefuseRedirects,
	}}
}

// NewClientWithHTTPClient builds a Client around a caller-provided
// *http.Client, bypassing NewClient's SSRF protections (blockPrivateDial,
// refuseRedirects) entirely. Used by tests in other packages that need Send
// to reach a local httptest server (and, for TLS servers, to trust its
// self-signed certificate via e.g. httptest.NewTLSServer's Client()) —
// NewClient's default transport has no way to do that from outside this
// package, since httpClient is unexported. Not for production use: any
// caller outside a _test.go file should use NewClient.
func NewClientWithHTTPClient(hc *http.Client) *Client {
	return &Client{httpClient: hc}
}

// Event is the JSON body POSTed to a monitor's webhook URL (PRD webhook-alerts).
type Event struct {
	EventType        string `json:"eventType"` // "down" | "recovery" | "test"
	MonitorName      string `json:"monitorName"`
	MonitorType      string `json:"monitorType,omitempty"`      // "cron" | "uptime" | "ssl"
	Reason           string `json:"reason,omitempty"`           // down only
	DowntimeDuration string `json:"downtimeDuration,omitempty"` // recovery only
	Timestamp        string `json:"timestamp"`
	// The monitor's project (ADR projects); left out when it has none, so
	// an event for a monitor without a project is what it was before.
	ProjectID   string `json:"projectId,omitempty"`
	ProjectName string `json:"projectName,omitempty"`
}

// Send POSTs event to url, signed with secret. Reports the response status
// code (0 if no response was ever received, e.g. timeout or connection
// error) so the caller can record delivery status (PRD webhook-alerts) without
// inspecting err's type. One attempt only — no retries, matching the
// no-retry-storm requirement in PRD webhook-alerts.
func (c *Client) Send(url, secret string, event Event) (statusCode int, err error) {
	body, err := json.Marshal(event)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(SignatureHeader, Sign(body, secret))

	return deliver.Do(c.httpClient, req, "webhook", func(resp *http.Response) error {
		return fmt.Errorf("webhook endpoint returned HTTP %d", resp.StatusCode)
	})
}

// Sign returns the hex-encoded HMAC-SHA256 signature of body using secret.
// Exported so the Settings verification snippet and tests use the exact
// same computation that goes on the wire.
func Sign(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// GenerateSecret returns a new random 256-bit signing secret, hex-encoded —
// generated automatically the first time a webhook channel is saved (PRD webhook-alerts)
// and on demand when the user regenerates it (PRD webhook-alerts).
func GenerateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
