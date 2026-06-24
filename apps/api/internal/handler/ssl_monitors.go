package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/checkmeup/checkmeup/internal/billing"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

// ─── response types ──────────────────────────────────────────────────────────

type sslMonitorResponse struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Hostname        string   `json:"hostname"`
	Status          string   `json:"status"`
	AlertsEnabled   bool     `json:"alertsEnabled"`
	ExpiresAt       *string  `json:"expiresAt"`
	Issuer          *string  `json:"issuer"`
	ErrorMsg        *string  `json:"errorMsg"`
	DaysUntilExpiry *int     `json:"daysUntilExpiry"`
	LastCheckedAt   *string  `json:"lastCheckedAt"`
	CreatedAt       string   `json:"createdAt"`
	ChannelIDs      []string `json:"channelIds,omitempty"`
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func sslMonitorToResponse(m db.SslMonitor) sslMonitorResponse {
	r := sslMonitorResponse{
		ID:            m.ID.String(),
		Name:          m.Name,
		Hostname:      m.Hostname,
		Status:        string(m.Status),
		AlertsEnabled: m.AlertsEnabled,
		CreatedAt:     m.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if m.ExpiresAt.Valid {
		t := m.ExpiresAt.Time.Format("2006-01-02T15:04:05Z")
		r.ExpiresAt = &t
		days := int(time.Until(m.ExpiresAt.Time).Hours() / 24)
		r.DaysUntilExpiry = &days
	}
	if m.Issuer.Valid {
		r.Issuer = &m.Issuer.String
	}
	if m.ErrorMsg.Valid {
		r.ErrorMsg = &m.ErrorMsg.String
	}
	if m.LastCheckedAt.Valid {
		t := m.LastCheckedAt.Time.Format("2006-01-02T15:04:05Z")
		r.LastCheckedAt = &t
	}
	return r
}

func sslMonitorIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
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

func parseHostname(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("hostname is required")
	}
	// Accept full URLs — strip scheme and path
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	host := strings.SplitN(raw, "/", 2)[0]
	host = strings.SplitN(host, ":", 2)[0] // strip port
	if host == "" {
		return "", errors.New("hostname is required")
	}
	return host, nil
}

// ─── handlers ────────────────────────────────────────────────────────────────

// ListSSLMonitors GET /api/v1/monitors/ssl
func (h *MonitorHandler) ListSSLMonitors(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	monitors, err := h.queries.ListSSLMonitors(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	result := make([]sslMonitorResponse, len(monitors))
	for i, m := range monitors {
		result[i] = sslMonitorToResponse(m)
	}
	respond.JSON(w, http.StatusOK, result)
}

type createSSLMonitorRequest struct {
	Name       string   `json:"name"`
	Hostname   string   `json:"hostname"`
	ChannelIDs []string `json:"channelIds"`
}

// CreateSSLMonitor POST /api/v1/monitors/ssl
func (h *MonitorHandler) CreateSSLMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req createSSLMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}
	hostname, err := parseHostname(req.Hostname)
	if err != nil {
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
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "ssl_monitor")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return
	}

	monitor, err := h.queries.CreateSSLMonitor(r.Context(), db.CreateSSLMonitorParams{
		OrgID:    orgID,
		Name:     req.Name,
		Hostname: hostname,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if len(req.ChannelIDs) > 0 {
		if err := h.setMonitorNotificationChannels(r.Context(), orgID, "ssl", monitor.ID, req.ChannelIDs); err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
			return
		}
	} else {
		h.attachDefaultNotificationChannels(r.Context(), orgID, "ssl", monitor.ID)
	}

	resp := sslMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "ssl", monitor.ID)
	respond.JSON(w, http.StatusCreated, resp)
}

// GetSSLMonitor GET /api/v1/monitors/ssl/{id}
func (h *MonitorHandler) GetSSLMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := sslMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.GetSSLMonitor(r.Context(), db.GetSSLMonitorParams{
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

	resp := sslMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "ssl", monitor.ID)
	respond.JSON(w, http.StatusOK, resp)
}

type updateSSLMonitorRequest struct {
	Name          string   `json:"name"`
	Hostname      string   `json:"hostname"`
	AlertsEnabled bool     `json:"alertsEnabled"`
	ChannelIDs    []string `json:"channelIds"`
}

// UpdateSSLMonitor PATCH /api/v1/monitors/ssl/{id}
func (h *MonitorHandler) UpdateSSLMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := sslMonitorIDs(w, r)
	if !ok {
		return
	}

	var req updateSSLMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}
	hostname, err := parseHostname(req.Hostname)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}

	monitor, err := h.queries.UpdateSSLMonitor(r.Context(), db.UpdateSSLMonitorParams{
		ID:            monitorID,
		OrgID:         orgID,
		Name:          req.Name,
		Hostname:      hostname,
		AlertsEnabled: req.AlertsEnabled,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if err := h.setMonitorNotificationChannels(r.Context(), orgID, "ssl", monitorID, req.ChannelIDs); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := sslMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "ssl", monitorID)
	respond.JSON(w, http.StatusOK, resp)
}

// PauseSSLMonitor POST /api/v1/monitors/ssl/{id}/pause
func (h *MonitorHandler) PauseSSLMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := sslMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.PauseSSLMonitor(r.Context(), db.PauseSSLMonitorParams{
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
	respond.JSON(w, http.StatusOK, sslMonitorToResponse(monitor))
}

// ResumeSSLMonitor POST /api/v1/monitors/ssl/{id}/resume
func (h *MonitorHandler) ResumeSSLMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := sslMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.ResumeSSLMonitor(r.Context(), db.ResumeSSLMonitorParams{
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
	respond.JSON(w, http.StatusOK, sslMonitorToResponse(monitor))
}

// DeleteSSLMonitor DELETE /api/v1/monitors/ssl/{id}
func (h *MonitorHandler) DeleteSSLMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := sslMonitorIDs(w, r)
	if !ok {
		return
	}

	if err := h.queries.DeleteSSLMonitor(r.Context(), db.DeleteSSLMonitorParams{
		ID: monitorID, OrgID: orgID,
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
