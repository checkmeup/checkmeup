package billing

import (
	"errors"

	"github.com/checkmeup/checkmeup/internal/db"
)

type Limits struct {
	MonitorTotal      int  // -1 = unlimited
	StatusPages       int  // -1 = unlimited
	MinIntervalMins   int  // minimum uptime check interval
	KeywordMonitoring bool // keyword/content checks on uptime monitors (EP-11)
}

var planLimits = map[db.Plan]Limits{
	db.PlanHobby:      {MonitorTotal: 10, StatusPages: 1, MinIntervalMins: 5, KeywordMonitoring: false},
	db.PlanSolo:       {MonitorTotal: 30, StatusPages: 3, MinIntervalMins: 1, KeywordMonitoring: true},
	db.PlanStartup:    {MonitorTotal: 100, StatusPages: 10, MinIntervalMins: 1, KeywordMonitoring: true},
	db.PlanEnterprise: {MonitorTotal: 1000, StatusPages: 100, MinIntervalMins: 1, KeywordMonitoring: true},
}

var (
	ErrMonitorLimit      = errors.New("monitor limit reached for your plan — upgrade to add more")
	ErrStatusPageLimit   = errors.New("status page limit reached for your plan — upgrade to add more")
	ErrKeywordMonitoring = errors.New("keyword monitoring is available on paid plans — upgrade to use it")
)

func GetLimits(plan db.Plan) Limits {
	if l, ok := planLimits[plan]; ok {
		return l
	}
	return planLimits[db.PlanHobby]
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

// CheckKeywordMonitoringAllowed returns an error if plan doesn't include
// keyword monitoring (EP-11, ADR-019: Hobby is excluded).
func CheckKeywordMonitoringAllowed(plan db.Plan) error {
	if !GetLimits(plan).KeywordMonitoring {
		return ErrKeywordMonitoring
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
