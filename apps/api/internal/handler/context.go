package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
)

// orgIDFrom and userIDFrom pull the authenticated org/user ID out of the
// request's JWT claims (apimiddleware.RequireAuth populates the context).
// Shared across the whole package rather than reimplemented per handler —
// suggestions.go used to parse claims.Subject inline instead of reusing
// this.

func orgIDFrom(r *http.Request) (uuid.UUID, error) {
	claims := apimiddleware.ClaimsFrom(r.Context())
	if claims == nil {
		return uuid.UUID{}, errors.New("no claims")
	}
	return uuid.Parse(claims.OrgID)
}

func userIDFrom(r *http.Request) (uuid.UUID, error) {
	claims := apimiddleware.ClaimsFrom(r.Context())
	if claims == nil {
		return uuid.UUID{}, errors.New("no claims")
	}
	return uuid.Parse(claims.Subject)
}
