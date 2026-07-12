package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

type IncidentHandler struct {
	queries *db.Queries
}

func NewIncidentHandler(pool *pgxpool.Pool) *IncidentHandler {
	return &IncidentHandler{queries: db.New(pool)}
}

// ─── request/response types ──────────────────────────────────────────────────

type incidentMonitorInput struct {
	MonitorType string `json:"monitorType"`
	MonitorID   string `json:"monitorId"`
}

type createIncidentRequest struct {
	Title          string                 `json:"title"`
	Message        string                 `json:"message"`
	Severity       string                 `json:"severity"`
	Monitors       []incidentMonitorInput `json:"monitors"`
	ConfirmOverlap bool                   `json:"confirmOverlap"`
}

type updateTitleRequest struct {
	Title string `json:"title"`
}

type postUpdateRequest struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type updateMessageRequest struct {
	Message string `json:"message"`
}

type incidentMonitorRef struct {
	MonitorType string `json:"monitorType"`
	MonitorID   string `json:"monitorId"`
	Name        string `json:"name"`
}

type incidentUpdateResponse struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type incidentResponse struct {
	ID           string                   `json:"id"`
	Title        string                   `json:"title"`
	Severity     string                   `json:"severity"`
	Status       string                   `json:"status"`
	Monitors     []incidentMonitorRef     `json:"monitors,omitempty"`
	MonitorCount int                      `json:"monitorCount"`
	Updates      []incidentUpdateResponse `json:"updates,omitempty"`
	CreatedAt    string                   `json:"createdAt"`
	UpdatedAt    string                   `json:"updatedAt"`
	ResolvedAt   *string                  `json:"resolvedAt"`
}

var validSeverities = map[string]db.IncidentSeverity{
	"minor":    db.IncidentSeverityMinor,
	"major":    db.IncidentSeverityMajor,
	"critical": db.IncidentSeverityCritical,
}

var validStatuses = map[string]db.IncidentStatus{
	"investigating": db.IncidentStatusInvestigating,
	"identified":    db.IncidentStatusIdentified,
	"monitoring":    db.IncidentStatusMonitoring,
	"resolved":      db.IncidentStatusResolved,
}

func incidentIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return uuid.UUID{}, uuid.UUID{}, false
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid id", "bad_request")
		return uuid.UUID{}, uuid.UUID{}, false
	}
	return orgID, id, true
}

func ptrTime(t time.Time, valid bool) *string {
	if !valid {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func toListItemResponse(row db.ListStatusPageIncidentsRow) incidentResponse {
	return incidentResponse{
		ID:           row.ID.String(),
		Title:        row.Title,
		Severity:     string(row.Severity),
		Status:       string(row.Status),
		MonitorCount: int(row.MonitorCount),
		CreatedAt:    row.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    row.UpdatedAt.Time.Format(time.RFC3339),
		ResolvedAt:   ptrTime(row.ResolvedAt.Time, row.ResolvedAt.Valid),
	}
}

// buildIncidentDetail loads an incident's monitors (resolved to display names)
// and its full update feed — shared by every handler that returns a full
// incident object (Get, Create, PostUpdate, UpdateTitle, UpdateUpdateMessage)
// so the frontend always gets a consistent, fully-hydrated shape back.
func (h *IncidentHandler) buildIncidentDetail(ctx context.Context, orgID uuid.UUID, incident db.StatusPageIncident) (incidentResponse, error) {
	refs, err := h.queries.ListStatusPageIncidentMonitors(ctx, incident.ID)
	if err != nil {
		return incidentResponse{}, err
	}
	monitors := make([]incidentMonitorRef, 0, len(refs))
	for _, ref := range refs {
		name := "(deleted monitor)"
		if _, n, rerr := resolveMonitorName(ctx, h.queries, orgID, ref.MonitorType, ref.MonitorID.String()); rerr == nil {
			name = n
		}
		monitors = append(monitors, incidentMonitorRef{MonitorType: ref.MonitorType, MonitorID: ref.MonitorID.String(), Name: name})
	}

	rows, err := h.queries.ListStatusPageIncidentUpdates(ctx, incident.ID)
	if err != nil {
		return incidentResponse{}, err
	}
	updates := make([]incidentUpdateResponse, len(rows))
	for i, u := range rows {
		updates[i] = incidentUpdateResponse{
			ID:        u.ID.String(),
			Message:   u.Message,
			Status:    string(u.Status),
			CreatedAt: u.CreatedAt.Time.Format(time.RFC3339),
		}
	}

	return incidentResponse{
		ID:           incident.ID.String(),
		Title:        incident.Title,
		Severity:     string(incident.Severity),
		Status:       string(incident.Status),
		Monitors:     monitors,
		MonitorCount: len(monitors),
		Updates:      updates,
		CreatedAt:    incident.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    incident.UpdatedAt.Time.Format(time.RFC3339),
		ResolvedAt:   ptrTime(incident.ResolvedAt.Time, incident.ResolvedAt.Valid),
	}, nil
}

// ─── handlers ─────────────────────────────────────────────────────────────────

// ListIncidents GET /api/v1/incidents
func (h *IncidentHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}
	rows, err := h.queries.ListStatusPageIncidents(r.Context(), orgID)
	if err != nil {
		respond.InternalError(w)
		return
	}
	result := make([]incidentResponse, len(rows))
	for i, row := range rows {
		result[i] = toListItemResponse(row)
	}
	respond.JSON(w, http.StatusOK, result)
}

// decodeAndValidateCreateIncidentRequest decodes the request body and
// validates its required fields, writing an error response and returning
// ok=false on the first failure.
func decodeAndValidateCreateIncidentRequest(w http.ResponseWriter, r *http.Request) (req createIncidentRequest, title, message string, severity db.IncidentSeverity, ok bool) {
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return req, "", "", "", false
	}

	title = strings.TrimSpace(req.Title)
	if title == "" {
		respond.Error(w, http.StatusBadRequest, "title is required", "bad_request")
		return req, "", "", "", false
	}
	message = strings.TrimSpace(req.Message)
	if message == "" {
		respond.Error(w, http.StatusBadRequest, "message is required", "bad_request")
		return req, "", "", "", false
	}
	severity, ok = validSeverities[req.Severity]
	if !ok {
		respond.Error(w, http.StatusBadRequest, "severity must be minor, major, or critical", "bad_request")
		return req, "", "", "", false
	}
	if len(req.Monitors) == 0 {
		respond.Error(w, http.StatusBadRequest, "at least one monitor is required", "bad_request")
		return req, "", "", "", false
	}
	return req, title, message, severity, true
}

// resolveIncidentMonitors resolves and de-duplicates the incoming monitor
// refs, writing an error response and returning ok=false on the first
// unresolvable monitor.
func (h *IncidentHandler) resolveIncidentMonitors(ctx context.Context, w http.ResponseWriter, orgID uuid.UUID, inputs []incidentMonitorInput) ([]resolvedMonitor, bool) {
	seen := map[string]bool{}
	var monitors []resolvedMonitor
	for _, m := range inputs {
		id, name, err := resolveMonitorName(ctx, h.queries, orgID, m.MonitorType, m.MonitorID)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
			return nil, false
		}
		key := m.MonitorType + ":" + id.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		monitors = append(monitors, resolvedMonitor{monitorType: m.MonitorType, monitorID: id, name: name})
	}
	return monitors, true
}

// maxActiveIncidents caps how many non-resolved incidents an org can have
// open at once, uniform across every plan (docs/reference/limits.md) — a
// flat safety cap, not a plan-gated one, since resolving an incident (not
// upgrading) is the only way past it.
const maxActiveIncidents = 100

// checkActiveIncidentCap rejects declaring a new incident once the org
// already has maxActiveIncidents non-resolved ones. Complements the 90-day
// retention on resolved incidents (ADR-015): that bounds long-term growth,
// this bounds how many can pile up before any of them are ever resolved.
func (h *IncidentHandler) checkActiveIncidentCap(ctx context.Context, w http.ResponseWriter, orgID uuid.UUID) bool {
	count, err := h.queries.CountActiveStatusPageIncidents(ctx, orgID)
	if err != nil {
		respond.InternalError(w)
		return false
	}
	if count >= maxActiveIncidents {
		respond.Error(w, http.StatusConflict,
			"too many active incidents — resolve some before declaring more",
			"too_many_active_incidents")
		return false
	}
	return true
}

// checkMaintenanceOverlap implements US-2405: warn (don't block) if a
// selected monitor is already under active maintenance. Returns false —
// having written a 409 response — only when overlap exists and the caller
// hasn't already confirmed it.
func (h *IncidentHandler) checkMaintenanceOverlap(ctx context.Context, w http.ResponseWriter, orgID uuid.UUID, monitors []resolvedMonitor, confirmed bool) bool {
	if confirmed {
		return true
	}
	active, err := h.queries.GetActiveMaintenanceForOrg(ctx, orgID)
	if err != nil {
		return true
	}
	activeSet := map[string]bool{}
	for _, a := range active {
		activeSet[a.MonitorType+":"+a.MonitorID.String()] = true
	}
	var overlapping []string
	for _, m := range monitors {
		if activeSet[m.monitorType+":"+m.monitorID.String()] {
			overlapping = append(overlapping, m.name)
		}
	}
	if len(overlapping) == 0 {
		return true
	}
	respond.Error(w, http.StatusConflict,
		"already under active maintenance: "+strings.Join(overlapping, ", "),
		"maintenance_overlap")
	return false
}

// createIncidentWithMonitors persists the incident, its monitor links, and
// its first update, writing an error response and returning ok=false on the
// first failure.
func (h *IncidentHandler) createIncidentWithMonitors(ctx context.Context, w http.ResponseWriter, orgID uuid.UUID, title, message string, severity db.IncidentSeverity, monitors []resolvedMonitor) (db.StatusPageIncident, bool) {
	incident, err := h.queries.CreateStatusPageIncident(ctx, db.CreateStatusPageIncidentParams{
		OrgID: orgID, Title: title, Severity: severity,
	})
	if err != nil {
		respond.InternalError(w)
		return db.StatusPageIncident{}, false
	}

	for _, m := range monitors {
		if _, err := h.queries.InsertStatusPageIncidentMonitor(ctx, db.InsertStatusPageIncidentMonitorParams{
			IncidentID: incident.ID, MonitorType: m.monitorType, MonitorID: m.monitorID,
		}); err != nil {
			respond.InternalError(w)
			return db.StatusPageIncident{}, false
		}
	}

	if _, err := h.queries.InsertStatusPageIncidentUpdate(ctx, db.InsertStatusPageIncidentUpdateParams{
		IncidentID: incident.ID, Message: message, Status: db.IncidentStatusInvestigating,
	}); err != nil {
		respond.InternalError(w)
		return db.StatusPageIncident{}, false
	}

	return incident, true
}

// CreateIncident POST /api/v1/incidents
func (h *IncidentHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	req, title, message, severity, ok := decodeAndValidateCreateIncidentRequest(w, r)
	if !ok {
		return
	}

	ctx := r.Context()
	if !h.checkActiveIncidentCap(ctx, w, orgID) {
		return
	}
	monitors, ok := h.resolveIncidentMonitors(ctx, w, orgID, req.Monitors)
	if !ok {
		return
	}
	if !h.checkMaintenanceOverlap(ctx, w, orgID, monitors, req.ConfirmOverlap) {
		return
	}

	incident, ok := h.createIncidentWithMonitors(ctx, w, orgID, title, message, severity, monitors)
	if !ok {
		return
	}

	resp, err := h.buildIncidentDetail(ctx, orgID, incident)
	if err != nil {
		respond.InternalError(w)
		return
	}
	respond.JSON(w, http.StatusCreated, resp)
}

// GetIncident GET /api/v1/incidents/:id
func (h *IncidentHandler) GetIncident(w http.ResponseWriter, r *http.Request) {
	orgID, id, ok := incidentIDs(w, r)
	if !ok {
		return
	}
	incident, err := h.queries.GetStatusPageIncident(r.Context(), db.GetStatusPageIncidentParams{ID: id, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "incident not found", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}
	resp, err := h.buildIncidentDetail(r.Context(), orgID, incident)
	if err != nil {
		respond.InternalError(w)
		return
	}
	respond.JSON(w, http.StatusOK, resp)
}

// UpdateIncidentTitle PATCH /api/v1/incidents/:id
func (h *IncidentHandler) UpdateIncidentTitle(w http.ResponseWriter, r *http.Request) {
	orgID, id, ok := incidentIDs(w, r)
	if !ok {
		return
	}
	var req updateTitleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		respond.Error(w, http.StatusBadRequest, "title is required", "bad_request")
		return
	}

	incident, err := h.queries.UpdateStatusPageIncidentTitle(r.Context(), db.UpdateStatusPageIncidentTitleParams{
		ID: id, OrgID: orgID, Title: title,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "incident not found", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}
	resp, err := h.buildIncidentDetail(r.Context(), orgID, incident)
	if err != nil {
		respond.InternalError(w)
		return
	}
	respond.JSON(w, http.StatusOK, resp)
}

// DeleteIncident DELETE /api/v1/incidents/:id
func (h *IncidentHandler) DeleteIncident(w http.ResponseWriter, r *http.Request) {
	orgID, id, ok := incidentIDs(w, r)
	if !ok {
		return
	}
	if err := h.queries.DeleteStatusPageIncident(r.Context(), db.DeleteStatusPageIncidentParams{ID: id, OrgID: orgID}); err != nil {
		respond.InternalError(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PostIncidentUpdate POST /api/v1/incidents/:id/updates
func (h *IncidentHandler) PostIncidentUpdate(w http.ResponseWriter, r *http.Request) {
	orgID, id, ok := incidentIDs(w, r)
	if !ok {
		return
	}

	incident, err := h.queries.GetStatusPageIncident(r.Context(), db.GetStatusPageIncidentParams{ID: id, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "incident not found", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}

	var req postUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		respond.Error(w, http.StatusBadRequest, "message is required", "bad_request")
		return
	}
	status, ok := validStatuses[req.Status]
	if !ok {
		respond.Error(w, http.StatusBadRequest, "status must be investigating, identified, monitoring, or resolved", "bad_request")
		return
	}

	ctx := r.Context()
	if _, err := h.queries.InsertStatusPageIncidentUpdate(ctx, db.InsertStatusPageIncidentUpdateParams{
		IncidentID: incident.ID, Message: message, Status: status,
	}); err != nil {
		respond.InternalError(w)
		return
	}

	updated, err := h.queries.UpdateStatusPageIncidentStatus(ctx, db.UpdateStatusPageIncidentStatusParams{
		ID: id, OrgID: orgID, Status: status,
	})
	if err != nil {
		respond.InternalError(w)
		return
	}

	resp, err := h.buildIncidentDetail(ctx, orgID, updated)
	if err != nil {
		respond.InternalError(w)
		return
	}
	respond.JSON(w, http.StatusCreated, resp)
}

// UpdateIncidentUpdateMessage PATCH /api/v1/incidents/:id/updates/:updateId
func (h *IncidentHandler) UpdateIncidentUpdateMessage(w http.ResponseWriter, r *http.Request) {
	orgID, id, ok := incidentIDs(w, r)
	if !ok {
		return
	}
	updateID, err := uuid.Parse(chi.URLParam(r, "updateId"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid update id", "bad_request")
		return
	}

	incident, err := h.queries.GetStatusPageIncident(r.Context(), db.GetStatusPageIncidentParams{ID: id, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "incident not found", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}

	var req updateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		respond.Error(w, http.StatusBadRequest, "message is required", "bad_request")
		return
	}

	ctx := r.Context()
	if _, err := h.queries.UpdateStatusPageIncidentUpdateMessage(ctx, db.UpdateStatusPageIncidentUpdateMessageParams{
		ID: updateID, IncidentID: incident.ID, Message: message,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "update not found", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}

	resp, err := h.buildIncidentDetail(ctx, orgID, incident)
	if err != nil {
		respond.InternalError(w)
		return
	}
	respond.JSON(w, http.StatusOK, resp)
}
