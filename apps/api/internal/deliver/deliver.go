// Package deliver provides the HTTP send-and-classify mechanics shared by
// every alert channel that POSTs directly to a REST endpoint (webhook,
// Slack, Twilio): a shared request timeout and the one-attempt, no-retry
// (statusCode, err) return convention (US-1404/US-1704/US-1904), so a
// caller can record delivery status without inspecting err's type. Telegram
// isn't included — its API always returns HTTP 200 and reports failure via
// a JSON "ok" field instead of the status code, a different enough response
// shape that forcing it through this same classification would be a
// force-fit, not a simplification.
package deliver

import (
	"fmt"
	"net/http"
	"time"
)

// Timeout is the shared per-request timeout for every outbound alert
// channel — no channel should hang indefinitely on network delays.
const Timeout = 10 * time.Second

// Do sends req via client and classifies the response: a 2xx status is
// success (nil error); anything else is an error built by errFn from the
// response, so each channel can shape its own message (Twilio decodes a
// JSON error body; webhook/Slack just report the status code). Returns the
// response status code (0 if no response was ever received — timeout,
// connection error, or refused by an SSRF dial guard).
func Do(client *http.Client, req *http.Request, chanName string, errFn func(*http.Response) error) (statusCode int, err error) {
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%s request failed: %w", chanName, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, errFn(resp)
	}
	return resp.StatusCode, nil
}
