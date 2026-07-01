package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

type MaintenanceHandler struct {
	queries *db.Queries
}

func NewMaintenanceHandler(pool *pgxpool.Pool) *MaintenanceHandler {
	return &MaintenanceHandler{queries: db.New(pool)}
}

// ─── request/response types ──────────────────────────────────────────────────

type maintenanceMonitorInput struct {
	MonitorType string `json:"monitorType"`
	MonitorID   string `json:"monitorId"`
}

type maintenanceWindowRequest struct {
	Title    string                    `json:"title"`
	Message  string                    `json:"message"`
	StartsAt string                    `json:"startsAt"`
	EndsAt   *string                   `json:"endsAt"`
	Monitors []maintenanceMonitorInput `json:"monitors"`
}

type maintenanceMonitorRef struct {
	MonitorType string `json:"monitorType"`
	MonitorID   string `json:"monitorId"`
	Name        string `json:"name"`
}

type maintenanceWindowResponse struct {
	ID           string                  `json:"id"`
	Title        string                  `json:"title"`
	Message      string                  `json:"message"`
	StartsAt     string                  `json:"startsAt"`
	EndsAt       *string                 `json:"endsAt"`
	Status       string                  `json:"status"` // upcoming | active | ended
	Monitors     []maintenanceMonitorRef `json:"monitors,omitempty"`
	MonitorCount int                     `json:"monitorCount"`
	CreatedAt    string                  `json:"createdAt"`
}

type resolvedMonitor struct {
	monitorType string
	monitorID   uuid.UUID
	name        string
}

func windowStatus(startsAt time.Time, endsAt pgtype.Timestamptz) string {
	now := time.Now()
	if now.Before(startsAt) {
		return "upcoming"
	}
	if endsAt.Valid && !now.Before(endsAt.Time) {
		return "ended"
	}
	return "active"
}

func toMaintenanceWindowResponse(win db.MaintenanceWindow, monitors []resolvedMonitor) maintenanceWindowResponse {
	resp := maintenanceWindowResponse{
		ID:           win.ID.String(),
		Title:        win.Title,
		Message:      win.Message,
		StartsAt:     win.StartsAt.Time.Format(time.RFC3339),
		Status:       windowStatus(win.StartsAt.Time, win.EndsAt),
		MonitorCount: len(monitors),
		CreatedAt:    win.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if win.EndsAt.Valid {
		t := win.EndsAt.Time.Format(time.RFC3339)
		resp.EndsAt = &t
	}
	for _, m := range monitors {
		resp.Monitors = append(resp.Monitors, maintenanceMonitorRef{
			MonitorType: m.monitorType,
			MonitorID:   m.monitorID.String(),
			Name:        m.name,
		})
	}
	return resp
}

func maintenanceWindowIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
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

// resolveMonitorName verifies a monitor belongs to the org and returns its display name.
func resolveMonitorName(ctx context.Context, queries *db.Queries, orgID uuid.UUID, monitorType, monitorIDStr string) (uuid.UUID, string, error) {
	id, err := uuid.Parse(monitorIDStr)
	if err != nil {
		return uuid.UUID{}, "", fmt.Errorf("invalid monitor id: %s", monitorIDStr)
	}
	switch monitorType {
	case "cron":
		m, err := queries.GetCronMonitor(ctx, db.GetCronMonitorParams{ID: id, OrgID: orgID})
		if err != nil {
			return uuid.UUID{}, "", fmt.Errorf("cron monitor not found: %s", monitorIDStr)
		}
		return id, m.Name, nil
	case "uptime":
		m, err := queries.GetUptimeMonitor(ctx, db.GetUptimeMonitorParams{ID: id, OrgID: orgID})
		if err != nil {
			return uuid.UUID{}, "", fmt.Errorf("uptime monitor not found: %s", monitorIDStr)
		}
		return id, m.Name, nil
	case "ssl":
		m, err := queries.GetSSLMonitor(ctx, db.GetSSLMonitorParams{ID: id, OrgID: orgID})
		if err != nil {
			return uuid.UUID{}, "", fmt.Errorf("ssl monitor not found: %s", monitorIDStr)
		}
		return id, m.Name, nil
	case "domain":
		m, err := queries.GetDomainMonitor(ctx, db.GetDomainMonitorParams{ID: id, OrgID: orgID})
		if err != nil {
			return uuid.UUID{}, "", fmt.Errorf("domain monitor not found: %s", monitorIDStr)
		}
		return id, m.Name, nil
	case "port":
		m, err := queries.GetPortMonitor(ctx, db.GetPortMonitorParams{ID: id, OrgID: orgID})
		if err != nil {
			return uuid.UUID{}, "", fmt.Errorf("port monitor not found: %s", monitorIDStr)
		}
		return id, m.Name, nil
	default:
		return uuid.UUID{}, "", fmt.Errorf("invalid monitor type: %s", monitorType)
	}
}

// validateWindowInput parses and validates a create/update request, resolving and
// de-duplicating the monitor list against the org. errMsg is non-empty on failure.
func (h *MaintenanceHandler) validateWindowInput(ctx context.Context, orgID uuid.UUID, req maintenanceWindowRequest) (title, message string, startsAt, endsAt pgtype.Timestamptz, monitors []resolvedMonitor, errMsg string) {
	title = strings.TrimSpace(req.Title)
	if title == "" {
		errMsg = "title is required"
		return
	}
	message = strings.TrimSpace(req.Message)

	if len(req.Monitors) == 0 {
		errMsg = "at least one monitor is required"
		return
	}

	start, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		errMsg = "startsAt must be an RFC3339 timestamp"
		return
	}
	startsAt = pgtype.Timestamptz{Time: start, Valid: true}

	if req.EndsAt != nil && strings.TrimSpace(*req.EndsAt) != "" {
		end, err := time.Parse(time.RFC3339, *req.EndsAt)
		if err != nil {
			errMsg = "endsAt must be an RFC3339 timestamp"
			return
		}
		if !end.After(start) {
			errMsg = "endsAt must be after startsAt"
			return
		}
		endsAt = pgtype.Timestamptz{Time: end, Valid: true}
	}

	seen := map[string]bool{}
	for _, m := range req.Monitors {
		id, name, rerr := resolveMonitorName(ctx, h.queries, orgID, m.MonitorType, m.MonitorID)
		if rerr != nil {
			errMsg = rerr.Error()
			return
		}
		key := m.MonitorType + ":" + id.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		monitors = append(monitors, resolvedMonitor{monitorType: m.MonitorType, monitorID: id, name: name})
	}
	return
}

// ─── handlers ─────────────────────────────────────────────────────────────────

// ListMaintenanceWindows GET /api/v1/maintenance-windows
func (h *MaintenanceHandler) ListMaintenanceWindows(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}
	rows, err := h.queries.ListMaintenanceWindows(r.Context(), orgID)
	if err != nil {
		respond.InternalError(w)
		return
	}
	result := make([]maintenanceWindowResponse, len(rows))
	for i, row := range rows {
		resp := maintenanceWindowResponse{
			ID:           row.ID.String(),
			Title:        row.Title,
			Message:      row.Message,
			StartsAt:     row.StartsAt.Time.Format(time.RFC3339),
			Status:       windowStatus(row.StartsAt.Time, row.EndsAt),
			MonitorCount: int(row.MonitorCount),
			CreatedAt:    row.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if row.EndsAt.Valid {
			t := row.EndsAt.Time.Format(time.RFC3339)
			resp.EndsAt = &t
		}
		result[i] = resp
	}
	respond.JSON(w, http.StatusOK, result)
}

// CreateMaintenanceWindow POST /api/v1/maintenance-windows
func (h *MaintenanceHandler) CreateMaintenanceWindow(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req maintenanceWindowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	title, message, startsAt, endsAt, monitors, errMsg := h.validateWindowInput(r.Context(), orgID, req)
	if errMsg != "" {
		respond.Error(w, http.StatusBadRequest, errMsg, "bad_request")
		return
	}

	win, err := h.queries.CreateMaintenanceWindow(r.Context(), db.CreateMaintenanceWindowParams{
		OrgID: orgID, Title: title, Message: message, StartsAt: startsAt, EndsAt: endsAt,
	})
	if err != nil {
		respond.InternalError(w)
		return
	}

	for _, m := range monitors {
		if _, err := h.queries.InsertMaintenanceWindowMonitor(r.Context(), db.InsertMaintenanceWindowMonitorParams{
			WindowID: win.ID, MonitorType: m.monitorType, MonitorID: m.monitorID,
		}); err != nil {
			respond.InternalError(w)
			return
		}
	}

	respond.JSON(w, http.StatusCreated, toMaintenanceWindowResponse(win, monitors))
}

// GetMaintenanceWindow GET /api/v1/maintenance-windows/:id
func (h *MaintenanceHandler) GetMaintenanceWindow(w http.ResponseWriter, r *http.Request) {
	orgID, id, ok := maintenanceWindowIDs(w, r)
	if !ok {
		return
	}

	win, err := h.queries.GetMaintenanceWindow(r.Context(), db.GetMaintenanceWindowParams{ID: id, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "maintenance window not found", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}

	refs, err := h.queries.ListMaintenanceWindowMonitors(r.Context(), win.ID)
	if err != nil {
		respond.InternalError(w)
		return
	}

	monitors := make([]resolvedMonitor, 0, len(refs))
	for _, ref := range refs {
		name := "(deleted monitor)"
		if _, n, rerr := resolveMonitorName(r.Context(), h.queries, orgID, ref.MonitorType, ref.MonitorID.String()); rerr == nil {
			name = n
		}
		monitors = append(monitors, resolvedMonitor{monitorType: ref.MonitorType, monitorID: ref.MonitorID, name: name})
	}

	respond.JSON(w, http.StatusOK, toMaintenanceWindowResponse(win, monitors))
}

// UpdateMaintenanceWindow PATCH /api/v1/maintenance-windows/:id
func (h *MaintenanceHandler) UpdateMaintenanceWindow(w http.ResponseWriter, r *http.Request) {
	orgID, id, ok := maintenanceWindowIDs(w, r)
	if !ok {
		return
	}

	if _, err := h.queries.GetMaintenanceWindow(r.Context(), db.GetMaintenanceWindowParams{ID: id, OrgID: orgID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "maintenance window not found", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}

	var req maintenanceWindowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	title, message, startsAt, endsAt, monitors, errMsg := h.validateWindowInput(r.Context(), orgID, req)
	if errMsg != "" {
		respond.Error(w, http.StatusBadRequest, errMsg, "bad_request")
		return
	}

	win, err := h.queries.UpdateMaintenanceWindow(r.Context(), db.UpdateMaintenanceWindowParams{
		ID: id, OrgID: orgID, Title: title, Message: message, StartsAt: startsAt, EndsAt: endsAt,
	})
	if err != nil {
		respond.InternalError(w)
		return
	}

	if err := h.queries.DeleteMaintenanceWindowMonitors(r.Context(), win.ID); err != nil {
		respond.InternalError(w)
		return
	}
	for _, m := range monitors {
		if _, err := h.queries.InsertMaintenanceWindowMonitor(r.Context(), db.InsertMaintenanceWindowMonitorParams{
			WindowID: win.ID, MonitorType: m.monitorType, MonitorID: m.monitorID,
		}); err != nil {
			respond.InternalError(w)
			return
		}
	}

	respond.JSON(w, http.StatusOK, toMaintenanceWindowResponse(win, monitors))
}

// DeleteMaintenanceWindow DELETE /api/v1/maintenance-windows/:id
func (h *MaintenanceHandler) DeleteMaintenanceWindow(w http.ResponseWriter, r *http.Request) {
	orgID, id, ok := maintenanceWindowIDs(w, r)
	if !ok {
		return
	}
	if err := h.queries.DeleteMaintenanceWindow(r.Context(), db.DeleteMaintenanceWindowParams{ID: id, OrgID: orgID}); err != nil {
		respond.InternalError(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// EndMaintenanceWindowNow POST /api/v1/maintenance-windows/:id/end
func (h *MaintenanceHandler) EndMaintenanceWindowNow(w http.ResponseWriter, r *http.Request) {
	orgID, id, ok := maintenanceWindowIDs(w, r)
	if !ok {
		return
	}

	win, err := h.queries.EndMaintenanceWindowNow(r.Context(), db.EndMaintenanceWindowNowParams{ID: id, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "maintenance window not found or already ended", "not_found")
			return
		}
		respond.InternalError(w)
		return
	}

	refs, err := h.queries.ListMaintenanceWindowMonitors(r.Context(), win.ID)
	if err != nil {
		respond.InternalError(w)
		return
	}

	resp := toMaintenanceWindowResponse(win, nil)
	resp.MonitorCount = len(refs)
	respond.JSON(w, http.StatusOK, resp)
}
