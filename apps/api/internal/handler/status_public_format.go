package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/checkmeup/checkmeup/internal/db"
)

// initials derives a 1-2 letter fallback badge from a status page's title
// (e.g. "Acme Corp" -> "AC"), shown when the org hasn't uploaded a custom
// logo — mirrors the avatar-badge pattern in the CheckMeUp Design mockup.
func initials(title string) string {
	fields := strings.Fields(title)
	out := ""
	for _, f := range fields {
		r := []rune(f)
		if len(r) == 0 {
			continue
		}
		out += strings.ToUpper(string(r[0]))
		if len(out) >= 2 {
			break
		}
	}
	if out == "" {
		return "?"
	}
	return out
}

// relativeTime formats a past timestamp the way the incident timeline design
// does ("58 min ago", "2h 10min ago", "3 days ago") — coarser than a clock
// time since visitors care about recency, not the exact minute.
func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		h := int(d.Hours())
		m := int(d.Minutes()) - h*60
		if m == 0 {
			return fmt.Sprintf("%dh ago", h)
		}
		return fmt.Sprintf("%dh %dmin ago", h, m)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// formatDuration renders an incident's total open time for the resolved
// history ("45 min", "1h 12min") — same coarse-grain style as relativeTime.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "< 1 min"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d min", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) - h*60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dmin", h, m)
}

// incidentSeverityLabel/-Color mirror the minor/major/critical vocabulary
// from US-2401 — minor is informational (blue) and doesn't escalate the
// overall banner; major/critical reuse the same red/amber tiers monitor
// outages already use, so computeOverallStatus treats them identically.
func incidentSeverityLabel(s db.IncidentSeverity) string {
	switch s {
	case db.IncidentSeverityCritical:
		return "Critical"
	case db.IncidentSeverityMajor:
		return "Major"
	default:
		return "Minor"
	}
}

func incidentSeverityColor(s db.IncidentSeverity) string {
	switch s {
	case db.IncidentSeverityCritical:
		return statusColorRed
	case db.IncidentSeverityMajor:
		return statusColorAmber
	default:
		return statusColorBlue
	}
}

func incidentStatusLabel(s db.IncidentStatus) string {
	switch s {
	case db.IncidentStatusIdentified:
		return "Identified"
	case db.IncidentStatusMonitoring:
		return "Monitoring"
	case db.IncidentStatusResolved:
		return "Resolved"
	default:
		return "Investigating"
	}
}

func expiryStatusDisplay(s string) (label, color string) {
	switch s {
	case "up":
		return "Valid", statusColorGreen
	case "expiring_soon":
		return "Expiring soon", statusColorAmber
	case "expired":
		return "Expired", statusColorRed
	case "error":
		return "Error", statusColorRed
	case "paused":
		return "Paused", statusColorGray
	default:
		return "Checking…", statusColorGray
	}
}

func solidColorBar(status string) []publicBar {
	color := statusColorGreen
	switch status {
	case "expiring_soon":
		color = statusColorAmber
	case "expired", "error":
		color = statusColorRed
	case "waiting", "paused":
		color = statusColorLight
	}
	bar := make([]publicBar, 90)
	for i := range bar {
		bar[i] = publicBar{Color: color}
	}
	return bar
}

func monitorStatusDisplay(s string) (label, color string) {
	switch s {
	case "up":
		return "Operational", statusColorGreen
	case "down":
		return "Down", statusColorRed
	case "paused":
		return "Paused", statusColorGray
	default:
		return "Checking…", statusColorGray
	}
}
