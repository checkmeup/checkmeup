package worker

// Integration tests for the background worker (goroutine-per-tick scheduler,
// ADR-001). Same conventions as the handler package's tests: real Postgres
// (ADR-010), DATABASE_URL from the environment with a devcontainer-matching
// fallback. This is a different package from internal/handler, so fixtures
// are created directly via db.Queries rather than through the HTTP handlers
// — that's also more honest about what the worker actually depends on.
//
// telegram.Client/email.Sender are constructed with no token/API key, same
// as the handler tests: telegram.SendMessage errors immediately (no network
// call) when there's no bot token, while email.Sender silently no-ops with a
// nil error when there's no Resend key (ADR-012). dispatchAlert's tests
// exploit that asymmetry deliberately — it's the only way to exercise a
// real "alert delivered" code path without live credentials.
//
// performTLSCheck's success path (a real, currently-valid certificate) isn't
// covered — it hardcodes the system trust store with no way to inject a
// custom root, so only the connection-error path is testable without a live
// network dependency on a real public host. Same boundary as billing.go's
// CreateCheckout and settings.go's TestTelegram.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/rdap"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/twilio"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://checkmeup:checkmeup@db:5432/checkmeup?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping test db: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// testOrg creates a bare org with no notification channels — tests that need
// a deliverable channel create one explicitly via testNotificationChannel,
// since dispatchAlert reads monitor_notification_channels, not org fields
// (ADR-023).
func testOrg(t *testing.T, queries *db.Queries, pool *pgxpool.Pool) db.Org {
	t.Helper()
	org, err := queries.CreateOrg(context.Background(), db.CreateOrgParams{
		Name: "test-" + uuid.NewString(),
	})
	if err != nil {
		t.Fatalf("create test org: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM orgs WHERE id = $1", org.ID)
	})
	return org
}

// testNotificationChannel creates an enabled channel for orgID. Caller still
// needs attachNotificationChannel to wire it to a specific monitor.
// upgradeOrgPlan bumps a test org to plan — sms credit consumption (ADR-032)
// is 0 on the default Hobby plan, so tests exercising the actual Twilio send
// path (rather than credit exhaustion itself) need a paid plan first.
func upgradeOrgPlan(t *testing.T, queries *db.Queries, orgID uuid.UUID, plan db.Plan) {
	t.Helper()
	if err := queries.UpdateOrgPlan(context.Background(), db.UpdateOrgPlanParams{
		ID: orgID, Plan: plan, BillingCycle: "monthly", SubscriptionStatus: "active",
	}); err != nil {
		t.Fatalf("upgrade org plan: %v", err)
	}
}

func testNotificationChannel(t *testing.T, queries *db.Queries, orgID uuid.UUID, channelType db.NotificationChannelType, config map[string]string) db.NotificationChannel {
	t.Helper()
	configBytes, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal channel config: %v", err)
	}
	c, err := queries.CreateNotificationChannel(context.Background(), db.CreateNotificationChannelParams{
		OrgID: orgID, Type: channelType, Name: string(channelType), Config: configBytes,
	})
	if err != nil {
		t.Fatalf("create test notification channel: %v", err)
	}
	return c
}

func attachNotificationChannel(t *testing.T, queries *db.Queries, channelID uuid.UUID, monitorType string, monitorID uuid.UUID) {
	t.Helper()
	if err := queries.InsertMonitorNotificationChannel(context.Background(), db.InsertMonitorNotificationChannelParams{
		ChannelID: channelID, MonitorType: monitorType, MonitorID: monitorID,
	}); err != nil {
		t.Fatalf("attach test notification channel: %v", err)
	}
}

// testUser creates a user for orgID, so dispatchAlert's no-channel fallback
// (ADR-023: email every user in the org) has someone to email.
func testUser(t *testing.T, queries *db.Queries, orgID uuid.UUID) db.User {
	t.Helper()
	u, err := queries.CreateUser(context.Background(), db.CreateUserParams{
		OrgID: orgID, Email: "user-" + uuid.NewString() + "@example.com", PasswordHash: "x",
	})
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return u
}

func testCronMonitor(t *testing.T, queries *db.Queries, orgID uuid.UUID, schedule string) db.CronMonitor {
	t.Helper()
	m, err := queries.CreateCronMonitor(context.Background(), db.CreateCronMonitorParams{
		OrgID: orgID, Name: "Cron monitor", Schedule: schedule, GracePeriodMins: 5,
		PingToken: uuid.NewString(), MaxAlertsPerIncident: 3,
	})
	if err != nil {
		t.Fatalf("create test cron monitor: %v", err)
	}
	return m
}

func testUptimeMonitor(t *testing.T, queries *db.Queries, orgID uuid.UUID, url string) db.UptimeMonitor {
	t.Helper()
	m, err := queries.CreateUptimeMonitor(context.Background(), db.CreateUptimeMonitorParams{
		OrgID: orgID, Name: "Uptime monitor", Url: url, IntervalMins: 10, MaxAlertsPerIncident: 3,
		KeywordMode: db.KeywordModeContains, JsonAssertions: []byte("[]"),
	})
	if err != nil {
		t.Fatalf("create test uptime monitor: %v", err)
	}
	return m
}

func testSSLMonitor(t *testing.T, queries *db.Queries, orgID uuid.UUID, hostname string) db.SslMonitor {
	t.Helper()
	m, err := queries.CreateSSLMonitor(context.Background(), db.CreateSSLMonitorParams{
		OrgID: orgID, Name: "SSL monitor", Hostname: hostname,
	})
	if err != nil {
		t.Fatalf("create test ssl monitor: %v", err)
	}
	return m
}

func testDomainMonitor(t *testing.T, queries *db.Queries, orgID uuid.UUID, domain string) db.DomainMonitor {
	t.Helper()
	m, err := queries.CreateDomainMonitor(context.Background(), db.CreateDomainMonitorParams{
		OrgID: orgID, Name: "Domain monitor", Domain: domain,
	})
	if err != nil {
		t.Fatalf("create test domain monitor: %v", err)
	}
	return m
}

// testPortMonitor creates a port monitor pointed at host:port with
// AlertAfterNFailures left at 0, matching testUptimeMonitor's convention
// (see TestCheckUptimeMonitors) — the first failing check is already enough
// to trip it down.
func testPortMonitor(t *testing.T, queries *db.Queries, orgID uuid.UUID, host string, port int32, expectedState db.PortExpectedState) db.PortMonitor {
	t.Helper()
	m, err := queries.CreatePortMonitor(context.Background(), db.CreatePortMonitorParams{
		OrgID: orgID, Name: "Port monitor", Host: host, Port: port,
		ExpectedState: expectedState, IntervalMins: 10, MaxAlertsPerIncident: 3,
	})
	if err != nil {
		t.Fatalf("create test port monitor: %v", err)
	}
	return m
}

// openTCPPort starts a local listener that accepts (and immediately drops)
// connections, simulating a reachable port for performTCPCheck. Returns the
// host/port to point a monitor at and a closer the caller must call.
func openTCPPort(t *testing.T) (host string, port int32, closeFn func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	host, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("split host/port: %v", err)
	}
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		t.Fatalf("parse port: %v", err)
	}
	return host, port, func() { _ = ln.Close() }
}

// freeTCPPort returns a host/port that nothing is listening on — obtained by
// binding then immediately releasing, so it's a real ephemeral port rather
// than a guessed number, keeping the "closed" case honest.
func freeTCPPort(t *testing.T) (host string, port int32) {
	t.Helper()
	host, port, closeFn := openTCPPort(t)
	closeFn()
	return host, port
}

// listenOnPort starts a listener on a specific, already-known host:port —
// used to simulate a monitored service coming back up on the same address a
// freeTCPPort call previously released. Small window between free and
// re-bind; accepted the same way the rest of this file accepts real sockets
// over mocks for TCP checks.
func listenOnPort(t *testing.T, host string, port int32) (string, int32, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err != nil {
		t.Fatalf("listen on %s:%d: %v", host, port, err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	return host, port, func() { _ = ln.Close() }
}

func getCronStatus(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) string {
	t.Helper()
	var status string
	if err := pool.QueryRow(context.Background(), "SELECT status FROM cron_monitors WHERE id = $1", id).Scan(&status); err != nil {
		t.Fatalf("query cron status: %v", err)
	}
	return status
}

func getUptimeRow(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) (status string, consecutiveFailures int32) {
	t.Helper()
	if err := pool.QueryRow(context.Background(),
		"SELECT status, consecutive_failures FROM uptime_monitors WHERE id = $1", id,
	).Scan(&status, &consecutiveFailures); err != nil {
		t.Fatalf("query uptime row: %v", err)
	}
	return status, consecutiveFailures
}

func getPortRow(t *testing.T, pool *pgxpool.Pool, id uuid.UUID) (status string, consecutiveFailures int32) {
	t.Helper()
	if err := pool.QueryRow(context.Background(),
		"SELECT status, consecutive_failures FROM port_monitors WHERE id = $1", id,
	).Scan(&status, &consecutiveFailures); err != nil {
		t.Fatalf("query port row: %v", err)
	}
	return status, consecutiveFailures
}

func forceDueNow(t *testing.T, pool *pgxpool.Pool, table string, id uuid.UUID) {
	t.Helper()
	col := "next_check_at"
	if table == "cron_monitors" {
		col = "next_ping_at"
	}
	if _, err := pool.Exec(context.Background(), "UPDATE "+table+" SET "+col+" = NOW() WHERE id = $1", id); err != nil {
		t.Fatalf("force due: %v", err)
	}
}

// ─── pure function tests ──────────────────────────────────────────────────

func TestChannelConfigValue(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		key  string
		want string
	}{
		{"present", `{"chatId":"123"}`, "chatId", "123"},
		{"trims whitespace", `{"email":"  a@b.com  "}`, "email", "a@b.com"},
		{"missing key", `{"chatId":"123"}`, "email", ""},
		{"invalid JSON", `not json`, "chatId", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := channelConfigValue([]byte(tc.raw), tc.key); got != tc.want {
				t.Fatalf("channelConfigValue(%q, %q) = %q, want %q", tc.raw, tc.key, got, tc.want)
			}
		})
	}
}

func TestKeywordCheckPasses(t *testing.T) {
	cases := []struct {
		name          string
		body, keyword string
		mode          db.KeywordMode
		caseSensitive bool
		want          bool
	}{
		{"contains, present", "Welcome back", "Welcome", db.KeywordModeContains, false, true},
		{"contains, absent", "Goodbye", "Welcome", db.KeywordModeContains, false, false},
		{"contains, case-insensitive by default", "welcome back", "Welcome", db.KeywordModeContains, false, true},
		{"contains, case-sensitive mismatch", "welcome back", "Welcome", db.KeywordModeContains, true, false},
		{"not_contains, absent passes", "Goodbye", "Welcome", db.KeywordModeNotContains, false, true},
		{"not_contains, present fails", "Welcome back", "Welcome", db.KeywordModeNotContains, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := keywordCheckPasses(tc.body, tc.keyword, tc.mode, tc.caseSensitive); got != tc.want {
				t.Fatalf("keywordCheckPasses(...) = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestKeywordFailureReason(t *testing.T) {
	if got := keywordFailureReason(db.KeywordModeContains); got != "Keyword not found" {
		t.Fatalf("want 'Keyword not found', got %q", got)
	}
	if got := keywordFailureReason(db.KeywordModeNotContains); got != "Keyword found" {
		t.Fatalf("want 'Keyword found', got %q", got)
	}
}

func TestHTTPStatusDesc(t *testing.T) {
	if got := httpStatusDesc(0); got != "timeout / connection error" {
		t.Fatalf("want timeout description, got %q", got)
	}
	if got := httpStatusDesc(404); got != "HTTP 404" {
		t.Fatalf("want 'HTTP 404', got %q", got)
	}
}

// ─── performHTTPCheck (local httptest.Server, no live network) ───────────

func TestPerformHTTPCheck(t *testing.T) {
	t.Run("200 with no keyword is up", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		code, _, isUp, reason := performHTTPCheck(db.UptimeMonitor{Url: srv.URL}, &http.Client{Timeout: 10 * time.Second})
		if code != 200 || !isUp || reason != "" {
			t.Fatalf("want (200, up, \"\"), got (%d, %v, %q)", code, isUp, reason)
		}
	})

	t.Run("non-200 is down with the status code in the reason", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		code, _, isUp, reason := performHTTPCheck(db.UptimeMonitor{Url: srv.URL}, &http.Client{Timeout: 10 * time.Second})
		if code != 500 || isUp || reason != "HTTP 500" {
			t.Fatalf("want (500, down, HTTP 500), got (%d, %v, %q)", code, isUp, reason)
		}
	})

	t.Run("connection error is down with no status code", func(t *testing.T) {
		code, _, isUp, reason := performHTTPCheck(db.UptimeMonitor{Url: "http://127.0.0.1:1"}, &http.Client{Timeout: 10 * time.Second})
		if code != 0 || isUp || reason != "timeout / connection error" {
			t.Fatalf("want (0, down, timeout), got (%d, %v, %q)", code, isUp, reason)
		}
	})

	t.Run("keyword present (contains mode) is up", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, "Welcome back, friend")
		}))
		defer srv.Close()

		_, _, isUp, reason := performHTTPCheck(db.UptimeMonitor{
			Url: srv.URL, Keyword: pgtype.Text{String: "Welcome", Valid: true}, KeywordMode: db.KeywordModeContains,
		}, &http.Client{Timeout: 10 * time.Second})
		if !isUp || reason != "" {
			t.Fatalf("want up with no failure reason, got isUp=%v reason=%q", isUp, reason)
		}
	})

	t.Run("keyword absent (contains mode) is down", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, "Goodbye")
		}))
		defer srv.Close()

		_, _, isUp, reason := performHTTPCheck(db.UptimeMonitor{
			Url: srv.URL, Keyword: pgtype.Text{String: "Welcome", Valid: true}, KeywordMode: db.KeywordModeContains,
		}, &http.Client{Timeout: 10 * time.Second})
		if isUp || reason != "Keyword not found" {
			t.Fatalf("want down/Keyword not found, got isUp=%v reason=%q", isUp, reason)
		}
	})

	t.Run("keyword present (not_contains mode) is down", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, "Under maintenance")
		}))
		defer srv.Close()

		_, _, isUp, reason := performHTTPCheck(db.UptimeMonitor{
			Url: srv.URL, Keyword: pgtype.Text{String: "maintenance", Valid: true}, KeywordMode: db.KeywordModeNotContains,
		}, &http.Client{Timeout: 10 * time.Second})
		if isUp || reason != "Keyword found" {
			t.Fatalf("want down/Keyword found, got isUp=%v reason=%q", isUp, reason)
		}
	})

	t.Run("non-200 short-circuits before the keyword check", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprint(w, "Welcome back") // keyword would pass if checked
		}))
		defer srv.Close()

		_, _, isUp, reason := performHTTPCheck(db.UptimeMonitor{
			Url: srv.URL, Keyword: pgtype.Text{String: "Welcome", Valid: true}, KeywordMode: db.KeywordModeContains,
		}, &http.Client{Timeout: 10 * time.Second})
		if isUp || reason != "HTTP 503" {
			t.Fatalf("want the status code to win over the keyword, got isUp=%v reason=%q", isUp, reason)
		}
	})

	t.Run("a passing JSON assertion is up", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `{"status":"ok"}`)
		}))
		defer srv.Close()

		_, _, isUp, reason := performHTTPCheck(db.UptimeMonitor{
			Url:            srv.URL,
			JsonAssertions: []byte(`[{"path":"$.status","comparator":"equals","expected":"ok"}]`),
		}, &http.Client{Timeout: 10 * time.Second})
		if !isUp || reason != "" {
			t.Fatalf("want up with no failure reason, got isUp=%v reason=%q", isUp, reason)
		}
	})

	t.Run("a failing JSON assertion is down with the assertion failure as the reason", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, `{"status":"degraded"}`)
		}))
		defer srv.Close()

		_, _, isUp, reason := performHTTPCheck(db.UptimeMonitor{
			Url:            srv.URL,
			JsonAssertions: []byte(`[{"path":"$.status","comparator":"equals","expected":"ok"}]`),
		}, &http.Client{Timeout: 10 * time.Second})
		if isUp || !strings.Contains(reason, "JSON assertion failed") {
			t.Fatalf("want down with a JSON assertion failure reason, got isUp=%v reason=%q", isUp, reason)
		}
	})
}

// ─── JSON assertions (pure functions) ─────────────────────────────────────

func TestEvaluateJsonAssertion(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		a       jsonAssertion
		wantOK  bool
		wantMsg string // only checked when non-empty
	}{
		{
			name:   "equals passes",
			body:   `{"status":"ok"}`,
			a:      jsonAssertion{Path: "$.status", Comparator: "equals", Expected: "ok"},
			wantOK: true,
		},
		{
			name:    "equals fails",
			body:    `{"status":"degraded"}`,
			a:       jsonAssertion{Path: "$.status", Comparator: "equals", Expected: "ok"},
			wantOK:  false,
			wantMsg: `JSON assertion failed: "$.status" equals "ok" (got "degraded")`,
		},
		{
			name:   "not_equals passes",
			body:   `{"status":"degraded"}`,
			a:      jsonAssertion{Path: "$.status", Comparator: "not_equals", Expected: "ok"},
			wantOK: true,
		},
		{
			name:   "contains passes (default comparator)",
			body:   `{"message":"all systems nominal"}`,
			a:      jsonAssertion{Path: "$.message", Comparator: "contains", Expected: "nominal"},
			wantOK: true,
		},
		{
			name:   "nested path resolves",
			body:   `{"data":{"health":{"status":"ok"}}}`,
			a:      jsonAssertion{Path: "$.data.health.status", Comparator: "equals", Expected: "ok"},
			wantOK: true,
		},
		{
			name:    "missing path fails with a not-found reason",
			body:    `{"status":"ok"}`,
			a:       jsonAssertion{Path: "$.missing", Comparator: "equals", Expected: "ok"},
			wantOK:  false,
			wantMsg: `JSON path "$.missing" not found`,
		},
		{
			name:    "path through a non-object segment fails",
			body:    `{"status":"ok"}`,
			a:       jsonAssertion{Path: "$.status.nested", Comparator: "equals", Expected: "ok"},
			wantOK:  false,
			wantMsg: `JSON path "$.status.nested" not found`,
		},
		{
			name:    "invalid JSON body fails",
			body:    `not json`,
			a:       jsonAssertion{Path: "$.status", Comparator: "equals", Expected: "ok"},
			wantOK:  false,
			wantMsg: "response is not valid JSON",
		},
		{
			name:   "greater_than passes",
			body:   `{"uptime":99.95}`,
			a:      jsonAssertion{Path: "$.uptime", Comparator: "greater_than", Expected: "99"},
			wantOK: true,
		},
		{
			name:   "less_than fails",
			body:   `{"latency":250}`,
			a:      jsonAssertion{Path: "$.latency", Comparator: "less_than", Expected: "100"},
			wantOK: false,
		},
		{
			name:   "greater_than with a non-numeric actual value fails closed",
			body:   `{"count":"not-a-number"}`,
			a:      jsonAssertion{Path: "$.count", Comparator: "greater_than", Expected: "5"},
			wantOK: false,
		},
		{
			name:   "boolean value coerces to true/false",
			body:   `{"healthy":true}`,
			a:      jsonAssertion{Path: "$.healthy", Comparator: "equals", Expected: "true"},
			wantOK: true,
		},
		{
			name:   "null value coerces to the string null",
			body:   `{"error":null}`,
			a:      jsonAssertion{Path: "$.error", Comparator: "equals", Expected: "null"},
			wantOK: true,
		},
		{
			name:   "integer-valued float renders without a decimal point",
			body:   `{"count":42}`,
			a:      jsonAssertion{Path: "$.count", Comparator: "equals", Expected: "42"},
			wantOK: true,
		},
		{
			name:   "a leading '.' on the path is tolerated the same as '$.'",
			body:   `{"status":"ok"}`,
			a:      jsonAssertion{Path: ".status", Comparator: "equals", Expected: "ok"},
			wantOK: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, ok := evaluateJsonAssertion(tc.body, tc.a)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v (msg: %q)", ok, tc.wantOK, msg)
			}
			if tc.wantMsg != "" && msg != tc.wantMsg {
				t.Fatalf("msg = %q, want %q", msg, tc.wantMsg)
			}
		})
	}
}

func TestJsonValueToString(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"string", "hello", "hello"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"nil", nil, "null"},
		{"whole-number float", float64(42), "42"},
		{"fractional float", 3.14, "3.14"},
		{"array falls back to JSON encoding", []any{"a", "b"}, `["a","b"]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := jsonValueToString(tc.in); got != tc.want {
				t.Fatalf("jsonValueToString(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseNumericPair(t *testing.T) {
	if _, _, ok := parseNumericPair("12.5", "10"); !ok {
		t.Fatal("want ok for two valid numbers")
	}
	if _, _, ok := parseNumericPair("abc", "10"); ok {
		t.Fatal("want not-ok when actual isn't numeric")
	}
	if _, _, ok := parseNumericPair("10", "abc"); ok {
		t.Fatal("want not-ok when expected isn't numeric")
	}
}

func TestPerformTLSCheck(t *testing.T) {
	t.Run("connection error", func(t *testing.T) {
		// .invalid is reserved (RFC 2606) to never resolve — a hermetic way
		// to force a DNS failure without depending on a live host being down.
		_, _, _, err := performTLSCheck("this-host-does-not-exist.invalid")
		if err == nil {
			t.Fatal("want an error for an unresolvable host")
		}
	})
}

// ─── dispatchAlert ─────────────────────────────────────────────────────────

func TestDispatchAlert(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	tg := telegram.NewClient("")  // no token: SendMessage always errors, no network call
	mailer := email.NewSender("") // no API key: SendAlertEmail no-ops with a nil error (ADR-012)
	// Webhook/Slack subtests below POST to a local httptest server, which
	// NewClient's SSRF protections (loopback blocked) would refuse to dial —
	// not exercised here, see webhook_test.go.
	wh := webhook.NewClientWithHTTPClient(&http.Client{Timeout: 10 * time.Second})
	sl := slack.NewClientWithHTTPClient(&http.Client{Timeout: 10 * time.Second})
	sm := twilio.NewClient("", "", "", "") // not configured: Send always errors, no network call
	logger := testLogger()
	n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, Slack: sl, SMS: sm, Logger: logger}
	msg := AlertMessage{Telegram: "down", EmailSubject: "down", EmailHTML: "<p>down</p>"}

	t.Run("no channel attached and no org user falls back to nothing", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: uuid.New()}, msg)
		if sent {
			t.Fatal("want no alert delivered with no channel attached and no org user")
		}
	})

	t.Run("no channel attached falls back to emailing every org user (ADR-023)", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		testUser(t, queries, org.ID)
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: uuid.New()}, msg)
		if !sent {
			t.Fatal("want the fallback email to org users to count as delivered")
		}
	})

	t.Run("telegram channel attached but unreachable in this environment does not count as delivered", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeTelegram, map[string]string{"chatId": "123"})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, msg)
		if sent {
			t.Fatal("want telegram-only delivery to fail with no bot token configured, and the email fallback to also report unsent since this org has no user")
		}
	})

	t.Run("email channel succeeds (no-op-safe in dev) even when a telegram channel on the same monitor fails", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		tgChannel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeTelegram, map[string]string{"chatId": "123"})
		emailChannel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeEmail, map[string]string{"email": "a@b.com"})
		attachNotificationChannel(t, queries, tgChannel.ID, "cron", monitorID)
		attachNotificationChannel(t, queries, emailChannel.ID, "cron", monitorID)

		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, msg)
		if !sent {
			t.Fatal("want the email channel's success to count as delivered, even though the telegram channel failed")
		}
	})

	t.Run("webhook channel without a Webhook event on the message does not count as delivered", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeWebhook, map[string]string{"url": "https://example.com/hook", "secret": "shh"})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		// msg has no Webhook field set — call sites that haven't been
		// updated to build one shouldn't crash, just skip the channel.
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, msg)
		if sent {
			t.Fatal("want no delivery when the message carries no webhook event")
		}
	})

	t.Run("webhook channel delivers and records success on the channel row", func(t *testing.T) {
		var gotSig string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotSig = r.Header.Get(webhook.SignatureHeader)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeWebhook, map[string]string{"url": srv.URL, "secret": "shh"})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		webhookMsg := AlertMessage{Webhook: &webhook.Event{EventType: "down", MonitorName: "X", MonitorType: "cron"}}
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, webhookMsg)
		if !sent {
			t.Fatal("want the webhook delivery to count as sent")
		}
		if gotSig == "" {
			t.Fatal("want the request to carry a signature header")
		}

		updated, err := queries.GetNotificationChannel(context.Background(), db.GetNotificationChannelParams{ID: channel.ID, OrgID: org.ID})
		if err != nil {
			t.Fatalf("get channel: %v", err)
		}
		if updated.LastDeliveryStatus.String != "success" {
			t.Fatalf("want last_delivery_status success, got %q", updated.LastDeliveryStatus.String)
		}
		if updated.LastDeliveryDetail.String != "200" {
			t.Fatalf("want last_delivery_detail 200, got %q", updated.LastDeliveryDetail.String)
		}
		if !updated.LastDeliveryAt.Valid {
			t.Fatal("want last_delivery_at set")
		}
	})

	t.Run("webhook channel records failure on a non-2xx response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeWebhook, map[string]string{"url": srv.URL, "secret": "shh"})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		webhookMsg := AlertMessage{Webhook: &webhook.Event{EventType: "down", MonitorName: "X", MonitorType: "cron"}}
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, webhookMsg)
		if sent {
			t.Fatal("want a 500 response to not count as delivered")
		}

		updated, err := queries.GetNotificationChannel(context.Background(), db.GetNotificationChannelParams{ID: channel.ID, OrgID: org.ID})
		if err != nil {
			t.Fatalf("get channel: %v", err)
		}
		if updated.LastDeliveryStatus.String != "failed" {
			t.Fatalf("want last_delivery_status failed, got %q", updated.LastDeliveryStatus.String)
		}
		if updated.LastDeliveryDetail.String != "500" {
			t.Fatalf("want last_delivery_detail 500, got %q", updated.LastDeliveryDetail.String)
		}
	})

	t.Run("slack channel without a Slack message on the alert does not count as delivered", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeSlack, map[string]string{"url": "https://hooks.slack.com/services/x"})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		// msg has no Slack field set — same guard as the webhook case above.
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, msg)
		if sent {
			t.Fatal("want no delivery when the message carries no Slack payload")
		}
	})

	t.Run("slack channel delivers and records success on the channel row", func(t *testing.T) {
		var gotBody slack.Message
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeSlack, map[string]string{"url": srv.URL})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		alertMsg := AlertMessage{Slack: &slack.Message{Text: "down"}}
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, alertMsg)
		if !sent {
			t.Fatal("want the Slack delivery to count as sent")
		}
		if gotBody.Text != "down" {
			t.Fatalf("want the posted body to carry the message text, got %+v", gotBody)
		}

		updated, err := queries.GetNotificationChannel(context.Background(), db.GetNotificationChannelParams{ID: channel.ID, OrgID: org.ID})
		if err != nil {
			t.Fatalf("get channel: %v", err)
		}
		if updated.LastDeliveryStatus.String != "success" {
			t.Fatalf("want last_delivery_status success, got %q", updated.LastDeliveryStatus.String)
		}
		if updated.LastDeliveryDetail.String != "200" {
			t.Fatalf("want last_delivery_detail 200, got %q", updated.LastDeliveryDetail.String)
		}
	})

	t.Run("slack channel records failure on a non-2xx response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeSlack, map[string]string{"url": srv.URL})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		alertMsg := AlertMessage{Slack: &slack.Message{Text: "down"}}
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, alertMsg)
		if sent {
			t.Fatal("want a 500 response to not count as delivered")
		}

		updated, err := queries.GetNotificationChannel(context.Background(), db.GetNotificationChannelParams{ID: channel.ID, OrgID: org.ID})
		if err != nil {
			t.Fatalf("get channel: %v", err)
		}
		if updated.LastDeliveryStatus.String != "failed" {
			t.Fatalf("want last_delivery_status failed, got %q", updated.LastDeliveryStatus.String)
		}
		if updated.LastDeliveryDetail.String != "500" {
			t.Fatalf("want last_delivery_detail 500, got %q", updated.LastDeliveryDetail.String)
		}
	})

	t.Run("slack channel with no configured URL does not count as delivered", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeSlack, map[string]string{})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		alertMsg := AlertMessage{Slack: &slack.Message{Text: "down"}}
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, alertMsg)
		if sent {
			t.Fatal("want no delivery when the channel has no configured URL")
		}
	})

	t.Run("sms channel without an SMS string on the alert does not count as delivered", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeSms, map[string]string{"phone_number": "+15005550006"})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		// msg has no SMS field set — same guard as the webhook/slack cases above.
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, msg)
		if sent {
			t.Fatal("want no delivery when the message carries no SMS text")
		}
	})

	t.Run("sms channel attempts a send and records failure when Twilio isn't configured", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		upgradeOrgPlan(t, queries, org.ID, db.PlanSolo) // Hobby has 0 sms credits — would fail on quota before ever reaching Twilio
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeSms, map[string]string{"phone_number": "+15005550006"})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		smsMsg := AlertMessage{SMS: "checkmeup: X is DOWN"}
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, smsMsg)
		if sent {
			t.Fatal("want no delivery with Twilio unconfigured in this test")
		}

		updated, err := queries.GetNotificationChannel(context.Background(), db.GetNotificationChannelParams{ID: channel.ID, OrgID: org.ID})
		if err != nil {
			t.Fatalf("get channel: %v", err)
		}
		if updated.LastDeliveryStatus.String != "failed" {
			t.Fatalf("want last_delivery_status failed, got %q", updated.LastDeliveryStatus.String)
		}
		if updated.LastDeliveryDetail.String != "timeout / connection error" {
			t.Fatalf("want the delivery to fail at the Twilio call itself (not quota), got detail %q", updated.LastDeliveryDetail.String)
		}
	})

	t.Run("sms channel on the Hobby plan is skipped for quota exhaustion without ever reaching Twilio", func(t *testing.T) {
		org := testOrg(t, queries, pool) // default plan: Hobby, 0 sms credits (ADR-032)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeSms, map[string]string{"phone_number": "+15005550006"})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		smsMsg := AlertMessage{SMS: "checkmeup: X is DOWN"}
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, smsMsg)
		if sent {
			t.Fatal("want no delivery on a 0-credit plan")
		}

		updated, err := queries.GetNotificationChannel(context.Background(), db.GetNotificationChannelParams{ID: channel.ID, OrgID: org.ID})
		if err != nil {
			t.Fatalf("get channel: %v", err)
		}
		if updated.LastDeliveryDetail.String != "sms credit quota exhausted" {
			t.Fatalf("want detail \"sms credit quota exhausted\", got %q", updated.LastDeliveryDetail.String)
		}
	})

	t.Run("sms send within the monthly quota consumes exactly one credit", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		upgradeOrgPlan(t, queries, org.ID, db.PlanSolo) // 10 credits/month
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeSms, map[string]string{"phone_number": "+15005550006"})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		smsMsg := AlertMessage{SMS: "checkmeup: X is DOWN"}
		DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, smsMsg)

		info, err := queries.GetOrgBillingInfo(context.Background(), org.ID)
		if err != nil {
			t.Fatalf("get billing info: %v", err)
		}
		if info.SmsCreditsUsedThisMonth != 1 {
			t.Fatalf("want 1 credit consumed, got %d", info.SmsCreditsUsedThisMonth)
		}
	})

	t.Run("sms channel with no configured phone number does not count as delivered", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeSms, map[string]string{})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		smsMsg := AlertMessage{SMS: "checkmeup: X is DOWN"}
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, smsMsg)
		if sent {
			t.Fatal("want no delivery when the channel has no configured phone number")
		}
	})

	t.Run("every attached channel failing falls back to org user email (generalized ADR-032 fallback)", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		testUser(t, queries, org.ID)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeSms, map[string]string{"phone_number": "+15005550006"})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		smsMsg := AlertMessage{SMS: "checkmeup: X is DOWN", EmailSubject: "down", EmailHTML: "<p>down</p>"}
		sent := DispatchAlert(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: monitorID}, smsMsg)
		if !sent {
			t.Fatal("want the email fallback to count as delivered when the only attached channel (sms) fails")
		}
	})
}

func TestConsumeSMSCredit(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)

	t.Run("blocks once the plan limit is reached, without going negative or over", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		limit := int32(3)
		for i := range limit {
			if _, err := queries.ConsumeSMSCredit(context.Background(), db.ConsumeSMSCreditParams{ID: org.ID, CreditCost: 1, CreditLimit: limit}); err != nil {
				t.Fatalf("credit %d: want success, got %v", i+1, err)
			}
		}
		if _, err := queries.ConsumeSMSCredit(context.Background(), db.ConsumeSMSCreditParams{ID: org.ID, CreditCost: 1, CreditLimit: limit}); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("want pgx.ErrNoRows once the limit is reached, got %v", err)
		}

		info, err := queries.GetOrgBillingInfo(context.Background(), org.ID)
		if err != nil {
			t.Fatalf("get billing info: %v", err)
		}
		if info.SmsCreditsUsedThisMonth != limit {
			t.Fatalf("want used stuck at %d, got %d", limit, info.SmsCreditsUsedThisMonth)
		}
	})

	t.Run("a limit of 0 blocks every send (Hobby plan)", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		if _, err := queries.ConsumeSMSCredit(context.Background(), db.ConsumeSMSCreditParams{ID: org.ID, CreditCost: 1, CreditLimit: 0}); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("want pgx.ErrNoRows for a 0 credit limit, got %v", err)
		}
	})

	t.Run("lazily resets once the reset date has passed", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		limit := int32(2)
		for range limit {
			if _, err := queries.ConsumeSMSCredit(context.Background(), db.ConsumeSMSCreditParams{ID: org.ID, CreditCost: 1, CreditLimit: limit}); err != nil {
				t.Fatalf("want success while under the limit, got %v", err)
			}
		}
		if _, err := queries.ConsumeSMSCredit(context.Background(), db.ConsumeSMSCreditParams{ID: org.ID, CreditCost: 1, CreditLimit: limit}); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("want the limit reached before the reset, got %v", err)
		}

		// Simulate the reset date having passed (no dedicated query for
		// this — it's not something production code ever sets directly).
		if _, err := pool.Exec(context.Background(), "UPDATE orgs SET sms_credits_reset_at = CURRENT_DATE - 1 WHERE id = $1", org.ID); err != nil {
			t.Fatalf("force reset date: %v", err)
		}

		used, err := queries.ConsumeSMSCredit(context.Background(), db.ConsumeSMSCreditParams{ID: org.ID, CreditCost: 1, CreditLimit: limit})
		if err != nil {
			t.Fatalf("want success after the reset date passes, got %v", err)
		}
		if used != 1 {
			t.Fatalf("want usage reset to 1 (this send), got %d", used)
		}

		info, err := queries.GetOrgBillingInfo(context.Background(), org.ID)
		if err != nil {
			t.Fatalf("get billing info: %v", err)
		}
		if !info.SmsCreditsResetAt.Valid || !info.SmsCreditsResetAt.Time.After(time.Now()) {
			t.Fatalf("want a fresh future reset date, got %+v", info.SmsCreditsResetAt)
		}
	})
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{45 * time.Second, "45s"},
		{90 * time.Second, "1m 30s"},
		{61 * time.Minute, "1h 1m"},
		{0, "0s"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if got := FormatDuration(tc.in); got != tc.want {
				t.Fatalf("FormatDuration(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestTruncateSMS(t *testing.T) {
	t.Run("short message passes through unchanged", func(t *testing.T) {
		if got := TruncateSMS("checkmeup: X is DOWN"); got != "checkmeup: X is DOWN" {
			t.Fatalf("want unchanged, got %q", got)
		}
	})

	t.Run("message over 160 chars is truncated with an ellipsis", func(t *testing.T) {
		long := strings.Repeat("a", 200)
		got := TruncateSMS(long)
		if len([]rune(got)) != smsSegmentLimit {
			t.Fatalf("want length %d, got %d", smsSegmentLimit, len([]rune(got)))
		}
		if !strings.HasSuffix(got, "…") {
			t.Fatalf("want truncated message to end with an ellipsis, got %q", got)
		}
	})

	t.Run("multi-byte characters are cut by rune, not by byte", func(t *testing.T) {
		long := strings.Repeat("é", 200) // é is 2 bytes in UTF-8, 1 rune
		got := TruncateSMS(long)
		for _, r := range got {
			if r != 'é' && r != '…' {
				t.Fatalf("want only é or the ellipsis rune, got a mangled rune %q in %q", r, got)
			}
		}
	})
}

// ─── checkOverdue (cron) ───────────────────────────────────────────────────

func TestCheckOverdue(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	tg := telegram.NewClient("")
	mailer := email.NewSender("")
	wh := webhook.NewClient()
	logger := testLogger()
	n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, Logger: logger}

	t.Run("marks an overdue monitor down, opens an incident, and delivers an alert", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		mon := testCronMonitor(t, queries, org.ID, "every 1h")
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeEmail, map[string]string{"email": "a@b.com"})
		attachNotificationChannel(t, queries, channel.ID, "cron", mon.ID)
		mustExecWorker(t, pool, "UPDATE cron_monitors SET status = 'up', next_ping_at = NOW() - INTERVAL '1 hour' WHERE id = $1", mon.ID)

		checkOverdue(context.Background(), n)

		if status := getCronStatus(t, pool, mon.ID); status != "down" {
			t.Fatalf("want status down, got %q", status)
		}
		var alertCount int32
		var resolvedAt pgtype.Timestamptz
		if err := pool.QueryRow(context.Background(),
			"SELECT alert_count, resolved_at FROM cron_incidents WHERE monitor_id = $1", mon.ID,
		).Scan(&alertCount, &resolvedAt); err != nil {
			t.Fatalf("query incident: %v", err)
		}
		if resolvedAt.Valid {
			t.Fatal("want a freshly opened, unresolved incident")
		}
		if alertCount != 1 {
			t.Fatalf("want alert_count 1 (delivered via the email channel), got %d", alertCount)
		}
	})

	t.Run("an incident still opens when alerts are disabled, but no alert is sent", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		mon := testCronMonitor(t, queries, org.ID, "every 1h")
		mustExecWorker(t, pool, "UPDATE cron_monitors SET status = 'up', next_ping_at = NOW() - INTERVAL '1 hour', alerts_enabled = false WHERE id = $1", mon.ID)

		checkOverdue(context.Background(), n)

		if status := getCronStatus(t, pool, mon.ID); status != "down" {
			t.Fatalf("want status down, got %q", status)
		}
		var alertCount int32
		if err := pool.QueryRow(context.Background(), "SELECT alert_count FROM cron_incidents WHERE monitor_id = $1", mon.ID).Scan(&alertCount); err != nil {
			t.Fatalf("query incident: %v", err)
		}
		if alertCount != 0 {
			t.Fatalf("want no alert sent with alerts disabled, got alert_count %d", alertCount)
		}
	})

	t.Run("a monitor under an active maintenance window is excluded", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		mon := testCronMonitor(t, queries, org.ID, "every 1h")
		mustExecWorker(t, pool, "UPDATE cron_monitors SET status = 'up', next_ping_at = NOW() - INTERVAL '1 hour' WHERE id = $1", mon.ID)

		var windowID uuid.UUID
		if err := pool.QueryRow(context.Background(),
			"INSERT INTO maintenance_windows (org_id, title, message, starts_at) VALUES ($1, 'Scheduled', '', NOW() - INTERVAL '1 minute') RETURNING id",
			org.ID,
		).Scan(&windowID); err != nil {
			t.Fatalf("seed maintenance window: %v", err)
		}
		mustExecWorker(t, pool, "INSERT INTO maintenance_window_monitors (window_id, monitor_type, monitor_id) VALUES ($1, 'cron', $2)", windowID, mon.ID)

		checkOverdue(context.Background(), n)

		if status := getCronStatus(t, pool, mon.ID); status != "up" {
			t.Fatalf("want status to remain up under maintenance, got %q", status)
		}
		var incidentCount int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM cron_incidents WHERE monitor_id = $1", mon.ID).Scan(&incidentCount); err != nil {
			t.Fatalf("count incidents: %v", err)
		}
		if incidentCount != 0 {
			t.Fatalf("want no incident opened under maintenance, got %d", incidentCount)
		}
	})

	t.Run("a monitor that isn't overdue is left alone", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		mon := testCronMonitor(t, queries, org.ID, "every 1h")
		// Freshly created: status 'waiting', not 'up' — never selected by
		// ListOverdueCronMonitors regardless of next_ping_at.
		checkOverdue(context.Background(), n)
		if status := getCronStatus(t, pool, mon.ID); status != "waiting" {
			t.Fatalf("want status unchanged (waiting), got %q", status)
		}
	})
}

// ─── checkUptimeMonitors / checkOneUptimeMonitor ──────────────────────────

func TestCheckUptimeMonitors(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	tg := telegram.NewClient("")
	mailer := email.NewSender("")
	// The recovery-webhook subtest below POSTs to a local httptest server,
	// which NewClient's SSRF protections (loopback blocked) would refuse to
	// dial — not exercised here, see webhook_test.go.
	wh := webhook.NewClientWithHTTPClient(&http.Client{Timeout: 10 * time.Second})
	logger := testLogger()
	n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, Logger: logger, HTTPClient: &http.Client{Timeout: 10 * time.Second}}

	t.Run("escalates to down after two consecutive failures, then recovers", func(t *testing.T) {
		var up atomic.Bool
		up.Store(false)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if up.Load() {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}))
		defer srv.Close()

		org := testOrg(t, queries, pool)
		// Use AlertAfterNFailures=1 to test "skip first failure, alert on second" behavior.
		mon, err := queries.CreateUptimeMonitor(context.Background(), db.CreateUptimeMonitorParams{
			OrgID: org.ID, Name: "Uptime monitor", Url: srv.URL, IntervalMins: 10,
			MaxAlertsPerIncident: 3, AlertAfterNFailures: 1,
			KeywordMode: db.KeywordModeContains, JsonAssertions: []byte("[]"),
		})
		if err != nil {
			t.Fatalf("create uptime monitor: %v", err)
		}
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeEmail, map[string]string{"email": "a@b.com"})
		attachNotificationChannel(t, queries, channel.ID, "uptime", mon.ID)

		// First failure: recorded, but alert_after_n_failures=1 suppresses the alert.
		checkUptimeMonitors(context.Background(), n)
		if status, failures := getUptimeRow(t, pool, mon.ID); status == "down" || failures != 1 {
			t.Fatalf("after 1st failure: want not-down with 1 consecutive failure, got status=%q failures=%d", status, failures)
		}

		// Second consecutive failure: trips down, opens an incident, alerts.
		forceDueNow(t, pool, "uptime_monitors", mon.ID)
		checkUptimeMonitors(context.Background(), n)
		status, failures := getUptimeRow(t, pool, mon.ID)
		if status != "down" || failures != 2 {
			t.Fatalf("after 2nd failure: want down with 2 consecutive failures, got status=%q failures=%d", status, failures)
		}
		var alertCount int32
		var resolvedAt pgtype.Timestamptz
		if err := pool.QueryRow(context.Background(),
			"SELECT alert_count, resolved_at FROM uptime_incidents WHERE monitor_id = $1", mon.ID,
		).Scan(&alertCount, &resolvedAt); err != nil {
			t.Fatalf("query incident: %v", err)
		}
		if resolvedAt.Valid {
			t.Fatal("want an unresolved incident after going down")
		}
		if alertCount != 1 {
			t.Fatalf("want 1 alert delivered, got %d", alertCount)
		}

		// Recovery: status flips back, consecutive_failures resets, incident resolves.
		up.Store(true)
		forceDueNow(t, pool, "uptime_monitors", mon.ID)
		checkUptimeMonitors(context.Background(), n)
		status, failures = getUptimeRow(t, pool, mon.ID)
		if status != "up" || failures != 0 {
			t.Fatalf("after recovery: want up with 0 consecutive failures, got status=%q failures=%d", status, failures)
		}
		if err := pool.QueryRow(context.Background(),
			"SELECT resolved_at FROM uptime_incidents WHERE monitor_id = $1", mon.ID,
		).Scan(&resolvedAt); err != nil {
			t.Fatalf("query incident: %v", err)
		}
		if !resolvedAt.Valid {
			t.Fatal("want the incident resolved after recovery")
		}
	})

	t.Run("delivers a recovery webhook event with a non-zero downtime duration", func(t *testing.T) {
		var up atomic.Bool
		up.Store(false)
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if up.Load() {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}))
		defer target.Close()

		var gotEvents []webhook.Event
		hookSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var e webhook.Event
			_ = json.NewDecoder(r.Body).Decode(&e)
			gotEvents = append(gotEvents, e)
			w.WriteHeader(http.StatusOK)
		}))
		defer hookSrv.Close()

		org := testOrg(t, queries, pool)
		mon := testUptimeMonitor(t, queries, org.ID, target.URL)
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeWebhook, map[string]string{"url": hookSrv.URL, "secret": "shh"})
		attachNotificationChannel(t, queries, channel.ID, "uptime", mon.ID)

		checkUptimeMonitors(context.Background(), n)
		forceDueNow(t, pool, "uptime_monitors", mon.ID)
		checkUptimeMonitors(context.Background(), n) // 2nd failure trips down + webhook

		up.Store(true)
		forceDueNow(t, pool, "uptime_monitors", mon.ID)
		checkUptimeMonitors(context.Background(), n) // recovery + webhook

		if len(gotEvents) != 2 {
			t.Fatalf("want 2 webhook events (down, recovery), got %d: %+v", len(gotEvents), gotEvents)
		}
		if gotEvents[0].EventType != "down" || gotEvents[0].Reason == "" {
			t.Fatalf("want a down event with a reason, got %+v", gotEvents[0])
		}
		if gotEvents[1].EventType != "recovery" || gotEvents[1].DowntimeDuration == "" {
			t.Fatalf("want a recovery event with a non-empty downtime duration, got %+v", gotEvents[1])
		}
	})

	t.Run("records a check row for every poll", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		org := testOrg(t, queries, pool)
		mon := testUptimeMonitor(t, queries, org.ID, srv.URL)
		checkUptimeMonitors(context.Background(), n)

		var count int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM uptime_checks WHERE monitor_id = $1", mon.ID).Scan(&count); err != nil {
			t.Fatalf("count checks: %v", err)
		}
		if count != 1 {
			t.Fatalf("want 1 check recorded, got %d", count)
		}
	})

	t.Run("a monitor under an active maintenance window is excluded", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		org := testOrg(t, queries, pool)
		mon := testUptimeMonitor(t, queries, org.ID, srv.URL)
		var windowID uuid.UUID
		if err := pool.QueryRow(context.Background(),
			"INSERT INTO maintenance_windows (org_id, title, message, starts_at) VALUES ($1, 'Scheduled', '', NOW() - INTERVAL '1 minute') RETURNING id",
			org.ID,
		).Scan(&windowID); err != nil {
			t.Fatalf("seed maintenance window: %v", err)
		}
		mustExecWorker(t, pool, "INSERT INTO maintenance_window_monitors (window_id, monitor_type, monitor_id) VALUES ($1, 'uptime', $2)", windowID, mon.ID)

		checkUptimeMonitors(context.Background(), n)

		var count int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM uptime_checks WHERE monitor_id = $1", mon.ID).Scan(&count); err != nil {
			t.Fatalf("count checks: %v", err)
		}
		if count != 0 {
			t.Fatalf("want no check performed under maintenance, got %d", count)
		}
	})
}

// ─── SSL alert building (pure functions) ──────────────────────────────────

func TestSslCrossedThreshold(t *testing.T) {
	t.Run("crosses the 7-day threshold and sets alerted7d", func(t *testing.T) {
		alerted30d, alerted14d, alerted7d := false, false, false
		if !sslCrossedThreshold(5, &alerted30d, &alerted14d, &alerted7d) {
			t.Fatal("want a crossing at 5 days left")
		}
		if !alerted7d || alerted14d || alerted30d {
			t.Fatalf("want only alerted7d set, got 30d=%v 14d=%v 7d=%v", alerted30d, alerted14d, alerted7d)
		}
	})

	t.Run("expired (negative days) also crosses the 7-day threshold", func(t *testing.T) {
		alerted30d, alerted14d, alerted7d := false, false, false
		if !sslCrossedThreshold(-1, &alerted30d, &alerted14d, &alerted7d) {
			t.Fatal("want a crossing when already expired")
		}
		if !alerted7d {
			t.Fatal("want alerted7d set for an expired cert")
		}
	})

	t.Run("does not re-alert once every threshold up to daysLeft has already fired", func(t *testing.T) {
		// Falling through to a higher, not-yet-alerted threshold (14d, 30d)
		// is itself a crossing, so all three flags need to already be set
		// for daysLeft=3 to genuinely produce no crossing.
		alerted30d, alerted14d, alerted7d := true, true, true
		if sslCrossedThreshold(3, &alerted30d, &alerted14d, &alerted7d) {
			t.Fatal("want no crossing once every threshold is already alerted")
		}
	})

	t.Run("crosses the 14-day threshold without touching 7d", func(t *testing.T) {
		alerted30d, alerted14d, alerted7d := false, false, false
		if !sslCrossedThreshold(10, &alerted30d, &alerted14d, &alerted7d) {
			t.Fatal("want a crossing at 10 days left")
		}
		if !alerted14d || alerted7d {
			t.Fatalf("want only alerted14d set, got 14d=%v 7d=%v", alerted14d, alerted7d)
		}
	})

	t.Run("crosses the 30-day threshold without touching 14d or 7d", func(t *testing.T) {
		alerted30d, alerted14d, alerted7d := false, false, false
		if !sslCrossedThreshold(25, &alerted30d, &alerted14d, &alerted7d) {
			t.Fatal("want a crossing at 25 days left")
		}
		if !alerted30d || alerted14d || alerted7d {
			t.Fatalf("want only alerted30d set, got 30d=%v 14d=%v 7d=%v", alerted30d, alerted14d, alerted7d)
		}
	})

	t.Run("beyond 30 days is not a crossing", func(t *testing.T) {
		alerted30d, alerted14d, alerted7d := false, false, false
		if sslCrossedThreshold(45, &alerted30d, &alerted14d, &alerted7d) {
			t.Fatal("want no crossing beyond the 30-day threshold")
		}
	})
}

func TestSslExpiredMessages(t *testing.T) {
	m := db.SslMonitor{Name: "Prod API", Hostname: "api.example.com"}
	subject, telegramMsg, emailHTML := sslExpiredMessages(m, "2026-01-01")

	if !strings.Contains(subject, "Prod API") || !strings.Contains(subject, "expired") {
		t.Fatalf("want subject to name the monitor and say expired, got %q", subject)
	}
	if !strings.Contains(telegramMsg, "api.example.com") || !strings.Contains(telegramMsg, "2026-01-01") {
		t.Fatalf("want telegram message to include hostname and date, got %q", telegramMsg)
	}
	if !strings.Contains(emailHTML, "api.example.com") {
		t.Fatalf("want email HTML to include hostname, got %q", emailHTML)
	}
}

func TestSslExpiringSoonMessages(t *testing.T) {
	m := db.SslMonitor{Name: "Prod API", Hostname: "api.example.com"}
	subject, telegramMsg, emailHTML := sslExpiringSoonMessages(m, 7, "2026-01-08")

	if !strings.Contains(subject, "Prod API") || !strings.Contains(subject, "7 days") {
		t.Fatalf("want subject to name the monitor and say 7 days, got %q", subject)
	}
	if !strings.Contains(telegramMsg, "api.example.com") || !strings.Contains(telegramMsg, "7 days") {
		t.Fatalf("want telegram message to include hostname and day count, got %q", telegramMsg)
	}
	if !strings.Contains(emailHTML, "2026-01-08") {
		t.Fatalf("want email HTML to include the expiry date, got %q", emailHTML)
	}
}

func TestSslThresholdAlert(t *testing.T) {
	m := db.SslMonitor{Name: "Prod API", Hostname: "api.example.com"}
	expiresAt := time.Now().Add(5 * 24 * time.Hour)

	t.Run("returns nil when no new threshold is crossed", func(t *testing.T) {
		alerted30d, alerted14d, alerted7d := true, true, true // already alerted at every threshold
		alert := sslThresholdAlert(m, 5, expiresAt, &alerted30d, &alerted14d, &alerted7d)
		if alert != nil {
			t.Fatalf("want nil, got %+v", alert)
		}
	})

	t.Run("builds an expired-phrased alert for negative days left", func(t *testing.T) {
		alerted30d, alerted14d, alerted7d := false, false, false
		alert := sslThresholdAlert(m, -2, expiresAt, &alerted30d, &alerted14d, &alerted7d)
		if alert == nil {
			t.Fatal("want a non-nil alert")
		}
		if !strings.Contains(alert.EmailSubject, "expired") {
			t.Fatalf("want an 'expired' subject, got %q", alert.EmailSubject)
		}
		if alert.Webhook == nil || alert.Webhook.EventType != "down" || alert.Webhook.MonitorType != "ssl" {
			t.Fatalf("want a down/ssl webhook event, got %+v", alert.Webhook)
		}
		if alert.Slack == nil {
			t.Fatal("want a non-nil Slack message")
		}
	})

	t.Run("builds an expiring-soon-phrased alert for non-negative days left", func(t *testing.T) {
		alerted30d, alerted14d, alerted7d := false, false, false
		alert := sslThresholdAlert(m, 5, expiresAt, &alerted30d, &alerted14d, &alerted7d)
		if alert == nil {
			t.Fatal("want a non-nil alert")
		}
		if !strings.Contains(alert.EmailSubject, "expires in 5 days") {
			t.Fatalf("want an 'expires in 5 days' subject, got %q", alert.EmailSubject)
		}
	})
}

func TestBuildSSLCheckResult(t *testing.T) {
	baseMonitor := db.SslMonitor{
		Name: "Prod API", Hostname: "api.example.com",
		AlertsEnabled: true, MaxAlertsPerIncident: 3,
	}

	t.Run("a check error sets status error and increments consecutive failures", func(t *testing.T) {
		m := baseMonitor
		m.ConsecutiveFailures = 2
		r := buildSSLCheckResult(m, time.Time{}, "", 0, fmt.Errorf("dial tcp: connection refused"))

		if r.status != db.SslMonitorStatusError {
			t.Fatalf("want status error, got %q", r.status)
		}
		if !r.errorMsgParam.Valid || r.errorMsgParam.String == "" {
			t.Fatal("want a non-empty error message")
		}
		if r.consecutiveFailures != 3 {
			t.Fatalf("want consecutiveFailures 3, got %d", r.consecutiveFailures)
		}
		if r.alert != nil {
			t.Fatal("want no alert on a check-error result")
		}
	})

	t.Run("more than 30 days left is up and resets alert state", func(t *testing.T) {
		m := baseMonitor
		m.Alerted30d, m.Alerted14d, m.Alerted7d = true, true, true
		m.ConsecutiveFailures = 4
		m.AlertCount = 2
		expiresAt := time.Now().Add(60 * 24 * time.Hour)

		r := buildSSLCheckResult(m, expiresAt, "Let's Encrypt", 60, nil)

		if r.status != db.SslMonitorStatusUp {
			t.Fatalf("want status up, got %q", r.status)
		}
		if r.alerted30d || r.alerted14d || r.alerted7d {
			t.Fatal("want all alerted flags reset on renewal")
		}
		if r.consecutiveFailures != 0 || r.alertCount != 0 {
			t.Fatalf("want consecutiveFailures and alertCount reset to 0, got %d and %d", r.consecutiveFailures, r.alertCount)
		}
		if r.alert != nil {
			t.Fatal("want no alert when the cert is comfortably valid")
		}
	})

	t.Run("expiring soon with alerts enabled crosses the threshold and alerts", func(t *testing.T) {
		m := baseMonitor
		expiresAt := time.Now().Add(5 * 24 * time.Hour)

		r := buildSSLCheckResult(m, expiresAt, "Let's Encrypt", 5, nil)

		if r.status != db.SslMonitorStatusExpiringSoon {
			t.Fatalf("want status expiring_soon, got %q", r.status)
		}
		if r.alert == nil {
			t.Fatal("want an alert on first crossing of the 7-day threshold")
		}
		if r.alertCount != 1 {
			t.Fatalf("want alertCount 1, got %d", r.alertCount)
		}
	})

	t.Run("expired sets status expired", func(t *testing.T) {
		m := baseMonitor
		expiresAt := time.Now().Add(-24 * time.Hour)

		r := buildSSLCheckResult(m, expiresAt, "Let's Encrypt", -1, nil)

		if r.status != db.SslMonitorStatusExpired {
			t.Fatalf("want status expired, got %q", r.status)
		}
	})

	t.Run("alerts disabled crosses the threshold but sends no alert", func(t *testing.T) {
		m := baseMonitor
		m.AlertsEnabled = false
		expiresAt := time.Now().Add(5 * 24 * time.Hour)

		r := buildSSLCheckResult(m, expiresAt, "Let's Encrypt", 5, nil)

		if r.alert != nil {
			t.Fatal("want no alert when alerts are disabled")
		}
	})

	t.Run("the per-incident alert cap suppresses further alerts", func(t *testing.T) {
		m := baseMonitor
		m.MaxAlertsPerIncident = 1
		m.AlertCount = 1 // already at the cap
		expiresAt := time.Now().Add(5 * 24 * time.Hour)

		r := buildSSLCheckResult(m, expiresAt, "Let's Encrypt", 5, nil)

		if r.alert != nil {
			t.Fatal("want no alert once the per-incident cap is reached")
		}
	})

	t.Run("alert_after_n_failures suppresses the alert until the threshold is met", func(t *testing.T) {
		m := baseMonitor
		m.AlertAfterNFailures = 2
		expiresAt := time.Now().Add(5 * 24 * time.Hour)

		r := buildSSLCheckResult(m, expiresAt, "Let's Encrypt", 5, nil)

		if r.alert != nil {
			t.Fatal("want no alert before consecutive failures exceed alert_after_n_failures")
		}
	})
}

// ─── checkSSLMonitors / checkOneSSLMonitor ────────────────────────────────

func TestCheckSSLMonitors(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	tg := telegram.NewClient("")
	mailer := email.NewSender("")
	wh := webhook.NewClient()
	logger := testLogger()
	n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, Logger: logger}

	t.Run("a connection failure is recorded as an error status", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		mon := testSSLMonitor(t, queries, org.ID, "this-host-does-not-exist.invalid")

		checkSSLMonitors(context.Background(), n)

		var status, errorMsg string
		var lastCheckedAt pgtype.Timestamptz
		if err := pool.QueryRow(context.Background(),
			"SELECT status, error_msg, last_checked_at FROM ssl_monitors WHERE id = $1", mon.ID,
		).Scan(&status, &errorMsg, &lastCheckedAt); err != nil {
			t.Fatalf("query monitor: %v", err)
		}
		if status != "error" {
			t.Fatalf("want status error, got %q", status)
		}
		if errorMsg == "" {
			t.Fatal("want a non-empty error message")
		}
		if !lastCheckedAt.Valid {
			t.Fatal("want last_checked_at set")
		}
	})

	t.Run("a monitor under an active maintenance window is excluded", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		mon := testSSLMonitor(t, queries, org.ID, "this-host-does-not-exist.invalid")
		var windowID uuid.UUID
		if err := pool.QueryRow(context.Background(),
			"INSERT INTO maintenance_windows (org_id, title, message, starts_at) VALUES ($1, 'Scheduled', '', NOW() - INTERVAL '1 minute') RETURNING id",
			org.ID,
		).Scan(&windowID); err != nil {
			t.Fatalf("seed maintenance window: %v", err)
		}
		mustExecWorker(t, pool, "INSERT INTO maintenance_window_monitors (window_id, monitor_type, monitor_id) VALUES ($1, 'ssl', $2)", windowID, mon.ID)

		checkSSLMonitors(context.Background(), n)

		var status string
		if err := pool.QueryRow(context.Background(), "SELECT status FROM ssl_monitors WHERE id = $1", mon.ID).Scan(&status); err != nil {
			t.Fatalf("query monitor: %v", err)
		}
		if status != "waiting" {
			t.Fatalf("want status unchanged (waiting) under maintenance, got %q", status)
		}
	})
}

// ─── checkDomainMonitors / checkOneDomainMonitor ──────────────────────────
//
// Unlike performTLSCheck, the RDAP lookup goes through an injectable
// *rdap.Client (rdap.NewClientWithHTTPClient), so both the success and
// error paths are testable against a local httptest server — no live
// network dependency.

func TestCheckDomainMonitors(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	tg := telegram.NewClient("")
	mailer := email.NewSender("")
	wh := webhook.NewClient()
	logger := testLogger()

	t.Run("a lookup failure is recorded as an error status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()
		n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, RDAP: rdap.NewClientWithHTTPClient(srv.Client(), srv.URL+"/domain/"), Logger: logger}

		org := testOrg(t, queries, pool)
		mon := testDomainMonitor(t, queries, org.ID, "this-domain-does-not-exist.invalid")

		checkDomainMonitors(context.Background(), n)

		var status, errorMsg string
		var lastCheckedAt pgtype.Timestamptz
		if err := pool.QueryRow(context.Background(),
			"SELECT status, error_msg, last_checked_at FROM domain_monitors WHERE id = $1", mon.ID,
		).Scan(&status, &errorMsg, &lastCheckedAt); err != nil {
			t.Fatalf("query monitor: %v", err)
		}
		if status != "error" {
			t.Fatalf("want status error, got %q", status)
		}
		if errorMsg == "" {
			t.Fatal("want a non-empty error message")
		}
		if !lastCheckedAt.Valid {
			t.Fatal("want last_checked_at set")
		}
	})

	t.Run("a successful lookup records registrar and expiry", func(t *testing.T) {
		expiresAt := time.Now().Add(90 * 24 * time.Hour).UTC().Truncate(time.Second)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintf(w, `{
				"events": [{"eventAction": "expiration", "eventDate": %q}],
				"entities": [{"roles": ["registrar"], "handle": "REG-1", "vcardArray": ["vcard", [["fn", {}, "text", "Test Registrar"]]]}]
			}`, expiresAt.Format(time.RFC3339))
		}))
		defer srv.Close()
		n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, RDAP: rdap.NewClientWithHTTPClient(srv.Client(), srv.URL+"/domain/"), Logger: logger}

		org := testOrg(t, queries, pool)
		mon := testDomainMonitor(t, queries, org.ID, "example.com")

		checkDomainMonitors(context.Background(), n)

		var status, registrar string
		var gotExpiresAt pgtype.Timestamptz
		if err := pool.QueryRow(context.Background(),
			"SELECT status, registrar, expires_at FROM domain_monitors WHERE id = $1", mon.ID,
		).Scan(&status, &registrar, &gotExpiresAt); err != nil {
			t.Fatalf("query monitor: %v", err)
		}
		if status != "up" {
			t.Fatalf("want status up, got %q", status)
		}
		if registrar != "Test Registrar" {
			t.Fatalf("want registrar Test Registrar, got %q", registrar)
		}
		if !gotExpiresAt.Valid || !gotExpiresAt.Time.Equal(expiresAt) {
			t.Fatalf("want expiresAt %v, got %v", expiresAt, gotExpiresAt.Time)
		}
	})

	t.Run("a monitor under an active maintenance window is excluded", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()
		n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, RDAP: rdap.NewClientWithHTTPClient(srv.Client(), srv.URL+"/domain/"), Logger: logger}

		org := testOrg(t, queries, pool)
		mon := testDomainMonitor(t, queries, org.ID, "this-domain-does-not-exist.invalid")
		var windowID uuid.UUID
		if err := pool.QueryRow(context.Background(),
			"INSERT INTO maintenance_windows (org_id, title, message, starts_at) VALUES ($1, 'Scheduled', '', NOW() - INTERVAL '1 minute') RETURNING id",
			org.ID,
		).Scan(&windowID); err != nil {
			t.Fatalf("seed maintenance window: %v", err)
		}
		mustExecWorker(t, pool, "INSERT INTO maintenance_window_monitors (window_id, monitor_type, monitor_id) VALUES ($1, 'domain', $2)", windowID, mon.ID)

		checkDomainMonitors(context.Background(), n)

		var status string
		if err := pool.QueryRow(context.Background(), "SELECT status FROM domain_monitors WHERE id = $1", mon.ID).Scan(&status); err != nil {
			t.Fatalf("query monitor: %v", err)
		}
		if status != "waiting" {
			t.Fatalf("want status unchanged (waiting) under maintenance, got %q", status)
		}
	})

	t.Run("crossing the 30-day threshold sends one alert via the fallback email", func(t *testing.T) {
		expiresAt := time.Now().Add(20 * 24 * time.Hour).UTC().Truncate(time.Second)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = fmt.Fprintf(w, `{"events": [{"eventAction": "expiration", "eventDate": %q}]}`, expiresAt.Format(time.RFC3339))
		}))
		defer srv.Close()
		n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, RDAP: rdap.NewClientWithHTTPClient(srv.Client(), srv.URL+"/domain/"), Logger: logger}

		org := testOrg(t, queries, pool)
		testUser(t, queries, org.ID) // fallback email recipient (ADR-023)
		mon := testDomainMonitor(t, queries, org.ID, "example.com")

		checkDomainMonitors(context.Background(), n)

		var status string
		var alerted30d, alerted14d, alerted7d bool
		if err := pool.QueryRow(context.Background(),
			"SELECT status, alerted_30d, alerted_14d, alerted_7d FROM domain_monitors WHERE id = $1", mon.ID,
		).Scan(&status, &alerted30d, &alerted14d, &alerted7d); err != nil {
			t.Fatalf("query monitor: %v", err)
		}
		if status != "expiring_soon" {
			t.Fatalf("want status expiring_soon, got %q", status)
		}
		if !alerted30d || alerted14d || alerted7d {
			t.Fatalf("want only alerted_30d set, got 30d=%v 14d=%v 7d=%v", alerted30d, alerted14d, alerted7d)
		}
	})
}

func TestDomainExpiredMessages(t *testing.T) {
	m := db.DomainMonitor{Name: "Prod domain", Domain: "example.com"}
	subject, telegramMsg, emailHTML := domainExpiredMessages(m, "2026-01-01")

	if !strings.Contains(subject, "Prod domain") || !strings.Contains(subject, "expired") {
		t.Fatalf("want subject to name the monitor and say expired, got %q", subject)
	}
	if !strings.Contains(telegramMsg, "example.com") || !strings.Contains(telegramMsg, "2026-01-01") {
		t.Fatalf("want telegram message to include the domain and date, got %q", telegramMsg)
	}
	if !strings.Contains(emailHTML, "example.com") {
		t.Fatalf("want email HTML to include the domain, got %q", emailHTML)
	}
}

func TestDomainExpiringSoonMessages(t *testing.T) {
	m := db.DomainMonitor{Name: "Prod domain", Domain: "example.com"}
	subject, telegramMsg, emailHTML := domainExpiringSoonMessages(m, 7, "2026-01-08")

	if !strings.Contains(subject, "Prod domain") || !strings.Contains(subject, "7 days") {
		t.Fatalf("want subject to name the monitor and say 7 days, got %q", subject)
	}
	if !strings.Contains(telegramMsg, "example.com") || !strings.Contains(telegramMsg, "7 days") {
		t.Fatalf("want telegram message to include the domain and day count, got %q", telegramMsg)
	}
	if !strings.Contains(emailHTML, "2026-01-08") {
		t.Fatalf("want email HTML to include the expiry date, got %q", emailHTML)
	}
}

// ─── pruneOldPings ─────────────────────────────────────────────────────────

func TestPruneOldPings(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	org := testOrg(t, queries, pool)
	mon := testCronMonitor(t, queries, org.ID, "every 1h")

	mustExecWorker(t, pool, "INSERT INTO cron_pings (monitor_id, received_at, source_ip) VALUES ($1, NOW() - INTERVAL '31 days', '1.1.1.1')", mon.ID)
	mustExecWorker(t, pool, "INSERT INTO cron_pings (monitor_id, received_at, source_ip) VALUES ($1, NOW() - INTERVAL '1 day', '1.1.1.2')", mon.ID)

	pruneOldPings(context.Background(), queries, testLogger())

	var remaining []string
	rows, err := pool.Query(context.Background(), "SELECT source_ip FROM cron_pings WHERE monitor_id = $1", mon.ID)
	if err != nil {
		t.Fatalf("query remaining pings: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			t.Fatalf("scan: %v", err)
		}
		remaining = append(remaining, ip)
	}
	if len(remaining) != 1 || remaining[0] != "1.1.1.2" {
		t.Fatalf("want only the recent ping (1.1.1.2) to survive, got %v", remaining)
	}
}

func mustExecWorker(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

// ─── checkPortMonitors / checkOnePortMonitor ──────────────────────────────

func TestCheckPortMonitors(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	tg := telegram.NewClient("")
	mailer := email.NewSender("")
	wh := webhook.NewClientWithHTTPClient(&http.Client{Timeout: 10 * time.Second})
	logger := testLogger()
	n := Notifiers{Queries: queries, Telegram: tg, Mailer: mailer, Webhook: wh, Logger: logger, TCPDialer: &net.Dialer{Timeout: 10 * time.Second}}

	t.Run("escalates to down after two consecutive failures, then recovers", func(t *testing.T) {
		host, port := freeTCPPort(t) // nothing listening yet — first checks fail

		org := testOrg(t, queries, pool)
		mon, err := queries.CreatePortMonitor(context.Background(), db.CreatePortMonitorParams{
			OrgID: org.ID, Name: "Port monitor", Host: host, Port: port,
			ExpectedState: db.PortExpectedStateOpen, IntervalMins: 10,
			MaxAlertsPerIncident: 3, AlertAfterNFailures: 1,
		})
		if err != nil {
			t.Fatalf("create port monitor: %v", err)
		}
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeEmail, map[string]string{"email": "a@b.com"})
		attachNotificationChannel(t, queries, channel.ID, "port", mon.ID)

		// First failure: recorded, but alert_after_n_failures=1 suppresses the alert.
		checkPortMonitors(context.Background(), n)
		if status, failures := getPortRow(t, pool, mon.ID); status == "down" || failures != 1 {
			t.Fatalf("after 1st failure: want not-down with 1 consecutive failure, got status=%q failures=%d", status, failures)
		}

		// Second consecutive failure: trips down, opens an incident, alerts.
		forceDueNow(t, pool, "port_monitors", mon.ID)
		checkPortMonitors(context.Background(), n)
		status, failures := getPortRow(t, pool, mon.ID)
		if status != "down" || failures != 2 {
			t.Fatalf("after 2nd failure: want down with 2 consecutive failures, got status=%q failures=%d", status, failures)
		}
		var alertCount int32
		var resolvedAt pgtype.Timestamptz
		if err := pool.QueryRow(context.Background(),
			"SELECT alert_count, resolved_at FROM port_incidents WHERE monitor_id = $1", mon.ID,
		).Scan(&alertCount, &resolvedAt); err != nil {
			t.Fatalf("query incident: %v", err)
		}
		if resolvedAt.Valid {
			t.Fatal("want an unresolved incident after going down")
		}
		if alertCount != 1 {
			t.Fatalf("want 1 alert delivered, got %d", alertCount)
		}

		// Recovery: start a listener on the exact same port, then recheck.
		_, _, closeFn := listenOnPort(t, host, port)
		defer closeFn()
		forceDueNow(t, pool, "port_monitors", mon.ID)
		checkPortMonitors(context.Background(), n)
		status, failures = getPortRow(t, pool, mon.ID)
		if status != "up" || failures != 0 {
			t.Fatalf("after recovery: want up with 0 consecutive failures, got status=%q failures=%d", status, failures)
		}
		if err := pool.QueryRow(context.Background(),
			"SELECT resolved_at FROM port_incidents WHERE monitor_id = $1", mon.ID,
		).Scan(&resolvedAt); err != nil {
			t.Fatalf("query incident: %v", err)
		}
		if !resolvedAt.Valid {
			t.Fatal("want the incident resolved after recovery")
		}
	})

	t.Run("delivers a recovery webhook event with a non-zero downtime duration", func(t *testing.T) {
		host, port := freeTCPPort(t)

		var gotEvents []webhook.Event
		hookSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var e webhook.Event
			_ = json.NewDecoder(r.Body).Decode(&e)
			gotEvents = append(gotEvents, e)
			w.WriteHeader(http.StatusOK)
		}))
		defer hookSrv.Close()

		org := testOrg(t, queries, pool)
		mon := testPortMonitor(t, queries, org.ID, host, port, db.PortExpectedStateOpen)
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeWebhook, map[string]string{"url": hookSrv.URL, "secret": "shh"})
		attachNotificationChannel(t, queries, channel.ID, "port", mon.ID)

		checkPortMonitors(context.Background(), n) // 1st failure trips down + webhook (alert_after_n_failures defaults to 0)

		_, _, closeFn := listenOnPort(t, host, port)
		defer closeFn()
		forceDueNow(t, pool, "port_monitors", mon.ID)
		checkPortMonitors(context.Background(), n) // recovery + webhook

		if len(gotEvents) != 2 {
			t.Fatalf("want 2 webhook events (down, recovery), got %d: %+v", len(gotEvents), gotEvents)
		}
		if gotEvents[0].EventType != "down" || gotEvents[0].Reason == "" {
			t.Fatalf("want a down event with a reason, got %+v", gotEvents[0])
		}
		if gotEvents[1].EventType != "recovery" || gotEvents[1].DowntimeDuration == "" {
			t.Fatalf("want a recovery event with a non-empty downtime duration, got %+v", gotEvents[1])
		}
	})

	t.Run("a closed-state monitor alerts when the port unexpectedly opens", func(t *testing.T) {
		host, port, closeFn := openTCPPort(t)
		defer closeFn()

		org := testOrg(t, queries, pool)
		mon := testPortMonitor(t, queries, org.ID, host, port, db.PortExpectedStateClosed)
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeEmail, map[string]string{"email": "a@b.com"})
		attachNotificationChannel(t, queries, channel.ID, "port", mon.ID)

		checkPortMonitors(context.Background(), n)

		status, _ := getPortRow(t, pool, mon.ID)
		if status != "down" {
			t.Fatalf("want down when a closed-state monitor's port is unexpectedly open, got %q", status)
		}
		var reason string
		if err := pool.QueryRow(context.Background(),
			"SELECT failure_reason FROM port_checks WHERE monitor_id = $1", mon.ID,
		).Scan(&reason); err != nil {
			t.Fatalf("query check: %v", err)
		}
		if reason != "port is unexpectedly open" {
			t.Fatalf("want the unexpectedly-open reason recorded, got %q", reason)
		}
	})

	t.Run("records a check row for every poll", func(t *testing.T) {
		host, port, closeFn := openTCPPort(t)
		defer closeFn()

		org := testOrg(t, queries, pool)
		mon := testPortMonitor(t, queries, org.ID, host, port, db.PortExpectedStateOpen)
		checkPortMonitors(context.Background(), n)

		var count int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM port_checks WHERE monitor_id = $1", mon.ID).Scan(&count); err != nil {
			t.Fatalf("count checks: %v", err)
		}
		if count != 1 {
			t.Fatalf("want 1 check recorded, got %d", count)
		}
	})

	t.Run("a monitor under an active maintenance window is excluded", func(t *testing.T) {
		host, port := freeTCPPort(t)

		org := testOrg(t, queries, pool)
		mon := testPortMonitor(t, queries, org.ID, host, port, db.PortExpectedStateOpen)
		var windowID uuid.UUID
		if err := pool.QueryRow(context.Background(),
			"INSERT INTO maintenance_windows (org_id, title, message, starts_at) VALUES ($1, 'Scheduled', '', NOW() - INTERVAL '1 minute') RETURNING id",
			org.ID,
		).Scan(&windowID); err != nil {
			t.Fatalf("seed maintenance window: %v", err)
		}
		mustExecWorker(t, pool, "INSERT INTO maintenance_window_monitors (window_id, monitor_type, monitor_id) VALUES ($1, 'port', $2)", windowID, mon.ID)

		checkPortMonitors(context.Background(), n)

		var count int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM port_checks WHERE monitor_id = $1", mon.ID).Scan(&count); err != nil {
			t.Fatalf("count checks: %v", err)
		}
		if count != 0 {
			t.Fatalf("want no check performed under maintenance, got %d", count)
		}
	})
}

func TestPerformTCPCheck(t *testing.T) {
	host, openPort, closeFn := openTCPPort(t)
	defer closeFn()
	_, closedPort := freeTCPPort(t)

	t.Run("expected open + reachable port is up", func(t *testing.T) {
		_, isUp, reason := performTCPCheck(db.PortMonitor{Host: host, Port: openPort, ExpectedState: db.PortExpectedStateOpen}, &net.Dialer{Timeout: 10 * time.Second})
		if !isUp || reason != "" {
			t.Fatalf("want (up, \"\"), got (%v, %q)", isUp, reason)
		}
	})

	t.Run("expected open + unreachable port is down", func(t *testing.T) {
		_, isUp, reason := performTCPCheck(db.PortMonitor{Host: host, Port: closedPort, ExpectedState: db.PortExpectedStateOpen}, &net.Dialer{Timeout: 10 * time.Second})
		if isUp || reason != "connection refused / timeout" {
			t.Fatalf("want (down, connection refused / timeout), got (%v, %q)", isUp, reason)
		}
	})

	t.Run("expected closed + reachable port is down (unexpectedly open)", func(t *testing.T) {
		_, isUp, reason := performTCPCheck(db.PortMonitor{Host: host, Port: openPort, ExpectedState: db.PortExpectedStateClosed}, &net.Dialer{Timeout: 10 * time.Second})
		if isUp || reason != "port is unexpectedly open" {
			t.Fatalf("want (down, port is unexpectedly open), got (%v, %q)", isUp, reason)
		}
	})

	t.Run("expected closed + unreachable port is up (matches expectation)", func(t *testing.T) {
		_, isUp, reason := performTCPCheck(db.PortMonitor{Host: host, Port: closedPort, ExpectedState: db.PortExpectedStateClosed}, &net.Dialer{Timeout: 10 * time.Second})
		if !isUp || reason != "" {
			t.Fatalf("want (up, \"\"), got (%v, %q)", isUp, reason)
		}
	})
}

func TestBuildPortDownAlert(t *testing.T) {
	t.Run("open state reads as a down service", func(t *testing.T) {
		alert := buildPortDownAlert(db.PortMonitor{Name: "DB", Host: "db.internal", Port: 5432, ExpectedState: db.PortExpectedStateOpen}, "connection refused / timeout")
		if !strings.Contains(alert.Telegram, "is down") {
			t.Fatalf("want 'is down' in open-state alert, got %q", alert.Telegram)
		}
	})

	t.Run("closed state reads as unexpectedly open, not down", func(t *testing.T) {
		alert := buildPortDownAlert(db.PortMonitor{Name: "DB", Host: "db.internal", Port: 5432, ExpectedState: db.PortExpectedStateClosed}, "port is unexpectedly open")
		if !strings.Contains(alert.Telegram, "unexpectedly open") {
			t.Fatalf("want 'unexpectedly open' in closed-state alert, got %q", alert.Telegram)
		}
		if strings.Contains(alert.Telegram, "is down") {
			t.Fatalf("want closed-state alert to avoid 'is down' phrasing, got %q", alert.Telegram)
		}
	})
}
