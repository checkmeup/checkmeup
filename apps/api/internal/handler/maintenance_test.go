package handler

// Integration tests for the maintenance-window handlers. Same conventions as
// auth_test.go/billing_test.go: real Postgres (ADR-010), package handler so
// the unexported request/response types (maintenanceWindowRequest,
// maintenanceWindowResponse, ...) can be reused directly instead of
// duplicating their JSON shape. These handlers read the "id" URL param via
// chi.URLParam, so requests that target a specific window go through
// withURLParam to inject a chi route context the same way chi's router
// would, since the handlers aren't exercised through a full chi.Mux here.

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

func testMaintenanceHandler(t *testing.T) (*AuthHandler, *MonitorHandler, *MaintenanceHandler, *pgxpool.Pool) {
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
	return NewAuthHandler(cfg, pool), NewMonitorHandler(cfg, pool, telegram.NewClient("")), NewMaintenanceHandler(pool), pool
}

type monitorRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func createCronMonitor(t *testing.T, h *MonitorHandler, access *http.Cookie, name string) monitorRef {
	t.Helper()
	w := doAuthed(t, http.MethodPost, h.CreateCronMonitor, access, map[string]any{
		"name": name, "schedule": "every 1h",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create cron monitor: want 201, got %d: %s", w.Code, w.Body.String())
	}
	return decodeBody[monitorRef](t, w)
}

func createUptimeMonitor(t *testing.T, h *MonitorHandler, access *http.Cookie, name string) monitorRef {
	t.Helper()
	w := doAuthed(t, http.MethodPost, h.CreateUptimeMonitor, access, map[string]any{
		"name": name, "url": "https://example.com/health", "intervalMins": 10,
		"maxResponseTimeMs": 10000, "httpMethod": "GET", "acceptedStatusCodes": []int32{200},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create uptime monitor: want 201, got %d: %s", w.Code, w.Body.String())
	}
	return decodeBody[monitorRef](t, w)
}

func createSSLMonitor(t *testing.T, h *MonitorHandler, access *http.Cookie, name string) monitorRef {
	t.Helper()
	w := doAuthed(t, http.MethodPost, h.CreateSSLMonitor, access, map[string]any{
		"name": name, "hostname": "example.com",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create ssl monitor: want 201, got %d: %s", w.Code, w.Body.String())
	}
	return decodeBody[monitorRef](t, w)
}

func createDomainMonitor(t *testing.T, h *MonitorHandler, access *http.Cookie, name string) monitorRef {
	t.Helper()
	w := doAuthed(t, http.MethodPost, h.CreateDomainMonitor, access, map[string]any{
		"name": name, "domain": "example.com",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create domain monitor: want 201, got %d: %s", w.Code, w.Body.String())
	}
	return decodeBody[monitorRef](t, w)
}

func createPortMonitor(t *testing.T, h *MonitorHandler, access *http.Cookie, name string) monitorRef {
	t.Helper()
	w := doAuthed(t, http.MethodPost, h.CreatePortMonitor, access, map[string]any{
		"name": name, "host": "example.com", "port": 443, "intervalMins": 10,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create port monitor: want 201, got %d: %s", w.Code, w.Body.String())
	}
	return decodeBody[monitorRef](t, w)
}

func createDNSMonitor(t *testing.T, h *MonitorHandler, access *http.Cookie, name string) monitorRef {
	t.Helper()
	w := doAuthed(t, http.MethodPost, h.CreateDNSMonitor, access, map[string]any{
		"name": name, "hostname": "example.com", "recordType": "A", "intervalMins": 10,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create dns monitor: want 201, got %d: %s", w.Code, w.Body.String())
	}
	return decodeBody[monitorRef](t, w)
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// doMaintenanceRequest authenticates via RequireAuth + the access cookie and
// injects the chi "id" URL param these handlers read via chi.URLParam.
func doMaintenanceRequest(t *testing.T, method string, handler http.HandlerFunc, access *http.Cookie, id string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader = http.NoBody
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "/api/v1/maintenance-windows/"+id, r)
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", id)
	req.AddCookie(access)
	w := httptest.NewRecorder()
	apimiddleware.RequireAuth(testJWTSecret)(handler).ServeHTTP(w, req)
	return w
}

func TestListMaintenanceWindows(t *testing.T) {
	authH, monitorH, maintH, pool := testMaintenanceHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/maintenance-windows", nil)
		w := httptest.NewRecorder()
		maintH.ListMaintenanceWindows(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("empty for a fresh org", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodGet, maintH.ListMaintenanceWindows, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		list := decodeBody[[]maintenanceWindowResponse](t, w)
		if len(list) != 0 {
			t.Fatalf("want empty list, got %d", len(list))
		}
	})

	t.Run("lists a created window with monitor count and status", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "DB backup")

		starts := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
		createW := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, u.access, maintenanceWindowRequest{
			Title:    "Ongoing migration",
			StartsAt: starts,
			Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: mon.ID}},
		})
		if createW.Code != http.StatusCreated {
			t.Fatalf("create: want 201, got %d: %s", createW.Code, createW.Body.String())
		}

		w := doAuthed(t, http.MethodGet, maintH.ListMaintenanceWindows, u.access, nil)
		list := decodeBody[[]maintenanceWindowResponse](t, w)
		if len(list) != 1 {
			t.Fatalf("want 1 window, got %d", len(list))
		}
		if list[0].Title != "Ongoing migration" {
			t.Fatalf("want title %q, got %q", "Ongoing migration", list[0].Title)
		}
		if list[0].MonitorCount != 1 {
			t.Fatalf("want monitor count 1, got %d", list[0].MonitorCount)
		}
		if list[0].Status != "active" {
			t.Fatalf("want status active (started in the past, no end), got %q", list[0].Status)
		}
	})
}

func TestCreateMaintenanceWindow(t *testing.T) {
	authH, monitorH, maintH, pool := testMaintenanceHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, maintH.CreateMaintenanceWindow, http.MethodPost, "/api/v1/maintenance-windows", maintenanceWindowRequest{Title: "x"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Validation monitor")
		validStart := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		validMonitors := []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: mon.ID}}
		badEnd := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339) // before start

		cases := []struct {
			name string
			req  maintenanceWindowRequest
		}{
			{"missing title", maintenanceWindowRequest{StartsAt: validStart, Monitors: validMonitors}},
			{"no monitors", maintenanceWindowRequest{Title: "x", StartsAt: validStart}},
			{"invalid startsAt", maintenanceWindowRequest{Title: "x", StartsAt: "not-a-timestamp", Monitors: validMonitors}},
			{"invalid endsAt", maintenanceWindowRequest{Title: "x", StartsAt: validStart, EndsAt: strPtr("not-a-timestamp"), Monitors: validMonitors}},
			{"endsAt before startsAt", maintenanceWindowRequest{Title: "x", StartsAt: validStart, EndsAt: &badEnd, Monitors: validMonitors}},
			{"unknown monitor type", maintenanceWindowRequest{Title: "x", StartsAt: validStart, Monitors: []maintenanceMonitorInput{{MonitorType: "carrier-pigeon", MonitorID: mon.ID}}}},
			{"monitor not found", maintenanceWindowRequest{Title: "x", StartsAt: validStart, Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: "00000000-0000-0000-0000-000000000000"}}}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, u.access, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}

		t.Run("malformed JSON body", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/maintenance-windows", bytes.NewReader([]byte("not json")))
			req.AddCookie(u.access)
			w := httptest.NewRecorder()
			apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(maintH.CreateMaintenanceWindow)).ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
			}
		})
	})

	t.Run("a monitor belonging to another org is rejected (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		monB := createCronMonitor(t, monitorH, uB.access, "Org B's monitor")

		starts := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		w := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, uA.access, maintenanceWindowRequest{
			Title:    "Cross-tenant attempt",
			StartsAt: starts,
			Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: monB.ID}},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400 (monitor not visible to org A), got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success resolves monitor names, dedups, and computes status", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cron := createCronMonitor(t, monitorH, u.access, "API cron")
		uptime := createUptimeMonitor(t, monitorH, u.access, "API uptime")
		ssl := createSSLMonitor(t, monitorH, u.access, "API ssl")

		starts := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		w := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, u.access, maintenanceWindowRequest{
			Title:    "Multi-monitor window",
			Message:  "scheduled work",
			StartsAt: starts,
			Monitors: []maintenanceMonitorInput{
				{MonitorType: "cron", MonitorID: cron.ID},
				{MonitorType: "uptime", MonitorID: uptime.ID},
				{MonitorType: "ssl", MonitorID: ssl.ID},
				{MonitorType: "cron", MonitorID: cron.ID}, // duplicate, should be deduped
			},
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[maintenanceWindowResponse](t, w)
		if resp.MonitorCount != 3 {
			t.Fatalf("want 3 monitors after dedup, got %d", resp.MonitorCount)
		}
		if resp.Status != "upcoming" {
			t.Fatalf("want status upcoming (future start), got %q", resp.Status)
		}
		names := map[string]bool{}
		for _, m := range resp.Monitors {
			names[m.Name] = true
		}
		for _, want := range []string{"API cron", "API uptime", "API ssl"} {
			if !names[want] {
				t.Fatalf("want monitor %q in response, got %+v", want, resp.Monitors)
			}
		}
	})

	t.Run("rejects creating a 101st window", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "x")
		orgID := u.resp.OrgID
		for i := 0; i < 100; i++ {
			mustExec(t, pool, "INSERT INTO maintenance_windows (org_id, title, starts_at) VALUES ($1, 'seed', NOW())", orgID)
		}

		w := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, u.access, maintenanceWindowRequest{
			Title: "101st", StartsAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
			Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: mon.ID}},
		})
		if w.Code != http.StatusConflict {
			t.Fatalf("want 409, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "too_many_maintenance_windows" {
			t.Fatalf("want code too_many_maintenance_windows, got %q", body["code"])
		}
	})
}

func TestGetMaintenanceWindow(t *testing.T) {
	authH, monitorH, maintH, pool := testMaintenanceHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/maintenance-windows/x", nil)
		w := httptest.NewRecorder()
		maintH.GetMaintenanceWindow(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doMaintenanceRequest(t, http.MethodGet, maintH.GetMaintenanceWindow, u.access, "not-a-uuid", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doMaintenanceRequest(t, http.MethodGet, maintH.GetMaintenanceWindow, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant window is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, uA.access, "Org A monitor")
		starts := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		createW := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, uA.access, maintenanceWindowRequest{
			Title: "Org A window", StartsAt: starts,
			Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: mon.ID}},
		})
		created := decodeBody[maintenanceWindowResponse](t, createW)

		w := doMaintenanceRequest(t, http.MethodGet, maintH.GetMaintenanceWindow, uB.access, created.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 fetching org A's window as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success, and a deleted monitor shows a placeholder name", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cron := createCronMonitor(t, monitorH, u.access, "Will be deleted")
		ssl := createSSLMonitor(t, monitorH, u.access, "Stays")

		starts := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
		createW := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, u.access, maintenanceWindowRequest{
			Title: "Get test", StartsAt: starts,
			Monitors: []maintenanceMonitorInput{
				{MonitorType: "cron", MonitorID: cron.ID},
				{MonitorType: "ssl", MonitorID: ssl.ID},
			},
		})
		created := decodeBody[maintenanceWindowResponse](t, createW)

		if _, err := pool.Exec(context.Background(), "DELETE FROM cron_monitors WHERE id = $1", cron.ID); err != nil {
			t.Fatalf("delete cron monitor: %v", err)
		}

		w := doMaintenanceRequest(t, http.MethodGet, maintH.GetMaintenanceWindow, u.access, created.ID, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[maintenanceWindowResponse](t, w)
		if resp.MonitorCount != 2 {
			t.Fatalf("want monitor count 2 (link survives monitor deletion), got %d", resp.MonitorCount)
		}
		var sawDeletedPlaceholder, sawSurvivor bool
		for _, m := range resp.Monitors {
			if m.Name == "(deleted monitor)" {
				sawDeletedPlaceholder = true
			}
			if m.Name == "Stays" {
				sawSurvivor = true
			}
		}
		if !sawDeletedPlaceholder {
			t.Fatalf("want a (deleted monitor) placeholder, got %+v", resp.Monitors)
		}
		if !sawSurvivor {
			t.Fatalf("want the surviving monitor's real name, got %+v", resp.Monitors)
		}
	})
}

func TestUpdateMaintenanceWindow(t *testing.T) {
	authH, monitorH, maintH, pool := testMaintenanceHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/maintenance-windows/x", nil)
		w := httptest.NewRecorder()
		maintH.UpdateMaintenanceWindow(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "x")
		starts := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		w := doMaintenanceRequest(t, http.MethodPatch, maintH.UpdateMaintenanceWindow, u.access, "00000000-0000-0000-0000-000000000000", maintenanceWindowRequest{
			Title: "x", StartsAt: starts, Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: mon.ID}},
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant update is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		monA := createCronMonitor(t, monitorH, uA.access, "Org A monitor")
		starts := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		createW := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, uA.access, maintenanceWindowRequest{
			Title: "Org A window", StartsAt: starts,
			Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: monA.ID}},
		})
		created := decodeBody[maintenanceWindowResponse](t, createW)

		w := doMaintenanceRequest(t, http.MethodPatch, maintH.UpdateMaintenanceWindow, uB.access, created.ID, maintenanceWindowRequest{
			Title: "hijacked", StartsAt: starts, Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: monA.ID}},
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 updating org A's window as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("validation failure on an existing window", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "x")
		starts := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		createW := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, u.access, maintenanceWindowRequest{
			Title: "Original", StartsAt: starts, Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: mon.ID}},
		})
		created := decodeBody[maintenanceWindowResponse](t, createW)

		w := doMaintenanceRequest(t, http.MethodPatch, maintH.UpdateMaintenanceWindow, u.access, created.ID, maintenanceWindowRequest{
			StartsAt: starts, Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: mon.ID}},
		}) // missing title
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success replaces title and monitor list", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cron := createCronMonitor(t, monitorH, u.access, "Cron")
		uptime := createUptimeMonitor(t, monitorH, u.access, "Uptime")
		starts := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		createW := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, u.access, maintenanceWindowRequest{
			Title: "Original", StartsAt: starts, Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: cron.ID}},
		})
		created := decodeBody[maintenanceWindowResponse](t, createW)

		w := doMaintenanceRequest(t, http.MethodPatch, maintH.UpdateMaintenanceWindow, u.access, created.ID, maintenanceWindowRequest{
			Title: "Renamed", StartsAt: starts, Monitors: []maintenanceMonitorInput{{MonitorType: "uptime", MonitorID: uptime.ID}},
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		updated := decodeBody[maintenanceWindowResponse](t, w)
		if updated.Title != "Renamed" {
			t.Fatalf("want title Renamed, got %q", updated.Title)
		}
		if updated.MonitorCount != 1 || len(updated.Monitors) != 1 || updated.Monitors[0].Name != "Uptime" {
			t.Fatalf("want monitor list replaced with just Uptime, got %+v", updated.Monitors)
		}
	})
}

func TestDeleteMaintenanceWindow(t *testing.T) {
	authH, monitorH, maintH, pool := testMaintenanceHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/maintenance-windows/x", nil)
		w := httptest.NewRecorder()
		maintH.DeleteMaintenanceWindow(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("cross-tenant delete is a no-op, owner delete succeeds", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, uA.access, "Org A monitor")
		starts := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		createW := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, uA.access, maintenanceWindowRequest{
			Title: "Org A window", StartsAt: starts,
			Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: mon.ID}},
		})
		created := decodeBody[maintenanceWindowResponse](t, createW)

		// Org B "deleting" org A's window must not affect it.
		wrongDelete := doMaintenanceRequest(t, http.MethodDelete, maintH.DeleteMaintenanceWindow, uB.access, created.ID, nil)
		if wrongDelete.Code != http.StatusNoContent {
			t.Fatalf("DELETE always reports success regardless of match (it's a no-op exec); want 204, got %d", wrongDelete.Code)
		}
		stillThere := doMaintenanceRequest(t, http.MethodGet, maintH.GetMaintenanceWindow, uA.access, created.ID, nil)
		if stillThere.Code != http.StatusOK {
			t.Fatalf("want org A's window to survive org B's delete attempt, got %d", stillThere.Code)
		}

		ownerDelete := doMaintenanceRequest(t, http.MethodDelete, maintH.DeleteMaintenanceWindow, uA.access, created.ID, nil)
		if ownerDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", ownerDelete.Code, ownerDelete.Body.String())
		}
		gone := doMaintenanceRequest(t, http.MethodGet, maintH.GetMaintenanceWindow, uA.access, created.ID, nil)
		if gone.Code != http.StatusNotFound {
			t.Fatalf("want 404 after delete, got %d", gone.Code)
		}
	})
}

func TestEndMaintenanceWindowNow(t *testing.T) {
	authH, monitorH, maintH, pool := testMaintenanceHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/maintenance-windows/x/end", nil)
		w := httptest.NewRecorder()
		maintH.EndMaintenanceWindowNow(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doMaintenanceRequest(t, http.MethodPost, maintH.EndMaintenanceWindowNow, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("ends an open-ended window, and ending it again 404s", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Open-ended")
		starts := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
		createW := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, u.access, maintenanceWindowRequest{
			Title: "Open-ended window", StartsAt: starts,
			Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: mon.ID}},
		})
		created := decodeBody[maintenanceWindowResponse](t, createW)
		if created.EndsAt != nil {
			t.Fatalf("want no end date initially, got %v", created.EndsAt)
		}

		endW := doMaintenanceRequest(t, http.MethodPost, maintH.EndMaintenanceWindowNow, u.access, created.ID, nil)
		if endW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", endW.Code, endW.Body.String())
		}
		ended := decodeBody[maintenanceWindowResponse](t, endW)
		if ended.EndsAt == nil {
			t.Fatal("want endsAt set after ending now")
		}
		if ended.Status != "ended" {
			t.Fatalf("want status ended, got %q", ended.Status)
		}
		if ended.MonitorCount != 1 {
			t.Fatalf("want monitor count 1 preserved, got %d", ended.MonitorCount)
		}

		again := doMaintenanceRequest(t, http.MethodPost, maintH.EndMaintenanceWindowNow, u.access, created.ID, nil)
		if again.Code != http.StatusNotFound {
			t.Fatalf("want 404 ending an already-ended window, got %d: %s", again.Code, again.Body.String())
		}
	})
}

func strPtr(s string) *string { return &s }
