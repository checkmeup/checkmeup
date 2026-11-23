package worker

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/checkmeup/checkmeup/internal/db"
)

// The incident lifecycle uptime, port and DNS monitors share: what a check's
// result does to the monitor, its incident and its alerts (ADR alerting).
// Each type supplies an incidentTracker, its queries bound to the monitor
// being checked and the check's result, and its alert wording. Cron, SSL and
// domain monitors alert more than once per incident, so they keep their own.

// trackedMonitor is what trackCheck reads of a monitor, as it was when the
// check began.
type trackedMonitor struct {
	Type                string // "uptime", "port" or "dns"
	ID                  uuid.UUID
	OrgID               uuid.UUID
	Status              db.MonitorStatus
	ConsecutiveFailures int32
	AlertAfterNFailures int32
	AlertsEnabled       bool
}

type incidentTracker interface {
	monitor() trackedMonitor
	latestCheckFailed(ctx context.Context) (bool, error)
	recordCheck(ctx context.Context, isUp bool) error
	scheduleRecheck(ctx context.Context) error
	recordUp(ctx context.Context) error
	// recordFailure returns the monitor's consecutive failures, this one counted.
	recordFailure(ctx context.Context) (int32, error)
	// openIncident marks the monitor down and opens its incident, both or
	// neither, keeping alert, the down alert as JSON, for its retries.
	openIncident(ctx context.Context, alert []byte) (uuid.UUID, error)
	// resolveIncident closes the monitor's open incident and returns when it started.
	resolveIncident(ctx context.Context) (time.Time, error)
	countAlert(ctx context.Context, incidentID uuid.UUID) error
	downAlert() AlertMessage
	recoveryAlert(downtime string) AlertMessage
}

// trackCheck records a check that found the monitor up or not, and moves the
// monitor and its incident on from it.
func trackCheck(ctx context.Context, n Notifiers, t incidentTracker, isUp bool) {
	m := t.monitor()
	// A failure right after a passing check is checked again about 30 seconds
	// later before it counts (ADR alerting): one request lost between this
	// server and the target shouldn't alert anyone. Asked before this check is
	// recorded, so the previous check is the one read.
	unconfirmed := !isUp && m.ConsecutiveFailures == 0 && !latestCheckFailed(ctx, n, t, m)
	if err := t.recordCheck(ctx, isUp); err != nil {
		m.logError(n, "create check", err)
		return
	}
	if unconfirmed {
		if err := t.scheduleRecheck(ctx); err != nil {
			m.logError(n, "schedule recheck", err)
		}
		return
	}
	if isUp {
		trackUp(ctx, n, t, m)
		return
	}
	trackDown(ctx, n, t, m)
}

// latestCheckFailed reports whether the monitor's last recorded check failed.
// On a read error it says true, so the failure counts right away, as it did
// before rechecks: a missed alert is worse than an early one.
func latestCheckFailed(ctx context.Context, n Notifiers, t incidentTracker, m trackedMonitor) bool {
	failed, err := t.latestCheckFailed(ctx)
	if err != nil {
		m.logError(n, "read latest check", err)
		return true
	}
	return failed
}

func trackUp(ctx context.Context, n Notifiers, t incidentTracker, m trackedMonitor) {
	if err := t.recordUp(ctx); err != nil {
		m.logError(n, "record up", err)
		return
	}
	if m.Status != db.MonitorStatusDown {
		return
	}
	startedAt, err := t.resolveIncident(ctx)
	if err != nil {
		m.logError(n, "resolve incident", err)
		return
	}
	if !m.AlertsEnabled {
		return
	}
	downtime := FormatDuration(time.Since(startedAt))
	GoAlert(ctx, func(ctx context.Context) {
		DispatchAlert(ctx, n, m.OrgID, m.ref(), t.recoveryAlert(downtime))
	})
}

// trackDown opens an incident once the failures pass alert_after_n_failures.
// Its down alert is the incident's only one, sent while its alert count is
// still 0, so max_alerts_per_incident can't apply here (ADR alerting).
func trackDown(ctx context.Context, n Notifiers, t incidentTracker, m trackedMonitor) {
	failures, err := t.recordFailure(ctx)
	if err != nil {
		m.logError(n, "record failure", err)
		return
	}
	if failures <= m.AlertAfterNFailures || m.Status == db.MonitorStatusDown {
		return
	}
	msg := t.downAlert()
	// If this fails the monitor stays up, so the next failing check tries again.
	incidentID, err := t.openIncident(ctx, alertJSON(n, m.ID, msg))
	if err != nil {
		m.logError(n, "open incident", err)
		return
	}
	if !m.AlertsEnabled {
		return
	}
	GoAlert(ctx, func(ctx context.Context) {
		if !DispatchAlert(ctx, n, m.OrgID, m.ref(), msg) {
			return
		}
		if err := t.countAlert(ctx, incidentID); err != nil {
			n.Logger.Error(m.Type+" worker: increment alert count", "incident_id", incidentID, "err", err)
		}
	})
}

func (m trackedMonitor) ref() MonitorRef {
	return MonitorRef{Type: m.Type, ID: m.ID}
}

func (m trackedMonitor) logError(n Notifiers, step string, err error) {
	n.Logger.Error(m.Type+" worker: "+step, "monitor_id", m.ID, "err", err)
}
