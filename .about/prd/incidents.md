# Manual incidents on status pages

## Problem

Automatic up/down status can't say "we're seeing slow checkouts, investigating"; users need to narrate an incident to their visitors and post progress.

## Users

Users who run public status pages.

## Success signal

A user declares an incident and posts updates, and visitors see each update on the status page at once.

## Scope

- Declaring an incident with a title, a first message, a severity (minor, major, critical) and affected monitors of any type.
- Append-only updates moving its status: investigating, identified, monitoring, resolved.
- Active incidents and their updates on the public status page, escalating the banner for major and critical.
- Editing titles and update messages; deleting an incident.
- A warning when an affected monitor is under maintenance.

## Non-goals

- Alerts on any channel for declared incidents or their updates.
- Email or SMS updates for status-page visitors.

## Acceptance criteria

- Declaring needs a title, a message and at least one monitor; the incident starts as "investigating" with that message as its first update and doesn't change any monitor's status.
  → test: apps/web/src/views/IncidentCreateView.test.ts "shows a validation error when no monitors are selected"
  → test: apps/api/internal/handler/incidents_test.go "TestCreateIncident/success resolves monitor names, dedups, and seeds the first update"
- Another org's monitor can't be attached.
  → test: apps/api/internal/handler/incidents_test.go "TestCreateIncident/a monitor belonging to another org is rejected (tenant isolation)"
- Declaring on a monitor under active maintenance warns first and goes ahead only when confirmed.
  → test: apps/api/internal/handler/incidents_test.go "TestCreateIncident/PRD incidents: declaring against a monitor under active maintenance warns unless confirmed"
  → test: apps/web/src/views/IncidentCreateView.test.ts "shows a maintenance-overlap warning and declares anyway on confirm"
- Each update appends a message and moves the status on, through to resolved; the form offers the next status by default and disappears once resolved.
  → test: apps/api/internal/handler/incidents_test.go "TestPostIncidentUpdate/success appends an update and advances status through to resolved"
  → test: apps/web/src/views/IncidentDetailView.test.ts "posts an update defaulting to the next status in sequence"
  → test: apps/web/src/views/IncidentDetailView.test.ts "hides the post-update form once the incident is resolved"
- The public page shows each active incident with all its updates, newest first.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeHTTP/active incidents render their full update timeline, newest first, batched across incidents"
- An active major incident raises the banner to "Partial outage" and a critical one to "Major outage", even with every monitor green; a minor one doesn't.
  → test: apps/api/internal/handler/status_public_test.go "TestComputeOverallStatus"
- The title and any update's message can be edited; deleting removes the incident and its updates.
  → test: apps/api/internal/handler/incidents_test.go "TestUpdateIncidentTitle/success renames and trims whitespace"
  → test: apps/api/internal/handler/incidents_test.go "TestUpdateIncidentUpdateMessage/success edits the message in place"
  → test: apps/api/internal/handler/incidents_test.go "TestDeleteIncident/cross-tenant delete is a no-op, owner delete succeeds and cascades"
- A 101st active incident or a 101st update on one incident is refused (ADR limits).
  → test: apps/api/internal/handler/incidents_test.go "TestCreateIncident/rejects declaring a 101st active incident, and resolving one frees a slot"
  → test: apps/api/internal/handler/incidents_test.go "TestPostIncidentUpdate/rejects a 101st update on the same incident"
- Resolved incidents are deleted 90 days after they resolve; active ones never are.
  → test: apps/api/internal/worker/worker_test.go "TestPruneOldStatusPageIncidents"
- Resolved incidents appear in a paginated history on the public page.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeHTTP/resolved incidents appear in a paginated history"
- When maintenance and an incident cover the same monitor, the public page shows both as distinct states.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeHTTP/maintenance and an incident on the same monitor show as distinct states"
