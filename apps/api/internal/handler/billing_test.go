package handler

// Integration tests for the billing handlers. Same conventions as
// auth_test.go: real Postgres (ADR-010), package handler (not handler_test)
// so webhook payloads can be signed with the same HMAC logic the handler
// itself uses, and the shared test helpers (testPool, doJSON, decodeBody,
// findCookie, signUpTestUser, testJWTSecret) defined there are reused here.
//
// CreateCheckout's success path (the actual call to the Paddle API) is not
// covered — http.DefaultClient is not injectable in createPaddleTransaction,
// and hitting the real Paddle API from a test isn't appropriate. The
// validation/configuration-gating branches that run before that call are
// covered instead.

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	testPaddleWebhookSecret = "test-paddle-webhook-secret"
	testSoloPriceID         = "pri_solo_test"
	testUnknownPriceID      = "pri_unknown_test"
)

// testBillingHandler builds an AuthHandler and BillingHandler sharing one
// pool/config, so a user signed up via the AuthHandler is visible to the
// BillingHandler's queries. The Paddle API key is left unset (CreateCheckout's
// "not configured" tests rely on that); only the webhook secret and the Solo
// price ID are configured.
func testBillingHandler(t *testing.T) (*AuthHandler, *BillingHandler, *pgxpool.Pool) {
	t.Helper()
	pool := testPool(t)
	cfg := &config.Config{
		Env:                 "development",
		JWTSecret:           testJWTSecret,
		JWTAccessTTL:        15 * time.Minute,
		JWTRefreshTTL:       7 * 24 * time.Hour,
		AppURL:              "http://localhost:5173",
		PaddleWebhookSecret: testPaddleWebhookSecret,
		PaddleSoloPriceID:   testSoloPriceID,
	}
	return NewAuthHandler(cfg, pool), NewBillingHandler(cfg, pool), pool
}

func signWebhookBody(ts string, body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + ":"))
	mac.Write(body)
	h1 := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("ts=%s;h1=%s", ts, h1)
}

// webhookEvent builds a payload shaped like a Paddle subscription webhook,
// matching the fields billing.go's Webhook handler reads.
func webhookEvent(eventType, orgID, status, priceID, customerID, subscriptionID string, nextBilledAt, periodEndsAt *string) map[string]any {
	data := map[string]any{
		"id":             subscriptionID,
		"status":         status,
		"customer_id":    customerID,
		"custom_data":    map[string]any{"org_id": orgID},
		"items":          []map[string]any{{"price": map[string]any{"id": priceID}}},
		"next_billed_at": nextBilledAt,
	}
	if periodEndsAt != nil {
		data["current_billing_period"] = map[string]any{"ends_at": *periodEndsAt}
	}
	return map[string]any{
		"event_type": eventType,
		"data":       data,
	}
}

func doWebhook(t *testing.T, h *BillingHandler, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal webhook payload: %v", err)
	}
	ts := fmt.Sprintf("%d", time.Now().Unix())
	req := httptest.NewRequest(http.MethodPost, "/webhook/paddle", bytes.NewReader(b))
	req.Header.Set("Paddle-Signature", signWebhookBody(ts, b, testPaddleWebhookSecret))
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
		req := httptest.NewRequest(http.MethodPost, "/webhook/paddle", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()
		h.Webhook(w, req)
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("want 503, got %d", w.Code)
		}
	})

	t.Run("invalid signature returns 401", func(t *testing.T) {
		_, h, _ := testBillingHandler(t)
		body := []byte(`{"event_type":"subscription.updated"}`)
		req := httptest.NewRequest(http.MethodPost, "/webhook/paddle", bytes.NewReader(body))
		req.Header.Set("Paddle-Signature", "ts=1700000000;h1=0000000000000000000000000000000000000000000000000000000000000000")
		w := httptest.NewRecorder()
		h.Webhook(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("malformed JSON body returns 400", func(t *testing.T) {
		_, h, _ := testBillingHandler(t)
		body := []byte("not json")
		ts := fmt.Sprintf("%d", time.Now().Unix())
		req := httptest.NewRequest(http.MethodPost, "/webhook/paddle", bytes.NewReader(body))
		req.Header.Set("Paddle-Signature", signWebhookBody(ts, body, testPaddleWebhookSecret))
		w := httptest.NewRecorder()
		h.Webhook(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})

	t.Run("non-subscription event is a no-op", func(t *testing.T) {
		_, h, _ := testBillingHandler(t)
		w := doWebhook(t, h, webhookEvent("transaction.completed", "not-a-uuid", "completed", testSoloPriceID, "cust-1", "sub-1", nil, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid org_id returns 400", func(t *testing.T) {
		_, h, _ := testBillingHandler(t)
		w := doWebhook(t, h, webhookEvent("subscription.updated", "not-a-uuid", "active", testSoloPriceID, "cust-1", "sub-1", nil, nil))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})

	t.Run("unknown price is a no-op", func(t *testing.T) {
		authH, billH, pool := testBillingHandler(t)
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)

		w := doWebhook(t, billH, webhookEvent("subscription.updated", u.resp.OrgID, "active", testUnknownPriceID, "cust-1", "sub-1", nil, nil))
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

	t.Run("known price upgrades the org's plan", func(t *testing.T) {
		authH, billH, pool := testBillingHandler(t)
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)

		nextBilledAt := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
		w := doWebhook(t, billH, webhookEvent("subscription.created", u.resp.OrgID, "active", testSoloPriceID, "ctm_4242", "sub_abc", &nextBilledAt, nil))
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
		if !info.PaddleCustomerID.Valid || info.PaddleCustomerID.String != "ctm_4242" {
			t.Fatalf("want paddle_customer_id ctm_4242, got %+v", info.PaddleCustomerID)
		}
		if !info.PaddleSubscriptionID.Valid || info.PaddleSubscriptionID.String != "sub_abc" {
			t.Fatalf("want paddle_subscription_id sub_abc, got %+v", info.PaddleSubscriptionID)
		}
		if !info.PlanRenewsAt.Valid {
			t.Fatal("want plan_renews_at set")
		}
	})

	t.Run("cancellation downgrades to hobby and clears subscription fields", func(t *testing.T) {
		authH, billH, pool := testBillingHandler(t)
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)

		nextBilledAt := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
		upW := doWebhook(t, billH, webhookEvent("subscription.created", u.resp.OrgID, "active", testSoloPriceID, "ctm_4242", "sub_abc", &nextBilledAt, nil))
		if upW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", upW.Code, upW.Body.String())
		}

		periodEndsAt := time.Now().Add(5 * 24 * time.Hour).UTC().Format(time.RFC3339)
		cancelW := doWebhook(t, billH, webhookEvent("subscription.canceled", u.resp.OrgID, "canceled", testSoloPriceID, "ctm_4242", "sub_abc", nil, &periodEndsAt))
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
		if info.PaddleCustomerID.Valid {
			t.Fatal("want paddle_customer_id cleared")
		}
		if info.PaddleSubscriptionID.Valid {
			t.Fatal("want paddle_subscription_id cleared")
		}
		if info.SubscriptionStatus != "canceled" {
			t.Fatalf("want subscription status canceled, got %q", info.SubscriptionStatus)
		}
		if !info.PlanRenewsAt.Valid {
			t.Fatal("want plan_renews_at set from current_billing_period.ends_at")
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
			t.Fatalf("want no customer portal URL with no paddle_customer_id, got %q", resp.CustomerPortalURL)
		}
	})

	// A non-empty paddle_customer_id would normally trigger a live call to
	// Paddle's portal-sessions API (createPaddlePortalSession) — not covered
	// here for the same reason CreateCheckout's success path isn't (no real
	// Paddle API access from a test). GetBillingInfo degrades gracefully
	// (empty portalURL, logged error) when that call fails, which is what
	// this case exercises against a fake customer ID.
	t.Run("plan reflects an upgrade even when the portal session call fails", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)
		if err := billH.queries.UpdateOrgPlan(context.Background(), db.UpdateOrgPlanParams{
			ID:                 orgID,
			Plan:               db.PlanSolo,
			BillingCycle:       cycleMonthly,
			PaddleCustomerID:   pgtype.Text{String: "ctm_fake", Valid: true},
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

	t.Run("billing not configured (no Paddle API key)", func(t *testing.T) {
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

	t.Run("plan not available (configured API key but no price for the plan)", func(t *testing.T) {
		pool := testPool(t)
		cfg := &config.Config{
			Env:           "development",
			JWTSecret:     testJWTSecret,
			JWTAccessTTL:  15 * time.Minute,
			JWTRefreshTTL: 7 * 24 * time.Hour,
			PaddleAPIKey:  "test-key",
			// No Paddle*PriceID set for any plan.
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

func TestVerifyPaddleSignature(t *testing.T) {
	body := []byte(`{"hello":"world"}`)
	ts := "1700000000"
	header := signWebhookBody(ts, body, testPaddleWebhookSecret)
	if !verifyPaddleSignature(body, header, testPaddleWebhookSecret) {
		t.Fatal("expected matching signature to verify")
	}
	if verifyPaddleSignature(body, header, "a-different-secret") {
		t.Fatal("expected signature verification to fail with the wrong secret")
	}
	if verifyPaddleSignature([]byte(`{"hello":"mars"}`), header, testPaddleWebhookSecret) {
		t.Fatal("expected signature verification to fail for a tampered body")
	}
	if verifyPaddleSignature(body, "not-a-valid-header", testPaddleWebhookSecret) {
		t.Fatal("expected signature verification to fail for a malformed header")
	}
}
