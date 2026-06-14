package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

// Run starts the missed-ping detection loop. It ticks every 30 seconds and
// transitions overdue monitors to "down", opens incidents, and fires alerts.
// Returns when ctx is cancelled.
func Run(ctx context.Context, queries *db.Queries, tg *telegram.Client, logger *slog.Logger) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	logger.Info("cron worker started")
	for {
		select {
		case <-ctx.Done():
			logger.Info("cron worker stopped")
			return
		case <-ticker.C:
			checkOverdue(ctx, queries, tg, logger)
		}
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

		if _, err := queries.CreateCronIncident(ctx, m.ID); err != nil {
			logger.Error("worker: create incident", "monitor_id", m.ID, "err", err)
		}

		if !m.AlertsEnabled {
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
		}
	}
}
