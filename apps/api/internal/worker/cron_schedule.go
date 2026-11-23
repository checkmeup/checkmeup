package worker

import (
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// CronParser reads a cron monitor's schedule (PRD cron-monitoring) and the
// free cron explainer's input (PRD free-tools): the 5 standard fields, with
// no seconds field and no @hourly-style descriptors.
var CronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

// NextPing returns when a cron monitor's next ping is expected, after which
// the grace period starts. Formula: next scheduled time after now + grace buffer.
func NextPing(schedule string, now time.Time, graceMins int) time.Time {
	s := strings.ToLower(strings.TrimSpace(schedule))

	if strings.HasPrefix(s, "every ") {
		d := parseEveryDuration(strings.TrimPrefix(s, "every "))
		if d > 0 {
			// next expected ping = now + interval + grace
			return now.Add(d + time.Duration(graceMins)*time.Minute)
		}
	}

	// Cron expression — use robfig/cron to find the next fire time.
	sched, err := CronParser.Parse(schedule)
	if err == nil {
		next := sched.Next(now)
		return next.Add(time.Duration(graceMins) * time.Minute)
	}

	// Fallback: 1 hour from now.
	return now.Add(time.Hour + time.Duration(graceMins)*time.Minute)
}

// parseEveryDuration parses "1h", "1hr", "1hour", "30m", "30min", "30mins",
// "1d", "1day", "1w", "1week" etc.
func parseEveryDuration(s string) time.Duration {
	s = normalizeDurationSuffix(strings.TrimSpace(strings.ToLower(s)))
	if len(s) < 2 {
		return 0
	}
	unit := s[len(s)-1]
	n, err := strconv.Atoi(strings.TrimSpace(s[:len(s)-1]))
	if err != nil || n <= 0 {
		return 0
	}
	return time.Duration(n) * durationUnit(unit)
}

// normalizeDurationSuffix rewrites a word suffix ("minutes", "hour", "days",
// "week", ...) to the single-char unit parseEveryDuration expects.
func normalizeDurationSuffix(s string) string {
	for _, r := range []struct{ old, new string }{
		{"minutes", "m"}, {"minute", "m"}, {"mins", "m"}, {"min", "m"},
		{"hours", "h"}, {"hour", "h"}, {"hrs", "h"}, {"hr", "h"},
		{"weeks", "w"}, {"week", "w"},
		{"days", "d"}, {"day", "d"},
	} {
		if strings.HasSuffix(s, r.old) {
			return strings.TrimSpace(s[:len(s)-len(r.old)]) + r.new
		}
	}
	return s
}

// durationUnit maps a single-char unit to its time.Duration multiplier, or
// 0 for an unrecognized unit (parseEveryDuration treats that as invalid).
func durationUnit(unit byte) time.Duration {
	switch unit {
	case 'm':
		return time.Minute
	case 'h':
		return time.Hour
	case 'd':
		return 24 * time.Hour
	case 'w':
		return 7 * 24 * time.Hour
	}
	return 0
}
