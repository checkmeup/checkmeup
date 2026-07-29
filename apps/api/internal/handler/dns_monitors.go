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
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/billing"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

// ─── response types ──────────────────────────────────────────────────────────

type dnsMonitorResponse struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Hostname             string   `json:"hostname"`
	RecordType           string   `json:"recordType"`
	ExpectedValue        *string  `json:"expectedValue"`
	BaselineCaptured     bool     `json:"baselineCaptured"`
	LastResolvedValue    *string  `json:"lastResolvedValue"`
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

type dnsCheckResponse struct {
	ID             string  `json:"id"`
	CheckedAt      string  `json:"checkedAt"`
	ResponseTimeMs int32   `json:"responseTimeMs"`
	IsUp           bool    `json:"isUp"`
	ResolvedValue  *string `json:"resolvedValue"`
	FailureReason  *string `json:"failureReason"`
}

type dnsIncidentResponse struct {
	ID         string  `json:"id"`
	StartedAt  string  `json:"startedAt"`
	ResolvedAt *string `json:"resolvedAt"`
}

type dnsStatsResponse struct {
	Uptime24h *float64 `json:"uptime24h"`
	Uptime7d  *float64 `json:"uptime7d"`
	Uptime30d *float64 `json:"uptime30d"`
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (h *MonitorHandler) dnsMonitorToResponse(m db.DnsMonitor) dnsMonitorResponse {
	r := dnsMonitorResponse{
		ID:                   m.ID.String(),
		Name:                 m.Name,
		Hostname:             m.Hostname,
		RecordType:           string(m.RecordType),
		BaselineCaptured:     m.BaselineCaptured,
		IntervalMins:         m.IntervalMins,
		Status:               string(m.Status),
		AlertsEnabled:        m.AlertsEnabled,
		MaxAlertsPerIncident: m.MaxAlertsPerIncident,
		AlertAfterNFailures:  m.AlertAfterNFailures,
		CreatedAt:            m.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if m.ExpectedValue.Valid {
		r.ExpectedValue = &m.ExpectedValue.String
	}
	if m.LastResolvedValue.Valid {
		r.LastResolvedValue = &m.LastResolvedValue.String
	}
	if m.LastCheckedAt.Valid {
		t := m.LastCheckedAt.Time.Format("2006-01-02T15:04:05Z")
		r.LastCheckedAt = &t
	}
	return r
}

func dnsCheckToResponse(c db.DnsCheck) dnsCheckResponse {
	r := dnsCheckResponse{
		ID:             c.ID.String(),
		CheckedAt:      c.CheckedAt.Time.Format("2006-01-02T15:04:05Z"),
		ResponseTimeMs: c.ResponseTimeMs,
		IsUp:           c.IsUp,
	}
	if c.ResolvedValue.Valid {
		r.ResolvedValue = &c.ResolvedValue.String
	}
	if c.FailureReason.Valid {
		r.FailureReason = &c.FailureReason.String
	}
	return r
}

func dnsIncidentToResponse(i db.DnsIncident) dnsIncidentResponse {
	r := dnsIncidentResponse{
		ID:        i.ID.String(),
		StartedAt: i.StartedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if i.ResolvedAt.Valid {
		t := i.ResolvedAt.Time.Format("2006-01-02T15:04:05Z")
		r.ResolvedAt = &t
	}
	return r
}

func dnsMonitorIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
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

var validDNSRecordTypes = map[string]db.DnsRecordType{
	string(db.DnsRecordTypeA):     db.DnsRecordTypeA,
	string(db.DnsRecordTypeAAAA):  db.DnsRecordTypeAAAA,
	string(db.DnsRecordTypeCNAME): db.DnsRecordTypeCNAME,
	string(db.DnsRecordTypeMX):    db.DnsRecordTypeMX,
	string(db.DnsRecordTypeTXT):   db.DnsRecordTypeTXT,
	string(db.DnsRecordTypeNS):    db.DnsRecordTypeNS,
}

func validateRecordType(raw string) (db.DnsRecordType, error) {
	rt, ok := validDNSRecordTypes[strings.ToUpper(strings.TrimSpace(raw))]
	if !ok {
		return "", errors.New("record type must be one of A, AAAA, CNAME, MX, TXT, NS")
	}
	return rt, nil
}

// validateDNSFields trims name and validates hostname/record type — shared
// by create and update. expectedValue is normalized separately since empty
// is a valid, meaningful input (baseline mode), not a validation error.
func validateDNSFields(name, rawHostname, rawRecordType string) (trimmedName, hostname string, recordType db.DnsRecordType, err error) {
	trimmedName = strings.TrimSpace(name)
	if trimmedName == "" {
		return "", "", "", errors.New("name is required")
	}
	hostname, err = parseHostname(rawHostname)
	if err != nil {
		return "", "", "", err
	}
	recordType, err = validateRecordType(rawRecordType)
	if err != nil {
		return "", "", "", err
	}
	return trimmedName, hostname, recordType, nil
}

// normalizeExpectedValue trims input and reports it as a pgtype.Text —
// empty means "not pinned" (baseline mode), not an error.
func normalizeExpectedValue(raw string) pgtype.Text {
	trimmed := strings.TrimSpace(raw)
	return pgtype.Text{String: trimmed, Valid: trimmed != ""}
}

// normalizeDNSMaxAlerts applies the same "negative means default" rule
// used by every other monitor type's create/update handler.
func normalizeDNSMaxAlerts(v int32) int32 {
	if v < 0 {
		return 3
	}
	return v
}

// getPlanAndCheckDNSLimit fetches the org's plan and verifies it hasn't hit
// its aggregate monitor limit — used only by create, since update doesn't
// add a new monitor.
func (h *MonitorHandler) getPlanAndCheckDNSLimit(w http.ResponseWriter, r *http.Request, orgID uuid.UUID) (db.Plan, bool) {
	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return "", false
	}
	total, err := h.queries.CountOrgMonitors(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return "", false
	}
	if err := billing.CheckMonitorLimit(plan, int(total)); err != nil {
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "dns_monitor")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return "", false
	}
	return plan, true
}

// clampDNSInterval enforces the plan's minimum check interval, writing a
// 402 response when the requested interval is too low.
func (h *MonitorHandler) clampDNSInterval(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, plan db.Plan, requestedMins int32) (int32, bool) {
	clamped, err := billing.ClampInterval(plan, int(requestedMins))
	if err != nil {
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "dns_monitor_interval")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return 0, false
	}
	return int32(clamped), true
}

// getPlanAndClampDNSInterval fetches the org's plan and clamps the
// requested interval in one step — used by update, which (unlike create)
// doesn't also need the plan for a monitor-limit check.
func (h *MonitorHandler) getPlanAndClampDNSInterval(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, requestedMins int32) (int32, bool) {
	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return 0, false
	}
	return h.clampDNSInterval(w, r, orgID, plan, requestedMins)
}

// respondDNSMonitorWriteError maps a failed lookup-and-write query error to
// the right HTTP status — 404 when the monitor doesn't exist (or belongs to
// another org), 500 otherwise.
func (h *MonitorHandler) respondDNSMonitorWriteError(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		respond.Error(w, http.StatusNotFound, "monitor not found", "not_found")
		return
	}
	respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
}

// attachDNSMonitorChannels attaches the requested channels to a newly
// created monitor, or falls back to every enabled org channel when none
// were explicitly selected (US-2802).
func (h *MonitorHandler) attachDNSMonitorChannels(r *http.Request, orgID, monitorID uuid.UUID, channelIDs []string) error {
	if len(channelIDs) > 0 {
		return h.setMonitorNotificationChannels(r.Context(), orgID, "dns", monitorID, channelIDs)
	}
	h.attachDefaultNotificationChannels(r.Context(), orgID, "dns", monitorID)
	return nil
}

// ─── handlers ────────────────────────────────────────────────────────────────

// ListDNSMonitors GET /api/v1/monitors/dns
func (h *MonitorHandler) ListDNSMonitors(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	monitors, err := h.queries.ListDNSMonitors(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	result := make([]dnsMonitorResponse, len(monitors))
	for i, m := range monitors {
		r2 := h.dnsMonitorToResponse(m)
		if stats, err := h.queries.GetDNSStats(r.Context(), m.ID); err == nil {
			r2.Uptime24h = uptimePct(stats.Up24h, stats.Total24h)
		}
		result[i] = r2
	}
	respond.JSON(w, http.StatusOK, result)
}

type createDNSMonitorRequest struct {
	Name                 string   `json:"name"`
	Hostname             string   `json:"hostname"`
	RecordType           string   `json:"recordType"`
	ExpectedValue        string   `json:"expectedValue"`
	IntervalMins         int32    `json:"intervalMins"`
	MaxAlertsPerIncident int32    `json:"maxAlertsPerIncident"`
	AlertAfterNFailures  int32    `json:"alertAfterNFailures"`
	ChannelIDs           []string `json:"channelIds"`
}

// CreateDNSMonitor POST /api/v1/monitors/dns
func (h *MonitorHandler) CreateDNSMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req createDNSMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	name, hostname, recordType, err := validateDNSFields(req.Name, req.Hostname, req.RecordType)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	req.Name = name

	plan, ok := h.getPlanAndCheckDNSLimit(w, r, orgID)
	if !ok {
		return
	}
	clampedInterval, ok := h.clampDNSInterval(w, r, orgID, plan, req.IntervalMins)
	if !ok {
		return
	}
	req.IntervalMins = clampedInterval
	req.MaxAlertsPerIncident = normalizeDNSMaxAlerts(req.MaxAlertsPerIncident)

	monitor, err := h.queries.CreateDNSMonitor(r.Context(), db.CreateDNSMonitorParams{
		OrgID:                orgID,
		Name:                 req.Name,
		Hostname:             hostname,
		RecordType:           recordType,
		ExpectedValue:        normalizeExpectedValue(req.ExpectedValue),
		IntervalMins:         req.IntervalMins,
		MaxAlertsPerIncident: req.MaxAlertsPerIncident,
		AlertAfterNFailures:  req.AlertAfterNFailures,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	if err := h.attachDNSMonitorChannels(r, orgID, monitor.ID, req.ChannelIDs); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := h.dnsMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "dns", monitor.ID)
	respond.JSON(w, http.StatusCreated, resp)
}

// GetDNSMonitor GET /api/v1/monitors/dns/{id}
func (h *MonitorHandler) GetDNSMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := dnsMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.GetDNSMonitor(r.Context(), db.GetDNSMonitorParams{
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

	detail, err := h.loadDNSMonitorDetail(r.Context(), monitorID, parsePageParam(r))
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := h.dnsMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "dns", monitor.ID)

	respond.JSON(w, http.StatusOK, map[string]any{
		"monitor":   resp,
		"chartData": detail.chart,
		"checks":    detail.checks,
		"incidents": detail.incidents,
		"stats": dnsStatsResponse{
			Uptime24h: uptimePct(detail.stats.Up24h, detail.stats.Total24h),
			Uptime7d:  uptimePct(detail.stats.Up7d, detail.stats.Total7d),
			Uptime30d: uptimePct(detail.stats.Up30d, detail.stats.Total30d),
		},
	})
}

type dnsMonitorDetail struct {
	chart     []dnsCheckResponse
	checks    []dnsCheckResponse
	incidents []dnsIncidentResponse
	stats     db.GetDNSStatsRow
}

// loadDNSMonitorDetail fetches the chart/checks/incidents/stats data shown
// on the DNS monitor detail page and converts it to response types.
func (h *MonitorHandler) loadDNSMonitorDetail(ctx context.Context, monitorID uuid.UUID, page int32) (dnsMonitorDetail, error) {
	chartData, err := h.queries.ListDNSChecks24h(ctx, monitorID)
	if err != nil {
		return dnsMonitorDetail{}, err
	}

	checks, err := h.queries.ListDNSChecks(ctx, db.ListDNSChecksParams{
		MonitorID: monitorID,
		Limit:     50,
		Offset:    page * 50,
	})
	if err != nil {
		return dnsMonitorDetail{}, err
	}

	incidents, err := h.queries.ListDNSIncidents(ctx, monitorID)
	if err != nil {
		return dnsMonitorDetail{}, err
	}

	stats, err := h.queries.GetDNSStats(ctx, monitorID)
	if err != nil {
		return dnsMonitorDetail{}, err
	}

	return dnsMonitorDetail{
		chart:     dnsChecksToResponse(chartData),
		checks:    dnsChecksToResponse(checks),
		incidents: dnsIncidentsToResponse(incidents),
		stats:     stats,
	}, nil
}

func dnsChecksToResponse(checks []db.DnsCheck) []dnsCheckResponse {
	resp := make([]dnsCheckResponse, len(checks))
	for i, c := range checks {
		resp[i] = dnsCheckToResponse(c)
	}
	return resp
}

func dnsIncidentsToResponse(incidents []db.DnsIncident) []dnsIncidentResponse {
	resp := make([]dnsIncidentResponse, len(incidents))
	for i, inc := range incidents {
		resp[i] = dnsIncidentToResponse(inc)
	}
	return resp
}

type updateDNSMonitorRequest struct {
	Name                 string   `json:"name"`
	Hostname             string   `json:"hostname"`
	RecordType           string   `json:"recordType"`
	ExpectedValue        string   `json:"expectedValue"`
	IntervalMins         int32    `json:"intervalMins"`
	AlertsEnabled        bool     `json:"alertsEnabled"`
	MaxAlertsPerIncident int32    `json:"maxAlertsPerIncident"`
	AlertAfterNFailures  int32    `json:"alertAfterNFailures"`
	ChannelIDs           []string `json:"channelIds"`
}

// UpdateDNSMonitor PATCH /api/v1/monitors/dns/{id}
func (h *MonitorHandler) UpdateDNSMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := dnsMonitorIDs(w, r)
	if !ok {
		return
	}

	var req updateDNSMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	name, hostname, recordType, err := validateDNSFields(req.Name, req.Hostname, req.RecordType)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	req.Name = name

	clampedInterval, ok := h.getPlanAndClampDNSInterval(w, r, orgID, req.IntervalMins)
	if !ok {
		return
	}
	req.IntervalMins = clampedInterval
	req.MaxAlertsPerIncident = normalizeDNSMaxAlerts(req.MaxAlertsPerIncident)

	monitor, err := h.queries.UpdateDNSMonitor(r.Context(), db.UpdateDNSMonitorParams{
		ID:                   monitorID,
		OrgID:                orgID,
		Name:                 req.Name,
		Hostname:             hostname,
		RecordType:           recordType,
		ExpectedValue:        normalizeExpectedValue(req.ExpectedValue),
		IntervalMins:         req.IntervalMins,
		AlertsEnabled:        req.AlertsEnabled,
		MaxAlertsPerIncident: req.MaxAlertsPerIncident,
		AlertAfterNFailures:  req.AlertAfterNFailures,
	})
	if err != nil {
		h.respondDNSMonitorWriteError(w, err)
		return
	}
	if err := h.setMonitorNotificationChannels(r.Context(), orgID, "dns", monitorID, req.ChannelIDs); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	resp := h.dnsMonitorToResponse(monitor)
	resp.ChannelIDs = h.loadNotificationChannelIDs(r.Context(), "dns", monitorID)
	respond.JSON(w, http.StatusOK, resp)
}

// PauseDNSMonitor POST /api/v1/monitors/dns/{id}/pause
func (h *MonitorHandler) PauseDNSMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := dnsMonitorIDs(w, r)
	if !ok {
		return
	}

	monitor, err := h.queries.PauseDNSMonitor(r.Context(), db.PauseDNSMonitorParams{
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
	respond.JSON(w, http.StatusOK, h.dnsMonitorToResponse(monitor))
}

// ResumeDNSMonitor POST /api/v1/monitors/dns/{id}/resume
func (h *MonitorHandler) ResumeDNSMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := dnsMonitorIDs(w, r)
	if !ok {
		return
	}
	if err := h.checkMonitorResumeLimit(r.Context(), orgID); err != nil {
		respondResumeLimitErr(w, err)
		return
	}

	monitor, err := h.queries.ResumeDNSMonitor(r.Context(), db.ResumeDNSMonitorParams{
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
	respond.JSON(w, http.StatusOK, h.dnsMonitorToResponse(monitor))
}

// DeleteDNSMonitor DELETE /api/v1/monitors/dns/{id}
func (h *MonitorHandler) DeleteDNSMonitor(w http.ResponseWriter, r *http.Request) {
	orgID, monitorID, ok := dnsMonitorIDs(w, r)
	if !ok {
		return
	}

	if err := h.queries.DeleteDNSMonitor(r.Context(), db.DeleteDNSMonitorParams{
		ID: monitorID, OrgID: orgID,
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
