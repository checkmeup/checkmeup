package worker

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/billing"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/rdap"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/twilio"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

// Maximum concurrent outbound checks per tick across uptime, SSL, and domain loops.
const checkConcurrency = 50

// Notifiers bundles the dependencies every alert-dispatch and monitor-check
// path needs. Passed as one value (rather than six separate params) to keep
// these functions' parameter counts down — every call site already needs
// the full set together (Lizard_parameter-count-medium).
type Notifiers struct {
	Queries  *db.Queries
	Telegram *telegram.Client
	Mailer   *email.Sender
	Webhook  *webhook.Client
	Slack    *slack.Client
	SMS      *twilio.Client
	RDAP     *rdap.Client
	Logger   *slog.Logger
}

// Run starts the background worker loops. Returns when ctx is cancelled.
//   - Every 30 s: missed-ping detection + uptime HTTP checks
//   - Every 24 h: ping retention cleanup (ADR-015)
func Run(ctx context.Context, n Notifiers) {
	ticker := time.NewTicker(30 * time.Second)
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	defer cleanupTicker.Stop()
	n.Logger.Info("worker started")
	for {
		select {
		case <-ctx.Done():
			n.Logger.Info("worker stopped")
			return
		case <-ticker.C:
			checkOverdue(ctx, n)
			checkUptimeMonitors(ctx, n)
			checkSSLMonitors(ctx, n)
			checkDomainMonitors(ctx, n)
			checkPortMonitors(ctx, n)
		case <-cleanupTicker.C:
			pruneOldPings(ctx, n.Queries, n.Logger)
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
	// Slack is nil for the same reason as Webhook above.
	Slack *slack.Message
	// SMS is "" for call sites that haven't been updated — sendSMSAlert skips
	// the sms channel type in that case rather than sending an empty text.
	// Always pre-truncated to one SMS segment via TruncateSMS (US-1902/1903).
	SMS string
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
// attached, enabled channels — or every attached channel failed to deliver
// (wrong number, exhausted SMS credit once ADR-032 ships, provider outage,
// etc.) — it falls back to emailing every user in the org instead of staying
// silent (ADR-023) — a monitor always has somewhere to send an alert.
// Reports whether the alert was delivered at all — per ADR-016 / US-2805,
// the per-incident alert cap counts one notification event regardless of how
// many channels (or the fallback) it went out on.
func DispatchAlert(ctx context.Context, n Notifiers, orgID uuid.UUID, mon MonitorRef, msg AlertMessage) bool {
	channels, err := n.Queries.ListMonitorNotificationChannels(ctx, db.ListMonitorNotificationChannelsParams{
		MonitorType: mon.Type, MonitorID: mon.ID,
	})
	if err != nil {
		n.Logger.Error("worker: list monitor notification channels", "monitor_id", mon.ID, "err", err)
		return false
	}

	sent := false
	for _, c := range channels {
		if sendToChannel(ctx, n, c, msg, mon.ID) {
			sent = true
		}
	}
	if !sent {
		return dispatchFallbackEmail(ctx, n, orgID, msg, mon.ID)
	}
	return sent
}

// sendToChannel delivers msg via a single channel's provider, logging (not
// returning) a failure so DispatchAlert can keep trying the rest.
func sendToChannel(ctx context.Context, n Notifiers, c db.NotificationChannel, msg AlertMessage, monitorID uuid.UUID) bool {
	switch c.Type {
	case db.NotificationChannelTypeTelegram:
		chatID := channelConfigValue(c.Config, "chatId")
		if chatID == "" {
			return false
		}
		if err := n.Telegram.SendMessage(chatID, msg.Telegram); err != nil {
			n.Logger.Error("worker: send telegram alert", "monitor_id", monitorID, "channel_id", c.ID, "err", err)
			return false
		}
		return true
	case db.NotificationChannelTypeEmail:
		addr := channelConfigValue(c.Config, "email")
		if addr == "" {
			return false
		}
		if err := n.Mailer.SendAlertEmail(addr, msg.EmailSubject, msg.EmailHTML); err != nil {
			n.Logger.Error("worker: send email alert", "monitor_id", monitorID, "channel_id", c.ID, "err", err)
			return false
		}
		return true
	case db.NotificationChannelTypeWebhook:
		return sendWebhookAlert(ctx, n, c, msg, monitorID)
	case db.NotificationChannelTypeSlack:
		return sendSlackAlert(ctx, n, c, msg, monitorID)
	case db.NotificationChannelTypeSms:
		return sendSMSAlert(ctx, n, c, msg, monitorID)
	}
	return false
}

// sendWebhookAlert POSTs msg.Webhook to the channel's configured URL and
// records the delivery outcome on the channel row (US-1404) regardless of
// success — "Last delivery: failed, 500, 2 min ago" needs a result even
// when nothing else attached to the monitor succeeded either.
func sendWebhookAlert(ctx context.Context, n Notifiers, c db.NotificationChannel, msg AlertMessage, monitorID uuid.UUID) bool {
	if msg.Webhook == nil {
		return false
	}
	url := channelConfigValue(c.Config, "url")
	if url == "" {
		return false
	}
	secret := channelConfigValue(c.Config, "secret")

	statusCode, sendErr := n.Webhook.Send(url, secret, *msg.Webhook)

	status, detail := "success", fmt.Sprintf("%d", statusCode)
	if sendErr != nil {
		status = "failed"
		if statusCode == 0 {
			detail = "timeout / connection error"
		}
	}
	if err := n.Queries.UpdateNotificationChannelDelivery(ctx, db.UpdateNotificationChannelDeliveryParams{
		ID:                 c.ID,
		LastDeliveryStatus: pgtype.Text{String: status, Valid: true},
		LastDeliveryDetail: pgtype.Text{String: detail, Valid: true},
	}); err != nil {
		n.Logger.Error("worker: record webhook delivery status", "channel_id", c.ID, "err", err)
	}

	if sendErr != nil {
		n.Logger.Error("worker: send webhook alert", "monitor_id", monitorID, "channel_id", c.ID, "status_code", statusCode, "err", sendErr)
		return false
	}
	return true
}

// sendSlackAlert POSTs msg.Slack to the channel's configured Incoming Webhook
// URL and records the delivery outcome on the channel row (US-1704), mirroring
// sendWebhookAlert's fire-and-forget, no-retry pattern.
func sendSlackAlert(ctx context.Context, n Notifiers, c db.NotificationChannel, msg AlertMessage, monitorID uuid.UUID) bool {
	if msg.Slack == nil {
		return false
	}
	url := channelConfigValue(c.Config, "url")
	if url == "" {
		return false
	}

	statusCode, sendErr := n.Slack.Send(url, *msg.Slack)

	status, detail := "success", fmt.Sprintf("%d", statusCode)
	if sendErr != nil {
		status = "failed"
		if statusCode == 0 {
			detail = "timeout / connection error"
		}
	}
	if err := n.Queries.UpdateNotificationChannelDelivery(ctx, db.UpdateNotificationChannelDeliveryParams{
		ID:                 c.ID,
		LastDeliveryStatus: pgtype.Text{String: status, Valid: true},
		LastDeliveryDetail: pgtype.Text{String: detail, Valid: true},
	}); err != nil {
		n.Logger.Error("worker: record slack delivery status", "channel_id", c.ID, "err", err)
	}

	if sendErr != nil {
		n.Logger.Error("worker: send slack alert", "monitor_id", monitorID, "channel_id", c.ID, "status_code", statusCode, "err", sendErr)
		return false
	}
	return true
}

// consumeSMSCredit enforces the org's monthly SMS credit quota (ADR-032)
// before a send is attempted — 1 credit per send in this simplified,
// unweighted pass (no per-destination cost band yet). Returns false and
// records a delivery failure if the org has no credit left this month,
// leaving DispatchAlert's existing fallback-email path (worker.go's
// DispatchAlert, once every attached channel including this one fails to
// deliver) as the safety net, same as any other sms send failure.
func consumeSMSCredit(ctx context.Context, n Notifiers, c db.NotificationChannel) bool {
	plan, err := n.Queries.GetOrgPlan(ctx, c.OrgID)
	if err != nil {
		n.Logger.Error("worker: get org plan for sms credit check", "channel_id", c.ID, "err", err)
		return false
	}
	limit := billing.GetLimits(plan).SMSCredits

	// CreditCost is hardcoded to 1 in this pass (no per-destination cost
	// band yet — ADR-032's "Implementation note"); a future weighted-cost
	// lookup would compute this from the phone number's E.164 calling code
	// instead.
	_, err = n.Queries.ConsumeSMSCredit(ctx, db.ConsumeSMSCreditParams{ID: c.OrgID, CreditCost: 1, CreditLimit: int32(limit)})
	if err != nil {
		status, detail := "failed", "sms credit quota exhausted"
		if !errors.Is(err, pgx.ErrNoRows) {
			n.Logger.Error("worker: consume sms credit", "channel_id", c.ID, "err", err)
			detail = "internal error checking sms credit"
		}
		if updErr := n.Queries.UpdateNotificationChannelDelivery(ctx, db.UpdateNotificationChannelDeliveryParams{
			ID:                 c.ID,
			LastDeliveryStatus: pgtype.Text{String: status, Valid: true},
			LastDeliveryDetail: pgtype.Text{String: detail, Valid: true},
		}); updErr != nil {
			n.Logger.Error("worker: record sms delivery status", "channel_id", c.ID, "err", updErr)
		}
		return false
	}
	return true
}

// sendSMSAlert texts msg.SMS to the channel's configured phone number via
// Twilio and records the delivery outcome on the channel row (US-1904),
// mirroring sendWebhookAlert/sendSlackAlert's fire-and-forget, no-retry
// pattern. Opt-out compliance is handled entirely by the in-app channel
// toggle, not by anything here (ADR-029) — a bounce from a number that
// opted out via STOP on a two-way sender just comes back as an ordinary
// delivery failure below, same as a carrier rejection or invalid number.
func sendSMSAlert(ctx context.Context, n Notifiers, c db.NotificationChannel, msg AlertMessage, monitorID uuid.UUID) bool {
	if msg.SMS == "" {
		return false
	}
	phone := channelConfigValue(c.Config, "phone_number")
	if phone == "" {
		return false
	}

	if !consumeSMSCredit(ctx, n, c) {
		return false
	}

	statusCode, sendErr := n.SMS.Send(phone, msg.SMS)

	status, detail := "success", fmt.Sprintf("%d", statusCode)
	if sendErr != nil {
		status = "failed"
		if statusCode == 0 {
			detail = "timeout / connection error"
		}
	}
	if err := n.Queries.UpdateNotificationChannelDelivery(ctx, db.UpdateNotificationChannelDeliveryParams{
		ID:                 c.ID,
		LastDeliveryStatus: pgtype.Text{String: status, Valid: true},
		LastDeliveryDetail: pgtype.Text{String: detail, Valid: true},
	}); err != nil {
		n.Logger.Error("worker: record sms delivery status", "channel_id", c.ID, "err", err)
	}

	if sendErr != nil {
		n.Logger.Error("worker: send sms alert", "monitor_id", monitorID, "channel_id", c.ID, "status_code", statusCode, "err", sendErr)
		return false
	}
	return true
}

// dispatchFallbackEmail emails every user in the org when a monitor has no
// attached channels. Every org has exactly one user today (EP-12 team
// invites aren't built), but this is written against org_id rather than a
// single user so it doesn't need touching when that ships.
func dispatchFallbackEmail(ctx context.Context, n Notifiers, orgID uuid.UUID, msg AlertMessage, monitorID uuid.UUID) bool {
	emails, err := n.Queries.ListOrgUserEmails(ctx, orgID)
	if err != nil {
		n.Logger.Error("worker: list org user emails", "org_id", orgID, "err", err)
		return false
	}
	sent := false
	for _, addr := range emails {
		if err := n.Mailer.SendAlertEmail(addr, msg.EmailSubject, msg.EmailHTML); err != nil {
			n.Logger.Error("worker: send fallback alert email", "monitor_id", monitorID, "err", err)
			continue
		}
		sent = true
	}
	return sent
}

// slackMsg returns a pointer to msg so it can be stored in AlertMessage.Slack
// without requiring the caller to take the address of a slack.Message literal.
func slackMsg(msg slack.Message) *slack.Message { return &msg }

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

// smsSegmentLimit is the length of a single GSM-7 SMS segment. Messages
// longer than this are truncated (US-1902/US-1903) rather than silently
// split across extra segments, which would multiply per-message cost.
const smsSegmentLimit = 160

// TruncateSMS trims s to fit a single SMS segment, appending an ellipsis if
// anything had to be cut. Exported so handler/ping.go's cron-recovery path
// (built outside worker.go's build*Alert functions) applies the same rule.
// Truncates by rune, not byte, so a monitor name containing multi-byte
// characters can't be cut mid-character.
func TruncateSMS(s string) string {
	runes := []rune(s)
	if len(runes) <= smsSegmentLimit {
		return s
	}
	return string(runes[:smsSegmentLimit-1]) + "…"
}

func pruneOldPings(ctx context.Context, queries *db.Queries, logger *slog.Logger) {
	if err := queries.DeleteOldCronPings(ctx); err != nil {
		logger.Error("worker: prune old pings", "err", err)
	} else {
		logger.Info("worker: pruned cron_pings older than 30 days")
	}
	if err := queries.DeleteOldUptimeChecks(ctx); err != nil {
		logger.Error("worker: prune old uptime checks", "err", err)
	} else {
		logger.Info("worker: pruned uptime_checks older than 90 days")
	}
	if err := queries.DeleteOldPortChecks(ctx); err != nil {
		logger.Error("worker: prune old port checks", "err", err)
	} else {
		logger.Info("worker: pruned port_checks older than 90 days")
	}
}

// ─── cron monitor checks ─────────────────────────────────────────────────────

func checkOverdue(ctx context.Context, n Notifiers) {
	monitors, err := n.Queries.ListOverdueCronMonitors(ctx)
	if err != nil {
		n.Logger.Error("worker: list overdue monitors", "err", err)
		return
	}

	for _, m := range monitors {
		processOverdueMonitor(ctx, n, m)
	}
}

func processOverdueMonitor(ctx context.Context, n Notifiers, m db.CronMonitor) {
	updated, err := n.Queries.IncrementCronConsecutiveFailures(ctx, m.ID)
	if err != nil {
		n.Logger.Error("worker: increment consecutive failures", "monitor_id", m.ID, "err", err)
		return
	}
	// Filter: suppress alert until more than N consecutive failures observed.
	// The monitor stays 'up' while filtering so the worker re-detects it every cycle.
	if updated.ConsecutiveFailures <= m.AlertAfterNFailures {
		return
	}

	if err := n.Queries.UpdateCronMonitorDown(ctx, m.ID); err != nil {
		n.Logger.Error("worker: mark down", "monitor_id", m.ID, "err", err)
		return
	}

	inc, err := n.Queries.CreateCronIncident(ctx, m.ID)
	if err != nil {
		n.Logger.Error("worker: create incident", "monitor_id", m.ID, "err", err)
		return
	}

	if !m.AlertsEnabled {
		return
	}
	if max := m.MaxAlertsPerIncident; max > 0 && inc.AlertCount >= max {
		return
	}

	msg := buildOverdueAlert(m)
	if !DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "cron", ID: m.ID}, msg) {
		return
	}
	if _, err := n.Queries.IncrementCronIncidentAlertCount(ctx, inc.ID); err != nil {
		n.Logger.Error("worker: increment alert count", "incident_id", inc.ID, "err", err)
	}
}

func buildOverdueAlert(m db.CronMonitor) AlertMessage {
	missedBy := time.Since(m.NextPingAt.Time).Round(time.Second)
	expectedAt := m.NextPingAt.Time.Format("15:04:05 MST")
	reason := fmt.Sprintf("missed its ping — expected at %s, missed by %s", expectedAt, missedBy)
	return AlertMessage{
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
			Reason:      reason,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.DownMessage(m.Name, "cron", reason)),
		SMS:   TruncateSMS(fmt.Sprintf("checkmeup: %s is DOWN (%s)", m.Name, reason)),
	}
}

// ─── uptime monitor checks ───────────────────────────────────────────────────

func checkUptimeMonitors(ctx context.Context, n Notifiers) {
	monitors, err := n.Queries.ListDueUptimeMonitors(ctx)
	if err != nil {
		n.Logger.Error("uptime worker: list due monitors", "err", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, checkConcurrency)
	for _, m := range monitors {
		m := m
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			checkOneUptimeMonitor(ctx, n, m)
		}()
	}
	wg.Wait()
}

func checkOneUptimeMonitor(ctx context.Context, n Notifiers, m db.UptimeMonitor) {
	statusCode, responseTimeMs, isUp, failureReason := performHTTPCheck(m)
	if !recordUptimeCheck(ctx, n, m, statusCode, responseTimeMs, isUp, failureReason) {
		return
	}

	prevStatus := m.Status
	if isUp {
		handleUptimeUp(ctx, n, m, prevStatus)
		return
	}
	handleUptimeDown(ctx, n, m, prevStatus, failureReason)
}

func recordUptimeCheck(ctx context.Context, n Notifiers, m db.UptimeMonitor, statusCode int, responseTimeMs int64, isUp bool, failureReason string) bool {
	var codeParam pgtype.Int4
	if statusCode > 0 {
		codeParam = pgtype.Int4{Int32: int32(statusCode), Valid: true}
	}
	if _, err := n.Queries.CreateUptimeCheck(ctx, db.CreateUptimeCheckParams{
		MonitorID:      m.ID,
		StatusCode:     codeParam,
		ResponseTimeMs: int32(responseTimeMs),
		IsUp:           isUp,
		FailureReason:  pgtype.Text{String: failureReason, Valid: failureReason != ""},
	}); err != nil {
		n.Logger.Error("uptime worker: create check", "monitor_id", m.ID, "err", err)
		return false
	}
	return true
}

func handleUptimeUp(ctx context.Context, n Notifiers, m db.UptimeMonitor, prevStatus db.MonitorStatus) {
	if _, err := n.Queries.RecordUptimeCheckUp(ctx, m.ID); err != nil {
		n.Logger.Error("uptime worker: record up", "monitor_id", m.ID, "err", err)
		return
	}
	if prevStatus != db.MonitorStatusDown {
		return
	}
	inc, err := n.Queries.ResolveLatestUptimeIncident(ctx, m.ID)
	if err != nil {
		n.Logger.Error("uptime worker: resolve incident", "monitor_id", m.ID, "err", err)
		return
	}
	if !m.AlertsEnabled {
		return
	}
	downtime := FormatDuration(time.Since(inc.StartedAt.Time))
	DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "uptime", ID: m.ID}, buildUptimeRecoveryAlert(m, downtime))
}

func buildUptimeRecoveryAlert(m db.UptimeMonitor, downtime string) AlertMessage {
	return AlertMessage{
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
		Slack: slackMsg(slack.RecoveryMessage(m.Name, "uptime", downtime)),
		SMS:   TruncateSMS(fmt.Sprintf("checkmeup: %s recovered after %s downtime", m.Name, downtime)),
	}
}

func handleUptimeDown(ctx context.Context, n Notifiers, m db.UptimeMonitor, prevStatus db.MonitorStatus, failureReason string) {
	updated, err := n.Queries.RecordUptimeCheckFailure(ctx, m.ID)
	if err != nil {
		n.Logger.Error("uptime worker: record failure", "monitor_id", m.ID, "err", err)
		return
	}
	if updated.ConsecutiveFailures <= m.AlertAfterNFailures || prevStatus == db.MonitorStatusDown {
		return
	}
	if err := n.Queries.MarkUptimeMonitorDown(ctx, m.ID); err != nil {
		n.Logger.Error("uptime worker: mark down", "monitor_id", m.ID, "err", err)
		return
	}
	inc, err := n.Queries.CreateUptimeIncident(ctx, m.ID)
	if err != nil {
		n.Logger.Error("uptime worker: create incident", "monitor_id", m.ID, "err", err)
		return
	}
	alertUptimeIncident(ctx, n, m, inc, failureReason)
}

// alertUptimeIncident dispatches the down alert for a freshly-recorded
// incident, honoring the monitor's alerts-enabled flag and per-incident cap.
func alertUptimeIncident(ctx context.Context, n Notifiers, m db.UptimeMonitor, inc db.UptimeIncident, failureReason string) {
	if !m.AlertsEnabled {
		return
	}
	if max := m.MaxAlertsPerIncident; max > 0 && inc.AlertCount >= max {
		return
	}
	if !DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "uptime", ID: m.ID}, buildUptimeDownAlert(m, failureReason)) {
		return
	}
	if _, err := n.Queries.IncrementUptimeIncidentAlertCount(ctx, inc.ID); err != nil {
		n.Logger.Error("uptime worker: increment alert count", "incident_id", inc.ID, "err", err)
	}
}

func buildUptimeDownAlert(m db.UptimeMonitor, failureReason string) AlertMessage {
	return AlertMessage{
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
		Slack: slackMsg(slack.DownMessage(m.Name, "uptime", failureReason)),
		SMS:   TruncateSMS(fmt.Sprintf("checkmeup: %s is DOWN (%s)", m.Name, failureReason)),
	}
}

// maxKeywordCheckBytes caps how much of the response body is read for a
// keyword search or JSON assertion, regardless of Content-Length (US-1102).
const maxKeywordCheckBytes = 512 * 1024

// performHTTPCheck runs the monitor's HTTP check and evaluates all configured
// assertions in order: status code → keyword → JSON assertions → response-time
// threshold. The first failing condition is the recorded failure reason.
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
	var assertions []jsonAssertion
	if len(m.JsonAssertions) > 0 {
		_ = json.Unmarshal(m.JsonAssertions, &assertions)
	}

	needsBody := keyword != "" || len(assertions) > 0
	var body string
	if needsBody {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxKeywordCheckBytes))
		body = string(raw)
	} else {
		_, _ = io.Copy(io.Discard, resp.Body)
	}

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, elapsed, false, httpStatusDesc(resp.StatusCode)
	}

	if keyword != "" && !keywordCheckPasses(body, keyword, m.KeywordMode, m.KeywordCaseSensitive) {
		return resp.StatusCode, elapsed, false, keywordFailureReason(m.KeywordMode)
	}

	for _, a := range assertions {
		if reason, ok := evaluateJsonAssertion(body, a); !ok {
			return resp.StatusCode, elapsed, false, reason
		}
	}

	if m.MaxResponseTimeMs.Valid && elapsed > int64(m.MaxResponseTimeMs.Int32) {
		return resp.StatusCode, elapsed, false, "response time exceeded"
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

type jsonAssertion struct {
	Path       string `json:"path"`
	Comparator string `json:"comparator"`
	Expected   string `json:"expected"`
}

// evaluateJsonAssertion resolves path in the JSON body and compares to
// expected. Returns ("", true) on pass or (failureReason, false) on fail.
func evaluateJsonAssertion(body string, a jsonAssertion) (string, bool) {
	var root map[string]any
	if err := json.Unmarshal([]byte(body), &root); err != nil {
		return "response is not valid JSON", false
	}
	actual, errMsg, ok := resolveJSONPath(root, a.Path)
	if !ok {
		return errMsg, false
	}
	return applyComparator(a.Path, a.Comparator, actual, a.Expected)
}

func resolveJSONPath(root map[string]any, path string) (value, errMsg string, ok bool) {
	p := strings.TrimPrefix(path, "$.")
	p = strings.TrimPrefix(p, ".")
	var cur any = root
	for _, seg := range strings.Split(p, ".") {
		m, isMap := cur.(map[string]any)
		if !isMap {
			return "", fmt.Sprintf("JSON path %q not found", path), false
		}
		cur, ok = m[seg]
		if !ok {
			return "", fmt.Sprintf("JSON path %q not found", path), false
		}
	}
	return jsonValueToString(cur), "", true
}

func applyComparator(path, comparator, actual, expected string) (string, bool) {
	switch comparator {
	case "greater_than", "less_than":
		return applyNumericComparator(path, comparator, actual, expected)
	default:
		return applyStringComparator(path, comparator, actual, expected)
	}
}

func applyStringComparator(path, comparator, actual, expected string) (string, bool) {
	var passes bool
	switch comparator {
	case "equals":
		passes = actual == expected
	case "not_equals":
		passes = actual != expected
	default: // "contains"
		passes = strings.Contains(actual, expected)
	}
	if !passes {
		return fmt.Sprintf("JSON assertion failed: %q %s %q (got %q)", path, comparator, expected, actual), false
	}
	return "", true
}

func applyNumericComparator(path, comparator, actual, expected string) (string, bool) {
	av, ae, parseOk := parseNumericPair(actual, expected)
	var passes bool
	if parseOk {
		if comparator == "greater_than" {
			passes = av > ae
		} else {
			passes = av < ae
		}
	}
	if !passes {
		return fmt.Sprintf("JSON assertion failed: %q %s %q (got %q)", path, comparator, expected, actual), false
	}
	return "", true
}

func jsonValueToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case nil:
		return "null"
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func parseNumericPair(actual, expected string) (float64, float64, bool) {
	av, err1 := strconv.ParseFloat(actual, 64)
	ae, err2 := strconv.ParseFloat(expected, 64)
	return av, ae, err1 == nil && err2 == nil
}

// ─── SSL monitor checks ──────────────────────────────────────────────────────

func checkSSLMonitors(ctx context.Context, n Notifiers) {
	monitors, err := n.Queries.ListDueSSLMonitors(ctx)
	if err != nil {
		n.Logger.Error("ssl worker: list due monitors", "err", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, checkConcurrency)
	for _, m := range monitors {
		m := m
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			checkOneSSLMonitor(ctx, n, m)
		}()
	}
	wg.Wait()
}

func checkOneSSLMonitor(ctx context.Context, n Notifiers, m db.SslMonitor) {
	expiresAt, issuer, daysLeft, err := performTLSCheck(m.Hostname)
	result := buildSSLCheckResult(m, expiresAt, issuer, daysLeft, err)

	if result.alert != nil {
		DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "ssl", ID: m.ID}, *result.alert)
	}

	if _, err := n.Queries.UpdateSSLMonitorCheck(ctx, db.UpdateSSLMonitorCheckParams{
		ID:                  m.ID,
		Status:              result.status,
		ExpiresAt:           result.expiresAtParam,
		Issuer:              result.issuerParam,
		ErrorMsg:            result.errorMsgParam,
		Alerted30d:          result.alerted30d,
		Alerted14d:          result.alerted14d,
		Alerted7d:           result.alerted7d,
		ConsecutiveFailures: result.consecutiveFailures,
		AlertCount:          result.alertCount,
	}); err != nil {
		n.Logger.Error("ssl worker: update check", "monitor_id", m.ID, "err", err)
	}
}

// sslCheckResult is the outcome of one performTLSCheck poll: the fields to
// persist on the monitor row, plus an alert to send when a threshold was
// freshly crossed (nil otherwise).
type sslCheckResult struct {
	status              db.SslMonitorStatus
	expiresAtParam      pgtype.Timestamptz
	issuerParam         pgtype.Text
	errorMsgParam       pgtype.Text
	alerted30d          bool
	alerted14d          bool
	alerted7d           bool
	consecutiveFailures int32
	alertCount          int32
	alert               *AlertMessage
}

func buildSSLCheckResult(m db.SslMonitor, expiresAt time.Time, issuer string, daysLeft int, checkErr error) sslCheckResult {
	r := sslCheckResult{alerted30d: m.Alerted30d, alerted14d: m.Alerted14d, alerted7d: m.Alerted7d, alertCount: m.AlertCount}

	if checkErr != nil {
		r.status = db.SslMonitorStatusError
		r.errorMsgParam = pgtype.Text{String: checkErr.Error(), Valid: true}
		r.consecutiveFailures = m.ConsecutiveFailures + 1
		return r
	}

	r.expiresAtParam = pgtype.Timestamptz{Time: expiresAt, Valid: true}
	r.issuerParam = pgtype.Text{String: issuer, Valid: true}

	switch {
	case daysLeft < 0:
		r.status = db.SslMonitorStatusExpired
	case daysLeft <= 30:
		r.status = db.SslMonitorStatusExpiringSoon
	default:
		r.status = db.SslMonitorStatusUp
		// Reset flags, failure counter, and alert count when cert is renewed.
		r.alerted30d, r.alerted14d, r.alerted7d = false, false, false
		r.consecutiveFailures = 0
		r.alertCount = 0
		return r
	}

	r.consecutiveFailures = m.ConsecutiveFailures + 1

	maxAlerts := m.MaxAlertsPerIncident
	withinLimit := maxAlerts == 0 || m.AlertCount < maxAlerts
	if m.AlertsEnabled && withinLimit && r.consecutiveFailures > m.AlertAfterNFailures {
		r.alert = sslThresholdAlert(m, daysLeft, expiresAt, &r.alerted30d, &r.alerted14d, &r.alerted7d)
		if r.alert != nil {
			r.alertCount = m.AlertCount + 1
		}
	}
	return r
}

// sslThresholdAlert returns the alert for the first newly-crossed expiry
// threshold (expired, then 7/14/30 days), setting the matching alerted flag
// so checkOneSSLMonitor only sends one notification per crossing.
func sslThresholdAlert(m db.SslMonitor, daysLeft int, expiresAt time.Time, alerted30d, alerted14d, alerted7d *bool) *AlertMessage {
	if !sslCrossedThreshold(daysLeft, alerted30d, alerted14d, alerted7d) {
		return nil
	}

	expiresStr := expiresAt.Format("2006-01-02")
	var subject, telegramMsg, emailHTML string
	if daysLeft < 0 {
		subject, telegramMsg, emailHTML = sslExpiredMessages(m, expiresStr)
	} else {
		subject, telegramMsg, emailHTML = sslExpiringSoonMessages(m, daysLeft, expiresStr)
	}

	return &AlertMessage{
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
		Slack: slackMsg(slack.DownMessage(m.Name, "ssl", subject)),
		SMS:   TruncateSMS("checkmeup: " + subject),
	}
}

// sslCrossedThreshold reports whether daysLeft just crossed a new expiry
// threshold (expired/7/14/30 days) and, if so, sets the matching alerted
// flag so the crossing only fires once. Expired and the 7-day threshold
// share alerted7d since expiring is a subset of the 7-day window.
func sslCrossedThreshold(daysLeft int, alerted30d, alerted14d, alerted7d *bool) bool {
	switch {
	case daysLeft <= 7 && !*alerted7d:
		*alerted7d = true
		return true
	case daysLeft <= 14 && !*alerted14d:
		*alerted14d = true
		return true
	case daysLeft <= 30 && !*alerted30d:
		*alerted30d = true
		return true
	default:
		return false
	}
}

// sslExpiredMessages builds the alert text for a certificate that has
// already expired (daysLeft < 0), distinct from sslExpiringSoonMessages'
// "expires in N days" phrasing.
func sslExpiredMessages(m db.SslMonitor, expiresStr string) (subject, telegramMsg, emailHTML string) {
	subject = fmt.Sprintf("DOWN: %s SSL certificate expired", m.Name)
	telegramMsg = fmt.Sprintf("🔴 <b>%s</b> SSL certificate has <b>expired</b>\n\nHost: <code>%s</code>\nExpired: %s",
		m.Name, m.Hostname, expiresStr)
	emailHTML = fmt.Sprintf("<p>🔴 <b>%s</b> SSL certificate has <b>expired</b></p><p>Host: <code>%s</code><br>Expired: %s</p>",
		m.Name, m.Hostname, expiresStr)
	return subject, telegramMsg, emailHTML
}

func sslExpiringSoonMessages(m db.SslMonitor, daysLeft int, expiresStr string) (subject, telegramMsg, emailHTML string) {
	subject = fmt.Sprintf("%s SSL certificate expires in %d days", m.Name, daysLeft)
	telegramMsg = fmt.Sprintf("🔐 <b>%s</b> SSL certificate expires in <b>%d days</b>\n\nHost: <code>%s</code>\nExpires: %s",
		m.Name, daysLeft, m.Hostname, expiresStr)
	emailHTML = fmt.Sprintf("<p>🔐 <b>%s</b> SSL certificate expires in <b>%d days</b></p><p>Host: <code>%s</code><br>Expires: %s</p>",
		m.Name, daysLeft, m.Hostname, expiresStr)
	return subject, telegramMsg, emailHTML
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

// ─── Domain monitor checks ───────────────────────────────────────────────────
//
// Mirrors the SSL monitor checks above (same 30/14/7-day threshold pattern,
// same flag-reset-on-renewal logic) — the only differences are the data
// source (RDAP lookup instead of a TLS handshake) and the field names
// (registrar instead of issuer).

func checkDomainMonitors(ctx context.Context, n Notifiers) {
	monitors, err := n.Queries.ListDueDomainMonitors(ctx)
	if err != nil {
		n.Logger.Error("domain worker: list due monitors", "err", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, checkConcurrency)
	for _, m := range monitors {
		m := m
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			checkOneDomainMonitor(ctx, n, m)
		}()
	}
	wg.Wait()
}

func checkOneDomainMonitor(ctx context.Context, n Notifiers, m db.DomainMonitor) {
	lookup, lookupErr := n.RDAP.Lookup(m.Domain)
	var daysLeft int
	if lookupErr == nil {
		daysLeft = int(time.Until(lookup.ExpiresAt).Hours() / 24)
	}
	result := buildDomainCheckResult(m, lookup.ExpiresAt, lookup.Registrar, daysLeft, lookupErr)

	if result.alert != nil {
		DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "domain", ID: m.ID}, *result.alert)
	}

	if _, err := n.Queries.UpdateDomainMonitorCheck(ctx, db.UpdateDomainMonitorCheckParams{
		ID:                  m.ID,
		Status:              result.status,
		ExpiresAt:           result.expiresAtParam,
		Registrar:           result.registrarParam,
		ErrorMsg:            result.errorMsgParam,
		Alerted30d:          result.alerted30d,
		Alerted14d:          result.alerted14d,
		Alerted7d:           result.alerted7d,
		ConsecutiveFailures: result.consecutiveFailures,
		AlertCount:          result.alertCount,
	}); err != nil {
		n.Logger.Error("domain worker: update check", "monitor_id", m.ID, "err", err)
	}
}

// domainCheckResult is the outcome of one RDAP lookup: the fields to persist
// on the monitor row, plus an alert to send when a threshold was freshly
// crossed (nil otherwise).
type domainCheckResult struct {
	status              db.DomainMonitorStatus
	expiresAtParam      pgtype.Timestamptz
	registrarParam      pgtype.Text
	errorMsgParam       pgtype.Text
	alerted30d          bool
	alerted14d          bool
	alerted7d           bool
	consecutiveFailures int32
	alertCount          int32
	alert               *AlertMessage
}

func buildDomainCheckResult(m db.DomainMonitor, expiresAt time.Time, registrar string, daysLeft int, checkErr error) domainCheckResult {
	r := domainCheckResult{alerted30d: m.Alerted30d, alerted14d: m.Alerted14d, alerted7d: m.Alerted7d, alertCount: m.AlertCount}

	if checkErr != nil {
		r.status = db.DomainMonitorStatusError
		r.errorMsgParam = pgtype.Text{String: checkErr.Error(), Valid: true}
		r.consecutiveFailures = m.ConsecutiveFailures + 1
		return r
	}

	r.expiresAtParam = pgtype.Timestamptz{Time: expiresAt, Valid: true}
	r.registrarParam = pgtype.Text{String: registrar, Valid: true}

	switch {
	case daysLeft < 0:
		r.status = db.DomainMonitorStatusExpired
	case daysLeft <= 30:
		r.status = db.DomainMonitorStatusExpiringSoon
	default:
		r.status = db.DomainMonitorStatusUp
		// Reset flags, failure counter, and alert count when domain is renewed.
		r.alerted30d, r.alerted14d, r.alerted7d = false, false, false
		r.consecutiveFailures = 0
		r.alertCount = 0
		return r
	}

	r.consecutiveFailures = m.ConsecutiveFailures + 1

	maxAlerts := m.MaxAlertsPerIncident
	withinLimit := maxAlerts == 0 || m.AlertCount < maxAlerts
	if m.AlertsEnabled && withinLimit && r.consecutiveFailures > m.AlertAfterNFailures {
		r.alert = domainThresholdAlert(m, daysLeft, expiresAt, &r.alerted30d, &r.alerted14d, &r.alerted7d)
		if r.alert != nil {
			r.alertCount = m.AlertCount + 1
		}
	}
	return r
}

// domainThresholdAlert returns the alert for the first newly-crossed expiry
// threshold (expired, then 7/14/30 days), setting the matching alerted flag
// so checkOneDomainMonitor only sends one notification per crossing.
func domainThresholdAlert(m db.DomainMonitor, daysLeft int, expiresAt time.Time, alerted30d, alerted14d, alerted7d *bool) *AlertMessage {
	if !sslCrossedThreshold(daysLeft, alerted30d, alerted14d, alerted7d) {
		return nil
	}

	expiresStr := expiresAt.Format("2006-01-02")
	var subject, telegramMsg, emailHTML string
	if daysLeft < 0 {
		subject, telegramMsg, emailHTML = domainExpiredMessages(m, expiresStr)
	} else {
		subject, telegramMsg, emailHTML = domainExpiringSoonMessages(m, daysLeft, expiresStr)
	}

	return &AlertMessage{
		Telegram:     telegramMsg,
		EmailSubject: subject,
		EmailHTML:    emailHTML,
		Webhook: &webhook.Event{
			EventType:   "down",
			MonitorName: m.Name,
			MonitorType: "domain",
			Reason:      subject,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.DownMessage(m.Name, "domain", subject)),
		SMS:   TruncateSMS("checkmeup: " + subject),
	}
}

// domainExpiredMessages builds the alert text for a domain registration that
// has already lapsed (daysLeft < 0), distinct from
// domainExpiringSoonMessages' "expires in N days" phrasing.
func domainExpiredMessages(m db.DomainMonitor, expiresStr string) (subject, telegramMsg, emailHTML string) {
	subject = fmt.Sprintf("DOWN: %s domain registration expired", m.Name)
	telegramMsg = fmt.Sprintf("🔴 <b>%s</b> domain registration has <b>expired</b>\n\nDomain: <code>%s</code>\nExpired: %s",
		m.Name, m.Domain, expiresStr)
	emailHTML = fmt.Sprintf("<p>🔴 <b>%s</b> domain registration has <b>expired</b></p><p>Domain: <code>%s</code><br>Expired: %s</p>",
		m.Name, m.Domain, expiresStr)
	return subject, telegramMsg, emailHTML
}

func domainExpiringSoonMessages(m db.DomainMonitor, daysLeft int, expiresStr string) (subject, telegramMsg, emailHTML string) {
	subject = fmt.Sprintf("%s domain registration expires in %d days", m.Name, daysLeft)
	telegramMsg = fmt.Sprintf("🌐 <b>%s</b> domain registration expires in <b>%d days</b>\n\nDomain: <code>%s</code>\nExpires: %s",
		m.Name, daysLeft, m.Domain, expiresStr)
	emailHTML = fmt.Sprintf("<p>🌐 <b>%s</b> domain registration expires in <b>%d days</b></p><p>Domain: <code>%s</code><br>Expires: %s</p>",
		m.Name, daysLeft, m.Domain, expiresStr)
	return subject, telegramMsg, emailHTML
}

// ─── port monitor checks ─────────────────────────────────────────────────────

func checkPortMonitors(ctx context.Context, n Notifiers) {
	monitors, err := n.Queries.ListDuePortMonitors(ctx)
	if err != nil {
		n.Logger.Error("port worker: list due monitors", "err", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, checkConcurrency)
	for _, m := range monitors {
		m := m
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			checkOnePortMonitor(ctx, n, m)
		}()
	}
	wg.Wait()
}

func checkOnePortMonitor(ctx context.Context, n Notifiers, m db.PortMonitor) {
	responseTimeMs, isUp, failureReason := performTCPCheck(m)
	if !recordPortCheck(ctx, n, m, responseTimeMs, isUp, failureReason) {
		return
	}

	prevStatus := m.Status
	if isUp {
		handlePortUp(ctx, n, m, prevStatus)
		return
	}
	handlePortDown(ctx, n, m, prevStatus, failureReason)
}

func recordPortCheck(ctx context.Context, n Notifiers, m db.PortMonitor, responseTimeMs int64, isUp bool, failureReason string) bool {
	if _, err := n.Queries.CreatePortCheck(ctx, db.CreatePortCheckParams{
		MonitorID:      m.ID,
		ResponseTimeMs: int32(responseTimeMs),
		IsUp:           isUp,
		FailureReason:  pgtype.Text{String: failureReason, Valid: failureReason != ""},
	}); err != nil {
		n.Logger.Error("port worker: create check", "monitor_id", m.ID, "err", err)
		return false
	}
	return true
}

func handlePortUp(ctx context.Context, n Notifiers, m db.PortMonitor, prevStatus db.MonitorStatus) {
	if _, err := n.Queries.RecordPortCheckUp(ctx, m.ID); err != nil {
		n.Logger.Error("port worker: record up", "monitor_id", m.ID, "err", err)
		return
	}
	if prevStatus != db.MonitorStatusDown {
		return
	}
	inc, err := n.Queries.ResolveLatestPortIncident(ctx, m.ID)
	if err != nil {
		n.Logger.Error("port worker: resolve incident", "monitor_id", m.ID, "err", err)
		return
	}
	if !m.AlertsEnabled {
		return
	}
	downtime := FormatDuration(time.Since(inc.StartedAt.Time))
	DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "port", ID: m.ID}, buildPortRecoveryAlert(m, downtime))
}

func buildPortRecoveryAlert(m db.PortMonitor, downtime string) AlertMessage {
	hostPort := fmt.Sprintf("%s:%d", m.Host, m.Port)
	return AlertMessage{
		Telegram:     fmt.Sprintf("✅ <b>%s</b> is back up\n\nHost: <code>%s</code>", m.Name, hostPort),
		EmailSubject: fmt.Sprintf("%s recovered", m.Name),
		EmailHTML:    fmt.Sprintf("<p>✅ <b>%s</b> is back up</p><p>Host: <code>%s</code></p>", m.Name, hostPort),
		Webhook: &webhook.Event{
			EventType:        "recovery",
			MonitorName:      m.Name,
			MonitorType:      "port",
			DowntimeDuration: downtime,
			Timestamp:        time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.RecoveryMessage(m.Name, "port", downtime)),
		SMS:   TruncateSMS(fmt.Sprintf("checkmeup: %s recovered after %s downtime", m.Name, downtime)),
	}
}

func handlePortDown(ctx context.Context, n Notifiers, m db.PortMonitor, prevStatus db.MonitorStatus, failureReason string) {
	updated, err := n.Queries.RecordPortCheckFailure(ctx, m.ID)
	if err != nil {
		n.Logger.Error("port worker: record failure", "monitor_id", m.ID, "err", err)
		return
	}
	if updated.ConsecutiveFailures <= m.AlertAfterNFailures || prevStatus == db.MonitorStatusDown {
		return
	}
	if err := n.Queries.MarkPortMonitorDown(ctx, m.ID); err != nil {
		n.Logger.Error("port worker: mark down", "monitor_id", m.ID, "err", err)
		return
	}
	inc, err := n.Queries.CreatePortIncident(ctx, m.ID)
	if err != nil {
		n.Logger.Error("port worker: create incident", "monitor_id", m.ID, "err", err)
		return
	}
	alertPortIncident(ctx, n, m, inc, failureReason)
}

// alertPortIncident dispatches the down alert for a freshly-recorded
// incident, honoring the monitor's alerts-enabled flag and per-incident cap.
func alertPortIncident(ctx context.Context, n Notifiers, m db.PortMonitor, inc db.PortIncident, failureReason string) {
	if !m.AlertsEnabled {
		return
	}
	if max := m.MaxAlertsPerIncident; max > 0 && inc.AlertCount >= max {
		return
	}
	if !DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "port", ID: m.ID}, buildPortDownAlert(m, failureReason)) {
		return
	}
	if _, err := n.Queries.IncrementPortIncidentAlertCount(ctx, inc.ID); err != nil {
		n.Logger.Error("port worker: increment alert count", "incident_id", inc.ID, "err", err)
	}
}

// buildPortDownAlert phrases the headline around expected_state: a
// closed-state monitor going "down" means the port unexpectedly became
// reachable, which reads very differently from an open-state monitor's
// service outage.
func buildPortDownAlert(m db.PortMonitor, failureReason string) AlertMessage {
	hostPort := fmt.Sprintf("%s:%d", m.Host, m.Port)
	if m.ExpectedState == db.PortExpectedStateClosed {
		subject := fmt.Sprintf("%s: port unexpectedly open", m.Name)
		return AlertMessage{
			Telegram: fmt.Sprintf("⚠️ <b>%s</b>: port unexpectedly open\n\nHost: <code>%s</code>",
				m.Name, hostPort),
			EmailSubject: subject,
			EmailHTML: fmt.Sprintf("<p>⚠️ <b>%s</b>: port unexpectedly open</p><p>Host: <code>%s</code></p>",
				m.Name, hostPort),
			Webhook: &webhook.Event{
				EventType:   "down",
				MonitorName: m.Name,
				MonitorType: "port",
				Reason:      failureReason,
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
			},
			Slack: slackMsg(slack.DownMessage(m.Name, "port", failureReason)),
			SMS:   TruncateSMS(fmt.Sprintf("checkmeup: %s", subject)),
		}
	}
	return AlertMessage{
		Telegram: fmt.Sprintf("🔴 <b>%s</b> is down\n\nHost: <code>%s</code>\nReason: %s",
			m.Name, hostPort, failureReason),
		EmailSubject: fmt.Sprintf("DOWN: %s", m.Name),
		EmailHTML: fmt.Sprintf("<p>🔴 <b>%s</b> is down</p><p>Host: <code>%s</code><br>Reason: %s</p>",
			m.Name, hostPort, failureReason),
		Webhook: &webhook.Event{
			EventType:   "down",
			MonitorName: m.Name,
			MonitorType: "port",
			Reason:      failureReason,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.DownMessage(m.Name, "port", failureReason)),
		SMS:   TruncateSMS(fmt.Sprintf("checkmeup: %s is DOWN (%s)", m.Name, failureReason)),
	}
}

// performTCPCheck opens a raw TCP connection to the monitor's host:port —
// no data sent or received, no protocol handshake. The dial outcome is
// interpreted against the monitor's expected state: an "open" monitor wants
// a successful connect, a "closed" monitor wants the opposite (US-3302).
func performTCPCheck(m db.PortMonitor) (responseTimeMs int64, isUp bool, failureReason string) {
	addr := net.JoinHostPort(m.Host, strconv.Itoa(int(m.Port)))
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	elapsed := time.Since(start).Milliseconds()
	connected := err == nil
	if connected {
		_ = conn.Close()
	}

	if m.ExpectedState == db.PortExpectedStateClosed {
		if connected {
			return elapsed, false, "port is unexpectedly open"
		}
		return elapsed, true, ""
	}
	if connected {
		return elapsed, true, ""
	}
	return elapsed, false, "connection refused / timeout"
}
