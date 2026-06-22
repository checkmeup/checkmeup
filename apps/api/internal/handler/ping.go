package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/webhook"
	"github.com/checkmeup/checkmeup/internal/worker"
)

type PingHandler struct {
	queries *db.Queries
	tg      *telegram.Client
	mailer  *email.Sender
	wh      *webhook.Client
}

func NewPingHandler(pool *pgxpool.Pool, tg *telegram.Client, mailer *email.Sender, wh *webhook.Client) *PingHandler {
	return &PingHandler{queries: db.New(pool), tg: tg, mailer: mailer, wh: wh}
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

	wasDown := monitor.Status == db.MonitorStatusDown

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

	// Recovery: monitor was down and just checked in again. The incident is
	// resolved regardless of AlertsEnabled — only the alert send itself is
	// gated by that setting (matches the uptime-monitor worker's pattern).
	// Routed through worker.DispatchAlert (not org.TelegramChatID/AlertEmail
	// directly) so cron recovery alerts respect the monitor's attached
	// notification_channels — including webhook (EP-14) — same as every
	// other alert path; this used to be a special case still wired to the
	// pre-EP-28 org-level fields.
	if wasDown {
		inc, err := h.queries.ResolveLatestCronIncident(r.Context(), monitor.ID)
		if err == nil && monitor.AlertsEnabled {
			downtime := worker.FormatDuration(now.Sub(inc.StartedAt.Time))
			msg := worker.AlertMessage{
				Telegram:     fmt.Sprintf("✅ <b>%s</b> recovered\n\nDown for: %s", monitor.Name, downtime),
				EmailSubject: fmt.Sprintf("%s recovered", monitor.Name),
				EmailHTML:    fmt.Sprintf("<p>✅ <b>%s</b> recovered</p><p>Down for: %s</p>", monitor.Name, downtime),
				Webhook: &webhook.Event{
					EventType:        "recovery",
					MonitorName:      monitor.Name,
					MonitorType:      "cron",
					DowntimeDuration: downtime,
					Timestamp:        now.UTC().Format(time.RFC3339),
				},
			}
			n := worker.Notifiers{Queries: h.queries, Telegram: h.tg, Mailer: h.mailer, Webhook: h.wh, Logger: slog.Default()}
			worker.DispatchAlert(r.Context(), n, monitor.OrgID, worker.MonitorRef{Type: "cron", ID: monitor.ID}, msg)
		}
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

// parseEveryDuration parses "1h", "1hr", "1hour", "30m", "30min", "30mins",
// "1d", "1day", "1w", "1week" etc.
func parseEveryDuration(s string) time.Duration {
	s = strings.TrimSpace(strings.ToLower(s))
	// Normalise word suffixes to single-char unit before parsing.
	for _, r := range []struct{ old, new string }{
		{"minutes", "m"}, {"minute", "m"}, {"mins", "m"}, {"min", "m"},
		{"hours", "h"}, {"hour", "h"}, {"hrs", "h"}, {"hr", "h"},
		{"weeks", "w"}, {"week", "w"},
		{"days", "d"}, {"day", "d"},
	} {
		if strings.HasSuffix(s, r.old) {
			s = strings.TrimSpace(s[:len(s)-len(r.old)]) + r.new
			break
		}
	}
	if len(s) < 2 {
		return 0
	}
	unit := s[len(s)-1]
	n, err := strconv.Atoi(strings.TrimSpace(s[:len(s)-1]))
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
