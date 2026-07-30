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

// checkStuckCronRuns implements EP-34 US-3403 — detecting a cron run that
// started but has exceeded its monitor's max_duration_mins without a
// completion ping ("zombie job"). Distinct from checkOverdue: this fires on
// an in-progress run exceeding a duration budget, not a missed ping.
func checkStuckCronRuns(ctx context.Context, n Notifiers) {
	runs, err := n.Queries.ListStuckCronRuns(ctx)
	if err != nil {
		n.Logger.Error("worker: list stuck cron runs", "err", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, checkConcurrency)
	for _, run := range runs {
		run := run
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			processStuckCronRun(ctx, n, run)
		}()
	}
	wg.Wait()
}

// processStuckCronRun marks the run alerted (so the next tick doesn't
// re-fire — "once per run" per US-3403) before dispatching, mirroring
// processOverdueMonitor's mark-then-alert order.
func processStuckCronRun(ctx context.Context, n Notifiers, run db.ListStuckCronRunsRow) {
	if _, err := n.Queries.MarkCronRunAlerted(ctx, run.ID); err != nil {
		n.Logger.Error("worker: mark cron run alerted", "run_id", run.ID, "err", err)
		return
	}
	if !run.MonitorAlertsEnabled {
		return
	}
	msg := buildStuckRunAlert(run)
	DispatchAlert(ctx, n, run.MonitorOrgID, MonitorRef{Type: "cron", ID: run.MonitorID}, msg)
}

func buildStuckRunAlert(run db.ListStuckCronRunsRow) AlertMessage {
	startedAt := run.StartedAt.Time
	runningFor := FormatDuration(time.Since(startedAt))
	maxDuration := FormatDuration(time.Duration(run.MonitorMaxDurationMins.Int32) * time.Minute)
	reason := fmt.Sprintf("stuck run — started at %s, running for %s (max %s)", startedAt.Format("15:04:05 MST"), runningFor, maxDuration)
	return AlertMessage{
		Telegram: fmt.Sprintf(
			"🐛 <b>%s</b> has a stuck run\n\nStarted at: %s\nRunning for: %s\nMax expected: %s",
			run.MonitorName, startedAt.Format("15:04:05 MST"), runningFor, maxDuration,
		),
		EmailSubject: fmt.Sprintf("STUCK: %s run exceeded max duration", run.MonitorName),
		EmailHTML: fmt.Sprintf(
			"<p>🐛 <b>%s</b> has a stuck run</p><p>Started at: %s<br>Running for: %s<br>Max expected: %s</p>",
			run.MonitorName, startedAt.Format("15:04:05 MST"), runningFor, maxDuration,
		),
		Webhook: &webhook.Event{
			EventType:   "down",
			MonitorName: run.MonitorName,
			MonitorType: "cron",
			Reason:      reason,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.DownMessage(run.MonitorName, "cron", reason)),
		SMS:   TruncateSMS(fmt.Sprintf("Checkmeup: %s stuck run (%s)", run.MonitorName, runningFor)),
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
		SMS:   TruncateSMS(fmt.Sprintf("Checkmeup: %s is DOWN (%s)", m.Name, reason)),
	}
}
