package handler

// Integration tests for the billing handlers. Same conventions as
// auth_test.go: real Postgres (ADR-010), package handler (not handler_test)
// so webhook payloads can be signed with the same HMAC logic the handler
// itself uses, and the shared test helpers (testPool, doJSON, decodeBody,
// findCookie, signUpTestUser, testJWTSecret) defined there are reused here.
//
// CreateCheckout's success path (the actual call to the LemonSqueezy API)
// is not covered — http.DefaultClient is not injectable in createLSCheckout,
// and hitting the real LemonSqueezy API from a test isn't appropriate. The
// validation/configuration-gating branches that run before that call are
// covered instead.

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
)

const (
	testLSWebhookSecret  = "test-ls-webhook-secret"
	testSoloVariantIDStr = "5551001"
	testSoloVariantID    = 5551001
	testUnknownVariantID = 9999999
)

// testBillingHandler builds an AuthHandler and BillingHandler sharing one
// pool/config, so a user signed up via the AuthHandler is visible to the
// BillingHandler's queries. LemonSqueezy API key/store are left unset
// (CreateCheckout's "not configured" tests rely on that); only the webhook
// secret and the Solo variant ID are configured.
func testBillingHandler(t *testing.T) (*AuthHandler, *BillingHandler, *pgxpool.Pool) {
	t.Helper()
	pool := testPool(t)
	cfg := &config.Config{
		Env:             "development",
		JWTSecret:       testJWTSecret,
		JWTAccessTTL:    15 * time.Minute,
		JWTRefreshTTL:   7 * 24 * time.Hour,
		AppURL:          "http://localhost:5173",
		LSWebhookSecret: testLSWebhookSecret,
		LSSoloVariantID: testSoloVariantIDStr,
	}
	return NewAuthHandler(cfg, pool), NewBillingHandler(cfg, pool), pool
}

func signWebhookBody(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// webhookEvent builds a payload shaped like a LemonSqueezy webhook, matching
// the fields billing.go's Webhook handler reads.
func webhookEvent(orgID, status string, variantID, customerID int64, subscriptionID string, renewsAt, endsAt *string) map[string]any {
	return map[string]any{
		"meta": map[string]any{
			"event_name":  "subscription_updated",
			"custom_data": map[string]any{"org_id": orgID},
		},
		"data": map[string]any{
			"id": subscriptionID,
			"attributes": map[string]any{
				"status":      status,
				"variant_id":  variantID,
				"customer_id": customerID,
				"renews_at":   renewsAt,
				"ends_at":     endsAt,
			},
		},
	}
}

func doWebhook(t *testing.T, h *BillingHandler, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal webhook payload: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/webhook/lemonsqueezy", bytes.NewReader(b))
	req.Header.Set("X-Signature", signWebhookBody(b, testLSWebhookSecret))
	w := httptest.NewRecorder()
	h.Webhook(w, req)
	return w
}

// doAuthed wraps handler with RequireAuth and sends the access cookie, for
// handlers that read the org from JWT claims (GetBillingInfo, CreateCheckout).
func doAuthed(t *testing.T, method string, handler http.HandlerFunc, access *http.Cookie, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader = http.NoBody
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "/", r)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(access)
	w := httptest.NewRecorder()
	apimiddleware.RequireAuth(testJWTSecret)(handler).ServeHTTP(w, req)
	return w
}

func TestWebhook(t *testing.T) {
	t.Run("missing secret config returns 503", func(t *testing.T) {
		pool := testPool(t)
		h := NewBillingHandler(&config.Config{Env: "development"}, pool)
		req := httptest.NewRequest(http.MethodPost, "/webhook/lemonsqueezy", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()
		h.Webhook(w, req)
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("want 503, got %d", w.Code)
		}
	})

	t.Run("invalid signature returns 401", func(t *testing.T) {
		_, h, _ := testBillingHandler(t)
		body := []byte(`{"meta":{"event_name":"subscription_updated"}}`)
		req := httptest.NewRequest(http.MethodPost, "/webhook/lemonsqueezy", bytes.NewReader(body))
		req.Header.Set("X-Signature", "0000000000000000000000000000000000000000000000000000000000000000")
		w := httptest.NewRecorder()
		h.Webhook(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body returns 400", func(t *testing.T) {
		_, h, _ := testBillingHandler(t)
		body := []byte("not json")
		req := httptest.NewRequest(http.MethodPost, "/webhook/lemonsqueezy", bytes.NewReader(body))
		req.Header.Set("X-Signature", signWebhookBody(body, testLSWebhookSecret))
		w := httptest.NewRecorder()
		h.Webhook(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})

	t.Run("invalid org_id returns 400", func(t *testing.T) {
		_, h, _ := testBillingHandler(t)
		w := doWebhook(t, h, webhookEvent("not-a-uuid", "active", testSoloVariantID, 1, "sub-1", nil, nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})

	t.Run("unknown variant is a no-op", func(t *testing.T) {
		authH, billH, pool := testBillingHandler(t)
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)

		w := doWebhook(t, billH, webhookEvent(u.resp.OrgID, "active", testUnknownVariantID, 1, "sub-1", nil, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		plan, err := billH.queries.GetOrgPlan(context.Background(), orgID)
		if err != nil {
			t.Fatalf("lookup plan: %v", err)
		}
		if plan != db.PlanHobby {
			t.Fatalf("want plan unchanged (hobby), got %s", plan)
		}
	})

	t.Run("known variant upgrades the org's plan", func(t *testing.T) {
		authH, billH, pool := testBillingHandler(t)
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)

		renewsAt := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
		w := doWebhook(t, billH, webhookEvent(u.resp.OrgID, "active", testSoloVariantID, 4242, "sub-abc", &renewsAt, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		info, err := billH.queries.GetOrgBillingInfo(context.Background(), orgID)
		if err != nil {
			t.Fatalf("lookup billing info: %v", err)
		}
		if info.Plan != db.PlanSolo {
			t.Fatalf("want plan solo, got %s", info.Plan)
		}
		if info.BillingCycle != cycleMonthly {
			t.Fatalf("want billing cycle monthly, got %q", info.BillingCycle)
		}
		if info.SubscriptionStatus != "active" {
			t.Fatalf("want subscription status active, got %q", info.SubscriptionStatus)
		}
		if !info.LsCustomerID.Valid || info.LsCustomerID.String != "4242" {
			t.Fatalf("want ls_customer_id 4242, got %+v", info.LsCustomerID)
		}
		if !info.LsSubscriptionID.Valid || info.LsSubscriptionID.String != "sub-abc" {
			t.Fatalf("want ls_subscription_id sub-abc, got %+v", info.LsSubscriptionID)
		}
		if !info.PlanRenewsAt.Valid {
			t.Fatal("want plan_renews_at set")
		}
	})

	t.Run("cancellation downgrades to hobby and clears subscription fields", func(t *testing.T) {
		authH, billH, pool := testBillingHandler(t)
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)

		renewsAt := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
		upW := doWebhook(t, billH, webhookEvent(u.resp.OrgID, "active", testSoloVariantID, 4242, "sub-abc", &renewsAt, nil))
		if upW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", upW.Code, upW.Body.String())
		}

		endsAt := time.Now().Add(5 * 24 * time.Hour).UTC().Format(time.RFC3339)
		cancelW := doWebhook(t, billH, webhookEvent(u.resp.OrgID, "cancelled", testSoloVariantID, 4242, "sub-abc", nil, &endsAt))
		if cancelW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", cancelW.Code, cancelW.Body.String())
		}

		info, err := billH.queries.GetOrgBillingInfo(context.Background(), orgID)
		if err != nil {
			t.Fatalf("lookup billing info: %v", err)
		}
		if info.Plan != db.PlanHobby {
			t.Fatalf("want plan reverted to hobby, got %s", info.Plan)
		}
		if info.BillingCycle != cycleMonthly {
			t.Fatalf("want billing cycle reset to monthly, got %q", info.BillingCycle)
		}
		if info.LsCustomerID.Valid {
			t.Fatal("want ls_customer_id cleared")
		}
		if info.LsSubscriptionID.Valid {
			t.Fatal("want ls_subscription_id cleared")
		}
		if info.SubscriptionStatus != "cancelled" {
			t.Fatalf("want subscription status cancelled, got %q", info.SubscriptionStatus)
		}
		if !info.PlanRenewsAt.Valid {
			t.Fatal("want plan_renews_at set from ends_at")
		}
	})
}

type billingInfoResponse struct {
	Plan              string `json:"plan"`
	BillingCycle      string `json:"billingCycle"`
	MonitorCount      int32  `json:"monitorCount"`
	CustomerPortalURL string `json:"customerPortalUrl"`
}

func TestGetBillingInfo(t *testing.T) {
	authH, billH, pool := testBillingHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/billing", nil)
		w := httptest.NewRecorder()
		billH.GetBillingInfo(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("fresh org defaults to hobby", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodGet, billH.GetBillingInfo, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[billingInfoResponse](t, w)
		if resp.Plan != "hobby" {
			t.Fatalf("want plan hobby, got %q", resp.Plan)
		}
		if resp.MonitorCount != 0 {
			t.Fatalf("want monitor count 0, got %d", resp.MonitorCount)
		}
		if resp.CustomerPortalURL != "" {
			t.Fatalf("want no customer portal URL with no ls_customer_id, got %q", resp.CustomerPortalURL)
		}
	})

	t.Run("reflects an upgraded plan and exposes a portal URL once billed", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)
		if err := billH.queries.UpdateOrgPlan(context.Background(), db.UpdateOrgPlanParams{
			ID:                 orgID,
			Plan:               db.PlanSolo,
			BillingCycle:       cycleMonthly,
			LsCustomerID:       pgtype.Text{String: "cust-1", Valid: true},
			SubscriptionStatus: "active",
		}); err != nil {
			t.Fatalf("seed plan: %v", err)
		}

		w := doAuthed(t, http.MethodGet, billH.GetBillingInfo, u.access, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := decodeBody[billingInfoResponse](t, w)
		if resp.Plan != "solo" {
			t.Fatalf("want plan solo, got %q", resp.Plan)
		}
		if resp.CustomerPortalURL == "" {
			t.Fatal("want a customer portal URL once ls_customer_id is set")
		}
	})
}

func TestCreateCheckout(t *testing.T) {
	authH, billH, pool := testBillingHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, billH.CreateCheckout, http.MethodPost, "/api/v1/billing/checkout", map[string]string{"plan": "solo"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid plan", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, billH.CreateCheckout, u.access, map[string]string{"plan": "not-a-real-plan"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid cycle", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, billH.CreateCheckout, u.access, map[string]string{"plan": "solo", "cycle": "weekly"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("billing not configured (no LemonSqueezy API key/store)", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, billH.CreateCheckout, u.access, map[string]string{"plan": "solo"})
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("want 503, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "not_configured" {
			t.Fatalf("want code not_configured, got %q", body["code"])
		}
	})

	t.Run("plan not available (configured store but no variant for the plan)", func(t *testing.T) {
		pool := testPool(t)
		cfg := &config.Config{
			Env:           "development",
			JWTSecret:     testJWTSecret,
			JWTAccessTTL:  15 * time.Minute,
			JWTRefreshTTL: 7 * 24 * time.Hour,
			LSAPIKey:      "test-key",
			LSStoreID:     "test-store",
			// No LS*VariantID set for any plan.
		}
		authH2 := NewAuthHandler(cfg, pool)
		billH2 := NewBillingHandler(cfg, pool)
		u := signUpTestUser(t, authH2, pool)

		w := doAuthed(t, http.MethodPost, billH2.CreateCheckout, u.access, map[string]string{"plan": "enterprise"})
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("want 503, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "not_configured" {
			t.Fatalf("want code not_configured, got %q", body["code"])
		}
	})
}

func TestVerifyLSSignature(t *testing.T) {
	body := []byte(`{"hello":"world"}`)
	sig := signWebhookBody(body, testLSWebhookSecret)
	if !verifyLSSignature(body, sig, testLSWebhookSecret) {
		t.Fatal("expected matching signature to verify")
	}
	if verifyLSSignature(body, sig, "a-different-secret") {
		t.Fatal("expected signature verification to fail with the wrong secret")
	}
	if verifyLSSignature([]byte(`{"hello":"mars"}`), sig, testLSWebhookSecret) {
		t.Fatal("expected signature verification to fail for a tampered body")
	}
}
