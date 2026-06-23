package handler

// Integration tests for the domain monitor handlers in domain_monitors.go.
// Mirrors ssl_monitors_test.go closely — same shape, same shared
// testMonitorHandler/createDomainMonitor helpers. Real Postgres (ADR-010),
// package handler so unexported request/response types are reused directly.

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

// doDomainMonitorRequest authenticates via RequireAuth + the access cookie
// and injects the chi "id" URL param these handlers read via chi.URLParam.
func doDomainMonitorRequest(t *testing.T, method string, handler http.HandlerFunc, access *http.Cookie, id string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader = http.NoBody
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "/api/v1/monitors/domain/"+id, r)
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", id)
	req.AddCookie(access)
	w := httptest.NewRecorder()
	apimiddleware.RequireAuth(testJWTSecret)(handler).ServeHTTP(w, req)
	return w
}

func TestListDomainMonitors(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/domain", nil)
		w := httptest.NewRecorder()
		monitorH.ListDomainMonitors(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("empty for a fresh org", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodGet, monitorH.ListDomainMonitors, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		list := decodeBody[[]domainMonitorResponse](t, w)
		if len(list) != 0 {
			t.Fatalf("want empty list, got %d", len(list))
		}
	})

	t.Run("lists created monitors", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		createDomainMonitor(t, monitorH, u.access, "example.com domain")
		createDomainMonitor(t, monitorH, u.access, "api.example.com domain")

		w := doAuthed(t, http.MethodGet, monitorH.ListDomainMonitors, u.access, nil)
		list := decodeBody[[]domainMonitorResponse](t, w)
		if len(list) != 2 {
			t.Fatalf("want 2 monitors, got %d", len(list))
		}
	})
}

func TestCreateDomainMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, monitorH.CreateDomainMonitor, http.MethodPost, "/api/v1/monitors/domain", createDomainMonitorRequest{Name: "x", Domain: "example.com"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/monitors/domain", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(monitorH.CreateDomainMonitor)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cases := []struct {
			name string
			req  createDomainMonitorRequest
		}{
			{"missing name", createDomainMonitorRequest{Domain: "example.com"}},
			{"missing domain", createDomainMonitorRequest{Name: "x"}},
			{"domain is just a scheme", createDomainMonitorRequest{Name: "x", Domain: "https://"}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doAuthed(t, http.MethodPost, monitorH.CreateDomainMonitor, u.access, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("domain is normalised from a full URL", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateDomainMonitor, u.access, createDomainMonitorRequest{
			Name: "Normalised", Domain: "https://example.com:8443/health?check=1",
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[domainMonitorResponse](t, w)
		if resp.Domain != "example.com" {
			t.Fatalf("want domain normalised to example.com, got %q", resp.Domain)
		}
	})

	t.Run("success defaults to waiting status with alerts enabled", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateDomainMonitor, u.access, createDomainMonitorRequest{
			Name: "Fresh monitor", Domain: "example.com",
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[domainMonitorResponse](t, w)
		if resp.Status != "waiting" {
			t.Fatalf("want status waiting, got %q", resp.Status)
		}
		if !resp.AlertsEnabled {
			t.Fatal("want alerts enabled by default")
		}
		if resp.ExpiresAt != nil || resp.Registrar != nil || resp.ErrorMsg != nil || resp.LastCheckedAt != nil {
			t.Fatalf("want all check-result fields nil before the first check, got %+v", resp)
		}
	})

	t.Run("plan limit enforced at 10 monitors on Hobby", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		for i := 0; i < 10; i++ {
			createDomainMonitor(t, monitorH, u.access, "Monitor")
		}
		w := doAuthed(t, http.MethodPost, monitorH.CreateDomainMonitor, u.access, createDomainMonitorRequest{
			Name: "One too many", Domain: "example.com",
		})
		if w.Code != http.StatusPaymentRequired {
			t.Fatalf("want 402, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "plan_limit_reached" {
			t.Fatalf("want code plan_limit_reached, got %q", body["code"])
		}
	})

	t.Run("counts toward the shared aggregate limit alongside other monitor types", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		for i := 0; i < 5; i++ {
			createCronMonitor(t, monitorH, u.access, "Cron")
		}
		for i := 0; i < 5; i++ {
			createDomainMonitor(t, monitorH, u.access, "Domain")
		}
		w := doAuthed(t, http.MethodPost, monitorH.CreateDomainMonitor, u.access, createDomainMonitorRequest{
			Name: "One too many", Domain: "example.com",
		})
		if w.Code != http.StatusPaymentRequired {
			t.Fatalf("want 402 once cron+domain monitors hit the shared limit of 10, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestGetDomainMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/domain/x", nil)
		w := httptest.NewRecorder()
		monitorH.GetDomainMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doDomainMonitorRequest(t, http.MethodGet, monitorH.GetDomainMonitor, u.access, "not-a-uuid", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doDomainMonitorRequest(t, http.MethodGet, monitorH.GetDomainMonitor, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant monitor is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createDomainMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doDomainMonitorRequest(t, http.MethodGet, monitorH.GetDomainMonitor, uB.access, mon.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 fetching org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success reflects the latest check result", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createDomainMonitor(t, monitorH, u.access, "Checked monitor")
		expiresAt := time.Now().Add(60 * 24 * time.Hour)
		seedDomainCheckResult(t, pool, mon.ID, expiresAt, "Example Registrar, LLC", "")

		w := doDomainMonitorRequest(t, http.MethodGet, monitorH.GetDomainMonitor, u.access, mon.ID, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[domainMonitorResponse](t, w)
		if resp.Registrar == nil || *resp.Registrar != "Example Registrar, LLC" {
			t.Fatalf("want registrar Example Registrar, LLC, got %v", resp.Registrar)
		}
		if resp.ExpiresAt == nil {
			t.Fatal("want expiresAt set")
		}
		if resp.LastCheckedAt == nil {
			t.Fatal("want lastCheckedAt set")
		}
		if resp.DaysUntilExpiry == nil {
			t.Fatal("want daysUntilExpiry computed")
		}
		want := int(time.Until(expiresAt).Hours() / 24)
		if diff := *resp.DaysUntilExpiry - want; diff < -1 || diff > 1 {
			t.Fatalf("want daysUntilExpiry ~%d, got %d", want, *resp.DaysUntilExpiry)
		}
	})

	t.Run("an error result is reflected without an expiry", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createDomainMonitor(t, monitorH, u.access, "Erroring monitor")
		if _, err := pool.Exec(context.Background(),
			"UPDATE domain_monitors SET status = 'error', error_msg = $2, last_checked_at = NOW() WHERE id = $1",
			mon.ID, "rdap lookup failed",
		); err != nil {
			t.Fatalf("seed error result: %v", err)
		}

		w := doDomainMonitorRequest(t, http.MethodGet, monitorH.GetDomainMonitor, u.access, mon.ID, nil)
		resp := decodeBody[domainMonitorResponse](t, w)
		if resp.Status != "error" {
			t.Fatalf("want status error, got %q", resp.Status)
		}
		if resp.ErrorMsg == nil || *resp.ErrorMsg != "rdap lookup failed" {
			t.Fatalf("want error message, got %v", resp.ErrorMsg)
		}
		if resp.ExpiresAt != nil || resp.DaysUntilExpiry != nil {
			t.Fatalf("want no expiry on an error result, got expiresAt=%v daysUntilExpiry=%v", resp.ExpiresAt, resp.DaysUntilExpiry)
		}
	})
}

func TestUpdateDomainMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/monitors/domain/x", nil)
		w := httptest.NewRecorder()
		monitorH.UpdateDomainMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createDomainMonitor(t, monitorH, u.access, "x")
		cases := []struct {
			name string
			req  updateDomainMonitorRequest
		}{
			{"missing name", updateDomainMonitorRequest{Domain: "example.com"}},
			{"missing domain", updateDomainMonitorRequest{Name: "x"}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doDomainMonitorRequest(t, http.MethodPatch, monitorH.UpdateDomainMonitor, u.access, mon.ID, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doDomainMonitorRequest(t, http.MethodPatch, monitorH.UpdateDomainMonitor, u.access, "00000000-0000-0000-0000-000000000000", updateDomainMonitorRequest{
			Name: "x", Domain: "example.com",
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant update is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createDomainMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doDomainMonitorRequest(t, http.MethodPatch, monitorH.UpdateDomainMonitor, uB.access, mon.ID, updateDomainMonitorRequest{
			Name: "hijacked", Domain: "example.com",
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 updating org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success applies changes and normalises the domain", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createDomainMonitor(t, monitorH, u.access, "Original name")

		w := doDomainMonitorRequest(t, http.MethodPatch, monitorH.UpdateDomainMonitor, u.access, mon.ID, updateDomainMonitorRequest{
			Name: "Renamed", Domain: "https://new-domain.example.com/path", AlertsEnabled: false,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[domainMonitorResponse](t, w)
		if resp.Name != "Renamed" {
			t.Fatalf("want name Renamed, got %q", resp.Name)
		}
		if resp.Domain != "new-domain.example.com" {
			t.Fatalf("want normalised domain, got %q", resp.Domain)
		}
		if resp.AlertsEnabled {
			t.Fatal("want alertsEnabled false")
		}
	})
}

func TestPauseResumeDomainMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/monitors/domain/x/pause", nil)
		w := httptest.NewRecorder()
		monitorH.PauseDomainMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doDomainMonitorRequest(t, http.MethodPost, monitorH.PauseDomainMonitor, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant pause is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createDomainMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doDomainMonitorRequest(t, http.MethodPost, monitorH.PauseDomainMonitor, uB.access, mon.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 pausing org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("pause then resume round-trips status", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createDomainMonitor(t, monitorH, u.access, "Togglable")

		pauseW := doDomainMonitorRequest(t, http.MethodPost, monitorH.PauseDomainMonitor, u.access, mon.ID, nil)
		if pauseW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", pauseW.Code, pauseW.Body.String())
		}
		paused := decodeBody[domainMonitorResponse](t, pauseW)
		if paused.Status != "paused" {
			t.Fatalf("want status paused, got %q", paused.Status)
		}

		resumeW := doDomainMonitorRequest(t, http.MethodPost, monitorH.ResumeDomainMonitor, u.access, mon.ID, nil)
		if resumeW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", resumeW.Code, resumeW.Body.String())
		}
		resumed := decodeBody[domainMonitorResponse](t, resumeW)
		if resumed.Status != "waiting" {
			t.Fatalf("want status waiting after resume, got %q", resumed.Status)
		}
	})
}

func TestDeleteDomainMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/monitors/domain/x", nil)
		w := httptest.NewRecorder()
		monitorH.DeleteDomainMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("cross-tenant delete is a no-op, owner delete succeeds", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createDomainMonitor(t, monitorH, uA.access, "Org A monitor")

		wrongDelete := doDomainMonitorRequest(t, http.MethodDelete, monitorH.DeleteDomainMonitor, uB.access, mon.ID, nil)
		if wrongDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204 (delete is a no-op exec regardless of match), got %d", wrongDelete.Code)
		}
		stillThere := doDomainMonitorRequest(t, http.MethodGet, monitorH.GetDomainMonitor, uA.access, mon.ID, nil)
		if stillThere.Code != http.StatusOK {
			t.Fatalf("want org A's monitor to survive org B's delete attempt, got %d", stillThere.Code)
		}

		ownerDelete := doDomainMonitorRequest(t, http.MethodDelete, monitorH.DeleteDomainMonitor, uA.access, mon.ID, nil)
		if ownerDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", ownerDelete.Code, ownerDelete.Body.String())
		}
		gone := doDomainMonitorRequest(t, http.MethodGet, monitorH.GetDomainMonitor, uA.access, mon.ID, nil)
		if gone.Code != http.StatusNotFound {
			t.Fatalf("want 404 after delete, got %d", gone.Code)
		}
	})
}

func seedDomainCheckResult(t *testing.T, pool *pgxpool.Pool, monitorID string, expiresAt time.Time, registrar, errorMsg string) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		"UPDATE domain_monitors SET status = 'up', expires_at = $2, registrar = $3, error_msg = NULLIF($4, ''), last_checked_at = NOW() WHERE id = $1",
		monitorID, expiresAt, registrar, errorMsg,
	); err != nil {
		t.Fatalf("seed domain check result: %v", err)
	}
}
