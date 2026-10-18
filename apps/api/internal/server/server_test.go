package server

// Tests for server.go. The router-wiring tests (TestRoutes_*, TestHandle*)
// hit a real PostgreSQL instance per ADR testing — DATABASE_URL must point at a
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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
)

const testJWTSecret = "test-jwt-secret-32-chars-minimum-xxx"

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://checkmeup:checkmeup@localhost:5432/checkmeup?sslmode=disable"
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

func TestRoutes_FreeToolsAreRateLimited(t *testing.T) {
	srv, _ := testServer(t, nil)
	// A non-hostname is refused before any lookup, so no check goes out; the
	// limiter counts it all the same.
	get := func(path, ip string) int {
		req := httptest.NewRequest(http.MethodGet, path+"exa%20mple", http.NoBody)
		req.RemoteAddr = ip + ":40000"
		w := httptest.NewRecorder()
		srv.router.ServeHTTP(w, req)
		return w.Code
	}

	t.Run("five checks a minute per IP", func(t *testing.T) {
		for i := 1; i <= 5; i++ {
			if code := get("/api/v1/tools/ssl?host=", "203.0.113.10"); code != http.StatusBadRequest {
				t.Fatalf("check %d: status = %d, want %d", i, code, http.StatusBadRequest)
			}
		}
		if code := get("/api/v1/tools/ssl?host=", "203.0.113.10"); code != http.StatusTooManyRequests {
			t.Fatalf("6th check: status = %d, want %d", code, http.StatusTooManyRequests)
		}
		if code := get("/api/v1/tools/ssl?host=", "203.0.113.11"); code != http.StatusBadRequest {
			t.Fatalf("another IP: status = %d, want %d", code, http.StatusBadRequest)
		}
	})

	t.Run("thirty domain lookups a minute across all visitors", func(t *testing.T) {
		for i := 1; i <= 30; i++ {
			if code := get("/api/v1/tools/domain?domain=", fmt.Sprintf("198.51.100.%d", i)); code != http.StatusBadRequest {
				t.Fatalf("lookup %d: status = %d, want %d", i, code, http.StatusBadRequest)
			}
		}
		if code := get("/api/v1/tools/domain?domain=", "198.51.100.31"); code != http.StatusTooManyRequests {
			t.Fatalf("31st lookup: status = %d, want %d", code, http.StatusTooManyRequests)
		}
	})
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

// kamal-proxy holds a deploy until GET /up answers 200; with the web app's
// routes known, the SPA fallback answers it with a 404, which failed a deploy.
func TestRoutes_KamalHealthCheckIsUpWithTheSPARoutesKnown(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"index.html":      "<html>app</html>",
		"404.html":        "<html>not-found</html>",
		"spa-routes.json": `["/"]`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	srv, _ := testServer(t, &config.Config{StaticDir: dir})

	req := httptest.NewRequest(http.MethodGet, "/up", http.NoBody)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /up status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
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

// Link checkers, SEO crawlers and uptime monitors often send HEAD; a 405
// reads to them as a broken page.
func TestRoutes_HEADOnAPageAnswersLikeGET(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>app</html>"), 0o600); err != nil {
		t.Fatalf("write index.html: %v", err)
	}
	srv, _ := testServer(t, &config.Config{StaticDir: dir})

	req := httptest.NewRequest(http.MethodHead, "/", http.NoBody)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// ─── redirectWWW ─────────────────────────────────────────────────────────────

func TestRedirectWWW(t *testing.T) {
	srv := &Server{cfg: &config.Config{AppURL: "https://checkmeup.net"}}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	t.Run("www host redirects to the canonical apex, preserving path and query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/blog/some-post?utm=1", http.NoBody)
		req.Host = "www.checkmeup.net"
		w := httptest.NewRecorder()
		srv.redirectWWW()(next).ServeHTTP(w, req)

		if w.Code != http.StatusMovedPermanently {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusMovedPermanently)
		}
		if got, want := w.Header().Get("Location"), "https://checkmeup.net/blog/some-post?utm=1"; got != want {
			t.Fatalf("Location = %q, want %q", got, want)
		}
	})

	t.Run("apex host passes through unchanged", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pricing", http.NoBody)
		req.Host = "checkmeup.net"
		w := httptest.NewRecorder()
		srv.redirectWWW()(next).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (non-www host should pass through)", w.Code, http.StatusOK)
		}
	})

	t.Run("unrelated host passes through unchanged", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pricing", http.NoBody)
		req.Host = "example.com"
		w := httptest.NewRecorder()
		srv.redirectWWW()(next).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (host other than www.<canonical> should pass through)", w.Code, http.StatusOK)
		}
	})
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

	t.Run("trailing slash 301s to the slash-less form, keeping the query", func(t *testing.T) {
		cases := map[string]string{
			"/pricing/":            "/pricing",
			"/blog/some-post/":     "/blog/some-post",
			"/blog/some-post/?a=1": "/blog/some-post?a=1",
			"/blog//":              "/blog",
			// A leading "//" must not survive into Location, where it would be
			// a protocol-relative URL to another host (open redirect).
			"//evil.example/": "/evil.example",
			// Browsers treat "/\\" like "//", so a leading backslash must not
			// survive either.
			"/\\evil.example/": "/evil.example",
		}
		for reqPath, want := range cases {
			req := httptest.NewRequest(http.MethodGet, reqPath, http.NoBody)
			w := httptest.NewRecorder()
			srv.handleSPA(w, req)

			if w.Code != http.StatusMovedPermanently {
				t.Fatalf("%s: status = %d, want %d", reqPath, w.Code, http.StatusMovedPermanently)
			}
			if got := w.Header().Get("Location"); got != want {
				t.Fatalf("%s: Location = %q, want %q", reqPath, got, want)
			}
		}
	})

	t.Run("root path is served, not redirected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		w := httptest.NewRecorder()
		srv.handleSPA(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
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

// ─── clientIPKey / pingTokenKey ─────────────────────────────────────────────

func TestClientIPKey(t *testing.T) {
	var gotKey string
	var gotErr error
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey, gotErr = clientIPKey(r)
		w.WriteHeader(http.StatusOK)
	})
	// clientIPKey reads middleware.GetClientIP from the request context, which
	// only chi's ClientIPFromRemoteAddr (installed globally in buildRouter)
	// populates — route a request through it rather than poking context directly.
	handler := middleware.ClientIPFromRemoteAddr(next)

	req := httptest.NewRequest(http.MethodGet, "/status/some-slug", http.NoBody)
	req.RemoteAddr = "203.0.113.7:54321"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if gotErr != nil {
		t.Fatalf("clientIPKey error: %v", gotErr)
	}
	if gotKey != "203.0.113.7" {
		t.Fatalf("key = %q, want %q (port stripped)", gotKey, "203.0.113.7")
	}
}

func TestClientIPFromProxy_KeysByTheVisitorBehindKamalProxy(t *testing.T) {
	cases := []struct {
		name, remoteAddr, xff, want string
	}{
		{"kamal-proxy's appended entry is the visitor", "172.18.0.2:40000", "198.51.100.9", "198.51.100.9"},
		{"a forged entry left of kamal-proxy's is ignored", "172.18.0.2:40000", "1.2.3.4, 198.51.100.9", "198.51.100.9"},
		{"a private peer with no header keeps its own address", "127.0.0.1:40000", "", "127.0.0.1"},
		{"a public peer's header is ignored", "203.0.113.7:54321", "198.51.100.9", "203.0.113.7"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got, _ = clientIPKey(r)
			})
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			req.RemoteAddr = tc.remoteAddr
			if tc.xff != "" {
				req.Header.Set("X-Forwarded-For", tc.xff)
			}
			middleware.ClientIPFromRemoteAddr(clientIPFromProxy(next)).ServeHTTP(httptest.NewRecorder(), req)
			if got != tc.want {
				t.Fatalf("key = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPingTokenKey(t *testing.T) {
	r := chi.NewRouter()
	var gotKey string
	var gotErr error
	r.Get("/ping/{token}", func(w http.ResponseWriter, r *http.Request) {
		gotKey, gotErr = pingTokenKey(r)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping/abc123", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if gotErr != nil {
		t.Fatalf("pingTokenKey error: %v", gotErr)
	}
	if gotKey != "abc123" {
		t.Fatalf("key = %q, want %q", gotKey, "abc123")
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

func TestHandleSPA_UnknownPathIsA404WhenTheRoutesAreKnown(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"index.html":      "<html>app</html>",
		"404.html":        "<html>not-found</html>",
		"spa-routes.json": `["/","/dashboard","/monitors/uptime/:id","/monitors/uptime/:id/edit"]`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	srv := &Server{cfg: &config.Config{StaticDir: dir}}

	for path, want := range map[string]struct {
		code int
		body string
	}{
		"/dashboard":                {http.StatusOK, "<html>app</html>"},
		"/monitors/uptime/abc":      {http.StatusOK, "<html>app</html>"},
		"/monitors/uptime/abc/edit": {http.StatusOK, "<html>app</html>"},
		"/nope-404":                 {http.StatusNotFound, "<html>not-found</html>"},
		"/monitors/uptime/abc/x":    {http.StatusNotFound, "<html>not-found</html>"},
		"/blog/no-such-post":        {http.StatusNotFound, "<html>not-found</html>"},
	} {
		req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
		w := httptest.NewRecorder()
		srv.handleSPA(w, req)

		if w.Code != want.code || w.Body.String() != want.body {
			t.Errorf("%s: got %d %q, want %d %q", path, w.Code, w.Body.String(), want.code, want.body)
		}
	}
}

// ─── rate limits (PRD accounts, public-api, security-hardening, status-badges) ───

// serve sends one request through the router from the given IP and returns
// the response.
func serve(srv *Server, method, path, ip string, body io.Reader, header http.Header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	req.RemoteAddr = ip + ":40000"
	for k, v := range header {
		req.Header[k] = v
	}
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)
	return w
}

// assertLimitedAfter sends allowed requests, each of which must not be
// rate limited, then one more, which must get 429 with Retry-After.
func assertLimitedAfter(t *testing.T, allowed int, send func() *httptest.ResponseRecorder) {
	t.Helper()
	for i := 1; i <= allowed; i++ {
		if w := send(); w.Code == http.StatusTooManyRequests {
			t.Fatalf("request %d: got 429 before the limit", i)
		}
	}
	w := send()
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("request %d: status = %d, want %d", allowed+1, w.Code, http.StatusTooManyRequests)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatalf("request %d: 429 has no Retry-After header", allowed+1)
	}
}

func TestRoutes_AuthEndpointsAreRateLimitedPerIP(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		allowed int
	}{
		{"a sixth sign-up within an hour gets 429", "/api/v1/auth/sign-up", 5},
		{"an eleventh sign-in within 10 minutes gets 429", "/api/v1/auth/sign-in", 10},
		{"a fourth reset request within 10 minutes gets 429", "/api/v1/auth/forgot-password", 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, _ := testServer(t, nil)
			// An empty body is refused before any account work; the
			// limiter counts it all the same.
			assertLimitedAfter(t, tc.allowed, func() *httptest.ResponseRecorder {
				return serve(srv, http.MethodPost, tc.path, "203.0.113.20", strings.NewReader("{}"), nil)
			})
			if w := serve(srv, http.MethodPost, tc.path, "203.0.113.21", strings.NewReader("{}"), nil); w.Code == http.StatusTooManyRequests {
				t.Fatal("another IP: got 429, want its own limit")
			}
		})
	}
}

// seedServerAPIKey inserts an org and an active API key directly, returning
// the raw key the public API expects in X-API-Key.
func seedServerAPIKey(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()
	var orgID uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO orgs (name) VALUES ($1) RETURNING id`,
		"server-test-org-"+uuid.NewString()).Scan(&orgID); err != nil {
		t.Fatalf("insert org: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM orgs WHERE id = $1", orgID)
	})
	raw := "cmu_live_" + uuid.NewString()
	sum := sha256.Sum256([]byte(raw))
	if _, err := pool.Exec(ctx,
		`INSERT INTO api_keys (org_id, key_hash, key_prefix, label) VALUES ($1, $2, $3, '')`,
		orgID, hex.EncodeToString(sum[:]), raw[:16]); err != nil {
		t.Fatalf("insert api key: %v", err)
	}
	return raw
}

func TestRoutes_PublicAPIIsRateLimitedPerKey(t *testing.T) {
	srv, pool := testServer(t, nil)
	path := "/api/v1/public/monitors/cron/" + uuid.NewString() + "/status"
	keyA, keyB := seedServerAPIKey(t, pool), seedServerAPIKey(t, pool)

	t.Run("a 61st request in a minute on one key gets 429 with Retry-After", func(t *testing.T) {
		// Requests rotate IPs: the limit follows the key, not the address.
		i := 0
		assertLimitedAfter(t, 60, func() *httptest.ResponseRecorder {
			i++
			return serve(srv, http.MethodGet, path, fmt.Sprintf("198.51.100.%d", i), http.NoBody, http.Header{"X-Api-Key": {keyA}})
		})
	})
	t.Run("another key has its own limit", func(t *testing.T) {
		if w := serve(srv, http.MethodGet, path, "198.51.100.1", http.NoBody, http.Header{"X-Api-Key": {keyB}}); w.Code == http.StatusTooManyRequests {
			t.Fatal("got 429 on a fresh key")
		}
	})
}

func TestRoutes_PingsAreRateLimitedPerToken(t *testing.T) {
	srv, _ := testServer(t, nil)
	token := uuid.NewString()

	t.Run("a 61st ping in a minute on one token gets 429", func(t *testing.T) {
		i := 0
		assertLimitedAfter(t, 60, func() *httptest.ResponseRecorder {
			i++
			return serve(srv, http.MethodGet, "/ping/"+token, fmt.Sprintf("198.51.100.%d", i), http.NoBody, nil)
		})
	})
	t.Run("another token has its own limit", func(t *testing.T) {
		if w := serve(srv, http.MethodGet, "/ping/"+uuid.NewString(), "198.51.100.1", http.NoBody, nil); w.Code == http.StatusTooManyRequests {
			t.Fatal("got 429 on a fresh token")
		}
	})
}

func TestRoutes_TelegramWebhookIsRateLimitedPerIP(t *testing.T) {
	// A wrong secret is answered 200 without touching Telegram.
	srv, _ := testServer(t, &config.Config{TelegramWebhookSecret: "the-real-secret"})

	t.Run("more than 60 calls a minute from one IP get 429", func(t *testing.T) {
		assertLimitedAfter(t, 60, func() *httptest.ResponseRecorder {
			return serve(srv, http.MethodPost, "/webhook/telegram", "203.0.113.30", strings.NewReader("{}"), nil)
		})
	})
	t.Run("another IP has its own limit", func(t *testing.T) {
		if w := serve(srv, http.MethodPost, "/webhook/telegram", "203.0.113.31", strings.NewReader("{}"), nil); w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestRoutes_BadgesAreRateLimitedPerIP(t *testing.T) {
	srv, _ := testServer(t, nil)
	slug := "no-such-page-" + uuid.NewString()

	t.Run("a 301st badge request in a minute from one IP gets 429", func(t *testing.T) {
		assertLimitedAfter(t, 300, func() *httptest.ResponseRecorder {
			return serve(srv, http.MethodGet, "/status/"+slug+"/badge.svg", "203.0.113.40", http.NoBody, nil)
		})
	})
	t.Run("another IP has its own limit", func(t *testing.T) {
		if w := serve(srv, http.MethodGet, "/status/"+slug+"/badge.svg", "203.0.113.41", http.NoBody, nil); w.Code == http.StatusTooManyRequests {
			t.Fatal("got 429 on a fresh IP")
		}
	})
}
