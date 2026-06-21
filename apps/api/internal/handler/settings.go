package handler

import (
	"encoding/json"
	"net/http"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

// SettingsHandler is now just the Telegram bot's incoming webhook (used for
// the /start chat-ID discovery flow). Saving/testing alert channels moved to
// NotificationChannelHandler — see EP-28 / ADR-023.
type SettingsHandler struct {
	cfg *config.Config
	tg  *telegram.Client
}

func NewSettingsHandler(cfg *config.Config, tg *telegram.Client) *SettingsHandler {
	return &SettingsHandler{cfg: cfg, tg: tg}
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
