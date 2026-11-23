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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/deliver"
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
			defer recoverPanic(n.Logger, "dns check", m.ID)
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
// record apart from one that's simply unreachable (PRD dns-monitoring).
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

	// A mismatch is rechecked like any failure (ADR alerting); a changed
	// record that is real still differs on the recheck.
	trackCheck(ctx, n, dnsTracker{
		q: n.Queries, m: m, responseTimeMs: responseTimeMs, resolvedValue: resolvedValue, failureReason: failureReason,
	}, isUp)
}

// dnsTracker is a DNS monitor's side of trackCheck, bound to one check's
// result. failureReason is empty for a record that resolved but differs.
type dnsTracker struct {
	q              *db.Queries
	m              db.DnsMonitor
	responseTimeMs int64
	resolvedValue  string
	failureReason  string
}

func (t dnsTracker) monitor() trackedMonitor {
	return trackedMonitor{
		Type: "dns", ID: t.m.ID, OrgID: t.m.OrgID, Status: t.m.Status,
		ConsecutiveFailures: t.m.ConsecutiveFailures, AlertAfterNFailures: t.m.AlertAfterNFailures,
		AlertsEnabled: t.m.AlertsEnabled,
	}
}

func (t dnsTracker) latestCheckFailed(ctx context.Context) (bool, error) {
	return t.q.LatestDNSCheckFailed(ctx, t.m.ID)
}

func (t dnsTracker) recordCheck(ctx context.Context, isUp bool) error {
	_, err := t.q.CreateDNSCheck(ctx, db.CreateDNSCheckParams{
		MonitorID:      t.m.ID,
		ResponseTimeMs: int32(t.responseTimeMs),
		IsUp:           isUp,
		ResolvedValue:  pgtype.Text{String: t.resolvedValue, Valid: t.resolvedValue != ""},
		FailureReason:  pgtype.Text{String: t.failureReason, Valid: t.failureReason != ""},
	})
	return err
}

func (t dnsTracker) scheduleRecheck(ctx context.Context) error {
	return t.q.ScheduleDNSRecheck(ctx, t.m.ID)
}

func (t dnsTracker) recordUp(ctx context.Context) (pgtype.Timestamptz, error) {
	return t.q.RecordDNSCheckUp(ctx, db.RecordDNSCheckUpParams{
		ID:                t.m.ID,
		LastResolvedValue: pgtype.Text{String: t.resolvedValue, Valid: true},
	})
}

func (t dnsTracker) recordFailure(ctx context.Context) (int32, error) {
	// A lookup failure has no resolvedValue — keep the monitor's last known
	// good value on display rather than blanking it out (PRD dns-monitoring); a
	// mismatch always carries a real resolvedValue to show instead.
	lastResolved := t.m.LastResolvedValue
	if t.resolvedValue != "" {
		lastResolved = pgtype.Text{String: t.resolvedValue, Valid: true}
	}
	updated, err := t.q.RecordDNSCheckFailure(ctx, db.RecordDNSCheckFailureParams{
		ID:                t.m.ID,
		LastResolvedValue: lastResolved,
	})
	return updated.ConsecutiveFailures, err
}

func (t dnsTracker) openIncident(ctx context.Context, alert []byte) (uuid.UUID, error) {
	inc, err := t.q.OpenDNSIncident(ctx, db.OpenDNSIncidentParams{MonitorID: t.m.ID, Alert: alert})
	return inc.ID, err
}

func (t dnsTracker) countAlert(ctx context.Context, incidentID uuid.UUID) error {
	_, err := t.q.IncrementDNSIncidentAlertCount(ctx, incidentID)
	return err
}

func (t dnsTracker) downAlert() AlertMessage {
	return buildDNSDownAlert(t.m, t.resolvedValue, t.failureReason)
}

func (t dnsTracker) recoveryAlert(downtime string) AlertMessage {
	return buildDNSRecoveryAlert(t.m, t.resolvedValue, downtime)
}

func buildDNSRecoveryAlert(m db.DnsMonitor, resolvedValue, downtime string) AlertMessage {
	return buildAlert(alertText{
		Event: "recovery", MonitorType: "dns", Name: m.Name,
		Icon: "✅", Headline: " DNS record is back to normal",
		Details: []string{
			codeLine("Hostname", fmt.Sprintf("%s (%s)", m.Hostname, m.RecordType)),
			codeLine("Value", html.EscapeString(resolvedValue)),
		},
		Subject:  m.Name + " recovered",
		Downtime: downtime,
		SMS:      fmt.Sprintf("%s recovered after %s downtime", m.Name, downtime),
	})
}

// buildDNSDownAlert branches on whether failureReason is set: a lookup
// error (NXDOMAIN/SERVFAIL/timeout) reads as "can't resolve", while an empty
// failureReason means the lookup succeeded but didn't match — a changed/
// hijacked record, phrased as old value → new value (PRD dns-monitoring).
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
	hostLine := codeLine("Hostname", fmt.Sprintf("%s (%s)", m.Hostname, m.RecordType))
	if failureReason != "" {
		return buildAlert(alertText{
			Event: "down", MonitorType: "dns", Name: m.Name,
			Icon: "🔴", Headline: ": DNS lookup failed",
			Details: []string{hostLine, "Reason: " + html.EscapeString(failureReason)},
			Subject: m.Name + ": DNS lookup failed",
			Reason:  failureReason,
			SMS:     fmt.Sprintf("%s DNS lookup failed (%s)", m.Name, failureReason),
		})
	}

	oldValue := "(unknown)"
	if m.ExpectedValue.Valid {
		oldValue = m.ExpectedValue.String
	}
	return buildAlert(alertText{
		Event: "down", MonitorType: "dns", Name: m.Name,
		Icon: "⚠️", Headline: ": DNS record changed",
		Details: []string{hostLine, fmt.Sprintf("%s → %s", html.EscapeString(oldValue), html.EscapeString(resolvedValue))},
		Subject: m.Name + ": DNS record changed",
		Reason:  fmt.Sprintf("%s → %s", oldValue, resolvedValue),
		SMS:     fmt.Sprintf("%s DNS record changed to %s", m.Name, resolvedValue),
	})
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

// DNSRecordSet is one record type's answer in the free DNS lookup
// (PRD free-tools): its values, none when the name has no such record, or
// why the lookup failed.
type DNSRecordSet struct {
	Type   string   `json:"type"`
	Values []string `json:"values"`
	Error  string   `json:"error,omitempty"`
}

// dnsLookupTypes are the record types a DNS monitor checks, in the order
// the free DNS lookup lists them.
var dnsLookupTypes = []db.DnsRecordType{
	db.DnsRecordTypeA, db.DnsRecordTypeAAAA, db.DnsRecordTypeCNAME,
	db.DnsRecordTypeMX, db.DnsRecordTypeTXT, db.DnsRecordTypeNS,
}

// LookupAllDNS resolves hostname for every record type a DNS monitor
// checks, all at once, for the free DNS lookup.
func LookupAllDNS(ctx context.Context, resolver *net.Resolver, hostname string) []DNSRecordSet {
	ctx, cancel := context.WithTimeout(ctx, deliver.Timeout)
	defer cancel()

	sets := make([]DNSRecordSet, len(dnsLookupTypes))
	var wg sync.WaitGroup
	for i, recordType := range dnsLookupTypes {
		wg.Go(func() {
			sets[i] = lookupRecordSet(ctx, resolver, recordType, hostname)
		})
	}
	wg.Wait()
	return sets
}

func lookupRecordSet(ctx context.Context, resolver *net.Resolver, recordType db.DnsRecordType, hostname string) DNSRecordSet {
	set := DNSRecordSet{Type: string(recordType), Values: []string{}}
	var values []string
	var err error
	if recordType == db.DnsRecordTypeMX {
		// Unlike a monitor's compared value, the lookup shows each mail
		// server's preference, in the resolver's preference order.
		var mxs []*net.MX
		mxs, err = resolver.LookupMX(ctx, hostname)
		for _, mx := range mxs {
			values = append(values, fmt.Sprintf("%d %s", mx.Pref, trimTrailingDot(mx.Host)))
		}
	} else {
		values, err = lookupRecord(ctx, resolver, recordType, hostname)
		sort.Strings(values)
	}
	var dnsErr *net.DNSError
	switch {
	case errors.As(err, &dnsErr) && dnsErr.IsNotFound:
		return set
	case err != nil:
		set.Error = classifyLookupError(err)
		return set
	}
	// LookupCNAME answers the name itself when it has no CNAME record.
	if recordType == db.DnsRecordTypeCNAME && len(values) == 1 && values[0] == hostname {
		return set
	}
	if values != nil {
		set.Values = values
	}
	return set
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
// doesn't exist" from "the lookup timed out" from anything else (PRD dns-monitoring).
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
