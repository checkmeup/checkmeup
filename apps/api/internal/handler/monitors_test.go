package handler

// Integration tests for the cron monitor handlers in monitors.go. Same
// conventions as auth_test.go/billing_test.go/maintenance_test.go: real
// Postgres (ADR-010), package handler so unexported request/response types
// and helpers (createCronMonitor et al. from maintenance_test.go,
// withURLParam) can be reused directly.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

func testMonitorHandler(t *testing.T) (*AuthHandler, *MonitorHandler, *pgxpool.Pool) {
	t.Helper()
	pool := testPool(t)
	cfg := &config.Config{
		Env:           "development",
		JWTSecret:     testJWTSecret,
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 7 * 24 * time.Hour,
		AppURL:        "http://localhost:5173",
		BaseURL:       "http://localhost:8080",
	}
	return NewAuthHandler(cfg, pool), NewMonitorHandler(cfg, pool, telegram.NewClient("")), pool
}

// doMonitorRequest authenticates via RequireAuth + the access cookie and
// injects the chi "id" URL param these handlers read via chi.URLParam.
func doMonitorRequest(t *testing.T, method string, handler http.HandlerFunc, access *http.Cookie, id string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader = http.NoBody
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "/api/v1/monitors/cron/"+id, r)
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", id)
	req.AddCookie(access)
	w := httptest.NewRecorder()
	apimiddleware.RequireAuth(testJWTSecret)(handler).ServeHTTP(w, req)
	return w
}

func TestListCronMonitors(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/cron", nil)
		w := httptest.NewRecorder()
		monitorH.ListCronMonitors(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("empty for a fresh org", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodGet, monitorH.ListCronMonitors, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		list := decodeBody[[]cronMonitorResponse](t, w)
		if len(list) != 0 {
			t.Fatalf("want empty list, got %d", len(list))
		}
	})

	t.Run("lists created monitors", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		createCronMonitor(t, monitorH, u.access, "Backup job")
		createCronMonitor(t, monitorH, u.access, "Nightly export")

		w := doAuthed(t, http.MethodGet, monitorH.ListCronMonitors, u.access, nil)
		list := decodeBody[[]cronMonitorResponse](t, w)
		if len(list) != 2 {
			t.Fatalf("want 2 monitors, got %d", len(list))
		}
	})
}

func TestCreateCronMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, monitorH.CreateCronMonitor, http.MethodPost, "/api/v1/monitors/cron", createCronMonitorRequest{Name: "x", Schedule: "every 1h"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/monitors/cron", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(monitorH.CreateCronMonitor)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cases := []struct {
			name string
			req  createCronMonitorRequest
		}{
			{"missing name", createCronMonitorRequest{Schedule: "every 1h"}},
			{"missing schedule", createCronMonitorRequest{Name: "x"}},
			{"invalid schedule", createCronMonitorRequest{Name: "x", Schedule: "not a valid schedule at all"}},
			{"zero maxDurationMins", createCronMonitorRequest{Name: "x", Schedule: "every 1h", MaxDurationMins: int32Ptr(0)}},
			{"negative maxDurationMins", createCronMonitorRequest{Name: "x", Schedule: "every 1h", MaxDurationMins: int32Ptr(-5)}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doAuthed(t, http.MethodPost, monitorH.CreateCronMonitor, u.access, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("success with an interval schedule, applying defaults", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateCronMonitor, u.access, createCronMonitorRequest{
			Name: "Interval job", Schedule: "every 1h", GracePeriodMins: 0, MaxAlertsPerIncident: -1,
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[cronMonitorResponse](t, w)
		if resp.GracePeriodMins != 5 {
			t.Fatalf("want default grace period 5, got %d", resp.GracePeriodMins)
		}
		if resp.MaxAlertsPerIncident != 3 {
			t.Fatalf("want default max alerts 3, got %d", resp.MaxAlertsPerIncident)
		}
		if resp.Status != "waiting" {
			t.Fatalf("want status waiting on creation, got %q", resp.Status)
		}
		if len(resp.PingToken) == 0 {
			t.Fatal("want a non-empty ping token")
		}
		wantURL := "http://localhost:8080/ping/" + resp.PingToken
		if resp.PingURL != wantURL {
			t.Fatalf("want ping URL %q, got %q", wantURL, resp.PingURL)
		}
		if resp.MaxDurationMins != nil {
			t.Fatalf("want maxDurationMins unset (zombie detection inactive) by default, got %v", resp.MaxDurationMins)
		}
	})

	t.Run("success with maxDurationMins set (US-3402)", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateCronMonitor, u.access, createCronMonitorRequest{
			Name: "Zombie-checked job", Schedule: "every 1h", MaxDurationMins: int32Ptr(30),
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[cronMonitorResponse](t, w)
		if resp.MaxDurationMins == nil || *resp.MaxDurationMins != 30 {
			t.Fatalf("want maxDurationMins 30, got %v", resp.MaxDurationMins)
		}
	})

	t.Run("success with a 5-field cron expression", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateCronMonitor, u.access, createCronMonitorRequest{
			Name: "Cron expr job", Schedule: "0 9 * * *",
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("plan limit enforced at 10 monitors on Hobby", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		for i := 0; i < 10; i++ {
			createCronMonitor(t, monitorH, u.access, "Monitor")
		}
		w := doAuthed(t, http.MethodPost, monitorH.CreateCronMonitor, u.access, createCronMonitorRequest{
			Name: "One too many", Schedule: "every 1h",
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

func TestGetCronMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/cron/x", nil)
		w := httptest.NewRecorder()
		monitorH.GetCronMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doMonitorRequest(t, http.MethodGet, monitorH.GetCronMonitor, u.access, "not-a-uuid", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doMonitorRequest(t, http.MethodGet, monitorH.GetCronMonitor, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant monitor is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doMonitorRequest(t, http.MethodGet, monitorH.GetCronMonitor, uB.access, mon.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 fetching org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success returns pings and incidents newest first", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Monitored job")

		now := time.Now()
		seedCronPing(t, pool, mon.ID, now.Add(-2*time.Hour), "10.0.0.1")
		seedCronPing(t, pool, mon.ID, now.Add(-1*time.Hour), "10.0.0.2")
		seedCronIncident(t, pool, mon.ID, now.Add(-3*time.Hour), nil)

		w := doMonitorRequest(t, http.MethodGet, monitorH.GetCronMonitor, u.access, mon.ID, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[struct {
			Monitor   cronMonitorResponse    `json:"monitor"`
			Pings     []cronPingResponse     `json:"pings"`
			Incidents []cronIncidentResponse `json:"incidents"`
		}](t, w)
		if resp.Monitor.ID != mon.ID {
			t.Fatalf("want monitor id %q, got %q", mon.ID, resp.Monitor.ID)
		}
		if len(resp.Pings) != 2 {
			t.Fatalf("want 2 pings, got %d", len(resp.Pings))
		}
		if resp.Pings[0].SourceIP != "10.0.0.2" {
			t.Fatalf("want most recent ping first, got source IP %q", resp.Pings[0].SourceIP)
		}
		if len(resp.Incidents) != 1 {
			t.Fatalf("want 1 incident, got %d", len(resp.Incidents))
		}
		if resp.Incidents[0].ResolvedAt != nil {
			t.Fatal("want an unresolved incident to have a nil resolvedAt")
		}
	})
}

func TestGetCronPings(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/cron/x/pings", nil)
		w := httptest.NewRecorder()
		monitorH.GetCronPings(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("not found for another org's monitor", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doMonitorRequest(t, http.MethodGet, monitorH.GetCronPings, uB.access, mon.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("pagination", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Paged job")
		now := time.Now()
		seedCronPing(t, pool, mon.ID, now.Add(-3*time.Minute), "1.1.1.1")
		seedCronPing(t, pool, mon.ID, now.Add(-2*time.Minute), "1.1.1.2")
		seedCronPing(t, pool, mon.ID, now.Add(-1*time.Minute), "1.1.1.3")

		w := doMonitorRequest(t, http.MethodGet, monitorH.GetCronPings, u.access, mon.ID, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		page1 := decodeBody[[]cronPingResponse](t, w)
		if len(page1) != 3 {
			t.Fatalf("want 3 pings on page 1, got %d", len(page1))
		}
		if page1[0].SourceIP != "1.1.1.3" {
			t.Fatalf("want newest ping first, got %q", page1[0].SourceIP)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/cron/"+mon.ID+"/pings?page=2", nil)
		req = withURLParam(req, "id", mon.ID)
		req.AddCookie(u.access)
		w2 := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(monitorH.GetCronPings)).ServeHTTP(w2, req)
		page2 := decodeBody[[]cronPingResponse](t, w2)
		if len(page2) != 0 {
			t.Fatalf("want 0 pings on page 2 (offset 50, only 3 rows exist), got %d", len(page2))
		}
	})

	t.Run("includes a ping's query-string metadata", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		wCreate := doAuthed(t, http.MethodPost, monitorH.CreateCronMonitor, u.access, map[string]any{
			"name": "CI job", "schedule": "every 1h",
		})
		if wCreate.Code != http.StatusCreated {
			t.Fatalf("create cron monitor: want 201, got %d: %s", wCreate.Code, wCreate.Body.String())
		}
		mon := decodeBody[cronMonitorRef](t, wCreate)

		pingReq := httptest.NewRequest(http.MethodGet, "/ping/"+mon.PingToken+"?build=142&state=success", nil)
		ping := NewPingHandler(pool, nil, nil, nil, nil, nil)
		pingW := httptest.NewRecorder()
		ping.ReceivePing(pingW, withURLParam(pingReq, "token", mon.PingToken))
		if pingW.Code != http.StatusOK {
			t.Fatalf("ping: want 200, got %d", pingW.Code)
		}

		w := doMonitorRequest(t, http.MethodGet, monitorH.GetCronPings, u.access, mon.ID, nil)
		pings := decodeBody[[]cronPingResponse](t, w)
		if len(pings) != 1 {
			t.Fatalf("want 1 ping, got %d", len(pings))
		}
		if pings[0].Metadata["build"] != "142" || pings[0].Metadata["state"] != "success" {
			t.Fatalf("metadata = %+v, want build=142 state=success", pings[0].Metadata)
		}
	})
}

func TestUpdateCronMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/monitors/cron/x", nil)
		w := httptest.NewRecorder()
		monitorH.UpdateCronMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "x")
		cases := []struct {
			name string
			req  updateCronMonitorRequest
		}{
			{"missing name", updateCronMonitorRequest{Schedule: "every 1h"}},
			{"invalid schedule", updateCronMonitorRequest{Name: "x", Schedule: "garbage"}},
			{"zero maxDurationMins", updateCronMonitorRequest{Name: "x", Schedule: "every 1h", MaxDurationMins: int32Ptr(0)}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doMonitorRequest(t, http.MethodPatch, monitorH.UpdateCronMonitor, u.access, mon.ID, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doMonitorRequest(t, http.MethodPatch, monitorH.UpdateCronMonitor, u.access, "00000000-0000-0000-0000-000000000000", updateCronMonitorRequest{
			Name: "x", Schedule: "every 1h",
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant update is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doMonitorRequest(t, http.MethodPatch, monitorH.UpdateCronMonitor, uB.access, mon.ID, updateCronMonitorRequest{
			Name: "hijacked", Schedule: "every 1h",
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 updating org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success applies changes and defaults", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Original name")

		w := doMonitorRequest(t, http.MethodPatch, monitorH.UpdateCronMonitor, u.access, mon.ID, updateCronMonitorRequest{
			Name: "Renamed", Schedule: "0 9 * * *", AlertsEnabled: false, GracePeriodMins: 0, MaxAlertsPerIncident: -1,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[cronMonitorResponse](t, w)
		if resp.Name != "Renamed" {
			t.Fatalf("want name Renamed, got %q", resp.Name)
		}
		if resp.Schedule != "0 9 * * *" {
			t.Fatalf("want schedule updated, got %q", resp.Schedule)
		}
		if resp.AlertsEnabled {
			t.Fatal("want alertsEnabled false")
		}
		if resp.GracePeriodMins != 5 {
			t.Fatalf("want default grace period 5, got %d", resp.GracePeriodMins)
		}
		if resp.MaxAlertsPerIncident != 3 {
			t.Fatalf("want default max alerts 3, got %d", resp.MaxAlertsPerIncident)
		}
		if resp.MaxDurationMins != nil {
			t.Fatalf("want maxDurationMins to stay unset, got %v", resp.MaxDurationMins)
		}
	})

	t.Run("success sets maxDurationMins (US-3402)", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Original name")

		w := doMonitorRequest(t, http.MethodPatch, monitorH.UpdateCronMonitor, u.access, mon.ID, updateCronMonitorRequest{
			Name: "Renamed", Schedule: "every 1h", MaxDurationMins: int32Ptr(45),
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[cronMonitorResponse](t, w)
		if resp.MaxDurationMins == nil || *resp.MaxDurationMins != 45 {
			t.Fatalf("want maxDurationMins 45, got %v", resp.MaxDurationMins)
		}
	})
}

func int32Ptr(v int32) *int32 { return &v }

func TestPauseResumeCronMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/monitors/cron/x/pause", nil)
		w := httptest.NewRecorder()
		monitorH.PauseCronMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doMonitorRequest(t, http.MethodPost, monitorH.PauseCronMonitor, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant pause is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doMonitorRequest(t, http.MethodPost, monitorH.PauseCronMonitor, uB.access, mon.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 pausing org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("pause then resume round-trips status", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Togglable")

		pauseW := doMonitorRequest(t, http.MethodPost, monitorH.PauseCronMonitor, u.access, mon.ID, nil)
		if pauseW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", pauseW.Code, pauseW.Body.String())
		}
		paused := decodeBody[cronMonitorResponse](t, pauseW)
		if paused.Status != "paused" {
			t.Fatalf("want status paused, got %q", paused.Status)
		}

		resumeW := doMonitorRequest(t, http.MethodPost, monitorH.ResumeCronMonitor, u.access, mon.ID, nil)
		if resumeW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", resumeW.Code, resumeW.Body.String())
		}
		resumed := decodeBody[cronMonitorResponse](t, resumeW)
		if resumed.Status != "waiting" {
			t.Fatalf("want status waiting after resume, got %q", resumed.Status)
		}
	})

	t.Run("resume is blocked once active monitors are already at the plan limit (ADR-019)", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)

		var monitors []monitorRef
		for i := range 10 { // Hobby's monitor limit
			monitors = append(monitors, createCronMonitor(t, monitorH, u.access, fmt.Sprintf("mon-%d", i)))
		}

		pauseW := doMonitorRequest(t, http.MethodPost, monitorH.PauseCronMonitor, u.access, monitors[0].ID, nil)
		if pauseW.Code != http.StatusOK {
			t.Fatalf("pause setup: want 200, got %d: %s", pauseW.Code, pauseW.Body.String())
		}

		// Directly create an 11th, already-active monitor — simulating an
		// org that's over its limit (e.g. from a prior higher plan) rather
		// than going through the full downgrade webhook flow (covered
		// separately in billing_test.go) — so active count is back at 10
		// (the limit) even with one paused.
		if _, err := monitorH.queries.CreateCronMonitor(context.Background(), db.CreateCronMonitorParams{
			OrgID: orgID, Name: "extra", Schedule: "every 1h", GracePeriodMins: 5,
			PingToken: uuid.NewString(), MaxAlertsPerIncident: 3,
		}); err != nil {
			t.Fatalf("create extra monitor: %v", err)
		}

		resumeW := doMonitorRequest(t, http.MethodPost, monitorH.ResumeCronMonitor, u.access, monitors[0].ID, nil)
		if resumeW.Code != http.StatusPaymentRequired {
			t.Fatalf("want 402, got %d: %s", resumeW.Code, resumeW.Body.String())
		}
		body := decodeBody[map[string]string](t, resumeW)
		if body["code"] != "plan_limit_reached" {
			t.Fatalf("want code plan_limit_reached, got %q", body["code"])
		}

		// The monitor stays paused — the blocked resume must not have applied.
		still, err := monitorH.queries.GetCronMonitor(context.Background(), db.GetCronMonitorParams{ID: uuid.MustParse(monitors[0].ID), OrgID: orgID})
		if err != nil {
			t.Fatalf("get monitor: %v", err)
		}
		if still.Status != db.MonitorStatusPaused {
			t.Fatalf("want monitor to remain paused after the blocked resume, got %q", still.Status)
		}
	})
}

func TestDeleteCronMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/monitors/cron/x", nil)
		w := httptest.NewRecorder()
		monitorH.DeleteCronMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("cross-tenant delete is a no-op, owner delete succeeds and cascades", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, uA.access, "Org A monitor")
		seedCronPing(t, pool, mon.ID, time.Now(), "1.2.3.4")

		wrongDelete := doMonitorRequest(t, http.MethodDelete, monitorH.DeleteCronMonitor, uB.access, mon.ID, nil)
		if wrongDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204 (delete is a no-op exec regardless of match), got %d", wrongDelete.Code)
		}
		stillThere := doMonitorRequest(t, http.MethodGet, monitorH.GetCronMonitor, uA.access, mon.ID, nil)
		if stillThere.Code != http.StatusOK {
			t.Fatalf("want org A's monitor to survive org B's delete attempt, got %d", stillThere.Code)
		}

		ownerDelete := doMonitorRequest(t, http.MethodDelete, monitorH.DeleteCronMonitor, uA.access, mon.ID, nil)
		if ownerDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", ownerDelete.Code, ownerDelete.Body.String())
		}
		gone := doMonitorRequest(t, http.MethodGet, monitorH.GetCronMonitor, uA.access, mon.ID, nil)
		if gone.Code != http.StatusNotFound {
			t.Fatalf("want 404 after delete, got %d", gone.Code)
		}

		var pingCount int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM cron_pings WHERE monitor_id = $1", mon.ID).Scan(&pingCount); err != nil {
			t.Fatalf("count pings: %v", err)
		}
		if pingCount != 0 {
			t.Fatalf("want pings cascade-deleted with the monitor, got %d remaining", pingCount)
		}
	})
}

func TestValidateSchedule(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"empty", "", true},
		{"interval minutes", "every 1m", false},
		{"interval hours", "every 12h", false},
		{"five-field cron", "0 9 * * *", false},
		{"wrong field count", "0 9 * *", true},
		{"gibberish", "whenever I feel like it doing stuff", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSchedule(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateSchedule(%q) error = %v, wantErr %v", tc.in, err, tc.wantErr)
			}
		})
	}
}

func seedCronPing(t *testing.T, pool *pgxpool.Pool, monitorID string, receivedAt time.Time, sourceIP string) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		"INSERT INTO cron_pings (monitor_id, received_at, source_ip) VALUES ($1, $2, $3)",
		monitorID, receivedAt, sourceIP,
	); err != nil {
		t.Fatalf("seed cron ping: %v", err)
	}
}

func seedCronIncident(t *testing.T, pool *pgxpool.Pool, monitorID string, startedAt time.Time, resolvedAt *time.Time) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		"INSERT INTO cron_incidents (monitor_id, started_at, resolved_at) VALUES ($1, $2, $3)",
		monitorID, startedAt, resolvedAt,
	); err != nil {
		t.Fatalf("seed cron incident: %v", err)
	}
}
