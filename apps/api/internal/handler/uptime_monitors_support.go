package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/billing"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

// JsonAssertion is one structured assertion evaluated against the response body.
type JsonAssertion struct {
	Path       string `json:"path"`
	Comparator string `json:"comparator"`
	Expected   string `json:"expected"`
}

var validComparators = map[string]bool{
	"equals": true, "not_equals": true, "contains": true,
	"greater_than": true, "less_than": true,
}

func validateJsonAssertions(assertions []JsonAssertion) error {
	for i, a := range assertions {
		if strings.TrimSpace(a.Path) == "" {
			return fmt.Errorf("assertion %d: path is required", i+1)
		}
		if !validComparators[a.Comparator] {
			return fmt.Errorf("assertion %d: invalid comparator %q", i+1, a.Comparator)
		}
	}
	return nil
}

// ─── response types ──────────────────────────────────────────────────────────

type uptimeMonitorResponse struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	URL                  string          `json:"url"`
	IntervalMins         int32           `json:"intervalMins"`
	Status               string          `json:"status"`
	AlertsEnabled        bool            `json:"alertsEnabled"`
	MaxAlertsPerIncident int32           `json:"maxAlertsPerIncident"`
	AlertAfterNFailures  int32           `json:"alertAfterNFailures"`
	LastCheckedAt        *string         `json:"lastCheckedAt"`
	CreatedAt            string          `json:"createdAt"`
	Uptime24h            *float64        `json:"uptime24h"`
	Keyword              *string         `json:"keyword"`
	KeywordMode          string          `json:"keywordMode"`
	KeywordCaseSensitive bool            `json:"keywordCaseSensitive"`
	JsonAssertions       []JsonAssertion `json:"jsonAssertions"`
	MaxResponseTimeMs    *int32          `json:"maxResponseTimeMs"`
	ChannelIDs           []string        `json:"channelIds,omitempty"`
}

type uptimeCheckResponse struct {
	ID             string  `json:"id"`
	CheckedAt      string  `json:"checkedAt"`
	StatusCode     *int32  `json:"statusCode"`
	ResponseTimeMs int32   `json:"responseTimeMs"`
	IsUp           bool    `json:"isUp"`
	FailureReason  *string `json:"failureReason"`
}

type uptimeIncidentResponse struct {
	ID         string  `json:"id"`
	StartedAt  string  `json:"startedAt"`
	ResolvedAt *string `json:"resolvedAt"`
}

type uptimeStatsResponse struct {
	Uptime24h *float64 `json:"uptime24h"`
	Uptime7d  *float64 `json:"uptime7d"`
	Uptime30d *float64 `json:"uptime30d"`
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func uptimePct(up, total int64) *float64 {
	if total == 0 {
		return nil
	}
	v := float64(up) / float64(total) * 100.0
	return &v
}

func (h *MonitorHandler) uptimeMonitorToResponse(m db.UptimeMonitor) uptimeMonitorResponse {
	r := uptimeMonitorResponse{
		ID:                   m.ID.String(),
		Name:                 m.Name,
		URL:                  m.Url,
		IntervalMins:         m.IntervalMins,
		Status:               string(m.Status),
		AlertsEnabled:        m.AlertsEnabled,
		MaxAlertsPerIncident: m.MaxAlertsPerIncident,
		AlertAfterNFailures:  m.AlertAfterNFailures,
		CreatedAt:            m.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		KeywordMode:          string(m.KeywordMode),
		KeywordCaseSensitive: m.KeywordCaseSensitive,
		JsonAssertions:       []JsonAssertion{},
	}
	if m.LastCheckedAt.Valid {
		t := m.LastCheckedAt.Time.Format("2006-01-02T15:04:05Z")
		r.LastCheckedAt = &t
	}
	if m.Keyword.Valid {
		r.Keyword = &m.Keyword.String
	}
	if len(m.JsonAssertions) > 0 {
		_ = json.Unmarshal(m.JsonAssertions, &r.JsonAssertions)
	}
	if m.MaxResponseTimeMs.Valid {
		r.MaxResponseTimeMs = &m.MaxResponseTimeMs.Int32
	}
	return r
}

func uptimeCheckToResponse(c db.UptimeCheck) uptimeCheckResponse {
	r := uptimeCheckResponse{
		ID:             c.ID.String(),
		CheckedAt:      c.CheckedAt.Time.Format("2006-01-02T15:04:05Z"),
		ResponseTimeMs: c.ResponseTimeMs,
		IsUp:           c.IsUp,
	}
	if c.StatusCode.Valid {
		r.StatusCode = &c.StatusCode.Int32
	}
	if c.FailureReason.Valid {
		r.FailureReason = &c.FailureReason.String
	}
	return r
}

func uptimeIncidentToResponse(i db.UptimeIncident) uptimeIncidentResponse {
	r := uptimeIncidentResponse{
		ID:        i.ID.String(),
		StartedAt: i.StartedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if i.ResolvedAt.Valid {
		t := i.ResolvedAt.Time.Format("2006-01-02T15:04:05Z")
		r.ResolvedAt = &t
	}
	return r
}

func uptimeMonitorIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return uuid.UUID{}, uuid.UUID{}, false
	}
	monitorID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid monitor id", "bad_request")
		return uuid.UUID{}, uuid.UUID{}, false
	}
	return orgID, monitorID, true
}

func validateURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return errors.New("url is required")
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return errors.New("url must start with http:// or https://")
	}
	return nil
}

// validateKeyword allows an empty keyword (US-1105: clearing it disables the
// check) but enforces the 1-500 char bound from US-1101 when one is set.
func validateKeyword(raw string) error {
	if raw == "" {
		return nil
	}
	if len(raw) > 500 {
		return errors.New("keyword must be 500 characters or fewer")
	}
	return nil
}

// validateUptimeMonitorRequest validates and normalizes the fields shared by
// create and update requests, responding with the first validation failure
// and returning false — same "respond internally, return ok" idiom as
// uptimeMonitorIDs above.
func validateUptimeMonitorRequest(w http.ResponseWriter, name *string, url string, keyword *string, jsonAssertions *[]JsonAssertion, maxResponseTimeMs *int32) bool {
	*name = strings.TrimSpace(*name)
	if *name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return false
	}
	if err := validateURL(url); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return false
	}
	*keyword = strings.TrimSpace(*keyword)
	if err := validateKeyword(*keyword); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return false
	}
	if *jsonAssertions == nil {
		*jsonAssertions = []JsonAssertion{}
	}
	if err := validateJsonAssertions(*jsonAssertions); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return false
	}
	if maxResponseTimeMs != nil && *maxResponseTimeMs <= 0 {
		respond.Error(w, http.StatusBadRequest, "maxResponseTimeMs must be a positive integer", "bad_request")
		return false
	}
	return true
}

// clampUptimeMonitorInterval applies the org's plan-aware interval clamp,
// responding with a plan_limit_reached error and returning false if the
// requested interval is below the plan's minimum.
func (h *MonitorHandler) clampUptimeMonitorInterval(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, plan db.Plan, requestedIntervalMins int32) (int32, bool) {
	clampedInterval, err := billing.ClampInterval(plan, int(requestedIntervalMins))
	if err != nil {
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "uptime_monitor_interval")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return 0, false
	}
	return int32(clampedInterval), true
}

// checkUptimeMonitorCreateLimits checks whether orgID can create another
// uptime monitor under its plan and returns the plan-clamped interval.
// Responds and returns false on any failure (query error or plan limit hit).
func (h *MonitorHandler) checkUptimeMonitorCreateLimits(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, requestedIntervalMins int32) (int32, bool) {
	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return 0, false
	}
	total, err := h.queries.CountOrgMonitors(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return 0, false
	}
	if err := billing.CheckMonitorLimit(plan, int(total)); err != nil {
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "uptime_monitor")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return 0, false
	}
	return h.clampUptimeMonitorInterval(w, r, orgID, plan, requestedIntervalMins)
}

func parseKeywordMode(raw string) db.KeywordMode {
	if raw == string(db.KeywordModeNotContains) {
		return db.KeywordModeNotContains
	}
	return db.KeywordModeContains
}

// parsePageParam parses the 1-based "page" query param into a 0-based page
// index, defaulting to 0 for missing/invalid/non-positive values.
func parsePageParam(r *http.Request) int32 {
	p := r.URL.Query().Get("page")
	if p == "" {
		return 0
	}
	n, err := strconv.Atoi(p)
	if err != nil || n <= 0 {
		return 0
	}
	return int32(n - 1)
}

type uptimeMonitorDetail struct {
	chart     []uptimeCheckResponse
	checks    []uptimeCheckResponse
	incidents []uptimeIncidentResponse
	stats     db.GetUptimeStatsRow
}

// loadUptimeMonitorDetail fetches the chart/checks/incidents/stats data shown
// on the uptime monitor detail page and converts it to response types.
func (h *MonitorHandler) loadUptimeMonitorDetail(ctx context.Context, monitorID uuid.UUID, page int32) (uptimeMonitorDetail, error) {
	chartData, err := h.queries.ListUptimeChecks24h(ctx, monitorID)
	if err != nil {
		return uptimeMonitorDetail{}, err
	}

	checks, err := h.queries.ListUptimeChecks(ctx, db.ListUptimeChecksParams{
		MonitorID: monitorID,
		Limit:     50,
		Offset:    page * 50,
	})
	if err != nil {
		return uptimeMonitorDetail{}, err
	}

	incidents, err := h.queries.ListUptimeIncidents(ctx, monitorID)
	if err != nil {
		return uptimeMonitorDetail{}, err
	}

	stats, err := h.queries.GetUptimeStats(ctx, monitorID)
	if err != nil {
		return uptimeMonitorDetail{}, err
	}

	return uptimeMonitorDetail{
		chart:     uptimeChecksToResponse(chartData),
		checks:    uptimeChecksToResponse(checks),
		incidents: uptimeIncidentsToResponse(incidents),
		stats:     stats,
	}, nil
}

func uptimeChecksToResponse(checks []db.UptimeCheck) []uptimeCheckResponse {
	resp := make([]uptimeCheckResponse, len(checks))
	for i, c := range checks {
		resp[i] = uptimeCheckToResponse(c)
	}
	return resp
}

func uptimeIncidentsToResponse(incidents []db.UptimeIncident) []uptimeIncidentResponse {
	resp := make([]uptimeIncidentResponse, len(incidents))
	for i, inc := range incidents {
		resp[i] = uptimeIncidentToResponse(inc)
	}
	return resp
}

type createUptimeMonitorRequest struct {
	Name                 string          `json:"name"`
	URL                  string          `json:"url"`
	IntervalMins         int32           `json:"intervalMins"`
	MaxAlertsPerIncident int32           `json:"maxAlertsPerIncident"`
	AlertAfterNFailures  int32           `json:"alertAfterNFailures"`
	Keyword              string          `json:"keyword"`
	KeywordMode          string          `json:"keywordMode"`
	KeywordCaseSensitive bool            `json:"keywordCaseSensitive"`
	JsonAssertions       []JsonAssertion `json:"jsonAssertions"`
	MaxResponseTimeMs    *int32          `json:"maxResponseTimeMs"`
	ChannelIDs           []string        `json:"channelIds"`
}

func buildCreateUptimeMonitorParams(orgID uuid.UUID, req createUptimeMonitorRequest) db.CreateUptimeMonitorParams {
	assertionsJSON, _ := json.Marshal(req.JsonAssertions)
	var maxRTParam pgtype.Int4
	if req.MaxResponseTimeMs != nil {
		maxRTParam = pgtype.Int4{Int32: *req.MaxResponseTimeMs, Valid: true}
	}
	return db.CreateUptimeMonitorParams{
		OrgID:                orgID,
		Name:                 req.Name,
		Url:                  strings.TrimSpace(req.URL),
		IntervalMins:         req.IntervalMins,
		MaxAlertsPerIncident: req.MaxAlertsPerIncident,
		AlertAfterNFailures:  req.AlertAfterNFailures,
		Keyword:              pgtype.Text{String: req.Keyword, Valid: req.Keyword != ""},
		KeywordMode:          parseKeywordMode(req.KeywordMode),
		KeywordCaseSensitive: req.KeywordCaseSensitive,
		JsonAssertions:       assertionsJSON,
		MaxResponseTimeMs:    maxRTParam,
	}
}

type updateUptimeMonitorRequest struct {
	Name                 string          `json:"name"`
	URL                  string          `json:"url"`
	IntervalMins         int32           `json:"intervalMins"`
	AlertsEnabled        bool            `json:"alertsEnabled"`
	MaxAlertsPerIncident int32           `json:"maxAlertsPerIncident"`
	AlertAfterNFailures  int32           `json:"alertAfterNFailures"`
	Keyword              string          `json:"keyword"`
	KeywordMode          string          `json:"keywordMode"`
	KeywordCaseSensitive bool            `json:"keywordCaseSensitive"`
	JsonAssertions       []JsonAssertion `json:"jsonAssertions"`
	MaxResponseTimeMs    *int32          `json:"maxResponseTimeMs"`
	ChannelIDs           []string        `json:"channelIds"`
}

func buildUpdateUptimeMonitorParams(orgID, monitorID uuid.UUID, req updateUptimeMonitorRequest) db.UpdateUptimeMonitorParams {
	assertionsJSON, _ := json.Marshal(req.JsonAssertions)
	var maxRTParam pgtype.Int4
	if req.MaxResponseTimeMs != nil {
		maxRTParam = pgtype.Int4{Int32: *req.MaxResponseTimeMs, Valid: true}
	}
	return db.UpdateUptimeMonitorParams{
		ID:                   monitorID,
		OrgID:                orgID,
		Name:                 req.Name,
		Url:                  strings.TrimSpace(req.URL),
		IntervalMins:         req.IntervalMins,
		AlertsEnabled:        req.AlertsEnabled,
		MaxAlertsPerIncident: req.MaxAlertsPerIncident,
		AlertAfterNFailures:  req.AlertAfterNFailures,
		Keyword:              pgtype.Text{String: req.Keyword, Valid: req.Keyword != ""},
		KeywordMode:          parseKeywordMode(req.KeywordMode),
		KeywordCaseSensitive: req.KeywordCaseSensitive,
		JsonAssertions:       assertionsJSON,
		MaxResponseTimeMs:    maxRTParam,
	}
}
