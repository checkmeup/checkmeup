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
		queries: db.New(pool),
		tg:      tg,
		mailer:  mailer,
		wh:      wh,
		sl:      sl,
		sm:      sm,
		// WithLimitCounter pins each limiter to wall-clock-aligned windows —
		// httprate's default (as of v0.16.0) instead offsets windows by
		// however long after process start the limiter was constructed, to
		// spread out resets across many differently-offset limiters. That
		// doesn't help here (one shared instance per limiter, keyed by org,
		// not many instances), and it broke tests that seed a window's count
		// directly via a wall-clock-truncated key.
		smsTestLimiter:       httprate.NewRateLimiter(10, time.Minute, httprate.WithLimitCounter(httprate.NewLocalLimitCounter(time.Minute))),
		smsTestHourlyLimiter: httprate.NewRateLimiter(10, time.Hour, httprate.WithLimitCounter(httprate.NewLocalLimitCounter(time.Hour))),
		smsTestDailyLimiter:  httprate.NewRateLimiter(20, 24*time.Hour, httprate.WithLimitCounter(httprate.NewLocalLimitCounter(24*time.Hour))),
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

// respondChannelNotFoundOrInternal writes a 404 if err is pgx.ErrNoRows, a
// 500 for any other non-nil error, and reports whether it wrote a response
// so the caller knows to return immediately.
func respondChannelNotFoundOrInternal(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, pgx.ErrNoRows) {
		respond.Error(w, http.StatusNotFound, "channel not found", "not_found")
		return true
	}
	respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
	return true
}

// decodeNotificationChannelRequest decodes and normalizes a create/update
// request body: strips any client-supplied consent_at (always server-
// stamped — finalizeSMSConsent, or carried forward explicitly in
// resolveUpdatedChannelConfig on update) and validates the name is
// non-empty. Shared by CreateNotificationChannel and UpdateNotificationChannel.
func decodeNotificationChannelRequest(w http.ResponseWriter, r *http.Request) (notificationChannelRequest, bool) {
	var req notificationChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return req, false
	}
	delete(req.Config, "consent_at")
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return req, false
	}
	return req, true
}

// checkNotificationChannelCreateLimit checks whether orgID can create
// another notification channel under its plan, returning the resolved plan
// for further per-type checks (e.g. SMS credits). Responds and returns false
// on any failure (query error or plan limit hit).
func (h *NotificationChannelHandler) checkNotificationChannelCreateLimit(w http.ResponseWriter, r *http.Request, orgID uuid.UUID) (db.Plan, bool) {
	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return "", false
	}
	total, err := h.queries.CountOrgNotificationChannels(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return "", false
	}
	if err := billing.CheckNotificationChannelLimit(plan, int(total)); err != nil {
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return "", false
	}
	return plan, true
}

// checkNotificationChannelReenableLimit checks whether re-enabling a
// disabled channel would push orgID over its plan's channel limit — same
// rule as creating one (ADR-019), checked against the enabled-only count
// since a disabled channel doesn't count against the limit at all.
func (h *NotificationChannelHandler) checkNotificationChannelReenableLimit(w http.ResponseWriter, r *http.Request, orgID uuid.UUID) bool {
	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return false
	}
	enabledCount, err := h.queries.CountEnabledNotificationChannelsForOrg(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return false
	}
	if err := billing.CheckNotificationChannelLimit(plan, int(enabledCount)); err != nil {
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return false
	}
	return true
}

// smsRequiresPaidPlan reports whether creating an sms channel should be
// blocked because the plan has no monthly SMS credits (Hobby, per ADR-032)
// — no point saving a channel that can never send.
func smsRequiresPaidPlan(channelType string, plan db.Plan) bool {
	return channelType == "sms" && billing.GetLimits(plan).SMSCredits <= 0
}

// channelIsBeingReenabled reports whether an update flips a previously-
// disabled channel back on.
func channelIsBeingReenabled(enabled bool, existing db.NotificationChannel) bool {
	return enabled && !existing.Enabled
}

// finalizeChannelConfig applies type-specific server-side config fields
// (webhook signing secret, SMS consent stamp) and marshals the result.
// Responds and reports ok=false on failure.
func finalizeChannelConfig(w http.ResponseWriter, channelType string, config map[string]any) ([]byte, bool) {
	if channelType == "webhook" {
		secret, err := webhook.GenerateSecret()
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
			return nil, false
		}
		config["secret"] = secret
	}
	if channelType == "sms" {
		finalizeSMSConsent(config)
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid config", "bad_request")
		return nil, false
	}
	return configBytes, true
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

	req, ok := decodeNotificationChannelRequest(w, r)
	if !ok {
		return
	}

	plan, ok := h.checkNotificationChannelCreateLimit(w, r, orgID)
	if !ok {
		return
	}
	if smsRequiresPaidPlan(req.Type, plan) {
		respond.Error(w, http.StatusPaymentRequired, "SMS alerts require a paid plan — upgrade to enable this channel", "plan_limit_reached")
		return
	}

	if err := validateChannelConfig(req.Type, req.Config); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	configBytes, ok := finalizeChannelConfig(w, req.Type, req.Config)
	if !ok {
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

	req, ok := decodeNotificationChannelRequest(w, r)
	if !ok {
		return
	}

	existing, err := h.queries.GetNotificationChannel(r.Context(), db.GetNotificationChannelParams{ID: channelID, OrgID: orgID})
	if respondChannelNotFoundOrInternal(w, err) {
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
	if channelIsBeingReenabled(enabled, existing) && !h.checkNotificationChannelReenableLimit(w, r, orgID) {
		return
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

// carryForwardWebhookSecret preserves an existing webhook channel's signing
// secret across an update — it can only change via RegenerateWebhookSecret
// (US-1403), so a regular PATCH (editing the URL or name) always carries it
// forward, ignoring anything the client sent for it.
func carryForwardWebhookSecret(existing db.NotificationChannel, config map[string]any) map[string]any {
	if config == nil {
		config = map[string]any{}
	}
	var existingCfg map[string]any
	_ = json.Unmarshal(existing.Config, &existingCfg)
	config["secret"] = existingCfg["secret"]
	return config
}

// carryForwardSMSConsent preserves an existing SMS channel's consent
// (ADR-029) across an update when the phone number is unchanged — consent is
// tied to the number, not to the act of editing, so re-saving the same
// number shouldn't force the user to re-check the opt-in box.
func carryForwardSMSConsent(existing db.NotificationChannel, config map[string]any) map[string]any {
	if config == nil {
		config = map[string]any{}
	}
	var existingCfg map[string]any
	_ = json.Unmarshal(existing.Config, &existingCfg)
	existingPhone, _ := existingCfg["phone_number"].(string)
	newPhone, _ := config["phone_number"].(string)
	if strings.TrimSpace(newPhone) == strings.TrimSpace(existingPhone) {
		config["consent"] = true
		config["consent_at"] = existingCfg["consent_at"]
	}
	return config
}

// resolveUpdatedChannelConfig validates req.Config against the channel's
// existing type (which never changes after creation) and returns the JSON
// config bytes plus the enabled value to write, preserving existing.Enabled
// when req.Enabled is omitted.
func resolveUpdatedChannelConfig(existing db.NotificationChannel, req notificationChannelRequest) ([]byte, bool, error) {
	config := req.Config
	switch existing.Type {
	case db.NotificationChannelTypeWebhook:
		config = carryForwardWebhookSecret(existing, config)
	case db.NotificationChannelTypeSms:
		config = carryForwardSMSConsent(existing, config)
	}
	if err := validateChannelConfig(string(existing.Type), config); err != nil {
		return nil, false, err
	}
	if existing.Type == db.NotificationChannelTypeSms {
		finalizeSMSConsent(config)
	}
	configBytes, err := json.Marshal(config)
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
	if respondChannelNotFoundOrInternal(w, err) {
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

// sendTestTelegram, sendTestEmail, etc. each perform one channel type's test
// send, returning the error (if any) for sendTestNotification to report
// against that type's error code. webhook is the only one with its own
// pre-step (a throwaway signing secret — the real one doesn't exist until
// the channel is saved, US-1401 — so the request shape on the wire still
// matches a real send even with nothing for the receiver to verify against).

func (h *NotificationChannelHandler) sendTestTelegram(config map[string]any) error {
	chatID, _ := config["chatId"].(string)
	return h.tg.SendMessage(strings.TrimSpace(chatID), "✅ Checkmeup is connected! You'll receive alerts here.")
}

func (h *NotificationChannelHandler) sendTestEmail(config map[string]any) error {
	addr, _ := config["email"].(string)
	return h.mailer.SendTestAlertEmail(strings.TrimSpace(addr))
}

// testSendInternalError wraps a failure that happened on our side (e.g.
// generating a throwaway signing secret) rather than the downstream
// provider, so sendTestNotification can report it as a 500 instead of the
// 502 used for provider failures.
type testSendInternalError struct{ err error }

func (e *testSendInternalError) Error() string { return e.err.Error() }
func (e *testSendInternalError) Unwrap() error { return e.err }

func (h *NotificationChannelHandler) sendTestWebhook(config map[string]any) error {
	url, _ := config["url"].(string)
	secret, err := webhook.GenerateSecret()
	if err != nil {
		return &testSendInternalError{err}
	}
	event := webhook.Event{
		EventType:   "test",
		MonitorName: "Test monitor",
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
	_, err = h.wh.Send(strings.TrimSpace(url), secret, event)
	return err
}

func (h *NotificationChannelHandler) sendTestSlack(config map[string]any) error {
	url, _ := config["url"].(string)
	_, err := h.sl.Send(strings.TrimSpace(url), slack.TestMessage())
	return err
}

func (h *NotificationChannelHandler) sendTestSMS(config map[string]any) error {
	phone, _ := config["phone_number"].(string)
	_, err := h.sm.Send(strings.TrimSpace(phone), "Checkmeup: this is a test SMS alert. You're all set!")
	return err
}

// sendTestNotification dispatches a test message for the given channel type,
// responding with a type-specific error and returning false on failure.
func (h *NotificationChannelHandler) sendTestNotification(w http.ResponseWriter, channelType string, config map[string]any) bool {
	senders := map[string]struct {
		send    func(map[string]any) error
		errCode string
	}{
		"telegram": {h.sendTestTelegram, "telegram_error"},
		"email":    {h.sendTestEmail, "email_error"},
		"webhook":  {h.sendTestWebhook, "webhook_error"},
		"slack":    {h.sendTestSlack, "slack_error"},
		"sms":      {h.sendTestSMS, "sms_error"},
	}
	sender, ok := senders[channelType]
	if !ok {
		return true
	}
	err := sender.send(config)
	if err == nil {
		return true
	}
	var internalErr *testSendInternalError
	if errors.As(err, &internalErr) {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return false
	}
	respond.Error(w, http.StatusBadGateway, err.Error(), sender.errCode)
	return false
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
