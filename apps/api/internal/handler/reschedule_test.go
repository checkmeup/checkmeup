package handler

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// nextRun reads a monitor's next scheduled check (or, for cron, next expected ping).
func nextRun(t *testing.T, pool *pgxpool.Pool, table, column, id string) time.Time {
	t.Helper()
	var at time.Time
	if err := pool.QueryRow(context.Background(), "SELECT "+column+" FROM "+table+" WHERE id = $1", id).Scan(&at); err != nil {
		t.Fatalf("read %s.%s: %v", table, column, err)
	}
	return at
}

func TestUpdateMonitor_ShorterIntervalReschedulesTheNextCheck(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	for _, tc := range []struct {
		table  string
		create func(*testing.T, *MonitorHandler, *http.Cookie, string) monitorRef
		update http.HandlerFunc
		body   any
	}{
		{"uptime_monitors", createUptimeMonitor, monitorH.UpdateUptimeMonitor, map[string]any{
			"name": "x", "url": "https://example.com/health", "intervalMins": 1,
			"maxResponseTimeMs": 10000, "httpMethod": "GET", "acceptedStatusCodes": []int32{200},
		}},
		{"port_monitors", createPortMonitor, monitorH.UpdatePortMonitor, map[string]any{
			"name": "x", "host": "example.com", "port": 443, "expectedState": "open", "intervalMins": 1,
		}},
		{"dns_monitors", createDNSMonitor, monitorH.UpdateDNSMonitor, map[string]any{
			"name": "x", "hostname": "example.com", "recordType": "A", "intervalMins": 1,
		}},
	} {
		t.Run(tc.table, func(t *testing.T) {
			u := signUpTestUser(t, authH, pool)
			mustExec(t, pool, "UPDATE orgs SET plan = 'startup' WHERE id = $1", u.resp.OrgID)
			mon := tc.create(t, monitorH, u.access, "x")
			patch := func(t *testing.T) {
				t.Helper()
				if w := doMonitorRequest(t, http.MethodPatch, tc.update, u.access, mon.ID, tc.body); w.Code != http.StatusOK {
					t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
				}
			}

			t.Run("moves a 10-minute schedule up to the new 1-minute one", func(t *testing.T) {
				mustExec(t, pool, "UPDATE "+tc.table+" SET last_checked_at = NOW() - INTERVAL '5 minutes', next_check_at = NOW() + INTERVAL '5 minutes' WHERE id = $1", mon.ID)
				patch(t)
				if next := nextRun(t, pool, tc.table, "next_check_at", mon.ID); next.After(time.Now()) {
					t.Fatalf("next_check_at = %v, want due now (last check 5 min ago + 1 min)", next)
				}
			})

			t.Run("leaves a pending first check alone", func(t *testing.T) {
				mustExec(t, pool, "UPDATE "+tc.table+" SET last_checked_at = NULL, next_check_at = NOW() WHERE id = $1", mon.ID)
				patch(t)
				if next := nextRun(t, pool, tc.table, "next_check_at", mon.ID); next.After(time.Now()) {
					t.Fatalf("next_check_at = %v, want still due now", next)
				}
			})
		})
	}
}

func TestUpdateCronMonitor_ScheduleChangeReschedulesTheNextPing(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)
	u := signUpTestUser(t, authH, pool)
	mon := createCronMonitor(t, monitorH, u.access, "x") // every 1h
	update := func(t *testing.T, schedule string) {
		t.Helper()
		w := doMonitorRequest(t, http.MethodPatch, monitorH.UpdateCronMonitor, u.access, mon.ID, updateCronMonitorRequest{Name: "renamed", Schedule: schedule})
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
		}
	}

	t.Run("a new schedule gets one full window from now", func(t *testing.T) {
		mustExec(t, pool, "UPDATE cron_monitors SET status = 'up', next_ping_at = NOW() + INTERVAL '50 minutes' WHERE id = $1", mon.ID)
		update(t, "every 1d")
		next := nextRun(t, pool, "cron_monitors", "next_ping_at", mon.ID)
		// every 1d plus the default 5-minute grace period
		if want := time.Now().Add(24*time.Hour + 5*time.Minute); next.Before(want.Add(-time.Minute)) || next.After(want.Add(time.Minute)) {
			t.Fatalf("next_ping_at = %v, want about %v", next, want)
		}
	})

	t.Run("an edit that keeps the schedule leaves the next ping alone", func(t *testing.T) {
		want := time.Now().Add(37 * time.Minute).Truncate(time.Second)
		mustExec(t, pool, "UPDATE cron_monitors SET next_ping_at = $2 WHERE id = $1", mon.ID, want)
		update(t, "every 1d")
		if next := nextRun(t, pool, "cron_monitors", "next_ping_at", mon.ID); !next.Equal(want) {
			t.Fatalf("next_ping_at = %v, want unchanged %v", next, want)
		}
	})

	t.Run("a monitor waiting for its first ping stays waiting", func(t *testing.T) {
		mustExec(t, pool, "UPDATE cron_monitors SET next_ping_at = NULL WHERE id = $1", mon.ID)
		update(t, "every 5m")
		var next *time.Time
		if err := pool.QueryRow(context.Background(), "SELECT next_ping_at FROM cron_monitors WHERE id = $1", mon.ID).Scan(&next); err != nil {
			t.Fatalf("read next_ping_at: %v", err)
		}
		if next != nil {
			t.Fatalf("next_ping_at = %v, want NULL until the first ping", *next)
		}
	})
}
