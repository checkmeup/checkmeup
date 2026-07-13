package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
)

const (
	statusColorGreen = "#1D9E75"
	statusColorRed   = "#EF4444"
	statusColorAmber = "#F59E0B"
	statusColorGray  = "#94A3B8"
	statusColorLight = "#E2E8F0"
	statusColorBlue  = "#3B82F6"
)

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
	case "domain":
		h.fillDomainRow(ctx, m.MonitorID, &row)
	case "port":
		h.fillPortRow(ctx, m.MonitorID, &row)
	}
	return row
}

func (h *StatusPublicHandler) fillUptimeRow(ctx context.Context, id uuid.UUID, row *publicMonitorRow) {
	mon, err := h.queries.GetUptimeMonitorPublic(ctx, id)
	if err == nil {
		row.StatusLabel, row.StatusColor = monitorStatusDisplay(string(mon.Status))
	}

	dailyRows, err := h.queries.GetUptimeDailyStatus90d(ctx, id)
	days := make([]dailyDownCount, len(dailyRows))
	for i, d := range dailyRows {
		days[i] = dailyDownCount{Day: d.Day, DownCount: d.DownCount}
	}
	row.Bar = build90DayBarFromDailyCounts(days, err)
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
	row.StatusLabel, row.StatusColor = expiryStatusDisplay(string(mon.Status))
	if mon.ExpiresAt.Valid {
		row.ExpiresAt = mon.ExpiresAt.Time.Format("2006-01-02")
		row.DaysLeft = int(time.Until(mon.ExpiresAt.Time).Hours() / 24)
	}
	row.Bar = solidColorBar(string(mon.Status))
}

func (h *StatusPublicHandler) fillDomainRow(ctx context.Context, id uuid.UUID, row *publicMonitorRow) {
	mon, err := h.queries.GetDomainMonitorPublic(ctx, id)
	if err != nil {
		return
	}
	row.StatusLabel, row.StatusColor = expiryStatusDisplay(string(mon.Status))
	if mon.ExpiresAt.Valid {
		row.ExpiresAt = mon.ExpiresAt.Time.Format("2006-01-02")
		row.DaysLeft = int(time.Until(mon.ExpiresAt.Time).Hours() / 24)
	}
	row.Bar = solidColorBar(string(mon.Status))
}

func (h *StatusPublicHandler) fillPortRow(ctx context.Context, id uuid.UUID, row *publicMonitorRow) {
	mon, err := h.queries.GetPortMonitorPublic(ctx, id)
	if err == nil {
		row.StatusLabel, row.StatusColor = monitorStatusDisplay(string(mon.Status))
	}

	dailyRows, err := h.queries.GetPortDailyStatus90d(ctx, id)
	days := make([]dailyDownCount, len(dailyRows))
	for i, d := range dailyRows {
		days[i] = dailyDownCount{Day: d.Day, DownCount: d.DownCount}
	}
	row.Bar = build90DayBarFromDailyCounts(days, err)
}

// ─── display helpers ─────────────────────────────────────────────────────────

// incidentSeverityLabel/-Color mirror the minor/major/critical vocabulary
// from US-2401 — minor is informational (blue) and doesn't escalate the
// overall banner; major/critical reuse the same red/amber tiers monitor
// outages already use, so computeOverallStatus treats them identically.
func incidentSeverityLabel(s db.IncidentSeverity) string {
	switch s {
	case db.IncidentSeverityCritical:
		return "Critical"
	case db.IncidentSeverityMajor:
		return "Major"
	default:
		return "Minor"
	}
}

func incidentSeverityColor(s db.IncidentSeverity) string {
	switch s {
	case db.IncidentSeverityCritical:
		return statusColorRed
	case db.IncidentSeverityMajor:
		return statusColorAmber
	default:
		return statusColorBlue
	}
}

func incidentStatusLabel(s db.IncidentStatus) string {
	switch s {
	case db.IncidentStatusIdentified:
		return "Identified"
	case db.IncidentStatusMonitoring:
		return "Monitoring"
	case db.IncidentStatusResolved:
		return "Resolved"
	default:
		return "Investigating"
	}
}

func expiryStatusDisplay(s string) (label, color string) {
	switch s {
	case "up":
		return "Valid", statusColorGreen
	case "expiring_soon":
		return "Expiring soon", statusColorAmber
	case "expired":
		return "Expired", statusColorRed
	case "error":
		return "Error", statusColorRed
	case "paused":
		return "Paused", statusColorGray
	default:
		return "Checking…", statusColorGray
	}
}

func solidColorBar(status string) []publicBar {
	color := statusColorGreen
	switch status {
	case "expiring_soon":
		color = statusColorAmber
	case "expired", "error":
		color = statusColorRed
	case "waiting", "paused":
		color = statusColorLight
	}
	bar := make([]publicBar, 90)
	for i := range bar {
		bar[i] = publicBar{Color: color}
	}
	return bar
}

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

// computeOverallStatus drives both the page banner and the badge routes
// (kept in sync so a badge always agrees with what the page shows). An
// active major/critical incident (US-2403) escalates it exactly like a
// monitor outage would, independent of any monitor's own up/down state.
func computeOverallStatus(rows []publicMonitorRow, activeIncidents []publicIncidentRow) (label, color string) {
	anyDown, anyDegraded := false, false
	for _, r := range rows {
		switch r.StatusColor {
		case statusColorRed:
			anyDown = true
		case statusColorAmber:
			anyDegraded = true
		}
	}
	for _, inc := range activeIncidents {
		switch inc.SeverityColor {
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

// dailyDownCount is the shape shared by GetUptimeDailyStatus90dRow and
// GetPortDailyStatus90dRow — distinct sqlc-generated types with identical
// fields, normalized here so fillUptimeRow/fillPortRow can share one bar
// builder instead of duplicating the same downDays/hasData loop.
type dailyDownCount struct {
	Day       pgtype.Date
	DownCount int64
}

// build90DayBarFromDailyCounts computes the 90-day bar from a per-day
// down-count query result — shared by uptime and port monitors, which both
// track a simple per-day count (unlike cron, which derives days from
// incident windows instead, or SSL/domain, which use a flat solid-color bar).
func build90DayBarFromDailyCounts(days []dailyDownCount, queryErr error) []publicBar {
	downDays := map[string]bool{}
	hasData := queryErr == nil && len(days) > 0
	if hasData {
		for _, d := range days {
			if d.DownCount > 0 {
				downDays[d.Day.Time.Format("2006-01-02")] = true
			}
		}
	}
	return build90DayBar(downDays, hasData)
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
