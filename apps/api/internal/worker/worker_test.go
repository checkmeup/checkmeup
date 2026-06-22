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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
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
	pool := testPool(t)
	queries := db.New(pool)
	tg := telegram.NewClient("")  // no token: SendMessage always errors, no network call
	mailer := email.NewSender("") // no API key: SendAlertEmail no-ops with a nil error (ADR-012)
	wh := webhook.NewClient()
	logger := testLogger()
	msg := AlertMessage{Telegram: "down", EmailSubject: "down", EmailHTML: "<p>down</p>"}

	t.Run("no channel attached and no org user falls back to nothing", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		sent := DispatchAlert(context.Background(), queries, tg, mailer, wh, org.ID, MonitorRef{Type: "cron", ID: uuid.New()}, msg, logger)
		if sent {
			t.Fatal("want no alert delivered with no channel attached and no org user")
		}
	})

	t.Run("no channel attached falls back to emailing every org user (ADR-023)", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		testUser(t, queries, org.ID)
		sent := DispatchAlert(context.Background(), queries, tg, mailer, wh, org.ID, MonitorRef{Type: "cron", ID: uuid.New()}, msg, logger)
		if !sent {
			t.Fatal("want the fallback email to org users to count as delivered")
		}
	})

	t.Run("telegram channel attached but unreachable in this environment does not count as delivered", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeTelegram, map[string]string{"chatId": "123"})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		sent := DispatchAlert(context.Background(), queries, tg, mailer, wh, org.ID, MonitorRef{Type: "cron", ID: monitorID}, msg, logger)
		if sent {
			t.Fatal("want telegram-only delivery to fail with no bot token configured, and no fallback since a channel is attached")
		}
	})

	t.Run("email channel succeeds (no-op-safe in dev) even when a telegram channel on the same monitor fails", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		monitorID := uuid.New()
		tgChannel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeTelegram, map[string]string{"chatId": "123"})
		emailChannel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeEmail, map[string]string{"email": "a@b.com"})
		attachNotificationChannel(t, queries, tgChannel.ID, "cron", monitorID)
		attachNotificationChannel(t, queries, emailChannel.ID, "cron", monitorID)

		sent := DispatchAlert(context.Background(), queries, tg, mailer, wh, org.ID, MonitorRef{Type: "cron", ID: monitorID}, msg, logger)
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
		sent := DispatchAlert(context.Background(), queries, tg, mailer, wh, org.ID, MonitorRef{Type: "cron", ID: monitorID}, msg, logger)
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
		sent := DispatchAlert(context.Background(), queries, tg, mailer, wh, org.ID, MonitorRef{Type: "cron", ID: monitorID}, webhookMsg, logger)
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
		sent := DispatchAlert(context.Background(), queries, tg, mailer, wh, org.ID, MonitorRef{Type: "cron", ID: monitorID}, webhookMsg, logger)
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

// ─── checkOverdue (cron) ───────────────────────────────────────────────────

func TestCheckOverdue(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	tg := telegram.NewClient("")
	mailer := email.NewSender("")
	wh := webhook.NewClient()
	logger := testLogger()

	t.Run("marks an overdue monitor down, opens an incident, and delivers an alert", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		mon := testCronMonitor(t, queries, org.ID, "every 1h")
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeEmail, map[string]string{"email": "a@b.com"})
		attachNotificationChannel(t, queries, channel.ID, "cron", mon.ID)
		mustExecWorker(t, pool, "UPDATE cron_monitors SET status = 'up', next_ping_at = NOW() - INTERVAL '1 hour' WHERE id = $1", mon.ID)

		checkOverdue(context.Background(), queries, tg, mailer, wh, logger)

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

		checkOverdue(context.Background(), queries, tg, mailer, wh, logger)

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

		checkOverdue(context.Background(), queries, tg, mailer, wh, logger)

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
		checkOverdue(context.Background(), queries, tg, mailer, wh, logger)
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
	wh := webhook.NewClient()
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
		mon := testUptimeMonitor(t, queries, org.ID, srv.URL)
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeEmail, map[string]string{"email": "a@b.com"})
		attachNotificationChannel(t, queries, channel.ID, "uptime", mon.ID)

		// First failure: recorded, but one failure alone doesn't trip "down".
		checkUptimeMonitors(context.Background(), queries, tg, mailer, wh, logger)
		if status, failures := getUptimeRow(t, pool, mon.ID); status == "down" || failures != 1 {
			t.Fatalf("after 1st failure: want not-down with 1 consecutive failure, got status=%q failures=%d", status, failures)
		}

		// Second consecutive failure: trips down, opens an incident, alerts.
		forceDueNow(t, pool, "uptime_monitors", mon.ID)
		checkUptimeMonitors(context.Background(), queries, tg, mailer, wh, logger)
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
		checkUptimeMonitors(context.Background(), queries, tg, mailer, wh, logger)
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

		checkUptimeMonitors(context.Background(), queries, tg, mailer, wh, logger)
		forceDueNow(t, pool, "uptime_monitors", mon.ID)
		checkUptimeMonitors(context.Background(), queries, tg, mailer, wh, logger) // 2nd failure trips down + webhook

		up.Store(true)
		forceDueNow(t, pool, "uptime_monitors", mon.ID)
		checkUptimeMonitors(context.Background(), queries, tg, mailer, wh, logger) // recovery + webhook

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
		checkUptimeMonitors(context.Background(), queries, tg, mailer, wh, logger)

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

		checkUptimeMonitors(context.Background(), queries, tg, mailer, wh, logger)

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
	wh := webhook.NewClient()
	logger := testLogger()

	t.Run("a connection failure is recorded as an error status", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		mon := testSSLMonitor(t, queries, org.ID, "this-host-does-not-exist.invalid")

		checkSSLMonitors(context.Background(), queries, tg, mailer, wh, logger)

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

		checkSSLMonitors(context.Background(), queries, tg, mailer, wh, logger)

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
