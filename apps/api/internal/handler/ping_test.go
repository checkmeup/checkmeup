package handler

// Tests for ping.go: the public, unauthenticated GET /ping/{token} endpoint
// cron jobs hit to check in, plus its pure helper functions. Integration
// tests follow the same conventions as the other *_test.go files in this
// package (real Postgres, ADR-010); the pure functions (computeNextPing,
// parseEveryDuration, realIP) are deterministic given their inputs (no
// internal time.Now() — "now" is always passed in), so they're tested with
// exact-value table tests, no tolerance windows needed. Downtime formatting
// (formerly a local formatDuration here) moved to worker.FormatDuration —
// see worker_test.go — once the cron-recovery path started sharing
// worker.DispatchAlert instead of sending alerts directly (EP-14).

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/twilio"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

func testPingHandler(t *testing.T) (*AuthHandler, *MonitorHandler, *PingHandler, *pgxpool.Pool) {
	t.Helper()
	pool := testPool(t)
	cfg := &config.Config{
		Env:           "development",
		JWTSecret:     testJWTSecret,
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 7 * 24 * time.Hour,
		AppURL:        "http://localhost:5173",
		BaseURL:       "http://localhost:8080",
	}
	authH := NewAuthHandler(cfg, pool)
	monitorH := NewMonitorHandler(cfg, pool, telegram.NewClient(""))
	// The recovery-webhook subtest in TestReceivePing POSTs to a local
	// httptest server, which webhook.NewClient's SSRF protections (loopback
	// blocked) would refuse to dial — not exercised here, see webhook_test.go.
	wh := webhook.NewClientWithHTTPClient(&http.Client{Timeout: 10 * time.Second})
	pingH := NewPingHandler(pool, telegram.NewClient(""), email.NewSender(""), wh, slack.NewClient(), twilio.NewClient("", "", "", ""))
	return authH, monitorH, pingH, pool
}

func doPing(t *testing.T, h *PingHandler, token string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/ping/"+token, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req = withURLParam(req, "token", token)
	w := httptest.NewRecorder()
	h.ReceivePing(w, req)
	return w
}

func TestReceivePing(t *testing.T) {
	authH, monitorH, pingH, pool := testPingHandler(t)

	t.Run("unknown token returns 404", func(t *testing.T) {
		w := doPing(t, pingH, "no-such-token", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("success records a ping and flips status to up", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Pinged job")
		token := cronPingToken(t, pool, mon.ID)

		w := doPing(t, pingH, token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		row := getCronMonitorRow(t, pool, mon.ID)
		if row.Status != db.MonitorStatusUp {
			t.Fatalf("want status up, got %q", row.Status)
		}
		if !row.LastPingAt.Valid {
			t.Fatal("want last_ping_at set")
		}
		if !row.NextPingAt.Valid {
			t.Fatal("want next_ping_at set")
		}

		var pingCount int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM cron_pings WHERE monitor_id = $1", mon.ID).Scan(&pingCount); err != nil {
			t.Fatalf("count pings: %v", err)
		}
		if pingCount != 1 {
			t.Fatalf("want 1 ping recorded, got %d", pingCount)
		}
	})

	t.Run("source IP resolution", func(t *testing.T) {
		cases := []struct {
			name    string
			headers map[string]string
			want    string
		}{
			{"falls back to RemoteAddr", nil, "192.0.2.1"},
			{"X-Real-IP wins", map[string]string{"X-Real-IP": "203.0.113.5"}, "203.0.113.5"},
			{"X-Forwarded-For takes the first hop", map[string]string{"X-Forwarded-For": "203.0.113.9, 10.0.0.1"}, "203.0.113.9"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				u := signUpTestUser(t, authH, pool)
				mon := createCronMonitor(t, monitorH, u.access, "IP test job")
				token := cronPingToken(t, pool, mon.ID)

				w := doPing(t, pingH, token, tc.headers)
				if w.Code != http.StatusOK {
					t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
				}

				var sourceIP string
				if err := pool.QueryRow(context.Background(),
					"SELECT source_ip FROM cron_pings WHERE monitor_id = $1", mon.ID,
				).Scan(&sourceIP); err != nil {
					t.Fatalf("query source_ip: %v", err)
				}
				if sourceIP != tc.want {
					t.Fatalf("want source IP %q, got %q", tc.want, sourceIP)
				}
			})
		}
	})

	t.Run("a paused monitor is not flipped to up or rescheduled", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Paused job")
		pauseW := doMonitorRequest(t, http.MethodPost, monitorH.PauseCronMonitor, u.access, mon.ID, nil)
		if pauseW.Code != http.StatusOK {
			t.Fatalf("setup: pause want 200, got %d", pauseW.Code)
		}

		token := cronPingToken(t, pool, mon.ID)
		w := doPing(t, pingH, token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		row := getCronMonitorRow(t, pool, mon.ID)
		if row.Status != db.MonitorStatusPaused {
			t.Fatalf("want status to remain paused, got %q", row.Status)
		}
		if row.LastPingAt.Valid {
			t.Fatal("want last_ping_at to remain unset while paused")
		}

		var pingCount int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM cron_pings WHERE monitor_id = $1", mon.ID).Scan(&pingCount); err != nil {
			t.Fatalf("count pings: %v", err)
		}
		if pingCount != 1 {
			t.Fatalf("want the ping itself still recorded even while paused, got %d", pingCount)
		}
	})

	t.Run("recovery resolves the open incident when alerts are enabled", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Recovering job")
		seedCronDown(t, pool, mon.ID, time.Now().Add(-30*time.Minute))
		token := cronPingToken(t, pool, mon.ID)

		w := doPing(t, pingH, token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		row := getCronMonitorRow(t, pool, mon.ID)
		if row.Status != db.MonitorStatusUp {
			t.Fatalf("want status up after recovery ping, got %q", row.Status)
		}

		var resolved bool
		if err := pool.QueryRow(context.Background(),
			"SELECT resolved_at IS NOT NULL FROM cron_incidents WHERE monitor_id = $1", mon.ID,
		).Scan(&resolved); err != nil {
			t.Fatalf("query incident: %v", err)
		}
		if !resolved {
			t.Fatal("want the open incident resolved on recovery")
		}
	})

	t.Run("recovery delivers a webhook event via the monitor's attached channel", func(t *testing.T) {
		var gotEvent webhook.Event
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&gotEvent)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Recovering job, webhook")
		seedCronDown(t, pool, mon.ID, time.Now().Add(-30*time.Minute))

		orgID, err := uuid.Parse(u.resp.OrgID)
		if err != nil {
			t.Fatalf("parse org id: %v", err)
		}
		monID, err := uuid.Parse(mon.ID)
		if err != nil {
			t.Fatalf("parse monitor id: %v", err)
		}

		queries := db.New(pool)
		channel, err := queries.CreateNotificationChannel(context.Background(), db.CreateNotificationChannelParams{
			OrgID: orgID, Type: db.NotificationChannelTypeWebhook, Name: "Hook",
			Config: []byte(`{"url":"` + srv.URL + `","secret":"shh"}`),
		})
		if err != nil {
			t.Fatalf("create webhook channel: %v", err)
		}
		if err := queries.InsertMonitorNotificationChannel(context.Background(), db.InsertMonitorNotificationChannelParams{
			ChannelID: channel.ID, MonitorType: "cron", MonitorID: monID,
		}); err != nil {
			t.Fatalf("attach webhook channel: %v", err)
		}

		token := cronPingToken(t, pool, mon.ID)
		w := doPing(t, pingH, token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		if gotEvent.EventType != "recovery" {
			t.Fatalf("want a recovery webhook event, got %+v", gotEvent)
		}
		if gotEvent.DowntimeDuration == "" {
			t.Fatal("want a non-empty downtime duration on the recovery event")
		}
	})

	t.Run("recovery resolves the incident even with alerts disabled", func(t *testing.T) {
		// Incident resolution must not depend on AlertsEnabled — only the
		// alert send itself is gated by that setting (matches the
		// uptime-monitor worker's pattern). Otherwise a monitor with alerts
		// off would accumulate permanently-open incidents on every down cycle.
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Recovering job, alerts off")
		seedCronDown(t, pool, mon.ID, time.Now().Add(-30*time.Minute))
		if _, err := pool.Exec(context.Background(), "UPDATE cron_monitors SET alerts_enabled = false WHERE id = $1", mon.ID); err != nil {
			t.Fatalf("disable alerts: %v", err)
		}
		token := cronPingToken(t, pool, mon.ID)

		w := doPing(t, pingH, token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		row := getCronMonitorRow(t, pool, mon.ID)
		if row.Status != db.MonitorStatusUp {
			t.Fatalf("want status up after recovery ping, got %q", row.Status)
		}

		var resolved bool
		if err := pool.QueryRow(context.Background(),
			"SELECT resolved_at IS NOT NULL FROM cron_incidents WHERE monitor_id = $1", mon.ID,
		).Scan(&resolved); err != nil {
			t.Fatalf("query incident: %v", err)
		}
		if !resolved {
			t.Fatal("want the open incident resolved on recovery even with alerts disabled")
		}
	})
}

func doPingStart(t *testing.T, h *PingHandler, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/ping/"+token+"/start", nil)
	req = withURLParam(req, "token", token)
	w := httptest.NewRecorder()
	h.ReceivePingStart(w, req)
	return w
}

// seedCronRun inserts a cron_runs row directly, simulating a run the worker
// has already flagged stuck (alerted=true) without needing a real 30s tick.
func seedCronRun(t *testing.T, pool *pgxpool.Pool, monitorID string, startedAt time.Time, alerted bool) {
	t.Helper()
	var alertedAt any
	if alerted {
		alertedAt = time.Now()
	}
	if _, err := pool.Exec(context.Background(),
		"INSERT INTO cron_runs (monitor_id, started_at, alerted_at) VALUES ($1, $2, $3)",
		monitorID, startedAt, alertedAt,
	); err != nil {
		t.Fatalf("seed cron run: %v", err)
	}
}

// attachWebhookChannel creates a webhook notification channel pointed at url
// and attaches it to the given cron monitor, so a test can assert on the
// dispatched alert's content.
func attachWebhookChannel(t *testing.T, pool *pgxpool.Pool, u signedUpUser, monitorID, url string) {
	t.Helper()
	orgID, err := uuid.Parse(u.resp.OrgID)
	if err != nil {
		t.Fatalf("parse org id: %v", err)
	}
	monID, err := uuid.Parse(monitorID)
	if err != nil {
		t.Fatalf("parse monitor id: %v", err)
	}
	queries := db.New(pool)
	channel, err := queries.CreateNotificationChannel(context.Background(), db.CreateNotificationChannelParams{
		OrgID: orgID, Type: db.NotificationChannelTypeWebhook, Name: "Hook",
		Config: []byte(`{"url":"` + url + `","secret":"shh"}`),
	})
	if err != nil {
		t.Fatalf("create webhook channel: %v", err)
	}
	if err := queries.InsertMonitorNotificationChannel(context.Background(), db.InsertMonitorNotificationChannelParams{
		ChannelID: channel.ID, MonitorType: "cron", MonitorID: monID,
	}); err != nil {
		t.Fatalf("attach webhook channel: %v", err)
	}
}

func TestReceivePingStart(t *testing.T) {
	authH, monitorH, pingH, pool := testPingHandler(t)

	t.Run("unknown token returns 404", func(t *testing.T) {
		w := doPingStart(t, pingH, "no-such-token")
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})

	t.Run("records a run without touching monitor ping state (ADR-039)", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Start ping job")
		token := cronPingToken(t, pool, mon.ID)

		w := doPingStart(t, pingH, token)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		var runCount int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM cron_runs WHERE monitor_id = $1", mon.ID).Scan(&runCount); err != nil {
			t.Fatalf("count runs: %v", err)
		}
		if runCount != 1 {
			t.Fatalf("want 1 run recorded, got %d", runCount)
		}

		row := getCronMonitorRow(t, pool, mon.ID)
		if row.Status != db.MonitorStatusWaiting {
			t.Fatalf("want status to remain untouched (waiting), got %q", row.Status)
		}
		if row.LastPingAt.Valid {
			t.Fatal("want last_ping_at untouched by the start ping")
		}
		if row.NextPingAt.Valid {
			t.Fatal("want next_ping_at untouched by the start ping")
		}
	})

	t.Run("completion ping closes the open run and stamps run_started_at (US-3401/US-3404)", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Round trip job")
		token := cronPingToken(t, pool, mon.ID)

		if w := doPingStart(t, pingH, token); w.Code != http.StatusOK {
			t.Fatalf("start: want 200, got %d", w.Code)
		}
		if w := doPing(t, pingH, token, nil); w.Code != http.StatusOK {
			t.Fatalf("complete: want 200, got %d", w.Code)
		}

		var completedCount int
		if err := pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM cron_runs WHERE monitor_id = $1 AND completed_at IS NOT NULL", mon.ID,
		).Scan(&completedCount); err != nil {
			t.Fatalf("count completed runs: %v", err)
		}
		if completedCount != 1 {
			t.Fatalf("want 1 completed run, got %d", completedCount)
		}

		var hasRunStartedAt bool
		if err := pool.QueryRow(context.Background(),
			"SELECT run_started_at IS NOT NULL FROM cron_pings WHERE monitor_id = $1", mon.ID,
		).Scan(&hasRunStartedAt); err != nil {
			t.Fatalf("query ping run_started_at: %v", err)
		}
		if !hasRunStartedAt {
			t.Fatal("want the completion ping to record run_started_at")
		}
	})

	t.Run("completion ping with no open run is accepted as-is (US-3401)", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "No start ping job")
		token := cronPingToken(t, pool, mon.ID)

		w := doPing(t, pingH, token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		var runStartedAtIsNull bool
		if err := pool.QueryRow(context.Background(),
			"SELECT run_started_at IS NULL FROM cron_pings WHERE monitor_id = $1", mon.ID,
		).Scan(&runStartedAtIsNull); err != nil {
			t.Fatalf("query ping run_started_at: %v", err)
		}
		if !runStartedAtIsNull {
			t.Fatal("want run_started_at to stay null with no preceding start ping")
		}
	})

	t.Run("recovery alert fires on completion when the closed run had already been flagged stuck (US-3403)", func(t *testing.T) {
		var gotEvent webhook.Event
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&gotEvent)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Stuck then completes")
		attachWebhookChannel(t, pool, u, mon.ID, srv.URL)

		token := cronPingToken(t, pool, mon.ID)
		seedCronRun(t, pool, mon.ID, time.Now().Add(-2*time.Hour), true)

		w := doPing(t, pingH, token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		if gotEvent.EventType != "recovery" {
			t.Fatalf("want a recovery webhook event, got %+v", gotEvent)
		}
	})

	t.Run("recovery alert fires when a new start ping supersedes an already-alerted open run (US-3403)", func(t *testing.T) {
		var gotEvent webhook.Event
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&gotEvent)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Superseded run")
		attachWebhookChannel(t, pool, u, mon.ID, srv.URL)

		token := cronPingToken(t, pool, mon.ID)
		seedCronRun(t, pool, mon.ID, time.Now().Add(-3*time.Hour), true)

		w := doPingStart(t, pingH, token)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		if gotEvent.EventType != "recovery" {
			t.Fatalf("want a recovery webhook event on supersede, got %+v", gotEvent)
		}

		var runCount int
		if err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM cron_runs WHERE monitor_id = $1", mon.ID).Scan(&runCount); err != nil {
			t.Fatalf("count runs: %v", err)
		}
		if runCount != 2 {
			t.Fatalf("want both the superseded run and the new run present, got %d", runCount)
		}
	})

	t.Run("no recovery alert when the open run was never flagged stuck", func(t *testing.T) {
		var gotEvent webhook.Event
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&gotEvent)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		u := signUpTestUser(t, authH, pool)
		mon := createCronMonitor(t, monitorH, u.access, "Never flagged")
		attachWebhookChannel(t, pool, u, mon.ID, srv.URL)

		token := cronPingToken(t, pool, mon.ID)
		if w := doPingStart(t, pingH, token); w.Code != http.StatusOK {
			t.Fatalf("start: want 200, got %d", w.Code)
		}

		w := doPing(t, pingH, token, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("complete: want 200, got %d", w.Code)
		}

		if gotEvent.EventType != "" {
			t.Fatalf("want no alert for a run that was never flagged stuck, got %+v", gotEvent)
		}
	})
}

func TestComputeNextPing(t *testing.T) {
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)

	t.Run("interval schedule adds the interval plus grace", func(t *testing.T) {
		got := computeNextPing("every 1h", now, 5)
		want := now.Add(time.Hour + 5*time.Minute)
		if !got.Equal(want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})

	t.Run("cron expression uses the next scheduled fire time plus grace", func(t *testing.T) {
		// "0 13 * * *" = 13:00 daily; from 12:00 the next fire is 13:00 the same day.
		got := computeNextPing("0 13 * * *", now, 10)
		want := time.Date(2026, 6, 19, 13, 10, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})

	t.Run("unparseable schedule falls back to 1 hour plus grace", func(t *testing.T) {
		got := computeNextPing("garbage schedule", now, 5)
		want := now.Add(time.Hour + 5*time.Minute)
		if !got.Equal(want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}

func TestParseEveryDuration(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"1h", time.Hour},
		{"30m", 30 * time.Minute},
		{"1d", 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"2hours", 2 * time.Hour},
		{"45mins", 45 * time.Minute},
		{"3hr", 3 * time.Hour},
		{"", 0},
		{"abc", 0},
		{"0h", 0},
		{"-5m", 0},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := parseEveryDuration(tc.in); got != tc.want {
				t.Fatalf("parseEveryDuration(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestRealIP(t *testing.T) {
	t.Run("X-Real-IP takes priority", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("X-Real-IP", "203.0.113.5")
		r.Header.Set("X-Forwarded-For", "10.0.0.1")
		if got := realIP(r); got != "203.0.113.5" {
			t.Fatalf("want 203.0.113.5, got %q", got)
		}
	})

	t.Run("X-Forwarded-For takes the first hop", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.1, 10.0.0.2")
		if got := realIP(r); got != "203.0.113.9" {
			t.Fatalf("want 203.0.113.9, got %q", got)
		}
	})

	t.Run("falls back to RemoteAddr, stripping the port", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "198.51.100.7:54321"
		if got := realIP(r); got != "198.51.100.7" {
			t.Fatalf("want 198.51.100.7, got %q", got)
		}
	})
}

func cronPingToken(t *testing.T, pool *pgxpool.Pool, monitorID string) string {
	t.Helper()
	var token string
	if err := pool.QueryRow(context.Background(), "SELECT ping_token FROM cron_monitors WHERE id = $1", monitorID).Scan(&token); err != nil {
		t.Fatalf("lookup ping token: %v", err)
	}
	return token
}

func getCronMonitorRow(t *testing.T, pool *pgxpool.Pool, monitorID string) db.CronMonitor {
	t.Helper()
	var m db.CronMonitor
	if err := pool.QueryRow(context.Background(),
		"SELECT id, org_id, name, schedule, grace_period_mins, ping_token, status, alerts_enabled, max_alerts_per_incident, last_ping_at, next_ping_at, created_at, updated_at FROM cron_monitors WHERE id = $1",
		monitorID,
	).Scan(
		&m.ID, &m.OrgID, &m.Name, &m.Schedule, &m.GracePeriodMins, &m.PingToken, &m.Status, &m.AlertsEnabled,
		&m.MaxAlertsPerIncident, &m.LastPingAt, &m.NextPingAt, &m.CreatedAt, &m.UpdatedAt,
	); err != nil {
		t.Fatalf("lookup monitor: %v", err)
	}
	return m
}

func seedCronDown(t *testing.T, pool *pgxpool.Pool, monitorID string, startedAt time.Time) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), "UPDATE cron_monitors SET status = 'down' WHERE id = $1", monitorID); err != nil {
		t.Fatalf("seed monitor down: %v", err)
	}
	if _, err := pool.Exec(context.Background(),
		"INSERT INTO cron_incidents (monitor_id, started_at) VALUES ($1, $2)", monitorID, startedAt,
	); err != nil {
		t.Fatalf("seed open incident: %v", err)
	}
}
