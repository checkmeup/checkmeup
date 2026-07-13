package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/checkmeup/checkmeup/internal/billing"
	"github.com/checkmeup/checkmeup/internal/respond"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

type testNotificationChannelRequest struct {
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
}

// TestNotificationChannel POST /api/v1/notification-channels/test
// Sends a test message using the given type+config without requiring it to
// be saved first, so the UI can verify a channel before saving (US-2801).
func (h *NotificationChannelHandler) TestNotificationChannel(w http.ResponseWriter, r *http.Request) {
	var req testNotificationChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	if err := validateChannelConfig(req.Type, req.Config); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	if req.Type == "sms" && !h.checkSMSTestAllowed(w, r) {
		return
	}
	if !h.sendTestNotification(w, req.Type, req.Config) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// checkSMSTestAllowed gates SMS test-sends behind plan credits and rate
// limits — same guard as CreateNotificationChannel. Without this, a
// zero-credit-plan org could rack up real Twilio spend via test sends alone,
// without ever being able to save/use an sms channel. Responds and returns
// false if the send should be blocked.
func (h *NotificationChannelHandler) checkSMSTestAllowed(w http.ResponseWriter, r *http.Request) bool {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return false
	}
	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return false
	}
	if billing.GetLimits(plan).SMSCredits <= 0 {
		respond.Error(w, http.StatusPaymentRequired, "SMS alerts require a paid plan — upgrade to enable this channel", "plan_limit_reached")
		return false
	}
	if h.smsTestLimiter.RespondOnLimit(w, r, orgID.String()) {
		return false
	}
	if h.smsTestHourlyLimiter.RespondOnLimit(w, r, orgID.String()) {
		return false
	}
	return !h.smsTestDailyLimiter.RespondOnLimit(w, r, orgID.String())
}

// sendTestNotification dispatches a test message for the given channel type,
// responding with a type-specific error and returning false on failure.
func (h *NotificationChannelHandler) sendTestNotification(w http.ResponseWriter, channelType string, config map[string]any) bool {
	switch channelType {
	case "telegram":
		chatID, _ := config["chatId"].(string)
		if err := h.tg.SendMessage(strings.TrimSpace(chatID), "✅ Checkmeup is connected! You'll receive alerts here."); err != nil {
			respond.Error(w, http.StatusBadGateway, err.Error(), "telegram_error")
			return false
		}
	case "email":
		addr, _ := config["email"].(string)
		if err := h.mailer.SendTestAlertEmail(strings.TrimSpace(addr)); err != nil {
			respond.Error(w, http.StatusBadGateway, err.Error(), "email_error")
			return false
		}
	case "webhook":
		url, _ := config["url"].(string)
		// The real signing secret doesn't exist until the channel is saved
		// (US-1401) — sign with a throwaway one so the request shape on the
		// wire matches a real send, even though there's nothing yet for the
		// receiver to verify against.
		secret, err := webhook.GenerateSecret()
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
			return false
		}
		event := webhook.Event{
			EventType:   "test",
			MonitorName: "Test monitor",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}
		if _, err := h.wh.Send(strings.TrimSpace(url), secret, event); err != nil {
			respond.Error(w, http.StatusBadGateway, err.Error(), "webhook_error")
			return false
		}
	case "slack":
		url, _ := config["url"].(string)
		if _, err := h.sl.Send(strings.TrimSpace(url), slack.TestMessage()); err != nil {
			respond.Error(w, http.StatusBadGateway, err.Error(), "slack_error")
			return false
		}
	case "sms":
		phone, _ := config["phone_number"].(string)
		if _, err := h.sm.Send(strings.TrimSpace(phone), "Checkmeup: this is a test SMS alert. You're all set!"); err != nil {
			respond.Error(w, http.StatusBadGateway, err.Error(), "sms_error")
			return false
		}
	}
	return true
}
