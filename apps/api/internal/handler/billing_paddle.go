package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func (h *BillingHandler) priceIDForPlan(plan, cycle string) string {
	annual := cycle == cycleAnnual
	switch plan {
	case "solo":
		if annual {
			return h.cfg.PaddleSoloAnnualPriceID
		}
		return h.cfg.PaddleSoloPriceID
	case "startup":
		if annual {
			return h.cfg.PaddleStartupAnnualPriceID
		}
		return h.cfg.PaddleStartupPriceID
	case "enterprise":
		if annual {
			return h.cfg.PaddleEnterpriseAnnualPriceID
		}
		return h.cfg.PaddleEnterprisePriceID
	}
	return ""
}

// paddleAPIBase returns Paddle's production or sandbox API host — these are
// entirely separate environments (separate API keys, price IDs, customers),
// so a sandbox key against the production host (or vice versa) just fails.
func (h *BillingHandler) paddleAPIBase() string {
	if h.cfg.PaddleEnvironment == "sandbox" {
		return "https://sandbox-api.paddle.com"
	}
	return "https://api.paddle.com"
}

// createPaddleTransaction creates a Paddle transaction server-side so
// custom_data.org_id comes from the authenticated session (orgIDFrom),
// never from client input — the frontend only ever sees the resulting
// transaction ID, which it hands to Paddle.js to open the checkout overlay.
func (h *BillingHandler) createPaddleTransaction(orgID uuid.UUID, priceID string) (string, error) {
	payload := map[string]any{
		"items": []map[string]any{
			{"price_id": priceID, "quantity": 1},
		},
		"custom_data": map[string]string{"org_id": orgID.String()},
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, h.paddleAPIBase()+"/transactions", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+h.cfg.PaddleAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Data.ID == "" {
		return "", fmt.Errorf("empty transaction ID from Paddle")
	}
	return result.Data.ID, nil
}

// paddleAPIError wraps a non-2xx Paddle response so callers can distinguish
// a client-side conflict (4xx — e.g. "subscription already has a pending
// scheduled change") from a real failure (5xx, network error), instead of
// collapsing every failure into the same generic 500.
type paddleAPIError struct {
	StatusCode int
	Body       string
}

func (e *paddleAPIError) Error() string {
	return fmt.Sprintf("paddle API error: status %d, body: %s", e.StatusCode, e.Body)
}

// detail extracts Paddle's own error code/detail from its standard error
// envelope (`{"error":{"code":"...","detail":"..."}}`), so callers can
// surface the real reason instead of guessing at one.
func (e *paddleAPIError) detail() (code, detail string) {
	var parsed struct {
		Error struct {
			Code   string `json:"code"`
			Detail string `json:"detail"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(e.Body), &parsed); err != nil {
		return "", ""
	}
	return parsed.Error.Code, parsed.Error.Detail
}

// updatePaddleSubscription changes an existing subscription to a new price —
// used for upgrades/downgrades between paid tiers. "prorated_immediately"
// charges/credits the difference right away rather than waiting for the
// next billing cycle, matching how the one-off CreateCheckout upgrade path
// (from Hobby) takes effect immediately too.
func (h *BillingHandler) updatePaddleSubscription(subscriptionID, priceID string) error {
	payload := map[string]any{
		"items": []map[string]any{
			{"price_id": priceID, "quantity": 1},
		},
		"proration_billing_mode": "prorated_immediately",
	}
	b, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/subscriptions/%s", h.paddleAPIBase(), subscriptionID)
	req, _ := http.NewRequest(http.MethodPatch, url, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+h.cfg.PaddleAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
		return &paddleAPIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}
	return nil
}

// cancelPaddleSubscription schedules cancellation for the end of the current
// billing period (not immediately) — the org keeps paid-tier access until
// then, matching the "Access until <date>" copy already shown in the billing
// UI for a cancelled subscription. The plan itself only flips to Hobby once
// Paddle's subscription.canceled webhook actually fires at period end.
func (h *BillingHandler) cancelPaddleSubscription(subscriptionID string) error {
	payload := map[string]any{"effective_from": "next_billing_period"}
	b, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/subscriptions/%s/cancel", h.paddleAPIBase(), subscriptionID)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+h.cfg.PaddleAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
		return &paddleAPIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}
	return nil
}

// createPaddlePortalSession generates a single-use, short-lived customer
// portal URL — Paddle explicitly documents these as not cacheable, unlike
// LemonSqueezy's static my-orders link, so this is called fresh every time.
func (h *BillingHandler) createPaddlePortalSession(customerID string) (string, error) {
	url := fmt.Sprintf("%s/customers/%s/portal-sessions", h.paddleAPIBase(), customerID)
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Authorization", "Bearer "+h.cfg.PaddleAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Data struct {
			URLs struct {
				General struct {
					Overview string `json:"overview"`
				} `json:"general"`
			} `json:"urls"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data.URLs.General.Overview, nil
}

// verifyPaddleSignature checks the Paddle-Signature header, formatted as
// "ts=<unix_timestamp>;h1=<hex_hmac>". The signed string is "ts:rawBody" —
// see https://developer.paddle.com/webhooks/signature-verification.
func verifyPaddleSignature(body []byte, header, secret string) bool {
	var ts, h1 string
	for _, part := range strings.Split(header, ";") {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		switch k {
		case "ts":
			ts = v
		case "h1":
			h1 = v
		}
	}
	if ts == "" || h1 == "" {
		return false
	}
	if _, err := strconv.ParseInt(ts, 10, 64); err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + ":"))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(h1))
}
