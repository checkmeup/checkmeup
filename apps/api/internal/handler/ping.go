package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"

	"github.com/checkmeup/checkmeup/internal/db"
)

type PingHandler struct {
	queries *db.Queries
}

func NewPingHandler(pool *pgxpool.Pool) *PingHandler {
	return &PingHandler{queries: db.New(pool)}
}

// ReceivePing handles GET /ping/{token}
// Always returns 200 so the calling job never fails due to monitoring being down.
// Returns 404 only when the token has been deleted (monitor no longer exists).
func (h *PingHandler) ReceivePing(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	monitor, err := h.queries.GetCronMonitorByToken(r.Context(), token)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	now := time.Now().UTC()
	sourceIP := realIP(r)

	if _, err := h.queries.CreateCronPing(r.Context(), db.CreateCronPingParams{
		MonitorID: monitor.ID,
		SourceIp:  sourceIP,
	}); err != nil {
		// Log but don't fail — the job must always get 200.
		w.WriteHeader(http.StatusOK)
		return
	}

	// Don't update next_ping_at for paused monitors; the worker ignores them anyway.
	if monitor.Status != db.MonitorStatusPaused {
		nextPing := computeNextPing(monitor.Schedule, now, int(monitor.GracePeriodMins))
		_, _ = h.queries.UpdateCronMonitorPing(r.Context(), db.UpdateCronMonitorPingParams{
			ID:         monitor.ID,
			OrgID:      monitor.OrgID,
			LastPingAt: pgtype.Timestamptz{Time: now, Valid: true},
			NextPingAt: pgtype.Timestamptz{Time: nextPing, Valid: true},
		})
	}

	w.WriteHeader(http.StatusOK)
}

// computeNextPing returns when the next ping is expected, after which the grace
// period starts. Formula: next scheduled time after now + grace buffer.
func computeNextPing(schedule string, now time.Time, graceMins int) time.Time {
	s := strings.ToLower(strings.TrimSpace(schedule))

	if strings.HasPrefix(s, "every ") {
		d := parseEveryDuration(strings.TrimPrefix(s, "every "))
		if d > 0 {
			// next expected ping = now + interval + grace
			return now.Add(d + time.Duration(graceMins)*time.Minute)
		}
	}

	// Cron expression — use robfig/cron to find the next fire time.
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(schedule)
	if err == nil {
		next := sched.Next(now)
		return next.Add(time.Duration(graceMins) * time.Minute)
	}

	// Fallback: 1 hour from now.
	return now.Add(time.Hour + time.Duration(graceMins)*time.Minute)
}

// parseEveryDuration parses "1h", "30m", "1d", "1w" etc.
func parseEveryDuration(s string) time.Duration {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return 0
	}
	unit := s[len(s)-1]
	n, err := strconv.Atoi(s[:len(s)-1])
	if err != nil || n <= 0 {
		return 0
	}
	switch unit {
	case 'm':
		return time.Duration(n) * time.Minute
	case 'h':
		return time.Duration(n) * time.Hour
	case 'd':
		return time.Duration(n) * 24 * time.Hour
	case 'w':
		return time.Duration(n) * 7 * 24 * time.Hour
	}
	return 0
}

// realIP extracts the client IP, respecting common proxy headers.
func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if i := strings.IndexByte(fwd, ','); i >= 0 {
			return strings.TrimSpace(fwd[:i])
		}
		return strings.TrimSpace(fwd)
	}
	addr := r.RemoteAddr
	if i := strings.LastIndexByte(addr, ':'); i >= 0 {
		return addr[:i]
	}
	return addr
}

