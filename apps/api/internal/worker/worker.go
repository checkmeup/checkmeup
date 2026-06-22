package worker

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

// Run starts the background worker loops. Returns when ctx is cancelled.
//   - Every 30 s: missed-ping detection + uptime HTTP checks
//   - Every 24 h: ping retention cleanup (ADR-015)
func Run(ctx context.Context, queries *db.Queries, tg *telegram.Client, mailer *email.Sender, wh *webhook.Client, logger *slog.Logger) {
	ticker := time.NewTicker(30 * time.Second)
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	defer cleanupTicker.Stop()
	logger.Info("worker started")
	for {
		select {
		case <-ctx.Done():
			logger.Info("worker stopped")
			return
		case <-ticker.C:
			checkOverdue(ctx, queries, tg, mailer, wh, logger)
			checkUptimeMonitors(ctx, queries, tg, mailer, wh, logger)
			checkSSLMonitors(ctx, queries, tg, mailer, wh, logger)
		case <-cleanupTicker.C:
			pruneOldPings(ctx, queries, logger)
		}
	}
}

// AlertMessage carries the per-channel content for a single alert event.
// Exported (along with MonitorRef and DispatchAlert) so handler/ping.go's
// cron-recovery path can route through the same channel dispatch as the
// worker's own down/recovery events, instead of duplicating it.
type AlertMessage struct {
	Telegram     string
	EmailSubject string
	EmailHTML    string
	// Webhook is nil for call sites that haven't been updated to build a
	// structured event — sendToChannel skips the webhook channel type in
	// that case rather than sending an empty payload.
	Webhook *webhook.Event
}

// MonitorRef identifies the monitor an alert is for. Bundled into one value
// (rather than two separate params) to keep DispatchAlert's parameter count
// down — every call site already has type+ID as a pair, matching the
// polymorphic monitor_type/monitor_id pattern used at the DB layer (ADR-023).
type MonitorRef struct {
	Type string
	ID   uuid.UUID
}

// DispatchAlert sends msg to every channel attached to the monitor, logging
// per-channel failures without aborting the others. If the monitor has no
// attached, enabled channels, it falls back to emailing every user in the
// org instead of staying silent (ADR-023) — a monitor always has somewhere
// to send an alert. Reports whether the alert was delivered at all — per
// ADR-016 / US-2805, the per-incident alert cap counts one notification
// event regardless of how many channels (or the fallback) it went out on.
func DispatchAlert(ctx context.Context, queries *db.Queries, tg *telegram.Client, mailer *email.Sender, wh *webhook.Client, orgID uuid.UUID, mon MonitorRef, msg AlertMessage, logger *slog.Logger) bool {
	channels, err := queries.ListMonitorNotificationChannels(ctx, db.ListMonitorNotificationChannelsParams{
		MonitorType: mon.Type, MonitorID: mon.ID,
	})
	if err != nil {
		logger.Error("worker: list monitor notification channels", "monitor_id", mon.ID, "err", err)
		return false
	}

	if len(channels) == 0 {
		return dispatchFallbackEmail(ctx, queries, mailer, orgID, msg, logger, mon.ID)
	}

	sent := false
	for _, c := range channels {
		if sendToChannel(ctx, queries, c, msg, tg, mailer, wh, logger, mon.ID) {
			sent = true
		}
	}
	return sent
}

// sendToChannel delivers msg via a single channel's provider, logging (not
// returning) a failure so DispatchAlert can keep trying the rest.
func sendToChannel(ctx context.Context, queries *db.Queries, c db.NotificationChannel, msg AlertMessage, tg *telegram.Client, mailer *email.Sender, wh *webhook.Client, logger *slog.Logger, monitorID uuid.UUID) bool {
	switch c.Type {
	case db.NotificationChannelTypeTelegram:
		chatID := channelConfigValue(c.Config, "chatId")
		if chatID == "" {
			return false
		}
		if err := tg.SendMessage(chatID, msg.Telegram); err != nil {
			logger.Error("worker: send telegram alert", "monitor_id", monitorID, "channel_id", c.ID, "err", err)
			return false
		}
		return true
	case db.NotificationChannelTypeEmail:
		addr := channelConfigValue(c.Config, "email")
		if addr == "" {
			return false
		}
		if err := mailer.SendAlertEmail(addr, msg.EmailSubject, msg.EmailHTML); err != nil {
			logger.Error("worker: send email alert", "monitor_id", monitorID, "channel_id", c.ID, "err", err)
			return false
		}
		return true
	case db.NotificationChannelTypeWebhook:
		return sendWebhookAlert(ctx, queries, c, msg, wh, logger, monitorID)
	}
	return false
}

// sendWebhookAlert POSTs msg.Webhook to the channel's configured URL and
// records the delivery outcome on the channel row (US-1404) regardless of
// success — "Last delivery: failed, 500, 2 min ago" needs a result even
// when nothing else attached to the monitor succeeded either.
func sendWebhookAlert(ctx context.Context, queries *db.Queries, c db.NotificationChannel, msg AlertMessage, wh *webhook.Client, logger *slog.Logger, monitorID uuid.UUID) bool {
	if msg.Webhook == nil {
		return false
	}
	url := channelConfigValue(c.Config, "url")
	if url == "" {
		return false
	}
	secret := channelConfigValue(c.Config, "secret")

	statusCode, sendErr := wh.Send(url, secret, *msg.Webhook)

	status, detail := "success", fmt.Sprintf("%d", statusCode)
	if sendErr != nil {
		status = "failed"
		if statusCode == 0 {
			detail = "timeout / connection error"
		}
	}
	if err := queries.UpdateNotificationChannelDelivery(ctx, db.UpdateNotificationChannelDeliveryParams{
		ID:                 c.ID,
		LastDeliveryStatus: pgtype.Text{String: status, Valid: true},
		LastDeliveryDetail: pgtype.Text{String: detail, Valid: true},
	}); err != nil {
		logger.Error("worker: record webhook delivery status", "channel_id", c.ID, "err", err)
	}

	if sendErr != nil {
		logger.Error("worker: send webhook alert", "monitor_id", monitorID, "channel_id", c.ID, "status_code", statusCode, "err", sendErr)
		return false
	}
	return true
}

// dispatchFallbackEmail emails every user in the org when a monitor has no
// attached channels. Every org has exactly one user today (EP-12 team
// invites aren't built), but this is written against org_id rather than a
// single user so it doesn't need touching when that ships.
func dispatchFallbackEmail(ctx context.Context, queries *db.Queries, mailer *email.Sender, orgID uuid.UUID, msg AlertMessage, logger *slog.Logger, monitorID uuid.UUID) bool {
	emails, err := queries.ListOrgUserEmails(ctx, orgID)
	if err != nil {
		logger.Error("worker: list org user emails", "org_id", orgID, "err", err)
		return false
	}
	sent := false
	for _, addr := range emails {
		if err := mailer.SendAlertEmail(addr, msg.EmailSubject, msg.EmailHTML); err != nil {
			logger.Error("worker: send fallback alert email", "monitor_id", monitorID, "err", err)
			continue
		}
		sent = true
	}
	return sent
}

// channelConfigValue reads a single string field out of a channel's config
// JSONB blob — e.g. "chatId" for telegram, "email" for email.
func channelConfigValue(raw []byte, key string) string {
	var cfg map[string]string
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return ""
	}
	return strings.TrimSpace(cfg[key])
}

// FormatDuration renders d as "1h 2m" / "2m 3s" / "3s" for downtime
// durations in alert messages. Exported so handler/ping.go's cron-recovery
// path (which needs the same formatting for its webhook event) doesn't
// duplicate it.
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func pruneOldPings(ctx context.Context, queries *db.Queries, logger *slog.Logger) {
	if err := queries.DeleteOldCronPings(ctx); err != nil {
		logger.Error("worker: prune old pings", "err", err)
	} else {
		logger.Info("worker: pruned cron_pings older than 30 days")
	}
}

// ─── cron monitor checks ─────────────────────────────────────────────────────

func checkOverdue(ctx context.Context, queries *db.Queries, tg *telegram.Client, mailer *email.Sender, wh *webhook.Client, logger *slog.Logger) {
	monitors, err := queries.ListOverdueCronMonitors(ctx)
	if err != nil {
		logger.Error("worker: list overdue monitors", "err", err)
		return
	}

	for _, m := range monitors {
		if err := queries.UpdateCronMonitorDown(ctx, m.ID); err != nil {
			logger.Error("worker: mark down", "monitor_id", m.ID, "err", err)
			continue
		}

		inc, err := queries.CreateCronIncident(ctx, m.ID)
		if err != nil {
			logger.Error("worker: create incident", "monitor_id", m.ID, "err", err)
			continue
		}

		if !m.AlertsEnabled {
			continue
		}

		max := m.MaxAlertsPerIncident
		if max > 0 && inc.AlertCount >= max {
			continue
		}

		missedBy := time.Since(m.NextPingAt.Time).Round(time.Second)
		expectedAt := m.NextPingAt.Time.Format("15:04:05 MST")
		msg := AlertMessage{
			Telegram: fmt.Sprintf(
				"🔴 <b>%s</b> missed its ping\n\nSchedule: <code>%s</code>\nExpected at: %s\nMissed by: %s",
				m.Name, m.Schedule, expectedAt, missedBy,
			),
			EmailSubject: fmt.Sprintf("DOWN: %s missed its ping", m.Name),
			EmailHTML: fmt.Sprintf(
				"<p>🔴 <b>%s</b> missed its ping</p><p>Schedule: <code>%s</code><br>Expected at: %s<br>Missed by: %s</p>",
				m.Name, m.Schedule, expectedAt, missedBy,
			),
			Webhook: &webhook.Event{
				EventType:   "down",
				MonitorName: m.Name,
				MonitorType: "cron",
				Reason:      fmt.Sprintf("missed its ping — expected at %s, missed by %s", expectedAt, missedBy),
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
			},
		}
		if !DispatchAlert(ctx, queries, tg, mailer, wh, m.OrgID, MonitorRef{Type: "cron", ID: m.ID}, msg, logger) {
			continue
		}
		if _, err := queries.IncrementCronIncidentAlertCount(ctx, inc.ID); err != nil {
			logger.Error("worker: increment alert count", "incident_id", inc.ID, "err", err)
		}
	}
}

// ─── uptime monitor checks ───────────────────────────────────────────────────

func checkUptimeMonitors(ctx context.Context, queries *db.Queries, tg *telegram.Client, mailer *email.Sender, wh *webhook.Client, logger *slog.Logger) {
	monitors, err := queries.ListDueUptimeMonitors(ctx)
	if err != nil {
		logger.Error("uptime worker: list due monitors", "err", err)
		return
	}

	var wg sync.WaitGroup
	for _, m := range monitors {
		m := m
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkOneUptimeMonitor(ctx, queries, tg, mailer, wh, logger, m)
		}()
	}
	wg.Wait()
}

func checkOneUptimeMonitor(ctx context.Context, queries *db.Queries, tg *telegram.Client, mailer *email.Sender, wh *webhook.Client, logger *slog.Logger, m db.UptimeMonitor) {
	statusCode, responseTimeMs, isUp, failureReason := performHTTPCheck(m)

	var codeParam pgtype.Int4
	if statusCode > 0 {
		codeParam = pgtype.Int4{Int32: int32(statusCode), Valid: true}
	}

	if _, err := queries.CreateUptimeCheck(ctx, db.CreateUptimeCheckParams{
		MonitorID:      m.ID,
		StatusCode:     codeParam,
		ResponseTimeMs: int32(responseTimeMs),
		IsUp:           isUp,
		FailureReason:  pgtype.Text{String: failureReason, Valid: failureReason != ""},
	}); err != nil {
		logger.Error("uptime worker: create check", "monitor_id", m.ID, "err", err)
		return
	}

	prevStatus := m.Status

	if isUp {
		if _, err := queries.RecordUptimeCheckUp(ctx, m.ID); err != nil {
			logger.Error("uptime worker: record up", "monitor_id", m.ID, "err", err)
			return
		}
		if prevStatus == db.MonitorStatusDown {
			inc, err := queries.ResolveLatestUptimeIncident(ctx, m.ID)
			if err != nil {
				logger.Error("uptime worker: resolve incident", "monitor_id", m.ID, "err", err)
			}
			if m.AlertsEnabled && err == nil {
				downtime := FormatDuration(time.Since(inc.StartedAt.Time))
				msg := AlertMessage{
					Telegram:     fmt.Sprintf("✅ <b>%s</b> is back up\n\nURL: <code>%s</code>", m.Name, m.Url),
					EmailSubject: fmt.Sprintf("%s recovered", m.Name),
					EmailHTML:    fmt.Sprintf("<p>✅ <b>%s</b> is back up</p><p>URL: <code>%s</code></p>", m.Name, m.Url),
					Webhook: &webhook.Event{
						EventType:        "recovery",
						MonitorName:      m.Name,
						MonitorType:      "uptime",
						DowntimeDuration: downtime,
						Timestamp:        time.Now().UTC().Format(time.RFC3339),
					},
				}
				DispatchAlert(ctx, queries, tg, mailer, wh, m.OrgID, MonitorRef{Type: "uptime", ID: m.ID}, msg, logger)
			}
		}
		return
	}

	// Check failed
	updated, err := queries.RecordUptimeCheckFailure(ctx, m.ID)
	if err != nil {
		logger.Error("uptime worker: record failure", "monitor_id", m.ID, "err", err)
		return
	}

	if updated.ConsecutiveFailures >= 2 && prevStatus != db.MonitorStatusDown {
		if err := queries.MarkUptimeMonitorDown(ctx, m.ID); err != nil {
			logger.Error("uptime worker: mark down", "monitor_id", m.ID, "err", err)
			return
		}
		inc, err := queries.CreateUptimeIncident(ctx, m.ID)
		if err != nil {
			logger.Error("uptime worker: create incident", "monitor_id", m.ID, "err", err)
			return
		}
		if !m.AlertsEnabled {
			return
		}
		max := m.MaxAlertsPerIncident
		if max > 0 && inc.AlertCount >= max {
			return
		}
		msg := AlertMessage{
			Telegram: fmt.Sprintf("🔴 <b>%s</b> is down\n\nURL: <code>%s</code>\nReason: %s",
				m.Name, m.Url, failureReason),
			EmailSubject: fmt.Sprintf("DOWN: %s", m.Name),
			EmailHTML: fmt.Sprintf("<p>🔴 <b>%s</b> is down</p><p>URL: <code>%s</code><br>Reason: %s</p>",
				m.Name, m.Url, failureReason),
			Webhook: &webhook.Event{
				EventType:   "down",
				MonitorName: m.Name,
				MonitorType: "uptime",
				Reason:      failureReason,
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
			},
		}
		if !DispatchAlert(ctx, queries, tg, mailer, wh, m.OrgID, MonitorRef{Type: "uptime", ID: m.ID}, msg, logger) {
			return
		}
		if _, err := queries.IncrementUptimeIncidentAlertCount(ctx, inc.ID); err != nil {
			logger.Error("uptime worker: increment alert count", "incident_id", inc.ID, "err", err)
		}
	}
}

// maxKeywordCheckBytes caps how much of the response body is read for a
// keyword search, regardless of Content-Length (US-1102).
const maxKeywordCheckBytes = 512 * 1024

// performHTTPCheck runs the monitor's HTTP check and, if a keyword is
// configured, searches the (capped) response body for it as part of the
// same request. failureReason is empty on success, or describes which
// check failed (HTTP status vs. keyword mismatch) for alerting/display.
func performHTTPCheck(m db.UptimeMonitor) (statusCode int, responseTimeMs int64, isUp bool, failureReason string) {
	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()
	resp, err := client.Get(m.Url)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return 0, elapsed, false, "timeout / connection error"
	}
	defer func() { _ = resp.Body.Close() }()

	keyword := strings.TrimSpace(m.Keyword.String)
	if keyword == "" {
		_, _ = io.Copy(io.Discard, resp.Body)
		if resp.StatusCode != http.StatusOK {
			return resp.StatusCode, elapsed, false, httpStatusDesc(resp.StatusCode)
		}
		return resp.StatusCode, elapsed, true, ""
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxKeywordCheckBytes))

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, elapsed, false, httpStatusDesc(resp.StatusCode)
	}
	if !keywordCheckPasses(string(body), keyword, m.KeywordMode, m.KeywordCaseSensitive) {
		return resp.StatusCode, elapsed, false, keywordFailureReason(m.KeywordMode)
	}
	return resp.StatusCode, elapsed, true, ""
}

// keywordCheckPasses reports whether body satisfies the monitor's keyword
// mode. Plain substring search only — no regex, to avoid a ReDoS surface.
func keywordCheckPasses(body, keyword string, mode db.KeywordMode, caseSensitive bool) bool {
	if !caseSensitive {
		body = strings.ToLower(body)
		keyword = strings.ToLower(keyword)
	}
	contains := strings.Contains(body, keyword)
	if mode == db.KeywordModeNotContains {
		return !contains
	}
	return contains
}

func keywordFailureReason(mode db.KeywordMode) string {
	if mode == db.KeywordModeNotContains {
		return "Keyword found"
	}
	return "Keyword not found"
}

func httpStatusDesc(code int) string {
	if code == 0 {
		return "timeout / connection error"
	}
	return fmt.Sprintf("HTTP %d", code)
}

// ─── SSL monitor checks ──────────────────────────────────────────────────────

func checkSSLMonitors(ctx context.Context, queries *db.Queries, tg *telegram.Client, mailer *email.Sender, wh *webhook.Client, logger *slog.Logger) {
	monitors, err := queries.ListDueSSLMonitors(ctx)
	if err != nil {
		logger.Error("ssl worker: list due monitors", "err", err)
		return
	}

	var wg sync.WaitGroup
	for _, m := range monitors {
		m := m
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkOneSSLMonitor(ctx, queries, tg, mailer, wh, logger, m)
		}()
	}
	wg.Wait()
}

func checkOneSSLMonitor(ctx context.Context, queries *db.Queries, tg *telegram.Client, mailer *email.Sender, wh *webhook.Client, logger *slog.Logger, m db.SslMonitor) {
	expiresAt, issuer, daysLeft, err := performTLSCheck(m.Hostname)

	var newStatus db.SslMonitorStatus
	var expiresAtParam pgtype.Timestamptz
	var issuerParam pgtype.Text
	var errorMsgParam pgtype.Text
	alerted30d, alerted14d, alerted7d := m.Alerted30d, m.Alerted14d, m.Alerted7d

	if err != nil {
		newStatus = db.SslMonitorStatusError
		errorMsgParam = pgtype.Text{String: err.Error(), Valid: true}
	} else {
		expiresAtParam = pgtype.Timestamptz{Time: expiresAt, Valid: true}
		issuerParam = pgtype.Text{String: issuer, Valid: true}

		switch {
		case daysLeft < 0:
			newStatus = db.SslMonitorStatusExpired
		case daysLeft <= 30:
			newStatus = db.SslMonitorStatusExpiringSoon
		default:
			newStatus = db.SslMonitorStatusUp
			// Reset flags when cert is renewed
			alerted30d, alerted14d, alerted7d = false, false, false
		}
	}

	// Send threshold alerts (one per crossing)
	if m.AlertsEnabled && (newStatus == db.SslMonitorStatusExpiringSoon || newStatus == db.SslMonitorStatusExpired) {
		var subject, telegramMsg, emailHTML string
		expiresStr := expiresAt.Format("2006-01-02")
		switch {
		case daysLeft < 0 && !alerted7d:
			subject = fmt.Sprintf("DOWN: %s SSL certificate expired", m.Name)
			telegramMsg = fmt.Sprintf("🔴 <b>%s</b> SSL certificate has <b>expired</b>\n\nHost: <code>%s</code>\nExpired: %s",
				m.Name, m.Hostname, expiresStr)
			emailHTML = fmt.Sprintf("<p>🔴 <b>%s</b> SSL certificate has <b>expired</b></p><p>Host: <code>%s</code><br>Expired: %s</p>",
				m.Name, m.Hostname, expiresStr)
			alerted7d = true
		case daysLeft <= 7 && !alerted7d:
			subject = fmt.Sprintf("%s SSL certificate expires in %d days", m.Name, daysLeft)
			telegramMsg = fmt.Sprintf("🔐 <b>%s</b> SSL certificate expires in <b>%d days</b>\n\nHost: <code>%s</code>\nExpires: %s",
				m.Name, daysLeft, m.Hostname, expiresStr)
			emailHTML = fmt.Sprintf("<p>🔐 <b>%s</b> SSL certificate expires in <b>%d days</b></p><p>Host: <code>%s</code><br>Expires: %s</p>",
				m.Name, daysLeft, m.Hostname, expiresStr)
			alerted7d = true
		case daysLeft <= 14 && !alerted14d:
			subject = fmt.Sprintf("%s SSL certificate expires in %d days", m.Name, daysLeft)
			telegramMsg = fmt.Sprintf("🔐 <b>%s</b> SSL certificate expires in <b>%d days</b>\n\nHost: <code>%s</code>\nExpires: %s",
				m.Name, daysLeft, m.Hostname, expiresStr)
			emailHTML = fmt.Sprintf("<p>🔐 <b>%s</b> SSL certificate expires in <b>%d days</b></p><p>Host: <code>%s</code><br>Expires: %s</p>",
				m.Name, daysLeft, m.Hostname, expiresStr)
			alerted14d = true
		case daysLeft <= 30 && !alerted30d:
			subject = fmt.Sprintf("%s SSL certificate expires in %d days", m.Name, daysLeft)
			telegramMsg = fmt.Sprintf("🔐 <b>%s</b> SSL certificate expires in <b>%d days</b>\n\nHost: <code>%s</code>\nExpires: %s",
				m.Name, daysLeft, m.Hostname, expiresStr)
			emailHTML = fmt.Sprintf("<p>🔐 <b>%s</b> SSL certificate expires in <b>%d days</b></p><p>Host: <code>%s</code><br>Expires: %s</p>",
				m.Name, daysLeft, m.Hostname, expiresStr)
			alerted30d = true
		}
		if telegramMsg != "" {
			DispatchAlert(ctx, queries, tg, mailer, wh, m.OrgID, MonitorRef{Type: "ssl", ID: m.ID}, AlertMessage{
				Telegram:     telegramMsg,
				EmailSubject: subject,
				EmailHTML:    emailHTML,
				Webhook: &webhook.Event{
					EventType:   "down",
					MonitorName: m.Name,
					MonitorType: "ssl",
					Reason:      subject,
					Timestamp:   time.Now().UTC().Format(time.RFC3339),
				},
			}, logger)
		}
	}

	if _, err := queries.UpdateSSLMonitorCheck(ctx, db.UpdateSSLMonitorCheckParams{
		ID:         m.ID,
		Status:     newStatus,
		ExpiresAt:  expiresAtParam,
		Issuer:     issuerParam,
		ErrorMsg:   errorMsgParam,
		Alerted30d: alerted30d,
		Alerted14d: alerted14d,
		Alerted7d:  alerted7d,
	}); err != nil {
		logger.Error("ssl worker: update check", "monitor_id", m.ID, "err", err)
	}
}

func performTLSCheck(hostname string) (expiresAt time.Time, issuer string, daysLeft int, err error) {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", hostname+":443", &tls.Config{
		ServerName: hostname,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return time.Time{}, "", 0, err
	}
	defer conn.Close() //nolint:errcheck // TLS close errors are not actionable

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return time.Time{}, "", 0, fmt.Errorf("no peer certificates")
	}

	leaf := certs[0]
	expiresAt = leaf.NotAfter
	issuer = leaf.Issuer.CommonName
	daysLeft = int(time.Until(expiresAt).Hours() / 24)
	return expiresAt, issuer, daysLeft, nil
}
