package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

const apiKeyPrefixLen = 16 // "cmu_live_" (9 chars) + 7 hex chars, enough to recognize in logs

// maxAPIKeys caps how many active API keys an org can have at once — a flat
// safety cap, uniform across every plan.
const maxAPIKeys = 100

type APIKeyHandler struct {
	queries *db.Queries
}

func NewAPIKeyHandler(pool *pgxpool.Pool) *APIKeyHandler {
	return &APIKeyHandler{queries: db.New(pool)}
}

type apiKeyResponse struct {
	ID         string  `json:"id"`
	Label      string  `json:"label"`
	KeyPrefix  string  `json:"keyPrefix"`
	CreatedAt  string  `json:"createdAt"`
	LastUsedAt *string `json:"lastUsedAt"`
}

func apiKeyToResponse(k db.ApiKey) apiKeyResponse {
	r := apiKeyResponse{
		ID:        k.ID.String(),
		Label:     k.Label,
		KeyPrefix: k.KeyPrefix,
		CreatedAt: k.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if k.LastUsedAt.Valid {
		t := k.LastUsedAt.Time.Format("2006-01-02T15:04:05Z")
		r.LastUsedAt = &t
	}
	return r
}

type createAPIKeyRequest struct {
	Label string `json:"label"`
}

type createAPIKeyResponse struct {
	apiKeyResponse
	Key string `json:"key"`
}

// CreateAPIKey generates a new public-API key for the caller's org. The raw
// key is returned once, here, and never again — only its hash is stored,
// same one-way pattern as password_hash / refresh_tokens.token_hash.
func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req createAPIKeyRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	count, err := h.queries.CountActiveAPIKeys(r.Context(), orgID)
	if err != nil {
		respond.InternalError(w)
		return
	}
	if count >= maxAPIKeys {
		respond.Error(w, http.StatusConflict,
			"too many active API keys — revoke one before creating more",
			"too_many_api_keys")
		return
	}

	raw, err := generateAPIKey()
	if err != nil {
		respond.InternalError(w)
		return
	}

	key, err := h.queries.CreateAPIKey(r.Context(), db.CreateAPIKeyParams{
		OrgID:     orgID,
		KeyHash:   hashToken(raw),
		KeyPrefix: raw[:apiKeyPrefixLen],
		Label:     req.Label,
	})
	if err != nil {
		respond.InternalError(w)
		return
	}

	respond.JSON(w, http.StatusCreated, createAPIKeyResponse{apiKeyToResponse(key), raw})
}

// ListAPIKeys returns active (non-revoked) keys for the caller's org, never
// including the raw key value — only what was shown once at creation.
func (h *APIKeyHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	keys, err := h.queries.ListAPIKeys(r.Context(), orgID)
	if err != nil {
		respond.InternalError(w)
		return
	}

	resp := make([]apiKeyResponse, len(keys))
	for i, k := range keys {
		resp[i] = apiKeyToResponse(k)
	}
	respond.JSON(w, http.StatusOK, resp)
}

// RevokeAPIKey takes effect immediately — the next request bearing this key
// fails auth, since RequireAPIKey only matches non-revoked keys.
func (h *APIKeyHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	keyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid key id", "bad_request")
		return
	}

	if err := h.queries.RevokeAPIKey(r.Context(), db.RevokeAPIKeyParams{ID: keyID, OrgID: orgID}); err != nil {
		respond.InternalError(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "cmu_live_" + hex.EncodeToString(b), nil
}
