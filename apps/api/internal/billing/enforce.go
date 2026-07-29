package billing

import (
	"context"

	"github.com/google/uuid"

	"github.com/checkmeup/checkmeup/internal/db"
)

// EnforceMonitorLimit pauses the newest active monitors (across all 6
// types) beyond limit, leaving the oldest `limit` active — called after a
// plan downgrade (ADR-019) so an org never has more than its new plan
// allows actually running, even though nothing is deleted. limit == -1
// (unlimited) is a no-op. Idempotent: a no-op when already at or under the
// limit, so it's safe to call unconditionally on every plan change, not
// just detected downgrades.
func EnforceMonitorLimit(ctx context.Context, q *db.Queries, orgID uuid.UUID, limit int) error {
	if limit == -1 {
		return nil
	}
	active, err := q.ListActiveMonitorsForOrg(ctx, orgID)
	if err != nil {
		return err
	}
	if len(active) <= limit {
		return nil
	}
	// active is newest-first (ORDER BY created_at DESC per the query), so
	// the first len(active)-limit entries are the newest — the overflow to
	// pause — leaving the oldest `limit` entries (the tail) active.
	toPause := active[:len(active)-limit]
	for _, m := range toPause {
		switch m.MonitorType {
		case "cron":
			_, err = q.PauseCronMonitor(ctx, db.PauseCronMonitorParams{ID: m.ID, OrgID: orgID})
		case "uptime":
			_, err = q.PauseUptimeMonitor(ctx, db.PauseUptimeMonitorParams{ID: m.ID, OrgID: orgID})
		case "ssl":
			_, err = q.PauseSSLMonitor(ctx, db.PauseSSLMonitorParams{ID: m.ID, OrgID: orgID})
		case "domain":
			_, err = q.PauseDomainMonitor(ctx, db.PauseDomainMonitorParams{ID: m.ID, OrgID: orgID})
		case "port":
			_, err = q.PausePortMonitor(ctx, db.PausePortMonitorParams{ID: m.ID, OrgID: orgID})
		case "dns":
			_, err = q.PauseDNSMonitor(ctx, db.PauseDNSMonitorParams{ID: m.ID, OrgID: orgID})
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// EnforceNotificationChannelLimit disables the newest enabled notification
// channels beyond limit, leaving the oldest `limit` enabled — same
// downgrade-time enforcement as EnforceMonitorLimit, for channels.
func EnforceNotificationChannelLimit(ctx context.Context, q *db.Queries, orgID uuid.UUID, limit int) error {
	if limit == -1 {
		return nil
	}
	channels, err := q.ListNotificationChannels(ctx, orgID) // newest-first
	if err != nil {
		return err
	}
	enabled := make([]db.NotificationChannel, 0, len(channels))
	for _, c := range channels {
		if c.Enabled {
			enabled = append(enabled, c)
		}
	}
	if len(enabled) <= limit {
		return nil
	}
	toDisable := enabled[:len(enabled)-limit]
	for _, c := range toDisable {
		if _, err := q.SetNotificationChannelEnabled(ctx, db.SetNotificationChannelEnabledParams{
			ID: c.ID, OrgID: orgID, Enabled: false,
		}); err != nil {
			return err
		}
	}
	return nil
}

// EnforceHideBrandingLimit clears status_pages.hide_branding across the org
// once it's no longer on a plan that allows it (ADR-035) — same
// downgrade-time enforcement as EnforceMonitorLimit/
// EnforceNotificationChannelLimit, for a boolean instead of a count. A no-op
// when the plan still allows it or no page has the flag set.
func EnforceHideBrandingLimit(ctx context.Context, q *db.Queries, orgID uuid.UUID, allowed bool) error {
	if allowed {
		return nil
	}
	return q.ClearHideBrandingForOrg(ctx, orgID)
}
