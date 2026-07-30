package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/billing"
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
	MaxDurationMins      *int32   `json:"maxDurationMins"`
	LastPingAt           *string  `json:"lastPingAt"`
	NextPingAt           *string  `json:"nextPingAt"`
	CreatedAt            string   `json:"createdAt"`
	ChannelIDs           []string `json:"channelIds,omitempty"`
}

type cronPingResponse struct {
	ID           string            `json:"id"`
	ReceivedAt   string            `json:"receivedAt"`
	SourceIP     string            `json:"sourceIp"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	RunStartedAt *string           `json:"runStartedAt"`
}

func cronPingToResponse(p db.CronPing) cronPingResponse {
	r := cronPingResponse{
		ID:         p.ID.String(),
		ReceivedAt: p.ReceivedAt.Time.Format("2006-01-02T15:04:05Z"),
		SourceIP:   p.SourceIp,
	}
	if len(p.Metadata) > 0 {
		meta := map[string]string{}
		if json.Unmarshal(p.Metadata, &meta) == nil {
			r.Metadata = meta
		}
	}
	if p.RunStartedAt.Valid {
		t := p.RunStartedAt.Time.Format("2006-01-02T15:04:05Z")
		r.RunStartedAt = &t
	}
	return r
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
	if m.MaxDurationMins.Valid {
		v := m.MaxDurationMins.Int32
		r.MaxDurationMins = &v
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

// attachDefaultNotificationChannels attaches every enabled channel the org
// currently has to a newly created monitor — a new monitor defaults to all
// of the org's enabled channels (US-2802), matching the pre-EP-28 implicit
// behavior of every monitor alerting on every org-level channel.
func (h *MonitorHandler) attachDefaultNotificationChannels(ctx context.Context, orgID uuid.UUID, monitorType string, monitorID uuid.UUID) {
	channels, err := h.queries.ListEnabledNotificationChannels(ctx, orgID)
	if err != nil {
		return
	}
	for _, c := range channels {
		_ = h.queries.InsertMonitorNotificationChannel(ctx, db.InsertMonitorNotificationChannelParams{
			ChannelID: c.ID, MonitorType: monitorType, MonitorID: monitorID,
		})
	}
}

// setMonitorNotificationChannels replaces a monitor's attached channels with
// channelIDs, dropping any ID that doesn't resolve to a channel owned by
// orgID — same ownership-scoping approach as resolveMonitorName.
func (h *MonitorHandler) setMonitorNotificationChannels(ctx context.Context, orgID uuid.UUID, monitorType string, monitorID uuid.UUID, channelIDs []string) error {
	owned, err := h.queries.ListNotificationChannels(ctx, orgID)
	if err != nil {
		return err
	}
	ownedSet := make(map[uuid.UUID]bool, len(owned))
	for _, c := range owned {
		ownedSet[c.ID] = true
	}

	if err := h.queries.DeleteMonitorNotificationChannels(ctx, db.DeleteMonitorNotificationChannelsParams{
		MonitorType: monitorType, MonitorID: monitorID,
	}); err != nil {
		return err
	}
	for _, idStr := range channelIDs {
		id, err := uuid.Parse(idStr)
		if err != nil || !ownedSet[id] {
			continue
		}
		if err := h.queries.InsertMonitorNotificationChannel(ctx, db.InsertMonitorNotificationChannelParams{
			ChannelID: id, MonitorType: monitorType, MonitorID: monitorID,
		}); err != nil {
			return err
		}
	}
	return nil
}

// applyMonitorNotificationChannels sets channelIDs on a newly created
// monitor, or falls back to every one of the org's enabled channels by
// default (US-2802) when none were specified. Shared by every monitor
// type's create handler.
func (h *MonitorHandler) applyMonitorNotificationChannels(ctx context.Context, orgID uuid.UUID, monitorType string, monitorID uuid.UUID, channelIDs []string) error {
	if len(channelIDs) > 0 {
		return h.setMonitorNotificationChannels(ctx, orgID, monitorType, monitorID, channelIDs)
	}
	h.attachDefaultNotificationChannels(ctx, orgID, monitorType, monitorID)
	return nil
}

// loadNotificationChannelIDs returns the channel IDs attached to a monitor,
// for inclusion in its API response (edit form pre-selection).
func (h *MonitorHandler) loadNotificationChannelIDs(ctx context.Context, monitorType string, monitorID uuid.UUID) []string {
	ids, err := h.queries.ListMonitorNotificationChannelIDs(ctx, db.ListMonitorNotificationChannelIDsParams{
		MonitorType: monitorType, MonitorID: monitorID,
	})
	if err != nil {
		return []string{}
	}
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}

// validateCronMonitorRequest validates and normalizes the fields shared by
// create and update requests, responding with the first validation failure
// and returning false — same "respond internally, return ok" idiom as
// validateUptimeMonitorRequest.
func validateCronMonitorRequest(w http.ResponseWriter, name, schedule *string, gracePeriodMins, maxAlertsPerIncident *int32, maxDurationMins *int32) bool {
	*name = strings.TrimSpace(*name)
	*schedule = strings.TrimSpace(*schedule)
	if *name == "" {
		respond.Error(w, http.StatusBadRequest, "name is required", "bad_request")
		return false
	}
	if err := validateSchedule(*schedule); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return false
	}
	if *gracePeriodMins < 1 {
		*gracePeriodMins = 5
	}
	if *maxAlertsPerIncident < 0 {
		*maxAlertsPerIncident = 3
	}
	// Unset (nil) means zombie detection is inactive (US-3402) — only a
	// provided-but-non-positive value is rejected, distinct from "disabled".
	if maxDurationMins != nil && *maxDurationMins <= 0 {
		respond.Error(w, http.StatusBadRequest, "maxDurationMins must be positive", "bad_request")
		return false
	}
	return true
}

// int32PtrToPgInt4 converts a nullable request field to sqlc's nullable
// int column type — nil stays NULL (US-3402's "disabled" state).
func int32PtrToPgInt4(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

// checkCronMonitorCreateLimit checks whether orgID can create another cron
// monitor under its plan. Responds and returns false on any failure (query
// error or plan limit hit).
func (h *MonitorHandler) checkCronMonitorCreateLimit(w http.ResponseWriter, r *http.Request, orgID uuid.UUID) bool {
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
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "cron_monitor")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return false
	}
	return true
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
	MaxDurationMins      *int32   `json:"maxDurationMins"`
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

	if !validateCronMonitorRequest(w, &req.Name, &req.Schedule, &req.GracePeriodMins, &req.MaxAlertsPerIncident, req.MaxDurationMins) {
		return
	}
	if !h.checkCronMonitorCreateLimit(w, r, orgID) {
		return
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
		AlertAfterNFailures:  req.AlertAfterNFailures,
		MaxDurationMins:      int32PtrToPgInt4(req.MaxDurationMins),
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if err := h.applyMonitorNotificationChannels(r.Context(), orgID, "cron", monitor.ID, req.ChannelIDs); err != nil {
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
	if respondMonitorNotFoundOrInternal(w, err) {
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
		pingResp[i] = cronPingToResponse(p)
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
	_, err := h.queries.GetCronMonitor(r.Context(), db.GetCronMonitorParams{
		ID: monitorID, OrgID: orgID,
	})
	if respondMonitorNotFoundOrInternal(w, err) {
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
	MaxDurationMins      *int32   `json:"maxDurationMins"`
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

	if !validateCronMonitorRequest(w, &req.Name, &req.Schedule, &req.GracePeriodMins, &req.MaxAlertsPerIncident, req.MaxDurationMins) {
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
		MaxDurationMins:      int32PtrToPgInt4(req.MaxDurationMins),
	})
	if respondMonitorNotFoundOrInternal(w, err) {
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
	if respondMonitorNotFoundOrInternal(w, err) {
		return
	}
	respond.JSON(w, http.StatusOK, h.monitorToResponse(monitor))
}

// checkMonitorResumeLimit returns billing.ErrMonitorLimit when resuming a
// paused monitor would push the org's active-monitor count over its plan
// limit (ADR-019) — after a downgrade, resuming doesn't create anything
// new, but it does grow the active count, so it's gated the same way
// creation is, just checked against CountActiveMonitorsForOrg (non-paused
// only) rather than the raw total CreateXMonitor checks. A non-nil error
// that isn't billing.ErrMonitorLimit means the check itself failed (DB
// error) — callers should treat that as an internal error, not a 402;
// distinguish with errors.Is(err, billing.ErrMonitorLimit).
func (h *MonitorHandler) checkMonitorResumeLimit(ctx context.Context, orgID uuid.UUID) error {
	plan, err := h.queries.GetOrgPlan(ctx, orgID)
	if err != nil {
		return err
	}
	active, err := h.queries.CountActiveMonitorsForOrg(ctx, orgID)
	if err != nil {
		return err
	}
	return billing.CheckMonitorLimit(plan, int(active))
}

// respondResumeLimitErr writes the right response for checkMonitorResumeLimit's
// error, telling a real plan-limit block (402) apart from an internal
// failure (500) that happened while checking.
func respondResumeLimitErr(w http.ResponseWriter, err error) {
	if errors.Is(err, billing.ErrMonitorLimit) {
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return
	}
	respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
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
	if respondMonitorNotFoundOrInternal(w, err) {
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
