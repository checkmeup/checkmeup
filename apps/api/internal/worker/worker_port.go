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
	"github.com/checkmeup/checkmeup/internal/deliver"
	"github.com/checkmeup/checkmeup/internal/httpsafe"
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
	dialer := n.TCPDialer
	if dialer == nil {
		dialer = portCheckDialer()
	}
	responseTimeMs, isUp, failureReason := performTCPCheck(m, dialer)
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
	return backUpAlert("port", m.Name, downtime, codeLine("Host", fmt.Sprintf("%s:%d", m.Host, m.Port)))
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
// incident, honoring the monitor's alerts-enabled flag. It runs once per
// incident, when its alert count is still 0, so max_alerts_per_incident
// can't apply here (ADR alerting).
func alertPortIncident(ctx context.Context, n Notifiers, m db.PortMonitor, inc db.PortIncident, failureReason string) {
	if !m.AlertsEnabled {
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
	hostLine := codeLine("Host", fmt.Sprintf("%s:%d", m.Host, m.Port))
	a := alertText{
		Event: "down", MonitorType: "port", Name: m.Name,
		Icon: "🔴", Headline: " is down",
		Details: []string{hostLine, "Reason: " + failureReason},
		Subject: "DOWN: " + m.Name,
		Reason:  failureReason,
		SMS:     fmt.Sprintf("%s is DOWN (%s)", m.Name, failureReason),
	}
	if m.ExpectedState == db.PortExpectedStateClosed {
		a.Icon, a.Headline, a.Details = "⚠️", ": port unexpectedly open", []string{hostLine}
		a.Subject = m.Name + ": port unexpectedly open"
		a.SMS = a.Subject
	}
	return buildAlert(a)
}

// sharedPortDialer is the *net.Dialer used for every real (non-test) port
// check, built once rather than per call — net.Dialer holds no per-dial
// mutable state, so one instance is safe to reuse across concurrent Dial
// calls. m.Host is user-supplied, so it goes through httpsafe.Dialer to
// block loopback/private/link-local/cloud-metadata targets (SSRF).
var sharedPortDialer = httpsafe.Dialer(deliver.Timeout)

func portCheckDialer() *net.Dialer {
	return sharedPortDialer
}

// performTCPCheck opens a raw TCP connection to the monitor's host:port —
// no data sent or received, no protocol handshake. The dial outcome is
// interpreted against the monitor's expected state: an "open" monitor wants
// a successful connect, a "closed" monitor wants the opposite (PRD port-monitoring).
// dialer is injected so tests can pass an unguarded dialer to reach a local
// listener — see portCheckDialer for the hardened dialer real checks use.
func performTCPCheck(m db.PortMonitor, dialer *net.Dialer) (responseTimeMs int64, isUp bool, failureReason string) {
	addr := net.JoinHostPort(m.Host, strconv.Itoa(int(m.Port)))
	start := time.Now()
	conn, err := dialer.Dial("tcp", addr)
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
