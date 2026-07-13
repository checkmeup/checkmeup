package handler

import (
	"context"
	"encoding/json"
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
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/twilio"
	"github.com/checkmeup/checkmeup/internal/webhook"
	"github.com/checkmeup/checkmeup/internal/worker"
)

// Caps on ping query-string metadata (e.g. ?build=142&state=success from a
// CI job) so a chatty caller can't bloat cron_pings storage — see ADR-015's
// retention math, which assumed near-empty rows.
const (
	maxPingMetadataPairs    = 20
	maxPingMetadataKeyLen   = 64
	maxPingMetadataValueLen = 256
)

type PingHandler struct {
	queries *db.Queries
	tg      *telegram.Client
	mailer  *email.Sender
	wh      *webhook.Client
	sl      *slack.Client
	sm      *twilio.Client
}

func NewPingHandler(pool *pgxpool.Pool, tg *telegram.Client, mailer *email.Sender, wh *webhook.Client, sl *slack.Client, sm *twilio.Client) *PingHandler {
	return &PingHandler{queries: db.New(pool), tg: tg, mailer: mailer, wh: wh, sl: sl, sm: sm}
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
		Metadata:  buildPingMetadata(r),
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
	if wasDown {
		h.sendCronRecoveryAlert(r.Context(), monitor, now)
	}

	w.WriteHeader(http.StatusOK)
}

// sendCronRecoveryAlert resolves the monitor's latest open incident and, if
// alerts are enabled, dispatches a recovery notification. Routed through
// worker.DispatchAlert (not org.TelegramChatID/AlertEmail directly) so cron
// recovery alerts respect the monitor's attached notification_channels —
// including webhook (EP-14) — same as every other alert path; this used to
// be a special case still wired to the pre-EP-28 org-level fields.
func (h *PingHandler) sendCronRecoveryAlert(ctx context.Context, monitor db.CronMonitor, now time.Time) {
	inc, err := h.queries.ResolveLatestCronIncident(ctx, monitor.ID)
	if err != nil || !monitor.AlertsEnabled {
		return
	}
	downtime := worker.FormatDuration(now.Sub(inc.StartedAt.Time))
	slackRecovery := slack.RecoveryMessage(monitor.Name, "cron", downtime)
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
		Slack: &slackRecovery,
		SMS:   worker.TruncateSMS(fmt.Sprintf("Checkmeup: %s recovered after %s", monitor.Name, downtime)),
	}
	n := worker.Notifiers{Queries: h.queries, Telegram: h.tg, Mailer: h.mailer, Webhook: h.wh, Slack: h.sl, SMS: h.sm, Logger: slog.Default()}
	worker.DispatchAlert(ctx, n, monitor.OrgID, worker.MonitorRef{Type: "cron", ID: monitor.ID}, msg)
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
	s = normalizeDurationSuffix(strings.TrimSpace(strings.ToLower(s)))
	if len(s) < 2 {
		return 0
	}
	unit := s[len(s)-1]
	n, err := strconv.Atoi(strings.TrimSpace(s[:len(s)-1]))
	if err != nil || n <= 0 {
		return 0
	}
	return time.Duration(n) * durationUnit(unit)
}

// normalizeDurationSuffix rewrites a word suffix ("minutes", "hour", "days",
// "week", ...) to the single-char unit parseEveryDuration expects.
func normalizeDurationSuffix(s string) string {
	for _, r := range []struct{ old, new string }{
		{"minutes", "m"}, {"minute", "m"}, {"mins", "m"}, {"min", "m"},
		{"hours", "h"}, {"hour", "h"}, {"hrs", "h"}, {"hr", "h"},
		{"weeks", "w"}, {"week", "w"},
		{"days", "d"}, {"day", "d"},
	} {
		if strings.HasSuffix(s, r.old) {
			return strings.TrimSpace(s[:len(s)-len(r.old)]) + r.new
		}
	}
	return s
}

// durationUnit maps a single-char unit to its time.Duration multiplier, or
// 0 for an unrecognized unit (parseEveryDuration treats that as invalid).
func durationUnit(unit byte) time.Duration {
	switch unit {
	case 'm':
		return time.Minute
	case 'h':
		return time.Hour
	case 'd':
		return 24 * time.Hour
	case 'w':
		return 7 * 24 * time.Hour
	}
	return 0
}

// buildPingMetadata captures the ping's query-string params (e.g. a CI job
// reporting ?build=142&state=success) as JSON, capped so one chatty caller
// can't bloat storage. Returns nil (no metadata column write) when the ping
// carries no query params. Never errors — an oversized or malformed param
// set is truncated, not rejected, since a ping must always succeed.
func buildPingMetadata(r *http.Request) []byte {
	query := r.URL.Query()
	if len(query) == 0 {
		return nil
	}

	meta := make(map[string]string, len(query))
	for key, values := range query {
		if len(meta) >= maxPingMetadataPairs {
			break
		}
		if len(key) > maxPingMetadataKeyLen {
			key = key[:maxPingMetadataKeyLen]
		}
		value := values[0]
		if len(value) > maxPingMetadataValueLen {
			value = value[:maxPingMetadataValueLen]
		}
		meta[key] = value
	}

	b, err := json.Marshal(meta)
	if err != nil {
		return nil
	}
	return b
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
