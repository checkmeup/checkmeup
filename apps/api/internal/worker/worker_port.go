package worker

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
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
	trackCheck(ctx, n, portTracker{q: n.Queries, m: m, responseTimeMs: responseTimeMs, failureReason: failureReason}, isUp)
}

// portTracker is a port monitor's side of trackCheck, bound to one check's result.
type portTracker struct {
	q              *db.Queries
	m              db.PortMonitor
	responseTimeMs int64
	failureReason  string
}

func (t portTracker) monitor() trackedMonitor {
	return trackedMonitor{
		Type: "port", ID: t.m.ID, OrgID: t.m.OrgID, Status: t.m.Status,
		ConsecutiveFailures: t.m.ConsecutiveFailures, AlertAfterNFailures: t.m.AlertAfterNFailures,
		AlertsEnabled: t.m.AlertsEnabled,
	}
}

func (t portTracker) latestCheckFailed(ctx context.Context) (bool, error) {
	return t.q.LatestPortCheckFailed(ctx, t.m.ID)
}

func (t portTracker) recordCheck(ctx context.Context, isUp bool) error {
	_, err := t.q.CreatePortCheck(ctx, db.CreatePortCheckParams{
		MonitorID:      t.m.ID,
		ResponseTimeMs: int32(t.responseTimeMs),
		IsUp:           isUp,
		FailureReason:  pgtype.Text{String: t.failureReason, Valid: t.failureReason != ""},
	})
	return err
}

func (t portTracker) scheduleRecheck(ctx context.Context) error {
	return t.q.SchedulePortRecheck(ctx, t.m.ID)
}

func (t portTracker) recordUp(ctx context.Context) error {
	_, err := t.q.RecordPortCheckUp(ctx, t.m.ID)
	return err
}

func (t portTracker) recordFailure(ctx context.Context) (int32, error) {
	updated, err := t.q.RecordPortCheckFailure(ctx, t.m.ID)
	return updated.ConsecutiveFailures, err
}

func (t portTracker) openIncident(ctx context.Context) (uuid.UUID, error) {
	inc, err := t.q.OpenPortIncident(ctx, t.m.ID)
	return inc.ID, err
}

func (t portTracker) resolveIncident(ctx context.Context) (time.Time, error) {
	inc, err := t.q.ResolveLatestPortIncident(ctx, t.m.ID)
	return inc.StartedAt.Time, err
}

func (t portTracker) countAlert(ctx context.Context, incidentID uuid.UUID) error {
	_, err := t.q.IncrementPortIncidentAlertCount(ctx, incidentID)
	return err
}

func (t portTracker) downAlert() AlertMessage {
	return buildPortDownAlert(t.m, t.failureReason)
}

func (t portTracker) recoveryAlert(downtime string) AlertMessage {
	return buildPortRecoveryAlert(t.m, downtime)
}

func buildPortRecoveryAlert(m db.PortMonitor, downtime string) AlertMessage {
	return backUpAlert("port", m.Name, downtime, codeLine("Host", fmt.Sprintf("%s:%d", m.Host, m.Port)))
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
