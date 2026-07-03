package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/google/uuid"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

const apiKeyHeader = "X-API-Key"

type apiKeyCtxKey int

const ctxAPIKeyOrgID apiKeyCtxKey = iota

// RequireAPIKey authenticates non-browser clients (the public API) via the
// X-API-Key header — deliberately separate from RequireAuth's session-cookie
// mechanism (ADR-028), so the two never silently accept each other's tokens.
func RequireAPIKey(queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get(apiKeyHeader)
			if raw == "" {
				respond.Error(w, http.StatusUnauthorized, "API key required", "unauthenticated")
				return
			}

			key, err := queries.GetActiveAPIKeyByHash(r.Context(), hashAPIKey(raw))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					respond.Error(w, http.StatusUnauthorized, "invalid or revoked API key", "invalid_api_key")
					return
				}
				respond.InternalError(w)
				return
			}

			go touchLastUsed(queries, key.ID)

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxAPIKeyOrgID, key.OrgID)))
		})
	}
}

// OrgIDFromAPIKey retrieves the org ID resolved by RequireAPIKey. Returns the
// zero UUID if the request was not authenticated via an API key.
func OrgIDFromAPIKey(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(ctxAPIKeyOrgID).(uuid.UUID)
	return id
}

func hashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// touchLastUsed updates last_used_at off the request path — it must never
// slow down or fail the caller's request.
func touchLastUsed(queries *db.Queries, keyID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = queries.TouchAPIKeyLastUsed(ctx, keyID)
}
