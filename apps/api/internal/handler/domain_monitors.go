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

type domainMonitorResponse struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Domain               string   `json:"domain"`
	Status               string   `json:"status"`
	AlertsEnabled        bool     `json:"alertsEnabled"`
	AlertAfterNFailures  int32    `json:"alertAfterNFailures"`
	MaxAlertsPerIncident int32    `json:"maxAlertsPerIncident"`
	ExpiresAt            *string  `json:"expiresAt"`
	Registrar            *string  `json:"registrar"`
	ErrorMsg             *string  `json:"errorMsg"`
	DaysUntilExpiry      *int     `json:"daysUntilExpiry"`
	LastCheckedAt        *string  `json:"lastCheckedAt"`
	CreatedAt            string   `json:"createdAt"`
	ChannelIDs           []string `json:"channelIds,omitempty"`
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func domainMonitorToResponse(m db.DomainMonitor) domainMonitorResponse {
	r := domainMonitorResponse{
		ID:                   m.ID.String(),
		Name:                 m.Name,
		Domain:               m.Domain,
		Status:               string(m.Status),
		AlertsEnabled:        m.AlertsEnabled,
		AlertAfterNFailures:  m.AlertAfterNFailures,
		MaxAlertsPerIncident: m.MaxAlertsPerIncident,
		CreatedAt:            m.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if m.ExpiresAt.Valid {
		t := m.ExpiresAt.Time.Format("2006-01-02T15:04:05Z")
		r.ExpiresAt = &t
		days := int(time.Until(m.ExpiresAt.Time).Hours() / 24)
		r.DaysUntilExpiry = &days
	}
	if m.Registrar.Valid {
		r.Registrar = &m.Registrar.String
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

func domainMonitorIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
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

// ─── handlers ────────────────────────────────────────────────────────────────

// ListDomainMonitors GET /api/v1/monitors/domain
func (h *MonitorHandler) ListDomainMonitors(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	monitors, err := h.queries.ListDomainMonitors(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	result := make([]domainMonitorResponse, len(monitors))
	for i, m := range monitors {
		result[i] = domainMonitorToResponse(m)
	}
	respond.JSON(w, http.StatusOK, result)
}

type createDomainMonitorRequest struct {
	Name       string   `json:"name"`
	Domain     string   `json:"domain"`
	ChannelIDs []string `json:"channelIds"`
}

// CreateDomainMonitor POST /api/v1/monitors/domain
func (h *MonitorHandler) CreateDomainMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req createDomainMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}
	// Apex domain has the same shape requirements as an SSL monitor's
	// hostname (no scheme, no path, no port) — reusing parseHostname avoids
	// duplicating that validation.
	domain, err := parseHostname(req.Domain)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}

	if !h.checkDomainMonitorLimit(w, r, orgID) {
		return
	}

	monitor, err := h.queries.CreateDomainMonitor(r.Context(), db.CreateDomainMonitorParams{
		OrgID:  orgID,
		Name:   req.Name,
		Domain: domain,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if len(req.ChannelIDs) > 0 {
		if err := h.setMonitorNotificationChannels(r.Context(), orgID, "domain", monitor.ID, req.ChannelIDs); err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
			return
		}
	} else {
		h.attachDefaultNotificationChannels(r.Context(), orgID, "domain", monitor.ID)
	}

	resp := domainMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "domain", monitor.ID)
	respond.JSON(w, http.StatusCreated, resp)
}

// checkDomainMonitorLimit enforces the org's plan limit before creating a
// domain monitor, writing the appropriate error response itself so the
// caller can just check the returned bool.
func (h *MonitorHandler) checkDomainMonitorLimit(w http.ResponseWriter, r *http.Request, orgID uuid.UUID) bool {
	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return false
	}
	total, err := h.queries.CountOrgMonitors(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return false
	}
	if err := billing.CheckMonitorLimit(plan, int(total)); err != nil {
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "domain_monitor")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return false
	}
	return true
}

// GetDomainMonitor GET /api/v1/monitors/domain/{id}
func (h *MonitorHandler) GetDomainMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := domainMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.GetDomainMonitor(r.Context(), db.GetDomainMonitorParams{
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

	resp := domainMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "domain", monitor.ID)
	respond.JSON(w, http.StatusOK, resp)
}

type updateDomainMonitorRequest struct {
	Name                 string   `json:"name"`
	Domain               string   `json:"domain"`
	AlertsEnabled        bool     `json:"alertsEnabled"`
	AlertAfterNFailures  int32    `json:"alertAfterNFailures"`
	MaxAlertsPerIncident int32    `json:"maxAlertsPerIncident"`
	ChannelIDs           []string `json:"channelIds"`
}

// UpdateDomainMonitor PATCH /api/v1/monitors/domain/{id}
func (h *MonitorHandler) UpdateDomainMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := domainMonitorIDs(w, r)
	if !ok {
		return
	}

	var req updateDomainMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}
	domain, err := parseHostname(req.Domain)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}

	monitor, err := h.queries.UpdateDomainMonitor(r.Context(), db.UpdateDomainMonitorParams{
		ID:                   monitorID,
		OrgID:                orgID,
		Name:                 req.Name,
		Domain:               domain,
		AlertsEnabled:        req.AlertsEnabled,
		AlertAfterNFailures:  req.AlertAfterNFailures,
		MaxAlertsPerIncident: req.MaxAlertsPerIncident,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if err := h.setMonitorNotificationChannels(r.Context(), orgID, "domain", monitorID, req.ChannelIDs); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := domainMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "domain", monitorID)
	respond.JSON(w, http.StatusOK, resp)
}

// PauseDomainMonitor POST /api/v1/monitors/domain/{id}/pause
func (h *MonitorHandler) PauseDomainMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := domainMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.PauseDomainMonitor(r.Context(), db.PauseDomainMonitorParams{
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
	respond.JSON(w, http.StatusOK, domainMonitorToResponse(monitor))
}

// ResumeDomainMonitor POST /api/v1/monitors/domain/{id}/resume
func (h *MonitorHandler) ResumeDomainMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := domainMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.ResumeDomainMonitor(r.Context(), db.ResumeDomainMonitorParams{
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
	respond.JSON(w, http.StatusOK, domainMonitorToResponse(monitor))
}

// DeleteDomainMonitor DELETE /api/v1/monitors/domain/{id}
func (h *MonitorHandler) DeleteDomainMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := domainMonitorIDs(w, r)
	if !ok {
		return
	}

	if err := h.queries.DeleteDomainMonitor(r.Context(), db.DeleteDomainMonitorParams{
		ID: monitorID, OrgID: orgID,
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
