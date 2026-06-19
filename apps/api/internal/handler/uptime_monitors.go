package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/billing"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

// ─── response types ──────────────────────────────────────────────────────────

type uptimeMonitorResponse struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	URL                  string   `json:"url"`
	IntervalMins         int32    `json:"intervalMins"`
	Status               string   `json:"status"`
	AlertsEnabled        bool     `json:"alertsEnabled"`
	MaxAlertsPerIncident int32    `json:"maxAlertsPerIncident"`
	LastCheckedAt        *string  `json:"lastCheckedAt"`
	CreatedAt            string   `json:"createdAt"`
	Uptime24h            *float64 `json:"uptime24h"`
	Keyword              *string  `json:"keyword"`
	KeywordMode          string   `json:"keywordMode"`
	KeywordCaseSensitive bool     `json:"keywordCaseSensitive"`
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
		CreatedAt:            m.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		KeywordMode:          string(m.KeywordMode),
		KeywordCaseSensitive: m.KeywordCaseSensitive,
	}
	if m.LastCheckedAt.Valid {
		t := m.LastCheckedAt.Time.Format("2006-01-02T15:04:05Z")
		r.LastCheckedAt = &t
	}
	if m.Keyword.Valid {
		r.Keyword = &m.Keyword.String
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

func parseKeywordMode(raw string) db.KeywordMode {
	if raw == string(db.KeywordModeNotContains) {
		return db.KeywordModeNotContains
	}
	return db.KeywordModeContains
}

// ─── handlers ────────────────────────────────────────────────────────────────

// ListUptimeMonitors GET /api/v1/monitors/uptime
func (h *MonitorHandler) ListUptimeMonitors(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	monitors, err := h.queries.ListUptimeMonitors(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	result := make([]uptimeMonitorResponse, len(monitors))
	for i, m := range monitors {
		r2 := h.uptimeMonitorToResponse(m)
		if stats, err := h.queries.GetUptimeStats(r.Context(), m.ID); err == nil {
			r2.Uptime24h = uptimePct(stats.Up24h, stats.Total24h)
		}
		result[i] = r2
	}
	respond.JSON(w, http.StatusOK, result)
}

type createUptimeMonitorRequest struct {
	Name                 string `json:"name"`
	URL                  string `json:"url"`
	IntervalMins         int32  `json:"intervalMins"`
	MaxAlertsPerIncident int32  `json:"maxAlertsPerIncident"`
	Keyword              string `json:"keyword"`
	KeywordMode          string `json:"keywordMode"`
	KeywordCaseSensitive bool   `json:"keywordCaseSensitive"`
}

// CreateUptimeMonitor POST /api/v1/monitors/uptime
func (h *MonitorHandler) CreateUptimeMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req createUptimeMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}
	if err := validateURL(req.URL); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	req.Keyword = strings.TrimSpace(req.Keyword)
	if err := validateKeyword(req.Keyword); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if req.Keyword != "" {
		if err := billing.CheckKeywordMonitoringAllowed(plan); err != nil {
			slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "uptime_monitor_keyword")
			respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
			return
		}
	}
	total, err := h.queries.CountOrgMonitors(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if err := billing.CheckMonitorLimit(plan, int(total)); err != nil {
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "uptime_monitor")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return
	}
	clampedInterval, err := billing.ClampInterval(plan, int(req.IntervalMins))
	if err != nil {
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "uptime_monitor_interval")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return
	}
	req.IntervalMins = int32(clampedInterval)

	if req.MaxAlertsPerIncident < 0 {
		req.MaxAlertsPerIncident = 3
	}

	monitor, err := h.queries.CreateUptimeMonitor(r.Context(), db.CreateUptimeMonitorParams{
		OrgID:                orgID,
		Name:                 req.Name,
		Url:                  strings.TrimSpace(req.URL),
		IntervalMins:         req.IntervalMins,
		MaxAlertsPerIncident: req.MaxAlertsPerIncident,
		Keyword:              pgtype.Text{String: req.Keyword, Valid: req.Keyword != ""},
		KeywordMode:          parseKeywordMode(req.KeywordMode),
		KeywordCaseSensitive: req.KeywordCaseSensitive,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	respond.JSON(w, http.StatusCreated, h.uptimeMonitorToResponse(monitor))
}

// GetUptimeMonitor GET /api/v1/monitors/uptime/{id}
func (h *MonitorHandler) GetUptimeMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := uptimeMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.GetUptimeMonitor(r.Context(), db.GetUptimeMonitorParams{
		ID: monitorID, OrgID: orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	detail, err := h.loadUptimeMonitorDetail(r.Context(), monitorID, parsePageParam(r))
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	respond.JSON(w, http.StatusOK, map[string]any{
		"monitor":   h.uptimeMonitorToResponse(monitor),
		"chartData": detail.chart,
		"checks":    detail.checks,
		"incidents": detail.incidents,
		"stats": uptimeStatsResponse{
			Uptime24h: uptimePct(detail.stats.Up24h, detail.stats.Total24h),
			Uptime7d:  uptimePct(detail.stats.Up7d, detail.stats.Total7d),
			Uptime30d: uptimePct(detail.stats.Up30d, detail.stats.Total30d),
		},
	})
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

type updateUptimeMonitorRequest struct {
	Name                 string `json:"name"`
	URL                  string `json:"url"`
	IntervalMins         int32  `json:"intervalMins"`
	AlertsEnabled        bool   `json:"alertsEnabled"`
	MaxAlertsPerIncident int32  `json:"maxAlertsPerIncident"`
	Keyword              string `json:"keyword"`
	KeywordMode          string `json:"keywordMode"`
	KeywordCaseSensitive bool   `json:"keywordCaseSensitive"`
}

// UpdateUptimeMonitor PATCH /api/v1/monitors/uptime/{id}
func (h *MonitorHandler) UpdateUptimeMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := uptimeMonitorIDs(w, r)
	if !ok {
		return
	}

	var req updateUptimeMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}
	if err := validateURL(req.URL); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	req.Keyword = strings.TrimSpace(req.Keyword)
	if err := validateKeyword(req.Keyword); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	if req.Keyword != "" {
		plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
			return
		}
		if err := billing.CheckKeywordMonitoringAllowed(plan); err != nil {
			slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "uptime_monitor_keyword")
			respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
			return
		}
	}
	if req.IntervalMins < 10 {
		req.IntervalMins = 10
	}
	if req.MaxAlertsPerIncident < 0 {
		req.MaxAlertsPerIncident = 3
	}

	monitor, err := h.queries.UpdateUptimeMonitor(r.Context(), db.UpdateUptimeMonitorParams{
		ID:                   monitorID,
		OrgID:                orgID,
		Name:                 req.Name,
		Url:                  strings.TrimSpace(req.URL),
		IntervalMins:         req.IntervalMins,
		AlertsEnabled:        req.AlertsEnabled,
		MaxAlertsPerIncident: req.MaxAlertsPerIncident,
		Keyword:              pgtype.Text{String: req.Keyword, Valid: req.Keyword != ""},
		KeywordMode:          parseKeywordMode(req.KeywordMode),
		KeywordCaseSensitive: req.KeywordCaseSensitive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	respond.JSON(w, http.StatusOK, h.uptimeMonitorToResponse(monitor))
}

// PauseUptimeMonitor POST /api/v1/monitors/uptime/{id}/pause
func (h *MonitorHandler) PauseUptimeMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := uptimeMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.PauseUptimeMonitor(r.Context(), db.PauseUptimeMonitorParams{
		ID: monitorID, OrgID: orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	respond.JSON(w, http.StatusOK, h.uptimeMonitorToResponse(monitor))
}

// ResumeUptimeMonitor POST /api/v1/monitors/uptime/{id}/resume
func (h *MonitorHandler) ResumeUptimeMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := uptimeMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.ResumeUptimeMonitor(r.Context(), db.ResumeUptimeMonitorParams{
		ID: monitorID, OrgID: orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	respond.JSON(w, http.StatusOK, h.uptimeMonitorToResponse(monitor))
}

// DeleteUptimeMonitor DELETE /api/v1/monitors/uptime/{id}
func (h *MonitorHandler) DeleteUptimeMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := uptimeMonitorIDs(w, r)
	if !ok {
		return
	}

	if err := h.queries.DeleteUptimeMonitor(r.Context(), db.DeleteUptimeMonitorParams{
		ID: monitorID, OrgID: orgID,
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
