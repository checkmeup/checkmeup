package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
	"github.com/checkmeup/checkmeup/internal/respond"
)

const maxSuggestionLength = 2000

type SuggestionHandler struct {
	queries *db.Queries
	mailer  *email.Sender
}

func NewSuggestionHandler(cfg *config.Config, pool *pgxpool.Pool) *SuggestionHandler {
	return &SuggestionHandler{queries: db.New(pool), mailer: email.NewSender(cfg.ResendAPIKey)}
}

type submitSuggestionRequest struct {
	Text string `json:"text"`
}

// SubmitSuggestion POST /api/v1/suggestions — stores the suggestion and emails
// the founder directly. No ticket statuses, no public board (EP-23).
func (h *SuggestionHandler) SubmitSuggestion(w http.ResponseWriter, r *http.Request) {
	claims := apimiddleware.ClaimsFrom(r.Context())
	if claims == nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid token", "invalid_token")
		return
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "invalid token", "invalid_token")
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
	_ = h.mailer.SendFeatureSuggestion(fromEmail, req.Text)

	w.WriteHeader(http.StatusNoContent)
}
