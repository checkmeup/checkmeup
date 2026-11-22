package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

// ─── response types ──────────────────────────────────────────────────────────

type sslMonitorResponse struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Hostname             string   `json:"hostname"`
	Status               string   `json:"status"`
	AlertsEnabled        bool     `json:"alertsEnabled"`
	AlertAfterNFailures  int32    `json:"alertAfterNFailures"`
	MaxAlertsPerIncident int32    `json:"maxAlertsPerIncident"`
	ExpiresAt            *string  `json:"expiresAt"`
	Issuer               *string  `json:"issuer"`
	ErrorMsg             *string  `json:"errorMsg"`
	DaysUntilExpiry      *int     `json:"daysUntilExpiry"`
	LastCheckedAt        *string  `json:"lastCheckedAt"`
	CreatedAt            string   `json:"createdAt"`
	ProjectID            *string  `json:"projectId"`
	ChannelIDs           []string `json:"channelIds,omitempty"`
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func sslMonitorToResponse(m db.SslMonitor) sslMonitorResponse {
	r := sslMonitorResponse{
		ID:                   m.ID.String(),
		Name:                 m.Name,
		Hostname:             m.Hostname,
		Status:               string(m.Status),
		AlertsEnabled:        m.AlertsEnabled,
		AlertAfterNFailures:  m.AlertAfterNFailures,
		MaxAlertsPerIncident: m.MaxAlertsPerIncident,
		CreatedAt:            m.CreatedAt.Time.UTC().Format("2006-01-02T15:04:05Z"),
		ProjectID:            projectIDString(m.ProjectID),
	}
	if m.ExpiresAt.Valid {
		t := m.ExpiresAt.Time.UTC().Format("2006-01-02T15:04:05Z")
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
		t := m.LastCheckedAt.Time.UTC().Format("2006-01-02T15:04:05Z")
		r.LastCheckedAt = &t
	}
	return r
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
	Name       string            `json:"name"`
	Hostname   string            `json:"hostname"`
	ChannelIDs []string          `json:"channelIds"`
	ProjectID  optionalProjectID `json:"projectId,omitzero"`
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

	if !h.allowNewMonitor(w, r, orgID, "ssl_monitor") {
		return
	}

	projectID, ok := parseProjectID(w, req.ProjectID)
	if !ok {
		return
	}
	monitor, err := h.queries.CreateSSLMonitor(r.Context(), db.CreateSSLMonitorParams{
		ProjectID: projectID,
		OrgID:     orgID,
		Name:      req.Name,
		Hostname:  hostname,
	})
	if respondProjectRefused(w, err) {
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if err := h.applyMonitorNotificationChannels(r.Context(), orgID, "ssl", monitor.ID, req.ChannelIDs); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := sslMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "ssl", monitor.ID)
	respond.JSON(w, http.StatusCreated, resp)
}

// GetSSLMonitor GET /api/v1/monitors/ssl/{id}
func (h *MonitorHandler) GetSSLMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := monitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.GetSSLMonitor(r.Context(), db.GetSSLMonitorParams{
		ID: monitorID, OrgID: orgID,
	})
	if respondMonitorNotFoundOrInternal(w, err) {
		return
	}

	resp := sslMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "ssl", monitor.ID)
	respond.JSON(w, http.StatusOK, resp)
}

type updateSSLMonitorRequest struct {
	Name                 string            `json:"name"`
	Hostname             string            `json:"hostname"`
	AlertsEnabled        bool              `json:"alertsEnabled"`
	AlertAfterNFailures  int32             `json:"alertAfterNFailures"`
	MaxAlertsPerIncident int32             `json:"maxAlertsPerIncident"`
	ChannelIDs           []string          `json:"channelIds"`
	ProjectID            optionalProjectID `json:"projectId,omitzero"`
}

// UpdateSSLMonitor PATCH /api/v1/monitors/ssl/{id}
func (h *MonitorHandler) UpdateSSLMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := monitorIDs(w, r)
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

	projectID, ok := parseProjectID(w, req.ProjectID)
	if !ok {
		return
	}
	monitor, err := h.queries.UpdateSSLMonitor(r.Context(), db.UpdateSSLMonitorParams{
		SetProject:           req.ProjectID.Set,
		ProjectID:            projectID,
		ID:                   monitorID,
		OrgID:                orgID,
		Name:                 req.Name,
		Hostname:             hostname,
		AlertsEnabled:        req.AlertsEnabled,
		AlertAfterNFailures:  req.AlertAfterNFailures,
		MaxAlertsPerIncident: req.MaxAlertsPerIncident,
	})
	if respondProjectRefused(w, err) {
		return
	}
	if respondMonitorNotFoundOrInternal(w, err) {
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
	orgID, monitorID, ok := monitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.PauseSSLMonitor(r.Context(), db.PauseSSLMonitorParams{
		ID: monitorID, OrgID: orgID,
	})
	if respondMonitorNotFoundOrInternal(w, err) {
		return
	}
	respond.JSON(w, http.StatusOK, sslMonitorToResponse(monitor))
}

// ResumeSSLMonitor POST /api/v1/monitors/ssl/{id}/resume
func (h *MonitorHandler) ResumeSSLMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := monitorIDs(w, r)
	if !ok {
		return
	}
	if !h.allowResume(w, r, orgID) {
		return
	}

	monitor, err := h.queries.ResumeSSLMonitor(r.Context(), db.ResumeSSLMonitorParams{
		ID: monitorID, OrgID: orgID,
	})
	if respondMonitorNotFoundOrInternal(w, err) {
		return
	}
	respond.JSON(w, http.StatusOK, sslMonitorToResponse(monitor))
}

// DeleteSSLMonitor DELETE /api/v1/monitors/ssl/{id}
func (h *MonitorHandler) DeleteSSLMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := monitorIDs(w, r)
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
