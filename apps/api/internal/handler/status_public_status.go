package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

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
	stats := map[string]dayCheckStats{}
	if err == nil {
		for _, d := range dailyRows {
			stats[d.Day.Time.Format("2006-01-02")] = dayCheckStats{checks: d.CheckCount, down: d.DownCount, avgMs: d.AvgResponseMs}
		}
	}
	row.Bar = barWithStats(stats, incidentDayKeys(h.queries.GetUptimeIncidentDays90d(ctx, id)))
}

func (h *StatusPublicHandler) fillCronRow(ctx context.Context, id uuid.UUID, row *publicMonitorRow) {
	mon, err := h.queries.GetCronMonitorPublic(ctx, id)
	if err == nil {
		row.StatusLabel, row.StatusColor = monitorStatusDisplay(string(mon.Status))
	}

	incidentDays, err := h.queries.GetCronIncidentDays90d(ctx, id)
	downDays := incidentDayKeys(incidentDays, err)
	// A cron monitor keeps no per-day history (pings are pruned after 30
	// days), so its days with data run from its creation, once it has pinged.
	var dataDays map[string]bool
	if mon.LastPingAt.Valid && downDays != nil {
		dataDays = daysSince(mon.CreatedAt.Time)
	}
	row.Bar = build90DayBar(downDays, dataDays)
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
	row.Bar = solidColorBar(string(mon.Status), mon.CreatedAt.Time)
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
	row.Bar = solidColorBar(string(mon.Status), mon.CreatedAt.Time)
}

func (h *StatusPublicHandler) fillPortRow(ctx context.Context, id uuid.UUID, row *publicMonitorRow) {
	mon, err := h.queries.GetPortMonitorPublic(ctx, id)
	if err == nil {
		row.StatusLabel, row.StatusColor = monitorStatusDisplay(string(mon.Status))
	}

	dailyRows, err := h.queries.GetPortDailyStatus90d(ctx, id)
	stats := map[string]dayCheckStats{}
	if err == nil {
		for _, d := range dailyRows {
			stats[d.Day.Time.Format("2006-01-02")] = dayCheckStats{checks: d.CheckCount, down: d.DownCount}
		}
	}
	row.Bar = barWithStats(stats, incidentDayKeys(h.queries.GetPortIncidentDays90d(ctx, id)))
}

func (h *StatusPublicHandler) fillDNSRow(ctx context.Context, id uuid.UUID, row *publicMonitorRow) {
	mon, err := h.queries.GetDNSMonitorPublic(ctx, id)
	if err == nil {
		row.StatusLabel, row.StatusColor = monitorStatusDisplay(string(mon.Status))
	}

	dailyRows, err := h.queries.GetDNSDailyStatus90d(ctx, id)
	stats := map[string]dayCheckStats{}
	if err == nil {
		for _, d := range dailyRows {
			stats[d.Day.Time.Format("2006-01-02")] = dayCheckStats{checks: d.CheckCount, down: d.DownCount}
		}
	}
	row.Bar = barWithStats(stats, incidentDayKeys(h.queries.GetDNSIncidentDays90d(ctx, id)))
}

// dayCheckStats is one day of a check-based monitor's results: how many
// checks ran, how many failed, and (uptime monitors only) their average
// response time in ms.
type dayCheckStats struct {
	checks, down int64
	avgMs        int32
}

// incidentDayKeys keys a monitor's incident days by date; nil when they
// couldn't be read, so its bar shows no data rather than 90 clean days.
func incidentDayKeys(days []pgtype.Date, err error) map[string]bool {
	if err != nil {
		return nil
	}
	keys := map[string]bool{}
	for _, d := range days {
		keys[d.Time.Format("2006-01-02")] = true
	}
	return keys
}

// barWithStats builds the 90-day bar of a check-based monitor: a day is down
// when an incident covered it (downDays), not when a check failed, since a
// failure its recheck cleared alerts no one. Each segment still carries its
// day's check counts, so summary.json reports uptime % and latency per day.
func barWithStats(stats map[string]dayCheckStats, downDays map[string]bool) []publicBar {
	// A day with no check (before the monitor existed, or while paused) has
	// no data; so does every day when the incidents couldn't be read.
	var dataDays map[string]bool
	if downDays != nil {
		dataDays = map[string]bool{}
		for day := range stats {
			dataDays[day] = true
		}
	}
	bar := build90DayBar(downDays, dataDays)
	for i := range bar {
		if st, ok := stats[bar[i].Date]; ok {
			bar[i].Checks, bar[i].Down, bar[i].AvgMs = st.checks, st.down, st.avgMs
		}
	}
	return bar
}

// computeOverallStatus drives both the page banner and the badge routes
// (kept in sync so a badge always agrees with what the page shows). An
// active major/critical incident (PRD incidents) escalates it exactly like a
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

// daysSince keys every UTC day from t's through today.
func daysSince(t time.Time) map[string]bool {
	days := map[string]bool{}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	for day := t.UTC().Truncate(24 * time.Hour); !day.After(today); day = day.AddDate(0, 0, 1) {
		days[day.Format("2006-01-02")] = true
	}
	return days
}

// build90DayBar colors the last 90 days, today last: red for a day in
// downDays, green for one in dataDays, and the no-data color for the rest.
func build90DayBar(downDays, dataDays map[string]bool) []publicBar {
	bar := make([]publicBar, 90)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	for i := range bar {
		day := today.AddDate(0, 0, i-89)
		key := day.Format("2006-01-02")
		label := day.Format("Jan 2")
		var color string
		switch {
		case downDays[key]:
			color = statusColorRed
		case dataDays[key]:
			color = statusColorGreen
		default:
			color = statusColorLight
		}
		bar[i] = publicBar{Color: color, Label: label, Date: key}
	}
	return bar
}
