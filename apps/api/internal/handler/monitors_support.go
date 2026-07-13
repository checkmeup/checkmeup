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
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/checkmeup/checkmeup/internal/billing"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

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
	return r
}

func cronIncidentToResponse(inc db.CronIncident) cronIncidentResponse {
	ir := cronIncidentResponse{
		ID:        inc.ID.String(),
		StartedAt: inc.StartedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	if inc.ResolvedAt.Valid {
		t := inc.ResolvedAt.Time.Format("2006-01-02T15:04:05Z")
		ir.ResolvedAt = &t
	}
	return ir
}

func normalizeAndValidateCreateCronMonitorRequest(req *createCronMonitorRequest) error {
	return normalizeAndValidateCronMonitorFields(&req.Name, &req.Schedule, &req.GracePeriodMins, &req.MaxAlertsPerIncident)
}

func normalizeAndValidateUpdateCronMonitorRequest(req *updateCronMonitorRequest) error {
	return normalizeAndValidateCronMonitorFields(&req.Name, &req.Schedule, &req.GracePeriodMins, &req.MaxAlertsPerIncident)
}

// normalizeAndValidateCronMonitorFields holds the field-level normalization
// and validation shared by create/update cron monitor requests — factored
// out (rather than duplicated per request type) since the two structs only
// differ in fields this validation doesn't touch (ChannelIDs,
// AlertAfterNFailures, and update's AlertsEnabled).
func normalizeAndValidateCronMonitorFields(name, schedule *string, gracePeriodMins, maxAlertsPerIncident *int32) error {
	*name = strings.TrimSpace(*name)
	*schedule = strings.TrimSpace(*schedule)

	if *name == "" {
		return errors.New("name is required")
	}
	if err := validateSchedule(*schedule); err != nil {
		return err
	}
	if *gracePeriodMins < 1 {
		*gracePeriodMins = 5
	}
	if *maxAlertsPerIncident < 0 {
		*maxAlertsPerIncident = 3
	}
	return nil
}

// respondMonitorCreateLimitErr maps a checkMonitorCreateLimit failure to an
// HTTP response — a plan-limit hit is a 402 with the billing error's own
// message, anything else (a query failure) is a generic 500.
func respondMonitorCreateLimitErr(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, resource string, err error) {
	if errors.Is(err, billing.ErrMonitorLimit) {
		slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "resource", resource)
		respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
		return
	}
	respond.Error(w, http.StatusInternalServerError, "internal error", "internal_error")
}

// createCronMonitorRow generates the monitor's ping token and creates the row.
func (h *MonitorHandler) createCronMonitorRow(ctx context.Context, orgID uuid.UUID, req createCronMonitorRequest) (db.CronMonitor, error) {
	token, err := generatePingToken()
	if err != nil {
		return db.CronMonitor{}, err
	}
	return h.queries.CreateCronMonitor(ctx, db.CreateCronMonitorParams{
		OrgID:                orgID,
		Name:                 req.Name,
		Schedule:             req.Schedule,
		GracePeriodMins:      req.GracePeriodMins,
		PingToken:            token,
		MaxAlertsPerIncident: req.MaxAlertsPerIncident,
		AlertAfterNFailures:  req.AlertAfterNFailures,
	})
}

// checkMonitorCreateLimit checks whether orgID can create another monitor
// under its plan's total-monitor cap — the same cap applies across all
// monitor types (cron, uptime, SSL, ...).
func (h *MonitorHandler) checkMonitorCreateLimit(ctx context.Context, orgID uuid.UUID) error {
	plan, err := h.queries.GetOrgPlan(ctx, orgID)
	if err != nil {
		return err
	}
	total, err := h.queries.CountOrgMonitors(ctx, orgID)
	if err != nil {
		return err
	}
	return billing.CheckMonitorLimit(plan, int(total))
}

func generatePingToken() (string, error) {
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

// loadCronMonitorDetail fetches the pings/incidents shown on the cron
// monitor detail page and converts them to response types.
func (h *MonitorHandler) loadCronMonitorDetail(ctx context.Context, monitorID uuid.UUID) ([]cronPingResponse, []cronIncidentResponse, error) {
	pings, err := h.queries.ListCronPings(ctx, db.ListCronPingsParams{
		MonitorID: monitorID,
		Limit:     50,
		Offset:    0,
	})
	if err != nil {
		return nil, nil, err
	}

	incidents, err := h.queries.ListCronIncidents(ctx, monitorID)
	if err != nil {
		return nil, nil, err
	}

	pingResp := make([]cronPingResponse, len(pings))
	for i, p := range pings {
		pingResp[i] = cronPingToResponse(p)
	}

	incidentResp := make([]cronIncidentResponse, len(incidents))
	for i, inc := range incidents {
		incidentResp[i] = cronIncidentToResponse(inc)
	}

	return pingResp, incidentResp, nil
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
