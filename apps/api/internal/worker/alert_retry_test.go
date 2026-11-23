package worker

// Tests for PRD alert-delivery: a down alert that didn't go out is sent again
// every 5 minutes while its incident is open and its monitor down.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/rdap"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/twilio"
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

// acceptingTwilio is a Twilio that accepts every SMS, each with a fresh SID,
// and lists the SIDs it gave out, oldest first.
func acceptingTwilio(t *testing.T) (client *twilio.Client, sids func() []string) {
	t.Helper()
	var mu sync.Mutex
	var given []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		sid := "SM" + strings.ReplaceAll(uuid.NewString(), "-", "")
		mu.Lock()
		given = append(given, sid)
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"sid": sid})
	}))
	t.Cleanup(srv.Close)
	return twilio.NewClientWithBaseURL(srv.URL, srv.Client()), func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), given...)
	}
}

// smsOnlyOrg is an org on a plan with SMS credits and no users, so a
// Fallback email has nobody to reach.
func smsOnlyOrg(t *testing.T, queries *db.Queries, pool *pgxpool.Pool) db.Org {
	t.Helper()
	org := testOrg(t, queries, pool)
	upgradeOrgPlan(t, queries, org.ID, db.PlanSolo)
	return org
}

// attachSMS attaches an SMS channel to the monitor, its only channel.
func attachSMS(t *testing.T, queries *db.Queries, orgID uuid.UUID, monitorType string, monitorID uuid.UUID) {
	t.Helper()
	c := testNotificationChannel(t, queries, orgID, db.NotificationChannelTypeSms, map[string]string{"phone_number": "+15005550006"})
	attachNotificationChannel(t, queries, c.ID, monitorType, monitorID)
}

// failAtCarrier delivers Twilio's callback that sid was undelivered, and
// reports whether its Fallback email went out.
func failAtCarrier(t *testing.T, n Notifiers, sid string) bool {
	t.Helper()
	return HandleSMSStatus(context.Background(), n, sid, "undelivered", "30003")
}

// lastSID is the SID of the latest SMS sent, failing the test when none was.
func lastSID(t *testing.T, sids []string) string {
	t.Helper()
	if len(sids) == 0 {
		t.Fatal("want an SMS sent, got none")
	}
	return sids[len(sids)-1]
}

// attemptedAgo sets back when the monitor's alert was last attempted: on its
// open incident, or for SSL and domain, on the monitor.
func attemptedAgo(t *testing.T, pool *pgxpool.Pool, monitorType string, monitorID uuid.UUID, interval string) {
	t.Helper()
	where := monitorType + "_incidents SET alert_attempted_at = NOW() - $2::interval WHERE resolved_at IS NULL AND monitor_id = $1"
	if monitorType == "ssl" || monitorType == "domain" {
		where = monitorType + "_monitors SET alert_attempted_at = NOW() - $2::interval WHERE id = $1"
	}
	mustExecWorker(t, pool, "UPDATE "+where, monitorID, interval)
}

// retryNow runs one retry pass and waits for what it sends.
func retryNow(n Notifiers) {
	retryAlerts(context.Background(), n)
	WaitForAlerts()
}

// smsRetryNotifiers are retryNotifiers whose SMS goes to an acceptingTwilio.
func smsRetryNotifiers(t *testing.T, queries *db.Queries) (Notifiers, func() []string) {
	t.Helper()
	n := retryNotifiers(queries)
	var sids func() []string
	n.SMS, sids = acceptingTwilio(t)
	return n, sids
}

func TestRetryAfterFailedSMS(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	ctx := context.Background()

	t.Run("a down alert whose only SMS and Fallback email fail is sent again 5 minutes later, until delivered, then not again", func(t *testing.T) {
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer target.Close()
		n := retryNotifiers(queries)
		var sids func() []string
		n.SMS, sids = acceptingTwilio(t)
		org := smsOnlyOrg(t, queries, pool)
		mon := testUptimeMonitor(t, queries, org.ID, target.URL)
		attachSMS(t, queries, org.ID, "uptime", mon.ID)

		checkUptimeMonitors(ctx, n) // rechecked before it counts
		forceDueNow(t, pool, "uptime_monitors", mon.ID)
		checkUptimeMonitors(ctx, n)
		WaitForAlerts()
		if got := len(sids()); got != 1 {
			t.Fatalf("want the down alert's SMS sent once, got %d", got)
		}
		if failAtCarrier(t, n, lastSID(t, sids())) {
			t.Fatal("want the Fallback email to fail: the org has no users")
		}

		retryAlerts(ctx, n)
		WaitForAlerts()
		if got := len(sids()); got != 1 {
			t.Fatalf("want no retry within 5 minutes of the attempt, got %d SMS", got)
		}

		for want := 2; want <= 3; want++ {
			mustExecWorker(t, pool, "UPDATE uptime_incidents SET alert_attempted_at = NOW() - INTERVAL '5 minutes' WHERE monitor_id = $1", mon.ID)
			retryAlerts(ctx, n)
			WaitForAlerts()
			if got := len(sids()); got != want {
				t.Fatalf("want the alert sent again 5 minutes after the attempt, got %d SMS, want %d", got, want)
			}
			if want == 2 {
				failAtCarrier(t, n, lastSID(t, sids()))
			}
		}
		if HandleSMSStatus(ctx, n, lastSID(t, sids()), "delivered", "") {
			t.Fatal("want no Fallback email for a delivered SMS")
		}
		if got := incidentAlertCount(t, pool, "uptime_incidents", mon.ID); got != 1 {
			t.Fatalf("want the delivered alert counted once, got alert_count %d", got)
		}

		mustExecWorker(t, pool, "UPDATE uptime_incidents SET alert_attempted_at = NOW() - INTERVAL '1 hour' WHERE monitor_id = $1", mon.ID)
		retryAlerts(ctx, n)
		WaitForAlerts()
		if got := len(sids()); got != 3 {
			t.Fatalf("want no retry once delivered, got %d SMS", got)
		}
	})

	t.Run("every incident-tracked type hands a failed only-delivery SMS back to its retry", func(t *testing.T) {
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
				n, sids := smsRetryNotifiers(t, queries)
				org := smsOnlyOrg(t, queries, pool)
				id := tc.create(org.ID)
				attachSMS(t, queries, org.ID, tc.monitorType, id)
				seedUndeliveredIncident(t, pool, tc.monitorType, id)

				retryNow(n)
				if got := incidentAlertCount(t, pool, tc.monitorType+"_incidents", id); got != 1 {
					t.Fatalf("want the accepted SMS counted as the delivery, got alert_count %d", got)
				}
				failAtCarrier(t, n, lastSID(t, sids()))
				if got := incidentAlertCount(t, pool, tc.monitorType+"_incidents", id); got != 0 {
					t.Fatalf("want the delivery undone, got alert_count %d", got)
				}
				attemptedAgo(t, pool, tc.monitorType, id, "5 minutes")
				retryNow(n)
				if got := len(sids()); got != 2 {
					t.Fatalf("want the down alert sent again, got %d SMS", got)
				}
			})
		}
	})

	t.Run("a cron missed-run alert whose only SMS and Fallback email fail is sent again", func(t *testing.T) {
		n, sids := smsRetryNotifiers(t, queries)
		org := smsOnlyOrg(t, queries, pool)
		mon := testCronMonitor(t, queries, org.ID, "every 1h")
		attachSMS(t, queries, org.ID, "cron", mon.ID)
		mustExecWorker(t, pool, "UPDATE cron_monitors SET status = 'up', next_ping_at = NOW() - INTERVAL '1 hour' WHERE id = $1", mon.ID)

		checkOverdue(ctx, n)
		WaitForAlerts()
		failAtCarrier(t, n, lastSID(t, sids()))
		attemptedAgo(t, pool, "cron", mon.ID, "5 minutes")
		retryNow(n)
		if got := len(sids()); got != 2 {
			t.Fatalf("want the missed-run alert sent again, got %d SMS", got)
		}
	})

	t.Run("an SSL error alert whose only SMS and Fallback email fail is sent again 5 minutes later, until delivered", func(t *testing.T) {
		n, sids := smsRetryNotifiers(t, queries)
		org := smsOnlyOrg(t, queries, pool)
		mon := testSSLMonitor(t, queries, org.ID, "this-host-does-not-exist.invalid")
		attachSMS(t, queries, org.ID, "ssl", mon.ID)
		// Already in error, so this failure isn't held for a recheck.
		mustExecWorker(t, pool, "UPDATE ssl_monitors SET status = 'error', consecutive_failures = 1 WHERE id = $1", mon.ID)

		checkSSLMonitors(ctx, n)
		WaitForAlerts()
		failAtCarrier(t, n, lastSID(t, sids()))
		retryNow(n)
		if got := len(sids()); got != 1 {
			t.Fatalf("want no retry within 5 minutes of the attempt, got %d SMS", got)
		}

		attemptedAgo(t, pool, "ssl", mon.ID, "5 minutes")
		retryNow(n)
		if got := len(sids()); got != 2 {
			t.Fatalf("want the SSL alert sent again, got %d SMS", got)
		}
		HandleSMSStatus(ctx, n, lastSID(t, sids()), "delivered", "")
		attemptedAgo(t, pool, "ssl", mon.ID, "1 hour")
		retryNow(n)
		if got := len(sids()); got != 2 {
			t.Fatalf("want no retry once delivered, got %d SMS", got)
		}
	})

	t.Run("a domain warning whose only SMS and Fallback email fail is sent again 5 minutes later, until delivered", func(t *testing.T) {
		expiresAt := time.Now().Add(20 * 24 * time.Hour).UTC().Truncate(time.Second)
		rdapSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"events": []map[string]string{{"eventAction": "expiration", "eventDate": expiresAt.Format(time.RFC3339)}}})
		}))
		defer rdapSrv.Close()
		n, sids := smsRetryNotifiers(t, queries)
		n.RDAP = rdap.NewClientWithHTTPClient(rdapSrv.Client(), rdapSrv.URL+"/domain/")
		org := smsOnlyOrg(t, queries, pool)
		mon := testDomainMonitor(t, queries, org.ID, "example.com")
		attachSMS(t, queries, org.ID, "domain", mon.ID)

		checkDomainMonitors(ctx, n)
		WaitForAlerts()
		failAtCarrier(t, n, lastSID(t, sids()))
		attemptedAgo(t, pool, "domain", mon.ID, "5 minutes")
		retryNow(n)
		if got := len(sids()); got != 2 {
			t.Fatalf("want the warning sent again, got %d SMS", got)
		}
		HandleSMSStatus(ctx, n, lastSID(t, sids()), "delivered", "")
		attemptedAgo(t, pool, "domain", mon.ID, "1 hour")
		retryNow(n)
		if got := len(sids()); got != 2 {
			t.Fatalf("want no retry once delivered, got %d SMS", got)
		}
	})

	t.Run("an alert handed back tries its SMS at most 3 times, the failed ones counted", func(t *testing.T) {
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer target.Close()
		n, sids := smsRetryNotifiers(t, queries)
		org := smsOnlyOrg(t, queries, pool)
		mon := testUptimeMonitor(t, queries, org.ID, target.URL)
		attachSMS(t, queries, org.ID, "uptime", mon.ID)
		checkUptimeMonitors(ctx, n)
		forceDueNow(t, pool, "uptime_monitors", mon.ID)
		checkUptimeMonitors(ctx, n)
		WaitForAlerts()

		for attempt := 2; attempt <= 5; attempt++ {
			if got := len(sids()); got == attempt-1 {
				failAtCarrier(t, n, lastSID(t, sids()))
			}
			attemptedAgo(t, pool, "uptime", mon.ID, "5 minutes")
			retryNow(n)
		}
		if got := len(sids()); got != 3 {
			t.Fatalf("want 3 SMS in all, got %d", got)
		}
		var attempts int32
		if err := pool.QueryRow(ctx, "SELECT alert_attempts FROM uptime_incidents WHERE monitor_id = $1", mon.ID).Scan(&attempts); err != nil {
			t.Fatalf("query incident: %v", err)
		}
		if attempts != 5 {
			t.Fatalf("want the alert still attempted without its SMS, got %d attempts", attempts)
		}
	})

	t.Run("a failed SMS whose Fallback email delivered isn't retried", func(t *testing.T) {
		n, sids := smsRetryNotifiers(t, queries)
		org := smsOnlyOrg(t, queries, pool)
		testUser(t, queries, org.ID) // the Fallback email reaches them
		mon := testUptimeMonitor(t, queries, org.ID, "http://example.invalid")
		attachSMS(t, queries, org.ID, "uptime", mon.ID)
		seedUndeliveredIncident(t, pool, "uptime", mon.ID)

		retryNow(n)
		if !failAtCarrier(t, n, lastSID(t, sids())) {
			t.Fatal("want the Fallback email sent")
		}
		attemptedAgo(t, pool, "uptime", mon.ID, "1 hour")
		retryNow(n)
		if got := len(sids()); got != 1 {
			t.Fatalf("want no retry once the Fallback email delivered it, got %d SMS", got)
		}
		if got := incidentAlertCount(t, pool, "uptime_incidents", mon.ID); got != 1 {
			t.Fatalf("want the alert still counted delivered, got alert_count %d", got)
		}
	})

	t.Run("a failed SMS on an incident no longer open hands nothing back", func(t *testing.T) {
		n, sids := smsRetryNotifiers(t, queries)
		org := smsOnlyOrg(t, queries, pool)
		mon := testUptimeMonitor(t, queries, org.ID, "http://example.invalid")
		attachSMS(t, queries, org.ID, "uptime", mon.ID)
		seedUndeliveredIncident(t, pool, "uptime", mon.ID)

		retryNow(n)
		mustExecWorker(t, pool, "UPDATE uptime_incidents SET resolved_at = NOW() WHERE monitor_id = $1", mon.ID)
		failAtCarrier(t, n, lastSID(t, sids()))
		retryNow(n)
		if got := incidentAlertCount(t, pool, "uptime_incidents", mon.ID); got != 1 {
			t.Fatalf("want the resolved incident left as it was, got alert_count %d", got)
		}
	})

	t.Run("an undo never takes back a delivery made after the SMS", func(t *testing.T) {
		n, sids := smsRetryNotifiers(t, queries)
		org := smsOnlyOrg(t, queries, pool)
		mon := testUptimeMonitor(t, queries, org.ID, "http://example.invalid")
		attachSMS(t, queries, org.ID, "uptime", mon.ID)
		seedUndeliveredIncident(t, pool, "uptime", mon.ID)

		retryNow(n)
		// The SMS is older than the attempt that delivered the alert.
		mustExecWorker(t, pool, "UPDATE sms_messages SET created_at = NOW() - INTERVAL '10 minutes' WHERE sid = $1", lastSID(t, sids()))
		failAtCarrier(t, n, lastSID(t, sids()))
		attemptedAgo(t, pool, "uptime", mon.ID, "6 minutes")
		retryNow(n)
		if got := incidentAlertCount(t, pool, "uptime_incidents", mon.ID); got != 1 {
			t.Fatalf("want the later delivery kept, got alert_count %d", got)
		}
		if got := len(sids()); got != 1 {
			t.Fatalf("want no retry, got %d SMS", got)
		}
	})

	t.Run("a handed-back SSL or domain alert never replaces a newer one", func(t *testing.T) {
		for _, tc := range []struct {
			monitorType string
			create      func(orgID uuid.UUID) uuid.UUID
		}{
			{"ssl", func(orgID uuid.UUID) uuid.UUID { return testSSLMonitor(t, queries, orgID, "example.invalid").ID }},
			{"domain", func(orgID uuid.UUID) uuid.UUID { return testDomainMonitor(t, queries, orgID, "example.com").ID }},
		} {
			t.Run(tc.monitorType, func(t *testing.T) {
				n, sids := smsRetryNotifiers(t, queries)
				org := smsOnlyOrg(t, queries, pool)
				id := tc.create(org.ID)
				attachSMS(t, queries, org.ID, tc.monitorType, id)
				table := tc.monitorType + "_monitors"
				older, err := json.Marshal(AlertMessage{EmailSubject: "older", SMS: "Checkmeup: older"})
				if err != nil {
					t.Fatalf("marshal alert: %v", err)
				}
				mustExecWorker(t, pool, "UPDATE "+table+" SET status = 'error', alert = $2, alert_attempted_at = NOW() - INTERVAL '6 minutes', alert_attempts = 1 WHERE id = $1", id, older)

				retryNow(n)
				// A newer check's alert, its attempt set back so only its
				// being there keeps the older one out.
				newer, err := json.Marshal(AlertMessage{EmailSubject: "newer", SMS: "Checkmeup: newer"})
				if err != nil {
					t.Fatalf("marshal alert: %v", err)
				}
				mustExecWorker(t, pool, "UPDATE "+table+" SET alert = $2, alert_attempted_at = NOW() - INTERVAL '1 hour' WHERE id = $1", id, newer)
				failAtCarrier(t, n, lastSID(t, sids()))
				undoPendingDeliveries(ctx, n)

				var subject string
				if err := pool.QueryRow(ctx, "SELECT alert->>'EmailSubject' FROM "+table+" WHERE id = $1", id).Scan(&subject); err != nil {
					t.Fatalf("query monitor: %v", err)
				}
				if subject != "newer" {
					t.Fatalf("want the newer alert kept, got %q", subject)
				}
			})
		}
	})

	t.Run("a failure callback that arrives before the alert is counted undoes it on the next tick", func(t *testing.T) {
		n, sids := smsRetryNotifiers(t, queries)
		org := smsOnlyOrg(t, queries, pool)
		mon := testUptimeMonitor(t, queries, org.ID, "http://example.invalid")
		attachSMS(t, queries, org.ID, "uptime", mon.ID)
		seedUndeliveredIncident(t, pool, "uptime", mon.ID)
		var incidentID, channelID uuid.UUID
		if err := pool.QueryRow(ctx, "SELECT id FROM uptime_incidents WHERE monitor_id = $1", mon.ID).Scan(&incidentID); err != nil {
			t.Fatalf("query incident: %v", err)
		}
		if err := pool.QueryRow(ctx, "SELECT channel_id FROM monitor_notification_channels WHERE monitor_id = $1", mon.ID).Scan(&channelID); err != nil {
			t.Fatalf("query channel: %v", err)
		}
		// The SMS of the alert's attempt 6 minutes ago, not yet counted delivered.
		sid := "SM" + strings.ReplaceAll(uuid.NewString(), "-", "")
		if err := queries.InsertSMSMessage(ctx, db.InsertSMSMessageParams{
			Sid: sid, OrgID: org.ID, ChannelID: channelID,
			FallbackSubject: pgtype.Text{String: "DOWN: Seeded", Valid: true}, FallbackHtml: pgtype.Text{String: "<p>down</p>", Valid: true},
			MonitorType: pgtype.Text{String: "uptime", Valid: true}, AlertSourceID: pgtype.UUID{Bytes: incidentID, Valid: true},
		}); err != nil {
			t.Fatalf("insert sms message: %v", err)
		}

		failAtCarrier(t, n, sid)
		mustExecWorker(t, pool, "UPDATE uptime_incidents SET alert_count = 1 WHERE id = $1", incidentID) // the count lands
		retryNow(n)
		if got := len(sids()); got != 1 {
			t.Fatalf("want the alert sent again on the next tick, got %d SMS", got)
		}
		var left int
		if err := pool.QueryRow(ctx, "SELECT count(*) FROM sms_messages WHERE sid = $1", sid).Scan(&left); err != nil {
			t.Fatalf("query sms messages: %v", err)
		}
		if left != 0 {
			t.Fatal("want the message dropped once its undo applied")
		}
	})

	t.Run("a failed recovery or stuck-run SMS stays one-shot", func(t *testing.T) {
		n, sids := smsRetryNotifiers(t, queries)
		org := smsOnlyOrg(t, queries, pool)
		mon := testCronMonitor(t, queries, org.ID, "every 1h")
		attachSMS(t, queries, org.ID, "cron", mon.ID)
		seedUndeliveredIncident(t, pool, "cron", mon.ID)
		mustExecWorker(t, pool, "UPDATE cron_incidents SET alert_count = 1 WHERE monitor_id = $1", mon.ID) // its down alert delivered
		mustExecWorker(t, pool, "UPDATE cron_monitors SET max_duration_mins = 30 WHERE id = $1", mon.ID)
		testCronRun(t, pool, mon.ID, time.Now().Add(-2*time.Hour), false)

		checkStuckCronRuns(ctx, n)
		WaitForAlerts()
		failAtCarrier(t, n, lastSID(t, sids()))
		retryNow(n)
		if got := incidentAlertCount(t, pool, "cron_incidents", mon.ID); got != 1 {
			t.Fatalf("want the down alert's delivery kept, got alert_count %d", got)
		}
		if got := len(sids()); got != 1 {
			t.Fatalf("want nothing sent again, got %d SMS", got)
		}
	})
}
