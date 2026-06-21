package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/respond"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

type NotificationChannelHandler struct {
	queries *db.Queries
	tg      *telegram.Client
	mailer  *email.Sender
}

func NewNotificationChannelHandler(pool *pgxpool.Pool, tg *telegram.Client, mailer *email.Sender) *NotificationChannelHandler {
	return &NotificationChannelHandler{queries: db.New(pool), tg: tg, mailer: mailer}
}

type notificationChannelResponse struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Name      string         `json:"name"`
	Config    map[string]any `json:"config"`
	Enabled   bool           `json:"enabled"`
	CreatedAt string         `json:"createdAt"`
}

func toNotificationChannelResponse(c db.NotificationChannel) notificationChannelResponse {
	var cfg map[string]any
	_ = json.Unmarshal(c.Config, &cfg)
	return notificationChannelResponse{
		ID:        c.ID.String(),
		Type:      string(c.Type),
		Name:      c.Name,
		Config:    cfg,
		Enabled:   c.Enabled,
		CreatedAt: c.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

// validateChannelConfig checks config has the field(s) required for type.
// Only telegram/email are supported today — other notification_channel_type
// values get added (their own migration + their own case here) when their
// epic actually ships (ADR-023).
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
	default:
		return errors.New("unsupported channel type")
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

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}
	if err := validateChannelConfig(req.Type, req.Config); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
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
	// A channel's type never changes after creation — only name/config/enabled.
	if err := validateChannelConfig(string(existing.Type), req.Config); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	configBytes, err := json.Marshal(req.Config)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid config", "bad_request")
		return
	}
	enabled := existing.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
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

	switch req.Type {
	case "telegram":
		chatID, _ := req.Config["chatId"].(string)
		if err := h.tg.SendMessage(strings.TrimSpace(chatID), "✅ checkmeup is connected! You'll receive alerts here."); err != nil {
			respond.Error(w, http.StatusBadGateway, err.Error(), "telegram_error")
			return
		}
	case "email":
		addr, _ := req.Config["email"].(string)
		if err := h.mailer.SendTestAlertEmail(strings.TrimSpace(addr)); err != nil {
			respond.Error(w, http.StatusBadGateway, err.Error(), "email_error")
			return
		}
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
