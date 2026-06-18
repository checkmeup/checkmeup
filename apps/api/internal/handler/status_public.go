package handler

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
)

// StatusPublicHandler serves the public /status/:slug page (no auth required).
type StatusPublicHandler struct {
	queries *db.Queries
	tmpl    *template.Template
}

func NewStatusPublicHandler(pool *pgxpool.Pool) *StatusPublicHandler {
	tmpl := template.Must(template.New("status").Funcs(template.FuncMap{
		"safeURL": func(s string) template.URL { return template.URL(s) },
	}).Parse(statusPageHTML))
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
	ExpiresAt          string // SSL only
	DaysLeft           int    // SSL only
	MaintenanceMessage string
}

type activeMaintenance struct {
	Title   string
	Message string
}

type publicPageData struct {
	Title        string
	Description  string
	LogoURL      string
	Overall      string
	OverallColor string
	Monitors     []publicMonitorRow
	UpdatedAt    string
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

	pageMonitors, err := h.queries.ListStatusPageMonitors(ctx, page.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
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

	overall, overallColor := computeOverallStatus(rows)
	data := publicPageData{
		Title:        page.Title,
		Description:  page.Description,
		LogoURL:      page.LogoUrl,
		Overall:      overall,
		OverallColor: overallColor,
		Monitors:     rows,
		UpdatedAt:    time.Now().UTC().Format("2006-01-02 15:04 UTC"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.Execute(w, data)
}

// ─── row builder ─────────────────────────────────────────────────────────────

func (h *StatusPublicHandler) buildRow(ctx context.Context, m db.StatusPageMonitor) publicMonitorRow {
	row := publicMonitorRow{
		DisplayName: m.DisplayName,
		Type:        m.MonitorType,
		StatusColor: statusColorGray,
		StatusLabel: "Unknown",
	}
	switch m.MonitorType {
	case "uptime":
		h.fillUptimeRow(ctx, m.MonitorID, &row)
	case "cron":
		h.fillCronRow(ctx, m.MonitorID, &row)
	case "ssl":
		h.fillSSLRow(ctx, m.MonitorID, &row)
	}
	return row
}

func (h *StatusPublicHandler) fillUptimeRow(ctx context.Context, id uuid.UUID, row *publicMonitorRow) {
	mon, err := h.queries.GetUptimeMonitorPublic(ctx, id)
	if err == nil {
		row.StatusLabel, row.StatusColor = monitorStatusDisplay(string(mon.Status))
	}

	dailyRows, err := h.queries.GetUptimeDailyStatus90d(ctx, id)
	downDays := map[string]bool{}
	hasData := false
	if err == nil && len(dailyRows) > 0 {
		hasData = true
		for _, d := range dailyRows {
			if d.DownCount > 0 {
				downDays[d.Day.Time.Format("2006-01-02")] = true
			}
		}
	}
	row.Bar = build90DayBar(downDays, hasData)
}

func (h *StatusPublicHandler) fillCronRow(ctx context.Context, id uuid.UUID, row *publicMonitorRow) {
	mon, err := h.queries.GetCronMonitorPublic(ctx, id)
	if err == nil {
		row.StatusLabel, row.StatusColor = monitorStatusDisplay(string(mon.Status))
	}

	incidentDays, err := h.queries.GetCronIncidentDays90d(ctx, id)
	downDays := map[string]bool{}
	hasData := err == nil
	for _, d := range incidentDays {
		downDays[d.Time.Format("2006-01-02")] = true
	}
	// For cron, days with no ping data should show as gray only if the monitor is brand new.
	// Use monitor creation date to determine "has data" window.
	if mon2, err2 := h.queries.GetCronMonitorPublic(ctx, id); err2 == nil {
		_ = mon2
		hasData = true // cron monitors always "have data" from creation
	}
	row.Bar = build90DayBar(downDays, hasData)
}

func (h *StatusPublicHandler) fillSSLRow(ctx context.Context, id uuid.UUID, row *publicMonitorRow) {
	mon, err := h.queries.GetSSLMonitorPublic(ctx, id)
	if err != nil {
		return
	}

	switch string(mon.Status) {
	case "up":
		row.StatusLabel, row.StatusColor = "Valid", statusColorGreen
	case "expiring_soon":
		row.StatusLabel, row.StatusColor = "Expiring soon", statusColorAmber
	case "expired":
		row.StatusLabel, row.StatusColor = "Expired", statusColorRed
	case "error":
		row.StatusLabel, row.StatusColor = "Error", statusColorRed
	case "paused":
		row.StatusLabel, row.StatusColor = "Paused", statusColorGray
	default:
		row.StatusLabel, row.StatusColor = "Checking…", statusColorGray
	}

	if mon.ExpiresAt.Valid {
		row.ExpiresAt = mon.ExpiresAt.Time.Format("2006-01-02")
		row.DaysLeft = int(time.Until(mon.ExpiresAt.Time).Hours() / 24)
	}

	// SSL bar: all green except current status reflected in last segment
	barColor := statusColorGreen
	switch string(mon.Status) {
	case "expiring_soon":
		barColor = statusColorAmber
	case "expired", "error":
		barColor = statusColorRed
	case "waiting", "paused":
		barColor = statusColorLight
	}
	bar := make([]publicBar, 90)
	for i := range bar {
		bar[i] = publicBar{Color: barColor}
	}
	row.Bar = bar
}

// ─── helpers ─────────────────────────────────────────────────────────────────

const (
	statusColorGreen = "#1D9E75"
	statusColorRed   = "#EF4444"
	statusColorAmber = "#F59E0B"
	statusColorGray  = "#94A3B8"
	statusColorLight = "#E2E8F0"
)

func monitorStatusDisplay(s string) (label, color string) {
	switch s {
	case "up":
		return "Operational", statusColorGreen
	case "down":
		return "Down", statusColorRed
	case "paused":
		return "Paused", statusColorGray
	default:
		return "Checking…", statusColorGray
	}
}

func computeOverallStatus(rows []publicMonitorRow) (label, color string) {
	anyDown, anyDegraded := false, false
	for _, r := range rows {
		switch r.StatusColor {
		case statusColorRed:
			anyDown = true
		case statusColorAmber:
			anyDegraded = true
		}
	}
	switch {
	case anyDown:
		return "Major outage", statusColorRed
	case anyDegraded:
		return "Partial outage", statusColorAmber
	default:
		return "All systems operational", statusColorGreen
	}
}

func build90DayBar(downDays map[string]bool, hasData bool) []publicBar {
	bar := make([]publicBar, 90)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	for i := range bar {
		day := today.AddDate(0, 0, i-89)
		key := day.Format("2006-01-02")
		label := day.Format("Jan 2")
		var color string
		switch {
		case !hasData:
			color = statusColorLight
		case downDays[key]:
			color = statusColorRed
		default:
			color = statusColorGreen
		}
		bar[i] = publicBar{Color: color, Label: label}
	}
	return bar
}

// ─── HTML template ────────────────────────────────────────────────────────────

const statusPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}} — Status</title>
{{if .LogoURL}}<link rel="icon" href="{{safeURL .LogoURL}}">{{else}}<link rel="icon" type="image/svg+xml" href="/favicon.svg">{{end}}
<style>
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:#f8fafc;color:#1e293b;min-height:100vh}
.page{max-width:740px;margin:0 auto;padding:48px 20px 80px}
.logo{display:block;max-height:40px;margin-bottom:16px}
h1{font-size:1.5rem;font-weight:700;color:#0f172a}
.subtitle{margin-top:6px;font-size:.875rem;color:#64748b}
.banner{margin-top:28px;border-radius:12px;padding:16px 20px;display:flex;align-items:center;gap:12px;font-weight:600;font-size:.9375rem}
.dot{width:12px;height:12px;border-radius:50%;flex-shrink:0}
.monitors{margin-top:32px;display:flex;flex-direction:column;gap:12px}
.monitor{background:#fff;border:1px solid #e2e8f0;border-radius:12px;padding:16px 20px}
.monitor-header{display:flex;align-items:center;justify-content:space-between;margin-bottom:10px}
.monitor-name{font-weight:600;font-size:.9375rem;color:#0f172a}
.chip{font-size:.6875rem;font-weight:700;padding:3px 8px;border-radius:9999px;color:#fff;letter-spacing:.02em;text-transform:uppercase}
.bar{display:flex;gap:2px}
.seg{height:28px;flex:1;border-radius:2px;cursor:default}
.bar-labels{display:flex;justify-content:space-between;margin-top:4px;font-size:.6875rem;color:#94a3b8}
.ssl-info{margin-top:4px;margin-bottom:8px;font-size:.8125rem;color:#64748b}
.footer{margin-top:48px;text-align:center;font-size:.8125rem;color:#94a3b8}
.footer a{color:#94a3b8;text-decoration:none}
.footer a:hover{text-decoration:underline}
</style>
</head>
<body>
<div class="page">
  {{if .LogoURL}}<img class="logo" src="{{safeURL .LogoURL}}" alt="">{{end}}
  <h1>{{.Title}}</h1>
  {{if .Description}}<p class="subtitle">{{.Description}}</p>{{end}}

  <div class="banner" style="background:{{.OverallColor}}18;border:1px solid {{.OverallColor}}44">
    <span class="dot" style="background:{{.OverallColor}}"></span>
    <span style="color:{{.OverallColor}}">{{.Overall}}</span>
  </div>

  <div class="monitors">
    {{range .Monitors}}
    <div class="monitor">
      <div class="monitor-header">
        <span class="monitor-name">{{.DisplayName}}</span>
        <span class="chip" style="background:{{.StatusColor}}">{{.StatusLabel}}</span>
      </div>
      {{if and (eq .Type "ssl") .ExpiresAt}}
      <div class="ssl-info">Certificate expires {{.ExpiresAt}} ({{.DaysLeft}} days)</div>
      {{end}}
      {{if .MaintenanceMessage}}
      <div class="ssl-info">{{.MaintenanceMessage}}</div>
      {{end}}
      <div class="bar">
        {{range .Bar}}<div class="seg" style="background:{{.Color}}" title="{{.Label}}"></div>{{end}}
      </div>
      <div class="bar-labels"><span>90 days ago</span><span>Today</span></div>
    </div>
    {{end}}
    {{if not .Monitors}}
    <p style="color:#94a3b8;font-size:.875rem;text-align:center;padding:32px 0">No monitors on this page yet.</p>
    {{end}}
  </div>

  <div class="footer">
    <p>Last updated {{.UpdatedAt}}</p>
    <p style="margin-top:6px">Powered by <a href="https://checkmeup.net">checkmeup</a></p>
    <p style="margin-top:6px"><a href="https://checkmeup.net/faq">FAQ</a> · <a href="https://checkmeup.net/terms">Terms</a> · <a href="https://checkmeup.net/privacy">Privacy</a></p>
  </div>
</div>
</body>
</html>`
