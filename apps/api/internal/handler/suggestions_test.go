package handler

// Integration tests for SubmitSuggestion. Same conventions as the other
// *_test.go files in this package: real Postgres (ADR-010).
//
// SubmitSuggestion checks auth via apimiddleware.ClaimsFrom directly (like
// Me/AcceptTerms in auth.go), not via the orgIDFrom helper most other
// handlers use — see the orgIDFrom-ownership note in docs/tech-debts.md.
// The per-IP/per-org rate limiting (EP-23: 5/hour per IP, 20/hour per org)
// is wired as chi/httprate middleware in server.go, not in this handler, so
// it's out of scope here — same boundary as the other handler test files.

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
)

func testSuggestionHandler(t *testing.T) (*AuthHandler, *SuggestionHandler, *pgxpool.Pool) {
	t.Helper()
	pool := testPool(t)
	cfg := &config.Config{
		Env:           "development",
		JWTSecret:     testJWTSecret,
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 7 * 24 * time.Hour,
		AppURL:        "http://localhost:5173",
	}
	return NewAuthHandler(cfg, pool), NewSuggestionHandler(cfg, pool), pool
}

func TestSubmitSuggestion(t *testing.T) {
	authH, suggestH, pool := testSuggestionHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/suggestions", bytes.NewReader([]byte(`{"text":"add dark mode"}`)))
		w := httptest.NewRecorder()
		suggestH.SubmitSuggestion(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/suggestions", bytes.NewReader([]byte("not json")))
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(suggestH.SubmitSuggestion)).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("empty text after trimming", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, suggestH.SubmitSuggestion, u.access, submitSuggestionRequest{Text: "   "})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "bad_request" {
			t.Fatalf("want code bad_request, got %q", body["code"])
		}
	})

	t.Run("text over the length limit is rejected", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, suggestH.SubmitSuggestion, u.access, submitSuggestionRequest{
			Text: strings.Repeat("a", maxSuggestionLength+1),
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("text exactly at the length limit is accepted", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, suggestH.SubmitSuggestion, u.access, submitSuggestionRequest{
			Text: strings.Repeat("a", maxSuggestionLength),
		})
		if w.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("success trims whitespace and persists org/user/text", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, suggestH.SubmitSuggestion, u.access, submitSuggestionRequest{
			Text: "  please add dark mode  ",
		})
		if w.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", w.Code, w.Body.String())
		}

		user, err := authH.queries.GetUserByEmail(context.Background(), u.email)
		if err != nil {
			t.Fatalf("lookup user: %v", err)
		}
		var orgID, userID, text string
		if err := pool.QueryRow(context.Background(),
			"SELECT org_id, user_id, text FROM feature_suggestions WHERE user_id = $1", user.ID,
		).Scan(&orgID, &userID, &text); err != nil {
			t.Fatalf("lookup stored suggestion: %v", err)
		}
		if text != "please add dark mode" {
			t.Fatalf("want trimmed text %q, got %q", "please add dark mode", text)
		}
		if orgID != user.OrgID.String() {
			t.Fatalf("want org_id %q, got %q", user.OrgID.String(), orgID)
		}
	})
}
