package handler

import (
	"context"
	"net/http"
	"testing"
)

// Pausing overwrites a down monitor's status, so nothing would see it as
// down again to resolve its incident: pausing resolves it, without a
// recovery alert, since the monitor wasn't seen to recover.
func TestPause_ResolvesTheOpenIncident(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)

	for _, tc := range []struct {
		monitorType string
		create      func(t *testing.T, access *http.Cookie) monitorRef
		pause       http.HandlerFunc
	}{
		{"uptime", func(t *testing.T, a *http.Cookie) monitorRef { return createUptimeMonitor(t, monitorH, a, "Down") }, monitorH.PauseUptimeMonitor},
		{"port", func(t *testing.T, a *http.Cookie) monitorRef { return createPortMonitor(t, monitorH, a, "Down") }, monitorH.PausePortMonitor},
		{"dns", func(t *testing.T, a *http.Cookie) monitorRef { return createDNSMonitor(t, monitorH, a, "Down") }, monitorH.PauseDNSMonitor},
		{"cron", func(t *testing.T, a *http.Cookie) monitorRef { return createCronMonitor(t, monitorH, a, "Down") }, monitorH.PauseCronMonitor},
	} {
		t.Run(tc.monitorType, func(t *testing.T) {
			u := signUpTestUser(t, authH, pool)
			mon := tc.create(t, u.access)
			mustExec(t, pool, "UPDATE "+tc.monitorType+"_monitors SET status = 'down' WHERE id = $1", mon.ID)
			mustExec(t, pool, "INSERT INTO "+tc.monitorType+"_incidents (monitor_id) VALUES ($1)", mon.ID)

			if w := doMonitorRequest(t, http.MethodPost, tc.pause, u.access, mon.ID, nil); w.Code != http.StatusOK {
				t.Fatalf("pause: want 200, got %d: %s", w.Code, w.Body.String())
			}
			var open int
			if err := pool.QueryRow(context.Background(),
				"SELECT COUNT(*) FROM "+tc.monitorType+"_incidents WHERE monitor_id = $1 AND resolved_at IS NULL", mon.ID,
			).Scan(&open); err != nil {
				t.Fatalf("query incidents: %v", err)
			}
			if open != 0 {
				t.Fatalf("want the incident resolved on pause, got %d open", open)
			}
		})
	}
}
