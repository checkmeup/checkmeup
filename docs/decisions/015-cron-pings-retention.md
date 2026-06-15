# ADR-015: cron_pings 30-day rolling retention

**Status:** Accepted  
**Date:** 2026-06-15

## Context

`cron_pings` has no TTL. A monitor pinging every minute generates ~43,200 rows/month. At Agency plan (300 monitors, 1-min interval) that is ~13M rows/month per customer — unbounded growth on an 80 GB SSD.

Options considered:
- **Keep last N rows per monitor** — count-based; 1,000 pings means 16 hours for minute-interval jobs but 2.7 years for daily jobs. Inconsistent user experience.
- **Keep last 30 days** — time-based; predictable and explainable to users. Worst case at Agency: 300 × 43,200 = 13M rows ≈ 1.3 GB/month, well within SSD capacity.
- **Keep last 90 days** — more history, 3× the storage.

## Decision

Delete `cron_pings` rows older than **30 days**, run once per day by the existing background worker.

A composite index `(monitor_id, received_at DESC)` is added to make the daily DELETE efficient and to speed up paginated ping queries in the detail view (which currently relies only on the `monitor_id` index and sorts in application memory).

Incidents are stored in a separate `cron_incidents` table with no FK to `cron_pings` — pruning pings does not affect incident history.

## Consequences

- Ping history visible in the UI is capped at 30 days. Sufficient for operational debugging; nobody audits 60-day-old ping logs.
- Storage worst case (Agency, 300 monitors, 1-min interval): ~1.3 GB/month, stable after the first 30 days.
- Cleanup runs in the existing goroutine worker — no new process, no Redis, no cron daemon (consistent with ADR-001).
- The daily ticker is independent of the 30-second missed-ping ticker; both run in the same `worker.Run` loop.
