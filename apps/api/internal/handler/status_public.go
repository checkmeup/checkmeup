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
// paginated "past incidents" history (US-2403) — plain query-param
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
	}
	tmpl := template.Must(template.New("status").Funcs(funcs).Parse(statusPageHTML))
	return &StatusPublicHandler{queries: db.New(pool), tmpl: tmpl}
}

// ─── view model ──────────────────────────────────────────────────────────────

type publicBar struct {
	Color string
	Label string // tooltip (used in title attr)
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

// publicIncidentRow is a manually-declared incident (EP-24), independent of
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

// loadActiveIncidents fetches manually-declared incidents (EP-24) that are
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
		CreatedAt:     r.CreatedAt.Time.Format("2006-01-02 15:04 UTC"),
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
		row.LatestUpdateAt = incidentUpdates[0].CreatedAt.Time.Format("2006-01-02 15:04 UTC")
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
// history, separate from the existing 90-day uptime bars (US-2403).
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
			CreatedAt:     r.CreatedAt.Time.Format("2006-01-02 15:04 UTC"),
			ResolvedAt:    r.ResolvedAt.Time.Format("2006-01-02 15:04 UTC"),
			Duration:      formatDuration(r.ResolvedAt.Time.Sub(r.CreatedAt.Time)),
		}
		if latest, err := h.queries.GetLatestStatusPageIncidentUpdate(ctx, r.ID); err == nil {
			row.LatestUpdateMessage = latest.Message
		}
		result[i] = row
	}
	return result, total, nil
}

// ─── badges (EP-30) ──────────────────────────────────────────────────────────

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
<script src="/status-theme.js"></script>
<style>
/* Mirrors apps/web/src/style.css — keep both in sync if the token set changes. */
:root{
  --bg:#0a0f0d;--surface:#101512;--surface-raised:#1a221d;--border:rgb(255 255 255 / 8%);
  --text-muted:rgb(242 245 243 / 28%);--text-dim:rgb(242 245 243 / 55%);--text:#f2f5f3;--text-strong:#f2f5f3;
  --status-up:#1d9e75;--status-degraded:#f59e0b;--status-down:#ef4444;--status-paused:#94a3b8;
  --card:rgb(255 255 255 / 3.5%);--accent:#1d9e75;--accent-wash:rgb(29 158 117 / 13%);--on-accent:#fff;
}
:root[data-theme='light']{
  --bg:#fbfdfc;--surface:#f2f6f4;--surface-raised:#e7ede9;--border:rgb(0 0 0 / 8%);
  --text-muted:rgb(11 15 12 / 45%);--text-dim:rgb(11 15 12 / 70%);--text:#0d1512;--text-strong:#0d1512;
  --card:rgb(0 0 0 / 3%);--accent:#0f6e56;--accent-wash:rgb(15 110 86 / 10%);
}
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
body{font-family:'Inter',-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:var(--bg);color:var(--text);min-height:100vh;-webkit-font-smoothing:antialiased}
.page{max-width:720px;margin:0 auto;padding:40px 20px 64px}
.head{display:flex;align-items:flex-start;justify-content:space-between;gap:16px;margin-bottom:26px}
.head-left{display:flex;align-items:center;gap:12px}
.logo{display:block;max-height:38px;width:auto;border-radius:10px}
.avatar{width:38px;height:38px;border-radius:10px;background:var(--accent);color:var(--on-accent);display:flex;align-items:center;justify-content:center;font-weight:700;font-size:14px;flex-shrink:0}
h1{font-size:20px;font-weight:700;letter-spacing:-.01em;color:var(--text-strong)}
.subtitle{margin-top:3px;font-size:13px;color:var(--text-muted)}
.theme-toggle{width:30px;height:30px;border-radius:8px;border:1px solid var(--border);background:var(--surface-raised);color:var(--text-dim);display:flex;align-items:center;justify-content:center;flex-shrink:0;cursor:pointer}
.theme-toggle .icon-moon{display:none}
:root[data-theme='light'] .theme-toggle .icon-sun{display:none}
:root[data-theme='light'] .theme-toggle .icon-moon{display:block}
.banner{border-radius:12px;padding:15px 18px;display:flex;align-items:center;gap:11px;font-weight:600;font-size:14px;margin-bottom:22px}
.dot{width:11px;height:11px;border-radius:50%;flex-shrink:0}
.banner-count{margin-left:auto;font-size:12px;font-weight:500;color:var(--text-muted)}
.incident{border-radius:12px;padding:18px 20px;margin-bottom:26px}
.incident-header{display:flex;align-items:center;gap:8px;margin-bottom:4px;flex-wrap:wrap}
.incident-title{font-size:14px;font-weight:700;color:var(--text-strong)}
.incident-status-chip{padding:2px 9px;border-radius:100px;background:var(--surface-raised);color:var(--text-dim);font-size:10.5px;font-weight:600}
.incident-affected{margin:8px 0 14px;font-size:12.5px;color:var(--text-muted)}
.timeline-row{display:flex;gap:12px;padding-bottom:14px}
.timeline-rail{display:flex;flex-direction:column;align-items:center;flex-shrink:0}
.timeline-dot{width:7px;height:7px;border-radius:50%;margin-top:5px}
.timeline-line{width:1px;flex:1;margin-top:4px}
.timeline-status{font-size:12.5px;font-weight:600;color:var(--text-strong)}
.timeline-time{font-weight:400;color:var(--text-muted);font-family:'IBM Plex Mono',ui-monospace,monospace;font-size:11px}
.timeline-message{font-size:12.5px;color:var(--text-dim);margin-top:3px;line-height:1.5}
.incidents-history{margin-top:24px;margin-bottom:8px}
.section-title{font-size:12.5px;font-weight:600;color:var(--text-dim);margin-bottom:10px}
.history-row{padding:12px 0;border-top:1px solid var(--border)}
.history-head{display:flex;align-items:center;gap:8px}
.history-title{font-size:13px;font-weight:600;color:var(--text-strong)}
.history-meta{margin-left:auto;font-family:'IBM Plex Mono',ui-monospace,monospace;font-size:11px;color:var(--text-muted);white-space:nowrap}
.history-summary{margin:4px 0 0;font-size:12px;color:var(--text-muted)}
.pagination{display:flex;justify-content:space-between;margin-top:10px;font-size:12.5px}
.pagination a{color:var(--accent);text-decoration:none}
.pagination a:hover{text-decoration:underline}
.monitors{display:flex;flex-direction:column;gap:12px;margin-bottom:8px}
.monitor{background:var(--card);border:1px solid var(--border);border-radius:12px;padding:16px 18px}
.monitor-header{display:flex;align-items:center;justify-content:space-between;gap:8px;margin-bottom:10px}
.monitor-name{font-weight:600;font-size:14px;color:var(--text-strong)}
.chip{font-size:10.5px;font-weight:700;padding:3px 10px;border-radius:100px;color:#fff;letter-spacing:.03em;text-transform:uppercase;flex-shrink:0}
.bar{display:flex;gap:1.5px}
.seg{height:24px;flex:1;border-radius:1px;cursor:default}
.bar-labels{display:flex;justify-content:space-between;margin-top:5px;font-size:10.5px;color:var(--text-muted)}
.monitor-note{margin:0 0 10px;font-size:12px;color:var(--text-muted)}
.footer{margin-top:36px;padding-top:22px;border-top:1px solid var(--border);text-align:center;font-size:12px;color:var(--text-muted)}
.footer p+p{margin-top:6px}
.footer a{color:var(--text-muted);text-decoration:none}
.footer a:hover{text-decoration:underline}
</style>
</head>
<body>
<div class="page">
  <div class="head">
    <div class="head-left">
      {{if .LogoURL}}<img class="logo" src="{{.LogoURL}}" alt="{{.Title}} logo">{{else}}<div class="avatar">{{.Initials}}</div>{{end}}
      <div>
        <h1>{{.Title}}</h1>
        {{if .Description}}<p class="subtitle">{{.Description}}</p>{{end}}
      </div>
    </div>
    <button class="theme-toggle" id="theme-toggle" title="Toggle theme" aria-label="Toggle theme">
      <svg class="icon-sun" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
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
      <svg class="icon-moon" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
      </svg>
    </button>
  </div>

  <div class="banner" style="background:{{.OverallColor}}18;border:1px solid {{.OverallColor}}44">
    <span class="dot" style="background:{{.OverallColor}}"></span>
    <span style="color:{{.OverallColor}}">{{.Overall}}</span>
    {{if .ActiveIncidents}}
    <span class="banner-count">{{len .ActiveIncidents}} active incident{{if ne (len .ActiveIncidents) 1}}s{{end}}</span>
    {{end}}
  </div>

  {{range .ActiveIncidents}}
  <div class="incident" style="border:1px solid {{.SeverityColor}}4D;background:{{.SeverityColor}}14">
    <div class="incident-header">
      <span class="chip" style="background:{{.SeverityColor}}">{{.Severity}}</span>
      <span class="incident-status-chip">{{.StatusLabel}}</span>
      <span class="incident-title">{{.Title}}</span>
    </div>
    {{if .Affected}}<p class="incident-affected">Affecting: {{.Affected}}</p>{{end}}
    {{$incidentColor := .SeverityColor}}
    {{range .Updates}}
    <div class="timeline-row">
      <div class="timeline-rail">
        <span class="timeline-dot" style="background:{{$incidentColor}}"></span>
        {{if not .IsLast}}<span class="timeline-line" style="background:{{$incidentColor}}59"></span>{{end}}
      </div>
      <div>
        <div class="timeline-status">{{.StatusLabel}} <span class="timeline-time">· {{.RelativeTime}}</span></div>
        <div class="timeline-message">{{.Message}}</div>
      </div>
    </div>
    {{end}}
  </div>
  {{end}}

  <div class="monitors">
    {{range .Monitors}}
    <div class="monitor">
      <div class="monitor-header">
        <span class="monitor-name">{{.DisplayName}}</span>
        <span class="chip" style="background:{{.StatusColor}}">{{.StatusLabel}}</span>
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
        {{range .Bar}}<div class="seg" style="background:{{.Color}}" title="{{.Label}}"></div>{{end}}
      </div>
      <div class="bar-labels"><span>90 days ago</span><span>Today</span></div>
    </div>
    {{end}}
    {{if not .Monitors}}
    <p style="color:var(--text-muted);font-size:14px;text-align:center;padding:32px 0">No monitors on this page yet.</p>
    {{end}}
  </div>

  {{if .ResolvedIncidents}}
  <div class="incidents-history">
    <h2 class="section-title">Past incidents</h2>
    {{range .ResolvedIncidents}}
    <div class="history-row">
      <div class="history-head">
        <span class="history-title">{{.Title}}</span>
        <span class="history-meta">{{.ResolvedAt}} · {{.Duration}}</span>
      </div>
      {{if .LatestUpdateMessage}}<p class="history-summary">{{.LatestUpdateMessage}}</p>{{end}}
    </div>
    {{end}}
    {{if or .HasPrevIncidentsPage .HasNextIncidentsPage}}
    <div class="pagination">
      {{if .HasPrevIncidentsPage}}<a href="?incidents_page={{sub .ResolvedIncidentsPage 1}}">← Newer</a>{{else}}<span></span>{{end}}
      {{if .HasNextIncidentsPage}}<a href="?incidents_page={{add .ResolvedIncidentsPage 1}}">Older →</a>{{end}}
    </div>
    {{end}}
  </div>
  {{end}}

  <div class="footer">
    <p>Last updated {{.UpdatedAt}}</p>
    {{if not .HideBranding}}
    <p>Powered by <a href="https://checkmeup.net">Checkmeup</a></p>
    <p><a href="https://checkmeup.net/faq">FAQ</a> · <a href="https://checkmeup.net/terms">Terms</a> · <a href="https://checkmeup.net/privacy">Privacy</a></p>
    {{end}}
  </div>
</div>
</body>
</html>`
