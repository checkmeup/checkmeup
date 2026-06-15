package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/config"
	"github.com/checkmeup/checkmeup/internal/db"
	apimiddleware "github.com/checkmeup/checkmeup/internal/middleware"
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
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	Schedule             string  `json:"schedule"`
	GracePeriodMins      int32   `json:"gracePeriodMins"`
	PingToken            string  `json:"pingToken"`
	PingURL              string  `json:"pingUrl"`
	Status               string  `json:"status"`
	AlertsEnabled        bool    `json:"alertsEnabled"`
	MaxAlertsPerIncident int32   `json:"maxAlertsPerIncident"`
	LastPingAt           *string `json:"lastPingAt"`
	NextPingAt           *string `json:"nextPingAt"`
	CreatedAt            string  `json:"createdAt"`
}

type cronPingResponse struct {
	ID         string `json:"id"`
	ReceivedAt string `json:"receivedAt"`
	SourceIP   string `json:"sourceIp"`
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

func orgIDFrom(r *http.Request) (uuid.UUID, error) {
	claims := apimiddleware.ClaimsFrom(r.Context())
	if claims == nil {
		return uuid.UUID{}, errors.New("no claims")
	}
	return uuid.Parse(claims.OrgID)
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
	Name                 string `json:"name"`
	Schedule             string `json:"schedule"`
	GracePeriodMins      int32  `json:"gracePeriodMins"`
	MaxAlertsPerIncident int32  `json:"maxAlertsPerIncident"`
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

	req.Name = strings.TrimSpace(req.Name)
	req.Schedule = strings.TrimSpace(req.Schedule)

	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}
	if err := validateSchedule(req.Schedule); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	if req.GracePeriodMins < 1 {
		req.GracePeriodMins = 5
	}
	if req.MaxAlertsPerIncident < 0 {
		req.MaxAlertsPerIncident = 3
	}

	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	token := hex.EncodeToString(tokenBytes)

	monitor, err := h.queries.CreateCronMonitor(r.Context(), db.CreateCronMonitorParams{
		OrgID:                orgID,
		Name:                 req.Name,
		Schedule:             req.Schedule,
		GracePeriodMins:      req.GracePeriodMins,
		PingToken:            token,
		MaxAlertsPerIncident: req.MaxAlertsPerIncident,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	respond.JSON(w, http.StatusCreated, h.monitorToResponse(monitor))
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

	pings, err := h.queries.ListCronPings(r.Context(), db.ListCronPingsParams{
		MonitorID: monitorID,
		Limit:     50,
		Offset:    0,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	incidents, err := h.queries.ListCronIncidents(r.Context(), monitorID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	pingResp := make([]cronPingResponse, len(pings))
	for i, p := range pings {
		pingResp[i] = cronPingResponse{
			ID:         p.ID.String(),
			ReceivedAt: p.ReceivedAt.Time.Format("2006-01-02T15:04:05Z"),
			SourceIP:   p.SourceIp,
		}
	}

	incidentResp := make([]cronIncidentResponse, len(incidents))
	for i, inc := range incidents {
		ir := cronIncidentResponse{
			ID:        inc.ID.String(),
			StartedAt: inc.StartedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if inc.ResolvedAt.Valid {
			t := inc.ResolvedAt.Time.Format("2006-01-02T15:04:05Z")
			ir.ResolvedAt = &t
		}
		incidentResp[i] = ir
	}

	respond.JSON(w, http.StatusOK, map[string]any{
		"monitor":   h.monitorToResponse(monitor),
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
		result[i] = cronPingResponse{
			ID:         p.ID.String(),
			ReceivedAt: p.ReceivedAt.Time.Format("2006-01-02T15:04:05Z"),
			SourceIP:   p.SourceIp,
		}
	}
	respond.JSON(w, http.StatusOK, result)
}

type updateCronMonitorRequest struct {
	Name                 string `json:"name"`
	Schedule             string `json:"schedule"`
	GracePeriodMins      int32  `json:"gracePeriodMins"`
	AlertsEnabled        bool   `json:"alertsEnabled"`
	MaxAlertsPerIncident int32  `json:"maxAlertsPerIncident"`
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

	req.Name = strings.TrimSpace(req.Name)
	req.Schedule = strings.TrimSpace(req.Schedule)

	if req.Name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return
	}
	if err := validateSchedule(req.Schedule); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	if req.GracePeriodMins < 1 {
		req.GracePeriodMins = 5
	}
	if req.MaxAlertsPerIncident < 0 {
		req.MaxAlertsPerIncident = 3
	}

	monitor, err := h.queries.UpdateCronMonitor(r.Context(), db.UpdateCronMonitorParams{
		ID:                   monitorID,
		OrgID:                orgID,
		Name:                 req.Name,
		Schedule:             req.Schedule,
		GracePeriodMins:      req.GracePeriodMins,
		AlertsEnabled:        req.AlertsEnabled,
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

	respond.JSON(w, http.StatusOK, h.monitorToResponse(monitor))
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

func monitorIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
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

func validateSchedule(s string) error {
	if s == "" {
		return fmt.Errorf("schedule is required")
	}
	s = strings.ToLower(s)

	// Plain interval: "every 1m", "every 30m", "every 1h", "every 12h", "every 1d"
	if strings.HasPrefix(s, "every ") {
		return nil
	}

	// Cron expression: 5 whitespace-separated fields
	fields := strings.Fields(s)
	if len(fields) == 5 {
		return nil
	}

	return fmt.Errorf("schedule must be a cron expression (e.g. \"0 * * * *\") or interval (e.g. \"every 1h\")")
}
