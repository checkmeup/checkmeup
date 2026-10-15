package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
)

// StatsHandler serves Checkmeup's platform-wide totals for the landing page
// (GET /api/v1/stats): monitors running now, and pings, checks and incidents
// over the last 30 days. Public, and cached in memory for an hour, so the
// counts' table scans run at most hourly however often the page is loaded.
type StatsHandler struct {
	queries *db.Queries
	ttl     time.Duration
	now     func() time.Time

	mu      sync.Mutex
	cached  []byte
	expires time.Time
}

type platformStats struct {
	Monitors     int64 `json:"monitors"`
	Pings30d     int64 `json:"pings30d"`
	Checks30d    int64 `json:"checks30d"`
	Incidents30d int64 `json:"incidents30d"`
}

func NewStatsHandler(pool *pgxpool.Pool) *StatsHandler {
	return &StatsHandler{queries: db.New(pool), ttl: time.Hour, now: time.Now}
}

// ServeStats handles GET /api/v1/stats.
func (h *StatsHandler) ServeStats(w http.ResponseWriter, r *http.Request) {
	body, err := h.load(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "max-age=3600")
	_, _ = w.Write(body)
}

// load returns the cached totals, recounting once they're an hour old. The
// lock is held through the recount so concurrent requests wait for one query
// rather than each starting their own; if the recount fails, the previous
// totals are served instead of an error.
func (h *StatsHandler) load(ctx context.Context) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cached != nil && h.now().Before(h.expires) {
		return h.cached, nil
	}
	row, err := h.queries.GetPlatformStats(ctx)
	if err != nil {
		if h.cached != nil {
			return h.cached, nil
		}
		return nil, err
	}
	body, err := json.Marshal(platformStats{
		Monitors:     row.Monitors,
		Pings30d:     row.Pings,
		Checks30d:    row.Checks,
		Incidents30d: row.Incidents,
	})
	if err != nil {
		return nil, err
	}
	h.cached, h.expires = body, h.now().Add(h.ttl)
	return body, nil
}
