package handler

// Integration tests for the port monitor handlers in port_monitors.go. Same
// conventions as uptime_monitors_test.go (which these largely mirror, minus
// the keyword/JSON-assertion machinery that doesn't apply to a TCP check).

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

	"github.com/checkmeup/checkmeup/internal/db"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
)

func doPortMonitorRequest(t *testing.T, method string, handler http.HandlerFunc, access *http.Cookie, id string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader = http.NoBody
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "/api/v1/monitors/port/"+id, r)
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", id)
	req.AddCookie(access)
	w := httptest.NewRecorder()
	apimiddleware.RequireAuth(testJWTSecret)(handler).ServeHTTP(w, req)
	return w
}

func seedPortCheck(t *testing.T, pool *pgxpool.Pool, monitorID string, checkedAt time.Time, responseTimeMs int32, isUp bool) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		"INSERT INTO port_checks (monitor_id, checked_at, response_time_ms, is_up) VALUES ($1, $2, $3, $4)",
		monitorID, checkedAt, responseTimeMs, isUp,
	); err != nil {
		t.Fatalf("seed port check: %v", err)
	}
}

func seedPortIncident(t *testing.T, pool *pgxpool.Pool, monitorID string, startedAt time.Time) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		"INSERT INTO port_incidents (monitor_id, started_at) VALUES ($1, $2)", monitorID, startedAt,
	); err != nil {
		t.Fatalf("seed port incident: %v", err)
	}
}

func TestListPortMonitors(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/port", nil)
		w := httptest.NewRecorder()
		monitorH.ListPortMonitors(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("empty for a fresh org", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodGet, monitorH.ListPortMonitors, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		list := decodeBody[[]portMonitorResponse](t, w)
		if len(list) != 0 {
			t.Fatalf("want empty list, got %d", len(list))
		}
	})

	t.Run("lists created monitors with no uptime stats yet", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		createPortMonitor(t, monitorH, u.access, "SMTP")
		createPortMonitor(t, monitorH, u.access, "Postgres")

		w := doAuthed(t, http.MethodGet, monitorH.ListPortMonitors, u.access, nil)
		list := decodeBody[[]portMonitorResponse](t, w)
		if len(list) != 2 {
			t.Fatalf("want 2 monitors, got %d", len(list))
		}
		if list[0].Uptime24h != nil {
			t.Fatalf("want nil uptime24h with no checks yet, got %v", *list[0].Uptime24h)
		}
	})
}

func TestCreatePortMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, monitorH.CreatePortMonitor, http.MethodPost, "/api/v1/monitors/port", createPortMonitorRequest{Name: "x", Host: "example.com", Port: 443, IntervalMins: 10})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/monitors/port", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(monitorH.CreatePortMonitor)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cases := []struct {
			name string
			req  createPortMonitorRequest
		}{
			{"missing name", createPortMonitorRequest{Host: "example.com", Port: 443, IntervalMins: 10}},
			{"missing host", createPortMonitorRequest{Name: "x", Port: 443, IntervalMins: 10}},
			{"port too low", createPortMonitorRequest{Name: "x", Host: "example.com", Port: 0, IntervalMins: 10}},
			{"port too high", createPortMonitorRequest{Name: "x", Host: "example.com", Port: 65536, IntervalMins: 10}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doAuthed(t, http.MethodPost, monitorH.CreatePortMonitor, u.access, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("an interval below the plan minimum is rejected, not silently clamped", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreatePortMonitor, u.access, createPortMonitorRequest{
			Name: "x", Host: "example.com", Port: 443, IntervalMins: 1, // Hobby's minimum is 5
		})
		if w.Code != http.StatusPaymentRequired {
			t.Fatalf("want 402, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success at the plan minimum interval, with defaults applied", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreatePortMonitor, u.access, createPortMonitorRequest{
			Name: "Fresh monitor", Host: "example.com/path:1234", Port: 5432, IntervalMins: 5, MaxAlertsPerIncident: -1,
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[portMonitorResponse](t, w)
		if resp.Host != "example.com" {
			t.Fatalf("want host stripped of scheme/path/port, got %q", resp.Host)
		}
		if resp.Port != 5432 {
			t.Fatalf("want port 5432, got %d", resp.Port)
		}
		if resp.ExpectedState != "open" {
			t.Fatalf("want default expected state open, got %q", resp.ExpectedState)
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

	t.Run("expected state closed is honored", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreatePortMonitor, u.access, createPortMonitorRequest{
			Name: "Firewalled DB", Host: "example.com", Port: 5432, ExpectedState: "closed", IntervalMins: 10,
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[portMonitorResponse](t, w)
		if resp.ExpectedState != "closed" {
			t.Fatalf("want expected state closed, got %q", resp.ExpectedState)
		}
	})

	t.Run("plan limit enforced at 10 monitors on Hobby", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		for i := 0; i < 10; i++ {
			createPortMonitor(t, monitorH, u.access, "Monitor")
		}
		w := doAuthed(t, http.MethodPost, monitorH.CreatePortMonitor, u.access, createPortMonitorRequest{
			Name: "One too many", Host: "example.com", Port: 443, IntervalMins: 10,
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

func TestGetPortMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/port/x", nil)
		w := httptest.NewRecorder()
		monitorH.GetPortMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doPortMonitorRequest(t, http.MethodGet, monitorH.GetPortMonitor, u.access, "not-a-uuid", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doPortMonitorRequest(t, http.MethodGet, monitorH.GetPortMonitor, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant monitor is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createPortMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doPortMonitorRequest(t, http.MethodGet, monitorH.GetPortMonitor, uB.access, mon.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 fetching org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success returns checks/incidents newest first and computed stats", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createPortMonitor(t, monitorH, u.access, "Checked monitor")

		now := time.Now()
		seedPortCheck(t, pool, mon.ID, now.Add(-3*time.Hour), 20, true)
		seedPortCheck(t, pool, mon.ID, now.Add(-2*time.Hour), 0, false)
		seedPortCheck(t, pool, mon.ID, now.Add(-1*time.Hour), 15, true)
		seedPortIncident(t, pool, mon.ID, now.Add(-2*time.Hour))

		w := doPortMonitorRequest(t, http.MethodGet, monitorH.GetPortMonitor, u.access, mon.ID, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[struct {
			Monitor   portMonitorResponse    `json:"monitor"`
			ChartData []portCheckResponse    `json:"chartData"`
			Checks    []portCheckResponse    `json:"checks"`
			Incidents []portIncidentResponse `json:"incidents"`
			Stats     portStatsResponse      `json:"stats"`
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

func TestUpdatePortMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/monitors/port/x", nil)
		w := httptest.NewRecorder()
		monitorH.UpdatePortMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createPortMonitor(t, monitorH, u.access, "x")
		cases := []struct {
			name string
			req  updatePortMonitorRequest
		}{
			{"missing name", updatePortMonitorRequest{Host: "example.com", Port: 443}},
			{"missing host", updatePortMonitorRequest{Name: "x", Port: 443}},
			{"invalid port", updatePortMonitorRequest{Name: "x", Host: "example.com", Port: 70000}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doPortMonitorRequest(t, http.MethodPatch, monitorH.UpdatePortMonitor, u.access, mon.ID, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doPortMonitorRequest(t, http.MethodPatch, monitorH.UpdatePortMonitor, u.access, "00000000-0000-0000-0000-000000000000", updatePortMonitorRequest{
			Name: "x", Host: "example.com", Port: 443, IntervalMins: 10,
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant update is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createPortMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doPortMonitorRequest(t, http.MethodPatch, monitorH.UpdatePortMonitor, uB.access, mon.ID, updatePortMonitorRequest{
			Name: "hijacked", Host: "example.com", Port: 443, IntervalMins: 10,
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 updating org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success applies changes and defaults, including host/port (unlike SSL/domain)", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createPortMonitor(t, monitorH, u.access, "Original name")

		w := doPortMonitorRequest(t, http.MethodPatch, monitorH.UpdatePortMonitor, u.access, mon.ID, updatePortMonitorRequest{
			Name: "Renamed", Host: "new.example.com", Port: 8443, ExpectedState: "closed", AlertsEnabled: false, IntervalMins: 30, MaxAlertsPerIncident: -1,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[portMonitorResponse](t, w)
		if resp.Name != "Renamed" {
			t.Fatalf("want name Renamed, got %q", resp.Name)
		}
		if resp.Host != "new.example.com" {
			t.Fatalf("want updated host, got %q", resp.Host)
		}
		if resp.Port != 8443 {
			t.Fatalf("want updated port, got %d", resp.Port)
		}
		if resp.ExpectedState != "closed" {
			t.Fatalf("want updated expected state closed, got %q", resp.ExpectedState)
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

	t.Run("an interval below the plan minimum is rejected on update too", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createPortMonitor(t, monitorH, u.access, "Hobby plan monitor")

		w := doPortMonitorRequest(t, http.MethodPatch, monitorH.UpdatePortMonitor, u.access, mon.ID, updatePortMonitorRequest{
			Name: "x", Host: "example.com", Port: 443, IntervalMins: 1, // Hobby's minimum is 5
		})
		if w.Code != http.StatusPaymentRequired {
			t.Fatalf("want 402, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestPauseResumePortMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doPortMonitorRequest(t, http.MethodPost, monitorH.PausePortMonitor, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("pause then resume round-trips status", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createPortMonitor(t, monitorH, u.access, "Togglable")

		pauseW := doPortMonitorRequest(t, http.MethodPost, monitorH.PausePortMonitor, u.access, mon.ID, nil)
		if pauseW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", pauseW.Code, pauseW.Body.String())
		}
		paused := decodeBody[portMonitorResponse](t, pauseW)
		if paused.Status != "paused" {
			t.Fatalf("want status paused, got %q", paused.Status)
		}

		resumeW := doPortMonitorRequest(t, http.MethodPost, monitorH.ResumePortMonitor, u.access, mon.ID, nil)
		if resumeW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", resumeW.Code, resumeW.Body.String())
		}
		resumed := decodeBody[portMonitorResponse](t, resumeW)
		if resumed.Status != "waiting" {
			t.Fatalf("want status waiting after resume, got %q", resumed.Status)
		}
	})
}

func TestDeletePortMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("cross-tenant delete is a no-op, owner delete succeeds and cascades", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createPortMonitor(t, monitorH, uA.access, "Org A monitor")
		seedPortCheck(t, pool, mon.ID, time.Now(), 10, true)

		wrongDelete := doPortMonitorRequest(t, http.MethodDelete, monitorH.DeletePortMonitor, uB.access, mon.ID, nil)
		if wrongDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204 (delete is a no-op exec regardless of match), got %d", wrongDelete.Code)
		}
		stillThere := doPortMonitorRequest(t, http.MethodGet, monitorH.GetPortMonitor, uA.access, mon.ID, nil)
		if stillThere.Code != http.StatusOK {
			t.Fatalf("want org A's monitor to survive org B's delete attempt, got %d", stillThere.Code)
		}

		ownerDelete := doPortMonitorRequest(t, http.MethodDelete, monitorH.DeletePortMonitor, uA.access, mon.ID, nil)
		if ownerDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", ownerDelete.Code, ownerDelete.Body.String())
		}
		gone := doPortMonitorRequest(t, http.MethodGet, monitorH.GetPortMonitor, uA.access, mon.ID, nil)
		if gone.Code != http.StatusNotFound {
			t.Fatalf("want 404 after delete, got %d", gone.Code)
		}

		var checkCount int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM port_checks WHERE monitor_id = $1", mon.ID).Scan(&checkCount); err != nil {
			t.Fatalf("count checks: %v", err)
		}
		if checkCount != 0 {
			t.Fatalf("want checks cascade-deleted with the monitor, got %d remaining", checkCount)
		}
	})
}

func TestValidatePort(t *testing.T) {
	cases := []struct {
		port    int32
		wantErr bool
	}{
		{1, false},
		{443, false},
		{65535, false},
		{0, true},
		{-1, true},
		{65536, true},
	}
	for _, tc := range cases {
		if err := validatePort(tc.port); (err != nil) != tc.wantErr {
			t.Fatalf("validatePort(%d) error = %v, wantErr %v", tc.port, err, tc.wantErr)
		}
	}
}

func TestParseExpectedState(t *testing.T) {
	if got := parseExpectedState("closed"); got != db.PortExpectedStateClosed {
		t.Fatalf("want closed, got %q", got)
	}
	for _, in := range []string{"open", "", "bogus"} {
		if got := parseExpectedState(in); got != db.PortExpectedStateOpen {
			t.Fatalf("parseExpectedState(%q) = %q, want open (default)", in, got)
		}
	}
}
