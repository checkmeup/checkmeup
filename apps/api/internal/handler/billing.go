package handler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/billing"
	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

// planCycle is what a Paddle price ID resolves to — a plan tier and its
// billing cycle (EP-27). "monthly" is the only cycle that existed before
// annual billing; it's also the zero-value/default for Hobby, which has none.
type planCycle struct {
	Plan  db.Plan
	Cycle string
}

const (
	cycleMonthly = "monthly"
	cycleAnnual  = "annual"
)

// normalizeCycle defaults an empty cycle to monthly and reports whether the
// resulting value is one of the two valid cycles — shared by CreateCheckout
// and decodeChangePlanRequest, which otherwise carried identical copies of
// this check.
func normalizeCycle(raw string) (string, bool) {
	if raw == "" {
		raw = cycleMonthly
	}
	return raw, raw == cycleMonthly || raw == cycleAnnual
}

// validPaidPlan reports whether plan is one of the three paid tiers — shared
// by CreateCheckout (which never allows "hobby") and decodeChangePlanRequest
// (which additionally allows "hobby", checked separately by its caller).
func validPaidPlan(plan string) bool {
	return plan == "solo" || plan == "startup" || plan == "enterprise"
}

type BillingHandler struct {
	cfg      *config.Config
	queries  *db.Queries
	priceMap map[string]planCycle // priceID → plan + cycle
}

func NewBillingHandler(cfg *config.Config, pool *pgxpool.Pool) *BillingHandler {
	m := map[string]planCycle{}
	add := func(priceID string, plan db.Plan, cycle string) {
		if priceID != "" {
			m[priceID] = planCycle{Plan: plan, Cycle: cycle}
		}
	}
	add(cfg.PaddleSoloPriceID, db.PlanSolo, cycleMonthly)
	add(cfg.PaddleStartupPriceID, db.PlanStartup, cycleMonthly)
	add(cfg.PaddleEnterprisePriceID, db.PlanEnterprise, cycleMonthly)
	add(cfg.PaddleSoloAnnualPriceID, db.PlanSolo, cycleAnnual)
	add(cfg.PaddleStartupAnnualPriceID, db.PlanStartup, cycleAnnual)
	add(cfg.PaddleEnterpriseAnnualPriceID, db.PlanEnterprise, cycleAnnual)
	return &BillingHandler{cfg: cfg, queries: db.New(pool), priceMap: m}
}

// currentSmsCreditsUsed returns the used-this-month SMS credit count,
// applying the lazy monthly reset (ADR-032/US-1907, see GetBillingInfo)
// early if the reset date has already passed.
func currentSmsCreditsUsed(info db.GetOrgBillingInfoRow) int32 {
	if info.SmsCreditsResetAt.Valid && !info.SmsCreditsResetAt.Time.After(time.Now()) {
		return 0
	}
	return info.SmsCreditsUsedThisMonth
}

// formatOptionalTimestamp formats a nullable timestamp as "2006-01-02", or
// nil if unset.
func formatOptionalTimestamp(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format("2006-01-02")
	return &s
}

// resolvePortalURL fetches a fresh Paddle customer-portal URL for orgID, or
// "" if there's no Paddle customer yet or the fetch fails (logged, non-fatal
// — see GetBillingInfo).
func (h *BillingHandler) resolvePortalURL(ctx context.Context, orgID uuid.UUID, customerID pgtype.Text) string {
	if !customerID.Valid || customerID.String == "" {
		return ""
	}
	url, err := h.createPaddlePortalSession(customerID.String)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create paddle portal session", "org_id", orgID, "error", err)
		return ""
	}
	return url
}

// GET /api/v1/billing
func (h *BillingHandler) GetBillingInfo(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	info, err := h.queries.GetOrgBillingInfo(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to load billing info", "internal_error")
		return
	}

	limits := billing.GetLimits(info.Plan)

	// The monthly reset (ADR-032/US-1907) is only physically applied lazily,
	// at the next SMS send (worker.go's ConsumeSMSCredit) — so a GET here
	// right after the reset date passes, with nothing sent yet this month,
	// would otherwise still show last month's stale used count. Mirror the
	// same "reset_at <= today" check for display so the dashboard/billing
	// page always shows the current month's true count.
	smsCreditsUsed := currentSmsCreditsUsed(info)

	type response struct {
		Plan                     string  `json:"plan"`
		BillingCycle             string  `json:"billingCycle"`
		SubscriptionStatus       string  `json:"subscriptionStatus"`
		PlanRenewsAt             *string `json:"planRenewsAt"`
		MonitorCount             int32   `json:"monitorCount"`
		MonitorLimit             int     `json:"monitorLimit"`
		StatusPageCount          int32   `json:"statusPageCount"`
		StatusPageLimit          int     `json:"statusPageLimit"`
		NotificationChannelCount int32   `json:"notificationChannelCount"`
		NotificationChannelLimit int     `json:"notificationChannelLimit"`
		SmsCreditsUsed           int32   `json:"smsCreditsUsed"`
		SmsCreditsLimit          int     `json:"smsCreditsLimit"`
		MinIntervalMins          int     `json:"minIntervalMins"`
		CustomerPortalURL        string  `json:"customerPortalUrl"`
	}

	// Paddle portal links are single-use, short-lived, and must be generated
	// on demand — unlike LemonSqueezy's static my-orders URL, there's no
	// fixed link to hand back, so this costs an API call on every
	// GetBillingInfo request for orgs with a paid plan.
	portalURL := h.resolvePortalURL(r.Context(), orgID, info.PaddleCustomerID)

	respond.JSON(w, http.StatusOK, response{
		Plan:                     string(info.Plan),
		BillingCycle:             info.BillingCycle,
		SubscriptionStatus:       info.SubscriptionStatus,
		PlanRenewsAt:             formatOptionalTimestamp(info.PlanRenewsAt),
		MonitorCount:             info.MonitorCount,
		MonitorLimit:             limits.MonitorTotal,
		StatusPageCount:          info.StatusPageCount,
		StatusPageLimit:          limits.StatusPages,
		NotificationChannelCount: info.NotificationChannelCount,
		NotificationChannelLimit: limits.NotificationChannels,
		SmsCreditsUsed:           smsCreditsUsed,
		SmsCreditsLimit:          limits.SMSCredits,
		MinIntervalMins:          limits.MinIntervalMins,
		CustomerPortalURL:        portalURL,
	})
}

// POST /api/v1/billing/checkout
func (h *BillingHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req struct {
		Plan  string `json:"plan"`
		Cycle string `json:"cycle"` // "monthly" or "annual"; defaults to monthly
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request", "bad_request")
		return
	}
	cycle, cycleOK := normalizeCycle(req.Cycle)
	if !cycleOK {
		respond.Error(w, http.StatusBadRequest, "invalid cycle", "bad_request")
		return
	}
	req.Cycle = cycle
	if !validPaidPlan(req.Plan) {
		respond.Error(w, http.StatusBadRequest, "invalid plan", "bad_request")
		return
	}

	// Check configuration before resolving a price ID, so an unconfigured
	// account reports "not configured" rather than the misleading "invalid
	// plan" for a plan name that was perfectly valid.
	if h.cfg.PaddleAPIKey == "" {
		respond.Error(w, http.StatusServiceUnavailable, "billing not configured", "not_configured")
		return
	}

	priceID := h.priceIDForPlan(req.Plan, req.Cycle)
	if priceID == "" {
		respond.Error(w, http.StatusServiceUnavailable, "this plan isn't available yet", "not_configured")
		return
	}

	transactionID, err := h.createPaddleTransaction(orgID, priceID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to create checkout", "internal_error")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"transactionId": transactionID})
}

// POST /api/v1/billing/change-plan
//
// Only handles changes to an *existing* Paddle subscription (upgrade or
// downgrade between paid tiers, or cancellation down to Hobby). Moving off
// Hobby onto a first paid plan has no subscription yet to modify, so that
// still goes through CreateCheckout + the Paddle.js overlay.
func (h *BillingHandler) ChangePlan(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	plan, cycle, errMsg, errCode := decodeChangePlanRequest(r)
	if errMsg != "" {
		respond.Error(w, http.StatusBadRequest, errMsg, errCode)
		return
	}

	if h.cfg.PaddleAPIKey == "" {
		respond.Error(w, http.StatusServiceUnavailable, "billing not configured", "not_configured")
		return
	}

	info, err := h.queries.GetOrgBillingInfo(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to load billing info", "internal_error")
		return
	}
	if !info.PaddleSubscriptionID.Valid || info.PaddleSubscriptionID.String == "" {
		respond.Error(w, http.StatusBadRequest, "no active subscription to change", "no_subscription")
		return
	}

	if plan == "hobby" {
		h.changePlanToHobby(w, r, orgID, info.PaddleSubscriptionID.String)
		return
	}
	h.changePlanToPaidTier(w, r, orgID, info.PaddleSubscriptionID.String, plan, cycle)
}

// changePlanToHobby cancels orgID's Paddle subscription — ChangePlan's
// downgrade-to-hobby path.
func (h *BillingHandler) changePlanToHobby(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, subscriptionID string) {
	if err := h.cancelPaddleSubscription(subscriptionID); err != nil {
		respondForPaddlePlanChangeError(w, r, orgID, "cancel", "failed to cancel subscription", err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// changePlanToPaidTier updates orgID's Paddle subscription to a new paid
// tier/cycle — ChangePlan's upgrade/downgrade-between-paid-tiers path.
func (h *BillingHandler) changePlanToPaidTier(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, subscriptionID, plan, cycle string) {
	priceID := h.priceIDForPlan(plan, cycle)
	if priceID == "" {
		respond.Error(w, http.StatusServiceUnavailable, "this plan isn't available yet", "not_configured")
		return
	}
	if err := h.updatePaddleSubscription(subscriptionID, priceID); err != nil {
		respondForPaddlePlanChangeError(w, r, orgID, "update", "failed to change plan", err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// decodeChangePlanRequest decodes and validates a ChangePlan request body.
// A non-empty errMsg means the request is invalid and the caller should
// respond with it (and errCode) directly instead of proceeding.
func decodeChangePlanRequest(r *http.Request) (plan, cycle, errMsg, errCode string) {
	var req struct {
		Plan  string `json:"plan"`
		Cycle string `json:"cycle"` // "monthly" or "annual"; ignored for hobby
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return "", "", "invalid request", "bad_request"
	}
	if req.Plan != "hobby" && !validPaidPlan(req.Plan) {
		return "", "", "invalid plan", "bad_request"
	}
	cycle, cycleOK := normalizeCycle(req.Cycle)
	if !cycleOK {
		return "", "", "invalid cycle", "bad_request"
	}
	return req.Plan, cycle, "", ""
}

// respondForPaddlePlanChangeError maps a Paddle API failure from ChangePlan
// to an HTTP response. A 4xx from Paddle is a business-rule rejection (bad
// state, bad input, permissions, etc.) rather than an infrastructure
// failure, so it's surfaced as 409 with Paddle's own code/detail — never a
// hardcoded guess — so the actual reason is visible without a log dive.
// Anything else (network error, 5xx, unparseable response) stays a generic
// 500, since those aren't actionable from the error text alone.
func respondForPaddlePlanChangeError(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, action, genericMsg string, err error) {
	var paddleErr *paddleAPIError
	if errors.As(err, &paddleErr) && paddleErr.StatusCode >= 400 && paddleErr.StatusCode < 500 {
		code, detail := paddleErr.detail()
		slog.WarnContext(r.Context(), "paddle rejected plan change", "org_id", orgID, "action", action, "paddle_code", code, "paddle_detail", detail)
		msg := "Paddle rejected this change"
		if detail != "" {
			msg = fmt.Sprintf("Paddle rejected this change: %s", detail)
		}
		respond.Error(w, http.StatusConflict, msg, "conflict")
		return
	}
	slog.ErrorContext(r.Context(), "failed to "+action+" paddle subscription", "org_id", orgID, "error", err)
	respond.Error(w, http.StatusInternalServerError, genericMsg, "internal_error")
}

// paddleWebhookPayload is the subset of a Paddle subscription webhook body
// (subscription.created/.updated/.canceled) that Webhook reads.
type paddleWebhookPayload struct {
	EventType string `json:"event_type"`
	Data      struct {
		ID                   string  `json:"id"`
		Status               string  `json:"status"`
		CustomerID           string  `json:"customer_id"`
		NextBilledAt         *string `json:"next_billed_at"`
		CurrentBillingPeriod *struct {
			EndsAt string `json:"ends_at"`
		} `json:"current_billing_period"`
		// ScheduledChange is present on subscription.updated the moment a
		// cancellation is scheduled (e.g. via ChangePlan's cancel-to-Hobby
		// path) — the subscription stays "active" until it actually takes
		// effect at ScheduledChange.EffectiveAt, so without reading this the
		// org still looks like a normal active subscription and the UI has
		// no way to know a cancellation is already pending.
		ScheduledChange *struct {
			Action      string `json:"action"`
			EffectiveAt string `json:"effective_at"`
		} `json:"scheduled_change"`
		CustomData struct {
			OrgID string `json:"org_id"`
		} `json:"custom_data"`
		Items []struct {
			Price struct {
				ID string `json:"id"`
			} `json:"price"`
		} `json:"items"`
	} `json:"data"`
}

// resolveOrgPlanUpdate turns a webhook's subscription data + resolved
// plan/cycle into the fields UpdateOrgPlan needs, handling the two special
// cases (actual cancellation, and a cancellation merely scheduled for
// period end) that don't just pass the plan/cycle straight through.
func resolveOrgPlanUpdate(pc planCycle, data *paddleWebhookPayload) (plan db.Plan, cycle, status, subscriptionID, customerID string, renewsAt pgtype.Timestamptz) {
	plan, cycle = pc.Plan, pc.Cycle
	status = data.Data.Status
	subscriptionID = data.Data.ID
	customerID = data.Data.CustomerID

	// On cancellation, downgrade to hobby — billing_cycle resets to the
	// column default since Hobby has no cycle of its own.
	if status == "canceled" {
		plan = db.PlanHobby
		cycle = cycleMonthly
		customerID = ""
		subscriptionID = ""
	}

	isCancelScheduled := status == "active" && data.Data.ScheduledChange != nil && data.Data.ScheduledChange.Action == "cancel"
	if isCancelScheduled {
		status = "cancel_scheduled"
	}

	renewsAt = resolveRenewsAt(data.Data.Status, isCancelScheduled, data)
	return plan, cycle, status, subscriptionID, customerID, renewsAt
}

// resolveRenewsAt picks which timestamp represents "when does this plan
// change take effect" from a webhook payload — the billing-period end for an
// actual cancellation, the scheduled-cancellation effective date for one
// that's merely pending, or the next regular renewal otherwise — and parses
// it into a nullable timestamp.
func resolveRenewsAt(rawStatus string, isCancelScheduled bool, data *paddleWebhookPayload) pgtype.Timestamptz {
	var tsStr *string
	switch {
	case rawStatus == "canceled" && data.Data.CurrentBillingPeriod != nil:
		tsStr = &data.Data.CurrentBillingPeriod.EndsAt
	case isCancelScheduled:
		tsStr = &data.Data.ScheduledChange.EffectiveAt
	default:
		tsStr = data.Data.NextBilledAt
	}
	if tsStr == nil {
		return pgtype.Timestamptz{}
	}
	t, err := time.Parse(time.RFC3339, *tsStr)
	if err != nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// POST /webhook/paddle
func (h *BillingHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, ok := h.authenticatePaddleWebhook(w, r)
	if !ok {
		return
	}
	orgID, pc, payload, ok := h.decodeWebhookTarget(w, body)
	if !ok {
		return
	}
	h.applyPaddlePlanUpdate(w, r, orgID, pc, &payload)
}

// authenticatePaddleWebhook reads the request body and verifies it's a
// genuine, correctly-signed Paddle webhook. ok is false if it already wrote
// a rejection response.
func (h *BillingHandler) authenticatePaddleWebhook(w http.ResponseWriter, r *http.Request) (body []byte, ok bool) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, false
	}
	if h.cfg.PaddleWebhookSecret == "" {
		slog.ErrorContext(r.Context(), "paddle webhook received but PADDLE_WEBHOOK_SECRET is not configured")
		w.WriteHeader(http.StatusServiceUnavailable)
		return nil, false
	}
	if !verifyPaddleSignature(body, r.Header.Get("Paddle-Signature"), h.cfg.PaddleWebhookSecret) {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}
	return body, true
}

// decodeWebhookTarget decodes an authenticated webhook body and resolves it
// down to the org + plan/cycle a subscription event applies to. ok is false
// if the webhook was malformed or ignorable (wrong event type, unknown
// price) — the caller has already written the appropriate response.
func (h *BillingHandler) decodeWebhookTarget(w http.ResponseWriter, body []byte) (orgID uuid.UUID, pc planCycle, payload paddleWebhookPayload, ok bool) {
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return orgID, pc, payload, false
	}

	// Only subscription lifecycle events carry the subscription shape parsed
	// above (transaction.* events have a different data shape) — anything
	// else is a no-op, same forgiving style as the unknown-price case below.
	if payload.EventType != "subscription.created" &&
		payload.EventType != "subscription.updated" &&
		payload.EventType != "subscription.canceled" {
		w.WriteHeader(http.StatusOK)
		return orgID, pc, payload, false
	}

	orgID, err := uuid.Parse(payload.Data.CustomData.OrgID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return orgID, pc, payload, false
	}

	if len(payload.Data.Items) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return orgID, pc, payload, false
	}
	pc, found := h.priceMap[payload.Data.Items[0].Price.ID]
	if !found {
		// Unknown price — ignore
		w.WriteHeader(http.StatusOK)
		return orgID, pc, payload, false
	}
	return orgID, pc, payload, true
}

// applyPaddlePlanUpdate persists the plan/cycle/status resolved from a
// webhook and re-enforces the new plan's resource limits.
func (h *BillingHandler) applyPaddlePlanUpdate(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, pc planCycle, payload *paddleWebhookPayload) {
	plan, cycle, status, subscriptionID, customerID, renewsAt := resolveOrgPlanUpdate(pc, payload)

	if err := h.queries.UpdateOrgPlan(r.Context(), db.UpdateOrgPlanParams{
		ID:                   orgID,
		Plan:                 plan,
		BillingCycle:         cycle,
		PaddleCustomerID:     pgtype.Text{String: customerID, Valid: customerID != ""},
		PaddleSubscriptionID: pgtype.Text{String: subscriptionID, Valid: subscriptionID != ""},
		SubscriptionStatus:   status,
		PlanRenewsAt:         renewsAt,
	}); err != nil {
		slog.ErrorContext(r.Context(), "failed to update org plan from paddle webhook", "org_id", orgID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Auto-pause/disable resources beyond the new plan's limits (ADR-019) —
	// newest-created first, oldest stays active. Run unconditionally rather
	// than only on a detected downgrade: it's a no-op when already under the
	// limit, and this avoids needing to diff old vs. new plan here. Logged
	// but non-fatal — Paddle shouldn't see a 500 because this internal
	// bookkeeping pass hiccuped; the plan change itself already succeeded.
	limits := billing.GetLimits(plan)
	if err := billing.EnforceMonitorLimit(r.Context(), h.queries, orgID, limits.MonitorTotal); err != nil {
		slog.ErrorContext(r.Context(), "failed to enforce monitor limit after plan change", "org_id", orgID, "error", err)
	}
	if err := billing.EnforceNotificationChannelLimit(r.Context(), h.queries, orgID, limits.NotificationChannels); err != nil {
		slog.ErrorContext(r.Context(), "failed to enforce notification channel limit after plan change", "org_id", orgID, "error", err)
	}
	if err := billing.EnforceHideBrandingLimit(r.Context(), h.queries, orgID, limits.HideBrandingAllowed); err != nil {
		slog.ErrorContext(r.Context(), "failed to enforce hide-branding limit after plan change", "org_id", orgID, "error", err)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillingHandler) priceIDForPlan(plan, cycle string) string {
	annual := cycle == cycleAnnual
	switch plan {
	case "solo":
		if annual {
			return h.cfg.PaddleSoloAnnualPriceID
		}
		return h.cfg.PaddleSoloPriceID
	case "startup":
		if annual {
			return h.cfg.PaddleStartupAnnualPriceID
		}
		return h.cfg.PaddleStartupPriceID
	case "enterprise":
		if annual {
			return h.cfg.PaddleEnterpriseAnnualPriceID
		}
		return h.cfg.PaddleEnterprisePriceID
	}
	return ""
}

// paddleAPIBase returns Paddle's production or sandbox API host — these are
// entirely separate environments (separate API keys, price IDs, customers),
// so a sandbox key against the production host (or vice versa) just fails.
func (h *BillingHandler) paddleAPIBase() string {
	if h.cfg.PaddleEnvironment == "sandbox" {
		return "https://sandbox-api.paddle.com"
	}
	return "https://api.paddle.com"
}

// createPaddleTransaction creates a Paddle transaction server-side so
// custom_data.org_id comes from the authenticated session (orgIDFrom),
// never from client input — the frontend only ever sees the resulting
// transaction ID, which it hands to Paddle.js to open the checkout overlay.
func (h *BillingHandler) createPaddleTransaction(orgID uuid.UUID, priceID string) (string, error) {
	payload := map[string]any{
		"items": []map[string]any{
			{"price_id": priceID, "quantity": 1},
		},
		"custom_data": map[string]string{"org_id": orgID.String()},
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, h.paddleAPIBase()+"/transactions", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+h.cfg.PaddleAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Data.ID == "" {
		return "", fmt.Errorf("empty transaction ID from Paddle")
	}
	return result.Data.ID, nil
}

// paddleAPIError wraps a non-2xx Paddle response so callers can distinguish
// a client-side conflict (4xx — e.g. "subscription already has a pending
// scheduled change") from a real failure (5xx, network error), instead of
// collapsing every failure into the same generic 500.
type paddleAPIError struct {
	StatusCode int
	Body       string
}

func (e *paddleAPIError) Error() string {
	return fmt.Sprintf("paddle API error: status %d, body: %s", e.StatusCode, e.Body)
}

// detail extracts Paddle's own error code/detail from its standard error
// envelope (`{"error":{"code":"...","detail":"..."}}`), so callers can
// surface the real reason instead of guessing at one.
func (e *paddleAPIError) detail() (code, detail string) {
	var parsed struct {
		Error struct {
			Code   string `json:"code"`
			Detail string `json:"detail"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(e.Body), &parsed); err != nil {
		return "", ""
	}
	return parsed.Error.Code, parsed.Error.Detail
}

// updatePaddleSubscription changes an existing subscription to a new price —
// used for upgrades/downgrades between paid tiers. "prorated_immediately"
// charges/credits the difference right away rather than waiting for the
// next billing cycle, matching how the one-off CreateCheckout upgrade path
// (from Hobby) takes effect immediately too.
func (h *BillingHandler) updatePaddleSubscription(subscriptionID, priceID string) error {
	payload := map[string]any{
		"items": []map[string]any{
			{"price_id": priceID, "quantity": 1},
		},
		"proration_billing_mode": "prorated_immediately",
	}
	b, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/subscriptions/%s", h.paddleAPIBase(), subscriptionID)
	req, _ := http.NewRequest(http.MethodPatch, url, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+h.cfg.PaddleAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
		return &paddleAPIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}
	return nil
}

// cancelPaddleSubscription schedules cancellation for the end of the current
// billing period (not immediately) — the org keeps paid-tier access until
// then, matching the "Access until <date>" copy already shown in the billing
// UI for a cancelled subscription. The plan itself only flips to Hobby once
// Paddle's subscription.canceled webhook actually fires at period end.
func (h *BillingHandler) cancelPaddleSubscription(subscriptionID string) error {
	payload := map[string]any{"effective_from": "next_billing_period"}
	b, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/subscriptions/%s/cancel", h.paddleAPIBase(), subscriptionID)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+h.cfg.PaddleAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
		return &paddleAPIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}
	return nil
}

// createPaddlePortalSession generates a single-use, short-lived customer
// portal URL — Paddle explicitly documents these as not cacheable, unlike
// LemonSqueezy's static my-orders link, so this is called fresh every time.
func (h *BillingHandler) createPaddlePortalSession(customerID string) (string, error) {
	url := fmt.Sprintf("%s/customers/%s/portal-sessions", h.paddleAPIBase(), customerID)
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Authorization", "Bearer "+h.cfg.PaddleAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Data struct {
			URLs struct {
				General struct {
					Overview string `json:"overview"`
				} `json:"general"`
			} `json:"urls"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data.URLs.General.Overview, nil
}

// verifyPaddleSignature checks the Paddle-Signature header, formatted as
// "ts=<unix_timestamp>;h1=<hex_hmac>". The signed string is "ts:rawBody" —
// see https://developer.paddle.com/webhooks/signature-verification.
func verifyPaddleSignature(body []byte, header, secret string) bool {
	var ts, h1 string
	for _, part := range strings.Split(header, ";") {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		switch k {
		case "ts":
			ts = v
		case "h1":
			h1 = v
		}
	}
	if ts == "" || h1 == "" {
		return false
	}
	if _, err := strconv.ParseInt(ts, 10, 64); err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + ":"))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(h1))
}
