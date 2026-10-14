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
}
