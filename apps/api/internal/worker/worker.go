package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

// Run starts the background worker loops. Returns when ctx is cancelled.
//   - Every 30 s: missed-ping detection — transitions overdue monitors to "down" and fires alerts.
//   - Every 24 h: ping retention cleanup — deletes cron_pings older than 30 days (ADR-015).
func Run(ctx context.Context, queries *db.Queries, tg *telegram.Client, logger *slog.Logger) {
	overdureTicker := time.NewTicker(30 * time.Second)
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer overdureTicker.Stop()
	defer cleanupTicker.Stop()
	logger.Info("cron worker started")
	for {
		select {
		case <-ctx.Done():
			logger.Info("cron worker stopped")
			return
		case <-overdureTicker.C:
			checkOverdue(ctx, queries, tg, logger)
		case <-cleanupTicker.C:
			pruneOldPings(ctx, queries, logger)
		}
	}
}

func pruneOldPings(ctx context.Context, queries *db.Queries, logger *slog.Logger) {
	if err := queries.DeleteOldCronPings(ctx); err != nil {
		logger.Error("worker: prune old pings", "err", err)
	} else {
		logger.Info("worker: pruned cron_pings older than 30 days")
	}
}

func checkOverdue(ctx context.Context, queries *db.Queries, tg *telegram.Client, logger *slog.Logger) {
	monitors, err := queries.ListOverdueCronMonitors(ctx)
	if err != nil {
		logger.Error("worker: list overdue monitors", "err", err)
		return
	}

	for _, m := range monitors {
		if err := queries.UpdateCronMonitorDown(ctx, m.ID); err != nil {
			logger.Error("worker: mark down", "monitor_id", m.ID, "err", err)
			continue
		}

		inc, err := queries.CreateCronIncident(ctx, m.ID)
		if err != nil {
			logger.Error("worker: create incident", "monitor_id", m.ID, "err", err)
			continue
		}

		if !m.AlertsEnabled {
			continue
		}

		max := m.MaxAlertsPerIncident
		if max > 0 && inc.AlertCount >= max {
			continue
		}

		org, err := queries.GetOrgByID(ctx, m.OrgID)
		if err != nil || !org.TelegramChatID.Valid {
			continue
		}

		missedBy := time.Since(m.NextPingAt.Time).Round(time.Second)
		msg := fmt.Sprintf(
			"🔴 <b>%s</b> missed its ping\n\nSchedule: <code>%s</code>\nExpected at: %s\nMissed by: %s",
			m.Name,
			m.Schedule,
			m.NextPingAt.Time.Format("15:04:05 MST"),
			missedBy,
		)
		if err := tg.SendMessage(org.TelegramChatID.String, msg); err != nil {
			logger.Error("worker: send down alert", "monitor_id", m.ID, "err", err)
			continue
		}
		if _, err := queries.IncrementCronIncidentAlertCount(ctx, inc.ID); err != nil {
			logger.Error("worker: increment alert count", "incident_id", inc.ID, "err", err)
		}
	}
}
