// Package rdap looks up domain registration data (registrar, expiry) via
// RDAP (RFC 9082/9083) — the structured-JSON successor to WHOIS, used by the
// domain expiry monitor (EP-29) the same way internal/webhook and
// internal/telegram wrap their respective external services.
package rdap

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string // domain is appended directly, e.g. baseURL + "example.com"
}

// NewClient builds a Client against rdap.org, a public bootstrap redirector
// that resolves the correct authoritative RDAP server per TLD — checkmeup
// doesn't need to maintain its own IANA bootstrap registry mapping.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://rdap.org/domain/",
	}
}

// NewClientWithHTTPClient builds a Client around a caller-provided
// *http.Client and base URL — used by tests to point Lookup at a local
// httptest server instead of the live rdap.org service, so domain checks
// (including the success path, unlike SSL's performTLSCheck) are testable
// without a live network dependency.
func NewClientWithHTTPClient(hc *http.Client, baseURL string) *Client {
	return &Client{httpClient: hc, baseURL: baseURL}
}

// Result is the subset of an RDAP domain response checkmeup cares about.
type Result struct {
	Registrar string
	ExpiresAt time.Time
}

type rdapEvent struct {
	Action string `json:"eventAction"`
	Date   string `json:"eventDate"`
}

type rdapEntity struct {
	Roles      []string        `json:"roles"`
	Handle     string          `json:"handle"`
	VcardArray json.RawMessage `json:"vcardArray"`
}

type rdapResponse struct {
	Events   []rdapEvent  `json:"events"`
	Entities []rdapEntity `json:"entities"`
}

// Lookup queries the RDAP server for domain and returns its registrar and
// registration expiry. Returns an error for any non-2xx response, a
// malformed body, or a response with no "expiration" event — a monitor
// can't report days-until-expiry without one.
func (c *Client) Lookup(domain string) (Result, error) {
	resp, err := c.fetch(domain)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	var parsed rdapResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return Result{}, fmt.Errorf("rdap response decode failed: %w", err)
	}

	expiresAt, err := expirationEvent(parsed.Events)
	if err != nil {
		return Result{}, err
	}

	return Result{Registrar: findRegistrar(parsed.Entities), ExpiresAt: expiresAt}, nil
}

// fetch performs the RDAP HTTP request and validates the response status,
// leaving the body open for the caller to decode and close.
func (c *Client) fetch(domain string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+domain, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/rdap+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rdap request failed: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()
		return nil, errors.New("domain not found in registry (rdap 404)")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("rdap endpoint returned HTTP %d", resp.StatusCode)
	}
	return resp, nil
}

// expirationEvent finds the "expiration" event among an RDAP response's
// events and parses its date.
func expirationEvent(events []rdapEvent) (time.Time, error) {
	var expiresAt time.Time
	for _, ev := range events {
		if ev.Action != "expiration" {
			continue
		}
		if t, err := time.Parse(time.RFC3339, ev.Date); err == nil {
			expiresAt = t
		}
	}
	if expiresAt.IsZero() {
		return time.Time{}, errors.New("rdap response has no expiration event")
	}
	return expiresAt, nil
}

// findRegistrar returns the formatted name of the first entity with a
// "registrar" role, or "" if none is present.
func findRegistrar(entities []rdapEntity) string {
	for _, e := range entities {
		if hasRole(e.Roles, "registrar") {
			return registrarName(e)
		}
	}
	return ""
}

func hasRole(roles []string, want string) bool {
	for _, r := range roles {
		if r == want {
			return true
		}
	}
	return false
}

// registrarName extracts the "fn" (formatted name) property from an RDAP
// entity's jCard vcardArray — ["vcard", [["fn", {}, "text", "Name"], ...]].
// Falls back to the entity's handle if the vCard is missing or unparsable,
// since registrar vCards are optional in RDAP responses.
func registrarName(e rdapEntity) string {
	props, err := vcardProperties(e.VcardArray)
	if err != nil {
		return e.Handle
	}
	for _, p := range props {
		if value, ok := fnPropertyValue(p); ok {
			return value
		}
	}
	return e.Handle
}

// vcardProperties unwraps a jCard vcardArray — ["vcard", [[...], ...]] —
// into its list of properties.
func vcardProperties(raw json.RawMessage) ([][]json.RawMessage, error) {
	var vcard []json.RawMessage
	if err := json.Unmarshal(raw, &vcard); err != nil || len(vcard) < 2 {
		return nil, errors.New("vcard missing properties array")
	}
	var props [][]json.RawMessage
	if err := json.Unmarshal(vcard[1], &props); err != nil {
		return nil, err
	}
	return props, nil
}

// fnPropertyValue reads a jCard property of the form ["fn", {}, "text",
// "Name"] and returns its value, reporting ok=false for any other property
// or a blank value.
func fnPropertyValue(p []json.RawMessage) (string, bool) {
	if len(p) < 4 {
		return "", false
	}
	var propName string
	if err := json.Unmarshal(p[0], &propName); err != nil || propName != "fn" {
		return "", false
	}
	var value string
	if err := json.Unmarshal(p[3], &value); err != nil || value == "" {
		return "", false
	}
	return value, true
}
