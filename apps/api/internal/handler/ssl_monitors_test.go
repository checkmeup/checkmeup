package handler

// Integration tests for the SSL monitor handlers in ssl_monitors.go. Same
// conventions as monitors_test.go (which these largely mirror, since SSL
// monitors share the MonitorHandler/testMonitorHandler/createSSLMonitor
// helpers already defined there and in maintenance_test.go): real Postgres
// (ADR-010), package handler so unexported request/response types are
// reused directly.

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

// doSSLMonitorRequest authenticates via RequireAuth + the access cookie and
// injects the chi "id" URL param these handlers read via chi.URLParam.
func doSSLMonitorRequest(t *testing.T, method string, handler http.HandlerFunc, access *http.Cookie, id string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader = http.NoBody
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "/api/v1/monitors/ssl/"+id, r)
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", id)
	req.AddCookie(access)
	w := httptest.NewRecorder()
	apimiddleware.RequireAuth(testJWTSecret)(handler).ServeHTTP(w, req)
	return w
}

func TestListSSLMonitors(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/ssl", nil)
		w := httptest.NewRecorder()
		monitorH.ListSSLMonitors(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("empty for a fresh org", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodGet, monitorH.ListSSLMonitors, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		list := decodeBody[[]sslMonitorResponse](t, w)
		if len(list) != 0 {
			t.Fatalf("want empty list, got %d", len(list))
		}
	})

	t.Run("lists created monitors", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		createSSLMonitor(t, monitorH, u.access, "example.com cert")
		createSSLMonitor(t, monitorH, u.access, "api.example.com cert")

		w := doAuthed(t, http.MethodGet, monitorH.ListSSLMonitors, u.access, nil)
		list := decodeBody[[]sslMonitorResponse](t, w)
		if len(list) != 2 {
			t.Fatalf("want 2 monitors, got %d", len(list))
		}
	})
}

func TestCreateSSLMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, monitorH.CreateSSLMonitor, http.MethodPost, "/api/v1/monitors/ssl", createSSLMonitorRequest{Name: "x", Hostname: "example.com"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/monitors/ssl", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(monitorH.CreateSSLMonitor)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cases := []struct {
			name string
			req  createSSLMonitorRequest
		}{
			{"missing name", createSSLMonitorRequest{Hostname: "example.com"}},
			{"missing hostname", createSSLMonitorRequest{Name: "x"}},
			{"hostname is just a scheme", createSSLMonitorRequest{Name: "x", Hostname: "https://"}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doAuthed(t, http.MethodPost, monitorH.CreateSSLMonitor, u.access, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("hostname is normalised from a full URL", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateSSLMonitor, u.access, createSSLMonitorRequest{
			Name: "Normalised", Hostname: "https://example.com:8443/health?check=1",
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[sslMonitorResponse](t, w)
		if resp.Hostname != "example.com" {
			t.Fatalf("want hostname normalised to example.com, got %q", resp.Hostname)
		}
	})

	t.Run("success defaults to waiting status with alerts enabled", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, monitorH.CreateSSLMonitor, u.access, createSSLMonitorRequest{
			Name: "Fresh monitor", Hostname: "example.com",
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[sslMonitorResponse](t, w)
		if resp.Status != "waiting" {
			t.Fatalf("want status waiting, got %q", resp.Status)
		}
		if !resp.AlertsEnabled {
			t.Fatal("want alerts enabled by default")
		}
		if resp.ExpiresAt != nil || resp.Issuer != nil || resp.ErrorMsg != nil || resp.LastCheckedAt != nil {
			t.Fatalf("want all check-result fields nil before the first check, got %+v", resp)
		}
	})

	t.Run("plan limit enforced at 10 monitors on Hobby", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		for i := 0; i < 10; i++ {
			createSSLMonitor(t, monitorH, u.access, "Monitor")
		}
		w := doAuthed(t, http.MethodPost, monitorH.CreateSSLMonitor, u.access, createSSLMonitorRequest{
			Name: "One too many", Hostname: "example.com",
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

func TestGetSSLMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/ssl/x", nil)
		w := httptest.NewRecorder()
		monitorH.GetSSLMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doSSLMonitorRequest(t, http.MethodGet, monitorH.GetSSLMonitor, u.access, "not-a-uuid", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doSSLMonitorRequest(t, http.MethodGet, monitorH.GetSSLMonitor, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant monitor is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createSSLMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doSSLMonitorRequest(t, http.MethodGet, monitorH.GetSSLMonitor, uB.access, mon.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 fetching org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success reflects the latest check result", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createSSLMonitor(t, monitorH, u.access, "Checked monitor")
		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		seedSSLCheckResult(t, pool, mon.ID, expiresAt, "Let's Encrypt", "")

		w := doSSLMonitorRequest(t, http.MethodGet, monitorH.GetSSLMonitor, u.access, mon.ID, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[sslMonitorResponse](t, w)
		if resp.Issuer == nil || *resp.Issuer != "Let's Encrypt" {
			t.Fatalf("want issuer Let's Encrypt, got %v", resp.Issuer)
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
		mon := createSSLMonitor(t, monitorH, u.access, "Erroring monitor")
		if _, err := pool.Exec(context.Background(),
			"UPDATE ssl_monitors SET status = 'error', error_msg = $2, last_checked_at = NOW() WHERE id = $1",
			mon.ID, "connection refused",
		); err != nil {
			t.Fatalf("seed error result: %v", err)
		}

		w := doSSLMonitorRequest(t, http.MethodGet, monitorH.GetSSLMonitor, u.access, mon.ID, nil)
		resp := decodeBody[sslMonitorResponse](t, w)
		if resp.Status != "error" {
			t.Fatalf("want status error, got %q", resp.Status)
		}
		if resp.ErrorMsg == nil || *resp.ErrorMsg != "connection refused" {
			t.Fatalf("want error message, got %v", resp.ErrorMsg)
		}
		if resp.ExpiresAt != nil || resp.DaysUntilExpiry != nil {
			t.Fatalf("want no expiry on an error result, got expiresAt=%v daysUntilExpiry=%v", resp.ExpiresAt, resp.DaysUntilExpiry)
		}
	})
}

func TestUpdateSSLMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/monitors/ssl/x", nil)
		w := httptest.NewRecorder()
		monitorH.UpdateSSLMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createSSLMonitor(t, monitorH, u.access, "x")
		cases := []struct {
			name string
			req  updateSSLMonitorRequest
		}{
			{"missing name", updateSSLMonitorRequest{Hostname: "example.com"}},
			{"missing hostname", updateSSLMonitorRequest{Name: "x"}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doSSLMonitorRequest(t, http.MethodPatch, monitorH.UpdateSSLMonitor, u.access, mon.ID, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doSSLMonitorRequest(t, http.MethodPatch, monitorH.UpdateSSLMonitor, u.access, "00000000-0000-0000-0000-000000000000", updateSSLMonitorRequest{
			Name: "x", Hostname: "example.com",
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant update is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createSSLMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doSSLMonitorRequest(t, http.MethodPatch, monitorH.UpdateSSLMonitor, uB.access, mon.ID, updateSSLMonitorRequest{
			Name: "hijacked", Hostname: "example.com",
		})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 updating org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success applies changes and normalises the hostname", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createSSLMonitor(t, monitorH, u.access, "Original name")

		w := doSSLMonitorRequest(t, http.MethodPatch, monitorH.UpdateSSLMonitor, u.access, mon.ID, updateSSLMonitorRequest{
			Name: "Renamed", Hostname: "https://new-host.example.com/path", AlertsEnabled: false,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[sslMonitorResponse](t, w)
		if resp.Name != "Renamed" {
			t.Fatalf("want name Renamed, got %q", resp.Name)
		}
		if resp.Hostname != "new-host.example.com" {
			t.Fatalf("want normalised hostname, got %q", resp.Hostname)
		}
		if resp.AlertsEnabled {
			t.Fatal("want alertsEnabled false")
		}
	})
}

func TestPauseResumeSSLMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/monitors/ssl/x/pause", nil)
		w := httptest.NewRecorder()
		monitorH.PauseSSLMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doSSLMonitorRequest(t, http.MethodPost, monitorH.PauseSSLMonitor, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant pause is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createSSLMonitor(t, monitorH, uA.access, "Org A monitor")

		w := doSSLMonitorRequest(t, http.MethodPost, monitorH.PauseSSLMonitor, uB.access, mon.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 pausing org A's monitor as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("pause then resume round-trips status", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createSSLMonitor(t, monitorH, u.access, "Togglable")

		pauseW := doSSLMonitorRequest(t, http.MethodPost, monitorH.PauseSSLMonitor, u.access, mon.ID, nil)
		if pauseW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", pauseW.Code, pauseW.Body.String())
		}
		paused := decodeBody[sslMonitorResponse](t, pauseW)
		if paused.Status != "paused" {
			t.Fatalf("want status paused, got %q", paused.Status)
		}

		resumeW := doSSLMonitorRequest(t, http.MethodPost, monitorH.ResumeSSLMonitor, u.access, mon.ID, nil)
		if resumeW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", resumeW.Code, resumeW.Body.String())
		}
		resumed := decodeBody[sslMonitorResponse](t, resumeW)
		if resumed.Status != "waiting" {
			t.Fatalf("want status waiting after resume, got %q", resumed.Status)
		}
	})
}

func TestDeleteSSLMonitor(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/monitors/ssl/x", nil)
		w := httptest.NewRecorder()
		monitorH.DeleteSSLMonitor(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("cross-tenant delete is a no-op, owner delete succeeds", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		mon := createSSLMonitor(t, monitorH, uA.access, "Org A monitor")

		wrongDelete := doSSLMonitorRequest(t, http.MethodDelete, monitorH.DeleteSSLMonitor, uB.access, mon.ID, nil)
		if wrongDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204 (delete is a no-op exec regardless of match), got %d", wrongDelete.Code)
		}
		stillThere := doSSLMonitorRequest(t, http.MethodGet, monitorH.GetSSLMonitor, uA.access, mon.ID, nil)
		if stillThere.Code != http.StatusOK {
			t.Fatalf("want org A's monitor to survive org B's delete attempt, got %d", stillThere.Code)
		}

		ownerDelete := doSSLMonitorRequest(t, http.MethodDelete, monitorH.DeleteSSLMonitor, uA.access, mon.ID, nil)
		if ownerDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", ownerDelete.Code, ownerDelete.Body.String())
		}
		gone := doSSLMonitorRequest(t, http.MethodGet, monitorH.GetSSLMonitor, uA.access, mon.ID, nil)
		if gone.Code != http.StatusNotFound {
			t.Fatalf("want 404 after delete, got %d", gone.Code)
		}
	})
}

func TestParseHostname(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"example.com", "example.com", false},
		{"https://example.com", "example.com", false},
		{"http://example.com/path", "example.com", false},
		{"example.com:443", "example.com", false},
		{"https://example.com:8443/foo/bar", "example.com", false},
		{"", "", true},
		{"   ", "", true},
		{"https://", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := parseHostname(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseHostname(%q) error = %v, wantErr %v", tc.in, err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Fatalf("parseHostname(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func seedSSLCheckResult(t *testing.T, pool *pgxpool.Pool, monitorID string, expiresAt time.Time, issuer, errorMsg string) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		"UPDATE ssl_monitors SET status = 'up', expires_at = $2, issuer = $3, error_msg = NULLIF($4, ''), last_checked_at = NOW() WHERE id = $1",
		monitorID, expiresAt, issuer, errorMsg,
	); err != nil {
		t.Fatalf("seed ssl check result: %v", err)
	}
}
