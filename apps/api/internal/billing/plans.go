package billing

import (
	"errors"

	"github.com/checkmeup/checkmeup/internal/db"
)

type Limits struct {
	MonitorTotal    int // -1 = unlimited
	StatusPages     int // -1 = unlimited
	MinIntervalMins int // minimum uptime check interval
}

var planLimits = map[db.Plan]Limits{
	db.PlanHobbyist: {MonitorTotal: 5, StatusPages: 1, MinIntervalMins: 10},
	db.PlanIndie:    {MonitorTotal: 20, StatusPages: 3, MinIntervalMins: 5},
	db.PlanStudio:   {MonitorTotal: 50, StatusPages: 10, MinIntervalMins: 1},
	db.PlanAgency:   {MonitorTotal: -1, StatusPages: -1, MinIntervalMins: 1},
}

var (
	ErrMonitorLimit    = errors.New("monitor limit reached for your plan — upgrade to add more")
	ErrStatusPageLimit = errors.New("status page limit reached for your plan — upgrade to add more")
)

func GetLimits(plan db.Plan) Limits {
	if l, ok := planLimits[plan]; ok {
		return l
	}
	return planLimits[db.PlanHobbyist]
}

func CheckMonitorLimit(plan db.Plan, current int) error {
	l := GetLimits(plan)
	if l.MonitorTotal != -1 && current >= l.MonitorTotal {
		return ErrMonitorLimit
	}
	return nil
}

func CheckStatusPageLimit(plan db.Plan, current int) error {
	l := GetLimits(plan)
	if l.StatusPages != -1 && current >= l.StatusPages {
		return ErrStatusPageLimit
	}
	return nil
}

// ClampInterval returns the interval to use, enforcing the plan minimum.
// Returns 0 and an error if the requested interval is below the plan minimum.
func ClampInterval(plan db.Plan, requestedMins int) (int, error) {
	l := GetLimits(plan)
	if requestedMins < l.MinIntervalMins {
		return l.MinIntervalMins, errors.New("check interval is below your plan minimum")
	}
	return requestedMins, nil
}
