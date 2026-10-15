package handler

import (
	"encoding/json"
	"math"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// statusSummary is a status page's headline and its monitors' 90-day bars
// folded into one, for embedding a live strip elsewhere — the landing page
// shows checkmeup.net's own.
type statusSummary struct {
	Title    string       `json:"title"`
	Overall  string       `json:"overall"`
	Tone     string       `json:"tone"`
	Monitors int          `json:"monitors"`
	Days     []summaryDay `json:"days"`
}

// summaryDay is one day of the strip. Uptime (% of checks that passed) and
// LatencyMs (uptime monitors' average response time) cover the page's
// check-based monitors and are left out on a day none of them ran.
type summaryDay struct {
	Date      string   `json:"date"`
	Tone      string   `json:"tone"`
	Uptime    *float64 `json:"uptime,omitempty"`
	LatencyMs *int32   `json:"latencyMs,omitempty"`
}

// toneRank orders tones by how bad they are, so a day takes its worst
// monitor's tone.
var toneRank = map[string]int{"none": 0, "up": 1, "paused": 2, "warn": 3, "down": 4}

// summarizeDays folds every monitor's 90-day bar into one, oldest day first:
// each day shows the worst tone any monitor had that day.
func summarizeDays(rows []publicMonitorRow) []summaryDay {
	days := make([]summaryDay, 0, 90)
	for i := range 90 {
		days = append(days, summarizeDay(rows, i))
	}
	return days
}

// summarizeDay folds day i of every monitor's bar: the worst tone, plus the
// share of checks that passed and the response time averaged over every
// uptime check that day.
func summarizeDay(rows []publicMonitorRow, i int) summaryDay {
	day := summaryDay{Tone: "none"}
	var checks, down, timedChecks, totalMs int64
	for _, row := range rows {
		if i >= len(row.Bar) {
			continue
		}
		seg := row.Bar[i]
		if day.Date == "" {
			day.Date = seg.Label
		}
		if tone := statusTone(seg.Color); toneRank[tone] > toneRank[day.Tone] {
			day.Tone = tone
		}
		checks += seg.Checks
		down += seg.Down
		if seg.AvgMs > 0 {
			timedChecks += seg.Checks
			totalMs += int64(seg.AvgMs) * seg.Checks
		}
	}
	if checks > 0 {
		uptime := math.Round(float64(checks-down)/float64(checks)*10000) / 100
		day.Uptime = &uptime
	}
	if timedChecks > 0 {
		latency := int32(math.Round(float64(totalMs) / float64(timedChecks)))
		day.LatencyMs = &latency
	}
	return day
}

// ServeSummary handles GET /status/:slug/summary.json — the page's overall
// status and a combined 90-day strip, public like the page itself.
func (h *StatusPublicHandler) ServeSummary(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	ctx := r.Context()

	page, err := h.queries.GetStatusPageBySlug(ctx, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	rows, err := h.loadRows(ctx, page)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	activeIncidents, err := h.loadActiveIncidents(ctx, page)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	overall, color := computeOverallStatus(rows, activeIncidents)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "max-age=60") // same short TTL as the badges
	_ = json.NewEncoder(w).Encode(statusSummary{
		Title:    page.Title,
		Overall:  overall,
		Tone:     statusTone(color),
		Monitors: len(rows),
		Days:     summarizeDays(rows),
	})
}
