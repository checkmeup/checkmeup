// Package webhook sends signed event notifications to a user-configured URL
// (EP-14) — the generic alert channel alongside telegram and email, for
// users who want to wire monitor events into their own automation.
package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

// SignatureHeader carries the hex-encoded HMAC-SHA256 signature of the raw
// request body (US-1403), so the receiver can verify a request really came
// from checkmeup before acting on it.
const SignatureHeader = "X-Checkmeup-Signature"

type Client struct {
	httpClient *http.Client
}

// NewClient builds a Client hardened against SSRF: a webhook URL is
// arbitrary, user-supplied data (US-1401), so the destination has to be
// restricted to the public internet, not just "https" — otherwise a
// customer could point a webhook at internal infrastructure (e.g. a cloud
// metadata endpoint) and have checkmeup's own server make the request for
// them.
//   - blockPrivateDial runs in net.Dialer.Control, which fires after DNS
//     resolution but before the TCP handshake, on the actual address being
//     connected to — so a hostname that resolves differently between
//     validation and connect (DNS rebinding) can't bypass it the way a
//     pre-flight URL/IP check could.
//   - refuseRedirects stops a 3xx response from retargeting the request to
//     a blocked address after the initial URL passed muster.
func NewClient() *Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second, Control: blockPrivateDial}
	return &Client{httpClient: &http.Client{
		Timeout:       10 * time.Second,
		Transport:     &http.Transport{DialContext: dialer.DialContext},
		CheckRedirect: refuseRedirects,
	}}
}

// blockPrivateDial rejects connections to loopback, private, link-local
// (which includes the 169.254.169.254 cloud-metadata address), unspecified,
// and multicast addresses.
func blockPrivateDial(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("webhook: refusing to dial non-IP address %q", host)
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return fmt.Errorf("webhook: refusing to dial restricted address %s", ip)
	}
	return nil
}

func refuseRedirects(*http.Request, []*http.Request) error {
	return errors.New("webhook: redirects are not followed")
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

// Event is the JSON body POSTed to a monitor's webhook URL (US-1402).
type Event struct {
	EventType        string `json:"eventType"` // "down" | "recovery" | "test"
	MonitorName      string `json:"monitorName"`
	MonitorType      string `json:"monitorType,omitempty"`      // "cron" | "uptime" | "ssl"
	Reason           string `json:"reason,omitempty"`           // down only
	DowntimeDuration string `json:"downtimeDuration,omitempty"` // recovery only
	Timestamp        string `json:"timestamp"`
}

// Send POSTs event to url, signed with secret. Reports the response status
// code (0 if no response was ever received, e.g. timeout or connection
// error) so the caller can record delivery status (US-1404) without
// inspecting err's type. One attempt only — no retries, matching the
// no-retry-storm requirement in US-1404.
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("webhook request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("webhook endpoint returned HTTP %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
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
// generated automatically the first time a webhook channel is saved (US-1401)
// and on demand when the user regenerates it (US-1403).
func GenerateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
