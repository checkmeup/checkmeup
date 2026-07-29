package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/checkmeup/checkmeup/internal/billing"
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
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	LogoURL      string `json:"logoUrl"`
	HideBranding bool   `json:"hideBranding"`
	Layout       string `json:"layout"`
	PublicURL    string `json:"publicUrl"`
	CreatedAt    string `json:"createdAt"`
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
		ID:           p.ID.String(),
		Slug:         p.Slug,
		Title:        p.Title,
		Description:  p.Description,
		LogoURL:      p.LogoUrl,
		HideBranding: p.HideBranding,
		Layout:       p.Layout,
		PublicURL:    baseURL + "/status/" + p.Slug,
		CreatedAt:    p.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
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

// validateLayout rejects anything but the two layouts the public template
// (status_public.go) knows how to render — ADR-038.
func validateLayout(s string) error {
	if s != "classic" && s != "grid" {
		return errors.New("layout must be \"classic\" or \"grid\"")
	}
	return nil
}

// validateLogoURL rejects anything but an absolute http(s) URL, so the
// public status page (status_public.go, unauthenticated) never has to
// render an org-supplied javascript:/data: URI into a <link>/<img> tag.
// An empty string (no logo) is valid.
func validateLogoURL(s string) error {
	if s == "" {
		return nil
	}
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("logo URL must be an absolute http:// or https:// URL")
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

// respondStatusPageNotFoundOrInternal writes a 404 if err is pgx.ErrNoRows,
// a 500 for any other non-nil error, and reports whether it wrote a
// response so the caller knows to return immediately.
func respondStatusPageNotFoundOrInternal(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, pgx.ErrNoRows) {
		respond.Error(w, http.StatusNotFound, "page not found", "not_found")
		return true
	}
	respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
	return true
}

// validateStatusPageTitleAndLogo trims and validates the title/logo fields
// shared by create and update requests, responding with the first
// validation failure and returning false.
func validateStatusPageTitleAndLogo(w http.ResponseWriter, title, logoURL *string) bool {
	*title = strings.TrimSpace(*title)
	if *title == "" {
		respond.Error(w, http.StatusBadRequest, "title is required", "bad_request")
		return false
	}
	*logoURL = strings.TrimSpace(*logoURL)
	if err := validateLogoURL(*logoURL); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return false
	}
	return true
}

// checkStatusPageCreateLimit checks whether orgID can create another status
// page under its plan. Responds and returns false on any failure (query
// error or plan limit hit).
func (h *StatusPageHandler) checkStatusPageCreateLimit(w http.ResponseWriter, r *http.Request, orgID uuid.UUID) bool {
	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return false
	}
	spCount, err := h.queries.CountOrgStatusPages(r.Context(), orgID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return false
	}
	if err := billing.CheckStatusPageLimit(plan, int(spCount)); err != nil {
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", "status_page")
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return false
	}
	return true
}

// resolvedStatusPageMonitor is a request monitor item after resolving its
// display name and confirming org ownership (see resolveStatusPageMonitors).
type resolvedStatusPageMonitor struct {
	monitorType  string
	monitorID    uuid.UUID
	displayName  string
	displayOrder int32
}

// resolveStatusPageMonitors resolves each requested monitor against orgID.
// resolveMonitorName (shared with maintenance.go) scopes its lookup by
// org_id, so this also confirms ownership — a status page can never
// reference another org's monitor. It also supplies the real monitor name as
// the displayName fallback (EP-06 US-0602: "defaults to monitor name"),
// rather than falling back to the raw UUID string.
func (h *StatusPageHandler) resolveStatusPageMonitors(ctx context.Context, orgID uuid.UUID, items []setMonitorItem) ([]resolvedStatusPageMonitor, error) {
	resolved := make([]resolvedStatusPageMonitor, 0, len(items))
	for _, m := range items {
		id, name, err := resolveMonitorName(ctx, h.queries, orgID, m.MonitorType, m.MonitorID)
		if err != nil {
			return nil, err
		}
		displayName := strings.TrimSpace(m.DisplayName)
		if displayName == "" {
			displayName = name
		}
		resolved = append(resolved, resolvedStatusPageMonitor{
			monitorType: m.MonitorType, monitorID: id, displayName: displayName, displayOrder: m.DisplayOrder,
		})
	}
	return resolved, nil
}

// replaceStatusPageMonitors atomically swaps pageID's monitor list for
// resolved (delete + insert).
func (h *StatusPageHandler) replaceStatusPageMonitors(ctx context.Context, pageID uuid.UUID, resolved []resolvedStatusPageMonitor) ([]statusPageMonitorResponse, error) {
	if err := h.queries.DeleteStatusPageMonitors(ctx, pageID); err != nil {
		return nil, err
	}
	result := make([]statusPageMonitorResponse, 0, len(resolved))
	for _, m := range resolved {
		inserted, err := h.queries.InsertStatusPageMonitor(ctx, db.InsertStatusPageMonitorParams{
			PageID:       pageID,
			MonitorType:  m.monitorType,
			MonitorID:    m.monitorID,
			DisplayName:  m.displayName,
			DisplayOrder: m.displayOrder,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, toStatusPageMonitorResponse(inserted))
	}
	return result, nil
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
	Layout      string `json:"layout"`
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
	if !validateStatusPageTitleAndLogo(w, &req.Title, &req.LogoURL) {
		return
	}
	if req.Layout == "" {
		req.Layout = "classic"
	}
	if err := validateLayout(req.Layout); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}
	if !h.checkStatusPageCreateLimit(w, r, orgID) {
		return
	}

	page, err := h.queries.CreateStatusPage(r.Context(), db.CreateStatusPageParams{
		OrgID:       orgID,
		Slug:        req.Slug,
		Title:       req.Title,
		Description: strings.TrimSpace(req.Description),
		LogoUrl:     req.LogoURL,
		Layout:      req.Layout,
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
	if respondStatusPageNotFoundOrInternal(w, err) {
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
	Title        string `json:"title"`
	Description  string `json:"description"`
	LogoURL      string `json:"logoUrl"`
	HideBranding bool   `json:"hideBranding"`
	Layout       string `json:"layout"`
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
	if !validateStatusPageTitleAndLogo(w, &req.Title, &req.LogoURL) {
		return
	}
	if req.Layout == "" {
		req.Layout = "classic"
	}
	if err := validateLayout(req.Layout); err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}

	// ADR-035: hide_branding can only be turned on for orgs on a paid plan.
	if req.HideBranding {
		plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
			return
		}
		if !billing.GetLimits(plan).HideBrandingAllowed {
			respond.Error(w, http.StatusPaymentRequired, "hiding branding requires a paid plan — upgrade to enable this", "plan_limit_reached")
			return
		}
	}

	page, err := h.queries.UpdateStatusPage(r.Context(), db.UpdateStatusPageParams{
		ID:           pageID,
		OrgID:        orgID,
		Title:        req.Title,
		Description:  strings.TrimSpace(req.Description),
		LogoUrl:      req.LogoURL,
		HideBranding: req.HideBranding,
		Layout:       req.Layout,
	})
	if respondStatusPageNotFoundOrInternal(w, err) {
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
	_, err := h.queries.GetStatusPage(r.Context(), db.GetStatusPageParams{
		ID: pageID, OrgID: orgID,
	})
	if respondStatusPageNotFoundOrInternal(w, err) {
		return
	}

	var req setMonitorsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body", "bad_request")
		return
	}

	resolved, err := h.resolveStatusPageMonitors(r.Context(), orgID, req.Monitors)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error(), "bad_request")
		return
	}

	result, err := h.replaceStatusPageMonitors(r.Context(), pageID, resolved)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
		return
	}
	respond.JSON(w, http.StatusOK, result)
}
