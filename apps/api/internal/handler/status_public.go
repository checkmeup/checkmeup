package handler

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
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

// initials derives a 1-2 letter fallback badge from a status page's title
// (e.g. "Acme Corp" -> "AC"), shown when the org hasn't uploaded a custom
// logo — mirrors the avatar-badge pattern in the CheckMeUp Design mockup.
func initials(title string) string {
	fields := strings.Fields(title)
	out := ""
	for _, f := range fields {
		r := []rune(f)
		if len(r) == 0 {
			continue
		}
		out += strings.ToUpper(string(r[0]))
		if len(out) >= 2 {
			break
		}
	}
	if out == "" {
		return "?"
	}
	return out
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

// relativeTime formats a past timestamp the way the incident timeline design
// does ("58 min ago", "2h 10min ago", "3 days ago") — coarser than a clock
// time since visitors care about recency, not the exact minute.
func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		h := int(d.Hours())
		m := int(d.Minutes()) - h*60
		if m == 0 {
			return fmt.Sprintf("%dh ago", h)
		}
		return fmt.Sprintf("%dh %dmin ago", h, m)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// formatDuration renders an incident's total open time for the resolved
// history ("45 min", "1h 12min") — same coarse-grain style as relativeTime.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "< 1 min"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d min", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) - h*60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dmin", h, m)
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
		row := publicIncidentRow{
			Title:         r.Title,
			Severity:      incidentSeverityLabel(r.Severity),
			SeverityColor: incidentSeverityColor(r.Severity),
			StatusLabel:   incidentStatusLabel(r.Status),
			Affected:      string(r.Affected),
			CreatedAt:     r.CreatedAt.Time.Format("2006-01-02 15:04 UTC"),
		}
		incidentUpdates := updatesByIncident[r.ID]
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
		result = append(result, row)
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
