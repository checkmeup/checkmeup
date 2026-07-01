package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/checkmeup/checkmeup/internal/billing"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

// ─── response types ──────────────────────────────────────────────────────────

type portMonitorResponse struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Host                 string   `json:"host"`
	Port                 int32    `json:"port"`
	ExpectedState        string   `json:"expectedState"`
	IntervalMins         int32    `json:"intervalMins"`
	Status               string   `json:"status"`
	AlertsEnabled        bool     `json:"alertsEnabled"`
	MaxAlertsPerIncident int32    `json:"maxAlertsPerIncident"`
	AlertAfterNFailures  int32    `json:"alertAfterNFailures"`
	LastCheckedAt        *string  `json:"lastCheckedAt"`
	CreatedAt            string   `json:"createdAt"`
	Uptime24h            *float64 `json:"uptime24h"`
	ChannelIDs           []string `json:"channelIds,omitempty"`
}

type portCheckResponse struct {
	ID             string  `json:"id"`
	CheckedAt      string  `json:"checkedAt"`
	ResponseTimeMs int32   `json:"responseTimeMs"`
	IsUp           bool    `json:"isUp"`
	FailureReason  *string `json:"failureReason"`
}

type portIncidentResponse struct {
	ID         string  `json:"id"`
	StartedAt  string  `json:"startedAt"`
	ResolvedAt *string `json:"resolvedAt"`
}

type portStatsResponse struct {
	Uptime24h *float64 `json:"uptime24h"`
	Uptime7d  *float64 `json:"uptime7d"`
	Uptime30d *float64 `json:"uptime30d"`
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (h *MonitorHandler) portMonitorToResponse(m db.PortMonitor) portMonitorResponse {
	r := portMonitorResponse{
		ID:                   m.ID.String(),
		Name:                 m.Name,
		Host:                 m.Host,
		Port:                 m.Port,
		ExpectedState:        string(m.ExpectedState),
		IntervalMins:         m.IntervalMins,
		Status:               string(m.Status),
		AlertsEnabled:        m.AlertsEnabled,
		MaxAlertsPerIncident: m.MaxAlertsPerIncident,
		AlertAfterNFailures:  m.AlertAfterNFailures,
		CreatedAt:            m.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if m.LastCheckedAt.Valid {
		t := m.LastCheckedAt.Time.Format("2006-01-02T15:04:05Z")
		r.LastCheckedAt = &t
	}
	return r
}

func portCheckToResponse(c db.PortCheck) portCheckResponse {
	r := portCheckResponse{
		ID:             c.ID.String(),
		CheckedAt:      c.CheckedAt.Time.Format("2006-01-02T15:04:05Z"),
		ResponseTimeMs: c.ResponseTimeMs,
		IsUp:           c.IsUp,
	}
	if c.FailureReason.Valid {
		r.FailureReason = &c.FailureReason.String
	}
	return r
}

func portIncidentToResponse(i db.PortIncident) portIncidentResponse {
	r := portIncidentResponse{
		ID:        i.ID.String(),
		StartedAt: i.StartedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if i.ResolvedAt.Valid {
		t := i.ResolvedAt.Time.Format("2006-01-02T15:04:05Z")
		r.ResolvedAt = &t
	}
	return r
}

func portMonitorIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
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

// validatePort reports whether port is in the valid TCP port range.
func validatePort(port int32) error {
	if port < 1 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	return nil
}

// parseExpectedState defaults to "open" for empty/invalid input — a port
// monitor is a normal uptime-style check unless the user opts into watching
// for unexpected exposure (US-3301).
func parseExpectedState(raw string) db.PortExpectedState {
	if raw == string(db.PortExpectedStateClosed) {
		return db.PortExpectedStateClosed
	}
	return db.PortExpectedStateOpen
}

// ─── handlers ────────────────────────────────────────────────────────────────

// ListPortMonitors GET /api/v1/monitors/port
func (h *MonitorHandler) ListPortMonitors(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	monitors, err := h.queries.ListPortMonitors(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	result := make([]portMonitorResponse, len(monitors))
	for i, m := range monitors {
		r2 := h.portMonitorToResponse(m)
		if stats, err := h.queries.GetPortStats(r.Context(), m.ID); err == nil {
			r2.Uptime24h = uptimePct(stats.Up24h, stats.Total24h)
		}
		result[i] = r2
	}
	respond.JSON(w, http.StatusOK, result)
}

type createPortMonitorRequest struct {
	Name                 string   `json:"name"`
	Host                 string   `json:"host"`
	Port                 int32    `json:"port"`
	ExpectedState        string   `json:"expectedState"`
	IntervalMins         int32    `json:"intervalMins"`
	MaxAlertsPerIncident int32    `json:"maxAlertsPerIncident"`
	AlertAfterNFailures  int32    `json:"alertAfterNFailures"`
	ChannelIDs           []string `json:"channelIds"`
}

// CreatePortMonitor POST /api/v1/monitors/port
func (h *MonitorHandler) CreatePortMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req createPortMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}
	host, err := parseHostname(req.Host)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	if err := validatePort(req.Port); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}

	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	total, err := h.queries.CountOrgMonitors(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if err := billing.CheckMonitorLimit(plan, int(total)); err != nil {
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "port_monitor")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return
	}
	clampedInterval, err := billing.ClampInterval(plan, int(req.IntervalMins))
	if err != nil {
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "port_monitor_interval")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return
	}
	req.IntervalMins = int32(clampedInterval)

	if req.MaxAlertsPerIncident < 0 {
		req.MaxAlertsPerIncident = 3
	}

	monitor, err := h.queries.CreatePortMonitor(r.Context(), db.CreatePortMonitorParams{
		OrgID:                orgID,
		Name:                 req.Name,
		Host:                 host,
		Port:                 req.Port,
		ExpectedState:        parseExpectedState(req.ExpectedState),
		IntervalMins:         req.IntervalMins,
		MaxAlertsPerIncident: req.MaxAlertsPerIncident,
		AlertAfterNFailures:  req.AlertAfterNFailures,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if len(req.ChannelIDs) > 0 {
		if err := h.setMonitorNotificationChannels(r.Context(), orgID, "port", monitor.ID, req.ChannelIDs); err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
			return
		}
	} else {
		h.attachDefaultNotificationChannels(r.Context(), orgID, "port", monitor.ID)
	}

	resp := h.portMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "port", monitor.ID)
	respond.JSON(w, http.StatusCreated, resp)
}

// GetPortMonitor GET /api/v1/monitors/port/{id}
func (h *MonitorHandler) GetPortMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := portMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.GetPortMonitor(r.Context(), db.GetPortMonitorParams{
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

	detail, err := h.loadPortMonitorDetail(r.Context(), monitorID, parsePageParam(r))
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := h.portMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "port", monitor.ID)

	respond.JSON(w, http.StatusOK, map[string]any{
		"monitor":   resp,
		"chartData": detail.chart,
		"checks":    detail.checks,
		"incidents": detail.incidents,
		"stats": portStatsResponse{
			Uptime24h: uptimePct(detail.stats.Up24h, detail.stats.Total24h),
			Uptime7d:  uptimePct(detail.stats.Up7d, detail.stats.Total7d),
			Uptime30d: uptimePct(detail.stats.Up30d, detail.stats.Total30d),
		},
	})
}

type portMonitorDetail struct {
	chart     []portCheckResponse
	checks    []portCheckResponse
	incidents []portIncidentResponse
	stats     db.GetPortStatsRow
}

// loadPortMonitorDetail fetches the chart/checks/incidents/stats data shown
// on the port monitor detail page and converts it to response types.
func (h *MonitorHandler) loadPortMonitorDetail(ctx context.Context, monitorID uuid.UUID, page int32) (portMonitorDetail, error) {
	chartData, err := h.queries.ListPortChecks24h(ctx, monitorID)
	if err != nil {
		return portMonitorDetail{}, err
	}

	checks, err := h.queries.ListPortChecks(ctx, db.ListPortChecksParams{
		MonitorID: monitorID,
		Limit:     50,
		Offset:    page * 50,
	})
	if err != nil {
		return portMonitorDetail{}, err
	}

	incidents, err := h.queries.ListPortIncidents(ctx, monitorID)
	if err != nil {
		return portMonitorDetail{}, err
	}

	stats, err := h.queries.GetPortStats(ctx, monitorID)
	if err != nil {
		return portMonitorDetail{}, err
	}

	return portMonitorDetail{
		chart:     portChecksToResponse(chartData),
		checks:    portChecksToResponse(checks),
		incidents: portIncidentsToResponse(incidents),
		stats:     stats,
	}, nil
}

func portChecksToResponse(checks []db.PortCheck) []portCheckResponse {
	resp := make([]portCheckResponse, len(checks))
	for i, c := range checks {
		resp[i] = portCheckToResponse(c)
	}
	return resp
}

func portIncidentsToResponse(incidents []db.PortIncident) []portIncidentResponse {
	resp := make([]portIncidentResponse, len(incidents))
	for i, inc := range incidents {
		resp[i] = portIncidentToResponse(inc)
	}
	return resp
}

type updatePortMonitorRequest struct {
	Name                 string   `json:"name"`
	Host                 string   `json:"host"`
	Port                 int32    `json:"port"`
	ExpectedState        string   `json:"expectedState"`
	IntervalMins         int32    `json:"intervalMins"`
	AlertsEnabled        bool     `json:"alertsEnabled"`
	MaxAlertsPerIncident int32    `json:"maxAlertsPerIncident"`
	AlertAfterNFailures  int32    `json:"alertAfterNFailures"`
	ChannelIDs           []string `json:"channelIds"`
}

// UpdatePortMonitor PATCH /api/v1/monitors/port/{id}
func (h *MonitorHandler) UpdatePortMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := portMonitorIDs(w, r)
	if !ok {
		return
	}

	var req updatePortMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}
	host, err := parseHostname(req.Host)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	if err := validatePort(req.Port); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}

	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	clampedInterval, err := billing.ClampInterval(plan, int(req.IntervalMins))
	if err != nil {
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "port_monitor_interval")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return
	}
	req.IntervalMins = int32(clampedInterval)

	if req.MaxAlertsPerIncident < 0 {
		req.MaxAlertsPerIncident = 3
	}

	monitor, err := h.queries.UpdatePortMonitor(r.Context(), db.UpdatePortMonitorParams{
		ID:                   monitorID,
		OrgID:                orgID,
		Name:                 req.Name,
		Host:                 host,
		Port:                 req.Port,
		ExpectedState:        parseExpectedState(req.ExpectedState),
		IntervalMins:         req.IntervalMins,
		AlertsEnabled:        req.AlertsEnabled,
		MaxAlertsPerIncident: req.MaxAlertsPerIncident,
		AlertAfterNFailures:  req.AlertAfterNFailures,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if err := h.setMonitorNotificationChannels(r.Context(), orgID, "port", monitorID, req.ChannelIDs); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := h.portMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "port", monitorID)
	respond.JSON(w, http.StatusOK, resp)
}

// PausePortMonitor POST /api/v1/monitors/port/{id}/pause
func (h *MonitorHandler) PausePortMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := portMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.PausePortMonitor(r.Context(), db.PausePortMonitorParams{
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
	respond.JSON(w, http.StatusOK, h.portMonitorToResponse(monitor))
}

// ResumePortMonitor POST /api/v1/monitors/port/{id}/resume
func (h *MonitorHandler) ResumePortMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := portMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.ResumePortMonitor(r.Context(), db.ResumePortMonitorParams{
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
	respond.JSON(w, http.StatusOK, h.portMonitorToResponse(monitor))
}

// DeletePortMonitor DELETE /api/v1/monitors/port/{id}
func (h *MonitorHandler) DeletePortMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := portMonitorIDs(w, r)
	if !ok {
		return
	}

	if err := h.queries.DeletePortMonitor(r.Context(), db.DeletePortMonitorParams{
		ID: monitorID, OrgID: orgID,
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
