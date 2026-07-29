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

// monitorNameLookups fetches a single monitor's name by ID+org, one closure
// per monitor type — each Get*MonitorParams is its own generated sqlc type,
// so this is the shared shape without generics. Used by resolveMonitorName.
var monitorNameLookups = map[string]func(ctx context.Context, q *db.Queries, id, orgID uuid.UUID) (string, error){
	"cron": func(ctx context.Context, q *db.Queries, id, orgID uuid.UUID) (string, error) {
		m, err := q.GetCronMonitor(ctx, db.GetCronMonitorParams{ID: id, OrgID: orgID})
		return m.Name, err
	},
	"uptime": func(ctx context.Context, q *db.Queries, id, orgID uuid.UUID) (string, error) {
		m, err := q.GetUptimeMonitor(ctx, db.GetUptimeMonitorParams{ID: id, OrgID: orgID})
		return m.Name, err
	},
	"ssl": func(ctx context.Context, q *db.Queries, id, orgID uuid.UUID) (string, error) {
		m, err := q.GetSSLMonitor(ctx, db.GetSSLMonitorParams{ID: id, OrgID: orgID})
		return m.Name, err
	},
	"domain": func(ctx context.Context, q *db.Queries, id, orgID uuid.UUID) (string, error) {
		m, err := q.GetDomainMonitor(ctx, db.GetDomainMonitorParams{ID: id, OrgID: orgID})
		return m.Name, err
	},
	"port": func(ctx context.Context, q *db.Queries, id, orgID uuid.UUID) (string, error) {
		m, err := q.GetPortMonitor(ctx, db.GetPortMonitorParams{ID: id, OrgID: orgID})
		return m.Name, err
	},
	"dns": func(ctx context.Context, q *db.Queries, id, orgID uuid.UUID) (string, error) {
		m, err := q.GetDNSMonitor(ctx, db.GetDNSMonitorParams{ID: id, OrgID: orgID})
		return m.Name, err
	},
}

// resolveMonitorName verifies a monitor belongs to the org and returns its display name.
func resolveMonitorName(ctx context.Context, queries *db.Queries, orgID uuid.UUID, monitorType, monitorIDStr string) (uuid.UUID, string, error) {
	id, err := uuid.Parse(monitorIDStr)
	if err != nil {
		return uuid.UUID{}, "", fmt.Errorf("invalid monitor id: %s", monitorIDStr)
	}
	lookup, ok := monitorNameLookups[monitorType]
	if !ok {
		return uuid.UUID{}, "", fmt.Errorf("invalid monitor type: %s", monitorType)
	}
	name, err := lookup(ctx, queries, id, orgID)
	if err != nil {
		return uuid.UUID{}, "", fmt.Errorf("%s monitor not found: %s", monitorType, monitorIDStr)
	}
	return id, name, nil
}

// parseWindowEndsAt parses the optional endsAt field, validating it's after
// start when present. Returns a zero-value (unset) Timestamptz and empty
// errMsg when endsAt is omitted — an open-ended window is valid.
func parseWindowEndsAt(rawEndsAt *string, start time.Time) (endsAt pgtype.Timestamptz, errMsg string) {
	if rawEndsAt == nil || strings.TrimSpace(*rawEndsAt) == "" {
		return pgtype.Timestamptz{}, ""
	}
	end, err := time.Parse(time.RFC3339, *rawEndsAt)
	if err != nil {
		return pgtype.Timestamptz{}, "endsAt must be an RFC3339 timestamp"
	}
	if !end.After(start) {
		return pgtype.Timestamptz{}, "endsAt must be after startsAt"
	}
	return pgtype.Timestamptz{Time: end, Valid: true}, ""
}

// resolveWindowMonitors resolves and de-duplicates reqMonitors against the
// org. errMsg is non-empty (with monitors nil) on the first resolution
// failure.
func (h *MaintenanceHandler) resolveWindowMonitors(ctx context.Context, orgID uuid.UUID, reqMonitors []maintenanceMonitorInput) (monitors []resolvedMonitor, errMsg string) {
	seen := map[string]bool{}
	for _, m := range reqMonitors {
		id, name, rerr := resolveMonitorName(ctx, h.queries, orgID, m.MonitorType, m.MonitorID)
		if rerr != nil {
			return nil, rerr.Error()
		}
		key := m.MonitorType + ":" + id.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		monitors = append(monitors, resolvedMonitor{monitorType: m.MonitorType, monitorID: id, name: name})
	}
	return monitors, ""
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

	endsAt, errMsg = parseWindowEndsAt(req.EndsAt, start)
	if errMsg != "" {
		return
	}

	monitors, errMsg = h.resolveWindowMonitors(ctx, orgID, req.Monitors)
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

// maxMaintenanceWindows caps how many maintenance windows an org can create
// in total — a flat safety cap, uniform across every plan. Unlike
// incidents, maintenance windows have no retention/pruning of old ones, so
// this bounds cumulative creation, not a concurrently-active count.
const maxMaintenanceWindows = 100

// respondMaintenanceWindowNotFoundOrInternal writes a 404 if err is
// pgx.ErrNoRows, a 500 for any other non-nil error, and reports whether it
// wrote a response so the caller knows to return immediately.
func respondMaintenanceWindowNotFoundOrInternal(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, pgx.ErrNoRows) {
		respond.Error(w, http.StatusNotFound, "maintenance window not found", "not_found")
		return true
	}
	respond.InternalError(w)
	return true
}

// checkMaintenanceWindowCreateLimit enforces maxMaintenanceWindows (a flat
// cumulative cap, uniform across every plan — see maxMaintenanceWindows).
// Responds and returns false on any failure (query error or limit hit).
func (h *MaintenanceHandler) checkMaintenanceWindowCreateLimit(w http.ResponseWriter, r *http.Request, orgID uuid.UUID) bool {
	count, err := h.queries.CountMaintenanceWindows(r.Context(), orgID)
	if err != nil {
		respond.InternalError(w)
		return false
	}
	if count >= maxMaintenanceWindows {
		respond.Error(w, http.StatusConflict,
			"too many maintenance windows — delete an old one before creating more",
			"too_many_maintenance_windows")
		return false
	}
	return true
}

// insertWindowMonitors attaches monitors to windowID.
func (h *MaintenanceHandler) insertWindowMonitors(ctx context.Context, windowID uuid.UUID, monitors []resolvedMonitor) error {
	for _, m := range monitors {
		if _, err := h.queries.InsertMaintenanceWindowMonitor(ctx, db.InsertMaintenanceWindowMonitorParams{
			WindowID: windowID, MonitorType: m.monitorType, MonitorID: m.monitorID,
		}); err != nil {
			return err
		}
	}
	return nil
}

// replaceWindowMonitors swaps out windowID's monitor attachments for
// monitors — used by UpdateMaintenanceWindow, where the new list may drop or
// add monitors compared to what's already attached.
func (h *MaintenanceHandler) replaceWindowMonitors(ctx context.Context, windowID uuid.UUID, monitors []resolvedMonitor) error {
	if err := h.queries.DeleteMaintenanceWindowMonitors(ctx, windowID); err != nil {
		return err
	}
	return h.insertWindowMonitors(ctx, windowID, monitors)
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

	if !h.checkMaintenanceWindowCreateLimit(w, r, orgID) {
		return
	}

	win, err := h.queries.CreateMaintenanceWindow(r.Context(), db.CreateMaintenanceWindowParams{
		OrgID: orgID, Title: title, Message: message, StartsAt: startsAt, EndsAt: endsAt,
	})
	if err != nil {
		respond.InternalError(w)
		return
	}

	if err := h.insertWindowMonitors(r.Context(), win.ID, monitors); err != nil {
		respond.InternalError(w)
		return
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
	if respondMaintenanceWindowNotFoundOrInternal(w, err) {
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

	_, err := h.queries.GetMaintenanceWindow(r.Context(), db.GetMaintenanceWindowParams{ID: id, OrgID: orgID})
	if respondMaintenanceWindowNotFoundOrInternal(w, err) {
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

	if err := h.replaceWindowMonitors(r.Context(), win.ID, monitors); err != nil {
		respond.InternalError(w)
		return
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
