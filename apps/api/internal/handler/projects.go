package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

// ProjectHandler serves an org's projects (ADR projects): named groups of
// monitors, typically one per client.
type ProjectHandler struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewProjectHandler(pool *pgxpool.Pool) *ProjectHandler {
	return &ProjectHandler{pool: pool, queries: db.New(pool)}
}

// maxProjects caps how many projects an org can have — a flat safety cap,
// the same on every plan (ADR limits). Projects aren't a plan feature.
const maxProjects = 1000

// maxProjectNameLen bounds a project name in characters; it is shown in the
// monitor list and at the start of every alert for the project's monitors.
const maxProjectNameLen = 100

// maxProjectMonitors bounds one assignment request: the most monitors any
// plan allows (billing.planLimits' Enterprise MonitorTotal).
const maxProjectMonitors = 1000

type projectRequest struct {
	Name string `json:"name"`
}

type projectResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	MonitorCount int    `json:"monitorCount"`
	CreatedAt    string `json:"createdAt"`
}

type projectMonitorInput struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type setProjectMonitorsRequest struct {
	Monitors []projectMonitorInput `json:"monitors"`
}

func toProjectResponse(p db.Project, monitorCount int) projectResponse {
	return projectResponse{
		ID:           p.ID.String(),
		Name:         p.Name,
		MonitorCount: monitorCount,
		CreatedAt:    p.CreatedAt.Time.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// validateProjectName trims name and returns an error message when it is
// empty or too long, "" when it's fine.
func validateProjectName(name *string) string {
	*name = strings.TrimSpace(*name)
	if *name == "" {
		return "name is required"
	}
	if utf8.RuneCountInString(*name) > maxProjectNameLen {
		return "name must be at most 100 characters"
	}
	return ""
}

func isProjectNameTaken(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "idx_projects_org_name"
}

func respondProjectNameTaken(w http.ResponseWriter) {
	respond.Error(w, http.StatusConflict, "a project with this name already exists", "project_name_taken")
}

// respondProjectNotFoundOrInternal writes a 404 if err is pgx.ErrNoRows, a
// 500 for any other non-nil error, and reports whether it wrote a response.
func respondProjectNotFoundOrInternal(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, pgx.ErrNoRows) {
		respond.Error(w, http.StatusNotFound, "project not found", "not_found")
		return true
	}
	respond.InternalError(w)
	return true
}

func projectIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return uuid.UUID{}, uuid.UUID{}, false
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid project id", "bad_request")
		return uuid.UUID{}, uuid.UUID{}, false
	}
	return orgID, id, true
}

// ListProjects GET /api/v1/projects
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}
	rows, err := h.queries.ListProjects(r.Context(), orgID)
	if err != nil {
		respond.InternalError(w)
		return
	}
	result := make([]projectResponse, len(rows))
	for i, row := range rows {
		result[i] = toProjectResponse(db.Project{
			ID: row.ID, OrgID: row.OrgID, Name: row.Name, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}, int(row.MonitorCount))
	}
	respond.JSON(w, http.StatusOK, result)
}

// CreateProject POST /api/v1/projects
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}

	var req projectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	if msg := validateProjectName(&req.Name); msg != "" {
		respond.Error(w, http.StatusBadRequest, msg, "bad_request")
		return
	}

	count, err := h.queries.CountProjects(r.Context(), orgID)
	if err != nil {
		respond.InternalError(w)
		return
	}
	if count >= maxProjects {
		respond.Error(w, http.StatusConflict,
			"too many projects — delete one before creating more",
			"too_many_projects")
		return
	}

	p, err := h.queries.CreateProject(r.Context(), db.CreateProjectParams{OrgID: orgID, Name: req.Name})
	if isProjectNameTaken(err) {
		respondProjectNameTaken(w)
		return
	}
	if err != nil {
		respond.InternalError(w)
		return
	}
	respond.JSON(w, http.StatusCreated, toProjectResponse(p, 0))
}

// RenameProject PATCH /api/v1/projects/{id}
func (h *ProjectHandler) RenameProject(w http.ResponseWriter, r *http.Request) {
	orgID, id, ok := projectIDs(w, r)
	if !ok {
		return
	}

	var req projectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	if msg := validateProjectName(&req.Name); msg != "" {
		respond.Error(w, http.StatusBadRequest, msg, "bad_request")
		return
	}

	p, err := h.queries.RenameProject(r.Context(), db.RenameProjectParams{ID: id, OrgID: orgID, Name: req.Name})
	if isProjectNameTaken(err) {
		respondProjectNameTaken(w)
		return
	}
	if respondProjectNotFoundOrInternal(w, err) {
		return
	}
	respond.JSON(w, http.StatusOK, toProjectResponse(p, h.countProjectMonitors(r.Context(), orgID, p.ID)))
}

// countProjectMonitors returns how many monitors are in a project, 0 if it
// can't be read — only the response's count is at stake.
func (h *ProjectHandler) countProjectMonitors(ctx context.Context, orgID, projectID uuid.UUID) int {
	n, err := h.queries.CountProjectMonitors(ctx, db.CountProjectMonitorsParams{OrgID: orgID, ProjectID: projectID})
	if err != nil {
		return 0
	}
	return int(n)
}

// DeleteProject DELETE /api/v1/projects/{id}
// The project's monitors stay, without a project (ADR projects).
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	orgID, id, ok := projectIDs(w, r)
	if !ok {
		return
	}
	n, err := h.queries.DeleteProject(r.Context(), db.DeleteProjectParams{ID: id, OrgID: orgID})
	if err != nil {
		respond.InternalError(w)
		return
	}
	if n == 0 {
		respond.Error(w, http.StatusNotFound, "project not found", "not_found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// projectAssigners puts exactly the given monitors of one type into a
// project, one closure per monitor type since each Assign*Params is its own
// generated sqlc type — same shape as monitorNameLookups.
var projectAssigners = map[string]func(ctx context.Context, q *db.Queries, orgID, projectID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error){
	"cron": func(ctx context.Context, q *db.Queries, orgID, projectID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
		return q.AssignCronMonitorsToProject(ctx, db.AssignCronMonitorsToProjectParams{OrgID: orgID, ProjectID: projectID, MonitorIds: ids})
	},
	"uptime": func(ctx context.Context, q *db.Queries, orgID, projectID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
		return q.AssignUptimeMonitorsToProject(ctx, db.AssignUptimeMonitorsToProjectParams{OrgID: orgID, ProjectID: projectID, MonitorIds: ids})
	},
	"ssl": func(ctx context.Context, q *db.Queries, orgID, projectID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
		return q.AssignSSLMonitorsToProject(ctx, db.AssignSSLMonitorsToProjectParams{OrgID: orgID, ProjectID: projectID, MonitorIds: ids})
	},
	"domain": func(ctx context.Context, q *db.Queries, orgID, projectID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
		return q.AssignDomainMonitorsToProject(ctx, db.AssignDomainMonitorsToProjectParams{OrgID: orgID, ProjectID: projectID, MonitorIds: ids})
	},
	"port": func(ctx context.Context, q *db.Queries, orgID, projectID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
		return q.AssignPortMonitorsToProject(ctx, db.AssignPortMonitorsToProjectParams{OrgID: orgID, ProjectID: projectID, MonitorIds: ids})
	},
	"dns": func(ctx context.Context, q *db.Queries, orgID, projectID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
		return q.AssignDNSMonitorsToProject(ctx, db.AssignDNSMonitorsToProjectParams{OrgID: orgID, ProjectID: projectID, MonitorIds: ids})
	},
}

// groupProjectMonitors validates and de-duplicates a request's monitors,
// grouped by type with every type present (an empty list takes all of that
// type out of the project). errMsg is non-empty on the first invalid entry.
func groupProjectMonitors(in []projectMonitorInput) (byType map[string][]uuid.UUID, total int, errMsg string) {
	if len(in) > maxProjectMonitors {
		return nil, 0, "at most 1000 monitors per request"
	}
	byType = make(map[string][]uuid.UUID, len(projectAssigners))
	for t := range projectAssigners {
		byType[t] = []uuid.UUID{}
	}
	seen := map[string]bool{}
	for _, m := range in {
		if _, ok := projectAssigners[m.Type]; !ok {
			return nil, 0, "invalid monitor type: " + m.Type
		}
		id, err := uuid.Parse(m.ID)
		if err != nil {
			return nil, 0, "invalid monitor id: " + m.ID
		}
		key := m.Type + ":" + id.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		byType[m.Type] = append(byType[m.Type], id)
	}
	return byType, len(seen), ""
}

// errProjectMonitorNotFound means a listed monitor isn't one of the org's.
var errProjectMonitorNotFound = errors.New("monitor not found")

// assignProjectMonitors runs every type's assignment in one transaction, so
// a request with an unknown monitor changes nothing.
func (h *ProjectHandler) assignProjectMonitors(ctx context.Context, orgID, projectID uuid.UUID, byType map[string][]uuid.UUID) error {
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := h.queries.WithTx(tx)

	for t, ids := range byType {
		touched, err := projectAssigners[t](ctx, qtx, orgID, projectID, ids)
		if err != nil {
			return err
		}
		found := make(map[uuid.UUID]bool, len(touched))
		for _, id := range touched {
			found[id] = true
		}
		for _, id := range ids {
			if !found[id] {
				return errProjectMonitorNotFound
			}
		}
	}
	return tx.Commit(ctx)
}

// SetProjectMonitors PUT /api/v1/projects/{id}/monitors
// Makes the listed monitors, of any types, exactly the project's monitors:
// listed ones move into it (out of any other project), the rest leave it.
func (h *ProjectHandler) SetProjectMonitors(w http.ResponseWriter, r *http.Request) {
	orgID, id, ok := projectIDs(w, r)
	if !ok {
		return
	}

	var req setProjectMonitorsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	byType, total, errMsg := groupProjectMonitors(req.Monitors)
	if errMsg != "" {
		respond.Error(w, http.StatusBadRequest, errMsg, "bad_request")
		return
	}

	p, err := h.queries.GetProject(r.Context(), db.GetProjectParams{ID: id, OrgID: orgID})
	if respondProjectNotFoundOrInternal(w, err) {
		return
	}

	if err := h.assignProjectMonitors(r.Context(), orgID, p.ID, byType); err != nil {
		if errors.Is(err, errProjectMonitorNotFound) {
			respond.Error(w, http.StatusBadRequest, "monitor not found", "bad_request")
			return
		}
		respond.InternalError(w)
		return
	}
	respond.JSON(w, http.StatusOK, toProjectResponse(p, total))
}

// ─── a monitor's project, on the monitor create/update requests ──────────────

// optionalProjectID is a monitor create/update request's projectId: absent
// leaves the monitor's project as it is, null takes the monitor out of its
// project, and an ID puts it into that project.
type optionalProjectID struct {
	Set   bool
	Value *string
}

// MarshalJSON lets tests send a request struct; with the field's omitzero an
// unset value leaves projectId out of the body altogether.
func (o optionalProjectID) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Value)
}

func (o *optionalProjectID) UnmarshalJSON(b []byte) error {
	o.Set = true
	if string(b) == "null" {
		o.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	o.Value = &s
	return nil
}

// resolveProjectID checks a request's projectId against the org before the
// monitor is written, responding 400 for a malformed or unknown project and
// returning false. The result is NULL when the request takes the monitor out
// of its project.
func (h *MonitorHandler) resolveProjectID(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, p optionalProjectID) (pgtype.UUID, bool) {
	if !p.Set || p.Value == nil || *p.Value == "" {
		return pgtype.UUID{}, true
	}
	id, err := uuid.Parse(*p.Value)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid projectId", "bad_request")
		return pgtype.UUID{}, false
	}
	if _, err := h.queries.GetProject(r.Context(), db.GetProjectParams{ID: id, OrgID: orgID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusBadRequest, "project not found", "bad_request")
			return pgtype.UUID{}, false
		}
		respond.InternalError(w)
		return pgtype.UUID{}, false
	}
	return pgtype.UUID{Bytes: id, Valid: true}, true
}

// monitorProjectSetters sets one monitor's project, one closure per monitor
// type — see projectAssigners.
var monitorProjectSetters = map[string]func(ctx context.Context, q *db.Queries, orgID, monitorID uuid.UUID, projectID pgtype.UUID) error{
	"cron": func(ctx context.Context, q *db.Queries, orgID, monitorID uuid.UUID, projectID pgtype.UUID) error {
		_, err := q.SetCronMonitorProject(ctx, db.SetCronMonitorProjectParams{ID: monitorID, OrgID: orgID, ProjectID: projectID})
		return err
	},
	"uptime": func(ctx context.Context, q *db.Queries, orgID, monitorID uuid.UUID, projectID pgtype.UUID) error {
		_, err := q.SetUptimeMonitorProject(ctx, db.SetUptimeMonitorProjectParams{ID: monitorID, OrgID: orgID, ProjectID: projectID})
		return err
	},
	"ssl": func(ctx context.Context, q *db.Queries, orgID, monitorID uuid.UUID, projectID pgtype.UUID) error {
		_, err := q.SetSSLMonitorProject(ctx, db.SetSSLMonitorProjectParams{ID: monitorID, OrgID: orgID, ProjectID: projectID})
		return err
	},
	"domain": func(ctx context.Context, q *db.Queries, orgID, monitorID uuid.UUID, projectID pgtype.UUID) error {
		_, err := q.SetDomainMonitorProject(ctx, db.SetDomainMonitorProjectParams{ID: monitorID, OrgID: orgID, ProjectID: projectID})
		return err
	},
	"port": func(ctx context.Context, q *db.Queries, orgID, monitorID uuid.UUID, projectID pgtype.UUID) error {
		_, err := q.SetPortMonitorProject(ctx, db.SetPortMonitorProjectParams{ID: monitorID, OrgID: orgID, ProjectID: projectID})
		return err
	},
	"dns": func(ctx context.Context, q *db.Queries, orgID, monitorID uuid.UUID, projectID pgtype.UUID) error {
		_, err := q.SetDNSMonitorProject(ctx, db.SetDNSMonitorProjectParams{ID: monitorID, OrgID: orgID, ProjectID: projectID})
		return err
	},
}

// applyMonitorProject writes a resolved projectId onto a monitor after its
// create or update, and into dst, the written monitor's ProjectID, so its
// response shows it. It does nothing when the request left projectId out.
func (h *MonitorHandler) applyMonitorProject(ctx context.Context, orgID uuid.UUID, monitorType string, monitorID uuid.UUID, p optionalProjectID, projectID pgtype.UUID, dst *pgtype.UUID) error {
	if !p.Set {
		return nil
	}
	if err := monitorProjectSetters[monitorType](ctx, h.queries, orgID, monitorID, projectID); err != nil {
		return err
	}
	*dst = projectID
	return nil
}

// projectIDString renders a monitor's nullable project_id for its response.
func projectIDString(id pgtype.UUID) *string {
	if !id.Valid {
		return nil
	}
	s := uuid.UUID(id.Bytes).String()
	return &s
}
