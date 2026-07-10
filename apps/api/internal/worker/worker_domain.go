package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

// ─── Domain monitor checks ───────────────────────────────────────────────────
//
// Mirrors the SSL monitor checks in worker_ssl.go (same 30/14/7-day threshold
// pattern via that file's sslCrossedThreshold, same flag-reset-on-renewal
// logic) — the only differences are the data source (RDAP lookup instead of
// a TLS handshake) and the field names (registrar instead of issuer).

func checkDomainMonitors(ctx context.Context, n Notifiers) {
	monitors, err := n.Queries.ListDueDomainMonitors(ctx)
	if err != nil {
		n.Logger.Error("domain worker: list due monitors", "err", err)
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
			checkOneDomainMonitor(ctx, n, m)
		}()
	}
	wg.Wait()
}

func checkOneDomainMonitor(ctx context.Context, n Notifiers, m db.DomainMonitor) {
	lookup, lookupErr := n.RDAP.Lookup(m.Domain)
	var daysLeft int
	if lookupErr == nil {
		daysLeft = int(time.Until(lookup.ExpiresAt).Hours() / 24)
	}
	result := buildDomainCheckResult(m, lookup.ExpiresAt, lookup.Registrar, daysLeft, lookupErr)

	if result.alert != nil {
		DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "domain", ID: m.ID}, *result.alert)
	}

	if _, err := n.Queries.UpdateDomainMonitorCheck(ctx, db.UpdateDomainMonitorCheckParams{
		ID:                  m.ID,
		Status:              result.status,
		ExpiresAt:           result.expiresAtParam,
		Registrar:           result.registrarParam,
		ErrorMsg:            result.errorMsgParam,
		Alerted30d:          result.alerted30d,
		Alerted14d:          result.alerted14d,
		Alerted7d:           result.alerted7d,
		ConsecutiveFailures: result.consecutiveFailures,
		AlertCount:          result.alertCount,
	}); err != nil {
		n.Logger.Error("domain worker: update check", "monitor_id", m.ID, "err", err)
	}
}

// domainCheckResult is the outcome of one RDAP lookup: the fields to persist
// on the monitor row, plus an alert to send when a threshold was freshly
// crossed (nil otherwise).
type domainCheckResult struct {
	status              db.DomainMonitorStatus
	expiresAtParam      pgtype.Timestamptz
	registrarParam      pgtype.Text
	errorMsgParam       pgtype.Text
	alerted30d          bool
	alerted14d          bool
	alerted7d           bool
	consecutiveFailures int32
	alertCount          int32
	alert               *AlertMessage
}

func buildDomainCheckResult(m db.DomainMonitor, expiresAt time.Time, registrar string, daysLeft int, checkErr error) domainCheckResult {
	r := domainCheckResult{alerted30d: m.Alerted30d, alerted14d: m.Alerted14d, alerted7d: m.Alerted7d, alertCount: m.AlertCount}

	if checkErr != nil {
		r.status = db.DomainMonitorStatusError
		r.errorMsgParam = pgtype.Text{String: checkErr.Error(), Valid: true}
		r.consecutiveFailures = m.ConsecutiveFailures + 1
		return r
	}

	r.expiresAtParam = pgtype.Timestamptz{Time: expiresAt, Valid: true}
	r.registrarParam = pgtype.Text{String: registrar, Valid: true}

	switch {
	case daysLeft < 0:
		r.status = db.DomainMonitorStatusExpired
	case daysLeft <= 30:
		r.status = db.DomainMonitorStatusExpiringSoon
	default:
		r.status = db.DomainMonitorStatusUp
		// Reset flags, failure counter, and alert count when domain is renewed.
		r.alerted30d, r.alerted14d, r.alerted7d = false, false, false
		r.consecutiveFailures = 0
		r.alertCount = 0
		return r
	}

	r.consecutiveFailures = m.ConsecutiveFailures + 1

	maxAlerts := m.MaxAlertsPerIncident
	withinLimit := maxAlerts == 0 || m.AlertCount < maxAlerts
	if m.AlertsEnabled && withinLimit && r.consecutiveFailures > m.AlertAfterNFailures {
		r.alert = domainThresholdAlert(m, daysLeft, expiresAt, &r.alerted30d, &r.alerted14d, &r.alerted7d)
		if r.alert != nil {
			r.alertCount = m.AlertCount + 1
		}
	}
	return r
}

// domainThresholdAlert returns the alert for the first newly-crossed expiry
// threshold (expired, then 7/14/30 days), setting the matching alerted flag
// so checkOneDomainMonitor only sends one notification per crossing.
func domainThresholdAlert(m db.DomainMonitor, daysLeft int, expiresAt time.Time, alerted30d, alerted14d, alerted7d *bool) *AlertMessage {
	if !sslCrossedThreshold(daysLeft, alerted30d, alerted14d, alerted7d) {
		return nil
	}

	expiresStr := expiresAt.Format("2006-01-02")
	var subject, telegramMsg, emailHTML string
	if daysLeft < 0 {
		subject, telegramMsg, emailHTML = domainExpiredMessages(m, expiresStr)
	} else {
		subject, telegramMsg, emailHTML = domainExpiringSoonMessages(m, daysLeft, expiresStr)
	}

	return &AlertMessage{
		Telegram:     telegramMsg,
		EmailSubject: subject,
		EmailHTML:    emailHTML,
		Webhook: &webhook.Event{
			EventType:   "down",
			MonitorName: m.Name,
			MonitorType: "domain",
			Reason:      subject,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.DownMessage(m.Name, "domain", subject)),
		SMS:   TruncateSMS("Checkmeup: " + subject),
	}
}

// domainExpiredMessages builds the alert text for a domain registration that
// has already lapsed (daysLeft < 0), distinct from
// domainExpiringSoonMessages' "expires in N days" phrasing.
func domainExpiredMessages(m db.DomainMonitor, expiresStr string) (subject, telegramMsg, emailHTML string) {
	subject = fmt.Sprintf("DOWN: %s domain registration expired", m.Name)
	telegramMsg = fmt.Sprintf("🔴 <b>%s</b> domain registration has <b>expired</b>\n\nDomain: <code>%s</code>\nExpired: %s",
		m.Name, m.Domain, expiresStr)
	emailHTML = fmt.Sprintf("<p>🔴 <b>%s</b> domain registration has <b>expired</b></p><p>Domain: <code>%s</code><br>Expired: %s</p>",
		m.Name, m.Domain, expiresStr)
	return subject, telegramMsg, emailHTML
}

func domainExpiringSoonMessages(m db.DomainMonitor, daysLeft int, expiresStr string) (subject, telegramMsg, emailHTML string) {
	subject = fmt.Sprintf("%s domain registration expires in %d days", m.Name, daysLeft)
	telegramMsg = fmt.Sprintf("🌐 <b>%s</b> domain registration expires in <b>%d days</b>\n\nDomain: <code>%s</code>\nExpires: %s",
		m.Name, daysLeft, m.Domain, expiresStr)
	emailHTML = fmt.Sprintf("<p>🌐 <b>%s</b> domain registration expires in <b>%d days</b></p><p>Domain: <code>%s</code><br>Expires: %s</p>",
		m.Name, daysLeft, m.Domain, expiresStr)
	return subject, telegramMsg, emailHTML
}
