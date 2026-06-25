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
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/billing"
	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

// planCycle is what a LemonSqueezy variant ID resolves to — a plan tier and
// its billing cycle (EP-27). "monthly" is the only cycle that existed before
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
	cfg        *config.Config
	queries    *db.Queries
	variantMap map[string]planCycle // variantID → plan + cycle
}

func NewBillingHandler(cfg *config.Config, pool *pgxpool.Pool) *BillingHandler {
	m := map[string]planCycle{}
	add := func(variantID string, plan db.Plan, cycle string) {
		if variantID != "" {
			m[variantID] = planCycle{Plan: plan, Cycle: cycle}
		}
	}
	add(cfg.LSSoloVariantID, db.PlanSolo, cycleMonthly)
	add(cfg.LSStartupVariantID, db.PlanStartup, cycleMonthly)
	add(cfg.LSEnterpriseVariantID, db.PlanEnterprise, cycleMonthly)
	add(cfg.LSSoloAnnualVariantID, db.PlanSolo, cycleAnnual)
	add(cfg.LSStartupAnnualVariantID, db.PlanStartup, cycleAnnual)
	add(cfg.LSEnterpriseAnnualVariantID, db.PlanEnterprise, cycleAnnual)
	return &BillingHandler{cfg: cfg, queries: db.New(pool), variantMap: m}
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
		Plan                        string  `json:"plan"`
		BillingCycle                string  `json:"billingCycle"`
		SubscriptionStatus          string  `json:"subscriptionStatus"`
		PlanRenewsAt                *string `json:"planRenewsAt"`
		MonitorCount                int32   `json:"monitorCount"`
		MonitorLimit                int     `json:"monitorLimit"`
		StatusPageCount             int32   `json:"statusPageCount"`
		StatusPageLimit             int     `json:"statusPageLimit"`
		NotificationChannelCount    int32   `json:"notificationChannelCount"`
		NotificationChannelLimit    int     `json:"notificationChannelLimit"`
		MinIntervalMins             int     `json:"minIntervalMins"`
		CustomerPortalURL           string  `json:"customerPortalUrl"`
	}

	var renewsAt *string
	if info.PlanRenewsAt.Valid {
		s := info.PlanRenewsAt.Time.Format("2006-01-02")
		renewsAt = &s
	}

	portalURL := ""
	if info.LsCustomerID.Valid && info.LsCustomerID.String != "" {
		portalURL = "https://app.lemonsqueezy.com/my-orders/"
	}

	respond.JSON(w, http.StatusOK, response{
		Plan:                        string(info.Plan),
		BillingCycle:                info.BillingCycle,
		SubscriptionStatus:          info.SubscriptionStatus,
		PlanRenewsAt:                renewsAt,
		MonitorCount:                info.MonitorCount,
		MonitorLimit:                limits.MonitorTotal,
		StatusPageCount:             info.StatusPageCount,
		StatusPageLimit:             limits.StatusPages,
		NotificationChannelCount:    info.NotificationChannelCount,
		NotificationChannelLimit:    limits.NotificationChannels,
		MinIntervalMins:             limits.MinIntervalMins,
		CustomerPortalURL:           portalURL,
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

	// Check configuration before resolving a variant ID, so an unconfigured
	// store reports "not configured" rather than the misleading "invalid plan"
	// for a plan name that was perfectly valid.
	if h.cfg.LSAPIKey == "" || h.cfg.LSStoreID == "" {
		respond.Error(w, http.StatusServiceUnavailable, "billing not configured", "not_configured")
		return
	}

	variantID := h.variantIDForPlan(req.Plan, req.Cycle)
	if variantID == "" {
		respond.Error(w, http.StatusServiceUnavailable, "this plan isn't available yet", "not_configured")
		return
	}

	checkoutURL, err := h.createLSCheckout(orgID, variantID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to create checkout", "internal_error")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"url": checkoutURL})
}

// POST /webhook/lemonsqueezy
func (h *BillingHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if h.cfg.LSWebhookSecret == "" {
		slog.ErrorContext(r.Context(), "lemonsqueezy webhook received but LS_WEBHOOK_SECRET is not configured")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	sig := r.Header.Get("X-Signature")
	if !verifyLSSignature(body, sig, h.cfg.LSWebhookSecret) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var payload struct {
		Meta struct {
			EventName  string `json:"event_name"`
			CustomData struct {
				OrgID string `json:"org_id"`
			} `json:"custom_data"`
		} `json:"meta"`
		Data struct {
			ID         string `json:"id"`
			Attributes struct {
				Status     string  `json:"status"`
				VariantID  int64   `json:"variant_id"`
				CustomerID int64   `json:"customer_id"`
				RenewsAt   *string `json:"renews_at"`
				EndsAt     *string `json:"ends_at"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orgID, err := uuid.Parse(payload.Meta.CustomData.OrgID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	variantIDStr := fmt.Sprintf("%d", payload.Data.Attributes.VariantID)
	pc, ok := h.variantMap[variantIDStr]
	if !ok {
		// Unknown variant — ignore
		w.WriteHeader(http.StatusOK)
		return
	}
	plan, cycle := pc.Plan, pc.Cycle

	status := payload.Data.Attributes.Status
	subscriptionID := payload.Data.ID
	customerID := fmt.Sprintf("%d", payload.Data.Attributes.CustomerID)

	// On cancellation, downgrade to hobby at period end — billing_cycle resets
	// to the column default since Hobby has no cycle of its own.
	if status == "cancelled" || status == "expired" {
		plan = db.PlanHobby
		cycle = cycleMonthly
		customerID = ""
		subscriptionID = ""
	}

	var renewsAt pgtype.Timestamptz
	tsStr := payload.Data.Attributes.RenewsAt
	if status == "cancelled" {
		tsStr = payload.Data.Attributes.EndsAt
	}
	if tsStr != nil {
		if t, err := time.Parse(time.RFC3339, *tsStr); err == nil {
			renewsAt = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	if err := h.queries.UpdateOrgPlan(r.Context(), db.UpdateOrgPlanParams{
		ID:                 orgID,
		Plan:               plan,
		BillingCycle:       cycle,
		LsCustomerID:       pgtype.Text{String: customerID, Valid: customerID != ""},
		LsSubscriptionID:   pgtype.Text{String: subscriptionID, Valid: subscriptionID != ""},
		SubscriptionStatus: status,
		PlanRenewsAt:       renewsAt,
	}); err != nil {
		slog.ErrorContext(r.Context(), "failed to update org plan from lemonsqueezy webhook", "org_id", orgID, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillingHandler) variantIDForPlan(plan, cycle string) string {
	annual := cycle == cycleAnnual
	switch plan {
	case "solo":
		if annual {
			return h.cfg.LSSoloAnnualVariantID
		}
		return h.cfg.LSSoloVariantID
	case "startup":
		if annual {
			return h.cfg.LSStartupAnnualVariantID
		}
		return h.cfg.LSStartupVariantID
	case "enterprise":
		if annual {
			return h.cfg.LSEnterpriseAnnualVariantID
		}
		return h.cfg.LSEnterpriseVariantID
	}
	return ""
}

func (h *BillingHandler) createLSCheckout(orgID uuid.UUID, variantID string) (string, error) {
	payload := map[string]any{
		"data": map[string]any{
			"type": "checkouts",
			"attributes": map[string]any{
				"checkout_data": map[string]any{
					"custom": map[string]string{"org_id": orgID.String()},
				},
				// Explicit success redirect rather than relying on the LemonSqueezy
				// store/product dashboard default (EP-07 US-0703). Failed payments
				// stay on LemonSqueezy's own hosted checkout page natively — no
				// redirect needed for that case.
				"product_options": map[string]any{
					"redirect_url": h.cfg.AppURL + "/billing?upgraded=true",
				},
			},
			"relationships": map[string]any{
				"store":   map[string]any{"data": map[string]string{"type": "stores", "id": h.cfg.LSStoreID}},
				"variant": map[string]any{"data": map[string]string{"type": "variants", "id": variantID}},
			},
		},
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "https://api.lemonsqueezy.com/v1/checkouts", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+h.cfg.LSAPIKey)
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Data struct {
			Attributes struct {
				URL string `json:"url"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Data.Attributes.URL == "" {
		return "", fmt.Errorf("empty checkout URL from LemonSqueezy")
	}
	return result.Data.Attributes.URL, nil
}

func verifyLSSignature(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
