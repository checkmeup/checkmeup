package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStatsServeStats(t *testing.T) {
	authH, monitorH, _, _, _, pool := testStatusPublicHandler(t)
	clock := time.Now()
	h := NewStatsHandler(pool)
	h.now = func() time.Time { return clock }

	get := func() platformStats {
		t.Helper()
		w := httptest.NewRecorder()
		h.ServeStats(w, httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		if cc := w.Header().Get("Cache-Control"); cc != "max-age=3600" {
			t.Fatalf("want an hour's cache TTL, got %q", cc)
		}
		var s platformStats
		if err := json.Unmarshal(w.Body.Bytes(), &s); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return s
	}

	before := get()

	u := signUpTestUser(t, authH, pool)
	cron := createCronMonitor(t, monitorH, u.access, "Stats job")
	mustExec(t, pool, "INSERT INTO cron_pings (monitor_id) VALUES ($1)", cron.ID)

	if cached := get(); cached != before {
		t.Fatalf("within the hour: want the cached totals %+v, got %+v", before, cached)
	}

	clock = clock.Add(time.Hour + time.Second)
	after := get()
	if after.Monitors <= before.Monitors || after.Pings30d <= before.Pings30d {
		t.Fatalf("after the hour: want the new monitor and ping counted, got %+v then %+v", before, after)
	}
}
