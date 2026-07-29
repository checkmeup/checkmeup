package handler

// Integration tests for the DNS monitor handlers in dns_monitors.go. Same
// conventions as port_monitors_test.go, which these largely mirror.

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
)

func doDNSMonitorRequest(t *testing.T, method string, handler http.HandlerFunc, access *http.Cookie, id string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader = http.NoBody
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "/api/v1/monitors/dns/"+id, r)
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", id)
	req.AddCookie(access)
	w := httptest.NewRecorder()
	apimiddleware.RequireAuth(testJWTSecret)(handler).ServeHTTP(w, req)
	return w
}

func seedDNSCheck(t *testing.T, pool *pgxpool.Pool, monitorID string, checkedAt time.Time, responseTimeMs int32, isUp bool) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		"INSERT INTO dns_checks (monitor_id, checked_at, response_time_ms, is_up) VALUES ($1, $2, $3, $4)",
		monitorID, checkedAt, responseTimeMs, isUp,
	); err != nil {
		t.Fatalf("seed dns check: %v", err)
	}
}

func seedDNSIncident(t *testing.T, pool *pgxpool.Pool, monitorID string, startedAt time.Time) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		"INSERT INTO dns_incidents (monitor_id, started_at) VALUES ($1, $2)", monitorID, startedAt,
	); err != nil {
		t.Fatalf("seed dns incident: %v", err)
	}
}

func TestListDNSMonitors(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/dns", nil)
		w := httptest.NewRecorder()
		monitorH.ListDNSMonitors(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("empty for a fresh org", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodGet, monitorH.ListDNSMonitors, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		list := decodeBody[[]dnsMonitorResponse](t, w)
		if len(list) != 0 {
			t.Fatalf("want empty list, got %d", len(list))
		}
	})

	t.Run("lists created monitors with no uptime stats yet", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		createDNSMonitor(t, monitorH, u.access, "Apex A record")
		createDNSMonitor(t, monitorH, u.access, "Mail MX")

		w := doAuthed(t, http.MethodGet, monitorH.ListDNSMonitors, u.access, nil)
		list := decodeBody[[]dnsMonitorResponse](t, w)
		if len(list) != 2 {
			t.Fatalf("want 2 monitors, got %d", len(list))
		}
		if list[0].Uptime24h != nil {
			t.Fatalf("want nil uptime24h with no checks yet, got %v", *list[0].Uptime24h)
		}
	})
}

func TestCreateDNSMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, monitorH.CreateDNSMonitor, http.MethodPost, "/api/v1/monitors/dns", createDNSMonitorRequest{Name: "x", Hostname: "example.com", RecordType: "A", IntervalMins: 10})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/monitors/dns", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(monitorH.CreateDNSMonitor)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cases := []struct {
			name string
			req  createDNSMonitorRequest
		}{
			{"missing name", createDNSMonitorRequest{Hostname: "example.com", RecordType: "A", IntervalMins: 10}},
			{"missing hostname", createDNSMonitorRequest{Name: "x", RecordType: "A", IntervalMins: 10}},
			{"invalid record type", createDNSMonitorRequest{Name: "x", Hostname: "example.com", RecordType: "PTR", IntervalMins: 10}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doAuthed(t, http.MethodPost, monitorH.CreateDNSMonitor, u.access, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("an interval below the plan minimum is rejected, not silently clamped", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateDNSMonitor, u.access, createDNSMonitorRequest{
			Name: "x", Hostname: "example.com", RecordType: "A", IntervalMins: 1, // Hobby's minimum is 5
		})
		if w.Code != http.StatusPaymentRequired {
			t.Fatalf("want 402, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("no expected value: baseline mode, with defaults applied", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateDNSMonitor, u.access, createDNSMonitorRequest{
			Name: "Fresh monitor", Hostname: "example.com/path:1234", RecordType: "a", IntervalMins: 5, MaxAlertsPerIncident: -1,
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[dnsMonitorResponse](t, w)
		if resp.Hostname != "example.com" {
			t.Fatalf("want hostname stripped of scheme/path/port, got %q", resp.Hostname)
		}
		if resp.RecordType != "A" {
			t.Fatalf("want record type normalized to uppercase A, got %q", resp.RecordType)
		}
		if resp.ExpectedValue != nil {
			t.Fatalf("want no expected value yet (baseline mode), got %q", *resp.ExpectedValue)
		}
		if resp.IntervalMins != 5 {
			t.Fatalf("want interval 5, got %d", resp.IntervalMins)
		}
		if resp.Status != "waiting" {
			t.Fatalf("want status waiting, got %q", resp.Status)
		}
		if resp.MaxAlertsPerIncident != 3 {
			t.Fatalf("want default max alerts 3, got %d", resp.MaxAlertsPerIncident)
		}
	})

	t.Run("an explicit expected value is pinned", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateDNSMonitor, u.access, createDNSMonitorRequest{
			Name: "Pinned A record", Hostname: "example.com", RecordType: "A", ExpectedValue: "1.2.3.4", IntervalMins: 10,
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[dnsMonitorResponse](t, w)
		if resp.ExpectedValue == nil || *resp.ExpectedValue != "1.2.3.4" {
			t.Fatalf("want expected value 1.2.3.4, got %v", resp.ExpectedValue)
		}
	})

	t.Run("plan limit enforced at 10 monitors on Hobby", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		for i := 0; i < 10; i++ {
			createDNSMonitor(t, monitorH, u.access, "Monitor")
		}
		w := doAuthed(t, http.MethodPost, monitorH.CreateDNSMonitor, u.access, createDNSMonitorRequest{
			Name: "One too many", Hostname: "example.com", RecordType: "A", IntervalMins: 10,
		})
		if w.Code != http.StatusPaymentRequired {
			t.Fatalf("want 402, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "plan_limit_reached" {
			t.Fatalf("want code plan_limit_reached, got %q", body["code"])
		}
	})
}

func TestGetDNSMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/dns/x", nil)
		w := httptest.NewRecorder()
		monitorH.GetDNSMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doDNSMonitorRequest(t, http.MethodGet, monitorH.GetDNSMonitor, u.access, "not-a-uuid", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doDNSMonitorRequest(t, http.MethodGet, monitorH.GetDNSMonitor, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant monitor is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createDNSMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doDNSMonitorRequest(t, http.MethodGet, monitorH.GetDNSMonitor, uB.access, mon.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 fetching org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success returns checks/incidents newest first and computed stats", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createDNSMonitor(t, monitorH, u.access, "Checked monitor")

		now := time.Now()
		seedDNSCheck(t, pool, mon.ID, now.Add(-3*time.Hour), 20, true)
		seedDNSCheck(t, pool, mon.ID, now.Add(-2*time.Hour), 0, false)
		seedDNSCheck(t, pool, mon.ID, now.Add(-1*time.Hour), 15, true)
		seedDNSIncident(t, pool, mon.ID, now.Add(-2*time.Hour))

		w := doDNSMonitorRequest(t, http.MethodGet, monitorH.GetDNSMonitor, u.access, mon.ID, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[struct {
			Monitor   dnsMonitorResponse    `json:"monitor"`
			ChartData []dnsCheckResponse    `json:"chartData"`
			Checks    []dnsCheckResponse    `json:"checks"`
			Incidents []dnsIncidentResponse `json:"incidents"`
			Stats     dnsStatsResponse      `json:"stats"`
		}](t, w)
		if resp.Monitor.ID != mon.ID {
			t.Fatalf("want monitor id %q, got %q", mon.ID, resp.Monitor.ID)
		}
		if len(resp.Checks) != 3 {
			t.Fatalf("want 3 checks, got %d", len(resp.Checks))
		}
		if !resp.Checks[0].IsUp {
			t.Fatal("want the most recent check (up) first")
		}
		if len(resp.Incidents) != 1 {
			t.Fatalf("want 1 incident, got %d", len(resp.Incidents))
		}
		if resp.Stats.Uptime24h == nil {
			t.Fatal("want uptime24h computed from the seeded checks")
		}
		want := float64(2) / float64(3) * 100.0
		if diff := *resp.Stats.Uptime24h - want; diff < -0.01 || diff > 0.01 {
			t.Fatalf("want uptime24h ~%.2f, got %.2f", want, *resp.Stats.Uptime24h)
		}
	})
}

func TestUpdateDNSMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/monitors/dns/x", nil)
		w := httptest.NewRecorder()
		monitorH.UpdateDNSMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createDNSMonitor(t, monitorH, u.access, "x")
		cases := []struct {
			name string
			req  updateDNSMonitorRequest
		}{
			{"missing name", updateDNSMonitorRequest{Hostname: "example.com", RecordType: "A"}},
			{"missing hostname", updateDNSMonitorRequest{Name: "x", RecordType: "A"}},
			{"invalid record type", updateDNSMonitorRequest{Name: "x", Hostname: "example.com", RecordType: "PTR"}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doDNSMonitorRequest(t, http.MethodPatch, monitorH.UpdateDNSMonitor, u.access, mon.ID, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doDNSMonitorRequest(t, http.MethodPatch, monitorH.UpdateDNSMonitor, u.access, "00000000-0000-0000-0000-000000000000", updateDNSMonitorRequest{
			Name: "x", Hostname: "example.com", RecordType: "A", IntervalMins: 10,
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant update is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createDNSMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doDNSMonitorRequest(t, http.MethodPatch, monitorH.UpdateDNSMonitor, uB.access, mon.ID, updateDNSMonitorRequest{
			Name: "hijacked", Hostname: "example.com", RecordType: "A", IntervalMins: 10,
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 updating org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success applies changes and defaults, including hostname/record type", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createDNSMonitor(t, monitorH, u.access, "Original name")

		w := doDNSMonitorRequest(t, http.MethodPatch, monitorH.UpdateDNSMonitor, u.access, mon.ID, updateDNSMonitorRequest{
			Name: "Renamed", Hostname: "new.example.com", RecordType: "MX", ExpectedValue: "mail.example.com",
			AlertsEnabled: false, IntervalMins: 30, MaxAlertsPerIncident: -1,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[dnsMonitorResponse](t, w)
		if resp.Name != "Renamed" {
			t.Fatalf("want name Renamed, got %q", resp.Name)
		}
		if resp.Hostname != "new.example.com" {
			t.Fatalf("want updated hostname, got %q", resp.Hostname)
		}
		if resp.RecordType != "MX" {
			t.Fatalf("want updated record type MX, got %q", resp.RecordType)
		}
		if resp.ExpectedValue == nil || *resp.ExpectedValue != "mail.example.com" {
			t.Fatalf("want updated expected value, got %v", resp.ExpectedValue)
		}
		if resp.BaselineCaptured {
			t.Fatal("want baselineCaptured false after an explicit edit (US-3905)")
		}
		if resp.AlertsEnabled {
			t.Fatal("want alertsEnabled false")
		}
		if resp.IntervalMins != 30 {
			t.Fatalf("want interval 30, got %d", resp.IntervalMins)
		}
		if resp.MaxAlertsPerIncident != 3 {
			t.Fatalf("want default max alerts 3, got %d", resp.MaxAlertsPerIncident)
		}
	})

	t.Run("clearing the expected value re-arms baseline mode", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateDNSMonitor, u.access, createDNSMonitorRequest{
			Name: "Pinned", Hostname: "example.com", RecordType: "A", ExpectedValue: "1.2.3.4", IntervalMins: 10,
		})
		mon := decodeBody[dnsMonitorResponse](t, w)

		editW := doDNSMonitorRequest(t, http.MethodPatch, monitorH.UpdateDNSMonitor, u.access, mon.ID, updateDNSMonitorRequest{
			Name: "Pinned", Hostname: "example.com", RecordType: "A", ExpectedValue: "", IntervalMins: 10,
		})
		if editW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", editW.Code, editW.Body.String())
		}
		resp := decodeBody[dnsMonitorResponse](t, editW)
		if resp.ExpectedValue != nil {
			t.Fatalf("want expected value cleared, got %q", *resp.ExpectedValue)
		}
		if resp.BaselineCaptured {
			t.Fatal("want baselineCaptured false until the next check re-captures it")
		}
	})

	t.Run("an interval below the plan minimum is rejected on update too", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createDNSMonitor(t, monitorH, u.access, "Hobby plan monitor")

		w := doDNSMonitorRequest(t, http.MethodPatch, monitorH.UpdateDNSMonitor, u.access, mon.ID, updateDNSMonitorRequest{
			Name: "x", Hostname: "example.com", RecordType: "A", IntervalMins: 1, // Hobby's minimum is 5
		})
		if w.Code != http.StatusPaymentRequired {
			t.Fatalf("want 402, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestPauseResumeDNSMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doDNSMonitorRequest(t, http.MethodPost, monitorH.PauseDNSMonitor, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("pause then resume round-trips status", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createDNSMonitor(t, monitorH, u.access, "Togglable")

		pauseW := doDNSMonitorRequest(t, http.MethodPost, monitorH.PauseDNSMonitor, u.access, mon.ID, nil)
		if pauseW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", pauseW.Code, pauseW.Body.String())
		}
		paused := decodeBody[dnsMonitorResponse](t, pauseW)
		if paused.Status != "paused" {
			t.Fatalf("want status paused, got %q", paused.Status)
		}

		resumeW := doDNSMonitorRequest(t, http.MethodPost, monitorH.ResumeDNSMonitor, u.access, mon.ID, nil)
		if resumeW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", resumeW.Code, resumeW.Body.String())
		}
		resumed := decodeBody[dnsMonitorResponse](t, resumeW)
		if resumed.Status != "waiting" {
			t.Fatalf("want status waiting after resume, got %q", resumed.Status)
		}
	})
}

func TestDeleteDNSMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("cross-tenant delete is a no-op, owner delete succeeds and cascades", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createDNSMonitor(t, monitorH, uA.access, "Org A monitor")
		seedDNSCheck(t, pool, mon.ID, time.Now(), 10, true)

		wrongDelete := doDNSMonitorRequest(t, http.MethodDelete, monitorH.DeleteDNSMonitor, uB.access, mon.ID, nil)
		if wrongDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204 (delete is a no-op exec regardless of match), got %d", wrongDelete.Code)
		}
		stillThere := doDNSMonitorRequest(t, http.MethodGet, monitorH.GetDNSMonitor, uA.access, mon.ID, nil)
		if stillThere.Code != http.StatusOK {
			t.Fatalf("want org A's monitor to survive org B's delete attempt, got %d", stillThere.Code)
		}

		ownerDelete := doDNSMonitorRequest(t, http.MethodDelete, monitorH.DeleteDNSMonitor, uA.access, mon.ID, nil)
		if ownerDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", ownerDelete.Code, ownerDelete.Body.String())
		}
		gone := doDNSMonitorRequest(t, http.MethodGet, monitorH.GetDNSMonitor, uA.access, mon.ID, nil)
		if gone.Code != http.StatusNotFound {
			t.Fatalf("want 404 after delete, got %d", gone.Code)
		}

		var checkCount int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM dns_checks WHERE monitor_id = $1", mon.ID).Scan(&checkCount); err != nil {
			t.Fatalf("count checks: %v", err)
		}
		if checkCount != 0 {
			t.Fatalf("want checks cascade-deleted with the monitor, got %d remaining", checkCount)
		}
	})
}

func TestValidateRecordType(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"A", false}, {"a", false}, {"AAAA", false}, {"CNAME", false},
		{"MX", false}, {"TXT", false}, {"NS", false},
		{"", true}, {"PTR", true}, {"SRV", true},
	}
	for _, tc := range cases {
		if _, err := validateRecordType(tc.in); (err != nil) != tc.wantErr {
			t.Fatalf("validateRecordType(%q) error = %v, wantErr %v", tc.in, err, tc.wantErr)
		}
	}
}
