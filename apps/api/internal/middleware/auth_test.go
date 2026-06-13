package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/checkmeup/checkmeup/internal/middleware"
)

const jwtSigningKey = "test-signing-key-32-chars-xxxxx"

func okHandler(w http.ResponseWriter, r *http.Request) {
	c := middleware.ClaimsFrom(r.Context())
	_, _ = fmt.Fprintf(w, "userID=%s orgID=%s", c.Subject, c.OrgID)
}

func validToken(t *testing.T, ttl time.Duration) string {
	t.Helper()
	claims := &middleware.Claims{OrgID: "org-abc"}
	claims.Subject = "user-xyz"
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(ttl))
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtSigningKey))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func serve(t *testing.T, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest("GET", "/", nil)
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	middleware.RequireAuth(jwtSigningKey)(http.HandlerFunc(okHandler)).ServeHTTP(w, r)
	return w
}

func TestRequireAuth_NoCookie(t *testing.T) {
	w := serve(t, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	w := serve(t, &http.Cookie{Name: "access_token", Value: validToken(t, 15*time.Minute)})
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	if got := w.Body.String(); got != "userID=user-xyz orgID=org-abc" {
		t.Fatalf("unexpected body: %s", got)
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	w := serve(t, &http.Cookie{Name: "access_token", Value: validToken(t, -time.Minute)})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestRequireAuth_WrongSecret(t *testing.T) {
	claims := &middleware.Claims{OrgID: "org-abc"}
	claims.Subject = "user-xyz"
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(15 * time.Minute))
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("different-signing-key"))
	w := serve(t, &http.Cookie{Name: "access_token", Value: tok})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}
