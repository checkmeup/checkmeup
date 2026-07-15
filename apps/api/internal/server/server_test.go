package server

// Tests for server.go. The router-wiring tests (TestRoutes_*, TestHandle*)
// hit a real PostgreSQL instance per ADR-010 — DATABASE_URL must point at a
// reachable Postgres with migrations applied (docker-compose db service
// locally, the CI service container in GitHub Actions) — because New()
// constructs every handler in the tree, several of which take a *pgxpool.Pool.
// The pure-logic helpers (handleSPA, requestLogger, authOrgKey,
// suggestionRateLimited) are exercised directly against a bare *Server
// without a pool, since none of them touch the database.
//
// package server (not server_test) so it can call the unexported
// handleHealth/handleSPA/authOrgKey/suggestionRateLimited/
// requestLogger directly instead of only through the router's public surface.

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
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

func testLoggerToBuf() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewTextHandler(&buf, nil)), &buf
}

// testServer builds a fully-wired Server backed by a real test DB pool, for
// exercising router wiring end to end.
func testServer(t *testing.T, cfg *config.Config) (*Server, *pgxpool.Pool) {
	t.Helper()
	pool := testPool(t)
	if cfg == nil {
		cfg = &config.Config{}
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = testJWTSecret
	}
	logger, _ := testLoggerToBuf()
	return New(cfg, logger, pool, "test-version"), pool
}

// ─── handleHealth ────────────────────────────────────────────────────────────

func TestHandleHealth_OK(t *testing.T) {
	srv, _ := testServer(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", http.NoBody)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", w.Code, http.StatusOK, w.Body.String())
	}

	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got["status"] != "ok" || got["version"] != "test-version" {
		t.Fatalf("body = %v, want status=ok version=test-version", got)
	}
}

func TestHandleHealth_DBUnavailable(t *testing.T) {
	srv, pool := testServer(t, nil)
	pool.Close() // closed pool makes Ping fail without needing a real outage

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", http.NoBody)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got["code"] != "db_unavailable" {
		t.Fatalf("code = %q, want db_unavailable", got["code"])
	}
}

// ─── routing wiring ──────────────────────────────────────────────────────────

func TestRoutes_ProtectedEndpointRequiresAuth(t *testing.T) {
	srv, _ := testServer(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", http.NoBody)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRoutes_PublicHealthDoesNotRequireAuth(t *testing.T) {
	srv, _ := testServer(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", http.NoBody)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatalf("status = %d, /health should not require auth", w.Code)
	}
}

func TestRoutes_UnknownAPIPathIs404(t *testing.T) {
	srv, _ := testServer(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/does-not-exist", http.NoBody)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRoutes_NoStaticDirMeansNoSPAFallback(t *testing.T) {
	srv, _ := testServer(t, &config.Config{StaticDir: ""})

	req := httptest.NewRequest(http.MethodGet, "/some/client/route", http.NoBody)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d (no StaticDir means no catch-all registered)", w.Code, http.StatusNotFound)
	}
}

func TestRoutes_SPAResponseIsCompressedWhenAccepted(t *testing.T) {
	dir := t.TempDir()
	// Large enough to clear chi's compress middleware minimum size threshold.
	body := bytes.Repeat([]byte("a"), 4096)
	if err := os.WriteFile(filepath.Join(dir, "app.js"), body, 0o600); err != nil {
		t.Fatalf("write app.js: %v", err)
	}

	srv, _ := testServer(t, &config.Config{StaticDir: dir})

	req := httptest.NewRequest(http.MethodGet, "/app.js", http.NoBody)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if got, want := w.Header().Get("Content-Encoding"), "gzip"; got != want {
		t.Fatalf("Content-Encoding = %q, want %q", got, want)
	}
	if w.Body.Len() >= len(body) {
		t.Fatalf("compressed body len = %d, want < uncompressed len %d", w.Body.Len(), len(body))
	}
}

// ─── handleSPA ───────────────────────────────────────────────────────────────

func TestHandleSPA(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>spa-shell</html>"), 0o600); err != nil {
		t.Fatalf("write index.html: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("console.log('hi')"), 0o600); err != nil {
		t.Fatalf("write app.js: %v", err)
	}

	srv := &Server{cfg: &config.Config{StaticDir: dir}}

	t.Run("existing static file is served as-is", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/app.js", http.NoBody)
		w := httptest.NewRecorder()
		srv.handleSPA(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
		if w.Body.String() != "console.log('hi')" {
			t.Fatalf("body = %q, want app.js contents", w.Body.String())
		}
	})

	t.Run("unknown path falls back to index.html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/some/client/route", http.NoBody)
		w := httptest.NewRecorder()
		srv.handleSPA(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
		if w.Body.String() != "<html>spa-shell</html>" {
			t.Fatalf("body = %q, want index.html contents", w.Body.String())
		}
	})

	t.Run("prerendered route directory serves its index.html without a redirect", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Join(dir, "pricing"), 0o750); err != nil {
			t.Fatalf("mkdir pricing: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "pricing", "index.html"), []byte("<html>pricing-page</html>"), 0o600); err != nil {
			t.Fatalf("write pricing/index.html: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/pricing", http.NoBody)
		w := httptest.NewRecorder()
		srv.handleSPA(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (no 301-to-trailing-slash redirect)", w.Code, http.StatusOK)
		}
		if w.Body.String() != "<html>pricing-page</html>" {
			t.Fatalf("body = %q, want pricing/index.html contents", w.Body.String())
		}
	})

	t.Run("hashed asset under /assets/ gets a long-lived immutable cache header", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o750); err != nil {
			t.Fatalf("mkdir assets: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "assets", "index-abc123.js"), []byte("console.log('hashed')"), 0o600); err != nil {
			t.Fatalf("write hashed asset: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/assets/index-abc123.js", http.NoBody)
		w := httptest.NewRecorder()
		srv.handleSPA(w, req)

		if got, want := w.Header().Get("Cache-Control"), "public, max-age=31536000, immutable"; got != want {
			t.Fatalf("Cache-Control = %q, want %q", got, want)
		}
	})

	t.Run("index.html fallback gets a no-cache header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/some/client/route", http.NoBody)
		w := httptest.NewRecorder()
		srv.handleSPA(w, req)

		if got, want := w.Header().Get("Cache-Control"), "no-cache"; got != want {
			t.Fatalf("Cache-Control = %q, want %q", got, want)
		}
	})

	t.Run("unhashed static file gets a no-cache header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/app.js", http.NoBody)
		w := httptest.NewRecorder()
		srv.handleSPA(w, req)

		if got, want := w.Header().Get("Cache-Control"), "no-cache"; got != want {
			t.Fatalf("Cache-Control = %q, want %q", got, want)
		}
	})

	t.Run("path traversal is rejected", func(t *testing.T) {
		// http.ServeFile rejects any request whose r.URL.Path contains ".."
		// with 400 before even looking at the resolved file name — this
		// fires ahead of, and independently from, the filepath.Clean("/"+
		// path) containment in handleSPA itself.
		req := httptest.NewRequest(http.MethodGet, "/../../../../etc/passwd", http.NoBody)
		w := httptest.NewRecorder()
		srv.handleSPA(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

// ─── authOrgKey / suggestionRateLimited ───────────────────────────────

func TestSuggestionOrgKey_NoClaimsInContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/suggestions", http.NoBody)

	key, err := authOrgKey(req)

	if err == nil {
		t.Fatal("want an error when no claims are in the request context")
	}
	if key != "" {
		t.Fatalf("key = %q, want empty", key)
	}
}

func TestSuggestionOrgKey_ReturnsOrgIDFromClaims(t *testing.T) {
	var gotKey string
	var gotErr error
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey, gotErr = authOrgKey(r)
		w.WriteHeader(http.StatusOK)
	})
	// RequireAuth is the only way claims end up in the request context
	// (the storage key is unexported in the middleware package), so route
	// a real signed token through it rather than poking context directly.
	handler := apimiddleware.RequireAuth(testJWTSecret)(next)

	claims := &apimiddleware.Claims{OrgID: "org-abc"}
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(15 * time.Minute))
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/suggestions", http.NoBody)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tok})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", w.Code, http.StatusOK, w.Body.String())
	}
	if gotErr != nil {
		t.Fatalf("authOrgKey error: %v", gotErr)
	}
	if gotKey != "org-abc" {
		t.Fatalf("key = %q, want org-abc", gotKey)
	}
}

func TestSuggestionRateLimited(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/suggestions", http.NoBody)

	suggestionRateLimited(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}

	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if got["code"] != "rate_limited" {
		t.Fatalf("code = %q, want rate_limited", got["code"])
	}
}

// ─── requestLogger ───────────────────────────────────────────────────────────

func TestRequestLogger_LogsAndPassesThrough(t *testing.T) {
	logger, buf := testLoggerToBuf()
	srv := &Server{logger: logger}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})

	handler := srv.requestLogger()(next)
	req := httptest.NewRequest(http.MethodGet, "/anything", http.NoBody)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Fatal("want the wrapped handler to be called")
	}
	if w.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusTeapot)
	}

	logLine := buf.String()
	for _, want := range []string{"GET", "/anything", "418"} {
		if !bytes.Contains([]byte(logLine), []byte(want)) {
			t.Fatalf("log line %q missing %q", logLine, want)
		}
	}
}

// ─── New ─────────────────────────────────────────────────────────────────────

func TestNew_RouterIsUsable(t *testing.T) {
	srv, _ := testServer(t, nil)
	if srv.router == nil {
		t.Fatal("want a non-nil router after New()")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", http.NoBody)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
