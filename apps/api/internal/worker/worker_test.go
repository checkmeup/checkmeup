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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/telegram"
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

// testOrg creates an org with an alert email already set and email alerts
// enabled, so dispatchAlert has a real (no-op-safe) channel to succeed on —
// see the package doc comment above for why that's the only deliverable
// channel available without live credentials.
func testOrg(t *testing.T, queries *db.Queries, pool *pgxpool.Pool) db.Org {
	t.Helper()
	org, err := queries.CreateOrg(context.Background(), db.CreateOrgParams{
		Name:       "test-" + uuid.NewString(),
		AlertEmail: pgtype.Text{String: "ops-" + uuid.NewString() + "@example.com", Valid: true},
	})
	if err != nil {
		t.Fatalf("create test org: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM orgs WHERE id = $1", org.ID)
	})
	return org
}

func enableEmailAlerts(t *testing.T, pool *pgxpool.Pool, orgID uuid.UUID) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), "UPDATE orgs SET email_alerts_enabled = true WHERE id = $1", orgID); err != nil {
		t.Fatalf("enable email alerts: %v", err)
	}
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
		KeywordMode: db.KeywordModeContains,
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

func TestHasAlertChannel(t *testing.T) {
	cases := []struct {
		name string
		org  db.Org
		want bool
	}{
		{"telegram configured", db.Org{TelegramChatID: pgtype.Text{String: "123", Valid: true}}, true},
		{"email enabled and address set", db.Org{EmailAlertsEnabled: true, AlertEmail: pgtype.Text{String: "a@b.com", Valid: true}}, true},
		{"email enabled but no address", db.Org{EmailAlertsEnabled: true}, false},
		{"email address set but not enabled", db.Org{AlertEmail: pgtype.Text{String: "a@b.com", Valid: true}}, false},
		{"nothing configured", db.Org{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasAlertChannel(tc.org); got != tc.want {
				t.Fatalf("hasAlertChannel(%+v) = %v, want %v", tc.org, got, tc.want)
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

		code, _, isUp, reason := performHTTPCheck(db.UptimeMonitor{Url: srv.URL})
		if code != 200 || !isUp || reason != "" {
			t.Fatalf("want (200, up, \"\"), got (%d, %v, %q)", code, isUp, reason)
		}
	})

	t.Run("non-200 is down with the status code in the reason", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		code, _, isUp, reason := performHTTPCheck(db.UptimeMonitor{Url: srv.URL})
		if code != 500 || isUp || reason != "HTTP 500" {
			t.Fatalf("want (500, down, HTTP 500), got (%d, %v, %q)", code, isUp, reason)
		}
	})

	t.Run("connection error is down with no status code", func(t *testing.T) {
		code, _, isUp, reason := performHTTPCheck(db.UptimeMonitor{Url: "http://127.0.0.1:1"})
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
		})
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
		})
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
		})
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
		})
		if isUp || reason != "HTTP 503" {
			t.Fatalf("want the status code to win over the keyword, got isUp=%v reason=%q", isUp, reason)
		}
	})
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
	tg := telegram.NewClient("")    // no token: SendMessage always errors, no network call
	mailer := email.NewSender("") // no API key: SendAlertEmail no-ops with a nil error (ADR-012)
	logger := testLogger()
	msg := alertMessage{telegram: "down", emailSubject: "down", emailHTML: "<p>down</p>"}

	t.Run("no channel configured", func(t *testing.T) {
		sent := dispatchAlert(tg, mailer, db.Org{}, msg, logger, uuid.New())
		if sent {
			t.Fatal("want no alert delivered with no channel configured")
		}
	})

	t.Run("telegram configured but unreachable in this environment does not count as delivered", func(t *testing.T) {
		org := db.Org{TelegramChatID: pgtype.Text{String: "123", Valid: true}}
		sent := dispatchAlert(tg, mailer, org, msg, logger, uuid.New())
		if sent {
			t.Fatal("want telegram-only delivery to fail with no bot token configured")
		}
	})

	t.Run("email channel succeeds (no-op-safe in dev) even when telegram fails", func(t *testing.T) {
		org := db.Org{
			TelegramChatID:     pgtype.Text{String: "123", Valid: true},
			EmailAlertsEnabled: true,
			AlertEmail:         pgtype.Text{String: "a@b.com", Valid: true},
		}
		sent := dispatchAlert(tg, mailer, org, msg, logger, uuid.New())
		if !sent {
			t.Fatal("want the email channel's success to count as delivered, even though telegram failed")
		}
	})
}

// ─── checkOverdue (cron) ───────────────────────────────────────────────────

func TestCheckOverdue(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	tg := telegram.NewClient("")
	mailer := email.NewSender("")
	logger := testLogger()

	t.Run("marks an overdue monitor down, opens an incident, and delivers an alert", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		enableEmailAlerts(t, pool, org.ID)
		mon := testCronMonitor(t, queries, org.ID, "every 1h")
		mustExecWorker(t, pool, "UPDATE cron_monitors SET status = 'up', next_ping_at = NOW() - INTERVAL '1 hour' WHERE id = $1", mon.ID)

		checkOverdue(context.Background(), queries, tg, mailer, logger)

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
		enableEmailAlerts(t, pool, org.ID)
		mon := testCronMonitor(t, queries, org.ID, "every 1h")
		mustExecWorker(t, pool, "UPDATE cron_monitors SET status = 'up', next_ping_at = NOW() - INTERVAL '1 hour', alerts_enabled = false WHERE id = $1", mon.ID)

		checkOverdue(context.Background(), queries, tg, mailer, logger)

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

		checkOverdue(context.Background(), queries, tg, mailer, logger)

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
		checkOverdue(context.Background(), queries, tg, mailer, logger)
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
	logger := testLogger()

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
		enableEmailAlerts(t, pool, org.ID)
		mon := testUptimeMonitor(t, queries, org.ID, srv.URL)

		// First failure: recorded, but one failure alone doesn't trip "down".
		checkUptimeMonitors(context.Background(), queries, tg, mailer, logger)
		if status, failures := getUptimeRow(t, pool, mon.ID); status == "down" || failures != 1 {
			t.Fatalf("after 1st failure: want not-down with 1 consecutive failure, got status=%q failures=%d", status, failures)
		}

		// Second consecutive failure: trips down, opens an incident, alerts.
		forceDueNow(t, pool, "uptime_monitors", mon.ID)
		checkUptimeMonitors(context.Background(), queries, tg, mailer, logger)
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
		checkUptimeMonitors(context.Background(), queries, tg, mailer, logger)
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

	t.Run("records a check row for every poll", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		org := testOrg(t, queries, pool)
		mon := testUptimeMonitor(t, queries, org.ID, srv.URL)
		checkUptimeMonitors(context.Background(), queries, tg, mailer, logger)

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

		checkUptimeMonitors(context.Background(), queries, tg, mailer, logger)

		var count int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM uptime_checks WHERE monitor_id = $1", mon.ID).Scan(&count); err != nil {
			t.Fatalf("count checks: %v", err)
		}
		if count != 0 {
			t.Fatalf("want no check performed under maintenance, got %d", count)
		}
	})
}

// ─── checkSSLMonitors / checkOneSSLMonitor ────────────────────────────────

func TestCheckSSLMonitors(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	tg := telegram.NewClient("")
	mailer := email.NewSender("")
	logger := testLogger()

	t.Run("a connection failure is recorded as an error status", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		mon := testSSLMonitor(t, queries, org.ID, "this-host-does-not-exist.invalid")

		checkSSLMonitors(context.Background(), queries, tg, mailer, logger)

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

		checkSSLMonitors(context.Background(), queries, tg, mailer, logger)

		var status string
		if err := pool.QueryRow(context.Background(), "SELECT status FROM ssl_monitors WHERE id = $1", mon.ID).Scan(&status); err != nil {
			t.Fatalf("query monitor: %v", err)
		}
		if status != "waiting" {
			t.Fatalf("want status unchanged (waiting) under maintenance, got %q", status)
		}
	})
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
