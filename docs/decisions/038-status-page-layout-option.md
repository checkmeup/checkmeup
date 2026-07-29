# ADR-038: Per-status-page layout choice (classic vs. grid)

**Date:** 2026-07-29
**Status:** Accepted

---

## Context

The public status page (EP-06, [ADR-017](017-status-page-ssr.md)) has only ever had one layout: a single-column stack of monitor cards with active incidents and past-incident history inline above/below them (design: `CheckMeUp Status Page.dc.html`, option 1a).

A second layout was designed alongside it (option 1b): a wider two-column monitor grid with active incidents and past-incident history moved into a right-hand sidebar — better suited to pages with many monitors, where the single-column list gets long. Rather than replacing the original layout, orgs should be able to pick whichever fits their page.

## Decision

- Add `status_pages.layout` (`TEXT NOT NULL DEFAULT 'classic'`, `CHECK (layout IN ('classic', 'grid'))`) — a per-page setting, same pattern as `hide_branding` ([ADR-035](035-status-page-hide-branding.md)): an org can run multiple status pages and may want a compact grid for an internal, monitor-heavy page while keeping the classic single-column layout for a client-facing one.
- Not plan-gated — unlike `hide_branding`, layout choice is a structural/cosmetic preference, not a paid feature, so it's available on every plan.
- Configurable from the status page edit form only (not at creation), mirroring `hide_branding`'s placement — new pages default to `classic`, preserving today's behavior for every existing page unchanged.
- The Go `html/template` in `status_public.go` renders both layouts from the same template set: the monitor card, incident card, and past-incident row markup are each factored into a `{{define}}` block reused by both layouts, so only the surrounding wrapper markup and layout-specific CSS (`.content-grid`, `.sidebar`, `.sidebar-card`) differ. This keeps the two layouts from silently drifting apart on a shared field (e.g. a new monitor-note case).

## Consequences

- No plan-limit or billing-webhook interaction, unlike `hide_branding` — nothing to enforce on downgrade.
- `status_public.go`'s unauthenticated render path stays a single query against `status_pages` plus the existing monitor/incident queries — `layout` is just one more field read off the same row, no new query.
- The grid layout's monitor cards omit the 90-day bar's "90 days ago / Today" labels (design 1b) to fit two columns at a readable width; the classic layout is unaffected.
- Adding a third layout later means one more `{{if}}` branch plus a CHECK-constraint migration — the shared `{{define}}` blocks mean it wouldn't need to re-implement monitor/incident rendering from scratch.

## Alternatives considered

1. Org-wide default layout instead of per-page — rejected for the same reason `hide_branding` is per-page: a single org can run status pages with different audiences and monitor counts, each better served by a different layout.
2. A client-side (Vue/JS) layout switcher on top of a single server-rendered payload — rejected; the public page's whole point is to render without JS ([ADR-017](017-status-page-ssr.md)), so layout has to be a server-side branch, not a runtime toggle.
