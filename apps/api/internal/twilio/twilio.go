// Package twilio sends SMS alerts via Twilio's Messages REST API (PRD sms-alerts,
// ADR alerting) — the eighth alert channel, alongside email/telegram/webhook/slack.
package twilio

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/checkmeup/checkmeup/internal/deliver"
)

const messagesPathFormat = "/2010-04-01/Accounts/%s/Messages.json"

type Client struct {
	accountSID          string
	apiKeySID           string
	apiKeySecret        string
	messagingServiceSID string
	// statusCallbackURL is where Twilio reports each message's delivery
	// status (ADR alerting); "" sends without one.
	statusCallbackURL string
	httpClient        *http.Client
	// baseURL defaults to Twilio's real API host; overridden in tests
	// (same package, so the field doesn't need to be exported) to point at
	// a local httptest server instead.
	baseURL string
}

// NewClient builds a Client authenticated with a scoped API Key (SID+Secret)
// rather than the primary Account Auth Token — see docs/twilio-setup.md and
// ADR alerting for why. accountSID is still required separately: it's part of the
// REST resource URL, not the auth pair.
func NewClient(accountSID, apiKeySID, apiKeySecret, messagingServiceSID, statusCallbackURL string) *Client {
	return &Client{
		accountSID:          accountSID,
		apiKeySID:           apiKeySID,
		apiKeySecret:        apiKeySecret,
		messagingServiceSID: messagingServiceSID,
		statusCallbackURL:   statusCallbackURL,
		httpClient:          &http.Client{Timeout: deliver.Timeout},
		baseURL:             "https://api.twilio.com",
	}
}

// NewClientWithBaseURL builds a configured Client that sends to baseURL
// through hc instead of Twilio's API, for tests outside this package that
// fake Twilio with an httptest server.
func NewClientWithBaseURL(baseURL string, hc *http.Client) *Client {
	c := NewClient("AC_test", "SK_test", "secret", "MG_test", "")
	c.baseURL, c.httpClient = baseURL, hc
	return c
}

// StatusCallbackToken is the secret path segment of the status callback URL,
// derived from key. It authenticates Twilio's callbacks in place of
// X-Twilio-Signature, which only the primary Auth Token can verify — and
// this app holds a scoped API key instead (ADR alerting).
func StatusCallbackToken(key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte("twilio-sms-status"))
	return hex.EncodeToString(mac.Sum(nil))
}

// StatusCallbackPath is the route Twilio posts status callbacks to, followed
// by "/" and the StatusCallbackToken.
const StatusCallbackPath = "/webhook/twilio/sms-status"

// StatusCallbackURL is the status callback URL for an app served at
// baseURL, or "" when baseURL isn't https — a local dev server Twilio
// can't reach.
func StatusCallbackURL(baseURL, key string) string {
	if !strings.HasPrefix(baseURL, "https://") {
		return ""
	}
	return strings.TrimSuffix(baseURL, "/") + StatusCallbackPath + "/" + StatusCallbackToken(key)
}

// APIError is Twilio's rejection of a send, carrying its message ("The 'To'
// number ... is not a valid phone number") for the channel's delivery status.
type APIError struct {
	Message string
}

func (e *APIError) Error() string { return "twilio: " + e.Message }

// messageResponse holds the fields Send reads from Twilio's reply: the new
// message's SID on success, the error message otherwise.
type messageResponse struct {
	SID     string `json:"sid"`
	Message string `json:"message"`
}

// Send texts body to the given E.164 phone number via the configured
// Messaging Service (so Twilio picks the right sender for the destination —
// toll-free/10DLC for US/Canada, alphanumeric elsewhere). Returns the
// message's SID, which its status callbacks carry, and the response status
// code (0 if no response was ever received, e.g. timeout or connection
// error) so the caller can record delivery status (PRD sms-alerts) without
// inspecting err's type. One attempt only — no retries, matching the
// no-retry-storm requirement shared with webhook (PRD webhook-alerts).
func (c *Client) Send(to, body string) (sid string, statusCode int, err error) {
	if c.accountSID == "" || c.apiKeySID == "" || c.apiKeySecret == "" {
		return "", 0, fmt.Errorf("twilio not configured")
	}

	form := url.Values{}
	form.Set("To", to)
	form.Set("MessagingServiceSid", c.messagingServiceSID)
	form.Set("Body", body)
	if c.statusCallbackURL != "" {
		form.Set("StatusCallback", c.statusCallbackURL)
	}

	endpoint := c.baseURL + fmt.Sprintf(messagesPathFormat, c.accountSID)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", 0, err
	}
	req.SetBasicAuth(c.apiKeySID, c.apiKeySecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("twilio request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var mr messageResponse
	_ = json.NewDecoder(resp.Body).Decode(&mr)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if mr.Message != "" {
			return "", resp.StatusCode, &APIError{Message: mr.Message}
		}
		return "", resp.StatusCode, fmt.Errorf("twilio returned HTTP %d", resp.StatusCode)
	}
	return mr.SID, resp.StatusCode, nil
}
