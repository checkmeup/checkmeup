package billing

import (
	"errors"

	"github.com/checkmeup/checkmeup/internal/db"
)

type Limits struct {
	MonitorTotal         int // -1 = unlimited
	StatusPages          int // -1 = unlimited
	NotificationChannels int // -1 = unlimited
	MinIntervalMins      int // minimum uptime check interval
	SMSCredits           int // monthly SMS sends allowed (ADR-032); 0 = sms unavailable on this plan
}

var planLimits = map[db.Plan]Limits{
	db.PlanHobby:      {MonitorTotal: 10, StatusPages: 1, NotificationChannels: 5, MinIntervalMins: 5, SMSCredits: 0},
	db.PlanSolo:       {MonitorTotal: 30, StatusPages: 3, NotificationChannels: 20, MinIntervalMins: 1, SMSCredits: 10},
	db.PlanStartup:    {MonitorTotal: 100, StatusPages: 10, NotificationChannels: 50, MinIntervalMins: 1, SMSCredits: 30},
	db.PlanEnterprise: {MonitorTotal: 1000, StatusPages: 100, NotificationChannels: 100, MinIntervalMins: 1, SMSCredits: 100},
}

var (
	ErrMonitorLimit             = errors.New("monitor limit reached for your plan — upgrade to add more")
	ErrStatusPageLimit          = errors.New("status page limit reached for your plan — upgrade to add more")
	ErrNotificationChannelLimit = errors.New("notification channel limit reached for your plan — upgrade to add more")
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

func CheckNotificationChannelLimit(plan db.Plan, current int) error {
	l := GetLimits(plan)
	if l.NotificationChannels != -1 && current >= l.NotificationChannels {
		return ErrNotificationChannelLimit
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
