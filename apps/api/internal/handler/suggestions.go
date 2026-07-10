package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/respond"
)

const maxSuggestionLength = 2000

type SuggestionHandler struct {
	queries  *db.Queries
	notifier *email.FounderNotifier
}

func NewSuggestionHandler(cfg *config.Config, pool *pgxpool.Pool) *SuggestionHandler {
	return &SuggestionHandler{queries: db.New(pool), notifier: email.NewFounderNotifier(cfg.ResendAPIKey)}
}

type submitSuggestionRequest struct {
	Text string `json:"text"`
}

// SubmitSuggestion POST /api/v1/suggestions — stores the suggestion and emails
// the founder directly. No ticket statuses, no public board (EP-23).
func (h *SuggestionHandler) SubmitSuggestion(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req submitSuggestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		respond.Error(w, http.StatusBadRequest, "text is required", "bad_request")
		return
	}
	if len(req.Text) > maxSuggestionLength {
		respond.Error(w, http.StatusBadRequest, "text must be 2000 characters or fewer", "bad_request")
		return
	}

	if _, err := h.queries.CreateFeatureSuggestion(r.Context(), db.CreateFeatureSuggestionParams{
		OrgID:  orgID,
		UserID: userID,
		Text:   req.Text,
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	// Best-effort from here — the suggestion is already stored, so a lookup or
	// email failure shouldn't turn into an error response for the user.
	fromEmail := "unknown"
	if user, err := h.queries.GetUserByID(r.Context(), userID); err == nil {
		fromEmail = user.Email
	}
	_ = h.notifier.SendFeatureSuggestion(fromEmail, req.Text)

	w.WriteHeader(http.StatusNoContent)
}
