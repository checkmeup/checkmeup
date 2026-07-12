package handler

// Integration tests for the manual incident handlers in incidents.go
// (EP-24). Same conventions as maintenance_test.go: real Postgres
// (ADR-010), package handler so the unexported request/response types can
// be reused directly. testIncidentHandler additionally returns the
// maintenance handler so US-2405 (maintenance-overlap warning) can be
// exercised end to end.

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

func testIncidentHandler(t *testing.T) (*AuthHandler, *MonitorHandler, *MaintenanceHandler, *IncidentHandler, *pgxpool.Pool) {
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
	return NewAuthHandler(cfg, pool), NewMonitorHandler(cfg, pool, telegram.NewClient("")), NewMaintenanceHandler(pool), NewIncidentHandler(pool), pool
}

// doIncidentRequest authenticates via RequireAuth + the access cookie and
// injects the chi URL params these handlers read via chi.URLParam.
func doIncidentRequest(t *testing.T, method string, handler http.HandlerFunc, access *http.Cookie, path string, params map[string]string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader = http.NoBody
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req.AddCookie(access)
	w := httptest.NewRecorder()
	apimiddleware.RequireAuth(testJWTSecret)(handler).ServeHTTP(w, req)
	return w
}

func createIncident(t *testing.T, h *IncidentHandler, access *http.Cookie, req createIncidentRequest) incidentResponse {
	t.Helper()
	w := doAuthed(t, http.MethodPost, h.CreateIncident, access, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create incident: want 201, got %d: %s", w.Code, w.Body.String())
	}
	return decodeBody[incidentResponse](t, w)
}

func TestListIncidents(t *testing.T) {
	authH, monitorH, _, incidentH, pool := testIncidentHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
		w := httptest.NewRecorder()
		incidentH.ListIncidents(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("empty for a fresh org", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodGet, incidentH.ListIncidents, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		list := decodeBody[[]incidentResponse](t, w)
		if len(list) != 0 {
			t.Fatalf("want empty list, got %d", len(list))
		}
	})

	t.Run("lists created incidents with monitor count", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "API")
		createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "Elevated latency", Message: "Investigating", Severity: "major",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})

		w := doAuthed(t, http.MethodGet, incidentH.ListIncidents, u.access, nil)
		list := decodeBody[[]incidentResponse](t, w)
		if len(list) != 1 {
			t.Fatalf("want 1 incident, got %d", len(list))
		}
		if list[0].Title != "Elevated latency" {
			t.Fatalf("want title %q, got %q", "Elevated latency", list[0].Title)
		}
		if list[0].Severity != "major" {
			t.Fatalf("want severity major, got %q", list[0].Severity)
		}
		if list[0].Status != "investigating" {
			t.Fatalf("want status investigating, got %q", list[0].Status)
		}
		if list[0].MonitorCount != 1 {
			t.Fatalf("want monitor count 1, got %d", list[0].MonitorCount)
		}
	})
}

func TestCreateIncident(t *testing.T) {
	authH, monitorH, maintH, incidentH, pool := testIncidentHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, incidentH.CreateIncident, http.MethodPost, "/api/v1/incidents", createIncidentRequest{Title: "x"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(incidentH.CreateIncident)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "Validation monitor")
		validMonitors := []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}}

		cases := []struct {
			name string
			req  createIncidentRequest
		}{
			{"missing title", createIncidentRequest{Message: "x", Severity: "major", Monitors: validMonitors}},
			{"missing message", createIncidentRequest{Title: "x", Severity: "major", Monitors: validMonitors}},
			{"invalid severity", createIncidentRequest{Title: "x", Message: "x", Severity: "apocalyptic", Monitors: validMonitors}},
			{"no monitors", createIncidentRequest{Title: "x", Message: "x", Severity: "major"}},
			{"unknown monitor type", createIncidentRequest{Title: "x", Message: "x", Severity: "major", Monitors: []incidentMonitorInput{{MonitorType: "carrier-pigeon", MonitorID: mon.ID}}}},
			{"monitor not found", createIncidentRequest{Title: "x", Message: "x", Severity: "major", Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: "00000000-0000-0000-0000-000000000000"}}}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doAuthed(t, http.MethodPost, incidentH.CreateIncident, u.access, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("a monitor belonging to another org is rejected (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		monB := createUptimeMonitor(t, monitorH, uB.access, "Org B's monitor")

		w := doAuthed(t, http.MethodPost, incidentH.CreateIncident, uA.access, createIncidentRequest{
			Title: "Cross-tenant attempt", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: monB.ID}},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400 (monitor not visible to org A), got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success resolves monitor names, dedups, and seeds the first update", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		uptime := createUptimeMonitor(t, monitorH, u.access, "API uptime")
		ssl := createSSLMonitor(t, monitorH, u.access, "API ssl")

		resp := createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "Multi-monitor outage", Message: "Looking into it", Severity: "critical",
			Monitors: []incidentMonitorInput{
				{MonitorType: "uptime", MonitorID: uptime.ID},
				{MonitorType: "ssl", MonitorID: ssl.ID},
				{MonitorType: "uptime", MonitorID: uptime.ID}, // duplicate, should be deduped
			},
		})
		if resp.Status != "investigating" {
			t.Fatalf("want initial status investigating, got %q", resp.Status)
		}
		if resp.MonitorCount != 2 {
			t.Fatalf("want 2 monitors after dedup, got %d", resp.MonitorCount)
		}
		if len(resp.Updates) != 1 || resp.Updates[0].Message != "Looking into it" {
			t.Fatalf("want a single seeded update, got %+v", resp.Updates)
		}
		if resp.Updates[0].Status != "investigating" {
			t.Fatalf("want seeded update status investigating, got %q", resp.Updates[0].Status)
		}
		names := map[string]bool{}
		for _, m := range resp.Monitors {
			names[m.Name] = true
		}
		for _, want := range []string{"API uptime", "API ssl"} {
			if !names[want] {
				t.Fatalf("want monitor %q in response, got %+v", want, resp.Monitors)
			}
		}
	})

	t.Run("US-2405: declaring against a monitor under active maintenance warns unless confirmed", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "Under maintenance")
		starts := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
		maintW := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, u.access, maintenanceWindowRequest{
			Title:    "Ongoing work",
			StartsAt: starts,
			Monitors: []maintenanceMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})
		if maintW.Code != http.StatusCreated {
			t.Fatalf("seed maintenance window: want 201, got %d: %s", maintW.Code, maintW.Body.String())
		}

		req := createIncidentRequest{
			Title: "Overlaps maintenance", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		}
		warned := doAuthed(t, http.MethodPost, incidentH.CreateIncident, u.access, req)
		if warned.Code != http.StatusConflict {
			t.Fatalf("want 409 maintenance_overlap, got %d: %s", warned.Code, warned.Body.String())
		}
		body := decodeBody[map[string]string](t, warned)
		if body["code"] != "maintenance_overlap" {
			t.Fatalf("want code maintenance_overlap, got %q", body["code"])
		}

		req.ConfirmOverlap = true
		confirmed := doAuthed(t, http.MethodPost, incidentH.CreateIncident, u.access, req)
		if confirmed.Code != http.StatusCreated {
			t.Fatalf("want 201 once overlap is confirmed, got %d: %s", confirmed.Code, confirmed.Body.String())
		}
	})
}

func TestGetIncident(t *testing.T) {
	authH, monitorH, _, incidentH, pool := testIncidentHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/x", nil)
		w := httptest.NewRecorder()
		incidentH.GetIncident(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doIncidentRequest(t, http.MethodGet, incidentH.GetIncident, u.access, "/api/v1/incidents/not-a-uuid", map[string]string{"id": "not-a-uuid"}, nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		id := "00000000-0000-0000-0000-000000000000"
		w := doIncidentRequest(t, http.MethodGet, incidentH.GetIncident, u.access, "/api/v1/incidents/"+id, map[string]string{"id": id}, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant incident is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, uA.access, "Org A monitor")
		created := createIncident(t, incidentH, uA.access, createIncidentRequest{
			Title: "Org A incident", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})

		w := doIncidentRequest(t, http.MethodGet, incidentH.GetIncident, uB.access, "/api/v1/incidents/"+created.ID, map[string]string{"id": created.ID}, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 fetching org A's incident as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success, and a deleted monitor shows a placeholder name", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		uptime := createUptimeMonitor(t, monitorH, u.access, "Will be deleted")
		ssl := createSSLMonitor(t, monitorH, u.access, "Stays")
		created := createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "Get test", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{
				{MonitorType: "uptime", MonitorID: uptime.ID},
				{MonitorType: "ssl", MonitorID: ssl.ID},
			},
		})

		if _, err := pool.Exec(context.Background(), "DELETE FROM uptime_monitors WHERE id = $1", uptime.ID); err != nil {
			t.Fatalf("delete uptime monitor: %v", err)
		}

		w := doIncidentRequest(t, http.MethodGet, incidentH.GetIncident, u.access, "/api/v1/incidents/"+created.ID, map[string]string{"id": created.ID}, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[incidentResponse](t, w)
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

func TestUpdateIncidentTitle(t *testing.T) {
	authH, monitorH, _, incidentH, pool := testIncidentHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/incidents/x", nil)
		w := httptest.NewRecorder()
		incidentH.UpdateIncidentTitle(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("empty title is rejected", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "x")
		created := createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "Original", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})
		w := doIncidentRequest(t, http.MethodPatch, incidentH.UpdateIncidentTitle, u.access, "/api/v1/incidents/"+created.ID, map[string]string{"id": created.ID}, updateTitleRequest{Title: "   "})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		id := "00000000-0000-0000-0000-000000000000"
		w := doIncidentRequest(t, http.MethodPatch, incidentH.UpdateIncidentTitle, u.access, "/api/v1/incidents/"+id, map[string]string{"id": id}, updateTitleRequest{Title: "x"})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant update is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, uA.access, "Org A monitor")
		created := createIncident(t, incidentH, uA.access, createIncidentRequest{
			Title: "Org A incident", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})

		w := doIncidentRequest(t, http.MethodPatch, incidentH.UpdateIncidentTitle, uB.access, "/api/v1/incidents/"+created.ID, map[string]string{"id": created.ID}, updateTitleRequest{Title: "hijacked"})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 updating org A's incident as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success renames and trims whitespace", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "x")
		created := createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "Original", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})

		w := doIncidentRequest(t, http.MethodPatch, incidentH.UpdateIncidentTitle, u.access, "/api/v1/incidents/"+created.ID, map[string]string{"id": created.ID}, updateTitleRequest{Title: "  Renamed  "})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[incidentResponse](t, w)
		if resp.Title != "Renamed" {
			t.Fatalf("want title Renamed (trimmed), got %q", resp.Title)
		}
	})
}

func TestDeleteIncident(t *testing.T) {
	authH, monitorH, _, incidentH, pool := testIncidentHandler(t)

	t.Run("cross-tenant delete is a no-op, owner delete succeeds and cascades", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, uA.access, "Org A monitor")
		created := createIncident(t, incidentH, uA.access, createIncidentRequest{
			Title: "Org A incident", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})

		wrongDelete := doIncidentRequest(t, http.MethodDelete, incidentH.DeleteIncident, uB.access, "/api/v1/incidents/"+created.ID, map[string]string{"id": created.ID}, nil)
		if wrongDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204 (delete is a no-op exec regardless of match), got %d", wrongDelete.Code)
		}
		stillThere := doIncidentRequest(t, http.MethodGet, incidentH.GetIncident, uA.access, "/api/v1/incidents/"+created.ID, map[string]string{"id": created.ID}, nil)
		if stillThere.Code != http.StatusOK {
			t.Fatalf("want org A's incident to survive org B's delete attempt, got %d", stillThere.Code)
		}

		ownerDelete := doIncidentRequest(t, http.MethodDelete, incidentH.DeleteIncident, uA.access, "/api/v1/incidents/"+created.ID, map[string]string{"id": created.ID}, nil)
		if ownerDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", ownerDelete.Code, ownerDelete.Body.String())
		}
		gone := doIncidentRequest(t, http.MethodGet, incidentH.GetIncident, uA.access, "/api/v1/incidents/"+created.ID, map[string]string{"id": created.ID}, nil)
		if gone.Code != http.StatusNotFound {
			t.Fatalf("want 404 after delete, got %d", gone.Code)
		}

		var updateCount int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM status_page_incident_updates WHERE incident_id = $1", created.ID).Scan(&updateCount); err != nil {
			t.Fatalf("count updates: %v", err)
		}
		if updateCount != 0 {
			t.Fatalf("want updates cascade-deleted with the incident, got %d remaining", updateCount)
		}
	})
}

func TestPostIncidentUpdate(t *testing.T) {
	authH, monitorH, _, incidentH, pool := testIncidentHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/x/updates", nil)
		w := httptest.NewRecorder()
		incidentH.PostIncidentUpdate(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		id := "00000000-0000-0000-0000-000000000000"
		w := doIncidentRequest(t, http.MethodPost, incidentH.PostIncidentUpdate, u.access, "/api/v1/incidents/"+id+"/updates", map[string]string{"id": id}, postUpdateRequest{Message: "x", Status: "identified"})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "x")
		created := createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "x", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})
		cases := []struct {
			name string
			req  postUpdateRequest
		}{
			{"missing message", postUpdateRequest{Status: "identified"}},
			{"invalid status", postUpdateRequest{Message: "x", Status: "raptured"}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doIncidentRequest(t, http.MethodPost, incidentH.PostIncidentUpdate, u.access, "/api/v1/incidents/"+created.ID+"/updates", map[string]string{"id": created.ID}, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("success appends an update and advances status through to resolved", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "x")
		created := createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "x", Message: "Initial report", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})

		identified := doIncidentRequest(t, http.MethodPost, incidentH.PostIncidentUpdate, u.access, "/api/v1/incidents/"+created.ID+"/updates", map[string]string{"id": created.ID}, postUpdateRequest{Message: "Found the cause", Status: "identified"})
		if identified.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", identified.Code, identified.Body.String())
		}
		identifiedResp := decodeBody[incidentResponse](t, identified)
		if identifiedResp.Status != "identified" {
			t.Fatalf("want status identified, got %q", identifiedResp.Status)
		}
		if len(identifiedResp.Updates) != 2 {
			t.Fatalf("want 2 updates (initial + identified), got %d", len(identifiedResp.Updates))
		}
		if identifiedResp.ResolvedAt != nil {
			t.Fatal("want resolvedAt still nil before resolution")
		}

		resolved := doIncidentRequest(t, http.MethodPost, incidentH.PostIncidentUpdate, u.access, "/api/v1/incidents/"+created.ID+"/updates", map[string]string{"id": created.ID}, postUpdateRequest{Message: "All clear", Status: "resolved"})
		if resolved.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", resolved.Code, resolved.Body.String())
		}
		resolvedResp := decodeBody[incidentResponse](t, resolved)
		if resolvedResp.Status != "resolved" {
			t.Fatalf("want status resolved, got %q", resolvedResp.Status)
		}
		if resolvedResp.ResolvedAt == nil {
			t.Fatal("want resolvedAt set once resolved")
		}
	})
}

func TestUpdateIncidentUpdateMessage(t *testing.T) {
	authH, monitorH, _, incidentH, pool := testIncidentHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/incidents/x/updates/y", nil)
		w := httptest.NewRecorder()
		incidentH.UpdateIncidentUpdateMessage(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid update id", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "x")
		created := createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "x", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})
		w := doIncidentRequest(t, http.MethodPatch, incidentH.UpdateIncidentUpdateMessage, u.access,
			"/api/v1/incidents/"+created.ID+"/updates/not-a-uuid",
			map[string]string{"id": created.ID, "updateId": "not-a-uuid"}, updateMessageRequest{Message: "x"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("incident not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		id := "00000000-0000-0000-0000-000000000000"
		updateID := "00000000-0000-0000-0000-000000000001"
		w := doIncidentRequest(t, http.MethodPatch, incidentH.UpdateIncidentUpdateMessage, u.access,
			"/api/v1/incidents/"+id+"/updates/"+updateID,
			map[string]string{"id": id, "updateId": updateID}, updateMessageRequest{Message: "x"})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("empty message is rejected", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "x")
		created := createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "x", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})
		updateID := created.Updates[0].ID
		w := doIncidentRequest(t, http.MethodPatch, incidentH.UpdateIncidentUpdateMessage, u.access,
			"/api/v1/incidents/"+created.ID+"/updates/"+updateID,
			map[string]string{"id": created.ID, "updateId": updateID}, updateMessageRequest{Message: "  "})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update belonging to a different incident is not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "x")
		incidentA := createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "A", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})
		incidentB := createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "B", Message: "x", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})
		updateFromB := incidentB.Updates[0].ID

		w := doIncidentRequest(t, http.MethodPatch, incidentH.UpdateIncidentUpdateMessage, u.access,
			"/api/v1/incidents/"+incidentA.ID+"/updates/"+updateFromB,
			map[string]string{"id": incidentA.ID, "updateId": updateFromB}, updateMessageRequest{Message: "hijacked"})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 (update belongs to a different incident), got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success edits the message in place", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "x")
		created := createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "x", Message: "Original message", Severity: "minor",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})
		updateID := created.Updates[0].ID

		w := doIncidentRequest(t, http.MethodPatch, incidentH.UpdateIncidentUpdateMessage, u.access,
			"/api/v1/incidents/"+created.ID+"/updates/"+updateID,
			map[string]string{"id": created.ID, "updateId": updateID}, updateMessageRequest{Message: "  Corrected message  "})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[incidentResponse](t, w)
		if len(resp.Updates) != 1 || resp.Updates[0].Message != "Corrected message" {
			t.Fatalf("want the single update edited in place (trimmed), got %+v", resp.Updates)
		}
	})
}
