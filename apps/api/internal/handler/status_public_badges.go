package handler

import (
	"fmt"
	"html"
	"net/http"
	"strings"
)

// badgeStatusWord maps a status color to the short badge vocabulary
// (US-3001: "operational" / "degraded" / "outage"). Gray covers paused,
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

func writeBadge(w http.ResponseWriter, label, value, color string) {
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "max-age=60") // US-3004: short TTL so embedders/CDNs don't hammer this on every page load
	_, _ = w.Write([]byte(renderBadgeSVG(label, value, color)))
}

// renderBadgeSVG draws a flat, shields.io-style two-segment badge: a gray
// label segment and a colored value segment. Built by hand (no badge
// library/external service — ADR's no-new-infra principle) using a rough
// per-character width estimate, which only needs to avoid clipping text,
// not be pixel-perfect.
func renderBadgeSVG(label, value, color string) string {
	labelW := badgeTextWidth(label)
	valueW := badgeTextWidth(value)
	totalW := labelW + valueW

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s">
<linearGradient id="s" x2="0" y2="100%%">
<stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
<stop offset="1" stop-opacity=".1"/>
</linearGradient>
<clipPath id="r"><rect width="%d" height="20" rx="3" fill="#fff"/></clipPath>
<g clip-path="url(#r)">
<rect width="%d" height="20" fill="#555"/>
<rect x="%d" width="%d" height="20" fill="%s"/>
<rect width="%d" height="20" fill="url(#s)"/>
</g>
<g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,sans-serif" font-size="11">
<text x="%d" y="14">%s</text>
<text x="%d" y="14">%s</text>
</g>
</svg>`,
		totalW, html.EscapeString(label), html.EscapeString(value),
		totalW,
		labelW,
		labelW, valueW, color,
		totalW,
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
