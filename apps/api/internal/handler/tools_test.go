package handler

import (
	"context"
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

// dnsSets builds a lookup's answer: values per record type, the rest empty.
func dnsSets(values map[string][]string) []worker.DNSRecordSet {
	var sets []worker.DNSRecordSet
	for _, typ := range []string{"A", "AAAA", "CNAME", "MX", "TXT", "NS"} {
		vs := values[typ]
		if vs == nil {
			vs = []string{}
		}
		sets = append(sets, worker.DNSRecordSet{Type: typ, Values: vs})
	}
	return sets
}

func TestCheckDNS_ListsEveryRecordTypeOfTheHostInAURL(t *testing.T) {
	var looked string
	h := NewToolsHandler(nil, nil)
	h.lookupDNS = func(_ context.Context, host string) []worker.DNSRecordSet {
		looked = host
		return dnsSets(map[string][]string{"A": {"93.184.215.14"}, "MX": {"10 mail.example.com"}})
	}

	w := toolRequest(t, h.CheckDNS, "/api/v1/tools/dns", "host", " https://Example.com/path ")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if looked != "example.com" {
		t.Fatalf("looked up %q, want %q", looked, "example.com")
	}
	want := `{"host":"example.com","records":[` +
		`{"type":"A","values":["93.184.215.14"]},{"type":"AAAA","values":[]},{"type":"CNAME","values":[]},` +
		`{"type":"MX","values":["10 mail.example.com"]},{"type":"TXT","values":[]},{"type":"NS","values":[]}]}`
	if got := strings.TrimSpace(w.Body.String()); got != want {
		t.Fatalf("body = %s\nwant   %s", got, want)
	}
}

func TestCheckDNS_RefusesWhatIsNotAHostname(t *testing.T) {
	h := NewToolsHandler(nil, nil)
	h.lookupDNS = func(context.Context, string) []worker.DNSRecordSet {
		t.Fatal("no lookup should run")
		return nil
	}
	for _, input := range []string{"", "localhost", "exa mple.com"} {
		t.Run(fmt.Sprintf("%q", input), func(t *testing.T) {
			w := toolRequest(t, h.CheckDNS, "/api/v1/tools/dns", "host", input)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestCheckDNS_AFailedLookupGivesItsReason(t *testing.T) {
	t.Run("a name with no records of any type", func(t *testing.T) {
		h := NewToolsHandler(nil, nil)
		h.lookupDNS = func(context.Context, string) []worker.DNSRecordSet { return dnsSets(nil) }

		w := toolRequest(t, h.CheckDNS, "/api/v1/tools/dns", "host", "missing.example.com")

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
		}
		if msg := errorMessage(t, w); msg != "missing.example.com has no DNS records" {
			t.Fatalf("error = %q, want %q", msg, "missing.example.com has no DNS records")
		}
	})

	t.Run("lookups that all failed", func(t *testing.T) {
		h := NewToolsHandler(nil, nil)
		h.lookupDNS = func(context.Context, string) []worker.DNSRecordSet {
			sets := dnsSets(nil)
			for i := range sets {
				sets[i].Error = "DNS lookup timeout"
			}
			return sets
		}

		w := toolRequest(t, h.CheckDNS, "/api/v1/tools/dns", "host", "slow.example.com")

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
		}
		if msg := errorMessage(t, w); msg != "DNS lookup timeout" {
			t.Fatalf("error = %q, want %q", msg, "DNS lookup timeout")
		}
	})
}

func explainCron(t *testing.T, h *ToolsHandler, expr string) (*httptest.ResponseRecorder, cronToolResponse) {
	t.Helper()
	w := toolRequest(t, h.ExplainCron, "/api/v1/tools/cron", "expression", expr)
	var got cronToolResponse
	if w.Code == http.StatusOK {
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode: %v", err)
		}
	}
	return w, got
}

func TestExplainCron_DescribesTheExpressionAndListsItsNextFiveRuns(t *testing.T) {
	h := NewToolsHandler(nil, nil)
	// A Friday, so the weekday schedule's next runs skip the weekend.
	h.now = func() time.Time { return time.Date(2026, 11, 20, 9, 30, 0, 0, time.UTC) }

	w, got := explainCron(t, h, "  0 9 * * MON-FRI ")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if got.Expression != "0 9 * * MON-FRI" || got.Description != "At 09:00, Monday through Friday" {
		t.Fatalf("got %q, %q; want %q, %q", got.Expression, got.Description, "0 9 * * MON-FRI", "At 09:00, Monday through Friday")
	}
	want := []time.Time{
		time.Date(2026, 11, 23, 9, 0, 0, 0, time.UTC),
		time.Date(2026, 11, 24, 9, 0, 0, 0, time.UTC),
		time.Date(2026, 11, 25, 9, 0, 0, 0, time.UTC),
		time.Date(2026, 11, 26, 9, 0, 0, 0, time.UTC),
		time.Date(2026, 11, 27, 9, 0, 0, 0, time.UTC),
	}
	if len(got.Runs) != len(want) {
		t.Fatalf("runs = %v, want %v", got.Runs, want)
	}
	for i := range want {
		if !got.Runs[i].Equal(want[i]) {
			t.Fatalf("runs = %v, want %v", got.Runs, want)
		}
	}
}

func TestExplainCron_DescribesEachCommonExpression(t *testing.T) {
	cases := map[string]string{
		"*/5 * * * *": "Every 5 minutes",
		"0 * * * *":   "Every hour",
		"0 0 * * *":   "At 00:00, every day",
		"0 9 * * 1-5": "At 09:00, Monday through Friday",
		"0 0 * * 0":   "At 00:00, on Sunday",
		"0 0 1 * *":   "At 00:00, on day 1 of the month",
	}
	h := NewToolsHandler(nil, nil)
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			if w, got := explainCron(t, h, expr); w.Code != http.StatusOK || got.Description != want {
				t.Fatalf("status %d, description %q; want %d, %q", w.Code, got.Description, http.StatusOK, want)
			}
		})
	}
}

func TestExplainCron_RefusesWhatTheScheduleParserCantRead(t *testing.T) {
	cases := map[string]string{
		"":                    "a cron expression has 5 fields (minute, hour, day of month, month, day of week), this one has 0",
		"0 9 * *":             "a cron expression has 5 fields (minute, hour, day of month, month, day of week), this one has 4",
		"60 * * * *":          "Minute: 60 is above the maximum, 59",
		"0 9 * * 7":           "Day of week: 7 is above the maximum, 6",
		"0 9 * FOO *":         `Month: "FOO" isn't a number or a name cron knows`,
		"every 1h":            "a cron expression has 5 fields (minute, hour, day of month, month, day of week), this one has 2",
		"CRON_TZ=UTC 0 * * *": `Minute: "CRON_TZ=UTC" isn't a value cron reads`,
		"@hourly":             "a cron expression has 5 fields (minute, hour, day of month, month, day of week), this one has 1",
		"0 0 30 2 *":          "this schedule never runs: none of its months has that day",
		"0 0 * * * *":         "a cron expression has 5 fields (minute, hour, day of month, month, day of week), this one has 6",
	}
	h := NewToolsHandler(nil, nil)
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			w, _ := explainCron(t, h, expr)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}
			if got := errorMessage(t, w); got != want {
				t.Fatalf("error = %q, want %q", got, want)
			}
		})
	}
}

func TestExplainCron_TheFirstRunIsWhenACronMonitorExpectsItsPing(t *testing.T) {
	now := time.Date(2026, 11, 21, 14, 7, 0, 0, time.UTC)
	h := NewToolsHandler(nil, nil)
	h.now = func() time.Time { return now }
	const grace = 5
	for _, expr := range []string{"*/5 * * * *", "0 9 * * 1-5", "15 3 1 * *", "0 0 1 1 *"} {
		t.Run(expr, func(t *testing.T) {
			_, got := explainCron(t, h, expr)
			want := computeNextPing(expr, now, grace)
			if len(got.Runs) == 0 || !got.Runs[0].Add(grace*time.Minute).Equal(want) {
				t.Fatalf("first run %v plus grace, want the monitor's %v", got.Runs, want)
			}
		})
	}
}
