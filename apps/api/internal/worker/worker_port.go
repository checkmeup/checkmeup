package worker

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

// ─── port monitor checks ─────────────────────────────────────────────────────

func checkPortMonitors(ctx context.Context, n Notifiers) {
	monitors, err := n.Queries.ListDuePortMonitors(ctx)
	if err != nil {
		n.Logger.Error("port worker: list due monitors", "err", err)
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
			checkOnePortMonitor(ctx, n, m)
		}()
	}
	wg.Wait()
}

func checkOnePortMonitor(ctx context.Context, n Notifiers, m db.PortMonitor) {
	responseTimeMs, isUp, failureReason := performTCPCheck(m)
	if !recordPortCheck(ctx, n, m, responseTimeMs, isUp, failureReason) {
		return
	}

	prevStatus := m.Status
	if isUp {
		handlePortUp(ctx, n, m, prevStatus)
		return
	}
	handlePortDown(ctx, n, m, prevStatus, failureReason)
}

func recordPortCheck(ctx context.Context, n Notifiers, m db.PortMonitor, responseTimeMs int64, isUp bool, failureReason string) bool {
	if _, err := n.Queries.CreatePortCheck(ctx, db.CreatePortCheckParams{
		MonitorID:      m.ID,
		ResponseTimeMs: int32(responseTimeMs),
		IsUp:           isUp,
		FailureReason:  pgtype.Text{String: failureReason, Valid: failureReason != ""},
	}); err != nil {
		n.Logger.Error("port worker: create check", "monitor_id", m.ID, "err", err)
		return false
	}
	return true
}

func handlePortUp(ctx context.Context, n Notifiers, m db.PortMonitor, prevStatus db.MonitorStatus) {
	if _, err := n.Queries.RecordPortCheckUp(ctx, m.ID); err != nil {
		n.Logger.Error("port worker: record up", "monitor_id", m.ID, "err", err)
		return
	}
	if prevStatus != db.MonitorStatusDown {
		return
	}
	inc, err := n.Queries.ResolveLatestPortIncident(ctx, m.ID)
	if err != nil {
		n.Logger.Error("port worker: resolve incident", "monitor_id", m.ID, "err", err)
		return
	}
	if !m.AlertsEnabled {
		return
	}
	downtime := FormatDuration(time.Since(inc.StartedAt.Time))
	DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "port", ID: m.ID}, buildPortRecoveryAlert(m, downtime))
}

func buildPortRecoveryAlert(m db.PortMonitor, downtime string) AlertMessage {
	hostPort := fmt.Sprintf("%s:%d", m.Host, m.Port)
	return AlertMessage{
		Telegram:     fmt.Sprintf("✅ <b>%s</b> is back up\n\nHost: <code>%s</code>", m.Name, hostPort),
		EmailSubject: fmt.Sprintf("%s recovered", m.Name),
		EmailHTML:    fmt.Sprintf("<p>✅ <b>%s</b> is back up</p><p>Host: <code>%s</code></p>", m.Name, hostPort),
		Webhook: &webhook.Event{
			EventType:        "recovery",
			MonitorName:      m.Name,
			MonitorType:      "port",
			DowntimeDuration: downtime,
			Timestamp:        time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.RecoveryMessage(m.Name, "port", downtime)),
		SMS:   TruncateSMS(fmt.Sprintf("checkmeup: %s recovered after %s downtime", m.Name, downtime)),
	}
}

func handlePortDown(ctx context.Context, n Notifiers, m db.PortMonitor, prevStatus db.MonitorStatus, failureReason string) {
	updated, err := n.Queries.RecordPortCheckFailure(ctx, m.ID)
	if err != nil {
		n.Logger.Error("port worker: record failure", "monitor_id", m.ID, "err", err)
		return
	}
	if updated.ConsecutiveFailures <= m.AlertAfterNFailures || prevStatus == db.MonitorStatusDown {
		return
	}
	if err := n.Queries.MarkPortMonitorDown(ctx, m.ID); err != nil {
		n.Logger.Error("port worker: mark down", "monitor_id", m.ID, "err", err)
		return
	}
	inc, err := n.Queries.CreatePortIncident(ctx, m.ID)
	if err != nil {
		n.Logger.Error("port worker: create incident", "monitor_id", m.ID, "err", err)
		return
	}
	alertPortIncident(ctx, n, m, inc, failureReason)
}

// alertPortIncident dispatches the down alert for a freshly-recorded
// incident, honoring the monitor's alerts-enabled flag and per-incident cap.
func alertPortIncident(ctx context.Context, n Notifiers, m db.PortMonitor, inc db.PortIncident, failureReason string) {
	if !m.AlertsEnabled {
		return
	}
	if max := m.MaxAlertsPerIncident; max > 0 && inc.AlertCount >= max {
		return
	}
	if !DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "port", ID: m.ID}, buildPortDownAlert(m, failureReason)) {
		return
	}
	if _, err := n.Queries.IncrementPortIncidentAlertCount(ctx, inc.ID); err != nil {
		n.Logger.Error("port worker: increment alert count", "incident_id", inc.ID, "err", err)
	}
}

// buildPortDownAlert phrases the headline around expected_state: a
// closed-state monitor going "down" means the port unexpectedly became
// reachable, which reads very differently from an open-state monitor's
// service outage.
func buildPortDownAlert(m db.PortMonitor, failureReason string) AlertMessage {
	hostPort := fmt.Sprintf("%s:%d", m.Host, m.Port)
	if m.ExpectedState == db.PortExpectedStateClosed {
		subject := fmt.Sprintf("%s: port unexpectedly open", m.Name)
		return AlertMessage{
			Telegram: fmt.Sprintf("⚠️ <b>%s</b>: port unexpectedly open\n\nHost: <code>%s</code>",
				m.Name, hostPort),
			EmailSubject: subject,
			EmailHTML: fmt.Sprintf("<p>⚠️ <b>%s</b>: port unexpectedly open</p><p>Host: <code>%s</code></p>",
				m.Name, hostPort),
			Webhook: &webhook.Event{
				EventType:   "down",
				MonitorName: m.Name,
				MonitorType: "port",
				Reason:      failureReason,
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
			},
			Slack: slackMsg(slack.DownMessage(m.Name, "port", failureReason)),
			SMS:   TruncateSMS(fmt.Sprintf("checkmeup: %s", subject)),
		}
	}
	return AlertMessage{
		Telegram: fmt.Sprintf("🔴 <b>%s</b> is down\n\nHost: <code>%s</code>\nReason: %s",
			m.Name, hostPort, failureReason),
		EmailSubject: fmt.Sprintf("DOWN: %s", m.Name),
		EmailHTML: fmt.Sprintf("<p>🔴 <b>%s</b> is down</p><p>Host: <code>%s</code><br>Reason: %s</p>",
			m.Name, hostPort, failureReason),
		Webhook: &webhook.Event{
			EventType:   "down",
			MonitorName: m.Name,
			MonitorType: "port",
			Reason:      failureReason,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.DownMessage(m.Name, "port", failureReason)),
		SMS:   TruncateSMS(fmt.Sprintf("checkmeup: %s is DOWN (%s)", m.Name, failureReason)),
	}
}

// performTCPCheck opens a raw TCP connection to the monitor's host:port —
// no data sent or received, no protocol handshake. The dial outcome is
// interpreted against the monitor's expected state: an "open" monitor wants
// a successful connect, a "closed" monitor wants the opposite (US-3302).
func performTCPCheck(m db.PortMonitor) (responseTimeMs int64, isUp bool, failureReason string) {
	addr := net.JoinHostPort(m.Host, strconv.Itoa(int(m.Port)))
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	elapsed := time.Since(start).Milliseconds()
	connected := err == nil
	if connected {
		_ = conn.Close()
	}

	if m.ExpectedState == db.PortExpectedStateClosed {
		if connected {
			return elapsed, false, "port is unexpectedly open"
		}
		return elapsed, true, ""
	}
	if connected {
		return elapsed, true, ""
	}
	return elapsed, false, "connection refused / timeout"
}
