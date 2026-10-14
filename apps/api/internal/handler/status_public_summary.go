package handler

import (
	"encoding/json"
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

type summaryDay struct {
	Date string `json:"date"`
	Tone string `json:"tone"`
}

// toneRank orders tones by how bad they are, so a day takes its worst
// monitor's tone.
var toneRank = map[string]int{"none": 0, "up": 1, "paused": 2, "warn": 3, "down": 4}

// summarizeDays folds every monitor's 90-day bar into one, oldest day first:
// each day shows the worst tone any monitor had that day.
func summarizeDays(rows []publicMonitorRow) []summaryDay {
	days := make([]summaryDay, 0, 90)
	for i := range 90 {
		day := summaryDay{Tone: "none"}
		for _, row := range rows {
			if i >= len(row.Bar) {
				continue
			}
			if day.Date == "" {
				day.Date = row.Bar[i].Label
			}
			if tone := statusTone(row.Bar[i].Color); toneRank[tone] > toneRank[day.Tone] {
				day.Tone = tone
			}
		}
		days = append(days, day)
	}
	return days
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
