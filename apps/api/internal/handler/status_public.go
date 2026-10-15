package handler

import (
	"context"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
)

// resolvedIncidentsPageSize is the page size for the public status page's
// paginated "past incidents" history (PRD incidents) — plain query-param
// pagination since this page is server-rendered with no JS.
const resolvedIncidentsPageSize = 10

// StatusPublicHandler serves the public /status/:slug page (no auth required).
type StatusPublicHandler struct {
	queries *db.Queries
	tmpl    *template.Template
}

func NewStatusPublicHandler(pool *pgxpool.Pool) *StatusPublicHandler {
	funcs := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		// The page styles each status by a tone class rather than by the hex
		// the view model carries (which also colors the SVG badges), so the
		// colors follow the page's light/dark tokens.
		"tone": statusTone,
	}
	tmpl := template.Must(template.New("status").Funcs(funcs).Parse(statusPageHTML))
	return &StatusPublicHandler{queries: db.New(pool), tmpl: tmpl}
}

// ─── view model ──────────────────────────────────────────────────────────────

type publicBar struct {
	Color string
	Label string // tooltip (used in title attr)
	Date  string // 2006-01-02, UTC
	// A check-based monitor's results that day (zero for cron, SSL and
	// domain): checks run, checks failed, and for uptime monitors the
	// average response time in ms.
	Checks, Down int64
	AvgMs        int32
}

type publicMonitorRow struct {
	DisplayName        string
	Type               string
	StatusLabel        string
	StatusColor        string
	Bar                []publicBar
	ExpiresAt          string // SSL/domain only
	DaysLeft           int    // SSL/domain only
	MaintenanceMessage string
}

type activeMaintenance struct {
	Title   string
	Message string
}

// publicIncidentUpdateRow is one entry in an active incident's timeline
// (design: CheckMeUp Status Page.dc.html, option 1a).
type publicIncidentUpdateRow struct {
	StatusLabel  string
	RelativeTime string
	Message      string
	IsLast       bool // true for the oldest entry — suppresses the connecting line below it
}

// publicIncidentRow is a manually-declared incident (PRD incidents), independent of
// the automatic up/down monitor rows above.
type publicIncidentRow struct {
	Title               string
	Severity            string // "Minor" / "Major" / "Critical"
	SeverityColor       string
	StatusLabel         string // "Investigating" / "Identified" / "Monitoring" / "Resolved"
	Affected            string // comma-joined display names of affected monitors
	Updates             []publicIncidentUpdateRow
	LatestUpdateMessage string
	LatestUpdateAt      string
	CreatedAt           string
	ResolvedAt          string
	Duration            string // resolved incidents only, e.g. "45 min" / "1h 12min"
}

type publicPageData struct {
	Title                 string
	Description           string
	LogoURL               string
	Initials              string
	Overall               string
	OverallColor          string
	Monitors              []publicMonitorRow
	ActiveIncidents       []publicIncidentRow
	ResolvedIncidents     []publicIncidentRow
	ResolvedIncidentsPage int
	HasPrevIncidentsPage  bool
	HasNextIncidentsPage  bool
	UpdatedAt             string
	HideBranding          bool
	Layout                string // "classic" or "grid" — ADR status-pages
}

// ─── ServeHTTP ───────────────────────────────────────────────────────────────

func (h *StatusPublicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	ctx := r.Context()

	page, err := h.queries.GetStatusPageBySlug(ctx, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	rows, err := h.loadRows(ctx, page)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	activeIncidents, err := h.loadActiveIncidents(ctx, page)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resolvedPage := parseIncidentsPageParam(r.URL.Query().Get("incidents_page"))
	resolvedIncidents, resolvedTotal, err := h.loadResolvedIncidents(ctx, page, resolvedPage)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	overall, overallColor := computeOverallStatus(rows, activeIncidents)
	data := publicPageData{
		Title:                 page.Title,
		Description:           page.Description,
		LogoURL:               page.LogoUrl,
		Initials:              initials(page.Title),
		Overall:               overall,
		OverallColor:          overallColor,
		Monitors:              rows,
		ActiveIncidents:       activeIncidents,
		ResolvedIncidents:     resolvedIncidents,
		ResolvedIncidentsPage: resolvedPage,
		HasPrevIncidentsPage:  resolvedPage > 1,
		HasNextIncidentsPage:  int64(resolvedPage*resolvedIncidentsPageSize) < resolvedTotal,
		UpdatedAt:             time.Now().UTC().Format("2006-01-02 15:04 UTC"),
		HideBranding:          page.HideBranding,
		Layout:                page.Layout,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.Execute(w, data)
}

// loadRows fetches every monitor attached to page and applies any active
// maintenance-window override — shared by the HTML page and both badge routes
// so a badge always agrees with what the page banner shows.
func (h *StatusPublicHandler) loadRows(ctx context.Context, page db.StatusPage) ([]publicMonitorRow, error) {
	pageMonitors, err := h.queries.ListStatusPageMonitors(ctx, page.ID)
	if err != nil {
		return nil, err
	}

	maintenance := map[string]activeMaintenance{}
	if active, err := h.queries.GetActiveMaintenanceForOrg(ctx, page.OrgID); err == nil {
		for _, a := range active {
			key := a.MonitorType + ":" + a.MonitorID.String()
			maintenance[key] = activeMaintenance{Title: a.Title, Message: a.Message}
		}
	}

	rows := make([]publicMonitorRow, 0, len(pageMonitors))
	for _, m := range pageMonitors {
		row := h.buildRow(ctx, m)
		if mw, ok := maintenance[m.MonitorType+":"+m.MonitorID.String()]; ok {
			row.StatusLabel = "Under maintenance"
			row.StatusColor = statusColorGray
			row.MaintenanceMessage = mw.Message
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// parseIncidentsPageParam parses the resolved-incidents 1-based page-number
// query param, defaulting to page 1 on anything invalid or out of range.
func parseIncidentsPageParam(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 1
	}
	return n
}

// loadActiveIncidents fetches manually-declared incidents (PRD incidents) that are
// still open and affect at least one monitor on this page, each with its
// full update timeline (design: CheckMeUp Status Page.dc.html, option 1a —
// the timeline shows every update, not just the latest one). Update
// timelines are fetched in one batched query for every incident on the
// page, not one query per incident — a page with many active incidents
// would otherwise issue that many separate round-trips on every
// unauthenticated page load.
// toPublicIncidentRow converts one incident row plus its (already
// time-ordered) updates into the public response shape, including the
// latest-update summary fields shown on the status page's top-level list.
func toPublicIncidentRow(r db.ListActiveStatusPageIncidentsForPageRow, incidentUpdates []db.StatusPageIncidentUpdate) publicIncidentRow {
	row := publicIncidentRow{
		Title:         r.Title,
		Severity:      incidentSeverityLabel(r.Severity),
		SeverityColor: incidentSeverityColor(r.Severity),
		StatusLabel:   incidentStatusLabel(r.Status),
		Affected:      string(r.Affected),
		CreatedAt:     r.CreatedAt.Time.UTC().Format("2006-01-02 15:04 UTC"),
	}
	row.Updates = make([]publicIncidentUpdateRow, len(incidentUpdates))
	for i, u := range incidentUpdates {
		row.Updates[i] = publicIncidentUpdateRow{
			StatusLabel:  incidentStatusLabel(u.Status),
			RelativeTime: relativeTime(u.CreatedAt.Time),
			Message:      u.Message,
			IsLast:       i == len(incidentUpdates)-1,
		}
	}
	if len(incidentUpdates) > 0 {
		row.LatestUpdateMessage = incidentUpdates[0].Message
		row.LatestUpdateAt = incidentUpdates[0].CreatedAt.Time.UTC().Format("2006-01-02 15:04 UTC")
	}
	return row
}

func (h *StatusPublicHandler) loadActiveIncidents(ctx context.Context, page db.StatusPage) ([]publicIncidentRow, error) {
	rows, err := h.queries.ListActiveStatusPageIncidentsForPage(ctx, page.ID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	updates, err := h.queries.ListStatusPageIncidentUpdatesForIncidents(ctx, ids)
	if err != nil {
		return nil, err
	}
	updatesByIncident := make(map[uuid.UUID][]db.StatusPageIncidentUpdate, len(rows))
	for _, u := range updates {
		updatesByIncident[u.IncidentID] = append(updatesByIncident[u.IncidentID], u)
	}

	result := make([]publicIncidentRow, 0, len(rows))
	for _, r := range rows {
		result = append(result, toPublicIncidentRow(r, updatesByIncident[r.ID]))
	}
	return result, nil
}

// loadResolvedIncidents fetches one page of this page's resolved-incident
// history, separate from the existing 90-day uptime bars (PRD incidents).
func (h *StatusPublicHandler) loadResolvedIncidents(ctx context.Context, page db.StatusPage, pageNum int) ([]publicIncidentRow, int64, error) {
	offset := (pageNum - 1) * resolvedIncidentsPageSize
	dbRows, err := h.queries.ListResolvedStatusPageIncidentsForPage(ctx, db.ListResolvedStatusPageIncidentsForPageParams{
		PageID: page.ID, Limit: resolvedIncidentsPageSize, Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := h.queries.CountResolvedStatusPageIncidentsForPage(ctx, page.ID)
	if err != nil {
		return nil, 0, err
	}

	result := make([]publicIncidentRow, len(dbRows))
	for i, r := range dbRows {
		row := publicIncidentRow{
			Title:         r.Title,
			Severity:      incidentSeverityLabel(r.Severity),
			SeverityColor: incidentSeverityColor(r.Severity),
			StatusLabel:   incidentStatusLabel(r.Status),
			CreatedAt:     r.CreatedAt.Time.UTC().Format("2006-01-02 15:04 UTC"),
			ResolvedAt:    r.ResolvedAt.Time.UTC().Format("2006-01-02 15:04 UTC"),
			Duration:      formatDuration(r.ResolvedAt.Time.Sub(r.CreatedAt.Time)),
		}
		if latest, err := h.queries.GetLatestStatusPageIncidentUpdate(ctx, r.ID); err == nil {
			row.LatestUpdateMessage = latest.Message
		}
		result[i] = row
	}
	return result, total, nil
}

// ─── badges (PRD status-badges) ──────────────────────────────────────────────────────────

// ServePageBadge handles GET /status/:slug/badge.svg — an embeddable badge
// showing the page's overall status, in the same three-tier wording as the
// page banner (operational/degraded/outage).
func (h *StatusPublicHandler) ServePageBadge(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	ctx := r.Context()

	page, err := h.queries.GetStatusPageBySlug(ctx, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	rows, err := h.loadRows(ctx, page)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	activeIncidents, err := h.loadActiveIncidents(ctx, page)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, color := computeOverallStatus(rows, activeIncidents)
	writeBadge(w, "status", badgeStatusWord(color), color)
}

// ServeMonitorBadge handles GET /status/:slug/badge/:monitor_id.svg — a badge
// for a single monitor attached to the page. 404s if the monitor isn't on
// that status page, so monitor status can't leak outside its page's visibility.
func (h *StatusPublicHandler) ServeMonitorBadge(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	ctx := r.Context()

	monitorID, err := uuid.Parse(chi.URLParam(r, "monitor_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	page, err := h.queries.GetStatusPageBySlug(ctx, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	pm, err := h.queries.GetStatusPageMonitor(ctx, db.GetStatusPageMonitorParams{PageID: page.ID, MonitorID: monitorID})
	if err != nil {
		http.NotFound(w, r)
		return
	}

	row := h.buildRow(ctx, pm)
	if active, err := h.queries.GetActiveMaintenanceForOrg(ctx, page.OrgID); err == nil {
		for _, a := range active {
			if a.MonitorType == pm.MonitorType && a.MonitorID == pm.MonitorID {
				row.StatusColor = statusColorGray
			}
		}
	}

	writeBadge(w, pm.DisplayName, badgeStatusWord(row.StatusColor), row.StatusColor)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

const (
	statusColorGreen = "#1D9E75"
	statusColorRed   = "#EF4444"
	statusColorAmber = "#F59E0B"
	statusColorGray  = "#94A3B8"
	statusColorLight = "#E2E8F0"
	statusColorBlue  = "#3B82F6"
)

// statusTone maps a status color to the tone class the page template styles
// it with. A minor incident is quiet (the paused gray), matching the app.
func statusTone(color string) string {
	switch color {
	case statusColorGreen:
		return "up"
	case statusColorRed:
		return "down"
	case statusColorAmber:
		return "warn"
	case statusColorLight:
		return "none"
	default:
		return "paused"
	}
}

// ─── HTML template ────────────────────────────────────────────────────────────

const statusPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}} — Status</title>
<meta name="description" content="{{if .Description}}{{.Description}}{{else}}{{.Title}} status page — live uptime, incidents, and maintenance updates.{{end}}">
<meta property="og:title" content="{{.Title}} — Status">
<meta property="og:description" content="{{if .Description}}{{.Description}}{{else}}{{.Title}} status page — live uptime, incidents, and maintenance updates.{{end}}">
<meta property="og:type" content="website">
{{if .LogoURL}}<link rel="icon" href="{{.LogoURL}}">{{else}}<link rel="icon" type="image/svg+xml" href="/favicon.svg">{{end}}
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Geist:wght@400;500;600&family=Geist+Mono:wght@400;500&display=swap" rel="stylesheet">
<script src="/status-theme.js"></script>
<style>
/* Mirrors the tokens in apps/web/src/style.css — keep both in sync if the token set changes. */
:root{
  color-scheme:dark;
  --bg:#0e1211;--surface:#131816;--surface-raised:#1a201e;--border:#232a27;--border-strong:#323b37;
  --text-muted:#7a847f;--text-dim:#9aa39e;--text:#ecf0ee;--text-strong:#ecf0ee;
  --status-up:#3fbf8f;--status-degraded:#ebae4a;--status-down:#f0605a;--status-paused:#7a847f;
  --tick-up:#2f6e58;--tick-empty:#1c2220;--accent:#4dc9a0;--on-accent:#06201a;--accent-wash:rgb(77 201 160 / 12%);--nav-bg:rgb(14 18 17 / 85%);
}
:root[data-theme='light']{
  color-scheme:light;
  --bg:#fff;--surface:#f7f9f7;--surface-raised:#f1f4f1;--border:#e4e8e4;--border-strong:#d6dcd7;
  --text-muted:#6f7873;--text-dim:#59625d;--text:#111714;--text-strong:#111714;
  --status-up:#12805f;--status-degraded:#a86200;--status-down:#c93a2e;--status-paused:#6f7873;
  --tick-up:#a9dcc7;--tick-empty:#ebeeeb;--accent:#0f6e56;--on-accent:#fff;--accent-wash:rgb(15 110 86 / 10%);--nav-bg:rgb(255 255 255 / 88%);
}
.tone-up{--tone:var(--status-up);--tick:var(--tick-up)}
.tone-down{--tone:var(--status-down);--tick:var(--status-down)}
.tone-warn{--tone:var(--status-degraded);--tick:var(--status-degraded)}
.tone-paused{--tone:var(--status-paused);--tick:var(--status-paused)}
.tone-none{--tone:var(--text-muted);--tick:var(--tick-empty)}
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
body{font-family:Geist,ui-sans-serif,system-ui,-apple-system,"Segoe UI",sans-serif;background:var(--bg);color:var(--text);min-height:100vh;-webkit-font-smoothing:antialiased;line-height:1.5}
.mono{font-family:'Geist Mono',ui-monospace,monospace}
.page{max-width:880px;margin:0 auto;padding:56px 24px 72px}
.page.layout-grid{max-width:1200px}
.topbar{position:sticky;top:0;z-index:10;border-bottom:1px solid var(--border);background:var(--nav-bg);backdrop-filter:blur(12px)}
.topbar-inner{max-width:880px;margin:0 auto;padding:0 24px;height:64px;display:flex;align-items:center;justify-content:space-between;gap:16px}
.topbar-inner.wide{max-width:1200px}
.head-left{display:flex;align-items:center;gap:12px;min-width:0}
.logo{display:block;max-height:32px;width:auto;border-radius:8px}
.avatar{width:36px;height:36px;border-radius:50%;background:var(--accent-wash);color:var(--accent);display:flex;align-items:center;justify-content:center;font-weight:600;font-size:13px;flex-shrink:0}
.page-name{font-size:15px;font-weight:600;color:var(--text-strong);white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.theme-toggle{width:36px;height:36px;border-radius:8px;border:0;background:transparent;color:var(--text-dim);display:flex;align-items:center;justify-content:center;flex-shrink:0;cursor:pointer}
.theme-toggle:hover{background:var(--surface-raised)}
.theme-toggle:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
.theme-toggle .icon-moon{display:none}
:root[data-theme='light'] .theme-toggle .icon-sun{display:none}
:root[data-theme='light'] .theme-toggle .icon-moon{display:block}
.overall{display:flex;flex-direction:column;gap:10px;margin-bottom:48px}
.overall h1{display:flex;align-items:center;gap:16px;font-size:44px;line-height:1.05;font-weight:600;letter-spacing:-.04em;color:var(--text-strong)}
.overall .description{font-size:17px;color:var(--text-dim);max-width:640px}
.overall .dot{width:12px;height:12px;border-radius:50%;flex-shrink:0;background:var(--tone);box-shadow:0 0 0 5px color-mix(in srgb,var(--tone) 18%,transparent)}
.overall p{font-size:15px;color:var(--text-dim)}
.section-title{font-size:22px;font-weight:600;letter-spacing:-.02em;color:var(--text-strong);margin-bottom:16px}
.eyebrow{font-size:13px;font-weight:600;color:var(--text-dim);margin-bottom:12px}
.incidents-active{margin-bottom:48px}
.incident{padding:20px 0 6px;border-top:1px solid var(--border)}
.incident-header{display:flex;align-items:center;gap:10px;margin-bottom:6px;flex-wrap:wrap}
.incident-title{font-size:17px;font-weight:600;letter-spacing:-.015em;color:var(--text-strong)}
.label{display:inline-flex;align-items:center;gap:7px;font-size:13px;font-weight:500;color:var(--tone);white-space:nowrap}
.label::before{content:"";width:8px;height:8px;border-radius:50%;background:var(--tone);flex-shrink:0}
.label.quiet{color:var(--text-dim)}
.incident-status{font-size:13px;color:var(--text-dim)}
.incident-affected{margin:4px 0 14px;font-size:14px;color:var(--text-dim)}
.timeline-row{display:flex;gap:14px;padding-bottom:14px}
.timeline-rail{display:flex;flex-direction:column;align-items:center;flex-shrink:0}
.timeline-dot{width:8px;height:8px;border-radius:50%;margin-top:7px;background:var(--tone)}
.timeline-line{width:1px;flex:1;margin-top:4px;background:color-mix(in srgb,var(--tone) 35%,transparent)}
.timeline-status{font-size:14px;font-weight:600;color:var(--text-strong)}
.timeline-time{font-weight:400;color:var(--text-dim);font-size:12.5px}
.timeline-message{font-size:14px;color:var(--text);margin-top:2px}
.monitors{display:flex;flex-direction:column}
.monitor{padding:18px 0;border-top:1px solid var(--border)}
.monitor-header{display:flex;align-items:baseline;justify-content:space-between;gap:12px;margin-bottom:10px}
.monitor-name{font-weight:600;font-size:15px;color:var(--text-strong);min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.bar{display:flex;gap:2px;height:32px}
.seg{flex:1;min-width:1px;border-radius:2px;background:var(--tick);cursor:default}
.bar-labels{display:flex;justify-content:space-between;margin-top:6px;font-size:12.5px;color:var(--text-dim)}
.monitor-note{margin:0 0 10px;font-size:13.5px;color:var(--text-dim)}
.empty{color:var(--text-dim);font-size:14px;text-align:center;padding:32px 0;border-top:1px solid var(--border)}
.incidents-history{margin-top:56px}
.history-row{padding:14px 0;border-top:1px solid var(--border)}
.history-head{display:flex;align-items:baseline;gap:12px;flex-wrap:wrap}
.history-title{font-size:15px;font-weight:500;color:var(--text-strong)}
.history-meta{margin-left:auto;font-size:12.5px;color:var(--text-dim);white-space:nowrap}
.history-summary{margin-top:4px;font-size:14px;color:var(--text-dim)}
.pagination{display:flex;justify-content:space-between;padding-top:14px;border-top:1px solid var(--border);font-size:14px}
.pagination a{color:var(--accent);text-decoration:none;font-weight:500}
.pagination a:hover{text-decoration:underline}
.footer{margin-top:56px;padding-top:22px;border-top:1px solid var(--border);display:flex;flex-wrap:wrap;justify-content:space-between;gap:8px 20px;font-size:13px;color:var(--text-dim)}
.footer a{color:var(--text-dim);text-decoration:none}
.footer a:hover{color:var(--text);text-decoration:underline}
.footer-links{display:flex;gap:14px;flex-wrap:wrap}
.content-grid{display:flex;gap:48px;align-items:flex-start}
.monitors-col{flex:1.6;min-width:0}
.content-grid .monitors{display:grid;grid-template-columns:1fr 1fr;column-gap:32px;min-width:0}
.content-grid .bar{height:20px}
.content-grid .bar-labels{display:none}
.sidebar{flex:1;min-width:280px;display:flex;flex-direction:column;gap:28px}
.sidebar .incidents-history{margin-top:0}
@media (max-width:860px){
  .content-grid{flex-direction:column;gap:32px}
  .content-grid .monitors{grid-template-columns:1fr}
  .monitors-col{width:100%}
  .sidebar{min-width:0;width:100%}
  .overall h1{font-size:34px}
}
</style>
</head>
<body>
<header class="topbar">
  <div class="topbar-inner{{if eq .Layout "grid"}} wide{{end}}">
    <div class="head-left">
      {{if .LogoURL}}<img class="logo" src="{{.LogoURL}}" alt="{{.Title}} logo">{{else}}<div class="avatar" aria-hidden="true">{{.Initials}}</div>{{end}}
      <div class="page-name">{{.Title}}</div>
    </div>
    <button class="theme-toggle" id="theme-toggle" title="Toggle theme" aria-label="Toggle theme">
      <svg class="icon-sun" aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="5"/>
        <line x1="12" y1="1" x2="12" y2="3"/>
        <line x1="12" y1="21" x2="12" y2="23"/>
        <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
        <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
        <line x1="1" y1="12" x2="3" y2="12"/>
        <line x1="21" y1="12" x2="23" y2="12"/>
        <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
        <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
      </svg>
      <svg class="icon-moon" aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
      </svg>
    </button>
  </div>
</header>
<div class="page {{if eq .Layout "grid"}}layout-grid{{end}}">

  <section class="overall tone-{{tone .OverallColor}}">
    <h1><span class="dot" aria-hidden="true"></span>{{.Overall}}</h1>
    {{if .Description}}<p class="description">{{.Description}}</p>{{end}}
    <p>{{if .ActiveIncidents}}{{len .ActiveIncidents}} active incident{{if ne (len .ActiveIncidents) 1}}s{{end}} · {{end}}Last updated {{.UpdatedAt}}</p>
  </section>

  {{if eq .Layout "grid"}}
  <div class="content-grid">
    <section class="monitors-col">
    <h2 class="section-title">Monitors</h2>
    <div class="monitors">
      {{range .Monitors}}{{template "monitorCard" .}}{{end}}
      {{if not .Monitors}}
      <p class="empty">No monitors on this page yet.</p>
      {{end}}
    </div>
    </section>
    <div class="sidebar">
      {{if .ActiveIncidents}}
      <section>
        <h2 class="eyebrow">Active incidents</h2>
        {{range .ActiveIncidents}}{{template "incidentCard" .}}{{end}}
      </section>
      {{end}}
      {{if .ResolvedIncidents}}
      <section class="incidents-history">
        <h2 class="section-title">Past incidents</h2>
        {{range .ResolvedIncidents}}{{template "historyRow" .}}{{end}}
        {{template "pagination" .}}
      </section>
      {{end}}
    </div>
  </div>
  {{else}}
  {{if .ActiveIncidents}}
  <section class="incidents-active">
    <h2 class="eyebrow">Active incidents</h2>
    {{range .ActiveIncidents}}{{template "incidentCard" .}}{{end}}
  </section>
  {{end}}

  <h2 class="section-title">Monitors</h2>
  <div class="monitors">
    {{range .Monitors}}{{template "monitorCard" .}}{{end}}
    {{if not .Monitors}}
    <p class="empty">No monitors on this page yet.</p>
    {{end}}
  </div>

  {{if .ResolvedIncidents}}
  <section class="incidents-history">
    <h2 class="section-title">Past incidents</h2>
    {{range .ResolvedIncidents}}{{template "historyRow" .}}{{end}}
    {{template "pagination" .}}
  </section>
  {{end}}
  {{end}}

  <footer class="footer">
    <span>Last updated {{.UpdatedAt}}</span>
    {{if not .HideBranding}}
    <span class="footer-links">
      <span>Powered by <a href="https://checkmeup.net">Checkmeup</a></span>
      <a href="https://checkmeup.net/faq">FAQ</a>
      <a href="https://checkmeup.net/terms">Terms</a>
      <a href="https://checkmeup.net/privacy">Privacy</a>
    </span>
    {{end}}
  </footer>
</div>
</body>
</html>

{{define "pagination"}}
  {{if or .HasPrevIncidentsPage .HasNextIncidentsPage}}
  <nav class="pagination" aria-label="Past incidents pages">
    {{if .HasPrevIncidentsPage}}<a href="?incidents_page={{sub .ResolvedIncidentsPage 1}}">← Newer</a>{{else}}<span></span>{{end}}
    {{if .HasNextIncidentsPage}}<a href="?incidents_page={{add .ResolvedIncidentsPage 1}}">Older →</a>{{end}}
  </nav>
  {{end}}
{{end}}

{{define "incidentCard"}}
  <section class="incident tone-{{tone .SeverityColor}}">
    <div class="incident-header">
      <span class="label">{{.Severity}}</span>
      <span class="incident-status">· {{.StatusLabel}}</span>
    </div>
    <h2 class="incident-title">{{.Title}}</h2>
    {{if .Affected}}<p class="incident-affected">Affecting: {{.Affected}}</p>{{end}}
    {{range .Updates}}
    <div class="timeline-row">
      <div class="timeline-rail" aria-hidden="true">
        <span class="timeline-dot"></span>
        {{if not .IsLast}}<span class="timeline-line"></span>{{end}}
      </div>
      <div>
        <div class="timeline-status">{{.StatusLabel}} <span class="timeline-time">· {{.RelativeTime}}</span></div>
        <div class="timeline-message">{{.Message}}</div>
      </div>
    </div>
    {{end}}
  </section>
{{end}}

{{define "monitorCard"}}
    <div class="monitor">
      <div class="monitor-header">
        <span class="monitor-name">{{.DisplayName}}</span>
        <span class="label tone-{{tone .StatusColor}}{{if eq (tone .StatusColor) "up"}} quiet{{end}}">{{.StatusLabel}}</span>
      </div>
      {{if and (eq .Type "ssl") .ExpiresAt}}
      <p class="monitor-note">Certificate expires {{.ExpiresAt}} ({{.DaysLeft}} days)</p>
      {{end}}
      {{if and (eq .Type "domain") .ExpiresAt}}
      <p class="monitor-note">Domain expires {{.ExpiresAt}} ({{.DaysLeft}} days)</p>
      {{end}}
      {{if .MaintenanceMessage}}
      <p class="monitor-note">{{.MaintenanceMessage}}</p>
      {{end}}
      <div class="bar">
        {{range .Bar}}<div class="seg tone-{{tone .Color}}" title="{{.Label}}"></div>{{end}}
      </div>
      <div class="bar-labels"><span>90 days ago</span><span>Today</span></div>
    </div>
{{end}}

{{define "historyRow"}}
    <div class="history-row">
      <div class="history-head">
        <span class="history-title">{{.Title}}</span>
        <span class="history-meta">{{.ResolvedAt}} · {{.Duration}}</span>
      </div>
      {{if .LatestUpdateMessage}}<p class="history-summary">{{.LatestUpdateMessage}}</p>{{end}}
    </div>
{{end}}`
