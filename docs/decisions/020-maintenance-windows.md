# ADR-020: Maintenance windows suppress checks via due-query exclusion

**Status:** Accepted
**Date:** 2026-06-16

## Context

The landing page and blog already promised "scheduled regular or unplanned maintenance — no alerts sent during maintenance windows" before any implementation existed. We needed a way for users to schedule a window (or start one ad-hoc) covering one or more monitors where, for the duration of the window: no alerts fire, no incidents are recorded, and the monitor's uptime stats stay unaffected.

Three implementation shapes were considered:

1. **Teach the worker about maintenance** — keep running checks as usual, but add an `if in maintenance` branch in `checkOverdue`/`checkOneUptimeMonitor`/`checkOneSSLMonitor` (`apps/api/internal/worker/worker.go`) to skip incident creation and alert sending.
2. **Exclude from the due/overdue queries** — make monitors with an active maintenance window invisible to `ListOverdueCronMonitors`, `ListDueUptimeMonitors`, and `ListDueSSLMonitors`, the same queries that already exclude `status = 'paused'` monitors.
3. **Recurring-schedule engine** — a full RRULE-style recurrence system with a "next occurrence" calculator.

## Decision

Went with **option 2**. Each of the three due/overdue queries gained an `AND NOT EXISTS (...)` clause checking `maintenance_window_monitors` joined to `maintenance_windows` for a row where `starts_at <= NOW() AND (ends_at IS NULL OR ends_at > NOW())`.

This means **zero changes to `worker.go`'s alert-decision logic** — a monitor under active maintenance simply never appears in the list the worker iterates over this tick, exactly like a paused monitor today. No check runs, no incident is created, no alert is sent, and the 90-day bar / uptime % stay clean automatically — there is no separate "exclude maintenance days from uptime" step needed.

Windows are one-off (single `starts_at` / nullable `ends_at`, where `NULL` means open-ended and is closed manually via an "end now" action) — recurring schedules (option 3) were explicitly deferred as a possible fast-follow, since they require a rule engine and a way to project upcoming occurrences that wasn't justified for v1.

A window can cover multiple monitors of any type (cron/uptime/ssl) via a polymorphic join table, `maintenance_window_monitors` — the same pattern as `status_page_monitors` (ADR-005's sibling table).

## Consequences

- New tables: `maintenance_windows` (org-scoped) and `maintenance_window_monitors` (polymorphic, no FK on `monitor_id` — same trade-off as `status_page_monitors`).
- The public status page (`status_public.go`) independently queries active windows per org and overrides the displayed status to "Under maintenance", reusing the existing gray `--status-paused` token rather than introducing a new color — the design system already documents that token as "Paused / maintenance" (`docs/design.md`). This keeps `computeOverallStatus` working unmodified, since it only escalates on red/amber.
- A maintenance window that starts while a monitor is already mid-incident freezes that incident in place (no new checks run to resolve or re-flag it) until the window ends, at which point normal checking resumes. For cron monitors specifically, if `next_ping_at` is still in the past when the window ends, the monitor can be marked down again immediately on the next tick — a known, accepted edge case for v1.
- If recurring maintenance is added later, it only needs to compute concrete `starts_at`/`ends_at` instances ahead of time and insert/extend rows in these same two tables — the worker-exclusion mechanism does not need to change.
