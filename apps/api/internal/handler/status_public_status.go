package handler

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/checkmeup/checkmeup/internal/db"
)

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
	case "dns":
		h.fillDNSRow(ctx, m.MonitorID, &row)
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

func (h *StatusPublicHandler) fillDNSRow(ctx context.Context, id uuid.UUID, row *publicMonitorRow) {
	mon, err := h.queries.GetDNSMonitorPublic(ctx, id)
	if err == nil {
		row.StatusLabel, row.StatusColor = monitorStatusDisplay(string(mon.Status))
	}

	dailyRows, err := h.queries.GetDNSDailyStatus90d(ctx, id)
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

// computeOverallStatus drives both the page banner and the badge routes
// (kept in sync so a badge always agrees with what the page shows). An
// active major/critical incident (US-2403) escalates it exactly like a
// monitor outage would, independent of any monitor's own up/down state.
func computeOverallStatus(rows []publicMonitorRow, activeIncidents []publicIncidentRow) (label, color string) {
	anyDown, anyDegraded := false, false
	for _, r := range rows {
		applySeverity(r.StatusColor, &anyDown, &anyDegraded)
	}
	for _, inc := range activeIncidents {
		applySeverity(inc.SeverityColor, &anyDown, &anyDegraded)
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

// applySeverity sets *anyDown or *anyDegraded when color represents a down
// (red) or degraded (amber) state — shared by the monitor-row and
// incident-severity loops in computeOverallStatus, which apply the same
// red/amber/other classification to two different fields. Out-params rather
// than a returned pair so the branching stays inside this one function
// instead of also needing an if/or at each call site.
func applySeverity(color string, anyDown, anyDegraded *bool) {
	switch color {
	case statusColorRed:
		*anyDown = true
	case statusColorAmber:
		*anyDegraded = true
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
