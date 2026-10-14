package handler

import (
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
	// when no check ran.
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
	ctx := r.Context()
	hours := int32(checkHistoryHours)
	var results []hourResult

	uptime, err := h.queries.ListUptimeCheckHistory(ctx, db.ListUptimeCheckHistoryParams{OrgID: orgID, Hours: hours})
	if err != nil {
		slog.ErrorContext(ctx, "list uptime check history failed", "err", err)
		respond.InternalError(w)
		return
	}
	for _, row := range uptime {
		results = append(results, hourResult{"uptime", row.MonitorID, row.Hour.Time, row.Checks, row.Failures})
	}

	port, err := h.queries.ListPortCheckHistory(ctx, db.ListPortCheckHistoryParams{OrgID: orgID, Hours: hours})
	if err != nil {
		slog.ErrorContext(ctx, "list port check history failed", "err", err)
		respond.InternalError(w)
		return
	}
	for _, row := range port {
		results = append(results, hourResult{"port", row.MonitorID, row.Hour.Time, row.Checks, row.Failures})
	}

	dns, err := h.queries.ListDNSCheckHistory(ctx, db.ListDNSCheckHistoryParams{OrgID: orgID, Hours: hours})
	if err != nil {
		slog.ErrorContext(ctx, "list dns check history failed", "err", err)
		respond.InternalError(w)
		return
	}
	for _, row := range dns {
		results = append(results, hourResult{"dns", row.MonitorID, row.Hour.Time, row.Checks, row.Failures})
	}

	pings, err := h.queries.ListCronPingHistory(ctx, db.ListCronPingHistoryParams{OrgID: orgID, Hours: hours})
	if err != nil {
		slog.ErrorContext(ctx, "list cron ping history failed", "err", err)
		respond.InternalError(w)
		return
	}
	for _, row := range pings {
		results = append(results, hourResult{"cron", row.MonitorID, row.Hour.Time, row.Checks, row.Failures})
	}

	incidentHours, err := h.queries.ListCronIncidentHours(ctx, db.ListCronIncidentHoursParams{OrgID: orgID, Hours: hours})
	if err != nil {
		slog.ErrorContext(ctx, "list cron incident hours failed", "err", err)
		respond.InternalError(w)
		return
	}
	for _, row := range incidentHours {
		results = append(results, hourResult{"cron", row.MonitorID, row.Hour.Time, 0, 1})
	}

	respond.JSON(w, http.StatusOK, checkHistoryResponse{
		Hours:    checkHistoryHours,
		Monitors: buildCheckHistory(time.Now(), checkHistoryHours, results),
	})
}
