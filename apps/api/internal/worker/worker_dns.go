package worker

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/deliver"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

// ─── dns monitor checks ──────────────────────────────────────────────────────

func checkDNSMonitors(ctx context.Context, n Notifiers) {
	monitors, err := n.Queries.ListDueDNSMonitors(ctx)
	if err != nil {
		n.Logger.Error("dns worker: list due monitors", "err", err)
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
			checkOneDNSMonitor(ctx, n, m)
		}()
	}
	wg.Wait()
}

// checkOneDNSMonitor resolves the monitor's record and classifies the
// outcome into exactly one of three states: up (matches expected/baseline,
// or nothing to compare against yet), mismatch-down (resolved but differs —
// failureReason stays empty), or lookup-error-down (failureReason set) —
// this three-way split is what lets the alert text tell a hijacked/changed
// record apart from one that's simply unreachable (US-3902).
func checkOneDNSMonitor(ctx context.Context, n Notifiers, m db.DnsMonitor) {
	resolver := n.DNSResolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	responseTimeMs, resolvedValue, lookupErr := performDNSLookup(ctx, m, resolver)

	var failureReason string
	isUp := lookupErr == nil
	switch {
	case lookupErr != nil:
		failureReason = classifyLookupError(lookupErr)
	case m.ExpectedValue.Valid && m.ExpectedValue.String != resolvedValue:
		isUp = false
	}

	prevStatus := m.Status
	if !recordDNSCheck(ctx, n, m, responseTimeMs, isUp, resolvedValue, failureReason) {
		return
	}
	if isUp {
		handleDNSUp(ctx, n, m, prevStatus, resolvedValue)
		return
	}
	handleDNSDown(ctx, n, m, prevStatus, resolvedValue, failureReason)
}

func recordDNSCheck(ctx context.Context, n Notifiers, m db.DnsMonitor, responseTimeMs int64, isUp bool, resolvedValue, failureReason string) bool {
	if _, err := n.Queries.CreateDNSCheck(ctx, db.CreateDNSCheckParams{
		MonitorID:      m.ID,
		ResponseTimeMs: int32(responseTimeMs),
		IsUp:           isUp,
		ResolvedValue:  pgtype.Text{String: resolvedValue, Valid: resolvedValue != ""},
		FailureReason:  pgtype.Text{String: failureReason, Valid: failureReason != ""},
	}); err != nil {
		n.Logger.Error("dns worker: create check", "monitor_id", m.ID, "err", err)
		return false
	}
	return true
}

func handleDNSUp(ctx context.Context, n Notifiers, m db.DnsMonitor, prevStatus db.MonitorStatus, resolvedValue string) {
	if _, err := n.Queries.RecordDNSCheckUp(ctx, db.RecordDNSCheckUpParams{
		ID:                m.ID,
		LastResolvedValue: pgtype.Text{String: resolvedValue, Valid: true},
	}); err != nil {
		n.Logger.Error("dns worker: record up", "monitor_id", m.ID, "err", err)
		return
	}
	if prevStatus != db.MonitorStatusDown {
		return
	}
	inc, err := n.Queries.ResolveLatestDNSIncident(ctx, m.ID)
	if err != nil {
		n.Logger.Error("dns worker: resolve incident", "monitor_id", m.ID, "err", err)
		return
	}
	if !m.AlertsEnabled {
		return
	}
	downtime := FormatDuration(time.Since(inc.StartedAt.Time))
	DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "dns", ID: m.ID}, buildDNSRecoveryAlert(m, resolvedValue, downtime))
}

func buildDNSRecoveryAlert(m db.DnsMonitor, resolvedValue, downtime string) AlertMessage {
	hostRecord := fmt.Sprintf("%s (%s)", m.Hostname, m.RecordType)
	safeValue := html.EscapeString(resolvedValue)
	return AlertMessage{
		Telegram: fmt.Sprintf("✅ <b>%s</b> DNS record is back to normal\n\nHostname: <code>%s</code>\nValue: <code>%s</code>",
			m.Name, hostRecord, safeValue),
		EmailSubject: fmt.Sprintf("%s recovered", m.Name),
		EmailHTML: fmt.Sprintf("<p>✅ <b>%s</b> DNS record is back to normal</p><p>Hostname: <code>%s</code><br>Value: <code>%s</code></p>",
			m.Name, hostRecord, safeValue),
		Webhook: &webhook.Event{
			EventType:        "recovery",
			MonitorName:      m.Name,
			MonitorType:      "dns",
			DowntimeDuration: downtime,
			Timestamp:        time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.RecoveryMessage(m.Name, "dns", downtime)),
		SMS:   TruncateSMS(fmt.Sprintf("Checkmeup: %s recovered after %s downtime", m.Name, downtime)),
	}
}

func handleDNSDown(ctx context.Context, n Notifiers, m db.DnsMonitor, prevStatus db.MonitorStatus, resolvedValue, failureReason string) {
	// A lookup failure has no resolvedValue — keep the monitor's last known
	// good value on display rather than blanking it out (US-3904); a
	// mismatch always carries a real resolvedValue to show instead.
	lastResolved := m.LastResolvedValue
	if resolvedValue != "" {
		lastResolved = pgtype.Text{String: resolvedValue, Valid: true}
	}
	updated, err := n.Queries.RecordDNSCheckFailure(ctx, db.RecordDNSCheckFailureParams{
		ID:                m.ID,
		LastResolvedValue: lastResolved,
	})
	if err != nil {
		n.Logger.Error("dns worker: record failure", "monitor_id", m.ID, "err", err)
		return
	}
	if updated.ConsecutiveFailures <= m.AlertAfterNFailures || prevStatus == db.MonitorStatusDown {
		return
	}
	if err := n.Queries.MarkDNSMonitorDown(ctx, m.ID); err != nil {
		n.Logger.Error("dns worker: mark down", "monitor_id", m.ID, "err", err)
		return
	}
	inc, err := n.Queries.CreateDNSIncident(ctx, m.ID)
	if err != nil {
		n.Logger.Error("dns worker: create incident", "monitor_id", m.ID, "err", err)
		return
	}
	alertDNSIncident(ctx, n, m, inc, resolvedValue, failureReason)
}

// alertDNSIncident dispatches the down alert for a freshly-recorded
// incident, honoring the monitor's alerts-enabled flag and per-incident cap.
func alertDNSIncident(ctx context.Context, n Notifiers, m db.DnsMonitor, inc db.DnsIncident, resolvedValue, failureReason string) {
	if !m.AlertsEnabled {
		return
	}
	if max := m.MaxAlertsPerIncident; max > 0 && inc.AlertCount >= max {
		return
	}
	if !DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "dns", ID: m.ID}, buildDNSDownAlert(m, resolvedValue, failureReason)) {
		return
	}
	if _, err := n.Queries.IncrementDNSIncidentAlertCount(ctx, inc.ID); err != nil {
		n.Logger.Error("dns worker: increment alert count", "incident_id", inc.ID, "err", err)
	}
}

// buildDNSDownAlert branches on whether failureReason is set: a lookup
// error (NXDOMAIN/SERVFAIL/timeout) reads as "can't resolve", while an empty
// failureReason means the lookup succeeded but didn't match — a changed/
// hijacked record, phrased as old value → new value (US-3902/US-3903).
//
// resolvedValue and a baseline-captured expectedValue both originate from
// the monitored domain's own DNS records — not from anything the alerted
// org typed — so they're untrusted from checkmeup's perspective even
// though nothing here is "user input" in the usual sense. Whoever controls
// that DNS zone could otherwise inject markup into another org's alert
// email/Telegram message; html.EscapeString neutralizes that before either
// value reaches Telegram's HTML parse mode or EmailHTML. Webhook/Slack/SMS
// use the raw values — Webhook is JSON (safe by construction), Slack/SMS
// aren't HTML, and escaping them would just show literal "&amp;" etc.
func buildDNSDownAlert(m db.DnsMonitor, resolvedValue, failureReason string) AlertMessage {
	hostRecord := fmt.Sprintf("%s (%s)", m.Hostname, m.RecordType)
	if failureReason != "" {
		subject := fmt.Sprintf("%s: DNS lookup failed", m.Name)
		safeReason := html.EscapeString(failureReason)
		return AlertMessage{
			Telegram: fmt.Sprintf("🔴 <b>%s</b>: DNS lookup failed\n\nHostname: <code>%s</code>\nReason: %s",
				m.Name, hostRecord, safeReason),
			EmailSubject: subject,
			EmailHTML: fmt.Sprintf("<p>🔴 <b>%s</b>: DNS lookup failed</p><p>Hostname: <code>%s</code><br>Reason: %s</p>",
				m.Name, hostRecord, safeReason),
			Webhook: &webhook.Event{
				EventType:   "down",
				MonitorName: m.Name,
				MonitorType: "dns",
				Reason:      failureReason,
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
			},
			Slack: slackMsg(slack.DownMessage(m.Name, "dns", failureReason)),
			SMS:   TruncateSMS(fmt.Sprintf("Checkmeup: %s DNS lookup failed (%s)", m.Name, failureReason)),
		}
	}

	oldValue := "(unknown)"
	if m.ExpectedValue.Valid {
		oldValue = m.ExpectedValue.String
	}
	changeDesc := fmt.Sprintf("%s → %s", oldValue, resolvedValue)
	safeChangeDesc := fmt.Sprintf("%s → %s", html.EscapeString(oldValue), html.EscapeString(resolvedValue))
	subject := fmt.Sprintf("%s: DNS record changed", m.Name)
	return AlertMessage{
		Telegram: fmt.Sprintf("⚠️ <b>%s</b>: DNS record changed\n\nHostname: <code>%s</code>\n%s",
			m.Name, hostRecord, safeChangeDesc),
		EmailSubject: subject,
		EmailHTML: fmt.Sprintf("<p>⚠️ <b>%s</b>: DNS record changed</p><p>Hostname: <code>%s</code><br>%s</p>",
			m.Name, hostRecord, safeChangeDesc),
		Webhook: &webhook.Event{
			EventType:   "down",
			MonitorName: m.Name,
			MonitorType: "dns",
			Reason:      changeDesc,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.DownMessage(m.Name, "dns", changeDesc)),
		SMS:   TruncateSMS(fmt.Sprintf("Checkmeup: %s DNS record changed to %s", m.Name, resolvedValue)),
	}
}

// performDNSLookup resolves m's hostname per its record type, sorting and
// joining multi-value answers ("; ", not "," — TXT content can itself
// contain commas) so answer-order jitter never looks like a change.
// resolver is injected so tests can point it at a fake DNS server — see
// Notifiers.DNSResolver.
func performDNSLookup(ctx context.Context, m db.DnsMonitor, resolver *net.Resolver) (responseTimeMs int64, resolvedValue string, lookupErr error) {
	ctx, cancel := context.WithTimeout(ctx, deliver.Timeout)
	defer cancel()

	start := time.Now()
	values, err := lookupRecord(ctx, resolver, m.RecordType, m.Hostname)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return elapsed, "", err
	}
	return elapsed, joinSortedValues(values), nil
}

// joinSortedValues sorts a multi-value DNS answer before joining so answer
// order jitter (e.g. round-robin A records) never looks like a change.
func joinSortedValues(values []string) string {
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	return strings.Join(sorted, "; ")
}

// lookupRecord dispatches to the net.Resolver method for m's record type.
// MX preference is dropped from the compared value (host only) — a
// simplification, not a full RFC-faithful comparison.
func lookupRecord(ctx context.Context, resolver *net.Resolver, recordType db.DnsRecordType, hostname string) ([]string, error) {
	switch recordType {
	case db.DnsRecordTypeA:
		ips, err := resolver.LookupIP(ctx, "ip4", hostname)
		if err != nil {
			return nil, err
		}
		return ipsToStrings(ips), nil
	case db.DnsRecordTypeAAAA:
		ips, err := resolver.LookupIP(ctx, "ip6", hostname)
		if err != nil {
			return nil, err
		}
		return ipsToStrings(ips), nil
	case db.DnsRecordTypeCNAME:
		cname, err := resolver.LookupCNAME(ctx, hostname)
		if err != nil {
			return nil, err
		}
		return []string{trimTrailingDot(cname)}, nil
	case db.DnsRecordTypeMX:
		mxs, err := resolver.LookupMX(ctx, hostname)
		if err != nil {
			return nil, err
		}
		out := make([]string, len(mxs))
		for i, mx := range mxs {
			out[i] = trimTrailingDot(mx.Host)
		}
		return out, nil
	case db.DnsRecordTypeTXT:
		return resolver.LookupTXT(ctx, hostname)
	case db.DnsRecordTypeNS:
		nss, err := resolver.LookupNS(ctx, hostname)
		if err != nil {
			return nil, err
		}
		out := make([]string, len(nss))
		for i, ns := range nss {
			out[i] = trimTrailingDot(ns.Host)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported dns record type: %s", recordType)
	}
}

func ipsToStrings(ips []net.IP) []string {
	out := make([]string, len(ips))
	for i, ip := range ips {
		out[i] = ip.String()
	}
	return out
}

func trimTrailingDot(s string) string {
	return strings.TrimSuffix(s, ".")
}

// classifyLookupError turns a lookup error into the short, user-facing
// reason shown in the check log and alert text — distinguishing "the name
// doesn't exist" from "the lookup timed out" from anything else (US-3902).
func classifyLookupError(err error) string {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		switch {
		case dnsErr.IsNotFound:
			return "NXDOMAIN"
		case dnsErr.IsTimeout:
			return "DNS lookup timeout"
		default:
			return "DNS lookup failed: " + dnsErr.Err
		}
	}
	return "DNS lookup failed: " + err.Error()
}
