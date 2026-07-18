package handler

// Integration tests for the uptime monitor handlers in uptime_monitors.go.
// Same conventions as monitors_test.go/ssl_monitors_test.go (which these
// largely mirror, since uptime monitors share the MonitorHandler/
// testMonitorHandler/createUptimeMonitor helpers already defined there and
// in maintenance_test.go): real Postgres (ADR-010), package handler so
// unexported request/response types are reused directly.

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
)

// doUptimeMonitorRequest authenticates via RequireAuth + the access cookie
// and injects the chi "id" URL param these handlers read via chi.URLParam.
func doUptimeMonitorRequest(t *testing.T, method string, handler http.HandlerFunc, access *http.Cookie, id string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader = http.NoBody
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "/api/v1/monitors/uptime/"+id, r)
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", id)
	req.AddCookie(access)
	w := httptest.NewRecorder()
	apimiddleware.RequireAuth(testJWTSecret)(handler).ServeHTTP(w, req)
	return w
}

func seedUptimeCheck(t *testing.T, pool *pgxpool.Pool, monitorID string, checkedAt time.Time, statusCode *int32, responseTimeMs int32, isUp bool) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		"INSERT INTO uptime_checks (monitor_id, checked_at, status_code, response_time_ms, is_up) VALUES ($1, $2, $3, $4, $5)",
		monitorID, checkedAt, statusCode, responseTimeMs, isUp,
	); err != nil {
		t.Fatalf("seed uptime check: %v", err)
	}
}

func seedUptimeIncident(t *testing.T, pool *pgxpool.Pool, monitorID string, startedAt time.Time) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		"INSERT INTO uptime_incidents (monitor_id, started_at) VALUES ($1, $2)", monitorID, startedAt,
	); err != nil {
		t.Fatalf("seed uptime incident: %v", err)
	}
}

func TestListUptimeMonitors(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/uptime", nil)
		w := httptest.NewRecorder()
		monitorH.ListUptimeMonitors(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("empty for a fresh org", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodGet, monitorH.ListUptimeMonitors, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		list := decodeBody[[]uptimeMonitorResponse](t, w)
		if len(list) != 0 {
			t.Fatalf("want empty list, got %d", len(list))
		}
	})

	t.Run("lists created monitors with no uptime stats yet", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		createUptimeMonitor(t, monitorH, u.access, "API")
		createUptimeMonitor(t, monitorH, u.access, "Website")

		w := doAuthed(t, http.MethodGet, monitorH.ListUptimeMonitors, u.access, nil)
		list := decodeBody[[]uptimeMonitorResponse](t, w)
		if len(list) != 2 {
			t.Fatalf("want 2 monitors, got %d", len(list))
		}
		if list[0].Uptime24h != nil {
			t.Fatalf("want nil uptime24h with no checks yet, got %v", *list[0].Uptime24h)
		}
	})
}

func TestCreateUptimeMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, monitorH.CreateUptimeMonitor, http.MethodPost, "/api/v1/monitors/uptime", createUptimeMonitorRequest{Name: "x", URL: "https://example.com", IntervalMins: 10})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/monitors/uptime", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(monitorH.CreateUptimeMonitor)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cases := []struct {
			name string
			req  createUptimeMonitorRequest
		}{
			{"missing name", createUptimeMonitorRequest{URL: "https://example.com", IntervalMins: 10}},
			{"missing URL", createUptimeMonitorRequest{Name: "x", IntervalMins: 10}},
			{"URL without scheme", createUptimeMonitorRequest{Name: "x", URL: "example.com", IntervalMins: 10}},
			{"keyword too long", createUptimeMonitorRequest{Name: "x", URL: "https://example.com", IntervalMins: 10, Keyword: strings.Repeat("a", 501)}},
			{"invalid HTTP method", createUptimeMonitorRequest{
				Name: "x", URL: "https://example.com", IntervalMins: 10,
				MaxResponseTimeMs: 10000, HttpMethod: "PUT", AcceptedStatusCodes: []int32{200},
			}},
			{"maxResponseTimeMs below 1000", createUptimeMonitorRequest{
				Name: "x", URL: "https://example.com", IntervalMins: 10,
				MaxResponseTimeMs: 999, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
			}},
			{"maxResponseTimeMs above 30000", createUptimeMonitorRequest{
				Name: "x", URL: "https://example.com", IntervalMins: 10,
				MaxResponseTimeMs: 30001, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
			}},
			{"status code out of range", createUptimeMonitorRequest{
				Name: "x", URL: "https://example.com", IntervalMins: 10,
				MaxResponseTimeMs: 10000, HttpMethod: "GET", AcceptedStatusCodes: []int32{600},
			}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doAuthed(t, http.MethodPost, monitorH.CreateUptimeMonitor, u.access, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("keyword monitoring works on the Hobby (free) plan", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateUptimeMonitor, u.access, createUptimeMonitorRequest{
			Name: "x", URL: "https://example.com", IntervalMins: 10, Keyword: "Welcome",
			MaxResponseTimeMs: 10000, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[uptimeMonitorResponse](t, w)
		if resp.Keyword == nil || *resp.Keyword != "Welcome" {
			t.Fatalf("want keyword set on Hobby, got %v", resp.Keyword)
		}
	})

	t.Run("an interval below the plan minimum is rejected, not silently clamped", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateUptimeMonitor, u.access, createUptimeMonitorRequest{
			Name: "x", URL: "https://example.com", IntervalMins: 1, // Hobby's minimum is 5
			MaxResponseTimeMs: 10000, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
		})
		if w.Code != http.StatusPaymentRequired {
			t.Fatalf("want 402, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success at the plan minimum interval, with defaults applied", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateUptimeMonitor, u.access, createUptimeMonitorRequest{
			Name: "Fresh monitor", URL: "https://example.com/health", IntervalMins: 5, MaxAlertsPerIncident: -1,
			MaxResponseTimeMs: 10000, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[uptimeMonitorResponse](t, w)
		if resp.IntervalMins != 5 {
			t.Fatalf("want interval 5, got %d", resp.IntervalMins)
		}
		if resp.Status != "waiting" {
			t.Fatalf("want status waiting, got %q", resp.Status)
		}
		if resp.MaxAlertsPerIncident != 3 {
			t.Fatalf("want default max alerts 3, got %d", resp.MaxAlertsPerIncident)
		}
		if resp.KeywordMode != "contains" {
			t.Fatalf("want default keyword mode contains, got %q", resp.KeywordMode)
		}
	})

	t.Run("keyword with non-default mode/case-sensitivity on a paid plan's faster interval", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mustExec(t, pool, "UPDATE orgs SET plan = 'solo' WHERE id = $1", u.resp.OrgID)

		w := doAuthed(t, http.MethodPost, monitorH.CreateUptimeMonitor, u.access, createUptimeMonitorRequest{
			Name: "Keyword monitor", URL: "https://example.com", IntervalMins: 1, Keyword: "Welcome back", KeywordMode: "not_contains", KeywordCaseSensitive: true,
			MaxResponseTimeMs: 10000, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[uptimeMonitorResponse](t, w)
		if resp.Keyword == nil || *resp.Keyword != "Welcome back" {
			t.Fatalf("want keyword set, got %v", resp.Keyword)
		}
		if resp.KeywordMode != "not_contains" {
			t.Fatalf("want keyword mode not_contains, got %q", resp.KeywordMode)
		}
		if !resp.KeywordCaseSensitive {
			t.Fatal("want keyword case sensitive true")
		}
		if resp.IntervalMins != 1 {
			t.Fatalf("want interval 1 (Solo's minimum), got %d", resp.IntervalMins)
		}
	})

	t.Run("plan limit enforced at 10 monitors on Hobby", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		for i := 0; i < 10; i++ {
			createUptimeMonitor(t, monitorH, u.access, "Monitor")
		}
		w := doAuthed(t, http.MethodPost, monitorH.CreateUptimeMonitor, u.access, createUptimeMonitorRequest{
			Name: "One too many", URL: "https://example.com", IntervalMins: 10,
			MaxResponseTimeMs: 10000, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
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

func TestGetUptimeMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/uptime/x", nil)
		w := httptest.NewRecorder()
		monitorH.GetUptimeMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doUptimeMonitorRequest(t, http.MethodGet, monitorH.GetUptimeMonitor, u.access, "not-a-uuid", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doUptimeMonitorRequest(t, http.MethodGet, monitorH.GetUptimeMonitor, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant monitor is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doUptimeMonitorRequest(t, http.MethodGet, monitorH.GetUptimeMonitor, uB.access, mon.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 fetching org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success returns checks/incidents newest first and computed stats", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "Checked monitor")

		now := time.Now()
		var ok int32 = 200
		var bad int32 = 500
		seedUptimeCheck(t, pool, mon.ID, now.Add(-3*time.Hour), &ok, 120, true)
		seedUptimeCheck(t, pool, mon.ID, now.Add(-2*time.Hour), &bad, 80, false)
		seedUptimeCheck(t, pool, mon.ID, now.Add(-1*time.Hour), &ok, 95, true)
		seedUptimeIncident(t, pool, mon.ID, now.Add(-2*time.Hour))

		w := doUptimeMonitorRequest(t, http.MethodGet, monitorH.GetUptimeMonitor, u.access, mon.ID, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[struct {
			Monitor   uptimeMonitorResponse    `json:"monitor"`
			ChartData []uptimeCheckResponse    `json:"chartData"`
			Checks    []uptimeCheckResponse    `json:"checks"`
			Incidents []uptimeIncidentResponse `json:"incidents"`
			Stats     uptimeStatsResponse      `json:"stats"`
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

	t.Run("pagination: an out-of-range page returns no checks", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "Paged monitor")
		var ok int32 = 200
		seedUptimeCheck(t, pool, mon.ID, time.Now(), &ok, 50, true)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/uptime/"+mon.ID+"?page=2", nil)
		req = withURLParam(req, "id", mon.ID)
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(monitorH.GetUptimeMonitor)).ServeHTTP(w, req)
		resp := decodeBody[struct {
			Checks []uptimeCheckResponse `json:"checks"`
		}](t, w)
		if len(resp.Checks) != 0 {
			t.Fatalf("want 0 checks on page 2 (offset 50, only 1 row exists), got %d", len(resp.Checks))
		}
	})
}

func TestUpdateUptimeMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/monitors/uptime/x", nil)
		w := httptest.NewRecorder()
		monitorH.UpdateUptimeMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "x")
		cases := []struct {
			name string
			req  updateUptimeMonitorRequest
		}{
			{"missing name", updateUptimeMonitorRequest{URL: "https://example.com"}},
			{"invalid URL", updateUptimeMonitorRequest{Name: "x", URL: "not-a-url"}},
			{"keyword too long", updateUptimeMonitorRequest{Name: "x", URL: "https://example.com", Keyword: strings.Repeat("a", 501)}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doUptimeMonitorRequest(t, http.MethodPatch, monitorH.UpdateUptimeMonitor, u.access, mon.ID, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("keyword monitoring works on the Hobby (free) plan on update too", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "x")
		w := doUptimeMonitorRequest(t, http.MethodPatch, monitorH.UpdateUptimeMonitor, u.access, mon.ID, updateUptimeMonitorRequest{
			Name: "x", URL: "https://example.com", IntervalMins: 5, Keyword: "Welcome",
			MaxResponseTimeMs: 10000, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[uptimeMonitorResponse](t, w)
		if resp.Keyword == nil || *resp.Keyword != "Welcome" {
			t.Fatalf("want keyword set on Hobby, got %v", resp.Keyword)
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doUptimeMonitorRequest(t, http.MethodPatch, monitorH.UpdateUptimeMonitor, u.access, "00000000-0000-0000-0000-000000000000", updateUptimeMonitorRequest{
			Name: "x", URL: "https://example.com", IntervalMins: 10,
			MaxResponseTimeMs: 10000, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant update is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doUptimeMonitorRequest(t, http.MethodPatch, monitorH.UpdateUptimeMonitor, uB.access, mon.ID, updateUptimeMonitorRequest{
			Name: "hijacked", URL: "https://example.com", IntervalMins: 10,
			MaxResponseTimeMs: 10000, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 updating org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success applies changes and defaults", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "Original name")

		w := doUptimeMonitorRequest(t, http.MethodPatch, monitorH.UpdateUptimeMonitor, u.access, mon.ID, updateUptimeMonitorRequest{
			Name: "Renamed", URL: "https://new.example.com/health", AlertsEnabled: false, IntervalMins: 30, MaxAlertsPerIncident: -1,
			MaxResponseTimeMs: 10000, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[uptimeMonitorResponse](t, w)
		if resp.Name != "Renamed" {
			t.Fatalf("want name Renamed, got %q", resp.Name)
		}
		if resp.URL != "https://new.example.com/health" {
			t.Fatalf("want updated URL, got %q", resp.URL)
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

	t.Run("a 1-minute interval is honored on a paid plan that allows it", func(t *testing.T) {
		// UpdateUptimeMonitor uses the same plan-aware billing.ClampInterval
		// as CreateUptimeMonitor — a Solo-plan org (1-minute minimum per
		// ADR-019) setting a 1-minute interval via Update must get the 1
		// minute they're paying for, not a hardcoded floor of 10.
		u := signUpTestUser(t, authH, pool)
		mustExec(t, pool, "UPDATE orgs SET plan = 'solo' WHERE id = $1", u.resp.OrgID)
		mon := createUptimeMonitor(t, monitorH, u.access, "Solo plan monitor")

		w := doUptimeMonitorRequest(t, http.MethodPatch, monitorH.UpdateUptimeMonitor, u.access, mon.ID, updateUptimeMonitorRequest{
			Name: "x", URL: "https://example.com", IntervalMins: 1,
			MaxResponseTimeMs: 10000, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[uptimeMonitorResponse](t, w)
		if resp.IntervalMins != 1 {
			t.Fatalf("want interval 1, got %d", resp.IntervalMins)
		}
	})

	t.Run("an interval below the plan minimum is rejected on update too", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "Hobby plan monitor")

		w := doUptimeMonitorRequest(t, http.MethodPatch, monitorH.UpdateUptimeMonitor, u.access, mon.ID, updateUptimeMonitorRequest{
			Name: "x", URL: "https://example.com", IntervalMins: 1, // Hobby's minimum is 5
			MaxResponseTimeMs: 10000, HttpMethod: "GET", AcceptedStatusCodes: []int32{200},
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

func TestPauseResumeUptimeMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/monitors/uptime/x/pause", nil)
		w := httptest.NewRecorder()
		monitorH.PauseUptimeMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doUptimeMonitorRequest(t, http.MethodPost, monitorH.PauseUptimeMonitor, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant pause is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doUptimeMonitorRequest(t, http.MethodPost, monitorH.PauseUptimeMonitor, uB.access, mon.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 pausing org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("pause then resume round-trips status", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "Togglable")

		pauseW := doUptimeMonitorRequest(t, http.MethodPost, monitorH.PauseUptimeMonitor, u.access, mon.ID, nil)
		if pauseW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", pauseW.Code, pauseW.Body.String())
		}
		paused := decodeBody[uptimeMonitorResponse](t, pauseW)
		if paused.Status != "paused" {
			t.Fatalf("want status paused, got %q", paused.Status)
		}

		resumeW := doUptimeMonitorRequest(t, http.MethodPost, monitorH.ResumeUptimeMonitor, u.access, mon.ID, nil)
		if resumeW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", resumeW.Code, resumeW.Body.String())
		}
		resumed := decodeBody[uptimeMonitorResponse](t, resumeW)
		if resumed.Status != "waiting" {
			t.Fatalf("want status waiting after resume, got %q", resumed.Status)
		}
	})
}

func TestDeleteUptimeMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/monitors/uptime/x", nil)
		w := httptest.NewRecorder()
		monitorH.DeleteUptimeMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("cross-tenant delete is a no-op, owner delete succeeds and cascades", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, uA.access, "Org A monitor")
		var ok int32 = 200
		seedUptimeCheck(t, pool, mon.ID, time.Now(), &ok, 50, true)

		wrongDelete := doUptimeMonitorRequest(t, http.MethodDelete, monitorH.DeleteUptimeMonitor, uB.access, mon.ID, nil)
		if wrongDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204 (delete is a no-op exec regardless of match), got %d", wrongDelete.Code)
		}
		stillThere := doUptimeMonitorRequest(t, http.MethodGet, monitorH.GetUptimeMonitor, uA.access, mon.ID, nil)
		if stillThere.Code != http.StatusOK {
			t.Fatalf("want org A's monitor to survive org B's delete attempt, got %d", stillThere.Code)
		}

		ownerDelete := doUptimeMonitorRequest(t, http.MethodDelete, monitorH.DeleteUptimeMonitor, uA.access, mon.ID, nil)
		if ownerDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", ownerDelete.Code, ownerDelete.Body.String())
		}
		gone := doUptimeMonitorRequest(t, http.MethodGet, monitorH.GetUptimeMonitor, uA.access, mon.ID, nil)
		if gone.Code != http.StatusNotFound {
			t.Fatalf("want 404 after delete, got %d", gone.Code)
		}

		var checkCount int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM uptime_checks WHERE monitor_id = $1", mon.ID).Scan(&checkCount); err != nil {
			t.Fatalf("count checks: %v", err)
		}
		if checkCount != 0 {
			t.Fatalf("want checks cascade-deleted with the monitor, got %d remaining", checkCount)
		}
	})
}

func TestValidateURL(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"https://example.com", false},
		{"http://example.com/health", false},
		{"", true},
		{"   ", true},
		{"example.com", true},
		{"ftp://example.com", true},
		{"not a url", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			err := validateURL(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateURL(%q) error = %v, wantErr %v", tc.in, err, tc.wantErr)
			}
		})
	}
}

func TestValidateKeyword(t *testing.T) {
	if err := validateKeyword(""); err != nil {
		t.Fatalf("want empty keyword allowed, got %v", err)
	}
	if err := validateKeyword(strings.Repeat("a", 500)); err != nil {
		t.Fatalf("want 500-char keyword allowed, got %v", err)
	}
	if err := validateKeyword(strings.Repeat("a", 501)); err == nil {
		t.Fatal("want a 501-char keyword rejected")
	}
}

func TestParseKeywordMode(t *testing.T) {
	if got := parseKeywordMode("not_contains"); got != db.KeywordModeNotContains {
		t.Fatalf("want not_contains, got %q", got)
	}
	for _, in := range []string{"contains", "", "bogus"} {
		if got := parseKeywordMode(in); got != db.KeywordModeContains {
			t.Fatalf("parseKeywordMode(%q) = %q, want contains (default)", in, got)
		}
	}
}

func TestParsePageParam(t *testing.T) {
	cases := []struct {
		query string
		want  int32
	}{
		{"", 0},
		{"page=1", 0},
		{"page=2", 1},
		{"page=0", 0},
		{"page=-1", 0},
		{"page=abc", 0},
	}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/x?"+tc.query, nil)
			if got := parsePageParam(req); got != tc.want {
				t.Fatalf("parsePageParam(%q) = %d, want %d", tc.query, got, tc.want)
			}
		})
	}
}

func TestUptimePct(t *testing.T) {
	if got := uptimePct(3, 0); got != nil {
		t.Fatalf("want nil for zero total, got %v", *got)
	}
	got := uptimePct(3, 4)
	if got == nil || *got != 75.0 {
		t.Fatalf("want 75.0, got %v", got)
	}
}
