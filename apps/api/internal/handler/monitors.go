package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

type MonitorHandler struct {
	cfg     *config.Config
	queries *db.Queries
	tg      *telegram.Client
}

func NewMonitorHandler(cfg *config.Config, pool *pgxpool.Pool, tg *telegram.Client) *MonitorHandler {
	return &MonitorHandler{cfg: cfg, queries: db.New(pool), tg: tg}
}

type cronMonitorResponse struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Schedule             string   `json:"schedule"`
	GracePeriodMins      int32    `json:"gracePeriodMins"`
	PingToken            string   `json:"pingToken"`
	PingURL              string   `json:"pingUrl"`
	Status               string   `json:"status"`
	AlertsEnabled        bool     `json:"alertsEnabled"`
	MaxAlertsPerIncident int32    `json:"maxAlertsPerIncident"`
	AlertAfterNFailures  int32    `json:"alertAfterNFailures"`
	LastPingAt           *string  `json:"lastPingAt"`
	NextPingAt           *string  `json:"nextPingAt"`
	CreatedAt            string   `json:"createdAt"`
	ChannelIDs           []string `json:"channelIds,omitempty"`
}

type cronPingResponse struct {
	ID         string            `json:"id"`
	ReceivedAt string            `json:"receivedAt"`
	SourceIP   string            `json:"sourceIp"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type cronIncidentResponse struct {
	ID         string  `json:"id"`
	StartedAt  string  `json:"startedAt"`
	ResolvedAt *string `json:"resolvedAt"`
}

func (h *MonitorHandler) monitorToResponse(m db.CronMonitor) cronMonitorResponse {
	r := cronMonitorResponse{
		ID:                   m.ID.String(),
		Name:                 m.Name,
		Schedule:             m.Schedule,
		GracePeriodMins:      m.GracePeriodMins,
		PingToken:            m.PingToken,
		PingURL:              fmt.Sprintf("%s/ping/%s", h.cfg.BaseURL, m.PingToken),
		Status:               string(m.Status),
		AlertsEnabled:        m.AlertsEnabled,
		MaxAlertsPerIncident: m.MaxAlertsPerIncident,
		AlertAfterNFailures:  m.AlertAfterNFailures,
		CreatedAt:            m.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if m.LastPingAt.Valid {
		t := m.LastPingAt.Time.Format("2006-01-02T15:04:05Z")
		r.LastPingAt = &t
	}
	if m.NextPingAt.Valid {
		t := m.NextPingAt.Time.Format("2006-01-02T15:04:05Z")
		r.NextPingAt = &t
	}
	return r
}

// ListCronMonitors GET /api/v1/monitors/cron
func (h *MonitorHandler) ListCronMonitors(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	monitors, err := h.queries.ListCronMonitors(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	result := make([]cronMonitorResponse, len(monitors))
	for i, m := range monitors {
		result[i] = h.monitorToResponse(m)
	}
	respond.JSON(w, http.StatusOK, result)
}

type createCronMonitorRequest struct {
	Name                 string   `json:"name"`
	Schedule             string   `json:"schedule"`
	GracePeriodMins      int32    `json:"gracePeriodMins"`
	MaxAlertsPerIncident int32    `json:"maxAlertsPerIncident"`
	AlertAfterNFailures  int32    `json:"alertAfterNFailures"`
	ChannelIDs           []string `json:"channelIds"`
}

// CreateCronMonitor POST /api/v1/monitors/cron
func (h *MonitorHandler) CreateCronMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req createCronMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	if err := normalizeAndValidateCreateCronMonitorRequest(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}

	if err := h.checkMonitorCreateLimit(r.Context(), orgID); err != nil {
		respondMonitorCreateLimitErr(w, r, orgID, "cron_monitor", err)
		return
	}

	monitor, err := h.createCronMonitorRow(r.Context(), orgID, req)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if err := h.attachMonitorChannels(r.Context(), orgID, "cron", monitor.ID, req.ChannelIDs); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := h.monitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "cron", monitor.ID)
	respond.JSON(w, http.StatusCreated, resp)
}

// GetCronMonitor GET /api/v1/monitors/cron/{id}
func (h *MonitorHandler) GetCronMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := monitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.GetCronMonitor(r.Context(), db.GetCronMonitorParams{
		ID:    monitorID,
		OrgID: orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	pingResp, incidentResp, err := h.loadCronMonitorDetail(r.Context(), monitorID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := h.monitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "cron", monitor.ID)

	respond.JSON(w, http.StatusOK, map[string]any{
		"monitor":   resp,
		"pings":     pingResp,
		"incidents": incidentResp,
	})
}

// GetCronPings GET /api/v1/monitors/cron/{id}/pings
func (h *MonitorHandler) GetCronPings(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := monitorIDs(w, r)
	if !ok {
		return
	}

	// Verify the monitor belongs to this org.
	if _, err := h.queries.GetCronMonitor(r.Context(), db.GetCronMonitorParams{
		ID: monitorID, OrgID: orgID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	page := int32(0)
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = int32(n - 1)
		}
	}
	const pageSize = 50

	pings, err := h.queries.ListCronPings(r.Context(), db.ListCronPingsParams{
		MonitorID: monitorID,
		Limit:     pageSize,
		Offset:    page * pageSize,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	result := make([]cronPingResponse, len(pings))
	for i, p := range pings {
		result[i] = cronPingToResponse(p)
	}
	respond.JSON(w, http.StatusOK, result)
}

type updateCronMonitorRequest struct {
	Name                 string   `json:"name"`
	Schedule             string   `json:"schedule"`
	GracePeriodMins      int32    `json:"gracePeriodMins"`
	AlertsEnabled        bool     `json:"alertsEnabled"`
	MaxAlertsPerIncident int32    `json:"maxAlertsPerIncident"`
	AlertAfterNFailures  int32    `json:"alertAfterNFailures"`
	ChannelIDs           []string `json:"channelIds"`
}

// UpdateCronMonitor PATCH /api/v1/monitors/cron/{id}
func (h *MonitorHandler) UpdateCronMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := monitorIDs(w, r)
	if !ok {
		return
	}

	var req updateCronMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	if err := normalizeAndValidateUpdateCronMonitorRequest(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}

	monitor, err := h.queries.UpdateCronMonitor(r.Context(), db.UpdateCronMonitorParams{
		ID:                   monitorID,
		OrgID:                orgID,
		Name:                 req.Name,
		Schedule:             req.Schedule,
		GracePeriodMins:      req.GracePeriodMins,
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
	if err := h.setMonitorNotificationChannels(r.Context(), orgID, "cron", monitorID, req.ChannelIDs); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := h.monitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "cron", monitorID)
	respond.JSON(w, http.StatusOK, resp)
}

// PauseCronMonitor POST /api/v1/monitors/cron/{id}/pause
func (h *MonitorHandler) PauseCronMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := monitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.PauseCronMonitor(r.Context(), db.PauseCronMonitorParams{
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
	respond.JSON(w, http.StatusOK, h.monitorToResponse(monitor))
}

// ResumeCronMonitor POST /api/v1/monitors/cron/{id}/resume
func (h *MonitorHandler) ResumeCronMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := monitorIDs(w, r)
	if !ok {
		return
	}
	if err := h.checkMonitorResumeLimit(r.Context(), orgID); err != nil {
		respondResumeLimitErr(w, err)
		return
	}

	monitor, err := h.queries.ResumeCronMonitor(r.Context(), db.ResumeCronMonitorParams{
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
	respond.JSON(w, http.StatusOK, h.monitorToResponse(monitor))
}

// DeleteCronMonitor DELETE /api/v1/monitors/cron/{id}
func (h *MonitorHandler) DeleteCronMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := monitorIDs(w, r)
	if !ok {
		return
	}

	if err := h.queries.DeleteCronMonitor(r.Context(), db.DeleteCronMonitorParams{
		ID: monitorID, OrgID: orgID,
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
