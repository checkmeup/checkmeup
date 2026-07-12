package billing

// Integration tests for enforce.go: real Postgres (ADR-010), same
// conventions as other packages' tests (testPool falls back to the
// devcontainer DSN when DATABASE_URL isn't set).

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
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

func testOrg(t *testing.T, queries *db.Queries, pool *pgxpool.Pool) db.Org {
	t.Helper()
	org, err := queries.CreateOrg(context.Background(), db.CreateOrgParams{Name: "test-" + uuid.NewString()})
	if err != nil {
		t.Fatalf("create test org: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM orgs WHERE id = $1", org.ID)
	})
	return org
}

// createAgedCronMonitor creates a cron monitor and backdates its created_at
// so ordering (newest-first) can be controlled deterministically in tests —
// CreateCronMonitor itself has no created_at param (it's always NOW()).
func createAgedCronMonitor(t *testing.T, queries *db.Queries, pool *pgxpool.Pool, orgID uuid.UUID, age time.Duration) db.CronMonitor {
	t.Helper()
	m, err := queries.CreateCronMonitor(context.Background(), db.CreateCronMonitorParams{
		OrgID: orgID, Name: "cron-" + uuid.NewString(), Schedule: "every 1h",
		GracePeriodMins: 5, PingToken: uuid.NewString(), MaxAlertsPerIncident: 3, AlertAfterNFailures: 0,
	})
	if err != nil {
		t.Fatalf("create cron monitor: %v", err)
	}
	if _, err := pool.Exec(context.Background(), "UPDATE cron_monitors SET created_at = $2 WHERE id = $1", m.ID, time.Now().Add(-age)); err != nil {
		t.Fatalf("backdate created_at: %v", err)
	}
	return m
}

func createAgedPortMonitor(t *testing.T, queries *db.Queries, pool *pgxpool.Pool, orgID uuid.UUID, age time.Duration) db.PortMonitor {
	t.Helper()
	m, err := queries.CreatePortMonitor(context.Background(), db.CreatePortMonitorParams{
		OrgID: orgID, Name: "port-" + uuid.NewString(), Host: "example.com", Port: 443,
		ExpectedState: db.PortExpectedStateOpen, IntervalMins: 5, MaxAlertsPerIncident: 3, AlertAfterNFailures: 0,
	})
	if err != nil {
		t.Fatalf("create port monitor: %v", err)
	}
	if _, err := pool.Exec(context.Background(), "UPDATE port_monitors SET created_at = $2 WHERE id = $1", m.ID, time.Now().Add(-age)); err != nil {
		t.Fatalf("backdate created_at: %v", err)
	}
	return m
}

func cronStatus(t *testing.T, queries *db.Queries, orgID, id uuid.UUID) db.MonitorStatus {
	t.Helper()
	m, err := queries.GetCronMonitor(context.Background(), db.GetCronMonitorParams{ID: id, OrgID: orgID})
	if err != nil {
		t.Fatalf("get cron monitor: %v", err)
	}
	return m.Status
}

func TestEnforceMonitorLimit(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)

	t.Run("pauses the newest active monitors beyond limit, leaves the oldest active", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		// Ages, oldest to newest: 5h, 4h, 3h, 2h, 1h — created in that order below.
		oldest := createAgedCronMonitor(t, queries, pool, org.ID, 5*time.Hour)
		second := createAgedCronMonitor(t, queries, pool, org.ID, 4*time.Hour)
		third := createAgedCronMonitor(t, queries, pool, org.ID, 3*time.Hour)
		newest2nd := createAgedCronMonitor(t, queries, pool, org.ID, 2*time.Hour)
		newest := createAgedCronMonitor(t, queries, pool, org.ID, 1*time.Hour)

		if err := EnforceMonitorLimit(context.Background(), queries, org.ID, 3); err != nil {
			t.Fatalf("enforce: %v", err)
		}

		for _, tc := range []struct {
			name string
			m    db.CronMonitor
			want db.MonitorStatus
		}{
			{"oldest stays active", oldest, db.MonitorStatusWaiting},
			{"2nd oldest stays active", second, db.MonitorStatusWaiting},
			{"3rd oldest stays active", third, db.MonitorStatusWaiting},
			{"2nd newest gets paused", newest2nd, db.MonitorStatusPaused},
			{"newest gets paused", newest, db.MonitorStatusPaused},
		} {
			if got := cronStatus(t, queries, org.ID, tc.m.ID); got != tc.want {
				t.Errorf("%s: want status %q, got %q", tc.name, tc.want, got)
			}
		}
	})

	t.Run("no-op when already at or under the limit", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		m := createAgedCronMonitor(t, queries, pool, org.ID, time.Hour)

		if err := EnforceMonitorLimit(context.Background(), queries, org.ID, 10); err != nil {
			t.Fatalf("enforce: %v", err)
		}
		if got := cronStatus(t, queries, org.ID, m.ID); got != db.MonitorStatusWaiting {
			t.Fatalf("want still active, got %q", got)
		}
	})

	t.Run("limit -1 (unlimited) is a no-op even when over any finite count", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		m := createAgedCronMonitor(t, queries, pool, org.ID, time.Hour)

		if err := EnforceMonitorLimit(context.Background(), queries, org.ID, -1); err != nil {
			t.Fatalf("enforce: %v", err)
		}
		if got := cronStatus(t, queries, org.ID, m.ID); got != db.MonitorStatusWaiting {
			t.Fatalf("want still active, got %q", got)
		}
	})

	t.Run("dispatches to the right monitor type across the aggregate (cron + port)", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		oldCron := createAgedCronMonitor(t, queries, pool, org.ID, 2*time.Hour)
		newPort := createAgedPortMonitor(t, queries, pool, org.ID, time.Hour)

		if err := EnforceMonitorLimit(context.Background(), queries, org.ID, 1); err != nil {
			t.Fatalf("enforce: %v", err)
		}

		if got := cronStatus(t, queries, org.ID, oldCron.ID); got != db.MonitorStatusWaiting {
			t.Fatalf("want the older cron monitor to stay active, got %q", got)
		}
		port, err := queries.GetPortMonitor(context.Background(), db.GetPortMonitorParams{ID: newPort.ID, OrgID: org.ID})
		if err != nil {
			t.Fatalf("get port monitor: %v", err)
		}
		if port.Status != db.MonitorStatusPaused {
			t.Fatalf("want the newer port monitor paused, got %q", port.Status)
		}
	})
}

func TestEnforceNotificationChannelLimit(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)

	createAgedChannel := func(t *testing.T, orgID uuid.UUID, age time.Duration) db.NotificationChannel {
		t.Helper()
		c, err := queries.CreateNotificationChannel(context.Background(), db.CreateNotificationChannelParams{
			OrgID: orgID, Type: db.NotificationChannelTypeEmail, Name: "chan-" + uuid.NewString(),
			Config: []byte(`{"email":"a@b.com"}`),
		})
		if err != nil {
			t.Fatalf("create channel: %v", err)
		}
		if _, err := pool.Exec(context.Background(), "UPDATE notification_channels SET created_at = $2 WHERE id = $1", c.ID, time.Now().Add(-age)); err != nil {
			t.Fatalf("backdate created_at: %v", err)
		}
		return c
	}

	t.Run("disables the newest enabled channels beyond limit, leaves the oldest enabled", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		oldest := createAgedChannel(t, org.ID, 3*time.Hour)
		middle := createAgedChannel(t, org.ID, 2*time.Hour)
		newest := createAgedChannel(t, org.ID, time.Hour)

		if err := EnforceNotificationChannelLimit(context.Background(), queries, org.ID, 2); err != nil {
			t.Fatalf("enforce: %v", err)
		}

		get := func(id uuid.UUID) db.NotificationChannel {
			c, err := queries.GetNotificationChannel(context.Background(), db.GetNotificationChannelParams{ID: id, OrgID: org.ID})
			if err != nil {
				t.Fatalf("get channel: %v", err)
			}
			return c
		}
		if !get(oldest.ID).Enabled {
			t.Error("want oldest channel to stay enabled")
		}
		if !get(middle.ID).Enabled {
			t.Error("want middle channel to stay enabled")
		}
		if get(newest.ID).Enabled {
			t.Error("want newest channel to be disabled")
		}
	})

	t.Run("already-disabled channels don't count against the limit or get touched", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		oldest := createAgedChannel(t, org.ID, 2*time.Hour)
		alreadyDisabled := createAgedChannel(t, org.ID, 90*time.Minute)
		if _, err := queries.SetNotificationChannelEnabled(context.Background(), db.SetNotificationChannelEnabledParams{
			ID: alreadyDisabled.ID, OrgID: org.ID, Enabled: false,
		}); err != nil {
			t.Fatalf("pre-disable: %v", err)
		}
		newest := createAgedChannel(t, org.ID, time.Hour)

		if err := EnforceNotificationChannelLimit(context.Background(), queries, org.ID, 2); err != nil {
			t.Fatalf("enforce: %v", err)
		}

		get := func(id uuid.UUID) db.NotificationChannel {
			c, err := queries.GetNotificationChannel(context.Background(), db.GetNotificationChannelParams{ID: id, OrgID: org.ID})
			if err != nil {
				t.Fatalf("get channel: %v", err)
			}
			return c
		}
		if !get(oldest.ID).Enabled {
			t.Error("want oldest channel to stay enabled")
		}
		if !get(newest.ID).Enabled {
			t.Error("want newest channel to stay enabled — only 2 were ever enabled, at the limit")
		}
	})
}

func TestEnforceHideBrandingLimit(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)

	createPage := func(t *testing.T, orgID uuid.UUID, hideBranding bool) db.StatusPage {
		t.Helper()
		p, err := queries.CreateStatusPage(context.Background(), db.CreateStatusPageParams{
			OrgID: orgID, Slug: "sp-" + uuid.NewString(), Title: "x",
		})
		if err != nil {
			t.Fatalf("create status page: %v", err)
		}
		p, err = queries.UpdateStatusPage(context.Background(), db.UpdateStatusPageParams{
			ID: p.ID, OrgID: orgID, Title: p.Title, HideBranding: hideBranding,
		})
		if err != nil {
			t.Fatalf("set hide_branding: %v", err)
		}
		return p
	}

	t.Run("clears hide_branding across every page in the org when not allowed", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		hidden1 := createPage(t, org.ID, true)
		hidden2 := createPage(t, org.ID, true)
		alreadyOff := createPage(t, org.ID, false)

		if err := EnforceHideBrandingLimit(context.Background(), queries, org.ID, false); err != nil {
			t.Fatalf("enforce: %v", err)
		}

		get := func(id uuid.UUID) db.StatusPage {
			p, err := queries.GetStatusPage(context.Background(), db.GetStatusPageParams{ID: id, OrgID: org.ID})
			if err != nil {
				t.Fatalf("get status page: %v", err)
			}
			return p
		}
		if get(hidden1.ID).HideBranding {
			t.Error("want hide_branding cleared")
		}
		if get(hidden2.ID).HideBranding {
			t.Error("want hide_branding cleared")
		}
		if get(alreadyOff.ID).HideBranding {
			t.Error("want an already-false page to stay false")
		}
	})

	t.Run("no-op when the plan still allows it", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		hidden := createPage(t, org.ID, true)

		if err := EnforceHideBrandingLimit(context.Background(), queries, org.ID, true); err != nil {
			t.Fatalf("enforce: %v", err)
		}

		got, err := queries.GetStatusPage(context.Background(), db.GetStatusPageParams{ID: hidden.ID, OrgID: org.ID})
		if err != nil {
			t.Fatalf("get status page: %v", err)
		}
		if !got.HideBranding {
			t.Error("want hide_branding to stay true when allowed=true")
		}
	})
}
