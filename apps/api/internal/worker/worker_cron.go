package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/slack"
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

// checkStuckCronRuns implements PRD zombie-job-detection — detecting a cron run that
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
// re-fire — "once per run" per PRD zombie-job-detection) before dispatching, mirroring
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
	startedAt := run.StartedAt.Time.Format("15:04:05 MST")
	runningFor := FormatDuration(time.Since(run.StartedAt.Time))
	maxDuration := FormatDuration(time.Duration(run.MonitorMaxDurationMins.Int32) * time.Minute)
	return buildAlert(alertText{
		Event: "down", MonitorType: "cron", Name: run.MonitorName,
		Icon: "🐛", Headline: " has a stuck run",
		Details: []string{"Started at: " + startedAt, "Running for: " + runningFor, "Max expected: " + maxDuration},
		Subject: fmt.Sprintf("STUCK: %s run exceeded max duration", run.MonitorName),
		Reason:  fmt.Sprintf("stuck run — started at %s, running for %s (max %s)", startedAt, runningFor, maxDuration),
		SMS:     fmt.Sprintf("%s stuck run (%s)", run.MonitorName, runningFor),
	})
}

func buildOverdueAlert(m db.CronMonitor) AlertMessage {
	missedBy := time.Since(m.NextPingAt.Time).Round(time.Second)
	expectedAt := m.NextPingAt.Time.Format("15:04:05 MST")
	reason := fmt.Sprintf("missed its ping — expected at %s, missed by %s", expectedAt, missedBy)
	return buildAlert(alertText{
		Event: "down", MonitorType: "cron", Name: m.Name,
		Icon: "🔴", Headline: " missed its ping",
		Details: []string{codeLine("Schedule", m.Schedule), "Expected at: " + expectedAt, fmt.Sprintf("Missed by: %s", missedBy)},
		Subject: fmt.Sprintf("DOWN: %s missed its ping", m.Name),
		Reason:  reason,
		SMS:     fmt.Sprintf("%s is DOWN (%s)", m.Name, reason),
	})
}

// CronRecoveryAlert is the alert for a down cron monitor pinging again —
// sent from the ping handler, which is where a cron monitor recovers.
func CronRecoveryAlert(name, downtime string) AlertMessage {
	return buildAlert(alertText{
		Event: "recovery", MonitorType: "cron", Name: name,
		Icon: "✅", Headline: " recovered",
		Details:  []string{"Down for: " + downtime},
		Subject:  name + " recovered",
		Downtime: downtime,
		SMS:      fmt.Sprintf("%s recovered after %s", name, downtime),
	})
}

// StuckRunRecoveryAlert is the alert for a stuck run that finished or was
// superseded by a new start ping; reason says which.
func StuckRunRecoveryAlert(name, reason, ranFor string) AlertMessage {
	msg := buildAlert(alertText{
		Event: "recovery", MonitorType: "cron", Name: name,
		Icon: "✅", Headline: " stuck run recovered — " + reason,
		Details:  []string{"Ran for: " + ranFor},
		Subject:  name + " stuck run recovered",
		Downtime: ranFor,
		SMS:      fmt.Sprintf("%s stuck run recovered (%s)", name, reason),
	})
	// Slack says why the run stopped being stuck along with how long it ran.
	slackMessage := slack.RecoveryMessage(name, "cron", fmt.Sprintf("%s (ran for %s)", reason, ranFor))
	msg.Slack = &slackMessage
	return msg
}
