package worker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/rdap"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/webhook"
	"github.com/checkmeup/checkmeup/internal/whois"
)

// ─── Domain monitor checks ───────────────────────────────────────────────────
//
// Mirrors the SSL monitor checks in worker_ssl.go (same 30/14/7-day threshold
// pattern via that file's sslCrossedThreshold, same flag-reset-on-renewal
// logic) — the only differences are the data source (an RDAP or WHOIS lookup
// instead of a TLS handshake) and the field names (registrar instead of issuer).

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
	lookup, lookupErr := LookupDomain(n.RDAP, n.WHOIS, m.Domain)
	// Rechecked about 30 seconds later before it shows, as SSL errors are (ADR alerting).
	if lookupErr != nil && !m.RecheckPending && m.Status != db.DomainMonitorStatusError {
		if err := n.Queries.ScheduleDomainRecheck(ctx, m.ID); err != nil {
			n.Logger.Error("domain worker: schedule recheck", "monitor_id", m.ID, "err", err)
		}
		return
	}
	var daysLeft int
	if lookupErr == nil {
		daysLeft = int(time.Until(lookup.ExpiresAt).Hours() / 24)
	}
	result := buildDomainCheckResult(m, lookup.ExpiresAt, lookup.Registrar, daysLeft, lookupErr)

	if result.alert != nil {
		GoAlert(ctx, func(ctx context.Context) {
			DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "domain", ID: m.ID}, *result.alert)
		})
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

// LookupDomain reads a domain's registrar and expiry over RDAP, or over WHOIS
// for the TLDs whose registries have no RDAP service (ADR domain-lookups).
// A TLD whose registry publishes no expiry date fails with the reason. Also
// behind the public domain expiry checker (PRD free-tools).
func LookupDomain(rd *rdap.Client, wi *whois.Client, domain string) (rdap.Result, error) {
	if msg := whois.Unsupported(domain); msg != "" {
		return rdap.Result{}, errors.New(msg)
	}
	if whois.Supports(domain) && wi != nil {
		res, err := wi.Lookup(domain)
		if errors.Is(err, whois.ErrNotRegistered) {
			return rdap.Result{}, errors.New("domain not found in registry (whois)")
		}
		return rdap.Result{Registrar: res.Registrar, ExpiresAt: res.ExpiresAt}, err
	}
	return rd.Lookup(domain)
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
		// Keeps the last known expiry and alerts on it: a registry's RDAP or
		// WHOIS can be down for days, and until 2026-11-23 a domain expiring
		// then alerted nobody.
		if !m.ExpiresAt.Valid {
			return r
		}
		r.expiresAtParam, r.registrarParam = m.ExpiresAt, m.Registrar
		known := m.ExpiresAt.Time
		if domainDaysLeft(known, time.Now()) < 0 {
			// The status stays Error, so it can't tell whether this was sent:
			// sent on the check the known date passed, when the last one was
			// still before it.
			if m.AlertsEnabled && m.LastCheckedAt.Valid && domainDaysLeft(known, m.LastCheckedAt.Time) >= 0 {
				r.sendExpired(m, known)
			}
			return r
		}
		r.warn(m, domainDaysLeft(known, time.Now()), known)
		return r
	}

	r.expiresAtParam = pgtype.Timestamptz{Time: expiresAt, Valid: true}
	r.registrarParam = pgtype.Text{String: registrar, Valid: true}

	switch {
	case daysLeft < 0:
		r.status = db.DomainMonitorStatusExpired
		r.consecutiveFailures = m.ConsecutiveFailures + 1
		if m.AlertsEnabled && m.Status != db.DomainMonitorStatusExpired {
			r.sendExpired(m, expiresAt)
		}
		return r
	case daysLeft <= 30:
		r.status = db.DomainMonitorStatusExpiringSoon
		r.consecutiveFailures = m.ConsecutiveFailures + 1
		r.warn(m, daysLeft, expiresAt)
	default:
		r.status = db.DomainMonitorStatusUp
		// Reset flags, failure counter, and alert count when domain is renewed.
		r.alerted30d, r.alerted14d, r.alerted7d = false, false, false
		r.consecutiveFailures = 0
		r.alertCount = 0
	}
	return r
}

// domainDaysLeft is the whole days from at to expiresAt, as a check at that
// time counts them.
func domainDaysLeft(expiresAt, at time.Time) int {
	return int(expiresAt.Sub(at).Hours() / 24)
}

// sendExpired sends the expired alert past the cap and alert_after_n_failures
// (PRD domain-expiry-monitoring: "at once when expired"): the 7-day warning
// shares alerted7d with it, and the three warnings fill the default cap of 3.
func (r *domainCheckResult) sendExpired(m db.DomainMonitor, expiresAt time.Time) {
	subject, telegramMsg, emailHTML := domainExpiredMessages(m, expiresAt.Format("2006-01-02"))
	a := domainAlert(m, subject, telegramMsg, emailHTML)
	r.alert = &a
	r.alertCount = m.AlertCount + 1
	r.alerted30d, r.alerted14d, r.alerted7d = true, true, true
}

// warn sends the warning for a newly crossed 30, 14 or 7-day threshold, within
// the cap and once past alert_after_n_failures.
func (r *domainCheckResult) warn(m db.DomainMonitor, daysLeft int, expiresAt time.Time) {
	withinLimit := m.MaxAlertsPerIncident == 0 || m.AlertCount < m.MaxAlertsPerIncident
	if !m.AlertsEnabled || !withinLimit || r.consecutiveFailures <= m.AlertAfterNFailures {
		return
	}
	r.alert = domainThresholdAlert(m, daysLeft, expiresAt, &r.alerted30d, &r.alerted14d, &r.alerted7d)
	if r.alert != nil {
		r.alertCount = m.AlertCount + 1
	}
}

// domainThresholdAlert returns the warning for the first newly-crossed
// threshold of a domain not yet expired (7/14/30 days), setting the matching
// alerted flag so checkOneDomainMonitor only sends one notification per crossing.
func domainThresholdAlert(m db.DomainMonitor, daysLeft int, expiresAt time.Time, alerted30d, alerted14d, alerted7d *bool) *AlertMessage {
	if !sslCrossedThreshold(daysLeft, alerted30d, alerted14d, alerted7d) {
		return nil
	}
	subject, telegramMsg, emailHTML := domainExpiringSoonMessages(m, daysLeft, expiresAt.Format("2006-01-02"))
	a := domainAlert(m, subject, telegramMsg, emailHTML)
	return &a
}

// domainAlert renders a domain alert's wording into every channel's format.
func domainAlert(m db.DomainMonitor, subject, telegramMsg, emailHTML string) AlertMessage {
	return AlertMessage{
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

// domainExpiredMessages/domainExpiringSoonMessages delegate to the shared
// expiredMessages/expiringSoonMessages (worker.go) — mirrored by
// sslExpiredMessages/sslExpiringSoonMessages in worker_ssl.go, which only
// differ in Thing/FieldLabel/Target/Emoji.
func domainExpiredMessages(m db.DomainMonitor, expiresStr string) (subject, telegramMsg, emailHTML string) {
	return expiredMessages(expiryAlertSubject{Name: m.Name, Thing: "domain registration", FieldLabel: "Domain", Target: m.Domain}, expiresStr)
}

func domainExpiringSoonMessages(m db.DomainMonitor, daysLeft int, expiresStr string) (subject, telegramMsg, emailHTML string) {
	return expiringSoonMessages(expiryAlertSubject{Name: m.Name, Thing: "domain registration", FieldLabel: "Domain", Target: m.Domain, Emoji: "🌐"}, daysLeft, expiresStr)
}
