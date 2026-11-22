package handler

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/checkmeup/checkmeup/internal/billing"
	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/respond"
)

// The plan's limits on monitors (ADR billing), checked at the three moments a
// monitor can grow what an org runs: creating one, changing how often it's
// checked, and resuming a paused one. Each check writes its own response, a
// 500 or a 402 plan_limit_reached, and reports whether the handler may go on.
// kind names the monitor type in the "plan limit hit" log, as "dns_monitor".

// allowNewMonitor reports whether orgID may create another monitor. Monitors
// are counted across every type, paused ones included.
func (h *MonitorHandler) allowNewMonitor(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, kind string) bool {
	_, ok := h.planForNewMonitor(w, r, orgID, kind)
	return ok
}

// allowNewMonitorEvery is allowNewMonitor for a monitor checked every
// intervalMins, which must also be no less than the plan's minimum; it
// returns the interval to store.
func (h *MonitorHandler) allowNewMonitorEvery(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, kind string, intervalMins int32) (int32, bool) {
	plan, ok := h.planForNewMonitor(w, r, orgID, kind)
	if !ok {
		return 0, false
	}
	return allowIntervalOnPlan(w, r, orgID, plan, kind, intervalMins)
}

// allowInterval reports whether orgID's plan allows checking a monitor every
// intervalMins, and returns the interval to store. A request below the plan's
// minimum is refused rather than raised to it, so the user sees why.
func (h *MonitorHandler) allowInterval(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, kind string, intervalMins int32) (int32, bool) {
	plan, ok := h.orgPlan(w, r, orgID)
	if !ok {
		return 0, false
	}
	return allowIntervalOnPlan(w, r, orgID, plan, kind, intervalMins)
}

// allowResume reports whether orgID may resume a paused monitor. Resuming
// creates nothing, but it grows what runs, so after a downgrade it is refused
// once the org's active monitors, paused ones left out, are at the limit.
func (h *MonitorHandler) allowResume(w http.ResponseWriter, r *http.Request, orgID uuid.UUID) bool {
	plan, ok := h.orgPlan(w, r, orgID)
	if !ok {
		return false
	}
	active, err := h.queries.CountActiveMonitorsForOrg(r.Context(), orgID)
	if err != nil {
		respond.InternalError(w)
		return false
	}
	if err := billing.CheckMonitorLimit(plan, int(active)); err != nil {
		refusePlanLimit(w, r, orgID, plan, "monitor_resume", err)
		return false
	}
	return true
}

func (h *MonitorHandler) orgPlan(w http.ResponseWriter, r *http.Request, orgID uuid.UUID) (db.Plan, bool) {
	plan, err := h.queries.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		respond.InternalError(w)
		return "", false
	}
	return plan, true
}

func (h *MonitorHandler) planForNewMonitor(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, kind string) (db.Plan, bool) {
	plan, ok := h.orgPlan(w, r, orgID)
	if !ok {
		return "", false
	}
	total, err := h.queries.CountOrgMonitors(r.Context(), orgID)
	if err != nil {
		respond.InternalError(w)
		return "", false
	}
	if err := billing.CheckMonitorLimit(plan, int(total)); err != nil {
		refusePlanLimit(w, r, orgID, plan, kind, err)
		return "", false
	}
	return plan, true
}

func allowIntervalOnPlan(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, plan db.Plan, kind string, intervalMins int32) (int32, bool) {
	clamped, err := billing.ClampInterval(plan, int(intervalMins))
	if err != nil {
		refusePlanLimit(w, r, orgID, plan, kind+"_interval", err)
		return 0, false
	}
	return int32(clamped), true
}

func refusePlanLimit(w http.ResponseWriter, r *http.Request, orgID uuid.UUID, plan db.Plan, resource string, err error) {
	slog.InfoContext(r.Context(), "plan limit hit", "org_id", orgID, "plan", plan, "resource", resource)
	respond.Error(w, http.StatusPaymentRequired, err.Error(), "plan_limit_reached")
}
