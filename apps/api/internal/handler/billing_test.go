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

	t.Run("scheduled cancellation marks status cancel_scheduled without downgrading yet", func(t *testing.T) {
		authH, billH, pool := testBillingHandler(t)
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)

		nextBilledAt := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
		upW := doWebhook(t, billH, webhookEvent("subscription.created", u.resp.OrgID, "active", testSoloPriceID, "ctm_4242", "sub_abc", &nextBilledAt, nil))
		if upW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", upW.Code, upW.Body.String())
		}

		effectiveAt := time.Now().Add(20 * 24 * time.Hour).UTC().Format(time.RFC3339)
		payload := map[string]any{
			"event_type": "subscription.updated",
			"data": map[string]any{
				"id":             "sub_abc",
				"status":         "active",
				"customer_id":    "ctm_4242",
				"custom_data":    map[string]any{"org_id": u.resp.OrgID},
				"items":          []map[string]any{{"price": map[string]any{"id": testSoloPriceID}}},
				"next_billed_at": nil,
				"scheduled_change": map[string]any{
					"action":       "cancel",
					"effective_at": effectiveAt,
				},
			},
		}
		w := doWebhook(t, billH, payload)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}

		info, err := billH.queries.GetOrgBillingInfo(context.Background(), orgID)
		if err != nil {
			t.Fatalf("lookup billing info: %v", err)
		}
		if info.Plan != db.PlanSolo {
			t.Fatalf("want plan still solo (not yet downgraded), got %s", info.Plan)
		}
		if info.SubscriptionStatus != "cancel_scheduled" {
			t.Fatalf("want subscription status cancel_scheduled, got %q", info.SubscriptionStatus)
		}
		if !info.PlanRenewsAt.Valid {
			t.Fatal("want plan_renews_at set from scheduled_change.effective_at")
		}
		if !info.PaddleSubscriptionID.Valid || info.PaddleSubscriptionID.String != "sub_abc" {
			t.Fatalf("want paddle_subscription_id still set (not cleared until actual cancellation), got %+v", info.PaddleSubscriptionID)
		}
	})

	t.Run("downgrade to hobby auto-pauses the newest monitors and disables the newest channels beyond the new limit (ADR-019)", func(t *testing.T) {
		authH, billH, pool := testBillingHandler(t)
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)

		nextBilledAt := time.Now().Add(30 * 24 * time.Hour).UTC().Format(time.RFC3339)
		upW := doWebhook(t, billH, webhookEvent("subscription.created", u.resp.OrgID, "active", testSoloPriceID, "ctm_downgrade", "sub_downgrade", &nextBilledAt, nil))
		if upW.Code != http.StatusOK {
			t.Fatalf("setup: want 200, got %d: %s", upW.Code, upW.Body.String())
		}

		// Hobby allows 10 monitors — create 11, oldest to newest, so exactly
		// one (the newest) should get paused by the downgrade below.
		var oldestMonitorID, newestMonitorID uuid.UUID
		for i := range 11 {
			m, err := billH.queries.CreateCronMonitor(context.Background(), db.CreateCronMonitorParams{
				OrgID: orgID, Name: fmt.Sprintf("mon-%d", i), Schedule: "every 1h",
				GracePeriodMins: 5, PingToken: uuid.NewString(), MaxAlertsPerIncident: 3,
			})
			if err != nil {
				t.Fatalf("create monitor %d: %v", i, err)
			}
			if i == 0 {
				oldestMonitorID = m.ID
			}
			newestMonitorID = m.ID
			// Backdate so creation order is deterministic regardless of clock resolution.
			if _, err := pool.Exec(context.Background(), "UPDATE cron_monitors SET created_at = $2 WHERE id = $1", m.ID, time.Now().Add(-time.Duration(11-i)*time.Hour)); err != nil {
				t.Fatalf("backdate monitor %d: %v", i, err)
			}
		}

		// Hobby allows 5 channels — create 6, so the newest gets disabled.
		var oldestChannelID, newestChannelID uuid.UUID
		for i := range 6 {
			c, err := billH.queries.CreateNotificationChannel(context.Background(), db.CreateNotificationChannelParams{
				OrgID: orgID, Type: db.NotificationChannelTypeEmail, Name: fmt.Sprintf("chan-%d", i),
				Config: []byte(`{"email":"a@b.com"}`),
			})
			if err != nil {
				t.Fatalf("create channel %d: %v", i, err)
			}
			if i == 0 {
				oldestChannelID = c.ID
			}
			newestChannelID = c.ID
			if _, err := pool.Exec(context.Background(), "UPDATE notification_channels SET created_at = $2 WHERE id = $1", c.ID, time.Now().Add(-time.Duration(6-i)*time.Hour)); err != nil {
				t.Fatalf("backdate channel %d: %v", i, err)
			}
		}

		// ADR-035: a status page with hide_branding set on Solo should have it
		// cleared once the downgrade below drops the org to Hobby.
		page, err := billH.queries.CreateStatusPage(context.Background(), db.CreateStatusPageParams{
			OrgID: orgID, Slug: "downgrade-branding-" + uuid.NewString(), Title: "x", Layout: "classic",
		})
		if err != nil {
			t.Fatalf("create status page: %v", err)
		}
		if _, err := billH.queries.UpdateStatusPage(context.Background(), db.UpdateStatusPageParams{
			ID: page.ID, OrgID: orgID, Title: "x", HideBranding: true, Layout: "classic",
		}); err != nil {
			t.Fatalf("set hide_branding: %v", err)
		}

		periodEndsAt := time.Now().Add(5 * 24 * time.Hour).UTC().Format(time.RFC3339)
		cancelW := doWebhook(t, billH, webhookEvent("subscription.canceled", u.resp.OrgID, "canceled", testSoloPriceID, "ctm_downgrade", "sub_downgrade", nil, &periodEndsAt))
		if cancelW.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", cancelW.Code, cancelW.Body.String())
		}

		oldestMonitor, err := billH.queries.GetCronMonitor(context.Background(), db.GetCronMonitorParams{ID: oldestMonitorID, OrgID: orgID})
		if err != nil {
			t.Fatalf("get oldest monitor: %v", err)
		}
		if oldestMonitor.Status == db.MonitorStatusPaused {
			t.Fatal("want the oldest monitor to stay active after downgrade")
		}
		newestMonitor, err := billH.queries.GetCronMonitor(context.Background(), db.GetCronMonitorParams{ID: newestMonitorID, OrgID: orgID})
		if err != nil {
			t.Fatalf("get newest monitor: %v", err)
		}
		if newestMonitor.Status != db.MonitorStatusPaused {
			t.Fatalf("want the newest monitor auto-paused after downgrading past the 10-monitor Hobby limit, got status %q", newestMonitor.Status)
		}

		oldestChannel, err := billH.queries.GetNotificationChannel(context.Background(), db.GetNotificationChannelParams{ID: oldestChannelID, OrgID: orgID})
		if err != nil {
			t.Fatalf("get oldest channel: %v", err)
		}
		if !oldestChannel.Enabled {
			t.Fatal("want the oldest channel to stay enabled after downgrade")
		}
		newestChannel, err := billH.queries.GetNotificationChannel(context.Background(), db.GetNotificationChannelParams{ID: newestChannelID, OrgID: orgID})
		if err != nil {
			t.Fatalf("get newest channel: %v", err)
		}
		if newestChannel.Enabled {
			t.Fatal("want the newest channel auto-disabled after downgrading past the 5-channel Hobby limit")
		}

		downgradedPage, err := billH.queries.GetStatusPage(context.Background(), db.GetStatusPageParams{ID: page.ID, OrgID: orgID})
		if err != nil {
			t.Fatalf("get status page: %v", err)
		}
		if downgradedPage.HideBranding {
			t.Fatal("want hide_branding cleared after downgrading to Hobby (ADR-035)")
		}
	})
}

type billingInfoResponse struct {
	Plan              string `json:"plan"`
	BillingCycle      string `json:"billingCycle"`
	MonitorCount      int32  `json:"monitorCount"`
	SmsCreditsUsed    int32  `json:"smsCreditsUsed"`
	SmsCreditsLimit   int    `json:"smsCreditsLimit"`
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
		if resp.SmsCreditsLimit != 0 {
			t.Fatalf("want sms credits limit 0 on Hobby (ADR-032), got %d", resp.SmsCreditsLimit)
		}
		if resp.SmsCreditsUsed != 0 {
			t.Fatalf("want sms credits used 0 for a fresh org, got %d", resp.SmsCreditsUsed)
		}
	})

	t.Run("a paid plan reports its sms credit limit", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)
		if err := billH.queries.UpdateOrgPlan(context.Background(), db.UpdateOrgPlanParams{
			ID: orgID, Plan: db.PlanSolo, BillingCycle: cycleMonthly, SubscriptionStatus: "active",
		}); err != nil {
			t.Fatalf("seed plan: %v", err)
		}

		w := doAuthed(t, http.MethodGet, billH.GetBillingInfo, u.access, nil)
		resp := decodeBody[billingInfoResponse](t, w)
		if resp.SmsCreditsLimit != 10 {
			t.Fatalf("want sms credits limit 10 on Solo (ADR-032), got %d", resp.SmsCreditsLimit)
		}
	})

	t.Run("reflects credits already consumed this month", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)
		if err := billH.queries.UpdateOrgPlan(context.Background(), db.UpdateOrgPlanParams{
			ID: orgID, Plan: db.PlanSolo, BillingCycle: cycleMonthly, SubscriptionStatus: "active",
		}); err != nil {
			t.Fatalf("seed plan: %v", err)
		}
		if _, err := billH.queries.ConsumeSMSCredit(context.Background(), db.ConsumeSMSCreditParams{ID: orgID, CreditCost: 1, CreditLimit: 10}); err != nil {
			t.Fatalf("consume credit: %v", err)
		}

		w := doAuthed(t, http.MethodGet, billH.GetBillingInfo, u.access, nil)
		resp := decodeBody[billingInfoResponse](t, w)
		if resp.SmsCreditsUsed != 1 {
			t.Fatalf("want sms credits used 1, got %d", resp.SmsCreditsUsed)
		}
	})

	t.Run("displays 0 used once the reset date has passed, even before any send applies it", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		orgID := uuid.MustParse(u.resp.OrgID)
		if err := billH.queries.UpdateOrgPlan(context.Background(), db.UpdateOrgPlanParams{
			ID: orgID, Plan: db.PlanSolo, BillingCycle: cycleMonthly, SubscriptionStatus: "active",
		}); err != nil {
			t.Fatalf("seed plan: %v", err)
		}
		if _, err := billH.queries.ConsumeSMSCredit(context.Background(), db.ConsumeSMSCreditParams{ID: orgID, CreditCost: 1, CreditLimit: 10}); err != nil {
			t.Fatalf("consume credit: %v", err)
		}
		if _, err := pool.Exec(context.Background(), "UPDATE orgs SET sms_credits_reset_at = CURRENT_DATE - 1 WHERE id = $1", orgID); err != nil {
			t.Fatalf("force reset date: %v", err)
		}

		w := doAuthed(t, http.MethodGet, billH.GetBillingInfo, u.access, nil)
		resp := decodeBody[billingInfoResponse](t, w)
		if resp.SmsCreditsUsed != 0 {
			t.Fatalf("want sms credits used 0 once the reset date has passed (lazy reset not yet physically applied), got %d", resp.SmsCreditsUsed)
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

func TestChangePlan(t *testing.T) {
	authH, billH, pool := testBillingHandler(t)

	t.Run("unauthenticated", func(t *testing.T) {
		w := doJSON(t, billH.ChangePlan, http.MethodPost, "/api/v1/billing/change-plan", map[string]string{"plan": "hobby"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})

	t.Run("invalid plan", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, billH.ChangePlan, u.access, map[string]string{"plan": "not-a-real-plan"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid cycle", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, billH.ChangePlan, u.access, map[string]string{"plan": "solo", "cycle": "weekly"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("billing not configured (no Paddle API key)", func(t *testing.T) {
		u := signUpTestUser(t, authH, pool)
		w := doAuthed(t, http.MethodPost, billH.ChangePlan, u.access, map[string]string{"plan": "hobby"})
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("want 503, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "not_configured" {
			t.Fatalf("want code not_configured, got %q", body["code"])
		}
	})

	// A Hobby org has no paddle_subscription_id to change — exercised against
	// a handler with a Paddle API key configured so the request gets past the
	// "not configured" gate and into the no-subscription check.
	t.Run("no active subscription", func(t *testing.T) {
		pool := testPool(t)
		cfg := &config.Config{
			Env:           "development",
			JWTSecret:     testJWTSecret,
			JWTAccessTTL:  15 * time.Minute,
			JWTRefreshTTL: 7 * 24 * time.Hour,
			PaddleAPIKey:  "test-key",
		}
		authH2 := NewAuthHandler(cfg, pool)
		billH2 := NewBillingHandler(cfg, pool)
		u := signUpTestUser(t, authH2, pool)

		w := doAuthed(t, http.MethodPost, billH2.ChangePlan, u.access, map[string]string{"plan": "hobby"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
		}
		body := decodeBody[map[string]string](t, w)
		if body["code"] != "no_subscription" {
			t.Fatalf("want code no_subscription, got %q", body["code"])
		}
	})

	// The actual Paddle subscription update/cancel API calls (updatePaddleSubscription,
	// cancelPaddleSubscription) aren't covered here for the same reason
	// CreateCheckout's success path isn't — http.DefaultClient isn't injectable,
	// and hitting the real Paddle API from a test isn't appropriate.
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
