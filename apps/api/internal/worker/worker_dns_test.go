package worker

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

// ─── fake DNS server ─────────────────────────────────────────────────────────
//
// fakeDNSResolver wires a *net.Resolver's Dial hook to an in-memory net.Pipe
// instead of a real socket — no port binding, no real network I/O, so the
// pure-Go resolver still speaks genuine DNS wire format but never leaves the
// process. ips == nil answers every query with NXDOMAIN; otherwise every
// query gets back an A-record answer set built from ips, regardless of the
// query's own qtype — sufficient for this package's tests, which only ever
// point an "A" record type monitor at it.

// fakeDNSResolver returns a resolver that answers any query with ips (or
// NXDOMAIN if ips is nil).
func fakeDNSResolver(t *testing.T, ips []net.IP) *net.Resolver {
	t.Helper()
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			client, server := net.Pipe()
			go func() {
				defer func() { _ = server.Close() }()
				buf := make([]byte, 512)
				n, err := server.Read(buf)
				if err != nil {
					return
				}
				resp := buildFakeDNSResponse(buf[:n], ips)
				if resp == nil {
					return
				}
				_, _ = server.Write(resp)
			}()
			return client, nil
		},
	}
}

// buildFakeDNSResponse crafts a minimal single-question DNS response,
// echoing query's ID and question section verbatim and appending one A
// record per ip (or none, with RCODE=3, when ips is nil — NXDOMAIN).
//
// query and the returned response are both prefixed with a 2-byte
// big-endian length, even though these are logically UDP exchanges — Go's
// resolver falls back to that stream framing whenever the Dial-returned
// net.Conn isn't also a net.PacketConn, which a net.Pipe() conn never is.
func buildFakeDNSResponse(query []byte, ips []net.IP) []byte {
	if len(query) < 14 {
		return nil
	}
	msg := query[2:] // drop the 2-byte length prefix

	i := 12
	for i < len(msg) && msg[i] != 0 {
		i += int(msg[i]) + 1
	}
	i++    // terminating zero label
	i += 4 // QTYPE + QCLASS
	if i > len(msg) {
		return nil
	}
	question := msg[12:i]

	flags := []byte{0x81, 0x80} // QR=1 RD=1 RA=1 RCODE=0
	ancount := len(ips)
	if ips == nil {
		flags = []byte{0x81, 0x83} // RCODE=3 NXDOMAIN
	}

	body := make([]byte, 0, 512)
	body = append(body, msg[0], msg[1]) // ID
	body = append(body, flags...)
	body = append(body, 0, 1) // QDCOUNT=1
	body = append(body, byte(ancount>>8), byte(ancount))
	body = append(body, 0, 0, 0, 0) // NSCOUNT, ARCOUNT
	body = append(body, question...)
	for _, ip := range ips {
		ip4 := ip.To4()
		body = append(body, 0xC0, 0x0C)       // NAME: pointer to offset 12
		body = append(body, 0, 1)             // TYPE A
		body = append(body, 0, 1)             // CLASS IN
		body = append(body, 0, 0, 0x01, 0x2C) // TTL 300
		body = append(body, 0, 4)             // RDLENGTH
		body = append(body, ip4...)
	}

	resp := make([]byte, 0, len(body)+2)
	resp = append(resp, byte(len(body)>>8), byte(len(body)))
	resp = append(resp, body...)
	return resp
}

// ─── test fixtures ────────────────────────────────────────────────────────

func testDNSMonitor(t *testing.T, queries *db.Queries, orgID uuid.UUID, hostname, expectedValue string) db.DnsMonitor {
	t.Helper()
	m, err := queries.CreateDNSMonitor(context.Background(), db.CreateDNSMonitorParams{
		OrgID: orgID, Name: "DNS monitor", Hostname: hostname, RecordType: db.DnsRecordTypeA,
		ExpectedValue: pgtype.Text{String: expectedValue, Valid: expectedValue != ""},
		IntervalMins:  10, MaxAlertsPerIncident: 3,
	})
	if err != nil {
		t.Fatalf("create test dns monitor: %v", err)
	}
	return m
}

func getDNSRow(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) (status string, consecutiveFailures int32, expectedValue pgtype.Text, baselineCaptured bool) {
	t.Helper()
	if err := pool.QueryRow(context.Background(),
		"SELECT status, consecutive_failures, expected_value, baseline_captured FROM dns_monitors WHERE id = $1", id,
	).Scan(&status, &consecutiveFailures, &expectedValue, &baselineCaptured); err != nil {
		t.Fatalf("query dns row: %v", err)
	}
	return status, consecutiveFailures, expectedValue, baselineCaptured
}

// ─── integration tests ───────────────────────────────────────────────────

func TestCheckDNSMonitors(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	tg := telegram.NewClient("")
	mailer := email.NewSender("")
	wh := webhook.NewClientWithHTTPClient(&http.Client{Timeout: 10 * time.Second})
	logger := testLogger()

	t.Run("baseline mode captures the first successful lookup as the comparison value", func(t *testing.T) {
		n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, Logger: logger,
			DNSResolver: fakeDNSResolver(t, []net.IP{net.ParseIP("1.2.3.4")})}

		org := testOrg(t, queries, pool)
		mon := testDNSMonitor(t, queries, org.ID, "example.com", "") // no expected value: baseline mode

		checkDNSMonitors(context.Background(), n)

		status, failures, expected, captured := getDNSRow(t, pool, mon.ID)
		if status != "up" || failures != 0 {
			t.Fatalf("want up with 0 failures after first baseline check, got status=%q failures=%d", status, failures)
		}
		if !expected.Valid || expected.String != "1.2.3.4" || !captured {
			t.Fatalf("want expected_value=1.2.3.4 baseline_captured=true, got %+v captured=%v", expected, captured)
		}
	})

	t.Run("a later check that differs from the captured baseline goes down as a mismatch, not a lookup error", func(t *testing.T) {
		n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, Logger: logger,
			DNSResolver: fakeDNSResolver(t, []net.IP{net.ParseIP("1.2.3.4")})}

		org := testOrg(t, queries, pool)
		mon := testDNSMonitor(t, queries, org.ID, "example.com", "")
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeEmail, map[string]string{"email": "a@b.com"})
		attachNotificationChannel(t, queries, channel.ID, "dns", mon.ID)

		checkDNSMonitors(context.Background(), n) // captures baseline 1.2.3.4

		n.DNSResolver = fakeDNSResolver(t, []net.IP{net.ParseIP("9.9.9.9")})
		forceDueNow(t, pool, "dns_monitors", mon.ID)
		checkDNSMonitors(context.Background(), n)

		status, _, expected, _ := getDNSRow(t, pool, mon.ID)
		if status != "down" {
			t.Fatalf("want down after the resolved value changed, got %q", status)
		}
		if expected.String != "1.2.3.4" {
			t.Fatalf("want expected_value to stay pinned at the baseline until acknowledged, got %q", expected.String)
		}
		var reason pgtype.Text
		var resolvedValue string
		if err := pool.QueryRow(context.Background(),
			"SELECT failure_reason, resolved_value FROM dns_checks WHERE monitor_id = $1 ORDER BY checked_at DESC LIMIT 1", mon.ID,
		).Scan(&reason, &resolvedValue); err != nil {
			t.Fatalf("query check: %v", err)
		}
		if reason.Valid {
			t.Fatalf("want no failure_reason on a mismatch (distinct from a lookup error), got %q", reason.String)
		}
		if resolvedValue != "9.9.9.9" {
			t.Fatalf("want the new resolved value recorded, got %q", resolvedValue)
		}
	})

	t.Run("NXDOMAIN is recorded as a lookup error, distinct from a mismatch", func(t *testing.T) {
		n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, Logger: logger,
			DNSResolver: fakeDNSResolver(t, nil)} // nil ips == NXDOMAIN

		org := testOrg(t, queries, pool)
		mon := testDNSMonitor(t, queries, org.ID, "does-not-exist.invalid", "1.2.3.4")

		checkDNSMonitors(context.Background(), n)

		status, _, _, _ := getDNSRow(t, pool, mon.ID)
		if status != "down" {
			t.Fatalf("want down on a lookup failure, got %q", status)
		}
		var reason pgtype.Text
		if err := pool.QueryRow(context.Background(),
			"SELECT failure_reason FROM dns_checks WHERE monitor_id = $1", mon.ID,
		).Scan(&reason); err != nil {
			t.Fatalf("query check: %v", err)
		}
		if !reason.Valid || reason.String != "NXDOMAIN" {
			t.Fatalf("want failure_reason=NXDOMAIN, got %+v", reason)
		}
	})

	t.Run("pinned expected value: matching lookup is up, recovers an existing incident", func(t *testing.T) {
		n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, Logger: logger,
			DNSResolver: fakeDNSResolver(t, nil)}

		org := testOrg(t, queries, pool)
		mon := testDNSMonitor(t, queries, org.ID, "example.com", "1.2.3.4")

		checkDNSMonitors(context.Background(), n) // NXDOMAIN -> down, opens an incident

		status, _, _, _ := getDNSRow(t, pool, mon.ID)
		if status != "down" {
			t.Fatalf("want down after NXDOMAIN, got %q", status)
		}

		n.DNSResolver = fakeDNSResolver(t, []net.IP{net.ParseIP("1.2.3.4")})
		forceDueNow(t, pool, "dns_monitors", mon.ID)
		checkDNSMonitors(context.Background(), n)

		status, failures, _, _ := getDNSRow(t, pool, mon.ID)
		if status != "up" || failures != 0 {
			t.Fatalf("want up with 0 failures after recovery, got status=%q failures=%d", status, failures)
		}
		var resolvedAt pgtype.Timestamptz
		if err := pool.QueryRow(context.Background(),
			"SELECT resolved_at FROM dns_incidents WHERE monitor_id = $1", mon.ID,
		).Scan(&resolvedAt); err != nil {
			t.Fatalf("query incident: %v", err)
		}
		if !resolvedAt.Valid {
			t.Fatal("want the incident resolved after recovery")
		}
	})

	t.Run("records a check row for every poll", func(t *testing.T) {
		n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, Logger: logger,
			DNSResolver: fakeDNSResolver(t, []net.IP{net.ParseIP("1.2.3.4")})}

		org := testOrg(t, queries, pool)
		mon := testDNSMonitor(t, queries, org.ID, "example.com", "")
		checkDNSMonitors(context.Background(), n)

		var count int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM dns_checks WHERE monitor_id = $1", mon.ID).Scan(&count); err != nil {
			t.Fatalf("count checks: %v", err)
		}
		if count != 1 {
			t.Fatalf("want 1 check recorded, got %d", count)
		}
	})

	t.Run("a monitor under an active maintenance window is excluded", func(t *testing.T) {
		n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, Logger: logger,
			DNSResolver: fakeDNSResolver(t, []net.IP{net.ParseIP("1.2.3.4")})}

		org := testOrg(t, queries, pool)
		mon := testDNSMonitor(t, queries, org.ID, "example.com", "")
		var windowID uuid.UUID
		if err := pool.QueryRow(context.Background(),
			"INSERT INTO maintenance_windows (org_id, title, message, starts_at) VALUES ($1, 'Scheduled', '', NOW() - INTERVAL '1 minute') RETURNING id",
			org.ID,
		).Scan(&windowID); err != nil {
			t.Fatalf("seed maintenance window: %v", err)
		}
		mustExecWorker(t, pool, "INSERT INTO maintenance_window_monitors (window_id, monitor_type, monitor_id) VALUES ($1, 'dns', $2)", windowID, mon.ID)

		checkDNSMonitors(context.Background(), n)

		var count int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM dns_checks WHERE monitor_id = $1", mon.ID).Scan(&count); err != nil {
			t.Fatalf("count checks: %v", err)
		}
		if count != 0 {
			t.Fatalf("want no check performed under maintenance, got %d", count)
		}
	})
}

// ─── pure function tests ──────────────────────────────────────────────────

func TestJoinSortedValues(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want string
	}{
		{"empty", nil, ""},
		{"single", []string{"1.2.3.4"}, "1.2.3.4"},
		{"sorts multi-value answers", []string{"9.9.9.9", "1.2.3.4"}, "1.2.3.4; 9.9.9.9"},
		{"txt content containing a comma stays intact", []string{"v=spf1, include:x"}, "v=spf1, include:x"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := joinSortedValues(tc.in); got != tc.want {
				t.Fatalf("joinSortedValues(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestTrimTrailingDot(t *testing.T) {
	if got := trimTrailingDot("mail.example.com."); got != "mail.example.com" {
		t.Fatalf("want trailing dot trimmed, got %q", got)
	}
	if got := trimTrailingDot("mail.example.com"); got != "mail.example.com" {
		t.Fatalf("want no-op without a trailing dot, got %q", got)
	}
}

func TestClassifyLookupError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"not found -> NXDOMAIN", &net.DNSError{IsNotFound: true}, "NXDOMAIN"},
		{"timeout", &net.DNSError{IsTimeout: true}, "DNS lookup timeout"},
		{"other DNS error", &net.DNSError{Err: "server misbehaving"}, "DNS lookup failed: server misbehaving"},
		{"non-DNS error", errors.New("boom"), "DNS lookup failed: boom"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyLookupError(tc.err); got != tc.want {
				t.Fatalf("classifyLookupError(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}

func TestBuildDNSDownAlert(t *testing.T) {
	t.Run("lookup error reads as unreachable, not a change", func(t *testing.T) {
		m := db.DnsMonitor{Name: "Example", Hostname: "example.com", RecordType: db.DnsRecordTypeA}
		msg := buildDNSDownAlert(m, "", "NXDOMAIN")
		if !strings.Contains(msg.EmailSubject, "DNS lookup failed") {
			t.Fatalf("want a lookup-failed subject, got %q", msg.EmailSubject)
		}
		if msg.Webhook.Reason != "NXDOMAIN" {
			t.Fatalf("want the lookup error as the webhook reason, got %q", msg.Webhook.Reason)
		}
	})

	t.Run("mismatch reads as a change, old value to new value", func(t *testing.T) {
		m := db.DnsMonitor{Name: "Example", Hostname: "example.com", RecordType: db.DnsRecordTypeA,
			ExpectedValue: pgtype.Text{String: "1.2.3.4", Valid: true}}
		msg := buildDNSDownAlert(m, "9.9.9.9", "")
		if !strings.Contains(msg.EmailSubject, "DNS record changed") {
			t.Fatalf("want a record-changed subject, got %q", msg.EmailSubject)
		}
		if !strings.Contains(msg.Webhook.Reason, "1.2.3.4") || !strings.Contains(msg.Webhook.Reason, "9.9.9.9") {
			t.Fatalf("want both old and new values in the reason, got %q", msg.Webhook.Reason)
		}
	})
}
