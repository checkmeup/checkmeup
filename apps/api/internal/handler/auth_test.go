package handler

// Integration tests for the auth handlers. Per ADR-010, these hit a real
// PostgreSQL instance (no DB mocks, to avoid hiding org_id-style bugs) —
// DATABASE_URL must point at a reachable Postgres with migrations applied
// (docker-compose db service locally, the CI service container in GitHub
// Actions). This file is `package handler` (not `handler_test`) so it can
// reuse the unexported hashToken helper to seed password-reset tokens
// directly, the same way ForgotPassword does internally.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/legal"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
)

const testJWTSecret = "test-jwt-secret-32-chars-minimum-xxx"

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://checkmeup:checkmeup@db:5432/checkmeup?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping test db: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func testAuthHandler(t *testing.T) (*AuthHandler, *pgxpool.Pool) {
	t.Helper()
	pool := testPool(t)
	cfg := &config.Config{
		Env:           "development",
		JWTSecret:     testJWTSecret,
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 7 * 24 * time.Hour,
		AppURL:        "http://localhost:5173",
	}
	return NewAuthHandler(cfg, pool), pool
}

func testEmail(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("test-%s@example.com", uuid.NewString())
}

func doJSON(t *testing.T, handler http.HandlerFunc, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader = http.NoBody
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func decodeBody[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(w.Body).Decode(&v); err != nil {
		t.Fatalf("decode response body: %v (body=%s)", err, w.Body.String())
	}
	return v
}

func findCookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// signedUpUser holds the result of signUpTestUser — the decoded response and
// the cookies issued, since httptest.ResponseRecorder's body can only be
// decoded once (it's a drained bytes.Buffer after the first read).
type signedUpUser struct {
	resp     userResponse
	email    string
	password string
	access   *http.Cookie
	refresh  *http.Cookie
}

// signUpTestUser signs up a fresh user with a unique email and registers
// cleanup of the org (cascades to users/refresh_tokens/password_reset_tokens)
// once the test completes.
func signUpTestUser(t *testing.T, h *AuthHandler, pool *pgxpool.Pool) signedUpUser {
	t.Helper()
	email := testEmail(t)
	password := "correct-horse-battery-staple"
	w := doJSON(t, h.SignUp, http.MethodPost, "/api/v1/auth/sign-up", signUpRequest{
		Email:         email,
		Password:      password,
		AcceptedTerms: true,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("sign up failed: %d %s", w.Code, w.Body.String())
	}
	resp := decodeBody[userResponse](t, w)
	orgID := uuid.MustParse(resp.OrgID)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM orgs WHERE id = $1", orgID)
	})
	return signedUpUser{
		resp:     resp,
		email:    email,
		password: password,
		access:   findCookie(w, "access_token"),
		refresh:  findCookie(w, "refresh_token"),
	}
}

func TestSignUp(t *testing.T) {
	h, pool := testAuthHandler(t)

	t.Run("success sets cookies and persists user", func(t *testing.T) {
		u := signUpTestUser(t, h, pool)
		if u.resp.Email != u.email {
			t.Fatalf("want email %q, got %q", u.email, u.resp.Email)
		}
		if u.resp.NeedsTermsAcceptance {
			t.Fatal("fresh sign-up should not need terms acceptance")
		}
		if u.access == nil || !u.access.HttpOnly {
			t.Fatal("expected an HttpOnly access_token cookie")
		}
		if u.refresh == nil || !u.refresh.HttpOnly {
			t.Fatal("expected an HttpOnly refresh_token cookie")
		}

		user, err := h.queries.GetUserByEmail(context.Background(), u.email)
		if err != nil {
			t.Fatalf("user not persisted: %v", err)
		}
		if user.Email != u.email {
			t.Fatalf("persisted email mismatch: got %q", user.Email)
		}
	})

	t.Run("duplicate email returns 409 and leaves no orphan org", func(t *testing.T) {
		u := signUpTestUser(t, h, pool)
		w := doJSON(t, h.SignUp, http.MethodPost, "/api/v1/auth/sign-up", signUpRequest{
			Email: u.email, Password: u.password, AcceptedTerms: true,
		})
		if w.Code != http.StatusConflict {
			t.Fatalf("want 409, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "email_taken" {
			t.Fatalf("want code email_taken, got %q", body["code"])
		}

		// CreateOrg + CreateUser run in one transaction (auth.go SignUp) —
		// a failed duplicate-email attempt must not leave an orphan org row.
		orgName := strings.SplitN(u.email, "@", 2)[0]
		var orgCount int
		if err := pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM orgs WHERE name = $1", orgName,
		).Scan(&orgCount); err != nil {
			t.Fatalf("count orgs: %v", err)
		}
		if orgCount != 1 {
			t.Fatalf("want exactly 1 org for %q (no orphan from the failed attempt), got %d", orgName, orgCount)
		}
	})

	t.Run("validation", func(t *testing.T) {
		cases := []struct {
			name string
			req  signUpRequest
		}{
			{"missing email", signUpRequest{Password: "longenough1", AcceptedTerms: true}},
			{"missing password", signUpRequest{Email: testEmail(t), AcceptedTerms: true}},
			{"short password", signUpRequest{Email: testEmail(t), Password: "short", AcceptedTerms: true}},
			{"terms not accepted", signUpRequest{Email: testEmail(t), Password: "longenough1"}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := doJSON(t, h.SignUp, http.MethodPost, "/api/v1/auth/sign-up", tc.req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
				}
			})
		}
	})
}

func TestSignIn(t *testing.T) {
	h, pool := testAuthHandler(t)

	t.Run("success", func(t *testing.T) {
		u := signUpTestUser(t, h, pool)
		w := doJSON(t, h.SignIn, http.MethodPost, "/api/v1/auth/sign-in", signInRequest{
			Email: u.email, Password: u.password,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		if findCookie(w, "access_token") == nil {
			t.Fatal("expected access_token cookie")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		u := signUpTestUser(t, h, pool)
		w := doJSON(t, h.SignIn, http.MethodPost, "/api/v1/auth/sign-in", signInRequest{
			Email: u.email, Password: "definitely-wrong",
		})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "invalid_credentials" {
			t.Fatalf("want code invalid_credentials, got %q", body["code"])
		}
	})

	t.Run("unknown email", func(t *testing.T) {
		w := doJSON(t, h.SignIn, http.MethodPost, "/api/v1/auth/sign-in", signInRequest{
			Email: testEmail(t), Password: "whatever-12345",
		})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestRefresh(t *testing.T) {
	h, pool := testAuthHandler(t)

	t.Run("rotates the refresh token and rejects reuse", func(t *testing.T) {
		u := signUpTestUser(t, h, pool)
		oldRefresh := u.refresh

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req.AddCookie(oldRefresh)
		w := httptest.NewRecorder()
		h.Refresh(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		newRefresh := findCookie(w, "refresh_token")
		if newRefresh == nil || newRefresh.Value == oldRefresh.Value {
			t.Fatal("expected a newly rotated refresh_token cookie")
		}

		// Reusing the old (now-rotated) refresh token must fail.
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req2.AddCookie(oldRefresh)
		w2 := httptest.NewRecorder()
		h.Refresh(w2, req2)
		if w2.Code != http.StatusUnauthorized {
			t.Fatalf("want 401 reusing a rotated refresh token, got %d", w2.Code)
		}
	})

	t.Run("missing cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		w := httptest.NewRecorder()
		h.Refresh(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("unknown token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "not-a-real-token"})
		w := httptest.NewRecorder()
		h.Refresh(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
}

func TestSignOut(t *testing.T) {
	h, pool := testAuthHandler(t)
	u := signUpTestUser(t, h, pool)
	refresh := u.refresh

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sign-out", nil)
	req.AddCookie(refresh)
	w := httptest.NewRecorder()
	h.SignOut(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", w.Code)
	}
	cleared := findCookie(w, "refresh_token")
	if cleared == nil || cleared.MaxAge >= 0 {
		t.Fatal("expected refresh_token cookie to be cleared (MaxAge < 0)")
	}

	// The refresh token must be invalidated server-side, not just the cookie.
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req2.AddCookie(refresh)
	w2 := httptest.NewRecorder()
	h.Refresh(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("signed-out refresh token should be rejected, got %d", w2.Code)
	}
}

func TestMe(t *testing.T) {
	h, pool := testAuthHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		w := httptest.NewRecorder()
		h.Me(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("authenticated", func(t *testing.T) {
		u := signUpTestUser(t, h, pool)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.AddCookie(u.access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(h.Me)).ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[userResponse](t, w)
		if resp.Email != u.email {
			t.Fatalf("want email %q, got %q", u.email, resp.Email)
		}
	})
}

func TestAcceptTerms(t *testing.T) {
	h, pool := testAuthHandler(t)

	t.Run("requires auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/accept-terms", nil)
		w := httptest.NewRecorder()
		h.AcceptTerms(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("re-accepting a stale version clears NeedsTermsAcceptance", func(t *testing.T) {
		u := signUpTestUser(t, h, pool)
		access := u.access
		userID := uuid.MustParse(u.resp.ID)

		// Simulate a user who accepted an older Terms/Privacy version.
		if _, err := pool.Exec(context.Background(),
			"UPDATE users SET terms_version = $1 WHERE id = $2", "2020-01-01", userID,
		); err != nil {
			t.Fatalf("seed stale terms version: %v", err)
		}

		meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		meReq.AddCookie(access)
		meW := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(h.Me)).ServeHTTP(meW, meReq)
		stale := decodeBody[userResponse](t, meW)
		if !stale.NeedsTermsAcceptance {
			t.Fatal("expected a stale terms_version to need re-acceptance")
		}

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/accept-terms", nil)
		req.AddCookie(access)
		w := httptest.NewRecorder()
		apimiddleware.RequireAuth(testJWTSecret)(http.HandlerFunc(h.AcceptTerms)).ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		after := decodeBody[userResponse](t, w)
		if after.NeedsTermsAcceptance {
			t.Fatal("expected NeedsTermsAcceptance to clear after accepting")
		}
		if after.TermsVersion == nil || *after.TermsVersion != legal.CurrentVersion {
			t.Fatalf("want terms version %q, got %v", legal.CurrentVersion, after.TermsVersion)
		}
	})
}

func TestForgotPassword(t *testing.T) {
	h, pool := testAuthHandler(t)

	t.Run("invalid body returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader([]byte("not json")))
		w := httptest.NewRecorder()
		h.ForgotPassword(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})

	t.Run("unknown email still returns 204 (no enumeration)", func(t *testing.T) {
		w := doJSON(t, h.ForgotPassword, http.MethodPost, "/api/v1/auth/forgot-password", forgotPasswordRequest{
			Email: testEmail(t),
		})
		if w.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d", w.Code)
		}
	})

	t.Run("known email returns 204 and creates a reset token", func(t *testing.T) {
		u := signUpTestUser(t, h, pool)
		w := doJSON(t, h.ForgotPassword, http.MethodPost, "/api/v1/auth/forgot-password", forgotPasswordRequest{
			Email: u.email,
		})
		if w.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d", w.Code)
		}

		user, err := h.queries.GetUserByEmail(context.Background(), u.email)
		if err != nil {
			t.Fatalf("lookup user: %v", err)
		}
		var count int
		if err := pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM password_reset_tokens WHERE user_id = $1", user.ID,
		).Scan(&count); err != nil {
			t.Fatalf("count reset tokens: %v", err)
		}
		if count != 1 {
			t.Fatalf("want 1 reset token, got %d", count)
		}
	})
}

func TestResetPassword(t *testing.T) {
	h, pool := testAuthHandler(t)

	t.Run("rotates password and revokes existing sessions", func(t *testing.T) {
		u := signUpTestUser(t, h, pool)
		email, oldPassword, refresh := u.email, u.password, u.refresh

		user, err := h.queries.GetUserByEmail(context.Background(), email)
		if err != nil {
			t.Fatalf("lookup user: %v", err)
		}

		// Seed a reset token the same way ForgotPassword does, so we know
		// the raw value (ForgotPassword only ever emails it).
		rawToken := "test-raw-reset-token-" + uuid.NewString()
		if _, err := h.queries.CreatePasswordResetToken(context.Background(), db.CreatePasswordResetTokenParams{
			UserID:    user.ID,
			TokenHash: hashToken(rawToken),
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
		}); err != nil {
			t.Fatalf("seed reset token: %v", err)
		}

		newPassword := "brand-new-password-1"
		w := doJSON(t, h.ResetPassword, http.MethodPost, "/api/v1/auth/reset-password", resetPasswordRequest{
			Token: rawToken, Password: newPassword,
		})
		if w.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d: %s", w.Code, w.Body.String())
		}

		oldW := doJSON(t, h.SignIn, http.MethodPost, "/api/v1/auth/sign-in", signInRequest{
			Email: email, Password: oldPassword,
		})
		if oldW.Code != http.StatusUnauthorized {
			t.Fatalf("old password should be rejected after reset, got %d", oldW.Code)
		}

		newW := doJSON(t, h.SignIn, http.MethodPost, "/api/v1/auth/sign-in", signInRequest{
			Email: email, Password: newPassword,
		})
		if newW.Code != http.StatusOK {
			t.Fatalf("new password should sign in, got %d: %s", newW.Code, newW.Body.String())
		}

		// A password reset must revoke pre-existing sessions.
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req.AddCookie(refresh)
		refreshW := httptest.NewRecorder()
		h.Refresh(refreshW, req)
		if refreshW.Code != http.StatusUnauthorized {
			t.Fatalf("pre-reset refresh token should be revoked, got %d", refreshW.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		w := doJSON(t, h.ResetPassword, http.MethodPost, "/api/v1/auth/reset-password", resetPasswordRequest{
			Token: "does-not-exist", Password: "longenoughpassword1",
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("short password rejected", func(t *testing.T) {
		w := doJSON(t, h.ResetPassword, http.MethodPost, "/api/v1/auth/reset-password", resetPasswordRequest{
			Token: "whatever-token", Password: "short",
		})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestValidateSignUp(t *testing.T) {
	cases := []struct {
		name string
		req  signUpRequest
		want string
	}{
		{"valid", signUpRequest{Email: "a@b.com", Password: "longenough1", AcceptedTerms: true}, ""},
		{"missing email", signUpRequest{Password: "longenough1", AcceptedTerms: true}, "email and password are required"},
		{"missing password", signUpRequest{Email: "a@b.com", AcceptedTerms: true}, "email and password are required"},
		{"terms not accepted", signUpRequest{Email: "a@b.com", Password: "longenough1"}, "you must accept the Terms of Service and Privacy Policy"},
		{"short password", signUpRequest{Email: "a@b.com", Password: "short", AcceptedTerms: true}, "password must be at least 8 characters"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := validateSignUp(tc.req); got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestHashToken(t *testing.T) {
	a := hashToken("raw-value-1")
	b := hashToken("raw-value-1")
	c := hashToken("raw-value-2")
	if a != b {
		t.Fatal("hashToken must be deterministic for the same input")
	}
	if a == c {
		t.Fatal("hashToken must differ for different inputs")
	}
}

func TestIsUniqueViolation(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{fmt.Errorf("ERROR: duplicate key value violates unique constraint \"users_email_key\""), true},
		{fmt.Errorf("ERROR: unique constraint violated"), true},
		{fmt.Errorf("connection refused"), false},
	}
	for _, tc := range cases {
		if got := isUniqueViolation(tc.err); got != tc.want {
			t.Fatalf("isUniqueViolation(%q) = %v, want %v", tc.err, got, tc.want)
		}
	}
}
