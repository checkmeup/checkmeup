package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
		MinIntervalMins          int     `json:"minIntervalMins"`
		CustomerPortalURL        string  `json:"customerPortalUrl"`
	}

	var renewsAt *string
	if info.PlanRenewsAt.Valid {
		s := info.PlanRenewsAt.Time.Format("2006-01-02")
		renewsAt = &s
	}

	portalURL := ""
	if info.PaddleCustomerID.Valid && info.PaddleCustomerID.String != "" {
		// Paddle portal links are single-use, short-lived, and must be
		// generated on demand — unlike LemonSqueezy's static my-orders URL,
		// there's no fixed link to hand back, so this costs an API call on
		// every GetBillingInfo request for orgs with a paid plan.
		url, err := h.createPaddlePortalSession(info.PaddleCustomerID.String)
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to create paddle portal session", "org_id", orgID, "error", err)
		} else {
			portalURL = url
		}
	}

	respond.JSON(w, http.StatusOK, response{
		Plan:                     string(info.Plan),
		BillingCycle:             info.BillingCycle,
		SubscriptionStatus:       info.SubscriptionStatus,
		PlanRenewsAt:             renewsAt,
		MonitorCount:             info.MonitorCount,
		MonitorLimit:             limits.MonitorTotal,
		StatusPageCount:          info.StatusPageCount,
		StatusPageLimit:          limits.StatusPages,
		NotificationChannelCount: info.NotificationChannelCount,
		NotificationChannelLimit: limits.NotificationChannels,
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
	if req.Cycle == "" {
		req.Cycle = cycleMonthly
	}
	if req.Cycle != cycleMonthly && req.Cycle != cycleAnnual {
		respond.Error(w, http.StatusBadRequest, "invalid cycle", "bad_request")
		return
	}
	if req.Plan != "solo" && req.Plan != "startup" && req.Plan != "enterprise" {
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

// POST /webhook/paddle
func (h *BillingHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if h.cfg.PaddleWebhookSecret == "" {
		slog.ErrorContext(r.Context(), "paddle webhook received but PADDLE_WEBHOOK_SECRET is not configured")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	sig := r.Header.Get("Paddle-Signature")
	if !verifyPaddleSignature(body, sig, h.cfg.PaddleWebhookSecret) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var payload struct {
		EventType string `json:"event_type"`
		Data      struct {
			ID                   string  `json:"id"`
			Status               string  `json:"status"`
			CustomerID           string  `json:"customer_id"`
			NextBilledAt         *string `json:"next_billed_at"`
			CurrentBillingPeriod *struct {
				EndsAt string `json:"ends_at"`
			} `json:"current_billing_period"`
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
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Only subscription lifecycle events carry the subscription shape parsed
	// above (transaction.* events have a different data shape) — anything
	// else is a no-op, same forgiving style as the unknown-price case below.
	if payload.EventType != "subscription.created" &&
		payload.EventType != "subscription.updated" &&
		payload.EventType != "subscription.canceled" {
		w.WriteHeader(http.StatusOK)
		return
	}

	orgID, err := uuid.Parse(payload.Data.CustomData.OrgID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(payload.Data.Items) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	priceID := payload.Data.Items[0].Price.ID
	pc, ok := h.priceMap[priceID]
	if !ok {
		// Unknown price — ignore
		w.WriteHeader(http.StatusOK)
		return
	}
	plan, cycle := pc.Plan, pc.Cycle

	status := payload.Data.Status
	subscriptionID := payload.Data.ID
	customerID := payload.Data.CustomerID

	// On cancellation, downgrade to hobby — billing_cycle resets to the
	// column default since Hobby has no cycle of its own.
	if status == "canceled" {
		plan = db.PlanHobby
		cycle = cycleMonthly
		customerID = ""
		subscriptionID = ""
	}

	var renewsAt pgtype.Timestamptz
	var tsStr *string
	if status == "canceled" && payload.Data.CurrentBillingPeriod != nil {
		tsStr = &payload.Data.CurrentBillingPeriod.EndsAt
	} else {
		tsStr = payload.Data.NextBilledAt
	}
	if tsStr != nil {
		if t, err := time.Parse(time.RFC3339, *tsStr); err == nil {
			renewsAt = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

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
