package worker

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/deliver"
	"github.com/checkmeup/checkmeup/internal/httpsafe"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

// ─── SSL monitor checks ──────────────────────────────────────────────────────

func checkSSLMonitors(ctx context.Context, n Notifiers) {
	monitors, err := n.Queries.ListDueSSLMonitors(ctx)
	if err != nil {
		n.Logger.Error("ssl worker: list due monitors", "err", err)
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
			checkOneSSLMonitor(ctx, n, m)
		}()
	}
	wg.Wait()
}

func checkOneSSLMonitor(ctx context.Context, n Notifiers, m db.SslMonitor) {
	expiresAt, issuer, daysLeft, err := performTLSCheck(m.Hostname)
	result := buildSSLCheckResult(m, expiresAt, issuer, daysLeft, err)

	if result.alert != nil {
		DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "ssl", ID: m.ID}, *result.alert)
	}

	if _, err := n.Queries.UpdateSSLMonitorCheck(ctx, db.UpdateSSLMonitorCheckParams{
		ID:                  m.ID,
		Status:              result.status,
		ExpiresAt:           result.expiresAtParam,
		Issuer:              result.issuerParam,
		ErrorMsg:            result.errorMsgParam,
		Alerted30d:          result.alerted30d,
		Alerted14d:          result.alerted14d,
		Alerted7d:           result.alerted7d,
		ConsecutiveFailures: result.consecutiveFailures,
		AlertCount:          result.alertCount,
	}); err != nil {
		n.Logger.Error("ssl worker: update check", "monitor_id", m.ID, "err", err)
	}
}

// sslCheckResult is the outcome of one performTLSCheck poll: the fields to
// persist on the monitor row, plus an alert to send when a threshold was
// freshly crossed (nil otherwise).
type sslCheckResult struct {
	status              db.SslMonitorStatus
	expiresAtParam      pgtype.Timestamptz
	issuerParam         pgtype.Text
	errorMsgParam       pgtype.Text
	alerted30d          bool
	alerted14d          bool
	alerted7d           bool
	consecutiveFailures int32
	alertCount          int32
	alert               *AlertMessage
}

func buildSSLCheckResult(m db.SslMonitor, expiresAt time.Time, issuer string, daysLeft int, checkErr error) sslCheckResult {
	r := sslCheckResult{alerted30d: m.Alerted30d, alerted14d: m.Alerted14d, alerted7d: m.Alerted7d, alertCount: m.AlertCount}

	if checkErr != nil {
		r.status = db.SslMonitorStatusError
		r.errorMsgParam = pgtype.Text{String: checkErr.Error(), Valid: true}
		r.consecutiveFailures = m.ConsecutiveFailures + 1
		return r
	}

	r.expiresAtParam = pgtype.Timestamptz{Time: expiresAt, Valid: true}
	r.issuerParam = pgtype.Text{String: issuer, Valid: true}

	switch {
	case daysLeft < 0:
		r.status = db.SslMonitorStatusExpired
	case daysLeft <= 30:
		r.status = db.SslMonitorStatusExpiringSoon
	default:
		r.status = db.SslMonitorStatusUp
		// Reset flags, failure counter, and alert count when cert is renewed.
		r.alerted30d, r.alerted14d, r.alerted7d = false, false, false
		r.consecutiveFailures = 0
		r.alertCount = 0
		return r
	}

	r.consecutiveFailures = m.ConsecutiveFailures + 1

	maxAlerts := m.MaxAlertsPerIncident
	withinLimit := maxAlerts == 0 || m.AlertCount < maxAlerts
	if m.AlertsEnabled && withinLimit && r.consecutiveFailures > m.AlertAfterNFailures {
		r.alert = sslThresholdAlert(m, daysLeft, expiresAt, &r.alerted30d, &r.alerted14d, &r.alerted7d)
		if r.alert != nil {
			r.alertCount = m.AlertCount + 1
		}
	}
	return r
}

// sslThresholdAlert returns the alert for the first newly-crossed expiry
// threshold (expired, then 7/14/30 days), setting the matching alerted flag
// so checkOneSSLMonitor only sends one notification per crossing.
func sslThresholdAlert(m db.SslMonitor, daysLeft int, expiresAt time.Time, alerted30d, alerted14d, alerted7d *bool) *AlertMessage {
	if !sslCrossedThreshold(daysLeft, alerted30d, alerted14d, alerted7d) {
		return nil
	}

	expiresStr := expiresAt.Format("2006-01-02")
	var subject, telegramMsg, emailHTML string
	if daysLeft < 0 {
		subject, telegramMsg, emailHTML = sslExpiredMessages(m, expiresStr)
	} else {
		subject, telegramMsg, emailHTML = sslExpiringSoonMessages(m, daysLeft, expiresStr)
	}

	return &AlertMessage{
		Telegram:     telegramMsg,
		EmailSubject: subject,
		EmailHTML:    emailHTML,
		Webhook: &webhook.Event{
			EventType:   "down",
			MonitorName: m.Name,
			MonitorType: "ssl",
			Reason:      subject,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.DownMessage(m.Name, "ssl", subject)),
		SMS:   TruncateSMS("Checkmeup: " + subject),
	}
}

// sslCrossedThreshold reports whether daysLeft just crossed a new expiry
// threshold (expired/7/14/30 days) and, if so, sets the matching alerted
// flag so the crossing only fires once. Expired and the 7-day threshold
// share alerted7d since expiring is a subset of the 7-day window.
//
// Shared with domainThresholdAlert in worker_domain.go — domain monitors
// mirror the same 30/14/7-day threshold pattern.
func sslCrossedThreshold(daysLeft int, alerted30d, alerted14d, alerted7d *bool) bool {
	switch {
	case daysLeft <= 7 && !*alerted7d:
		*alerted7d = true
		return true
	case daysLeft <= 14 && !*alerted14d:
		*alerted14d = true
		return true
	case daysLeft <= 30 && !*alerted30d:
		*alerted30d = true
		return true
	default:
		return false
	}
}

// sslExpiredMessages builds the alert text for a certificate that has
// already expired (daysLeft < 0), distinct from sslExpiringSoonMessages'
// "expires in N days" phrasing.
func sslExpiredMessages(m db.SslMonitor, expiresStr string) (subject, telegramMsg, emailHTML string) {
	subject = fmt.Sprintf("DOWN: %s SSL certificate expired", m.Name)
	telegramMsg = fmt.Sprintf("🔴 <b>%s</b> SSL certificate has <b>expired</b>\n\nHost: <code>%s</code>\nExpired: %s",
		m.Name, m.Hostname, expiresStr)
	emailHTML = fmt.Sprintf("<p>🔴 <b>%s</b> SSL certificate has <b>expired</b></p><p>Host: <code>%s</code><br>Expired: %s</p>",
		m.Name, m.Hostname, expiresStr)
	return subject, telegramMsg, emailHTML
}

func sslExpiringSoonMessages(m db.SslMonitor, daysLeft int, expiresStr string) (subject, telegramMsg, emailHTML string) {
	subject = fmt.Sprintf("%s SSL certificate expires in %d days", m.Name, daysLeft)
	telegramMsg = fmt.Sprintf("🔐 <b>%s</b> SSL certificate expires in <b>%d days</b>\n\nHost: <code>%s</code>\nExpires: %s",
		m.Name, daysLeft, m.Hostname, expiresStr)
	emailHTML = fmt.Sprintf("<p>🔐 <b>%s</b> SSL certificate expires in <b>%d days</b></p><p>Host: <code>%s</code><br>Expires: %s</p>",
		m.Name, daysLeft, m.Hostname, expiresStr)
	return subject, telegramMsg, emailHTML
}

// sharedSSLDialer is the *net.Dialer used for every SSL check, built once
// rather than per call — net.Dialer holds no per-dial mutable state, so one
// instance is safe to reuse across concurrent Dial calls. hostname is
// user-supplied, so it goes through httpsafe.Dialer to block
// loopback/private/link-local/cloud-metadata targets (SSRF).
var sharedSSLDialer = httpsafe.Dialer(deliver.Timeout)

func performTLSCheck(hostname string) (expiresAt time.Time, issuer string, daysLeft int, err error) {
	conn, err := tls.DialWithDialer(sharedSSLDialer, "tcp", hostname+":443", &tls.Config{
		ServerName: hostname,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return time.Time{}, "", 0, err
	}
	defer conn.Close() //nolint:errcheck // TLS close errors are not actionable

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return time.Time{}, "", 0, fmt.Errorf("no peer certificates")
	}

	leaf := certs[0]
	expiresAt = leaf.NotAfter
	issuer = leaf.Issuer.CommonName
	daysLeft = int(time.Until(expiresAt).Hours() / 24)
	return expiresAt, issuer, daysLeft, nil
}
