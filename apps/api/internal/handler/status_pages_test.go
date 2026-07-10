package handler

// Integration tests for the status page admin handlers. Same conventions as
// the other *_test.go files in this package: real Postgres (ADR-010),
// package handler so unexported request/response types and the
// MonitorHandler/createCronMonitor-style helpers (from monitors_test.go,
// maintenance_test.go) are reused directly.
//
// CheckSlug doesn't check auth internally (no orgIDFrom call — it's a
// read-only slug-availability lookup) but is routed behind RequireAuth in
// server.go, same pattern already noted for settings.go's TestTelegram/
// TestEmail. Its "unauthenticated" case goes through the middleware
// explicitly rather than calling the handler directly.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

func testStatusPageHandler(t *testing.T) (*AuthHandler, *MonitorHandler, *StatusPageHandler, *pgxpool.Pool) {
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
	authH := NewAuthHandler(cfg, pool)
	monitorH := NewMonitorHandler(cfg, pool, telegram.NewClient(""))
	statusH := NewStatusPageHandler(pool)
	return authH, monitorH, statusH, pool
}

// doStatusPageRequest authenticates via RequireAuth + the access cookie and
// injects the chi "id" URL param these handlers read via chi.URLParam.
func doStatusPageRequest(t *testing.T, method string, handler http.HandlerFunc, access *http.Cookie, id string, body any) *httptest.ResponseRecorder {
	t.Helper()
	r := bytes.NewReader(nil)
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "/api/v1/status-pages/"+id, r)
	req.Header.Set("Content-Type", "application/json")
	req = withURLParam(req, "id", id)
	req.AddCookie(access)
	w := httptest.NewRecorder()
	apimiddleware.RequireAuth(testJWTSecret)(handler).ServeHTTP(w, req)
	return w
}

func uniqueSlug(t *testing.T) string {
	t.Helper()
	// Slugs are lowercase letters/digits/hyphens only — uuid.NewString() is
	// already lowercase hex with hyphens, so it's a valid slug as-is.
	return "page-" + uuid.NewString()
}

func TestCheckSlug(t *testing.T) {
	authH, _, statusH, pool := testStatusPageHandler(t)

	t.Run("unauthenticated (routed behind RequireAuth, like TestTelegram/TestEmail)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/status-pages/check-slug?slug=foo", nil)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(statusH.CheckSlug)).ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid slug format reports unavailable with a reason", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/status-pages/check-slug?slug=AB", nil)
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(statusH.CheckSlug)).ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[map[string]any](t, w)
		if avail, _ := resp["available"].(bool); avail {
			t.Fatal("want an invalid-format slug reported unavailable")
		}
		if resp["reason"] == "" {
			t.Fatal("want a non-empty reason for an invalid slug")
		}
	})

	t.Run("a fresh valid slug is available", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		slug := uniqueSlug(t)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/status-pages/check-slug?slug="+slug, nil)
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(statusH.CheckSlug)).ServeHTTP(w, req)
		resp := decodeBody[map[string]any](t, w)
		if avail, _ := resp["available"].(bool); !avail {
			t.Fatalf("want a fresh slug available, got %+v", resp)
		}
	})

	t.Run("a taken slug is reported unavailable", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "x"})
		if createW.Code != http.StatusCreated {
			t.Fatalf("setup: want 201, got %d: %s", createW.Code, createW.Body.String())
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/status-pages/check-slug?slug="+slug, nil)
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(statusH.CheckSlug)).ServeHTTP(w, req)
		resp := decodeBody[map[string]any](t, w)
		if avail, _ := resp["available"].(bool); avail {
			t.Fatal("want a taken slug reported unavailable")
		}
		if resp["reason"] != "slug is already taken" {
			t.Fatalf("want reason 'slug is already taken', got %v", resp["reason"])
		}
	})
}

func TestListStatusPages(t *testing.T) {
	authH, _, statusH, pool := testStatusPageHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/status-pages", nil)
		w := httptest.NewRecorder()
		statusH.ListStatusPages(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("empty for a fresh org", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodGet, statusH.ListStatusPages, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		list := decodeBody[[]statusPageResponse](t, w)
		if len(list) != 0 {
			t.Fatalf("want empty list, got %d", len(list))
		}
	})

	t.Run("lists a created page", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "My status"})
		if createW.Code != http.StatusCreated {
			t.Fatalf("setup: want 201, got %d", createW.Code)
		}

		w := doAuthed(t, http.MethodGet, statusH.ListStatusPages, u.access, nil)
		list := decodeBody[[]statusPageResponse](t, w)
		if len(list) != 1 {
			t.Fatalf("want 1 page, got %d", len(list))
		}
		if list[0].Slug != slug {
			t.Fatalf("want slug %q, got %q", slug, list[0].Slug)
		}
	})
}

func TestCreateStatusPage(t *testing.T) {
	authH, _, statusH, pool := testStatusPageHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, statusH.CreateStatusPage, http.MethodPost, "/api/v1/status-pages", createStatusPageRequest{Slug: uniqueSlug(t), Title: "x"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/status-pages", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(statusH.CreateStatusPage)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cases := []struct {
			name string
			req  createStatusPageRequest
		}{
			{"slug too short", createStatusPageRequest{Slug: "ab", Title: "x"}},
			{"slug with invalid characters", createStatusPageRequest{Slug: "abc_def!", Title: "x"}},
			{"slug starting with hyphen", createStatusPageRequest{Slug: "-abcdef", Title: "x"}},
			{"missing title", createStatusPageRequest{Slug: uniqueSlug(t)}},
			{"javascript: logo URL", createStatusPageRequest{Slug: uniqueSlug(t), Title: "x", LogoURL: "javascript:alert(1)"}},
			{"relative logo URL", createStatusPageRequest{Slug: uniqueSlug(t), Title: "x", LogoURL: "/evil"}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})

	t.Run("success returns a public URL ending in the slug", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		slug := uniqueSlug(t)
		w := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{
			Slug: slug, Title: "Acme status", Description: "all systems", LogoURL: "https://example.com/logo.png",
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[statusPageResponse](t, w)
		if resp.Slug != slug {
			t.Fatalf("want slug %q, got %q", slug, resp.Slug)
		}
		wantSuffix := "/status/" + slug
		if len(resp.PublicURL) < len(wantSuffix) || resp.PublicURL[len(resp.PublicURL)-len(wantSuffix):] != wantSuffix {
			t.Fatalf("want public URL ending in %q, got %q", wantSuffix, resp.PublicURL)
		}
	})

	t.Run("duplicate slug returns 409", func(t *testing.T) {
		// Slugs are globally unique (the public URL is /status/:slug with no
		// per-org namespacing), so this needs two different orgs — Hobby's
		// 1-status-page limit means a single org can't create a second page
		// at all, let alone reach the slug-uniqueness check.
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		slug := uniqueSlug(t)
		firstW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, uA.access, createStatusPageRequest{Slug: slug, Title: "First"})
		if firstW.Code != http.StatusCreated {
			t.Fatalf("setup: want 201, got %d", firstW.Code)
		}

		w := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, uB.access, createStatusPageRequest{Slug: slug, Title: "Second"})
		if w.Code != http.StatusConflict {
			t.Fatalf("want 409, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "slug_taken" {
			t.Fatalf("want code slug_taken, got %q", body["code"])
		}
	})

	t.Run("plan limit enforced at 1 status page on Hobby", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		firstW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "First"})
		if firstW.Code != http.StatusCreated {
			t.Fatalf("setup: want 201, got %d", firstW.Code)
		}

		w := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "Second"})
		if w.Code != http.StatusPaymentRequired {
			t.Fatalf("want 402, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "plan_limit_reached" {
			t.Fatalf("want code plan_limit_reached, got %q", body["code"])
		}
	})
}

func TestGetStatusPage(t *testing.T) {
	authH, monitorH, statusH, pool := testStatusPageHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/status-pages/x", nil)
		w := httptest.NewRecorder()
		statusH.GetStatusPage(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doStatusPageRequest(t, http.MethodGet, statusH.GetStatusPage, u.access, "not-a-uuid", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doStatusPageRequest(t, http.MethodGet, statusH.GetStatusPage, u.access, "00000000-0000-0000-0000-000000000000", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant page is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, uA.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "Org A page"})
		page := decodeBody[statusPageResponse](t, createW)

		w := doStatusPageRequest(t, http.MethodGet, statusH.GetStatusPage, uB.access, page.ID, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 fetching org A's page as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success includes attached monitors", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Attached monitor")
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "x"})
		page := decodeBody[statusPageResponse](t, createW)

		setW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "cron", MonitorID: mon.ID, DisplayName: "Backups", DisplayOrder: 0}},
		})
		if setW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", setW.Code, setW.Body.String())
		}

		w := doStatusPageRequest(t, http.MethodGet, statusH.GetStatusPage, u.access, page.ID, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[statusPageDetailResponse](t, w)
		if len(resp.Monitors) != 1 {
			t.Fatalf("want 1 monitor, got %d", len(resp.Monitors))
		}
		if resp.Monitors[0].DisplayName != "Backups" {
			t.Fatalf("want display name Backups, got %q", resp.Monitors[0].DisplayName)
		}
	})
}

func TestUpdateStatusPage(t *testing.T) {
	authH, _, statusH, pool := testStatusPageHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/status-pages/x", nil)
		w := httptest.NewRecorder()
		statusH.UpdateStatusPage(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("missing title", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "x"})
		page := decodeBody[statusPageResponse](t, createW)

		w := doStatusPageRequest(t, http.MethodPatch, statusH.UpdateStatusPage, u.access, page.ID, updateStatusPageRequest{})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doStatusPageRequest(t, http.MethodPatch, statusH.UpdateStatusPage, u.access, "00000000-0000-0000-0000-000000000000", updateStatusPageRequest{Title: "x"})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant update is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, uA.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "Org A page"})
		page := decodeBody[statusPageResponse](t, createW)

		w := doStatusPageRequest(t, http.MethodPatch, statusH.UpdateStatusPage, uB.access, page.ID, updateStatusPageRequest{Title: "hijacked"})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 updating org A's page as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success applies changes", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "Original"})
		page := decodeBody[statusPageResponse](t, createW)

		w := doStatusPageRequest(t, http.MethodPatch, statusH.UpdateStatusPage, u.access, page.ID, updateStatusPageRequest{
			Title: "Renamed", Description: "new description", LogoURL: "https://example.com/new-logo.png",
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[statusPageResponse](t, w)
		if resp.Title != "Renamed" {
			t.Fatalf("want title Renamed, got %q", resp.Title)
		}
		if resp.Description != "new description" {
			t.Fatalf("want updated description, got %q", resp.Description)
		}
		if resp.Slug != page.Slug {
			t.Fatalf("want slug unchanged (no slug field in update request), got %q vs %q", resp.Slug, page.Slug)
		}
	})
}

func TestDeleteStatusPage(t *testing.T) {
	authH, _, statusH, pool := testStatusPageHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/status-pages/x", nil)
		w := httptest.NewRecorder()
		statusH.DeleteStatusPage(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("cross-tenant delete is a no-op, owner delete succeeds", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, uA.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "Org A page"})
		page := decodeBody[statusPageResponse](t, createW)

		wrongDelete := doStatusPageRequest(t, http.MethodDelete, statusH.DeleteStatusPage, uB.access, page.ID, nil)
		if wrongDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204 (delete is a no-op exec regardless of match), got %d", wrongDelete.Code)
		}
		stillThere := doStatusPageRequest(t, http.MethodGet, statusH.GetStatusPage, uA.access, page.ID, nil)
		if stillThere.Code != http.StatusOK {
			t.Fatalf("want org A's page to survive org B's delete attempt, got %d", stillThere.Code)
		}

		ownerDelete := doStatusPageRequest(t, http.MethodDelete, statusH.DeleteStatusPage, uA.access, page.ID, nil)
		if ownerDelete.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", ownerDelete.Code, ownerDelete.Body.String())
		}
		gone := doStatusPageRequest(t, http.MethodGet, statusH.GetStatusPage, uA.access, page.ID, nil)
		if gone.Code != http.StatusNotFound {
			t.Fatalf("want 404 after delete, got %d", gone.Code)
		}
	})
}

func TestSetStatusPageMonitors(t *testing.T) {
	authH, monitorH, statusH, pool := testStatusPageHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/status-pages/x/monitors", nil)
		w := httptest.NewRecorder()
		statusH.SetStatusPageMonitors(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("page not found", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, "00000000-0000-0000-0000-000000000000", setMonitorsRequest{})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("cross-tenant page is not found (tenant isolation)", func(t *testing.T) {
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, uA.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "Org A page"})
		page := decodeBody[statusPageResponse](t, createW)

		w := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, uB.access, page.ID, setMonitorsRequest{})
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 setting monitors on org A's page as org B, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("validation", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "x")
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "x"})
		page := decodeBody[statusPageResponse](t, createW)

		t.Run("invalid monitor type", func(t *testing.T) {
			w := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
				Monitors: []setMonitorItem{{MonitorType: "carrier-pigeon", MonitorID: mon.ID}},
			})
			if w.Code != http.StatusBadRequest {
				t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
			}
		})

		t.Run("invalid monitor id", func(t *testing.T) {
			w := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
				Monitors: []setMonitorItem{{MonitorType: "cron", MonitorID: "not-a-uuid"}},
			})
			if w.Code != http.StatusBadRequest {
				t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
			}
		})

		t.Run("malformed JSON body", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/api/v1/status-pages/"+page.ID+"/monitors", bytes.NewReader([]byte("not json")))
			req = withURLParam(req, "id", page.ID)
			req.AddCookie(u.access)
			w := httptest.NewRecorder()
			apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(statusH.SetStatusPageMonitors)).ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
			}
		})
	})

	t.Run("success attaches multiple monitor types and fully replaces on a second call", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cron := createCronMonitor(t, monitorH, u.access, "Cron")
		uptime := createUptimeMonitor(t, monitorH, u.access, "Uptime")
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "x"})
		page := decodeBody[statusPageResponse](t, createW)

		firstW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{
				{MonitorType: "cron", MonitorID: cron.ID, DisplayName: "Backups", DisplayOrder: 0},
				{MonitorType: "uptime", MonitorID: uptime.ID, DisplayName: "API", DisplayOrder: 1},
			},
		})
		if firstW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", firstW.Code, firstW.Body.String())
		}
		first := decodeBody[[]statusPageMonitorResponse](t, firstW)
		if len(first) != 2 {
			t.Fatalf("want 2 monitors, got %d", len(first))
		}

		// Replace with just one monitor — the old set must be fully gone, not merged.
		secondW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "uptime", MonitorID: uptime.ID, DisplayName: "API only", DisplayOrder: 0}},
		})
		if secondW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", secondW.Code, secondW.Body.String())
		}
		second := decodeBody[[]statusPageMonitorResponse](t, secondW)
		if len(second) != 1 {
			t.Fatalf("want 1 monitor after replace, got %d", len(second))
		}
		if second[0].DisplayName != "API only" {
			t.Fatalf("want API only, got %q", second[0].DisplayName)
		}

		// An empty list clears all monitors.
		clearW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{})
		if clearW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", clearW.Code, clearW.Body.String())
		}
		cleared := decodeBody[[]statusPageMonitorResponse](t, clearW)
		if len(cleared) != 0 {
			t.Fatalf("want 0 monitors after clearing, got %d", len(cleared))
		}
	})

	t.Run("SECURITY: rejects a monitor belonging to a different org (tenant isolation)", func(t *testing.T) {
		// SetStatusPageMonitors resolves each monitor via resolveMonitorName
		// (shared with maintenance.go), which scopes its lookup by org_id —
		// this confirms ownership, not just that the type/UUID are well
		// formed. Without this, org A could attach org B's monitor ID to org
		// A's own status page, and org B's live monitor status/90-day uptime
		// bar would render on org A's public status page (status_public.go's
		// public lookups are org_id-blind by design, trusting that
		// status_page_monitors rows are always same-org).
		uA := signUpTestUser(t, authH, pool)
		uB := signUpTestUser(t, authH, pool)
		victimMon := createCronMonitor(t, monitorH, uB.access, "Org B's private monitor")
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, uA.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "Org A page"})
		page := decodeBody[statusPageResponse](t, createW)

		w := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, uA.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "cron", MonitorID: victimMon.ID, DisplayName: "Borrowed", DisplayOrder: 0}},
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400 rejecting a foreign org's monitor id, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("an empty display name falls back to the monitor's real name, not its UUID", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Real Monitor Name")
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: uniqueSlug(t), Title: "x"})
		page := decodeBody[statusPageResponse](t, createW)

		w := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "cron", MonitorID: mon.ID, DisplayName: "  ", DisplayOrder: 0}},
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		attached := decodeBody[[]statusPageMonitorResponse](t, w)
		if len(attached) != 1 || attached[0].DisplayName != "Real Monitor Name" {
			t.Fatalf("want display name to default to the monitor's real name, got %+v", attached)
		}
	})
}
