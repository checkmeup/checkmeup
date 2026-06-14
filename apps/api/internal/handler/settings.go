package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

type SettingsHandler struct {
	cfg     *config.Config
	queries *db.Queries
	tg      *telegram.Client
}

func NewSettingsHandler(cfg *config.Config, pool *pgxpool.Pool, tg *telegram.Client) *SettingsHandler {
	return &SettingsHandler{cfg: cfg, queries: db.New(pool), tg: tg}
}

type settingsResponse struct {
	TelegramChatID *string `json:"telegramChatId"`
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

	resp := settingsResponse{}
	if org.TelegramChatID.Valid {
		resp.TelegramChatID = &org.TelegramChatID.String
	}
	respond.JSON(w, http.StatusOK, resp)
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

	resp := settingsResponse{}
	if org.TelegramChatID.Valid {
		resp.TelegramChatID = &org.TelegramChatID.String
	}
	respond.JSON(w, http.StatusOK, resp)
}

// HandleTelegramWebhook POST /webhook/telegram  (no auth — called by Telegram's servers)
func (h *SettingsHandler) HandleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
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
