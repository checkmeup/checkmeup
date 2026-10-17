package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/checkmeup/checkmeup/internal/rdap"
	"github.com/checkmeup/checkmeup/internal/whois"
	"github.com/checkmeup/checkmeup/internal/worker"
)

func toolRequest(t *testing.T, handle http.HandlerFunc, path, param, value string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	handle(w, httptest.NewRequest(http.MethodGet, path+"?"+param+"="+url.QueryEscape(value), nil))
	return w
}

func errorMessage(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body %q: %v", w.Body.String(), err)
	}
	return body.Error
}

func TestCheckSSL_ReadsTheCertificateOfTheHostInAURL(t *testing.T) {
	expires := time.Date(2027, 1, 2, 0, 0, 0, 0, time.UTC)
	var checked string
	h := NewToolsHandler(nil, nil)
	h.checkTLS = func(host string) (time.Time, string, int, error) {
		checked = host
		return expires, "R11", 77, nil
	}

	w := toolRequest(t, h.CheckSSL, "/api/v1/tools/ssl", "host", " https://Example.com:8443/path ")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if checked != "example.com" {
		t.Fatalf("checked host = %q, want %q", checked, "example.com")
	}
	var got sslToolResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := sslToolResponse{Host: "example.com", ExpiresAt: expires, DaysLeft: 77, Issuer: "R11"}
	if got != want {
		t.Fatalf("body = %+v, want %+v", got, want)
	}
}

func TestCheckSSL_RefusesWhatIsNotAHostname(t *testing.T) {
	h := NewToolsHandler(nil, nil)
	h.checkTLS = func(string) (time.Time, string, int, error) {
		t.Fatal("no check should run")
		return time.Time{}, "", 0, nil
	}
	for _, input := range []string{"", "   ", "localhost", "exa mple.com", "example.com?x=1", "user@example.com"} {
		t.Run(fmt.Sprintf("%q", input), func(t *testing.T) {
			w := toolRequest(t, h.CheckSSL, "/api/v1/tools/ssl", "host", input)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestCheckSSL_AFailedCheckGivesItsReason(t *testing.T) {
	h := NewToolsHandler(nil, nil)
	h.checkTLS = func(string) (time.Time, string, int, error) {
		return time.Time{}, "", 0, errors.New("x509: certificate has expired")
	}

	w := toolRequest(t, h.CheckSSL, "/api/v1/tools/ssl", "host", "expired.example.com")

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
	if msg := errorMessage(t, w); msg != "x509: certificate has expired" {
		t.Fatalf("error = %q, want the check's reason", msg)
	}
}

func TestCheckSSL_RefusesAPrivateAddress(t *testing.T) {
	h := NewToolsHandler(nil, nil)
	if h.checkTLS == nil {
		t.Fatal("want the real TLS check by default")
	}
	h.checkTLS = worker.CheckTLS

	w := toolRequest(t, h.CheckSSL, "/api/v1/tools/ssl", "host", "127.0.0.1")

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
	if msg := errorMessage(t, w); !strings.Contains(msg, "restricted address") {
		t.Fatalf("error = %q, want a refusal of the restricted address", msg)
	}
}

// fakeRDAP answers every lookup with body and counts the lookups.
func fakeRDAP(t *testing.T, body string) (*rdap.Client, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	return rdap.NewClientWithHTTPClient(srv.Client(), srv.URL+"/domain/"), &hits
}

func TestCheckDomain_ReadsExpiryAndRegistrar(t *testing.T) {
	rd, _ := fakeRDAP(t, `{
		"events": [{"eventAction": "expiration", "eventDate": "2027-06-01T00:00:00Z"}],
		"entities": [{"roles": ["registrar"], "vcardArray": ["vcard", [["fn", {}, "text", "Example Registrar"]]]}]
	}`)
	h := NewToolsHandler(rd, whois.NewClient())

	w := toolRequest(t, h.CheckDomain, "/api/v1/tools/domain", "domain", "https://example.com/")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}
	var got domainToolResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Domain != "example.com" || got.Registrar != "Example Registrar" ||
		!got.ExpiresAt.Equal(time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC)) || got.DaysLeft <= 0 {
		t.Fatalf("body = %+v, want example.com expiring 2027-06-01 at Example Registrar", got)
	}
}

func TestCheckDomain_ATLDWithNoPublishedExpiryIsRefusedWithoutALookup(t *testing.T) {
	rd, hits := fakeRDAP(t, `{}`)
	h := NewToolsHandler(rd, whois.NewClientWithAddr(func(string) string {
		t.Fatal("no WHOIS query should be made")
		return ""
	}))

	w := toolRequest(t, h.CheckDomain, "/api/v1/tools/domain", "domain", "example.de")

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
	if msg := errorMessage(t, w); msg == "" {
		t.Fatal("want the reason in the error")
	}
	if n := hits.Load(); n != 0 {
		t.Fatalf("RDAP lookups = %d, want 0", n)
	}
}

func TestCheckDomain_AnEmptyInputIsRefusedWithoutALookup(t *testing.T) {
	rd, hits := fakeRDAP(t, `{}`)
	h := NewToolsHandler(rd, whois.NewClient())

	w := toolRequest(t, h.CheckDomain, "/api/v1/tools/domain", "domain", "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if n := hits.Load(); n != 0 {
		t.Fatalf("RDAP lookups = %d, want 0", n)
	}
}
