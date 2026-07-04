// Package twilio sends SMS alerts via Twilio's Messages REST API (EP-19,
// ADR-029) — the eighth alert channel, alongside email/telegram/webhook/slack.
package twilio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const messagesPathFormat = "/2010-04-01/Accounts/%s/Messages.json"

type Client struct {
	accountSID          string
	apiKeySID           string
	apiKeySecret        string
	messagingServiceSID string
	httpClient          *http.Client
	// baseURL defaults to Twilio's real API host; overridden in tests
	// (same package, so the field doesn't need to be exported) to point at
	// a local httptest server instead.
	baseURL string
}

// NewClient builds a Client authenticated with a scoped API Key (SID+Secret)
// rather than the primary Account Auth Token — see docs/twilio-setup.md and
// ADR-029 for why. accountSID is still required separately: it's part of the
// REST resource URL, not the auth pair.
func NewClient(accountSID, apiKeySID, apiKeySecret, messagingServiceSID string) *Client {
	return &Client{
		accountSID:          accountSID,
		apiKeySID:           apiKeySID,
		apiKeySecret:        apiKeySecret,
		messagingServiceSID: messagingServiceSID,
		httpClient:          &http.Client{Timeout: 10 * time.Second},
		baseURL:             "https://api.twilio.com",
	}
}

type errorBody struct {
	Message string `json:"message"`
}

// Send texts body to the given E.164 phone number via the configured
// Messaging Service (so Twilio picks the right sender for the destination —
// toll-free/10DLC for US/Canada, alphanumeric elsewhere). Reports the
// response status code (0 if no response was ever received, e.g. timeout or
// connection error) so the caller can record delivery status (US-1904)
// without inspecting err's type. One attempt only — no retries, matching the
// no-retry-storm requirement shared with webhook (US-1404).
func (c *Client) Send(to, body string) (statusCode int, err error) {
	if c.accountSID == "" || c.apiKeySID == "" || c.apiKeySecret == "" {
		return 0, fmt.Errorf("twilio not configured")
	}

	form := url.Values{}
	form.Set("To", to)
	form.Set("MessagingServiceSid", c.messagingServiceSID)
	form.Set("Body", body)

	endpoint := c.baseURL + fmt.Sprintf(messagesPathFormat, c.accountSID)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(c.apiKeySID, c.apiKeySecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("twilio request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var eb errorBody
		_ = json.NewDecoder(resp.Body).Decode(&eb)
		if eb.Message != "" {
			return resp.StatusCode, fmt.Errorf("twilio: %s", eb.Message)
		}
		return resp.StatusCode, fmt.Errorf("twilio returned HTTP %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}
