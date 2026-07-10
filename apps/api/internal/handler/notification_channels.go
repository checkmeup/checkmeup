package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/billing"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/respond"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/twilio"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

type NotificationChannelHandler struct {
	queries *db.Queries
	tg      *telegram.Client
	mailer  *email.Sender
	wh      *webhook.Client
	sl      *slack.Client
	sm      *twilio.Client
	// smsTestLimiter/smsTestHourlyLimiter/smsTestDailyLimiter bound "send test
	// SMS" to 10/minute, 10/hour, and 20/day per org — SMS is the only
	// test-send with a real per-message cost, so it gets tighter, org-scoped
	// limits (burst + sustained + daily ceiling) on top of the route's
	// blanket per-IP one. Built directly (not via httprate.Limit middleware)
	// since the limit only applies to one channel type within a shared
	// endpoint — see TestNotificationChannel, which supplies the org-derived
	// key itself rather than a route-level KeyFunc.
	smsTestLimiter       *httprate.RateLimiter
	smsTestHourlyLimiter *httprate.RateLimiter
	smsTestDailyLimiter  *httprate.RateLimiter
}

func NewNotificationChannelHandler(pool *pgxpool.Pool, tg *telegram.Client, mailer *email.Sender, wh *webhook.Client, sl *slack.Client, sm *twilio.Client) *NotificationChannelHandler {
	return &NotificationChannelHandler{
		queries:              db.New(pool),
		tg:                   tg,
		mailer:               mailer,
		wh:                   wh,
		sl:                   sl,
		sm:                   sm,
		smsTestLimiter:       httprate.NewRateLimiter(10, time.Minute),
		smsTestHourlyLimiter: httprate.NewRateLimiter(10, time.Hour),
		smsTestDailyLimiter:  httprate.NewRateLimiter(20, 24*time.Hour),
	}
}

type notificationChannelResponse struct {
	ID                 string         `json:"id"`
	Type               string         `json:"type"`
	Name               string         `json:"name"`
	Config             map[string]any `json:"config"`
	Enabled            bool           `json:"enabled"`
	CreatedAt          string         `json:"createdAt"`
	LastDeliveryStatus string         `json:"lastDeliveryStatus,omitempty"`
	LastDeliveryDetail string         `json:"lastDeliveryDetail,omitempty"`
	LastDeliveryAt     string         `json:"lastDeliveryAt,omitempty"`
}

// toNotificationChannelResponse round-trips config verbatim, including a
// webhook channel's signing secret. ADR-023 flags secrets in config JSONB
// as something that "should never round-trip unmasked" in general (Slack
// webhook URLs, future OAuth tokens) — but a checkmeup-issued HMAC signing
// secret is different from a bearer credential: US-1403 requires it stay
// "viewable... in Settings" so the user can configure their endpoint to
// verify X-Checkmeup-Signature. Masking it would break that requirement.
func toNotificationChannelResponse(c db.NotificationChannel) notificationChannelResponse {
	var cfg map[string]any
	_ = json.Unmarshal(c.Config, &cfg)
	return notificationChannelResponse{
		ID:                 c.ID.String(),
		Type:               string(c.Type),
		Name:               c.Name,
		Config:             cfg,
		Enabled:            c.Enabled,
		CreatedAt:          c.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		LastDeliveryStatus: c.LastDeliveryStatus.String,
		LastDeliveryDetail: c.LastDeliveryDetail.String,
		LastDeliveryAt:     formatDeliveryAt(c.LastDeliveryAt),
	}
}

func formatDeliveryAt(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("2006-01-02T15:04:05Z")
}

// validateChannelConfig checks config has the field(s) required for type.
// Other notification_channel_type values get added (their own migration +
// their own case here) when their epic actually ships (ADR-023).
func validateChannelConfig(channelType string, config map[string]any) error {
	switch channelType {
	case "telegram":
		chatID, _ := config["chatId"].(string)
		if strings.TrimSpace(chatID) == "" {
			return errors.New("chatId is required")
		}
	case "email":
		addr, _ := config["email"].(string)
		if strings.TrimSpace(addr) == "" {
			return errors.New("email is required")
		}
	case "webhook":
		return validateWebhookURL(config)
	case "slack":
		return validateSlackURL(config)
	case "sms":
		return validateSMSConfig(config)
	default:
		return errors.New("unsupported channel type")
	}
	return nil
}

// e164Pattern matches E.164 phone numbers (US-1901): a leading +, no leading
// zero, up to 15 digits total.
var e164Pattern = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// validateSMSConfig enforces US-1901: a valid E.164 phone number, and an
// explicit opt-in checkbox — a TCPA-style regulatory requirement for
// automated texts, not satisfied just by providing a number (ADR-029).
// consent is required true on every request that reaches here; callers that
// want to carry an existing consent forward without re-prompting the user
// (see resolveUpdatedChannelConfig, for an unchanged phone number on update)
// inject consent: true themselves before calling validateChannelConfig.
func validateSMSConfig(config map[string]any) error {
	phone, _ := config["phone_number"].(string)
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return errors.New("phone_number is required")
	}
	if !e164Pattern.MatchString(phone) {
		return errors.New("phone_number must be in E.164 format (e.g. +14155551234)")
	}
	if !consentGiven(config["consent"]) {
		return errors.New("consent is required — check the opt-in box before saving")
	}
	return nil
}

// consentGiven reports true for either a JSON boolean true or the string
// "true" — config values arrive as any (json.Unmarshal into map[string]any),
// but every other channel's config is plain strings on the wire (chatId,
// email, url...), so the frontend sends "true" here too rather than being
// the one field that needs a non-string type.
func consentGiven(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true"
	default:
		return false
	}
}

// finalizeSMSConsent strips the client-supplied consent flag (already
// validated true by validateSMSConfig) and stamps a server-set consent_at —
// never trusting a client-supplied timestamp — unless one was already
// carried forward from an existing channel (see resolveUpdatedChannelConfig).
func finalizeSMSConsent(config map[string]any) {
	if _, ok := config["consent_at"]; !ok {
		config["consent_at"] = time.Now().UTC().Format(time.RFC3339)
	}
	delete(config, "consent")
}

// validateWebhookURL enforces the US-1401 AC that a webhook URL must be
// https://. Doesn't require config["secret"] — that's generated server-side
// (see ensureWebhookSecret), never supplied by the client.
func validateWebhookURL(config map[string]any) error {
	rawURL, _ := config["url"].(string)
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return errors.New("url is required")
	}
	if !strings.HasPrefix(rawURL, "https://") {
		return errors.New("url must start with https://")
	}
	return nil
}

// validateSlackURL enforces US-1701: the URL must match the Slack Incoming
// Webhook pattern (https://hooks.slack.com/...). The URL itself is the
// credential — no separate secret like the generic webhook channel (US-1401).
func validateSlackURL(config map[string]any) error {
	rawURL, _ := config["url"].(string)
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return errors.New("url is required")
	}
	if !strings.HasPrefix(rawURL, "https://hooks.slack.com/") {
		return errors.New("url must be a Slack Incoming Webhook URL (https://hooks.slack.com/...)")
	}
	return nil
}

// ListNotificationChannels GET /api/v1/notification-channels
func (h *NotificationChannelHandler) ListNotificationChannels(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	channels, err := h.queries.ListNotificationChannels(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	result := make([]notificationChannelResponse, len(channels))
	for i, c := range channels {
		result[i] = toNotificationChannelResponse(c)
	}
	respond.JSON(w, http.StatusOK, result)
}

type notificationChannelRequest struct {
	Type    string         `json:"type"`
	Name    string         `json:"name"`
	Config  map[string]any `json:"config"`
	Enabled *bool          `json:"enabled"`
}

// CreateNotificationChannel POST /api/v1/notification-channels
func (h *NotificationChannelHandler) CreateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req notificationChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	// consent_at is always server-stamped (finalizeSMSConsent) — never trust
	// a client-supplied value here, or a forged timestamp would persist.
	delete(req.Config, "consent_at")

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}

	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	total, err := h.queries.CountOrgNotificationChannels(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if err := billing.CheckNotificationChannelLimit(plan, int(total)); err != nil {
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return
	}
	// A plan with 0 monthly SMS credits (Hobby, per ADR-032) can't create an
	// sms channel at all — no point saving one that can never send.
	if req.Type == "sms" && billing.GetLimits(plan).SMSCredits <= 0 {
		respond.Error(w, http.StatusPaymentRequired, "SMS alerts require a paid plan — upgrade to enable this channel", "plan_limit_reached")
		return
	}

	if err := validateChannelConfig(req.Type, req.Config); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	if req.Type == "webhook" {
		secret, err := webhook.GenerateSecret()
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
			return
		}
		req.Config["secret"] = secret
	}
	if req.Type == "sms" {
		finalizeSMSConsent(req.Config)
	}
	configBytes, err := json.Marshal(req.Config)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid config", "bad_request")
		return
	}

	channel, err := h.queries.CreateNotificationChannel(r.Context(), db.CreateNotificationChannelParams{
		OrgID:  orgID,
		Type:   db.NotificationChannelType(req.Type),
		Name:   req.Name,
		Config: configBytes,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	respond.JSON(w, http.StatusCreated, toNotificationChannelResponse(channel))
}

// UpdateNotificationChannel PATCH /api/v1/notification-channels/{id}
func (h *NotificationChannelHandler) UpdateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	orgID, channelID, ok := notificationChannelIDs(w, r)
	if !ok {
		return
	}

	var req notificationChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	// consent_at is always server-stamped (finalizeSMSConsent, or carried
	// forward explicitly in resolveUpdatedChannelConfig) — never trust a
	// client-supplied value here, or a forged timestamp would persist.
	delete(req.Config, "consent_at")
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}

	existing, err := h.queries.GetNotificationChannel(r.Context(), db.GetNotificationChannelParams{ID: channelID, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "channel not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	configBytes, enabled, err := resolveUpdatedChannelConfig(existing, req)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	// Re-enabling a previously-disabled channel grows the org's active
	// channel count, same as creating one — gated the same way (ADR-019),
	// checked against enabled-only count since a disabled channel doesn't
	// count against the limit at all.
	if enabled && !existing.Enabled {
		plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
			return
		}
		enabledCount, err := h.queries.CountEnabledNotificationChannelsForOrg(r.Context(), orgID)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
			return
		}
		if err := billing.CheckNotificationChannelLimit(plan, int(enabledCount)); err != nil {
			respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
			return
		}
	}

	channel, err := h.queries.UpdateNotificationChannel(r.Context(), db.UpdateNotificationChannelParams{
		ID:      channelID,
		OrgID:   orgID,
		Name:    req.Name,
		Config:  configBytes,
		Enabled: enabled,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	respond.JSON(w, http.StatusOK, toNotificationChannelResponse(channel))
}

// resolveUpdatedChannelConfig validates req.Config against the channel's
// existing type (which never changes after creation) and returns the JSON
// config bytes plus the enabled value to write, preserving existing.Enabled
// when req.Enabled is omitted.
func resolveUpdatedChannelConfig(existing db.NotificationChannel, req notificationChannelRequest) ([]byte, bool, error) {
	if existing.Type == db.NotificationChannelTypeWebhook {
		// The signing secret can only change via RegenerateWebhookSecret
		// (US-1403) — a regular PATCH (editing the URL or name) always
		// carries the existing secret forward, ignoring anything the client
		// sent for it.
		var existingCfg map[string]any
		_ = json.Unmarshal(existing.Config, &existingCfg)
		if req.Config == nil {
			req.Config = map[string]any{}
		}
		req.Config["secret"] = existingCfg["secret"]
	}
	if existing.Type == db.NotificationChannelTypeSms {
		// Consent (ADR-029) is tied to the phone number, not to the act of
		// editing — re-saving the same number shouldn't force the user to
		// re-check the opt-in box. Only require fresh consent when the
		// number itself is changing (a new recipient).
		var existingCfg map[string]any
		_ = json.Unmarshal(existing.Config, &existingCfg)
		existingPhone, _ := existingCfg["phone_number"].(string)
		if req.Config == nil {
			req.Config = map[string]any{}
		}
		newPhone, _ := req.Config["phone_number"].(string)
		if strings.TrimSpace(newPhone) == strings.TrimSpace(existingPhone) {
			req.Config["consent"] = true
			req.Config["consent_at"] = existingCfg["consent_at"]
		}
	}
	if err := validateChannelConfig(string(existing.Type), req.Config); err != nil {
		return nil, false, err
	}
	if existing.Type == db.NotificationChannelTypeSms {
		finalizeSMSConsent(req.Config)
	}
	configBytes, err := json.Marshal(req.Config)
	if err != nil {
		return nil, false, errors.New("invalid config")
	}
	enabled := existing.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	return configBytes, enabled, nil
}

// RegenerateWebhookSecret POST /api/v1/notification-channels/{id}/regenerate-secret
// Rotates a webhook channel's HMAC signing secret in place (US-1403).
// Existing requests already in flight keep validating against the old
// secret until they're sent — only future sends use the new one, since
// nothing is retroactively re-signed.
func (h *NotificationChannelHandler) RegenerateWebhookSecret(w http.ResponseWriter, r *http.Request) {
	orgID, channelID, ok := notificationChannelIDs(w, r)
	if !ok {
		return
	}

	existing, err := h.queries.GetNotificationChannel(r.Context(), db.GetNotificationChannelParams{ID: channelID, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "channel not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if existing.Type != db.NotificationChannelTypeWebhook {
		respond.Error(w, http.StatusBadRequest, "only webhook channels have a signing secret", "bad_request")
		return
	}

	var cfg map[string]any
	_ = json.Unmarshal(existing.Config, &cfg)
	secret, err := webhook.GenerateSecret()
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	cfg["secret"] = secret
	configBytes, err := json.Marshal(cfg)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	channel, err := h.queries.UpdateNotificationChannelConfig(r.Context(), db.UpdateNotificationChannelConfigParams{
		ID: channelID, OrgID: orgID, Config: configBytes,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	respond.JSON(w, http.StatusOK, toNotificationChannelResponse(channel))
}

// DeleteNotificationChannel DELETE /api/v1/notification-channels/{id}
func (h *NotificationChannelHandler) DeleteNotificationChannel(w http.ResponseWriter, r *http.Request) {
	orgID, channelID, ok := notificationChannelIDs(w, r)
	if !ok {
		return
	}

	if err := h.queries.DeleteNotificationChannel(r.Context(), db.DeleteNotificationChannelParams{
		ID: channelID, OrgID: orgID,
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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

func notificationChannelIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return uuid.UUID{}, uuid.UUID{}, false
	}
	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid channel id", "bad_request")
		return uuid.UUID{}, uuid.UUID{}, false
	}
	return orgID, channelID, true
}
