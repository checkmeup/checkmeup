package handler

// Integration tests for the API key management handlers (api_keys.go).
// Same conventions as auth_test.go/monitors_test.go: real Postgres
// (ADR-010), package handler so doAuthed/withURLParam/signUpTestUser can be
// reused directly.

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
)

func testAPIKeyHandler(t *testing.T) (*AuthHandler, *APIKeyHandler, *pgxpool.Pool) {
	t.Helper()
	pool := testPool(t)
	cfg := &config.Config{
		Env:           "development",
		JWTSecret:     testJWTSecret,
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 7 * 24 * time.Hour,
		AppURL:        "http://localhost:5173",
	}
	return NewAuthHandler(cfg, pool), NewAPIKeyHandler(pool), pool
}

func TestCreateAPIKey_ReturnsRawKeyOnceAndPrefix(t *testing.T) {
	auth, h, pool := testAPIKeyHandler(t)
	user := signUpTestUser(t, auth, pool)

	w := doAuthed(t, http.MethodPost, h.CreateAPIKey, user.access, map[string]any{"label": "CI integration"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create api key: want 201, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeBody[createAPIKeyResponse](t, w)
	if !strings.HasPrefix(resp.Key, "cmu_live_") {
		t.Fatalf("key = %q, want cmu_live_ prefix", resp.Key)
	}
	if resp.KeyPrefix != resp.Key[:apiKeyPrefixLen] {
		t.Fatalf("keyPrefix = %q, want prefix of %q", resp.KeyPrefix, resp.Key)
	}
	if resp.Label != "CI integration" {
		t.Fatalf("label = %q, want %q", resp.Label, "CI integration")
	}
}

func TestListAPIKeys_ExcludesRevoked(t *testing.T) {
	auth, h, pool := testAPIKeyHandler(t)
	user := signUpTestUser(t, auth, pool)

	w1 := doAuthed(t, http.MethodPost, h.CreateAPIKey, user.access, map[string]any{"label": "keep"})
	kept := decodeBody[createAPIKeyResponse](t, w1)
	w2 := doAuthed(t, http.MethodPost, h.CreateAPIKey, user.access, map[string]any{"label": "revoke-me"})
	revoked := decodeBody[createAPIKeyResponse](t, w2)

	if w := doMaintenanceRequest(t, http.MethodDelete, h.RevokeAPIKey, user.access, revoked.ID, nil); w.Code != http.StatusNoContent {
		t.Fatalf("revoke api key: want 204, got %d: %s", w.Code, w.Body.String())
	}

	w := doAuthed(t, http.MethodGet, h.ListAPIKeys, user.access, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list api keys: want 200, got %d: %s", w.Code, w.Body.String())
	}
	keys := decodeBody[[]apiKeyResponse](t, w)
	if len(keys) != 1 || keys[0].ID != kept.ID {
		t.Fatalf("keys = %+v, want only %q", keys, kept.ID)
	}
}

func TestRevokeAPIKey_CrossOrgIsolation(t *testing.T) {
	auth, h, pool := testAPIKeyHandler(t)
	userA := signUpTestUser(t, auth, pool)
	userB := signUpTestUser(t, auth, pool)

	w := doAuthed(t, http.MethodPost, h.CreateAPIKey, userA.access, map[string]any{"label": "org-a-key"})
	created := decodeBody[createAPIKeyResponse](t, w)

	// Org B attempts to revoke org A's key — must not affect it.
	if w := doMaintenanceRequest(t, http.MethodDelete, h.RevokeAPIKey, userB.access, created.ID, nil); w.Code != http.StatusNoContent {
		t.Fatalf("revoke (other org): want 204, got %d: %s", w.Code, w.Body.String())
	}

	wList := doAuthed(t, http.MethodGet, h.ListAPIKeys, userA.access, nil)
	keys := decodeBody[[]apiKeyResponse](t, wList)
	if len(keys) != 1 || keys[0].ID != created.ID {
		t.Fatalf("org A's key was affected by org B's revoke call: %+v", keys)
	}
}

func TestCreateAPIKey_RejectsA101stActiveKey(t *testing.T) {
	auth, h, pool := testAPIKeyHandler(t)
	user := signUpTestUser(t, auth, pool)

	for i := 0; i < 100; i++ {
		mustExec(t, pool, "INSERT INTO api_keys (org_id, key_hash, key_prefix, label) VALUES ($1, gen_random_uuid()::text, 'cmu_live_seed', 'seed')", user.resp.OrgID)
	}

	w := doAuthed(t, http.MethodPost, h.CreateAPIKey, user.access, map[string]any{"label": "101st"})
	if w.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d: %s", w.Code, w.Body.String())
	}
	body := decodeBody[map[string]string](t, w)
	if body["code"] != "too_many_api_keys" {
		t.Fatalf("want code too_many_api_keys, got %q", body["code"])
	}
}
