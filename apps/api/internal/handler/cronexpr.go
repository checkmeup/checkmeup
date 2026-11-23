package handler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/checkmeup/checkmeup/internal/worker"
)

// cronParser reads the free cron explainer's input (PRD free-tools) as a
// cron monitor's schedule is read, so the explainer's next runs are the ones
// the monitor expects.
var cronParser = worker.CronParser

// cronUnit is one of the 5 fields: its name in error messages, its range,
// and how its values and steps read in a description.
type cronUnit struct {
	name             string
	singular, plural string
	min, max         int
	names            []string // value names from min upward, for month and day of week
}

var (
	monthNames   = []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	weekdayNames = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

	cronUnits = [5]cronUnit{
		{name: "Minute", singular: "minute", plural: "minutes", max: 59},
		{name: "Hour", singular: "hour", plural: "hours", max: 23},
		{name: "Day of month", singular: "day", plural: "days", min: 1, max: 31},
		{name: "Month", singular: "month", plural: "months", min: 1, max: 12, names: monthNames},
		{name: "Day of week", singular: "day of the week", plural: "days of the week", max: 6, names: weekdayNames},
	}
)

// parseCronExpression reads expr as 5 whitespace-separated cron fields and
// returns its schedule and fields, or an error naming the field it can't
// read ("Minute: 60 is above the maximum, 59"). It also refuses an
// expression no date ever matches, such as 30 February, which would
// otherwise never fire.
func parseCronExpression(expr string) (cron.Schedule, []string, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, nil, fmt.Errorf("a cron expression has 5 fields (minute, hour, day of month, month, day of week), this one has %d", len(fields))
	}
	for i, f := range fields {
		// Each field on its own, every other field "*", so an error names it.
		probe := []string{"*", "*", "*", "*", "*"}
		probe[i] = f
		if strings.ContainsAny(f, "=@") {
			return nil, nil, fmt.Errorf("%s: %q isn't a value cron reads", cronUnits[i].name, f)
		}
		if _, err := cronParser.Parse(strings.Join(probe, " ")); err != nil {
			return nil, nil, fmt.Errorf("%s: %s", cronUnits[i].name, readableCronError(err))
		}
	}
	sched, err := cronParser.Parse(strings.Join(fields, " "))
	if err != nil {
		return nil, nil, fmt.Errorf("can't read this expression: %s", readableCronError(err))
	}
	if sched.Next(time.Now()).IsZero() {
		return nil, nil, fmt.Errorf("this schedule never runs: none of its months has that day")
	}
	return sched, fields, nil
}

var cronErrorRewrites = []struct {
	pattern *regexp.Regexp
	format  string
}{
	{regexp.MustCompile(`^end of range \((\d+)\) above maximum \((\d+)\)`), "%s is above the maximum, %s"},
	{regexp.MustCompile(`^beginning of range \((\d+)\) below minimum \((\d+)\)`), "%s is below the minimum, %s"},
	{regexp.MustCompile(`^beginning of range \((\d+)\) beyond end of range \((\d+)\)`), "a range can't start at %s, after its end, %s"},
	{regexp.MustCompile(`^failed to parse int from ([^:]+):`), "%q isn't a number or a name cron knows"},
	{regexp.MustCompile(`^failed to parse int from :`), "a value is missing"},
	{regexp.MustCompile(`^step of range should be a positive number`), "a step must be 1 or more"},
	{regexp.MustCompile(`^negative number`), "negative numbers aren't allowed"},
	{regexp.MustCompile(`^too many hyphens`), "a range has one hyphen"},
	{regexp.MustCompile(`^too many slashes`), "a step has one slash"},
}

// readableCronError rewrites one of robfig/cron's field errors in plain words.
func readableCronError(err error) string {
	msg := err.Error()
	for _, rw := range cronErrorRewrites {
		if m := rw.pattern.FindStringSubmatch(msg); m != nil {
			args := make([]any, len(m)-1)
			for i, s := range m[1:] {
				args[i] = s
			}
			return fmt.Sprintf(rw.format, args...)
		}
	}
	return msg
}

var cronNamePattern = regexp.MustCompile(`[A-Za-z]+`)

// normalizeCronField writes a field's month or weekday names as numbers and
// "?" as "*", as robfig/cron reads them.
func normalizeCronField(f string, u cronUnit) string {
	f = strings.ReplaceAll(f, "?", "*")
	if u.names == nil {
		return f
	}
	return cronNamePattern.ReplaceAllStringFunc(f, func(name string) string {
		for i, n := range u.names {
			if strings.EqualFold(n[:3], name) {
				return strconv.Itoa(u.min + i)
			}
		}
		return name
	})
}

// describeCron says in plain English when a parsed expression's fields fire:
// "Every 5 minutes", "At 09:00, Monday through Friday".
func describeCron(fields []string) string {
	var f [5]string
	for i := range f {
		f[i] = normalizeCronField(fields[i], cronUnits[i])
	}
	parts := []string{describeCronTime(f[0], f[1])}
	if days := describeCronDays(f[2], f[4]); days != "" {
		parts = append(parts, days)
	}
	if cronRestricted(f[3]) {
		months := describeCronField(f[3], cronUnits[3])
		if !strings.HasPrefix(months, "every ") {
			months = "in " + months
		}
		parts = append(parts, months)
	}
	if len(parts) == 1 && isCronNumber(f[0]) && isCronList(f[1]) {
		parts = append(parts, "every day")
	}
	return strings.Join(parts, ", ")
}

func describeCronTime(minute, hour string) string {
	if desc, ok := describeCronCommonTime(minute, hour); ok {
		return desc
	}
	desc := describeCronField(minute, cronUnits[0])
	if !strings.HasPrefix(desc, "every ") {
		desc = "at " + desc
	}
	return strings.ToUpper(desc[:1]) + desc[1:] + describeCronHours(hour)
}

// describeCronCommonTime reads the minute and hour together where they make
// a plain phrase: a clock time, "Every hour", "Every 2 hours".
func describeCronCommonTime(minute, hour string) (string, bool) {
	switch {
	case isCronNumber(minute) && isCronList(hour):
		times := strings.Split(hour, ",")
		for i, h := range times {
			times[i] = cronClock(h, minute)
		}
		return "At " + joinAnd(times), true
	case minute == "*" && hour == "*":
		return "Every minute", true
	case minute == "0" && hour == "*":
		return "Every hour", true
	case isCronNumber(minute) && hour == "*":
		return fmt.Sprintf("At %s past every hour", plural(minute, "minute", "minutes")), true
	case minute == "0" && isCronStep(hour, "*"):
		return "Every " + plural(cronStepOf(hour), "hour", "hours"), true
	}
	return "", false
}

// describeCronHours is the hour field after a minute phrase: ", between
// 09:00 and 17:59", ", every 3 hours", or nothing for "*".
func describeCronHours(hour string) string {
	switch {
	case hour == "*":
		return ""
	case isCronNumber(hour):
		return ", between " + cronClock(hour, "0") + " and " + cronClock(hour, "59")
	case isCronRange(hour):
		lo, hi, _ := strings.Cut(hour, "-")
		return ", between " + cronClock(lo, "0") + " and " + cronClock(hi, "59")
	}
	h := describeCronField(hour, cronUnits[1])
	if !strings.HasPrefix(h, "every ") {
		h = "during " + h
	}
	return ", " + h
}

// describeCronDays describes the day-of-month and day-of-week fields. When
// both are restricted, robfig/cron fires on a day matching either, as Vixie
// cron does.
func describeCronDays(dom, dow string) string {
	var parts []string
	if cronRestricted(dom) {
		d := describeCronField(dom, cronUnits[2])
		if !strings.HasPrefix(d, "every ") {
			d = "on " + d + " of the month"
		}
		parts = append(parts, d)
	}
	if cronRestricted(dow) {
		d := describeCronField(dow, cronUnits[4])
		if !strings.HasPrefix(d, "every ") && !isCronRange(dow) {
			d = "on " + d
		}
		parts = append(parts, d)
	}
	return strings.Join(parts, " or ")
}

// describeCronField describes one field in lower case: "every 5 minutes",
// "minutes 0, 15 and 30", "days 1 through 15", "Monday through Friday".
func describeCronField(f string, u cronUnit) string {
	if items := strings.Split(f, ","); len(items) > 1 {
		for i, item := range items {
			items[i] = cronItem(item, u)
		}
		if u.names != nil {
			return joinAnd(items)
		}
		return u.plural + " " + joinAnd(items)
	}
	if base, step, ok := strings.Cut(f, "/"); ok && step != "1" {
		every := "every " + plural(step, u.singular, u.plural)
		switch {
		case base == "*":
			return every
		case isCronRange(base):
			lo, hi, _ := strings.Cut(base, "-")
			return every + " from " + cronValue(lo, u) + " through " + cronValue(hi, u)
		default:
			return every + " from " + cronValue(base, u) + " through " + cronValue(strconv.Itoa(u.max), u)
		}
	} else if ok {
		f = base
	}
	if f == "*" {
		return "every " + u.singular
	}
	if isCronRange(f) {
		if u.names != nil {
			return cronItem(f, u)
		}
		return u.plural + " " + cronItem(f, u)
	}
	if u.names != nil {
		return cronValue(f, u)
	}
	return u.singular + " " + f
}

// cronItem is one list item: a value, a range, or a stepped range as is.
func cronItem(item string, u cronUnit) string {
	if strings.Contains(item, "/") {
		return item
	}
	if lo, hi, ok := strings.Cut(item, "-"); ok {
		return cronValue(lo, u) + " through " + cronValue(hi, u)
	}
	return cronValue(item, u)
}

func cronValue(v string, u cronUnit) string {
	n, err := strconv.Atoi(v)
	if u.names == nil || err != nil || n < u.min || n-u.min >= len(u.names) {
		return v
	}
	return u.names[n-u.min]
}

// cronRestricted is whether a field limits when the schedule fires, as
// robfig/cron's star bit has it: "*" and "*/1" don't, "*/2" does.
func cronRestricted(f string) bool {
	return f != "*" && f != "*/1"
}

func isCronNumber(f string) bool {
	_, err := strconv.Atoi(f)
	return err == nil
}

func isCronList(f string) bool {
	for item := range strings.SplitSeq(f, ",") {
		if !isCronNumber(item) {
			return false
		}
	}
	return true
}

func isCronRange(f string) bool {
	lo, hi, ok := strings.Cut(f, "-")
	return ok && isCronNumber(lo) && isCronNumber(hi)
}

// isCronStep is whether f is base/step with a step over 1.
func isCronStep(f, base string) bool {
	b, step, ok := strings.Cut(f, "/")
	return ok && step != "1" && b == base
}

func cronStepOf(f string) string {
	_, step, _ := strings.Cut(f, "/")
	return step
}

func cronClock(hour, minute string) string {
	h, _ := strconv.Atoi(hour)
	m, _ := strconv.Atoi(minute)
	return fmt.Sprintf("%02d:%02d", h, m)
}

func plural(n, one, many string) string {
	if n == "1" {
		return n + " " + one
	}
	return n + " " + many
}

func joinAnd(items []string) string {
	if len(items) == 1 {
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}
