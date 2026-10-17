# Public status pages

## Problem

When something breaks, clients and users ask "is it down for everyone?"; a public page answers before they ask, and makes a freelancer look professional to their clients.

## Users

Freelancers and teams showing their own or their clients' service status to the public.

## Success signal

A user publishes a page with their monitors in a few minutes and can send clients its link.

## Scope

- A page at `checkmeup.net/status/<slug>` with a title, description, logo URL and chosen monitors with display names and order.
- The public page: overall banner, per-monitor status and 90-day bar, active manual incidents (PRD incidents), maintenance shown as such, last-updated time, working without JavaScript.
- Per-page layout (classic or grid) and, on paid plans, hiding the Checkmeup branding (ADR status-pages).
- Deleting a page.

## Non-goals

- Custom domains or subdomains (ADR status-pages).
- Email subscriptions for visitors.

## Acceptance criteria

- A slug must be 3–48 lowercase letters, digits or hyphens and unused; availability shows while typing.
  → test: apps/api/internal/handler/status_pages_test.go "TestCheckSlug/invalid slug format reports unavailable with a reason"
  → test: apps/api/internal/handler/status_pages_test.go "TestCheckSlug/a taken slug is reported unavailable"
  → test: apps/web/src/views/StatusPageCreateView.test.ts "shows the taken reason when the slug is unavailable"
- Creating a page returns its public URL; a taken slug gets `409`.
  → test: apps/api/internal/handler/status_pages_test.go "TestCreateStatusPage/success returns a public URL ending in the slug"
  → test: apps/api/internal/handler/status_pages_test.go "TestCreateStatusPage/duplicate slug returns 409"
- Monitors of any type can be added, given display names, reordered and removed; another org's monitor is refused.
  → test: apps/api/internal/handler/status_pages_test.go "TestSetStatusPageMonitors/success attaches multiple monitor types and fully replaces on a second call"
  → test: apps/api/internal/handler/status_pages_test.go "TestSetStatusPageMonitors/SECURITY: rejects a monitor belonging to a different org (tenant isolation)"
  → test: apps/web/src/views/StatusPageDetailView.test.ts "reorders selected monitors with the move up/down controls"
- An empty display name shows the monitor's own name.
  → test: apps/api/internal/handler/status_pages_test.go "TestSetStatusPageMonitors/an empty display name falls back to the monitor's real name, not its UUID"
- The public page needs no login, shows each monitor's status and a 90-day bar (red for days with downtime, today last) and an overall banner, and works without JavaScript.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeHTTP/renders monitor statuses and aggregates overall status"
  → test: apps/api/internal/handler/status_public_test.go "TestBuild90DayBar/has data: down days are red, others green, today is the last segment"
  → test: apps/api/internal/handler/status_public_test.go "TestComputeOverallStatus"
- A monitor under an active maintenance window shows "Under maintenance" instead of its real status.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeHTTP/an active maintenance window overrides the monitor's real status"
- An unknown or deleted slug returns `404`.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeHTTP/unknown slug returns 404"
- The grid layout puts monitors in a grid with incidents in a sidebar; it can be picked in the edit form.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeHTTP/ADR status-pages: grid layout renders the sidebar/content-grid wrapper with monitors and incidents"
  → test: apps/web/src/views/StatusPageEditView.test.ts "switches to the grid layout and submits it"
- Unless branding is hidden, the footer reads "Powered by Checkmeup" and links "Get a free status page like this" to checkmeup.net, UTM-tagged (`utm_source=status_page`) so GA4 counts the visits and sign-ups it brings.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeHTTP/a page with no monitors renders the empty state"
- On a paid plan, hiding branding removes the "Powered by Checkmeup" footer and its links; on Hobby the option is disabled with an upgrade hint and the API refuses it.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeHTTP/ADR status-pages: hide_branding suppresses the footer's branding and FAQ/Terms/Privacy links"
  → test: apps/api/internal/handler/status_pages_test.go "TestUpdateStatusPage/ADR status-pages: hideBranding requires a paid plan"
  → test: apps/web/src/views/StatusPageEditView.test.ts "disables the hide-branding checkbox on Hobby and shows an upgrade hint"
- Deleting a page asks for confirmation and leaves its monitors in place.
  → test: apps/web/src/views/StatusPageDetailView.test.ts "deletes the page and navigates back to the list on confirm"
  → test: apps/api/internal/handler/status_pages_test.go "TestDeleteStatusPage/cross-tenant delete is a no-op, owner delete succeeds"
- Another org's page can't be read or changed.
  → test: apps/api/internal/handler/status_pages_test.go "TestGetStatusPage/cross-tenant page is not found (tenant isolation)"
