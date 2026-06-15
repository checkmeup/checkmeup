package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,46}[a-z0-9]$`)

// StatusPageHandler handles status page admin endpoints.
type StatusPageHandler struct {
	queries *db.Queries
}

func NewStatusPageHandler(pool *pgxpool.Pool) *StatusPageHandler {
	return &StatusPageHandler{queries: db.New(pool)}
}

// ─── response types ──────────────────────────────────────────────────────────

type statusPageResponse struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	LogoURL     string `json:"logoUrl"`
	PublicURL   string `json:"publicUrl"`
	CreatedAt   string `json:"createdAt"`
}

type statusPageMonitorResponse struct {
	ID           string `json:"id"`
	MonitorType  string `json:"monitorType"`
	MonitorID    string `json:"monitorId"`
	DisplayName  string `json:"displayName"`
	DisplayOrder int32  `json:"displayOrder"`
}

type statusPageDetailResponse struct {
	statusPageResponse
	Monitors []statusPageMonitorResponse `json:"monitors"`
}

func toStatusPageResponse(p db.StatusPage, baseURL string) statusPageResponse {
	return statusPageResponse{
		ID:          p.ID.String(),
		Slug:        p.Slug,
		Title:       p.Title,
		Description: p.Description,
		LogoURL:     p.LogoUrl,
		PublicURL:   baseURL + "/status/" + p.Slug,
		CreatedAt:   p.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

func toStatusPageMonitorResponse(m db.StatusPageMonitor) statusPageMonitorResponse {
	return statusPageMonitorResponse{
		ID:           m.ID.String(),
		MonitorType:  m.MonitorType,
		MonitorID:    m.MonitorID.String(),
		DisplayName:  m.DisplayName,
		DisplayOrder: m.DisplayOrder,
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func validateSlug(s string) error {
	if !slugRe.MatchString(s) {
		return errors.New("slug must be 3–48 lowercase letters, numbers, or hyphens (not starting/ending with a hyphen)")
	}
	return nil
}

func statusPageIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return uuid.UUID{}, uuid.UUID{}, false
	}
	pageID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid page id", "bad_request")
		return uuid.UUID{}, uuid.UUID{}, false
	}
	return orgID, pageID, true
}

func baseURL(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	return scheme + "://" + r.Host
}

// ─── handlers ────────────────────────────────────────────────────────────────

// CheckSlug GET /api/v1/status-pages/check-slug?slug=foo
func (h *StatusPageHandler) CheckSlug(w http.ResponseWriter, r *http.Request) {
	slug := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("slug")))
	if err := validateSlug(slug); err != nil {
		respond.JSON(w, http.StatusOK, map[string]any{"available": false, "reason": err.Error()})
		return
	}
	available, err := h.queries.SlugAvailable(r.Context(), slug)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	reason := ""
	if !available {
		reason = "slug is already taken"
	}
	respond.JSON(w, http.StatusOK, map[string]any{"available": available, "reason": reason})
}

// ListStatusPages GET /api/v1/status-pages
func (h *StatusPageHandler) ListStatusPages(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}
	pages, err := h.queries.ListStatusPages(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	base := baseURL(r)
	result := make([]statusPageResponse, len(pages))
	for i, p := range pages {
		result[i] = toStatusPageResponse(p, base)
	}
	respond.JSON(w, http.StatusOK, result)
}

type createStatusPageRequest struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	LogoURL     string `json:"logoUrl"`
}

// CreateStatusPage POST /api/v1/status-pages
func (h *StatusPageHandler) CreateStatusPage(w http.ResponseWriter, r *http.Request) {
	orgID, err := orgIDFrom(r)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "authentication required", "unauthenticated")
		return
	}
	var req createStatusPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	req.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	if err := validateSlug(req.Slug); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		respond.Error(w, http.StatusBadRequest, "title is required", "bad_request")
		return
	}

	page, err := h.queries.CreateStatusPage(r.Context(), db.CreateStatusPageParams{
		OrgID:       orgID,
		Slug:        req.Slug,
		Title:       req.Title,
		Description: strings.TrimSpace(req.Description),
		LogoUrl:     strings.TrimSpace(req.LogoURL),
	})
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			respond.Error(w, http.StatusConflict, "slug is already taken", "slug_taken")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	respond.JSON(w, http.StatusCreated, toStatusPageResponse(page, baseURL(r)))
}

// GetStatusPage GET /api/v1/status-pages/:id
func (h *StatusPageHandler) GetStatusPage(w http.ResponseWriter, r *http.Request) {
	orgID, pageID, ok := statusPageIDs(w, r)
	if !ok {
		return
	}
	page, err := h.queries.GetStatusPage(r.Context(), db.GetStatusPageParams{ID: pageID, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "page not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	monitors, err := h.queries.ListStatusPageMonitors(r.Context(), pageID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	monResp := make([]statusPageMonitorResponse, len(monitors))
	for i, m := range monitors {
		monResp[i] = toStatusPageMonitorResponse(m)
	}
	resp := statusPageDetailResponse{
		statusPageResponse: toStatusPageResponse(page, baseURL(r)),
		Monitors:           monResp,
	}
	respond.JSON(w, http.StatusOK, resp)
}

type updateStatusPageRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	LogoURL     string `json:"logoUrl"`
}

// UpdateStatusPage PATCH /api/v1/status-pages/:id
func (h *StatusPageHandler) UpdateStatusPage(w http.ResponseWriter, r *http.Request) {
	orgID, pageID, ok := statusPageIDs(w, r)
	if !ok {
		return
	}
	var req updateStatusPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		respond.Error(w, http.StatusBadRequest, "title is required", "bad_request")
		return
	}
	page, err := h.queries.UpdateStatusPage(r.Context(), db.UpdateStatusPageParams{
		ID:          pageID,
		OrgID:       orgID,
		Title:       req.Title,
		Description: strings.TrimSpace(req.Description),
		LogoUrl:     strings.TrimSpace(req.LogoURL),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "page not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	respond.JSON(w, http.StatusOK, toStatusPageResponse(page, baseURL(r)))
}

// DeleteStatusPage DELETE /api/v1/status-pages/:id
func (h *StatusPageHandler) DeleteStatusPage(w http.ResponseWriter, r *http.Request) {
	orgID, pageID, ok := statusPageIDs(w, r)
	if !ok {
		return
	}
	if err := h.queries.DeleteStatusPage(r.Context(), db.DeleteStatusPageParams{
		ID: pageID, OrgID: orgID,
	}); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type setMonitorsRequest struct {
	Monitors []setMonitorItem `json:"monitors"`
}

type setMonitorItem struct {
	MonitorType  string `json:"monitorType"`
	MonitorID    string `json:"monitorId"`
	DisplayName  string `json:"displayName"`
	DisplayOrder int32  `json:"displayOrder"`
}

// SetStatusPageMonitors PUT /api/v1/status-pages/:id/monitors
func (h *StatusPageHandler) SetStatusPageMonitors(w http.ResponseWriter, r *http.Request) {
	orgID, pageID, ok := statusPageIDs(w, r)
	if !ok {
		return
	}

	// Verify page belongs to org
	if _, err := h.queries.GetStatusPage(r.Context(), db.GetStatusPageParams{
		ID: pageID, OrgID: orgID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond.Error(w, http.StatusNotFound, "page not found", "not_found")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}

	var req setMonitorsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	validTypes := map[string]bool{"cron": true, "uptime": true, "ssl": true}
	for _, m := range req.Monitors {
		if !validTypes[m.MonitorType] {
			respond.Error(w, http.StatusBadRequest, "invalid monitor type: "+m.MonitorType, "bad_request")
			return
		}
		if _, err := uuid.Parse(m.MonitorID); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid monitor id", "bad_request")
			return
		}
	}

	// Replace all monitors atomically (delete + insert)
	if err := h.queries.DeleteStatusPageMonitors(r.Context(), pageID); err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	result := make([]statusPageMonitorResponse, 0, len(req.Monitors))
	for _, m := range req.Monitors {
		monID, _ := uuid.Parse(m.MonitorID)
		displayName := strings.TrimSpace(m.DisplayName)
		if displayName == "" {
			displayName = m.MonitorID // fallback, shouldn't happen
		}
		inserted, err := h.queries.InsertStatusPageMonitor(r.Context(), db.InsertStatusPageMonitorParams{
			PageID:       pageID,
			MonitorType:  m.MonitorType,
			MonitorID:    monID,
			DisplayName:  displayName,
			DisplayOrder: m.DisplayOrder,
		})
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
			return
		}
		result = append(result, toStatusPageMonitorResponse(inserted))
	}
	respond.JSON(w, http.StatusOK, result)
}
