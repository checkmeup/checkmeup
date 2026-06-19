package handler

// Integration tests for the public, unauthenticated /status/:slug page
// (ADR-017: server-rendered Go html/template, no JS dependency). Same
// conventions as the other *_test.go files: real Postgres (ADR-010),
// package handler so unexported types/helpers are reused directly. Since
// this handler renders HTML rather than JSON, assertions are substring
// checks against the rendered body rather than JSON decoding.
//
// The pure helper functions (monitorStatusDisplay, computeOverallStatus,
// build90DayBar) are deterministic and tested directly with table tests,
// no DB needed.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

func testStatusPublicHandler(t *testing.T) (*AuthHandler, *MonitorHandler, *StatusPageHandler, *MaintenanceHandler, *StatusPublicHandler, *pgxpool.Pool) {
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
	maintH := NewMaintenanceHandler(pool)
	publicH := NewStatusPublicHandler(pool)
	return authH, monitorH, statusH, maintH, publicH, pool
}

func doPublicStatusPage(publicH *StatusPublicHandler, slug string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/status/"+slug, nil)
	req = withURLParam(req, "slug", slug)
	w := httptest.NewRecorder()
	publicH.ServeHTTP(w, req)
	return w
}

func TestStatusPublicServeHTTP(t *testing.T) {
	authH, monitorH, statusH, maintH, publicH, pool := testStatusPublicHandler(t)

	t.Run("unknown slug returns 404", func(t *testing.T) {
		w := doPublicStatusPage(publicH, "no-such-slug-"+uuid.NewString())
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("a page with no monitors renders the empty state", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{
			Slug: slug, Title: "Empty Co", Description: "nothing here yet",
		})
		if createW.Code != http.StatusCreated {
			t.Fatalf("setup: want 201, got %d", createW.Code)
		}

		w := doPublicStatusPage(publicH, slug)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
			t.Fatalf("want text/html content type, got %q", ct)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Empty Co") {
			t.Fatal("want the page title rendered")
		}
		if !strings.Contains(body, "nothing here yet") {
			t.Fatal("want the page description rendered")
		}
		if !strings.Contains(body, "No monitors on this page yet.") {
			t.Fatal("want the empty-monitors message")
		}
		if !strings.Contains(body, "All systems operational") {
			t.Fatal("want an empty page to report all systems operational")
		}
	})

	t.Run("renders monitor statuses and aggregates overall status", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cron := createCronMonitor(t, monitorH, u.access, "Cron job")
		uptimeMon := createUptimeMonitor(t, monitorH, u.access, "API")
		sslMon := createSSLMonitor(t, monitorH, u.access, "Cert")

		// Cron: up. Uptime: down (drives the overall status to "Major outage").
		// SSL: expiring soon, with a real expiry date.
		mustExec(t, pool, "UPDATE cron_monitors SET status = 'up' WHERE id = $1", cron.ID)
		mustExec(t, pool, "UPDATE uptime_monitors SET status = 'down' WHERE id = $1", uptimeMon.ID)
		expiresAt := time.Now().Add(10 * 24 * time.Hour)
		mustExec(t, pool, "UPDATE ssl_monitors SET status = 'expiring_soon', expires_at = $2 WHERE id = $1", sslMon.ID, expiresAt)

		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Multi Co"})
		page := decodeBody[statusPageResponse](t, createW)
		setW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{
				{MonitorType: "cron", MonitorID: cron.ID, DisplayName: "Backups", DisplayOrder: 0},
				{MonitorType: "uptime", MonitorID: uptimeMon.ID, DisplayName: "API", DisplayOrder: 1},
				{MonitorType: "ssl", MonitorID: sslMon.ID, DisplayName: "Certificate", DisplayOrder: 2},
			},
		})
		if setW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", setW.Code, setW.Body.String())
		}

		w := doPublicStatusPage(publicH, slug)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		body := w.Body.String()

		for _, want := range []string{"Backups", "API", "Certificate"} {
			if !strings.Contains(body, want) {
				t.Fatalf("want display name %q in body", want)
			}
		}
		if !strings.Contains(body, "Operational") {
			t.Fatal("want cron's Operational label")
		}
		if !strings.Contains(body, "Down") {
			t.Fatal("want uptime's Down label")
		}
		if !strings.Contains(body, "Expiring soon") {
			t.Fatal("want ssl's Expiring soon label")
		}
		if !strings.Contains(body, "Major outage") {
			t.Fatal("want overall status Major outage (a red row is present)")
		}
		if !strings.Contains(body, expiresAt.Format("2006-01-02")) {
			t.Fatal("want the SSL certificate expiry date rendered")
		}
	})

	t.Run("an active maintenance window overrides the monitor's real status", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Under maintenance monitor")
		mustExec(t, pool, "UPDATE cron_monitors SET status = 'down' WHERE id = $1", mon.ID)

		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Maintained Co"})
		page := decodeBody[statusPageResponse](t, createW)
		setW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "cron", MonitorID: mon.ID, DisplayName: "Maintained", DisplayOrder: 0}},
		})
		if setW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", setW.Code, setW.Body.String())
		}

		starts := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
		windowW := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, u.access, maintenanceWindowRequest{
			Title:    "Scheduled work",
			Message:  "We are upgrading the database.",
			StartsAt: starts,
			Monitors: []maintenanceMonitorInput{{MonitorType: "cron", MonitorID: mon.ID}},
		})
		if windowW.Code != http.StatusCreated {
			t.Fatalf("setup: want 201, got %d: %s", windowW.Code, windowW.Body.String())
		}

		w := doPublicStatusPage(publicH, slug)
		body := w.Body.String()
		if !strings.Contains(body, "Under maintenance") {
			t.Fatal("want the maintenance override label")
		}
		if !strings.Contains(body, "We are upgrading the database.") {
			t.Fatal("want the maintenance message rendered")
		}
		if strings.Contains(body, ">Down<") {
			t.Fatal("want the real Down status hidden behind the maintenance override")
		}
	})
}

func TestMonitorStatusDisplay(t *testing.T) {
	cases := []struct {
		status    string
		wantLabel string
		wantColor string
	}{
		{"up", "Operational", statusColorGreen},
		{"down", "Down", statusColorRed},
		{"paused", "Paused", statusColorGray},
		{"waiting", "Checking…", statusColorGray},
		{"something-unexpected", "Checking…", statusColorGray},
	}
	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			label, color := monitorStatusDisplay(tc.status)
			if label != tc.wantLabel || color != tc.wantColor {
				t.Fatalf("monitorStatusDisplay(%q) = (%q, %q), want (%q, %q)", tc.status, label, color, tc.wantLabel, tc.wantColor)
			}
		})
	}
}

func TestComputeOverallStatus(t *testing.T) {
	cases := []struct {
		name      string
		rows      []publicMonitorRow
		wantLabel string
		wantColor string
	}{
		{"no monitors", nil, "All systems operational", statusColorGreen},
		{"all green", []publicMonitorRow{{StatusColor: statusColorGreen}, {StatusColor: statusColorGreen}}, "All systems operational", statusColorGreen},
		{"one amber", []publicMonitorRow{{StatusColor: statusColorGreen}, {StatusColor: statusColorAmber}}, "Partial outage", statusColorAmber},
		{"one red", []publicMonitorRow{{StatusColor: statusColorGreen}, {StatusColor: statusColorRed}}, "Major outage", statusColorRed},
		{"red wins over amber", []publicMonitorRow{{StatusColor: statusColorAmber}, {StatusColor: statusColorRed}}, "Major outage", statusColorRed},
		{"gray (unknown/paused) alone doesn't trigger an outage", []publicMonitorRow{{StatusColor: statusColorGray}}, "All systems operational", statusColorGreen},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			label, color := computeOverallStatus(tc.rows)
			if label != tc.wantLabel || color != tc.wantColor {
				t.Fatalf("computeOverallStatus(%v) = (%q, %q), want (%q, %q)", tc.rows, label, color, tc.wantLabel, tc.wantColor)
			}
		})
	}
}

func TestBuild90DayBar(t *testing.T) {
	t.Run("no data renders all bars as the light placeholder color", func(t *testing.T) {
		bar := build90DayBar(map[string]bool{}, false)
		if len(bar) != 90 {
			t.Fatalf("want 90 segments, got %d", len(bar))
		}
		for i, seg := range bar {
			if seg.Color != statusColorLight {
				t.Fatalf("segment %d: want light placeholder color, got %q", i, seg.Color)
			}
		}
	})

	t.Run("has data: down days are red, others green, today is the last segment", func(t *testing.T) {
		today := time.Now().UTC().Truncate(24 * time.Hour)
		downKey := today.AddDate(0, 0, -5).Format("2006-01-02")
		bar := build90DayBar(map[string]bool{downKey: true}, true)
		if len(bar) != 90 {
			t.Fatalf("want 90 segments, got %d", len(bar))
		}
		// Segment 89 (last) is today; segment 84 is today-5.
		if bar[84].Color != statusColorRed {
			t.Fatalf("want the seeded down day red, got %q", bar[84].Color)
		}
		if bar[89].Color != statusColorGreen {
			t.Fatalf("want today green (no data seeded for it), got %q", bar[89].Color)
		}
		if bar[89].Label != today.Format("Jan 2") {
			t.Fatalf("want today's label %q, got %q", today.Format("Jan 2"), bar[89].Label)
		}
	})
}

func mustExec(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}
