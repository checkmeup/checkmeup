package handler

// Integration tests for the public (X-API-Key authenticated) status
// endpoints in public_status.go. Same conventions as monitors_test.go: real
// Postgres (ADR-010), package handler so unexported helpers are reusable.
// Requests are routed through the real RequireAPIKey middleware (not a
// context-injection shortcut) so these tests exercise the same auth path
// production traffic does.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
)

type cronMonitorRef struct {
	ID        string `json:"id"`
	PingToken string `json:"pingToken"`
}

// createAPIKeyForOrg inserts an active API key for orgID directly (bypassing
// the handler layer, which is tested separately in api_keys_test.go) and
// returns the raw key value to send as X-API-Key.
func createAPIKeyForOrg(t *testing.T, pool *pgxpool.Pool, orgID uuid.UUID) string {
	t.Helper()
	raw := "cmu_live_" + uuid.NewString()
	sum := sha256.Sum256([]byte(raw))
	hash := hex.EncodeToString(sum[:])
	_, err := pool.Exec(context.Background(),
		`INSERT INTO api_keys (org_id, key_hash, key_prefix, label) VALUES ($1, $2, $3, '')`,
		orgID, hash, raw[:16])
	if err != nil {
		t.Fatalf("insert api key: %v", err)
	}
	return raw
}

func doPublicStatusRequest(t *testing.T, pool *pgxpool.Pool, handler http.HandlerFunc, rawKey, monitorID string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/monitors/x/"+monitorID+"/status", nil)
	req = withURLParam(req, "id", monitorID)
	req.Header.Set("X-API-Key", rawKey)
	w := httptest.NewRecorder()
	apimiddleware.RequireAPIKey(db.New(pool))(handler).ServeHTTP(w, req)
	return w
}

func TestGetCronStatus_ReturnsStatusAndPingMetadata(t *testing.T) {
	auth, h, pool := testMonitorHandler(t)
	user := signUpTestUser(t, auth, pool)

	wCreate := doAuthed(t, http.MethodPost, h.CreateCronMonitor, user.access, map[string]any{
		"name": "CI Build", "schedule": "every 1h",
	})
	if wCreate.Code != http.StatusCreated {
		t.Fatalf("create cron monitor: want 201, got %d: %s", wCreate.Code, wCreate.Body.String())
	}
	monitor := decodeBody[cronMonitorRef](t, wCreate)
	rawKey := createAPIKeyForOrg(t, pool, uuid.MustParse(user.resp.OrgID))

	// Ping with build metadata, same as a CI job would.
	pingReq := httptest.NewRequest(http.MethodGet, "/ping/"+monitor.PingToken+"?build=142&state=success", nil)
	ping := NewPingHandler(pool, nil, nil, nil, nil)
	pingW := httptest.NewRecorder()
	ping.ReceivePing(pingW, withURLParam(pingReq, "token", monitor.PingToken))
	if pingW.Code != http.StatusOK {
		t.Fatalf("ping: want 200, got %d", pingW.Code)
	}

	w := doPublicStatusRequest(t, pool, h.GetCronStatus, rawKey, monitor.ID)
	if w.Code != http.StatusOK {
		t.Fatalf("get cron status: want 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeBody[monitorStatusResponse](t, w)
	if resp.Type != "cron" || resp.Status != "up" {
		t.Fatalf("resp = %+v, want type=cron status=up", resp)
	}
	if resp.LastPingMetadata["build"] != "142" || resp.LastPingMetadata["state"] != "success" {
		t.Fatalf("lastPingMetadata = %+v, want build=142 state=success", resp.LastPingMetadata)
	}
}

func TestGetCronStatus_WrongOrgNotFound(t *testing.T) {
	auth, h, pool := testMonitorHandler(t)
	user := signUpTestUser(t, auth, pool)
	monitor := createCronMonitor(t, h, user.access, "Isolated")

	// A key belonging to a different org must not see this monitor.
	otherOrgUser := signUpTestUser(t, auth, pool)
	rawKey := createAPIKeyForOrg(t, pool, uuid.MustParse(otherOrgUser.resp.OrgID))

	w := doPublicStatusRequest(t, pool, h.GetCronStatus, rawKey, monitor.ID)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404 for a monitor belonging to a different org, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetUptimeStatus_ReturnsStatus(t *testing.T) {
	auth, h, pool := testMonitorHandler(t)
	user := signUpTestUser(t, auth, pool)
	monitor := createUptimeMonitor(t, h, user.access, "Ping Test")
	rawKey := createAPIKeyForOrg(t, pool, uuid.MustParse(user.resp.OrgID))

	w := doPublicStatusRequest(t, pool, h.GetUptimeStatus, rawKey, monitor.ID)
	if w.Code != http.StatusOK {
		t.Fatalf("get uptime status: want 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := decodeBody[monitorStatusResponse](t, w)
	if resp.Type != "uptime" {
		t.Fatalf("resp = %+v, want type=uptime", resp)
	}
}

func TestGetSSLStatus_InvalidMonitorID(t *testing.T) {
	auth, h, pool := testMonitorHandler(t)
	user := signUpTestUser(t, auth, pool)
	rawKey := createAPIKeyForOrg(t, pool, uuid.MustParse(user.resp.OrgID))

	w := doPublicStatusRequest(t, pool, h.GetSSLStatus, rawKey, "not-a-uuid")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for invalid monitor id, got %d: %s", w.Code, w.Body.String())
	}
}
