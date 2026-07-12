# ADR-015: cron_pings 30-day rolling retention

**Status:** Accepted  
**Date:** 2026-06-15  
**Revised:** 2026-07-12 — extended the same daily-prune mechanism to resolved manual incidents ([EP-24](../stories/ep-24-incident-management.md)'s `status_page_incidents`, not the automatic `cron_incidents`/`uptime_incidents`/`port_incidents` this ADR's Context section discusses): 90-day retention, uniform across every plan, in place of a per-plan creation limit — see Consequences
**Revised:** 2026-07-12 — added a flat 100-active-incident cap per org, closing the gap the retention change alone left open (an org that never resolves anything could still grow the table unboundedly) — see Consequences. Three sibling caps on other unbounded-creation resources (incident updates, maintenance windows, API keys) followed in the same pass — see [ADR-036](036-flat-safety-caps.md).

## Context

`cron_pings` has no TTL. A monitor pinging every minute generates ~43,200 rows/month. At Enterprise plan (300 monitors, 1-min interval) that is ~13M rows/month per customer — unbounded growth on an 80 GB SSD.

Options considered:

- **Keep last N rows per monitor** — count-based; 1,000 pings means 16 hours for minute-interval jobs but 2.7 years for daily jobs. Inconsistent user experience.
- **Keep last 30 days** — time-based; predictable and explainable to users. Worst case at Enterprise: 300 × 43,200 = 13M rows ≈ 1.3 GB/month, well within SSD capacity.
- **Keep last 90 days** — more history, 3× the storage.

## Decision

Delete `cron_pings` rows older than **30 days**, run once per day by the existing background worker.

A composite index `(monitor_id, received_at DESC)` is added to make the daily DELETE efficient and to speed up paginated ping queries in the detail view (which currently relies only on the `monitor_id` index and sorts in application memory).

Incidents are stored in a separate `cron_incidents` table with no FK to `cron_pings` — pruning pings does not affect incident history.

## Consequences

- Ping history visible in the UI is capped at 30 days. Sufficient for operational debugging; nobody audits 60-day-old ping logs.
- Storage worst case (Enterprise, 300 monitors, 1-min interval): ~1.3 GB/month, stable after the first 30 days.
- Cleanup runs in the existing goroutine worker — no new process, no Redis, no cron daemon (consistent with ADR-001).
- The daily ticker is independent of the 30-second missed-ping ticker; both run in the same `worker.Run` loop.
- Manual incidents (`status_page_incidents`) grow by human action, not per-check, so 90 days of history is generous rather than tight — chosen to match the existing `uptime_checks`/`port_checks` window rather than invent a new number. Only **resolved** incidents age out (`status = 'resolved' AND resolved_at < NOW() - INTERVAL '90 days'`); a still-active incident is exempt regardless of how old it is, so it can never silently disappear from a status page while genuinely ongoing. No per-plan differentiation — every plan gets the same 90 days, closing the [`docs/reference/limits.md`](../reference/limits.md) open finding about unbounded incident creation without needing a `billing.Check*Limit`-style count cap.
- Time-based retention alone doesn't bound an org that just never resolves anything — those incidents are permanently exempt from the 90-day prune by design. `CreateIncident` therefore also rejects declaring a new incident once an org already has 100 non-resolved ones (`maxActiveIncidents`, `409 too_many_active_incidents`) — a flat safety cap, not a plan limit, since resolving one (not upgrading) is the only way past it. Same uniform-across-every-plan approach as the retention window.
