package worker

// Tests for PRD alert-delivery: a down alert that didn't go out is sent again
// every 5 minutes while its incident is open and its monitor down.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/rdap"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

// statusWebhook attaches a webhook channel to the monitor whose receiver
// answers with the status code set (200 at first), and counts what it got.
func statusWebhook(t *testing.T, queries *db.Queries, orgID uuid.UUID, monitorType string, monitorID uuid.UUID) (setStatus func(int), received func() int) {
	t.Helper()
	var code atomic.Int32
	code.Store(http.StatusOK)
	var mu sync.Mutex
	count := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		count++
		mu.Unlock()
		w.WriteHeader(int(code.Load()))
	}))
	t.Cleanup(srv.Close)
	channel := testNotificationChannel(t, queries, orgID, db.NotificationChannelTypeWebhook, map[string]string{"url": srv.URL, "secret": "shh"})
	attachNotificationChannel(t, queries, channel.ID, monitorType, monitorID)
	return func(c int) { code.Store(int32(c)) }, func() int {
		mu.Lock()
		defer mu.Unlock()
		return count
	}
}

func retryNotifiers(queries *db.Queries) Notifiers {
	// The webhook receivers are local servers, which webhook.NewClient refuses to dial.
	return Notifiers{
		Queries: queries, Telegram: telegram.NewClient(""), Mailer: email.NewSender(""), Logger: testLogger(),
		Webhook:    webhook.NewClientWithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func incidentAlertCount(t *testing.T, pool *pgxpool.Pool, table string, monitorID uuid.UUID) int32 {
	t.Helper()
	var n int32
	if err := pool.QueryRow(context.Background(), "SELECT alert_count FROM "+table+" WHERE monitor_id = $1", monitorID).Scan(&n); err != nil {
		t.Fatalf("query %s: %v", table, err)
	}
	return n
}

// seedUndeliveredIncident opens an incident on a down monitor whose down
// alert was last attempted 6 minutes ago and never delivered.
func seedUndeliveredIncident(t *testing.T, pool *pgxpool.Pool, monitorType string, monitorID uuid.UUID) {
	t.Helper()
	alert, err := json.Marshal(buildAlert(alertText{Event: "down", MonitorType: monitorType, Name: "Seeded", Icon: "🔴", Headline: " is down", Subject: "DOWN: Seeded", SMS: "Seeded is down"}))
	if err != nil {
		t.Fatalf("marshal alert: %v", err)
	}
	mustExecWorker(t, pool, "UPDATE "+monitorType+"_monitors SET status = 'down' WHERE id = $1", monitorID)
	mustExecWorker(t, pool, "INSERT INTO "+monitorType+"_incidents (monitor_id, alert, alert_attempted_at, alert_attempts) VALUES ($1, $2, NOW() - INTERVAL '6 minutes', 1)", monitorID, alert)
}

func TestRetryDownAlerts(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	n := retryNotifiers(queries)

	t.Run("a down alert that didn't go out is sent again 5 minutes later, until delivered, then not again", func(t *testing.T) {
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer target.Close()
		org := testOrg(t, queries, pool) // no users: the Fallback email has nobody to reach
		mon := testUptimeMonitor(t, queries, org.ID, target.URL)
		setStatus, received := statusWebhook(t, queries, org.ID, "uptime", mon.ID)
		setStatus(http.StatusInternalServerError)

		checkUptimeMonitors(context.Background(), n) // rechecked before it counts
		forceDueNow(t, pool, "uptime_monitors", mon.ID)
		checkUptimeMonitors(context.Background(), n)
		WaitForAlerts()
		if got := received(); got != 1 {
			t.Fatalf("want the first down alert attempted once, got %d", got)
		}

		retryAlerts(context.Background(), n)
		WaitForAlerts()
		if got := received(); got != 1 {
			t.Fatalf("want no retry within 5 minutes of the attempt, got %d attempts", got)
		}

		setStatus(http.StatusOK)
		mustExecWorker(t, pool, "UPDATE uptime_incidents SET alert_attempted_at = NOW() - INTERVAL '5 minutes' WHERE monitor_id = $1", mon.ID)
		retryAlerts(context.Background(), n)
		WaitForAlerts()
		if got := received(); got != 2 {
			t.Fatalf("want a retry 5 minutes after the attempt, got %d attempts", got)
		}
		if got := incidentAlertCount(t, pool, "uptime_incidents", mon.ID); got != 1 {
			t.Fatalf("want the delivered alert counted once, got alert_count %d", got)
		}

		mustExecWorker(t, pool, "UPDATE uptime_incidents SET alert_attempted_at = NOW() - INTERVAL '1 hour' WHERE monitor_id = $1", mon.ID)
		retryAlerts(context.Background(), n)
		WaitForAlerts()
		if got := received(); got != 2 {
			t.Fatalf("want no retry once delivered, got %d attempts", got)
		}
	})

	t.Run("every incident-tracked type retries its undelivered down alert", func(t *testing.T) {
		for _, tc := range []struct {
			monitorType string
			create      func(orgID uuid.UUID) uuid.UUID
		}{
			{"uptime", func(orgID uuid.UUID) uuid.UUID {
				return testUptimeMonitor(t, queries, orgID, "http://example.invalid").ID
			}},
			{"port", func(orgID uuid.UUID) uuid.UUID {
				return testPortMonitor(t, queries, orgID, "example.invalid", 443, db.PortExpectedStateOpen).ID
			}},
			{"dns", func(orgID uuid.UUID) uuid.UUID { return testDNSMonitor(t, queries, orgID, "example.invalid", "").ID }},
			{"cron", func(orgID uuid.UUID) uuid.UUID { return testCronMonitor(t, queries, orgID, "every 1h").ID }},
		} {
			t.Run(tc.monitorType, func(t *testing.T) {
				org := testOrg(t, queries, pool)
				id := tc.create(org.ID)
				_, received := statusWebhook(t, queries, org.ID, tc.monitorType, id)
				seedUndeliveredIncident(t, pool, tc.monitorType, id)

				retryAlerts(context.Background(), n)
				WaitForAlerts()
				if got := received(); got != 1 {
					t.Fatalf("want the down alert sent again, got %d", got)
				}
				if got := incidentAlertCount(t, pool, tc.monitorType+"_incidents", id); got != 1 {
					t.Fatalf("want it counted once delivered, got alert_count %d", got)
				}
			})
		}
	})

	t.Run("no retry once the problem it reports is gone or muted", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			after func(orgID, monitorID uuid.UUID)
		}{
			{"incident resolved", func(_, id uuid.UUID) {
				mustExecWorker(t, pool, "UPDATE uptime_incidents SET resolved_at = NOW() WHERE monitor_id = $1", id)
			}},
			{"monitor up again", func(_, id uuid.UUID) {
				mustExecWorker(t, pool, "UPDATE uptime_monitors SET status = 'up' WHERE id = $1", id)
			}},
			{"monitor paused", func(_, id uuid.UUID) {
				mustExecWorker(t, pool, "UPDATE uptime_monitors SET status = 'paused' WHERE id = $1", id)
			}},
			{"alerts turned off", func(_, id uuid.UUID) {
				mustExecWorker(t, pool, "UPDATE uptime_monitors SET alerts_enabled = false WHERE id = $1", id)
			}},
			{"under a maintenance window", func(orgID, id uuid.UUID) {
				var windowID uuid.UUID
				if err := pool.QueryRow(context.Background(),
					"INSERT INTO maintenance_windows (org_id, title, message, starts_at) VALUES ($1, 'Scheduled', '', NOW() - INTERVAL '1 minute') RETURNING id", orgID,
				).Scan(&windowID); err != nil {
					t.Fatalf("create maintenance window: %v", err)
				}
				mustExecWorker(t, pool, "INSERT INTO maintenance_window_monitors (window_id, monitor_type, monitor_id) VALUES ($1, 'uptime', $2)", windowID, id)
			}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				org := testOrg(t, queries, pool)
				mon := testUptimeMonitor(t, queries, org.ID, "http://example.invalid")
				_, received := statusWebhook(t, queries, org.ID, "uptime", mon.ID)
				seedUndeliveredIncident(t, pool, "uptime", mon.ID)
				tc.after(org.ID, mon.ID)

				retryAlerts(context.Background(), n)
				WaitForAlerts()
				if got := received(); got != 0 {
					t.Fatalf("want no retry, got %d", got)
				}
			})
		}
	})

	t.Run("a down alert opened with alerts off goes out once they are turned on", func(t *testing.T) {
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer target.Close()
		org := testOrg(t, queries, pool)
		mon := testUptimeMonitor(t, queries, org.ID, target.URL)
		_, received := statusWebhook(t, queries, org.ID, "uptime", mon.ID)
		disableAlerts(t, pool, "uptime_monitors", mon.ID)

		checkUptimeMonitors(context.Background(), n)
		forceDueNow(t, pool, "uptime_monitors", mon.ID)
		checkUptimeMonitors(context.Background(), n)
		WaitForAlerts()

		mustExecWorker(t, pool, "UPDATE uptime_monitors SET alerts_enabled = true WHERE id = $1", mon.ID)
		retryAlerts(context.Background(), n)
		WaitForAlerts()
		if got := received(); got != 1 {
			t.Fatalf("want the down alert sent once alerts are on, got %d", got)
		}
	})
}

func TestRetryMessage_SMSTriedAtMostThreeTimes(t *testing.T) {
	alert, err := json.Marshal(AlertMessage{EmailSubject: "DOWN: api", SMS: "Checkmeup: api is down"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, tc := range []struct {
		attempt int32
		wantSMS string
	}{
		{2, "Checkmeup: api is down"},
		{3, "Checkmeup: api is down"},
		{4, ""},
	} {
		msg, err := retryMessage(alert, tc.attempt)
		if err != nil {
			t.Fatalf("attempt %d: %v", tc.attempt, err)
		}
		if msg.SMS != tc.wantSMS || msg.EmailSubject != "DOWN: api" {
			t.Fatalf("attempt %d: got SMS %q subject %q, want SMS %q", tc.attempt, msg.SMS, msg.EmailSubject, tc.wantSMS)
		}
	}
}

// nextCheckIn is how long until the monitor's next check, in minutes.
func nextCheckIn(t *testing.T, pool *pgxpool.Pool, table string, id uuid.UUID) float64 {
	t.Helper()
	var mins float64
	if err := pool.QueryRow(context.Background(),
		"SELECT EXTRACT(EPOCH FROM next_check_at - NOW()) / 60 FROM "+table+" WHERE id = $1", id,
	).Scan(&mins); err != nil {
		t.Fatalf("query next check: %v", err)
	}
	return mins
}

func TestRetryExpiryAndSSLAlerts(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)

	t.Run("an SSL error alert that fails is sent again 5 minutes later, until delivered, and counts once", func(t *testing.T) {
		n := retryNotifiers(queries)
		org := testOrg(t, queries, pool)
		mon := testSSLMonitor(t, queries, org.ID, "this-host-does-not-exist.invalid")
		setStatus, received := statusWebhook(t, queries, org.ID, "ssl", mon.ID)
		setStatus(http.StatusInternalServerError)
		// Already in error, so this failure isn't held for a recheck.
		mustExecWorker(t, pool, "UPDATE ssl_monitors SET status = 'error', consecutive_failures = 1 WHERE id = $1", mon.ID)

		checkSSLMonitors(context.Background(), n)
		WaitForAlerts()
		retryAlerts(context.Background(), n)
		WaitForAlerts()
		if got := received(); got != 1 {
			t.Fatalf("want one attempt and no retry within 5 minutes, got %d", got)
		}

		setStatus(http.StatusOK)
		mustExecWorker(t, pool, "UPDATE ssl_monitors SET alert_attempted_at = NOW() - INTERVAL '5 minutes' WHERE id = $1", mon.ID)
		retryAlerts(context.Background(), n)
		WaitForAlerts()
		if got := received(); got != 2 {
			t.Fatalf("want the alert sent again, got %d attempts", got)
		}

		mustExecWorker(t, pool, "UPDATE ssl_monitors SET alert_attempted_at = NOW() - INTERVAL '1 hour' WHERE id = $1", mon.ID)
		retryAlerts(context.Background(), n)
		WaitForAlerts()
		if got := received(); got != 2 {
			t.Fatalf("want no retry once delivered, got %d attempts", got)
		}
		var alertCount int32
		if err := pool.QueryRow(context.Background(), "SELECT alert_count FROM ssl_monitors WHERE id = $1", mon.ID).Scan(&alertCount); err != nil {
			t.Fatalf("query monitor: %v", err)
		}
		if alertCount != 1 {
			t.Fatalf("want the alert counted once whatever its attempts, got alert_count %d", alertCount)
		}
		if mins := nextCheckIn(t, pool, "ssl_monitors", mon.ID); mins < 23*60 {
			t.Fatalf("want the next check a day later, got %.0f minutes", mins)
		}
	})

	t.Run("an SSL alert isn't sent again once the certificate is fine", func(t *testing.T) {
		n := retryNotifiers(queries)
		org := testOrg(t, queries, pool)
		mon := testSSLMonitor(t, queries, org.ID, "this-host-does-not-exist.invalid")
		setStatus, received := statusWebhook(t, queries, org.ID, "ssl", mon.ID)
		setStatus(http.StatusInternalServerError)
		mustExecWorker(t, pool, "UPDATE ssl_monitors SET status = 'error', consecutive_failures = 1 WHERE id = $1", mon.ID)
		checkSSLMonitors(context.Background(), n)
		WaitForAlerts()

		mustExecWorker(t, pool, "UPDATE ssl_monitors SET status = 'up', alert_attempted_at = NOW() - INTERVAL '5 minutes' WHERE id = $1", mon.ID)
		retryAlerts(context.Background(), n)
		WaitForAlerts()
		if got := received(); got != 1 {
			t.Fatalf("want no retry, got %d attempts", got)
		}
	})

	t.Run("a domain warning that fails is sent again 5 minutes later, until delivered", func(t *testing.T) {
		expiresAt := time.Now().Add(20 * 24 * time.Hour).UTC().Truncate(time.Second)
		rdapSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"events": []map[string]string{{"eventAction": "expiration", "eventDate": expiresAt.Format(time.RFC3339)}}})
		}))
		defer rdapSrv.Close()
		n := retryNotifiers(queries)
		n.RDAP = rdap.NewClientWithHTTPClient(rdapSrv.Client(), rdapSrv.URL+"/domain/")
		org := testOrg(t, queries, pool)
		mon := testDomainMonitor(t, queries, org.ID, "example.com")
		setStatus, received := statusWebhook(t, queries, org.ID, "domain", mon.ID)
		setStatus(http.StatusInternalServerError)

		checkDomainMonitors(context.Background(), n)
		WaitForAlerts()
		setStatus(http.StatusOK)
		mustExecWorker(t, pool, "UPDATE domain_monitors SET alert_attempted_at = NOW() - INTERVAL '5 minutes' WHERE id = $1", mon.ID)
		retryAlerts(context.Background(), n)
		WaitForAlerts()
		if got := received(); got != 2 {
			t.Fatalf("want the warning sent again, got %d attempts", got)
		}

		forceDueNow(t, pool, "domain_monitors", mon.ID)
		checkDomainMonitors(context.Background(), n)
		mustExecWorker(t, pool, "UPDATE domain_monitors SET alert_attempted_at = NOW() - INTERVAL '1 hour' WHERE id = $1", mon.ID)
		retryAlerts(context.Background(), n)
		WaitForAlerts()
		if got := received(); got != 2 {
			t.Fatalf("want the warning not sent again once delivered, got %d attempts", got)
		}
	})

	t.Run("a domain whose registry stops answering the day it expires still alerts", func(t *testing.T) {
		// The first error after a working check is rechecked 30 s later,
		// which moves last_checked_at past the expiry.
		rdapSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer rdapSrv.Close()
		n := retryNotifiers(queries)
		n.RDAP = rdap.NewClientWithHTTPClient(rdapSrv.Client(), rdapSrv.URL+"/domain/")
		org := testOrg(t, queries, pool)
		mon := testDomainMonitor(t, queries, org.ID, "example.com")
		_, received := statusWebhook(t, queries, org.ID, "domain", mon.ID)
		mustExecWorker(t, pool, `UPDATE domain_monitors SET status = 'expiring_soon', expires_at = NOW() - INTERVAL '25 hours',
			last_checked_at = NOW() - INTERVAL '24 hours', alerted_30d = true, alerted_14d = true, alerted_7d = true,
			alert_count = 3 WHERE id = $1`, mon.ID)

		checkDomainMonitors(context.Background(), n) // first error: rechecked
		WaitForAlerts()
		forceDueNow(t, pool, "domain_monitors", mon.ID)
		checkDomainMonitors(context.Background(), n)
		WaitForAlerts()
		if got := received(); got != 1 {
			t.Fatalf("want the expired alert sent, got %d", got)
		}
	})
}

// An alert whose channels can't be read isn't delivered, so it stays due a
// retry: the retry counts an alert only when DispatchAlert says it went out.
func TestDispatchAlert_ChannelsUnreadableIsNotDelivered(t *testing.T) {
	pool := testPool(t)
	pool.Close() // a closed pool fails every query, as a dropped connection would
	n := retryNotifiers(db.New(pool))
	if DispatchAlert(context.Background(), n, uuid.New(), MonitorRef{Type: "uptime", ID: uuid.New()}, AlertMessage{EmailSubject: "DOWN: api"}) {
		t.Fatal("want an alert whose channels can't be read reported undelivered")
	}
}
