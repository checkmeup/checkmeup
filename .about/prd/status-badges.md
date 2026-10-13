# Status badges

## Problem

Users want to show their status in a README, footer or docs without visitors clicking through to the status page; a badge also advertises Checkmeup.

## Users

Status-page owners embedding their status elsewhere.

## Success signal

A user copies a badge snippet into a README and it shows the live status, linking back to the page.

## Scope

- An SVG badge per status page and per monitor on that page, rendered by the server.
- A "Badges" section in the status page editor with previews and Markdown and HTML snippets.
- Short caching and per-IP rate limiting (ADR limits).

## Non-goals

- Badge styles or colors chosen by the user.
- Badges for monitors not on a status page.

## Acceptance criteria

- The page badge says "operational", "degraded" or "outage", matching the page banner.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServePageBadge/all-green page renders an operational badge"
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServePageBadge/a down monitor renders an outage badge"
  → test: apps/api/internal/handler/status_public_test.go "TestBadgeStatusWord"
- A monitor badge shows that monitor's status, but only for a monitor on that page; otherwise, or for an unknown slug, it's `404`.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeMonitorBadge/attached monitor renders its own status badge"
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeMonitorBadge/monitor not attached to the page returns 404 (no leaking status outside the page)"
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServePageBadge/unknown slug returns 404"
- Text in a badge is escaped so it can't break the SVG.
  → test: apps/api/internal/handler/status_public_test.go "TestRenderBadgeSVG/escapes text that could break the SVG markup"
- The editor lists the page badge and one per attached monitor, and copies Markdown or HTML (an image linking to the page) in one click.
  → test: apps/web/src/views/StatusPageDetailView.test.ts "renders a page-level badge and one per attached monitor"
  → test: apps/web/src/views/StatusPageDetailView.test.ts "copies the Markdown snippet for a badge"
  → test: apps/web/src/views/StatusPageDetailView.test.ts "copies the HTML snippet for a monitor badge"
- Badge responses carry `Cache-Control: max-age=60`.
- A 301st badge request in a minute from one IP gets `429`.
