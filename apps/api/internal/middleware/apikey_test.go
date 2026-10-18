package middleware_test

// Integration tests for RequireAPIKey. Per ADR testing, these hit a real
// PostgreSQL instance (docker-compose db service locally, the CI service
// container in GitHub Actions) — no DB mocks, to avoid hiding org_id-style
// bugs. Keys are seeded directly via SQL rather than through the handler
// layer, since this package tests the middleware in isolation.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/middleware"
)

func testAPIKeyPool(t *testing.T) *pgxpool.Pool {
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

// seedAPIKey inserts an org + API key directly, returning the org ID and
// the raw key value RequireAPIKey expects in the X-API-Key header.
func seedAPIKey(t *testing.T, pool *pgxpool.Pool, revoked bool) (orgID uuid.UUID, rawKey string) {
	t.Helper()
	ctx := context.Background()

	if err := pool.QueryRow(ctx, `INSERT INTO orgs (name) VALUES ($1) RETURNING id`,
		"apikey-test-org-"+uuid.NewString()).Scan(&orgID); err != nil {
		t.Fatalf("insert org: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM orgs WHERE id = $1", orgID)
	})

	rawKey = "cmu_live_" + uuid.NewString()
	sum := sha256.Sum256([]byte(rawKey))
	hash := hex.EncodeToString(sum[:])

	revokedExpr := "NULL"
	if revoked {
		revokedExpr = "NOW()"
	}
	_, err := pool.Exec(ctx, fmt.Sprintf(
		`INSERT INTO api_keys (org_id, key_hash, key_prefix, label, revoked_at) VALUES ($1, $2, $3, '', %s)`, revokedExpr),
		orgID, hash, rawKey[:16])
	if err != nil {
		t.Fatalf("insert api key: %v", err)
	}
	return orgID, rawKey
}

func apiKeyOKHandler(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.OrgIDFromAPIKey(r.Context())
	_, _ = fmt.Fprintf(w, "orgID=%s", orgID)
}

func TestRequireAPIKey_NoHeader(t *testing.T) {
	queries := db.New(testAPIKeyPool(t))
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	middleware.RequireAPIKey(queries)(http.HandlerFunc(apiKeyOKHandler)).ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestRequireAPIKey_InvalidKey(t *testing.T) {
	queries := db.New(testAPIKeyPool(t))
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-API-Key", "cmu_live_does-not-exist")
	w := httptest.NewRecorder()
	middleware.RequireAPIKey(queries)(http.HandlerFunc(apiKeyOKHandler)).ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestRequireAPIKey_ValidKey(t *testing.T) {
	pool := testAPIKeyPool(t)
	queries := db.New(pool)
	orgID, rawKey := seedAPIKey(t, pool, false)

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-API-Key", rawKey)
	w := httptest.NewRecorder()
	middleware.RequireAPIKey(queries)(http.HandlerFunc(apiKeyOKHandler)).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	if want := "orgID=" + orgID.String(); w.Body.String() != want {
		t.Fatalf("body = %q, want %q", w.Body.String(), want)
	}
}

func TestRequireAPIKey_RevokedKey(t *testing.T) {
	pool := testAPIKeyPool(t)
	queries := db.New(pool)
	_, rawKey := seedAPIKey(t, pool, true)

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-API-Key", rawKey)
	w := httptest.NewRecorder()
	middleware.RequireAPIKey(queries)(http.HandlerFunc(apiKeyOKHandler)).ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestRequireAPIKey_RecordingLastUseNeverSlowsTheRequest(t *testing.T) {
	pool := testAPIKeyPool(t)
	queries := db.New(pool)
	orgID, rawKey := seedAPIKey(t, pool, false)
	ctx := context.Background()

	// Hold the key's row lock so the last-use UPDATE has to wait for it; the
	// key lookup is a plain read and isn't blocked by the lock.
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SELECT 1 FROM api_keys WHERE org_id = $1 FOR UPDATE`, orgID); err != nil {
		t.Fatalf("lock key row: %v", err)
	}

	done := make(chan int, 1)
	go func() {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-API-Key", rawKey)
		w := httptest.NewRecorder()
		middleware.RequireAPIKey(queries)(http.HandlerFunc(apiKeyOKHandler)).ServeHTTP(w, r)
		done <- w.Code
	}()
	select {
	case code := <-done:
		if code != http.StatusOK {
			t.Fatalf("want 200, got %d", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("request waited on recording the key's last use")
	}

	// Once the lock is gone, the last use is still recorded.
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		var recorded bool
		if err := pool.QueryRow(ctx, `SELECT last_used_at IS NOT NULL FROM api_keys WHERE org_id = $1`, orgID).Scan(&recorded); err != nil {
			t.Fatalf("read last_used_at: %v", err)
		}
		if recorded {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("last_used_at was never recorded")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
