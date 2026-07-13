package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
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

	if err := h.checkNotificationChannelCreateLimits(r.Context(), orgID, req.Type); err != nil {
		if errors.Is(err, billing.ErrNotificationChannelLimit) || errors.Is(err, errSMSRequiresPaidPlan) {
			respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	configBytes, err := buildNotificationChannelConfig(&req)
	if err != nil {
		if errors.Is(err, errWebhookSecretGeneration) {
			respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
			return
		}
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
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

// errSMSRequiresPaidPlan is returned when an org on a plan with 0 monthly
// SMS credits (Hobby, per ADR-032) tries to create an sms channel — no
// point saving one that can never send.
var errSMSRequiresPaidPlan = errors.New("SMS alerts require a paid plan — upgrade to enable this channel")

func (h *NotificationChannelHandler) checkNotificationChannelCreateLimits(ctx context.Context, orgID uuid.UUID, channelType string) error {
	plan, err := h.queries.GetOrgPlan(ctx, orgID)
	if err != nil {
		return err
	}
	total, err := h.queries.CountOrgNotificationChannels(ctx, orgID)
	if err != nil {
		return err
	}
	if err := billing.CheckNotificationChannelLimit(plan, int(total)); err != nil {
		return err
	}
	if channelType == "sms" && billing.GetLimits(plan).SMSCredits <= 0 {
		return errSMSRequiresPaidPlan
	}
	return nil
}

// errWebhookSecretGeneration wraps a webhook.GenerateSecret failure so the
// caller can distinguish it (an internal/crypto failure, 500) from the
// validation-shaped errors buildNotificationChannelConfig otherwise returns
// (400) — see the errors.Is check at the CreateNotificationChannel call site.
var errWebhookSecretGeneration = errors.New("failed to generate webhook secret")

// buildNotificationChannelConfig validates req.Config against req.Type,
// fills in server-generated fields (webhook secret, SMS consent timestamp),
// and marshals the result for storage.
func buildNotificationChannelConfig(req *notificationChannelRequest) ([]byte, error) {
	if err := validateChannelConfig(req.Type, req.Config); err != nil {
		return nil, err
	}
	if req.Type == "webhook" {
		secret, err := webhook.GenerateSecret()
		if err != nil {
			return nil, errWebhookSecretGeneration
		}
		req.Config["secret"] = secret
	}
	if req.Type == "sms" {
		finalizeSMSConsent(req.Config)
	}
	configBytes, err := json.Marshal(req.Config)
	if err != nil {
		return nil, errors.New("invalid config")
	}
	return configBytes, nil
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
		respondChannelLookupErr(w, err)
		return
	}
	configBytes, enabled, err := resolveUpdatedChannelConfig(existing, req)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	if err := h.checkReEnableLimitIfNeeded(r.Context(), orgID, existing.Enabled, enabled); err != nil {
		respondNotificationChannelLimitErr(w, err)
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

func respondChannelLookupErr(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		respond.Error(w, http.StatusNotFound, "channel not found", "not_found")
		return
	}
	respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
}

func respondNotificationChannelLimitErr(w http.ResponseWriter, err error) {
	if errors.Is(err, billing.ErrNotificationChannelLimit) {
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return
	}
	respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
}

// checkReEnableLimitIfNeeded checks the plan's enabled-channel cap only when
// this update actually re-enables a previously-disabled channel — that
// grows the org's active channel count, same as creating one, gated the
// same way (ADR-019). A disabled channel doesn't count against the limit at
// all, so no other transition needs this check.
func (h *NotificationChannelHandler) checkReEnableLimitIfNeeded(ctx context.Context, orgID uuid.UUID, wasEnabled, nowEnabled bool) error {
	if !nowEnabled || wasEnabled {
		return nil
	}
	plan, err := h.queries.GetOrgPlan(ctx, orgID)
	if err != nil {
		return err
	}
	enabledCount, err := h.queries.CountEnabledNotificationChannelsForOrg(ctx, orgID)
	if err != nil {
		return err
	}
	return billing.CheckNotificationChannelLimit(plan, int(enabledCount))
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
