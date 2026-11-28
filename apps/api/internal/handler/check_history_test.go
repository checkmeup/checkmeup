package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func tickString(p *string) string {
	if p == nil {
		return "-"
	}
	return *p
}

func ticksString(ticks []*string) string {
	out := ""
	for _, tk := range ticks {
		switch tickString(tk) {
		case "up":
			out += "u"
		case "down":
			out += "d"
		case "warn":
			out += "w"
		default:
			out += "-"
		}
	}
	return out
}

func TestBuildCheckHistory(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 30, 0, 0, time.UTC)
	id := uuid.New()
	hour := func(hoursAgo int) time.Time { return time.Date(2026, 10, 15, 12-hoursAgo, 0, 0, 0, time.UTC) }

	t.Run("places hours oldest first ending with the current hour", func(t *testing.T) {
		got := buildCheckHistory(now, 4, []hourResult{
			{"uptime", id, hour(0), 5, 0},
			{"uptime", id, hour(2), 5, 1},
		})
		if len(got) != 1 {
			t.Fatalf("monitors = %d, want 1", len(got))
		}
		if s := ticksString(got[0].Ticks); s != "-d-u" {
			t.Fatalf("ticks = %q, want %q", s, "-d-u")
		}
	})

	t.Run("ignores hours outside the window", func(t *testing.T) {
		got := buildCheckHistory(now, 4, []hourResult{
			{"uptime", id, hour(4), 5, 5},
			{"uptime", id, hour(-1), 5, 5},
		})
		if len(got) != 0 {
			t.Fatalf("monitors = %d, want 0", len(got))
		}
	})

	t.Run("a failure in an hour wins over passing checks in the same hour", func(t *testing.T) {
		got := buildCheckHistory(now, 2, []hourResult{
			{"cron", id, hour(0), 0, 1},
			{"cron", id, hour(0), 3, 0},
		})
		if s := ticksString(got[0].Ticks); s != "-d" {
			t.Fatalf("ticks = %q, want %q", s, "-d")
		}
	})

	t.Run("keeps monitors of different types with the same id apart", func(t *testing.T) {
		got := buildCheckHistory(now, 1, []hourResult{
			{"uptime", id, hour(0), 1, 0},
			{"port", id, hour(0), 1, 1},
		})
		if len(got) != 2 {
			t.Fatalf("monitors = %d, want 2", len(got))
		}
	})
}

func TestExpiryTicks(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 30, 0, 0, time.UTC)
	cases := []struct {
		name    string
		created time.Time
		status  string
		want    string
	}{
		{"valid since before the window: every hour up", now.Add(-100 * time.Hour), "up", "uuuu"},
		{"created 2 hours ago: the hours before have no data", now.Add(-2 * time.Hour), "up", "-uuu"},
		{"expiring soon is warn", now.Add(-100 * time.Hour), "expiring_soon", "wwww"},
		{"expired is down", now.Add(-100 * time.Hour), "expired", "dddd"},
		{"a lookup error is down", now.Add(-100 * time.Hour), "error", "dddd"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ticksString(expiryTicks(now, 4, tc.created, tc.status)); got != tc.want {
				t.Fatalf("ticks = %q, want %q", got, tc.want)
			}
		})
	}
	for _, status := range []string{"waiting", "paused"} {
		if got := expiryTicks(now, 4, now.Add(-100*time.Hour), status); got != nil {
			t.Fatalf("%s: ticks = %v, want no strip", status, got)
		}
	}
}

func TestListCheckHistory(t *testing.T) {
	authH, monitorH, pool := testMonitorHandler(t)
	ctx := context.Background()

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/monitors/check-history", nil)
		w := httptest.NewRecorder()
		monitorH.ListCheckHistory(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("returns hourly uptime results and cron incident hours for the caller's org only", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		other := signUpTestUser(t, authH, pool)
		mine := createUptimeMonitor(t, monitorH, u.access, "API")
		theirs := createUptimeMonitor(t, monitorH, other.access, "Theirs")
		job := createCronMonitor(t, monitorH, u.access, "Backup")

		insertCheck := func(monitorID string, ago time.Duration, up bool) {
			t.Helper()
			if _, err := pool.Exec(ctx,
				`INSERT INTO uptime_checks (monitor_id, response_time_ms, is_up, checked_at) VALUES ($1, 100, $2, NOW() - $3::interval)`,
				monitorID, up, ago.String()); err != nil {
				t.Fatalf("insert uptime check: %v", err)
			}
		}
		insertCheck(mine.ID, 0, true)
		insertCheck(mine.ID, 3*time.Hour, false)
		insertCheck(mine.ID, 100*time.Hour, false) // outside the 48h window
		insertCheck(theirs.ID, 0, false)
		if _, err := pool.Exec(ctx,
			`INSERT INTO cron_incidents (monitor_id, started_at, resolved_at) VALUES ($1, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '4 hours')`,
			job.ID); err != nil {
			t.Fatalf("insert cron incident: %v", err)
		}

		w := doAuthed(t, http.MethodGet, monitorH.ListCheckHistory, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
		}
		body := decodeBody[checkHistoryResponse](t, w)
		if body.Hours != 48 {
			t.Fatalf("hours = %d, want 48", body.Hours)
		}

		found := map[string]checkHistoryMonitor{}
		for _, m := range body.Monitors {
			found[m.Type+":"+m.ID] = m
		}
		if _, ok := found["uptime:"+theirs.ID]; ok {
			t.Fatalf("another org's monitor leaked into the history")
		}

		api, ok := found["uptime:"+mine.ID]
		if !ok {
			t.Fatalf("own uptime monitor missing from %v", body.Monitors)
		}
		if len(api.Ticks) != 48 {
			t.Fatalf("ticks = %d, want 48", len(api.Ticks))
		}
		if got := tickString(api.Ticks[47]); got != "up" {
			t.Fatalf("current hour = %q, want up", got)
		}
		if got := tickString(api.Ticks[44]); got != "down" {
			t.Fatalf("3 hours ago = %q, want down", got)
		}
		if got := tickString(api.Ticks[0]); got != "-" {
			t.Fatalf("47 hours ago = %q, want no checks", got)
		}

		backup, ok := found["cron:"+job.ID]
		if !ok {
			t.Fatalf("cron monitor with an incident missing from %v", body.Monitors)
		}
		if got := tickString(backup.Ticks[42]); got != "down" {
			t.Fatalf("cron incident hour (5h ago) = %q, want down", got)
		}
	})

	t.Run("a cron monitor is up every hour from its first ping on, not only the hours a ping landed in", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		daily := createCronMonitor(t, monitorH, u.access, "Daily")
		paused := createCronMonitor(t, monitorH, u.access, "Paused")
		never := createCronMonitor(t, monitorH, u.access, "Never pinged")
		ping := func(monitorID string, ago string) {
			t.Helper()
			if _, err := pool.Exec(ctx,
				`INSERT INTO cron_pings (monitor_id, received_at) VALUES ($1, NOW() - $2::interval)`,
				monitorID, ago); err != nil {
				t.Fatalf("insert cron ping: %v", err)
			}
		}
		ping(daily.ID, "10 hours")
		ping(paused.ID, "20 hours")
		if _, err := pool.Exec(ctx,
			`UPDATE cron_monitors SET status = 'paused', last_ping_at = NOW() - INTERVAL '10 hours' WHERE id = $1`,
			paused.ID); err != nil {
			t.Fatalf("pause: %v", err)
		}

		body := decodeBody[checkHistoryResponse](t, doAuthed(t, http.MethodGet, monitorH.ListCheckHistory, u.access, nil))
		found := map[string][]*string{}
		for _, m := range body.Monitors {
			found[m.ID] = m.Ticks
		}

		d := found[daily.ID]
		if got := tickString(d[36]); got != "-" {
			t.Fatalf("daily, 11h ago (before its first ping) = %q, want no data", got)
		}
		for i := 37; i <= 47; i++ {
			if got := tickString(d[i]); got != "up" {
				t.Fatalf("daily, %dh ago = %q, want up", 47-i, got)
			}
		}
		p := found[paused.ID]
		if got := tickString(p[37]); got != "up" {
			t.Fatalf("paused, 10h ago (its last ping) = %q, want up", got)
		}
		if got := tickString(p[38]); got != "-" {
			t.Fatalf("paused, 9h ago (after its last ping) = %q, want no data", got)
		}
		if _, ok := found[never.ID]; ok {
			t.Fatalf("a cron monitor with no ping has a check strip: %v", found[never.ID])
		}
	})

	t.Run("an SSL or domain monitor's strip is its status since it was created, and only once checked", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		other := signUpTestUser(t, authH, pool)
		cert := createSSLMonitor(t, monitorH, u.access, "Cert")
		dom := createDomainMonitor(t, monitorH, u.access, "Domain")
		unchecked := createSSLMonitor(t, monitorH, u.access, "Not checked yet")
		theirs := createSSLMonitor(t, monitorH, other.access, "Theirs")
		exec := func(q, id string) {
			t.Helper()
			if _, err := pool.Exec(ctx, q, id); err != nil {
				t.Fatalf("%s: %v", q, err)
			}
		}
		exec(`UPDATE ssl_monitors SET status = 'up', created_at = NOW() - INTERVAL '5 hours' WHERE id = $1`, cert.ID)
		exec(`UPDATE domain_monitors SET status = 'expiring_soon', created_at = NOW() - INTERVAL '100 hours' WHERE id = $1`, dom.ID)
		exec(`UPDATE ssl_monitors SET status = 'up' WHERE id = $1`, theirs.ID)

		body := decodeBody[checkHistoryResponse](t, doAuthed(t, http.MethodGet, monitorH.ListCheckHistory, u.access, nil))
		found := map[string][]*string{}
		for _, m := range body.Monitors {
			found[m.Type+":"+m.ID] = m.Ticks
		}
		c := found["ssl:"+cert.ID]
		if got := tickString(c[41]); got != "-" {
			t.Fatalf("cert, 6h ago (before it was created) = %q, want no data", got)
		}
		if got := tickString(c[42]); got != "up" {
			t.Fatalf("cert, 5h ago (the hour it was created) = %q, want up", got)
		}
		if got := tickString(found["domain:"+dom.ID][0]); got != "warn" {
			t.Fatalf("expiring domain, 47h ago = %q, want warn", got)
		}
		if _, ok := found["ssl:"+unchecked.ID]; ok {
			t.Fatalf("an SSL monitor not checked yet has a strip")
		}
		if _, ok := found["ssl:"+theirs.ID]; ok {
			t.Fatalf("another org's SSL monitor leaked into the history")
		}
	})
}
