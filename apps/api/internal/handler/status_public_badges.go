package handler

import (
	"fmt"
	"html"
	"net/http"
	"strings"
)

// badgeStatusWord maps a status color to the short badge vocabulary
// (PRD status-badges: "operational" / "degraded" / "outage"). Gray covers paused,
// under-maintenance, and not-yet-checked monitors — none of those are an
// outage, but "operational" would overstate it, so badges call it "unknown".
func badgeStatusWord(color string) string {
	switch color {
	case statusColorGreen:
		return "operational"
	case statusColorAmber:
		return "degraded"
	case statusColorRed:
		return "outage"
	default:
		return "unknown"
	}
}

// The badge palette is the light theme's status colors (docs/reference/design.md):
// each carries white text at AA contrast, and a badge sits on whatever page
// embeds it, so it can't switch with a theme. The label segment is the light
// theme's dim text color.
const (
	badgeLabelFill    = "#59625d"
	badgeFillUp       = "#12805f"
	badgeFillDegraded = "#a86200"
	badgeFillDown     = "#c93a2e"
	badgeFillUnknown  = "#6f7873"
)

// badgeFill maps a status color to the fill of the badge's value segment.
func badgeFill(color string) string {
	switch color {
	case statusColorGreen:
		return badgeFillUp
	case statusColorAmber:
		return badgeFillDegraded
	case statusColorRed:
		return badgeFillDown
	default:
		return badgeFillUnknown
	}
}

func writeBadge(w http.ResponseWriter, label, value, color string) {
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "max-age=60") // PRD status-badges: short TTL so embedders/CDNs don't hammer this on every page load
	_, _ = w.Write([]byte(renderBadgeSVG(label, value, color)))
}

// renderBadgeSVG draws a flat, shields.io-style two-segment badge: a gray
// label segment and a colored value segment, without the glossy gradient.
// Built by hand (no badge library/external service — ADR's no-new-infra
// principle) using a rough
// per-character width estimate, which only needs to avoid clipping text,
// not be pixel-perfect.
func renderBadgeSVG(label, value, color string) string {
	labelW := badgeTextWidth(label)
	valueW := badgeTextWidth(value)
	totalW := labelW + valueW

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s">
<clipPath id="r"><rect width="%d" height="20" rx="3" fill="#fff"/></clipPath>
<g clip-path="url(#r)">
<rect width="%d" height="20" fill="%s"/>
<rect x="%d" width="%d" height="20" fill="%s"/>
</g>
<g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,sans-serif" font-size="11">
<text x="%d" y="14">%s</text>
<text x="%d" y="14">%s</text>
</g>
</svg>`,
		totalW, html.EscapeString(label), html.EscapeString(value),
		totalW,
		labelW, badgeLabelFill,
		labelW, valueW, badgeFill(color),
		labelW/2, html.EscapeString(label),
		labelW+valueW/2, html.EscapeString(value),
	)
}

// badgeTextWidth estimates a string's rendered width at 11px Verdana in
// pixels: narrow characters get a smaller allowance, wide ones more, the
// rest an average — plus 10px padding on each side, mirroring the badge
// services this format is modeled on.
func badgeTextWidth(s string) int {
	width := 0
	for _, r := range s {
		switch {
		case strings.ContainsRune("iIl.:,;'!|", r):
			width += 4
		case strings.ContainsRune("mwMW", r):
			width += 10
		case r == ' ':
			width += 4
		default:
			width += 7
		}
	}
	return width + 20
}
