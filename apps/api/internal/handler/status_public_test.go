package handler

// Integration tests for the public, unauthenticated /status/:slug page
// (ADR status-pages: server-rendered Go html/template, no JS dependency). Same
// conventions as the other *_test.go files: real Postgres (ADR testing),
// package handler so unexported types/helpers are reused directly. Since
// this handler renders HTML rather than JSON, assertions are substring
// checks against the rendered body rather than JSON decoding.
//
// The pure helper functions (monitorStatusDisplay, computeOverallStatus,
// build90DayBar) are deterministic and tested directly with table tests,
// no DB needed.

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
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
		if !strings.Contains(body, "Powered by") {
			t.Fatal("want the branding footer by default (hide_branding defaults false)")
		}
		if !strings.Contains(body, `<a href="https://checkmeup.net/?utm_source=status_page&utm_medium=referral&utm_campaign=powered_by">Get a free status page like this</a>`) {
			t.Fatal("want the branding footer to offer a free status page, tagged for GA4")
		}
	})

	t.Run("ADR status-pages: grid layout renders the sidebar/content-grid wrapper with monitors and incidents", func(t *testing.T) {
		incidentH := NewIncidentHandler(pool)
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "API")
		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Grid Co", Layout: "grid"})
		page := decodeBody[statusPageResponse](t, createW)
		setW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "uptime", MonitorID: mon.ID, DisplayName: "API", DisplayOrder: 0}},
		})
		if setW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", setW.Code, setW.Body.String())
		}
		createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "Grid incident", Message: "Looking into it", Severity: "major",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})

		w := doPublicStatusPage(publicH, slug)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, `class="page layout-grid"`) {
			t.Fatal("want the layout-grid page class")
		}
		if !strings.Contains(body, `class="content-grid"`) {
			t.Fatal("want the content-grid wrapper")
		}
		if !strings.Contains(body, `class="sidebar"`) {
			t.Fatal("want the sidebar wrapper")
		}
		if !strings.Contains(body, "API") {
			t.Fatal("want the monitor rendered inside the grid")
		}
		if !strings.Contains(body, "Grid incident") {
			t.Fatal("want the active incident rendered inside the sidebar")
		}
	})

	t.Run("ADR status-pages: hide_branding suppresses the footer's branding and FAQ/Terms/Privacy links", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mustExec(t, pool, "UPDATE orgs SET plan = 'solo' WHERE id = $1", u.resp.OrgID)
		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Branded Co"})
		page := decodeBody[statusPageResponse](t, createW)

		updateW := doStatusPageRequest(t, http.MethodPatch, statusH.UpdateStatusPage, u.access, page.ID, updateStatusPageRequest{
			Title: "Branded Co", HideBranding: true,
		})
		if updateW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", updateW.Code, updateW.Body.String())
		}

		w := doPublicStatusPage(publicH, slug)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		body := w.Body.String()
		if strings.Contains(body, "Powered by") {
			t.Fatal("want the branding line hidden")
		}
		if strings.Contains(body, "checkmeup.net/faq") {
			t.Fatal("want the FAQ/Terms/Privacy links hidden")
		}
		if !strings.Contains(body, "Last updated") {
			t.Fatal("want the footer's Last updated line to stay, only branding hidden")
		}
	})

	t.Run("active incidents render their full update timeline, newest first, batched across incidents", func(t *testing.T) {
		incidentH := NewIncidentHandler(pool)
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "API")
		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Incident Co"})
		page := decodeBody[statusPageResponse](t, createW)
		setW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "uptime", MonitorID: mon.ID, DisplayName: "API", DisplayOrder: 0}},
		})
		if setW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", setW.Code, setW.Body.String())
		}

		created := createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "Elevated latency", Message: "Investigating reports", Severity: "major",
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})
		followUp := doIncidentRequest(t, http.MethodPost, incidentH.PostIncidentUpdate, u.access, "/api/v1/incidents/"+created.ID+"/updates", map[string]string{"id": created.ID}, postUpdateRequest{
			Message: "Found the cause", Status: "identified",
		})
		if followUp.Code != http.StatusCreated {
			t.Fatalf("setup: want 201, got %d: %s", followUp.Code, followUp.Body.String())
		}

		w := doPublicStatusPage(publicH, slug)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Elevated latency") {
			t.Fatal("want the incident title rendered")
		}
		if !strings.Contains(body, "Investigating reports") {
			t.Fatal("want the initial update message rendered")
		}
		if !strings.Contains(body, "Found the cause") {
			t.Fatal("want the follow-up update message rendered")
		}
		idxInitial := strings.Index(body, "Investigating reports")
		idxFollowUp := strings.Index(body, "Found the cause")
		if idxFollowUp > idxInitial {
			t.Fatal("want the newest update (Found the cause) rendered before the older one (Investigating reports)")
		}
	})

	t.Run("renders monitor statuses and aggregates overall status", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cron := createCronMonitor(t, monitorH, u.access, "Cron job")
		uptimeMon := createUptimeMonitor(t, monitorH, u.access, "API")
		sslMon := createSSLMonitor(t, monitorH, u.access, "Cert")
		domainMon := createDomainMonitor(t, monitorH, u.access, "Domain")
		portMon := createPortMonitor(t, monitorH, u.access, "Port")
		dnsMon := createDNSMonitor(t, monitorH, u.access, "DNS")

		// Cron: up. Uptime: down (drives the overall status to "Major outage").
		// SSL and domain: expiring soon, with a real expiry date each. Port
		// and DNS: up.
		mustExec(t, pool, "UPDATE cron_monitors SET status = 'up' WHERE id = $1", cron.ID)
		mustExec(t, pool, "UPDATE uptime_monitors SET status = 'down' WHERE id = $1", uptimeMon.ID)
		expiresAt := time.Now().Add(10 * 24 * time.Hour)
		mustExec(t, pool, "UPDATE ssl_monitors SET status = 'expiring_soon', expires_at = $2 WHERE id = $1", sslMon.ID, expiresAt)
		domainExpiresAt := time.Now().Add(20 * 24 * time.Hour)
		mustExec(t, pool, "UPDATE domain_monitors SET status = 'expiring_soon', expires_at = $2 WHERE id = $1", domainMon.ID, domainExpiresAt)
		mustExec(t, pool, "UPDATE port_monitors SET status = 'up' WHERE id = $1", portMon.ID)
		mustExec(t, pool, "UPDATE dns_monitors SET status = 'up' WHERE id = $1", dnsMon.ID)

		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Multi Co"})
		page := decodeBody[statusPageResponse](t, createW)
		setW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{
				{MonitorType: "cron", MonitorID: cron.ID, DisplayName: "Backups", DisplayOrder: 0},
				{MonitorType: "uptime", MonitorID: uptimeMon.ID, DisplayName: "API", DisplayOrder: 1},
				{MonitorType: "ssl", MonitorID: sslMon.ID, DisplayName: "Certificate", DisplayOrder: 2},
				{MonitorType: "domain", MonitorID: domainMon.ID, DisplayName: "Domain expiry", DisplayOrder: 3},
				{MonitorType: "port", MonitorID: portMon.ID, DisplayName: "Port check", DisplayOrder: 4},
				{MonitorType: "dns", MonitorID: dnsMon.ID, DisplayName: "DNS record", DisplayOrder: 5},
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

		for _, want := range []string{"Backups", "API", "Certificate", "Domain expiry", "Port check", "DNS record"} {
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
			t.Fatal("want ssl's and domain's Expiring soon label")
		}
		if !strings.Contains(body, "Major outage") {
			t.Fatal("want overall status Major outage (a red row is present)")
		}
		if !strings.Contains(body, expiresAt.Format("2006-01-02")) {
			t.Fatal("want the SSL certificate expiry date rendered")
		}
		if !strings.Contains(body, "Domain expires "+domainExpiresAt.Format("2006-01-02")) {
			t.Fatal("want the domain expiry date rendered")
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
	t.Run("PRD legal-pages: the footer links to Terms and Privacy", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Footer Co"})
		if createW.Code != http.StatusCreated {
			t.Fatalf("setup: want 201, got %d", createW.Code)
		}

		body := doPublicStatusPage(publicH, slug).Body.String()
		footer := between(body, `<footer class="footer">`, "</footer>")
		if !strings.Contains(footer, `<a href="https://checkmeup.net/terms">Terms</a>`) {
			t.Fatal("want a Terms link in the footer")
		}
		if !strings.Contains(footer, `<a href="https://checkmeup.net/privacy">Privacy</a>`) {
			t.Fatal("want a Privacy link in the footer")
		}
	})

	t.Run("resolved incidents appear in a paginated history", func(t *testing.T) {
		incidentH := NewIncidentHandler(pool)
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "API")
		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "History Co"})
		page := decodeBody[statusPageResponse](t, createW)
		setW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "uptime", MonitorID: mon.ID, DisplayName: "API", DisplayOrder: 0}},
		})
		if setW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", setW.Code, setW.Body.String())
		}

		// Eleven resolved incidents, "Past incident 01" resolved most recently.
		const total = 11
		for i := 1; i <= total; i++ {
			created := createIncident(t, incidentH, u.access, createIncidentRequest{
				Title: fmt.Sprintf("Past incident %02d", i), Message: "Looking into it", Severity: "minor",
				Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
			})
			resolveW := doIncidentRequest(t, http.MethodPost, incidentH.PostIncidentUpdate, u.access, "/api/v1/incidents/"+created.ID+"/updates", map[string]string{"id": created.ID}, postUpdateRequest{
				Message: fmt.Sprintf("Fixed %02d", i), Status: "resolved",
			})
			if resolveW.Code != http.StatusCreated {
				t.Fatalf("setup: want 201, got %d: %s", resolveW.Code, resolveW.Body.String())
			}
			mustExec(t, pool, "UPDATE status_page_incidents SET resolved_at = NOW() - make_interval(mins => $2) WHERE id = $1", created.ID, i)
		}

		pageBody := func(query string) string {
			req := httptest.NewRequest(http.MethodGet, "/status/"+slug+query, nil)
			req = withURLParam(req, "slug", slug)
			w := httptest.NewRecorder()
			publicH.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s: want 200, got %d", query, w.Code)
			}
			return w.Body.String()
		}

		first := pageBody("")
		if !strings.Contains(first, "Past incidents") || !strings.Contains(first, "Past incident 01") || !strings.Contains(first, "Fixed 01") {
			t.Fatal("want the most recently resolved incident in the history on the first page")
		}
		if strings.Contains(first, "Past incident 11") {
			t.Fatal("want the oldest incident pushed off the first page")
		}
		if !strings.Contains(first, `href="?incidents_page=2"`) || strings.Contains(first, "Newer") {
			t.Fatal("want the first page to link to older incidents only")
		}

		second := pageBody("?incidents_page=2")
		if !strings.Contains(second, "Past incident 11") {
			t.Fatal("want the oldest incident on the second page")
		}
		if strings.Contains(second, "Past incident 01") {
			t.Fatal("want the newest incident left on the first page")
		}
		if !strings.Contains(second, `href="?incidents_page=1"`) {
			t.Fatal("want the second page to link back to newer incidents")
		}
	})

	t.Run("maintenance and an incident on the same monitor show as distinct states", func(t *testing.T) {
		incidentH := NewIncidentHandler(pool)
		u := signUpTestUser(t, authH, pool)
		mon := createUptimeMonitor(t, monitorH, u.access, "API")
		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Both Co"})
		page := decodeBody[statusPageResponse](t, createW)
		setW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "uptime", MonitorID: mon.ID, DisplayName: "API", DisplayOrder: 0}},
		})
		if setW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", setW.Code, setW.Body.String())
		}
		windowW := doAuthed(t, http.MethodPost, maintH.CreateMaintenanceWindow, u.access, maintenanceWindowRequest{
			Title:    "Database upgrade",
			StartsAt: time.Now().Add(-time.Hour).UTC().Format(time.RFC3339),
			Monitors: []maintenanceMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})
		if windowW.Code != http.StatusCreated {
			t.Fatalf("setup: want 201, got %d: %s", windowW.Code, windowW.Body.String())
		}
		createIncident(t, incidentH, u.access, createIncidentRequest{
			Title: "Login errors", Message: "Users cannot sign in", Severity: "major", ConfirmOverlap: true,
			Monitors: []incidentMonitorInput{{MonitorType: "uptime", MonitorID: mon.ID}},
		})

		body := doPublicStatusPage(publicH, slug).Body.String()
		if !strings.Contains(between(body, `<span class="monitor-name">API</span>`, `class="bar"`), "Under maintenance") {
			t.Fatal("want the monitor shown as Under maintenance")
		}
		incidentCard := between(body, `<section class="incident `, "</section>")
		for _, want := range []string{"Login errors", "Investigating", "Users cannot sign in", "Affecting: API"} {
			if !strings.Contains(incidentCard, want) {
				t.Fatalf("want the incident card to show %q", want)
			}
		}
	})
}

// between returns the part of s from the first start up to the next end, or
// "" when either is missing.
func between(s, start, end string) string {
	i := strings.Index(s, start)
	if i < 0 {
		return ""
	}
	j := strings.Index(s[i:], end)
	if j < 0 {
		return ""
	}
	return s[i : i+j]
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
		incidents []publicIncidentRow
		wantLabel string
		wantColor string
	}{
		{"no monitors", nil, nil, "All systems operational", statusColorGreen},
		{"all green", []publicMonitorRow{{StatusColor: statusColorGreen}, {StatusColor: statusColorGreen}}, nil, "All systems operational", statusColorGreen},
		{"one amber", []publicMonitorRow{{StatusColor: statusColorGreen}, {StatusColor: statusColorAmber}}, nil, "Partial outage", statusColorAmber},
		{"one red", []publicMonitorRow{{StatusColor: statusColorGreen}, {StatusColor: statusColorRed}}, nil, "Major outage", statusColorRed},
		{"red wins over amber", []publicMonitorRow{{StatusColor: statusColorAmber}, {StatusColor: statusColorRed}}, nil, "Major outage", statusColorRed},
		{"gray (unknown/paused) alone doesn't trigger an outage", []publicMonitorRow{{StatusColor: statusColorGray}}, nil, "All systems operational", statusColorGreen},
		{"minor incident alone doesn't escalate", nil, []publicIncidentRow{{SeverityColor: statusColorBlue}}, "All systems operational", statusColorGreen},
		{"major incident escalates even with all-green monitors", []publicMonitorRow{{StatusColor: statusColorGreen}}, []publicIncidentRow{{SeverityColor: statusColorAmber}}, "Partial outage", statusColorAmber},
		{"critical incident escalates even with all-green monitors", []publicMonitorRow{{StatusColor: statusColorGreen}}, []publicIncidentRow{{SeverityColor: statusColorRed}}, "Major outage", statusColorRed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			label, color := computeOverallStatus(tc.rows, tc.incidents)
			if label != tc.wantLabel || color != tc.wantColor {
				t.Fatalf("computeOverallStatus(%v, %v) = (%q, %q), want (%q, %q)", tc.rows, tc.incidents, label, color, tc.wantLabel, tc.wantColor)
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

// ─── badges (PRD status-badges) ──────────────────────────────────────────────────────────

func doPageBadge(publicH *StatusPublicHandler, slug string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/status/"+slug+"/badge.svg", nil)
	req = withURLParam(req, "slug", slug)
	w := httptest.NewRecorder()
	publicH.ServePageBadge(w, req)
	return w
}

func doMonitorBadge(publicH *StatusPublicHandler, slug, monitorID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/status/"+slug+"/badge/"+monitorID+".svg", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", slug)
	rctx.URLParams.Add("monitor_id", monitorID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	publicH.ServeMonitorBadge(w, req)
	return w
}

func TestStatusPublicServePageBadge(t *testing.T) {
	authH, monitorH, statusH, _, publicH, pool := testStatusPublicHandler(t)

	t.Run("unknown slug returns 404", func(t *testing.T) {
		w := doPageBadge(publicH, "no-such-slug-"+uuid.NewString())
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("all-green page renders an operational badge", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cron := createCronMonitor(t, monitorH, u.access, "Cron job")
		mustExec(t, pool, "UPDATE cron_monitors SET status = 'up' WHERE id = $1", cron.ID)

		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Badge Co"})
		page := decodeBody[statusPageResponse](t, createW)
		setW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "cron", MonitorID: cron.ID, DisplayName: "Backups", DisplayOrder: 0}},
		})
		if setW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", setW.Code, setW.Body.String())
		}

		w := doPageBadge(publicH, slug)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		if ct := w.Header().Get("Content-Type"); ct != "image/svg+xml; charset=utf-8" {
			t.Fatalf("want SVG content type, got %q", ct)
		}
		if cc := w.Header().Get("Cache-Control"); cc != "max-age=60" {
			t.Fatalf("want short cache TTL (PRD status-badges), got %q", cc)
		}
		body := w.Body.String()
		if !strings.Contains(body, "operational") {
			t.Fatalf("want badge text 'operational', got %s", body)
		}
		if !strings.Contains(body, badgeFillUp) {
			t.Fatalf("want the green fill color, got %s", body)
		}
	})

	t.Run("a down monitor renders an outage badge", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		uptimeMon := createUptimeMonitor(t, monitorH, u.access, "API")
		mustExec(t, pool, "UPDATE uptime_monitors SET status = 'down' WHERE id = $1", uptimeMon.ID)

		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Down Co"})
		page := decodeBody[statusPageResponse](t, createW)
		setW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "uptime", MonitorID: uptimeMon.ID, DisplayName: "API", DisplayOrder: 0}},
		})
		if setW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", setW.Code, setW.Body.String())
		}

		w := doPageBadge(publicH, slug)
		body := w.Body.String()
		if !strings.Contains(body, "outage") {
			t.Fatalf("want badge text 'outage', got %s", body)
		}
	})
}

func TestStatusPublicServeMonitorBadge(t *testing.T) {
	authH, monitorH, statusH, _, publicH, pool := testStatusPublicHandler(t)

	t.Run("unknown slug returns 404", func(t *testing.T) {
		w := doMonitorBadge(publicH, "no-such-slug-"+uuid.NewString(), uuid.NewString())
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("invalid monitor id returns 404", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Bad Id Co"})
		page := decodeBody[statusPageResponse](t, createW)
		_ = page
		w := doMonitorBadge(publicH, slug, "not-a-uuid")
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("monitor not attached to the page returns 404 (no leaking status outside the page)", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cron := createCronMonitor(t, monitorH, u.access, "Off-page cron")
		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Sparse Co"})
		page := decodeBody[statusPageResponse](t, createW)
		_ = page

		w := doMonitorBadge(publicH, slug, cron.ID)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("attached monitor renders its own status badge", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		sslMon := createSSLMonitor(t, monitorH, u.access, "Cert")
		mustExec(t, pool, "UPDATE ssl_monitors SET status = 'expiring_soon' WHERE id = $1", sslMon.ID)

		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Cert Co"})
		page := decodeBody[statusPageResponse](t, createW)
		setW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "ssl", MonitorID: sslMon.ID, DisplayName: "Certificate", DisplayOrder: 0}},
		})
		if setW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", setW.Code, setW.Body.String())
		}

		w := doMonitorBadge(publicH, slug, sslMon.ID)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "degraded") {
			t.Fatalf("want badge text 'degraded' for an expiring-soon cert, got %s", body)
		}
		if !strings.Contains(body, "Certificate") {
			t.Fatalf("want the monitor's display name as the badge label, got %s", body)
		}
	})
}

func TestBadgeStatusWord(t *testing.T) {
	cases := []struct {
		color string
		want  string
	}{
		{statusColorGreen, "operational"},
		{statusColorAmber, "degraded"},
		{statusColorRed, "outage"},
		{statusColorGray, "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if got := badgeStatusWord(tc.color); got != tc.want {
				t.Fatalf("badgeStatusWord(%q) = %q, want %q", tc.color, got, tc.want)
			}
		})
	}
}

func TestRenderBadgeSVG(t *testing.T) {
	svg := renderBadgeSVG("status", "operational", statusColorGreen)
	if !strings.Contains(svg, "<svg") || !strings.Contains(svg, "</svg>") {
		t.Fatalf("want a well-formed SVG document, got %s", svg)
	}
	if !strings.Contains(svg, "status") || !strings.Contains(svg, "operational") {
		t.Fatalf("want label and value text present, got %s", svg)
	}
	if !strings.Contains(svg, badgeFillUp) {
		t.Fatalf("want the up fill color present, got %s", svg)
	}

	t.Run("escapes text that could break the SVG markup", func(t *testing.T) {
		svg := renderBadgeSVG(`<script>&"'`, "operational", statusColorGreen)
		if strings.Contains(svg, "<script>") {
			t.Fatalf("want the label HTML-escaped, got %s", svg)
		}
	})
}

func TestBadgeFill_UsesTheLightThemeStatusColors(t *testing.T) {
	cases := []struct{ color, want string }{
		{statusColorGreen, badgeFillUp},
		{statusColorAmber, badgeFillDegraded},
		{statusColorRed, badgeFillDown},
		{statusColorGray, badgeFillUnknown},
		{statusColorLight, badgeFillUnknown},
	}
	for _, tc := range cases {
		if got := badgeFill(tc.color); got != tc.want {
			t.Fatalf("badgeFill(%q) = %q, want %q", tc.color, got, tc.want)
		}
	}
}

func mustExec(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

func TestStatusTone_MapsEachStatusColorToItsTone(t *testing.T) {
	cases := []struct{ color, want string }{
		{statusColorGreen, "up"},
		{statusColorRed, "down"},
		{statusColorAmber, "warn"},
		{statusColorLight, "none"},
		{statusColorGray, "paused"},
		{statusColorBlue, "paused"},
	}
	for _, c := range cases {
		if got := statusTone(c.color); got != c.want {
			t.Fatalf("statusTone(%s) = %q, want %q", c.color, got, c.want)
		}
	}
}

func TestSummarizeDays_TakesEachDaysWorstTone(t *testing.T) {
	bar := func(colors ...string) []publicBar {
		b := make([]publicBar, 90)
		for i := range b {
			b[i] = publicBar{Color: statusColorGreen, Label: "day"}
		}
		for i, c := range colors {
			b[i] = publicBar{Color: c, Label: "day"}
		}
		return b
	}
	rows := []publicMonitorRow{
		{Bar: bar(statusColorRed, statusColorGreen, statusColorLight)},
		{Bar: bar(statusColorGreen, statusColorAmber, statusColorLight)},
	}

	days := summarizeDays(rows)
	if len(days) != 90 {
		t.Fatalf("want 90 days, got %d", len(days))
	}
	for i, want := range []string{"down", "warn", "none", "up"} {
		if days[i].Tone != want {
			t.Fatalf("day %d: tone = %q, want %q", i, days[i].Tone, want)
		}
	}
	if days[0].Date != "day" {
		t.Fatalf("want the bar's date label, got %q", days[0].Date)
	}
	if got := summarizeDays(nil); len(got) != 90 || got[0].Tone != "none" {
		t.Fatalf("a page with no monitors: want 90 empty days, got %+v", got[:1])
	}
}

// SSL and domain monitors keep no daily history, so their bar repeats today's
// status on every day; only today is real.
func TestSummarizeDays_CountsSSLAndDomainMonitorsOnlyToday(t *testing.T) {
	bar := func(color string) []publicBar {
		b := make([]publicBar, 90)
		for i := range b {
			b[i] = publicBar{Color: color, Label: "day"}
		}
		return b
	}
	rows := []publicMonitorRow{
		{Type: "uptime", Bar: bar(statusColorGreen)},
		{Type: "ssl", Bar: bar(statusColorRed)},
		{Type: "domain", Bar: bar(statusColorAmber)},
	}

	days := summarizeDays(rows)
	if days[0].Tone != "up" || days[88].Tone != "up" {
		t.Fatalf("past days: want up from the uptime monitor alone, got %q / %q", days[0].Tone, days[88].Tone)
	}
	if days[89].Tone != "down" {
		t.Fatalf("today: want down from the failing SSL monitor, got %q", days[89].Tone)
	}
	if got := summarizeDays(rows[1:2]); got[0].Tone != "none" || got[0].Date != "day" {
		t.Fatalf("a page with only an SSL monitor: want past days dated but without data, got %+v", got[0])
	}
}

func TestSummarizeDay_UptimeAndLatency(t *testing.T) {
	seg := func(checks, down int64, avgMs int32) []publicBar {
		return []publicBar{{Color: statusColorGreen, Label: "Oct 14", Checks: checks, Down: down, AvgMs: avgMs}}
	}

	t.Run("pools checks across monitors and weights latency by uptime checks", func(t *testing.T) {
		rows := []publicMonitorRow{
			{Bar: seg(1000, 1, 100)}, // uptime
			{Bar: seg(3000, 0, 300)}, // uptime
			{Bar: seg(1000, 4, 0)},   // port: counts toward uptime %, not latency
		}
		day := summarizeDay(rows, 0)
		if day.Uptime == nil || *day.Uptime != 99.9 {
			t.Fatalf("want 99.9%% (4995 of 5000 checks passed), got %v", day.Uptime)
		}
		if day.LatencyMs == nil || *day.LatencyMs != 250 {
			t.Fatalf("want 250 ms ((1000×100 + 3000×300) / 4000), got %v", day.LatencyMs)
		}
	})

	t.Run("leaves both out on a day no check ran", func(t *testing.T) {
		day := summarizeDay([]publicMonitorRow{{Bar: seg(0, 0, 0)}}, 0)
		if day.Uptime != nil || day.LatencyMs != nil {
			t.Fatalf("want no uptime or latency, got %v / %v", day.Uptime, day.LatencyMs)
		}
	})
}

func TestBarWithStats(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	downKey := today.AddDate(0, 0, -1).Format("2006-01-02")
	upKey := today.Format("2006-01-02")
	bar := barWithStats(map[string]dayCheckStats{
		downKey: {checks: 1440, down: 2, avgMs: 180},
		upKey:   {checks: 720, avgMs: 120},
	})

	if bar[88].Color != statusColorRed || bar[89].Color != statusColorGreen {
		t.Fatalf("want yesterday red and today green, got %q / %q", bar[88].Color, bar[89].Color)
	}
	if bar[88].Checks != 1440 || bar[88].Down != 2 || bar[88].AvgMs != 180 {
		t.Fatalf("want yesterday's stats on its segment, got %+v", bar[88])
	}
	if bar[89].Date != upKey || bar[89].Checks != 720 {
		t.Fatalf("want today's stats on the last segment, got %+v", bar[89])
	}
	if bar[0].Checks != 0 {
		t.Fatalf("a day without stats carries none, got %+v", bar[0])
	}
	if empty := barWithStats(map[string]dayCheckStats{}); empty[0].Color != statusColorLight {
		t.Fatalf("no stats at all: want the no-data color, got %q", empty[0].Color)
	}
}

func TestStatusPublicServeSummary(t *testing.T) {
	authH, monitorH, statusH, _, publicH, pool := testStatusPublicHandler(t)
	doSummary := func(slug string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/status/"+slug+"/summary.json", nil)
		req = withURLParam(req, "slug", slug)
		w := httptest.NewRecorder()
		publicH.ServeSummary(w, req)
		return w
	}

	t.Run("unknown slug returns 404", func(t *testing.T) {
		if w := doSummary("no-such-slug-" + uuid.NewString()); w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("returns the overall status and a 90-day strip", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		cron := createCronMonitor(t, monitorH, u.access, "Cron job")
		mustExec(t, pool, "UPDATE cron_monitors SET status = 'up' WHERE id = $1", cron.ID)
		slug := uniqueSlug(t)
		createW := doAuthed(t, http.MethodPost, statusH.CreateStatusPage, u.access, createStatusPageRequest{Slug: slug, Title: "Summary Co"})
		page := decodeBody[statusPageResponse](t, createW)
		setW := doStatusPageRequest(t, http.MethodPut, statusH.SetStatusPageMonitors, u.access, page.ID, setMonitorsRequest{
			Monitors: []setMonitorItem{{MonitorType: "cron", MonitorID: cron.ID, DisplayName: "Backups", DisplayOrder: 0}},
		})
		if setW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", setW.Code, setW.Body.String())
		}

		w := doSummary(slug)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		if cc := w.Header().Get("Cache-Control"); cc != "max-age=60" {
			t.Fatalf("want a short cache TTL, got %q", cc)
		}
		got := decodeBody[statusSummary](t, w)
		if got.Title != "Summary Co" || got.Overall != "All systems operational" || got.Tone != "up" {
			t.Fatalf("unexpected headline: %+v", got)
		}
		if got.Monitors != 1 || len(got.Days) != 90 {
			t.Fatalf("want 1 monitor and 90 days, got %d and %d", got.Monitors, len(got.Days))
		}
	})
}
