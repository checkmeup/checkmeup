package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

// checkHistoryHours is how far back the dashboard's check strips reach: one
// tick per hour.
const checkHistoryHours = 48

type checkHistoryMonitor struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	// One entry per hour, oldest first, ending with the current hour:
	// "up" when every check that hour passed, "down" when any failed, null
	// when no check ran. A cron monitor is "up" every hour from its first
	// ping on that no incident was open (ListCronPingHistory). SSL and domain
	// monitors keep no check history: every hour since they were created is
	// their current status, "warn" when expiring soon (expiryTicks).
	Ticks []*string `json:"ticks"`
}

type checkHistoryResponse struct {
	Hours    int                   `json:"hours"`
	Monitors []checkHistoryMonitor `json:"monitors"`
}

// hourResult is one monitor's checks in one hour, whatever the monitor type.
type hourResult struct {
	monitorType string
	monitorID   uuid.UUID
	hour        time.Time
	checks      int32
	failures    int32
}

var (
	tickUp   = "up"
	tickWarn = "warn"
	tickDown = "down"
)

// buildCheckHistory turns sparse per-hour results into a dense tick list per
// monitor covering the `hours` hours that end with the one `now` falls in.
// Results outside that window are ignored; monitors with no results are
// left out.
func buildCheckHistory(now time.Time, hours int, results []hourResult) []checkHistoryMonitor {
	current := now.UTC().Truncate(time.Hour)
	start := current.Add(-time.Duration(hours-1) * time.Hour)

	type key struct {
		monitorType string
		id          uuid.UUID
	}
	byMonitor := map[key][]*string{}
	var order []key

	for _, res := range results {
		idx := int(res.hour.UTC().Truncate(time.Hour).Sub(start) / time.Hour)
		if idx < 0 || idx >= hours || res.checks+res.failures == 0 {
			continue
		}
		k := key{res.monitorType, res.monitorID}
		ticks, ok := byMonitor[k]
		if !ok {
			ticks = make([]*string, hours)
			order = append(order, k)
		}
		// A failure anywhere in the hour wins over passing checks, including
		// a cron incident row landing in the same hour as a ping.
		if res.failures > 0 {
			ticks[idx] = &tickDown
		} else if ticks[idx] == nil {
			ticks[idx] = &tickUp
		}
		byMonitor[k] = ticks
	}

	out := make([]checkHistoryMonitor, 0, len(order))
	for _, k := range order {
		out = append(out, checkHistoryMonitor{Type: k.monitorType, ID: k.id.String(), Ticks: byMonitor[k]})
	}
	return out
}

// ListCheckHistory returns the last 48 hours of check results for every
// uptime, port, DNS and cron monitor in the org. SSL and domain monitors keep
// no check history, so they never appear.
func (h *MonitorHandler) ListCheckHistory(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}
	results, err := h.loadHourResults(r.Context(), orgID, int32(checkHistoryHours))
	if err != nil {
		slog.ErrorContext(r.Context(), "list check history failed", "err", err)
		respond.InternalError(w)
		return
	}
	now := time.Now()
	monitors := buildCheckHistory(now, checkHistoryHours, results)
	expiry, err := h.expiryHistory(r.Context(), orgID, now, checkHistoryHours)
	if err != nil {
		slog.ErrorContext(r.Context(), "list check history failed", "err", err)
		respond.InternalError(w)
		return
	}
	respond.JSON(w, http.StatusOK, checkHistoryResponse{
		Hours:    checkHistoryHours,
		Monitors: append(monitors, expiry...),
	})
}

// expiryHistory gives the org's SSL and domain monitors their strips, which
// expiryTicks draws from the status alone: they keep no check history.
func (h *MonitorHandler) expiryHistory(ctx context.Context, orgID uuid.UUID, now time.Time, hours int) ([]checkHistoryMonitor, error) {
	var out []checkHistoryMonitor
	add := func(monitorType string, id uuid.UUID, status string, created time.Time) {
		if ticks := expiryTicks(now, hours, created, status); ticks != nil {
			out = append(out, checkHistoryMonitor{Type: monitorType, ID: id.String(), Ticks: ticks})
		}
	}
	ssl, err := h.queries.ListSSLMonitors(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("ssl: %w", err)
	}
	for _, m := range ssl {
		add("ssl", m.ID, string(m.Status), m.CreatedAt.Time)
	}
	domains, err := h.queries.ListDomainMonitors(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("domain: %w", err)
	}
	for _, m := range domains {
		add("domain", m.ID, string(m.Status), m.CreatedAt.Time)
	}
	return out, nil
}

// expiryTicks colors every hour since an SSL or domain monitor was created by
// its current status, as the status page's bar does (solidColorBar): a
// certificate or registration changes state in days, not hours, and the
// monitor keeps no record of when it did. A monitor not yet checked, or
// paused, has no strip (nil).
func expiryTicks(now time.Time, hours int, created time.Time, status string) []*string {
	var tick *string
	switch status {
	case "up":
		tick = &tickUp
	case "expiring_soon":
		tick = &tickWarn
	case "expired", "error":
		tick = &tickDown
	default:
		return nil
	}
	current := now.UTC().Truncate(time.Hour)
	first := hours - 1 - int(current.Sub(created.UTC().Truncate(time.Hour))/time.Hour)
	ticks := make([]*string, hours)
	for i := max(first, 0); i < hours; i++ {
		ticks[i] = tick
	}
	return ticks
}

// loadHourResults collects the org's hourly check counts from every monitor
// type that keeps a check history. A cron incident hour counts as one failed
// check, since a missed ping leaves no row of its own.
func (h *MonitorHandler) loadHourResults(ctx context.Context, orgID uuid.UUID, hours int32) ([]hourResult, error) {
	var results []hourResult

	uptime, err := h.queries.ListUptimeCheckHistory(ctx, db.ListUptimeCheckHistoryParams{OrgID: orgID, Hours: hours})
	if err != nil {
		return nil, fmt.Errorf("uptime: %w", err)
	}
	for _, row := range uptime {
		results = append(results, hourResult{"uptime", row.MonitorID, row.Hour.Time, row.Checks, row.Failures})
	}

	port, err := h.queries.ListPortCheckHistory(ctx, db.ListPortCheckHistoryParams{OrgID: orgID, Hours: hours})
	if err != nil {
		return nil, fmt.Errorf("port: %w", err)
	}
	for _, row := range port {
		results = append(results, hourResult{"port", row.MonitorID, row.Hour.Time, row.Checks, row.Failures})
	}

	dns, err := h.queries.ListDNSCheckHistory(ctx, db.ListDNSCheckHistoryParams{OrgID: orgID, Hours: hours})
	if err != nil {
		return nil, fmt.Errorf("dns: %w", err)
	}
	for _, row := range dns {
		results = append(results, hourResult{"dns", row.MonitorID, row.Hour.Time, row.Checks, row.Failures})
	}

	pings, err := h.queries.ListCronPingHistory(ctx, db.ListCronPingHistoryParams{OrgID: orgID, Hours: hours})
	if err != nil {
		return nil, fmt.Errorf("cron pings: %w", err)
	}
	for _, row := range pings {
		results = append(results, hourResult{"cron", row.MonitorID, row.Hour.Time, row.Checks, row.Failures})
	}

	incidentHours, err := h.queries.ListCronIncidentHours(ctx, db.ListCronIncidentHoursParams{OrgID: orgID, Hours: hours})
	if err != nil {
		return nil, fmt.Errorf("cron incidents: %w", err)
	}
	for _, row := range incidentHours {
		results = append(results, hourResult{"cron", row.MonitorID, row.Hour.Time, 0, 1})
	}
	return results, nil
}
