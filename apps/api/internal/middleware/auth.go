package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	"github.com/checkmeup/checkmeup/internal/respond"
)

const accessCookie = "access_token"

type ctxKey int

const (
	ctxClaims ctxKey = iota
)

type Claims struct {
	OrgID string `json:"org"`
	jwt.RegisteredClaims
}

// RequireAuth validates the JWT access token from the httpOnly cookie.
// On success it stores *Claims in the request context; on failure returns 401.
func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(accessCookie)
			if err != nil {
				respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			}, jwt.WithExpirationRequired())

			if err != nil || !token.Valid {
				respond.Error(w, http.StatusUnauthorized, "invalid or expired token", "invalid_token")
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxClaims, claims)))
		})
	}
}

// ClaimsFrom retrieves the authenticated claims from the request context.
// Returns nil if the request was not authenticated.
func ClaimsFrom(ctx context.Context) *Claims {
	c, _ := ctx.Value(ctxClaims).(*Claims)
	return c
}
