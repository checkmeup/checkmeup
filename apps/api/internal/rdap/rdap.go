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
	req, err := http.NewRequest(http.MethodGet, c.baseURL+domain, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Accept", "application/rdap+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("rdap request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return Result{}, errors.New("domain not found in registry (rdap 404)")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, fmt.Errorf("rdap endpoint returned HTTP %d", resp.StatusCode)
	}

	var parsed rdapResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return Result{}, fmt.Errorf("rdap response decode failed: %w", err)
	}

	var expiresAt time.Time
	for _, ev := range parsed.Events {
		if ev.Action != "expiration" {
			continue
		}
		t, err := time.Parse(time.RFC3339, ev.Date)
		if err == nil {
			expiresAt = t
		}
	}
	if expiresAt.IsZero() {
		return Result{}, errors.New("rdap response has no expiration event")
	}

	var registrar string
	for _, e := range parsed.Entities {
		if hasRole(e.Roles, "registrar") {
			registrar = registrarName(e)
			break
		}
	}

	return Result{Registrar: registrar, ExpiresAt: expiresAt}, nil
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
	var vcard []json.RawMessage
	if err := json.Unmarshal(e.VcardArray, &vcard); err != nil || len(vcard) < 2 {
		return e.Handle
	}
	var props [][]json.RawMessage
	if err := json.Unmarshal(vcard[1], &props); err != nil {
		return e.Handle
	}
	for _, p := range props {
		if len(p) < 4 {
			continue
		}
		var propName string
		if err := json.Unmarshal(p[0], &propName); err != nil || propName != "fn" {
			continue
		}
		var value string
		if err := json.Unmarshal(p[3], &value); err == nil && value != "" {
			return value
		}
	}
	return e.Handle
}
