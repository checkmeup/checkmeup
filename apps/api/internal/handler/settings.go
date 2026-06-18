package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/respond"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

type SettingsHandler struct {
	cfg     *config.Config
	queries *db.Queries
	tg      *telegram.Client
	mailer  *email.Sender
}

func NewSettingsHandler(cfg *config.Config, pool *pgxpool.Pool, tg *telegram.Client, mailer *email.Sender) *SettingsHandler {
	return &SettingsHandler{cfg: cfg, queries: db.New(pool), tg: tg, mailer: mailer}
}

type settingsResponse struct {
	TelegramChatID     *string `json:"telegramChatId"`
	AlertEmail         *string `json:"alertEmail"`
	EmailAlertsEnabled bool    `json:"emailAlertsEnabled"`
}

func toSettingsResponse(org db.Org) settingsResponse {
	resp := settingsResponse{EmailAlertsEnabled: org.EmailAlertsEnabled}
	if org.TelegramChatID.Valid {
		resp.TelegramChatID = &org.TelegramChatID.String
	}
	if org.AlertEmail.Valid {
		resp.AlertEmail = &org.AlertEmail.String
	}
	return resp
}

// GetSettings GET /api/v1/settings
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	org, err := h.queries.GetOrgByID(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	respond.JSON(w, http.StatusOK, toSettingsResponse(org))
}

type saveTelegramRequest struct {
	ChatID string `json:"chatId"`
}

// SaveTelegram PUT /api/v1/settings/telegram
func (h *SettingsHandler) SaveTelegram(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req saveTelegramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	req.ChatID = strings.TrimSpace(req.ChatID)

	org, err := h.queries.UpdateOrgTelegramChatID(r.Context(), db.UpdateOrgTelegramChatIDParams{
		ID:             orgID,
		TelegramChatID: pgtype.Text{String: req.ChatID, Valid: req.ChatID != ""},
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	respond.JSON(w, http.StatusOK, toSettingsResponse(org))
}

type saveEmailRequest struct {
	Email string `json:"email"`
}

// SaveEmail PUT /api/v1/settings/email
func (h *SettingsHandler) SaveEmail(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req saveEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	req.Email = strings.TrimSpace(req.Email)

	org, err := h.queries.UpdateOrgAlertEmail(r.Context(), db.UpdateOrgAlertEmailParams{
		ID:         orgID,
		AlertEmail: pgtype.Text{String: req.Email, Valid: req.Email != ""},
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	respond.JSON(w, http.StatusOK, toSettingsResponse(org))
}

type setEmailAlertsEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

// SetEmailAlertsEnabled PUT /api/v1/settings/email/enabled
func (h *SettingsHandler) SetEmailAlertsEnabled(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req setEmailAlertsEnabledRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	org, err := h.queries.UpdateOrgEmailAlertsEnabled(r.Context(), db.UpdateOrgEmailAlertsEnabledParams{
		ID:                 orgID,
		EmailAlertsEnabled: req.Enabled,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	respond.JSON(w, http.StatusOK, toSettingsResponse(org))
}

// HandleTelegramWebhook POST /webhook/telegram  (no auth — called by Telegram's servers)
func (h *SettingsHandler) HandleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
	// Reject requests that don't carry the expected secret token.
	// The token is set via setWebhook at startup; legitimate Telegram calls always include it.
	if secret := h.cfg.TelegramWebhookSecret; secret != "" {
		if r.Header.Get("X-Telegram-Bot-Api-Secret-Token") != secret {
			w.WriteHeader(http.StatusOK) // 200 so random scanners don't learn the path exists
			return
		}
	}

	var update telegram.WebhookUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		w.WriteHeader(http.StatusOK) // always 200 so Telegram doesn't retry
		return
	}
	h.tg.HandleUpdate(update)
	w.WriteHeader(http.StatusOK)
}

type testTelegramRequest struct {
	ChatID string `json:"chatId"`
}

// TestTelegram POST /api/v1/settings/telegram/test
func (h *SettingsHandler) TestTelegram(w http.ResponseWriter, r *http.Request) {
	var req testTelegramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	req.ChatID = strings.TrimSpace(req.ChatID)
	if req.ChatID == "" {
		respond.Error(w, http.StatusBadRequest, "chatId is required", "bad_request")
		return
	}

	if err := h.tg.SendMessage(req.ChatID, "✅ checkmeup is connected! You'll receive alerts here."); err != nil {
		respond.Error(w, http.StatusBadGateway, err.Error(), "telegram_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type testEmailRequest struct {
	Email string `json:"email"`
}

// TestEmail POST /api/v1/settings/email/test
func (h *SettingsHandler) TestEmail(w http.ResponseWriter, r *http.Request) {
	var req testEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		respond.Error(w, http.StatusBadRequest, "email is required", "bad_request")
		return
	}

	if err := h.mailer.SendTestAlertEmail(req.Email); err != nil {
		respond.Error(w, http.StatusBadGateway, err.Error(), "email_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
