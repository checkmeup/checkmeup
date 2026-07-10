// Package slack sends alert notifications to a Slack Incoming Webhook URL
// (EP-17) — a thin Slack-specific formatter on top of the same HTTP delivery
// mechanics used by the generic webhook channel (EP-14). Config shape:
// {"url": "https://hooks.slack.com/services/..."}.
package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

// Message is the JSON body POSTed to a Slack Incoming Webhook URL.
// Block Kit blocks give a formatted card in Slack; the top-level text field
// is the plain-text fallback shown in notifications.
type Message struct {
	Text   string  `json:"text"`
	Blocks []block `json:"blocks,omitempty"`
}

type block struct {
	Type string   `json:"type"`
	Text *textObj `json:"text,omitempty"`
}

type textObj struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func headerBlock(text string) block {
	return block{Type: "header", Text: &textObj{Type: "plain_text", Text: text}}
}

func sectionBlock(md string) block {
	return block{Type: "section", Text: &textObj{Type: "mrkdwn", Text: md}}
}

// DownMessage builds a Slack alert for a monitor going down.
func DownMessage(monitorName, monitorType, reason string) Message {
	header := fmt.Sprintf("🔴 %s is down", monitorName)
	body := fmt.Sprintf("*Monitor:* %s\n*Type:* %s\n*Reason:* %s", monitorName, monitorType, reason)
	return Message{
		Text:   header,
		Blocks: []block{headerBlock(header), sectionBlock(body)},
	}
}

// RecoveryMessage builds a Slack alert for a monitor recovering.
func RecoveryMessage(monitorName, monitorType, downtime string) Message {
	header := fmt.Sprintf("✅ %s recovered", monitorName)
	body := fmt.Sprintf("*Monitor:* %s\n*Type:* %s\n*Downtime:* %s", monitorName, monitorType, downtime)
	return Message{
		Text:   header,
		Blocks: []block{headerBlock(header), sectionBlock(body)},
	}
}

// TestMessage returns a Slack message used by the "Send test message" button
// (US-1701) before a channel is saved.
func TestMessage() Message {
	return Message{
		Text:   "✅ Checkmeup is connected! You'll receive alerts here.",
		Blocks: []block{sectionBlock("✅ *Checkmeup is connected!* You'll receive alerts here.")},
	}
}

func isRestrictedIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast()
}

// blockPrivateDial rejects connections to loopback, private, link-local
// (which includes 169.254.169.254 cloud-metadata), unspecified, and multicast
// addresses. Mirrors the same guard in the webhook package — Slack webhook
// URLs are user-supplied, so the destination must be restricted to the public
// internet even though the URL is validated against hooks.slack.com on save.
func blockPrivateDial(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("slack: refusing to dial non-IP address %q", host)
	}
	if isRestrictedIP(ip) {
		return fmt.Errorf("slack: refusing to dial restricted address %s", ip)
	}
	return nil
}

func refuseRedirects(*http.Request, []*http.Request) error {
	return errors.New("slack: redirects are not followed")
}

// Client sends Slack messages via Incoming Webhooks.
type Client struct {
	httpClient *http.Client
}

// NewClient builds a Client hardened against SSRF: blockPrivateDial fires
// after DNS resolution (DNS-rebinding-safe) and refuseRedirects prevents a
// 3xx chain from retargeting to a blocked address. Consistent with the
// webhook channel client (US-1402 / EP-14).
func NewClient() *Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second, Control: blockPrivateDial}
	return &Client{httpClient: &http.Client{
		Timeout:       10 * time.Second,
		Transport:     &http.Transport{DialContext: dialer.DialContext},
		CheckRedirect: refuseRedirects,
	}}
}

// NewClientWithHTTPClient builds a Client around a caller-provided
// *http.Client — used by tests that need Send to reach a local httptest
// server. Not for production use.
func NewClientWithHTTPClient(hc *http.Client) *Client {
	return &Client{httpClient: hc}
}

// Send POSTs msg to the Slack Incoming Webhook at url. Returns the HTTP
// status code (0 if no response was received) so the caller can record
// delivery status (US-1704). One attempt only — no retries (US-1704).
func (c *Client) Send(url string, msg Message) (statusCode int, err error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("slack request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("slack webhook returned HTTP %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}
