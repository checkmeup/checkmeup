package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

// ─── cron monitor checks ─────────────────────────────────────────────────────

func checkOverdue(ctx context.Context, n Notifiers) {
	monitors, err := n.Queries.ListOverdueCronMonitors(ctx)
	if err != nil {
		n.Logger.Error("worker: list overdue monitors", "err", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, checkConcurrency)
	for _, m := range monitors {
		m := m
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			processOverdueMonitor(ctx, n, m)
		}()
	}
	wg.Wait()
}

func processOverdueMonitor(ctx context.Context, n Notifiers, m db.CronMonitor) {
	updated, err := n.Queries.IncrementCronConsecutiveFailures(ctx, m.ID)
	if err != nil {
		n.Logger.Error("worker: increment consecutive failures", "monitor_id", m.ID, "err", err)
		return
	}
	// Filter: suppress alert until more than N consecutive failures observed.
	// The monitor stays 'up' while filtering so the worker re-detects it every cycle.
	if updated.ConsecutiveFailures <= m.AlertAfterNFailures {
		return
	}

	if err := n.Queries.UpdateCronMonitorDown(ctx, m.ID); err != nil {
		n.Logger.Error("worker: mark down", "monitor_id", m.ID, "err", err)
		return
	}

	inc, err := n.Queries.CreateCronIncident(ctx, m.ID)
	if err != nil {
		n.Logger.Error("worker: create incident", "monitor_id", m.ID, "err", err)
		return
	}

	if !m.AlertsEnabled {
		return
	}
	if max := m.MaxAlertsPerIncident; max > 0 && inc.AlertCount >= max {
		return
	}

	msg := buildOverdueAlert(m)
	if !DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "cron", ID: m.ID}, msg) {
		return
	}
	if _, err := n.Queries.IncrementCronIncidentAlertCount(ctx, inc.ID); err != nil {
		n.Logger.Error("worker: increment alert count", "incident_id", inc.ID, "err", err)
	}
}

func buildOverdueAlert(m db.CronMonitor) AlertMessage {
	missedBy := time.Since(m.NextPingAt.Time).Round(time.Second)
	expectedAt := m.NextPingAt.Time.Format("15:04:05 MST")
	reason := fmt.Sprintf("missed its ping — expected at %s, missed by %s", expectedAt, missedBy)
	return AlertMessage{
		Telegram: fmt.Sprintf(
			"🔴 <b>%s</b> missed its ping\n\nSchedule: <code>%s</code>\nExpected at: %s\nMissed by: %s",
			m.Name, m.Schedule, expectedAt, missedBy,
		),
		EmailSubject: fmt.Sprintf("DOWN: %s missed its ping", m.Name),
		EmailHTML: fmt.Sprintf(
			"<p>🔴 <b>%s</b> missed its ping</p><p>Schedule: <code>%s</code><br>Expected at: %s<br>Missed by: %s</p>",
			m.Name, m.Schedule, expectedAt, missedBy,
		),
		Webhook: &webhook.Event{
			EventType:   "down",
			MonitorName: m.Name,
			MonitorType: "cron",
			Reason:      reason,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.DownMessage(m.Name, "cron", reason)),
		SMS:   TruncateSMS(fmt.Sprintf("checkmeup: %s is DOWN (%s)", m.Name, reason)),
	}
}
