package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/checkmeup/checkmeup/internal/db"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
	"github.com/checkmeup/checkmeup/internal/respond"
)

// monitorStatusResponse is the public API's read-only status projection —
// intentionally narrower than the session-authenticated detail responses in
// monitors.go/ssl_monitors.go/etc: this is an external contract (ADR-028)
// versioned like any other public endpoint (ADR-007), so it only carries
// what a third-party integration (CI dashboard, status LED) actually needs.
type monitorStatusResponse struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Type             string            `json:"type"`
	Status           string            `json:"status"`
	LastCheckedAt    *string           `json:"lastCheckedAt"`
	ExpiresAt        *string           `json:"expiresAt,omitempty"`
	DaysUntilExpiry  *int              `json:"daysUntilExpiry,omitempty"`
	LastPingMetadata map[string]string `json:"lastPingMetadata,omitempty"`
}

func publicMonitorIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	orgID := apimiddleware.OrgIDFromAPIKey(r.Context())
	monitorID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid monitor id", "bad_request")
		return uuid.UUID{}, uuid.UUID{}, false
	}
	return orgID, monitorID, true
}

// GetCronStatus handles GET /api/v1/public/monitors/cron/{id}/status.
func (h *MonitorHandler) GetCronStatus(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := publicMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.GetCronMonitor(r.Context(), db.GetCronMonitorParams{ID: monitorID, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}

	resp := monitorStatusResponse{
		ID:     monitor.ID.String(),
		Name:   monitor.Name,
		Type:   "cron",
		Status: string(monitor.Status),
	}
	if monitor.LastPingAt.Valid {
		t := monitor.LastPingAt.Time.Format("2006-01-02T15:04:05Z")
		resp.LastCheckedAt = &t
	}

	if ping, err := h.queries.GetLatestCronPing(r.Context(), monitorID); err == nil && len(ping.Metadata) > 0 {
		meta := map[string]string{}
		if json.Unmarshal(ping.Metadata, &meta) == nil {
			resp.LastPingMetadata = meta
		}
	}

	respond.JSON(w, http.StatusOK, resp)
}

// GetUptimeStatus handles GET /api/v1/public/monitors/uptime/{id}/status.
func (h *MonitorHandler) GetUptimeStatus(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := publicMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.GetUptimeMonitor(r.Context(), db.GetUptimeMonitorParams{ID: monitorID, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}

	resp := monitorStatusResponse{
		ID:     monitor.ID.String(),
		Name:   monitor.Name,
		Type:   "uptime",
		Status: string(monitor.Status),
	}
	if monitor.LastCheckedAt.Valid {
		t := monitor.LastCheckedAt.Time.Format("2006-01-02T15:04:05Z")
		resp.LastCheckedAt = &t
	}
	respond.JSON(w, http.StatusOK, resp)
}

// GetPortStatus handles GET /api/v1/public/monitors/port/{id}/status.
func (h *MonitorHandler) GetPortStatus(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := publicMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.GetPortMonitor(r.Context(), db.GetPortMonitorParams{ID: monitorID, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}

	resp := monitorStatusResponse{
		ID:     monitor.ID.String(),
		Name:   monitor.Name,
		Type:   "port",
		Status: string(monitor.Status),
	}
	if monitor.LastCheckedAt.Valid {
		t := monitor.LastCheckedAt.Time.Format("2006-01-02T15:04:05Z")
		resp.LastCheckedAt = &t
	}
	respond.JSON(w, http.StatusOK, resp)
}

// GetSSLStatus handles GET /api/v1/public/monitors/ssl/{id}/status.
func (h *MonitorHandler) GetSSLStatus(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := publicMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.GetSSLMonitor(r.Context(), db.GetSSLMonitorParams{ID: monitorID, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}

	resp := monitorStatusResponse{
		ID:     monitor.ID.String(),
		Name:   monitor.Name,
		Type:   "ssl",
		Status: string(monitor.Status),
	}
	if monitor.LastCheckedAt.Valid {
		t := monitor.LastCheckedAt.Time.Format("2006-01-02T15:04:05Z")
		resp.LastCheckedAt = &t
	}
	if monitor.ExpiresAt.Valid {
		t := monitor.ExpiresAt.Time.Format("2006-01-02T15:04:05Z")
		resp.ExpiresAt = &t
		days := int(time.Until(monitor.ExpiresAt.Time).Hours() / 24)
		resp.DaysUntilExpiry = &days
	}
	respond.JSON(w, http.StatusOK, resp)
}

// GetDomainStatus handles GET /api/v1/public/monitors/domain/{id}/status.
func (h *MonitorHandler) GetDomainStatus(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := publicMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.GetDomainMonitor(r.Context(), db.GetDomainMonitorParams{ID: monitorID, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}

	resp := monitorStatusResponse{
		ID:     monitor.ID.String(),
		Name:   monitor.Name,
		Type:   "domain",
		Status: string(monitor.Status),
	}
	if monitor.LastCheckedAt.Valid {
		t := monitor.LastCheckedAt.Time.Format("2006-01-02T15:04:05Z")
		resp.LastCheckedAt = &t
	}
	if monitor.ExpiresAt.Valid {
		t := monitor.ExpiresAt.Time.Format("2006-01-02T15:04:05Z")
		resp.ExpiresAt = &t
		days := int(time.Until(monitor.ExpiresAt.Time).Hours() / 24)
		resp.DaysUntilExpiry = &days
	}
	respond.JSON(w, http.StatusOK, resp)
}
