package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
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

	// HTTPClient and TCPDialer back the uptime and port checks respectively.
	// A monitor's Host/URL is user-supplied, so leaving these nil (the
	// production default) wires them through httpsafe.Dialer to block
	// loopback/private/link-local/cloud-metadata targets (SSRF) — see
	// uptimeCheckClient/portCheckDialer. Tests that need to reach a local
	// httptest server set these explicitly to an unguarded client/dialer.
	HTTPClient *http.Client
	TCPDialer  *net.Dialer
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

// expiryAlertSubject describes the per-monitor-type wording for
// expiredMessages/expiringSoonMessages below, shared by the SSL and domain
// checks — both follow the same 30/14/7-day threshold pattern
// (sslCrossedThreshold) and only differ in what's being checked and which
// field to display.
type expiryAlertSubject struct {
	Name       string // monitor name
	Thing      string // "SSL certificate" / "domain registration"
	FieldLabel string // "Host" / "Domain"
	Target     string // the checked value: hostname / domain
	Emoji      string // expiringSoonMessages' icon: "🔐" / "🌐" (expiredMessages always uses 🔴)
}

// expiredMessages builds the DOWN alert text for an expiry-threshold check
// (SSL/domain) that has already lapsed, distinct from expiringSoonMessages'
// "expires in N days" phrasing.
func expiredMessages(s expiryAlertSubject, expiresStr string) (subject, telegramMsg, emailHTML string) {
	subject = fmt.Sprintf("DOWN: %s %s expired", s.Name, s.Thing)
	telegramMsg = fmt.Sprintf("🔴 <b>%s</b> %s has <b>expired</b>\n\n%s: <code>%s</code>\nExpired: %s",
		s.Name, s.Thing, s.FieldLabel, s.Target, expiresStr)
	emailHTML = fmt.Sprintf("<p>🔴 <b>%s</b> %s has <b>expired</b></p><p>%s: <code>%s</code><br>Expired: %s</p>",
		s.Name, s.Thing, s.FieldLabel, s.Target, expiresStr)
	return
}

// expiringSoonMessages builds the threshold-crossing alert text for an
// expiry-threshold check (SSL/domain) still ahead of expiry.
func expiringSoonMessages(s expiryAlertSubject, daysLeft int, expiresStr string) (subject, telegramMsg, emailHTML string) {
	subject = fmt.Sprintf("%s %s expires in %d days", s.Name, s.Thing, daysLeft)
	telegramMsg = fmt.Sprintf("%s <b>%s</b> %s expires in <b>%d days</b>\n\n%s: <code>%s</code>\nExpires: %s",
		s.Emoji, s.Name, s.Thing, daysLeft, s.FieldLabel, s.Target, expiresStr)
	emailHTML = fmt.Sprintf("<p>%s <b>%s</b> %s expires in <b>%d days</b></p><p>%s: <code>%s</code><br>Expires: %s</p>",
		s.Emoji, s.Name, s.Thing, daysLeft, s.FieldLabel, s.Target, expiresStr)
	return
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

// The remaining monitor-type-specific check loops (cron, uptime, SSL, domain,
// port) live in worker_<type>.go — split out of this file so each check loop
// is reviewable on its own (architecture-guardrails audit: this file alone
// was ~1070 logical lines). checkConcurrency, Notifiers, AlertMessage,
// MonitorRef, DispatchAlert, and the helpers above are shared by all of them.
