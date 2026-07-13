package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

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

	if !validateUptimeMonitorRequest(w, &req.Name, req.URL, &req.Keyword, &req.JsonAssertions, req.MaxResponseTimeMs) {
		return
	}
	clampedInterval, ok := h.checkUptimeMonitorCreateLimits(w, r, orgID, req.IntervalMins)
	if !ok {
		return
	}
	req.IntervalMins = clampedInterval

	if req.MaxAlertsPerIncident < 0 {
		req.MaxAlertsPerIncident = 3
	}

	monitor, err := h.queries.CreateUptimeMonitor(r.Context(), buildCreateUptimeMonitorParams(orgID, req))
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if err := h.attachMonitorChannels(r.Context(), orgID, "uptime", monitor.ID, req.ChannelIDs); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := h.uptimeMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "uptime", monitor.ID)
	respond.JSON(w, http.StatusCreated, resp)
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

	resp := h.uptimeMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "uptime", monitor.ID)

	respond.JSON(w, http.StatusOK, map[string]any{
		"monitor":   resp,
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

	if !validateUptimeMonitorRequest(w, &req.Name, req.URL, &req.Keyword, &req.JsonAssertions, req.MaxResponseTimeMs) {
		return
	}

	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	// Same plan-aware clamp as CreateUptimeMonitor — a request below the
	// org's plan minimum is rejected, not silently floored to a fixed value
	// that could be either too strict (Hobby's 5-min minimum) or too loose
	// (denying a paid plan's 1-min minimum).
	clampedInterval, ok := h.clampUptimeMonitorInterval(w, r, orgID, plan, req.IntervalMins)
	if !ok {
		return
	}
	req.IntervalMins = clampedInterval

	if req.MaxAlertsPerIncident < 0 {
		req.MaxAlertsPerIncident = 3
	}

	monitor, err := h.queries.UpdateUptimeMonitor(r.Context(), buildUpdateUptimeMonitorParams(orgID, monitorID, req))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if err := h.setMonitorNotificationChannels(r.Context(), orgID, "uptime", monitorID, req.ChannelIDs); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := h.uptimeMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "uptime", monitorID)
	respond.JSON(w, http.StatusOK, resp)
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
	if err := h.checkMonitorResumeLimit(r.Context(), orgID); err != nil {
		respondResumeLimitErr(w, err)
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
