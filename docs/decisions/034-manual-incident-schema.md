# ADR-034: Manual incidents live in a new `status_page_incidents` table

**Date:** 2026-07-12
**Status:** Accepted

---

## Context

`cron_incidents` and `uptime_incidents` are created and resolved entirely automatically by the worker from up/down transitions ([ADR-016](016-alert-debounce.md)) — each row is 1:1 with a single monitor's check-failure window (`monitor_id`, `started_at`, `resolved_at`). [EP-24](../stories/ep-24-incident-management.md) adds manually-declared incidents: a user narrating something affecting visitors (e.g. degraded performance) that isn't a hard monitor-down, shown on the public status page ([EP-06](../stories/ep-06-status-page.md)) alongside the existing automatic state.

The open question (tracked in [decision backlog](backlog.md)): should manual incidents extend the existing `cron_incidents`/`uptime_incidents` tables, or live in a new table decoupled from monitors and monitor-type?

---

## Alternatives considered

| Option | Spans multiple monitors of different types | Reuses existing lifecycle columns | Ruled out because |
|---|---|---|---|
| Extend `cron_incidents`/`uptime_incidents` | ❌ (each table is 1:1 with one monitor of one type) | Partial — `started_at`/`resolved_at` fit, but nothing else does | Would need a nullable `monitor_id`, a way to link one incident to several monitors across two tables, and new columns (title, message, severity, status, append-only updates) bolted onto tables whose existing rows are purely automatic. Mixing automatic and manual rows in the same table also complicates every query/worker path that currently assumes "row exists = worker-detected transition." |
| New `status_page_incidents` table | ✅ (via a join table to monitors) | N/A — new lifecycle designed for the manual case | **Chosen** |

---

## Decision

Manual incidents get their own table, independent of `cron_incidents`/`uptime_incidents` and of monitor type:

- `status_page_incidents` — `id`, `org_id`, `title`, `severity` (`minor` / `major` / `critical`), `status` (`investigating` / `identified` / `monitoring` / `resolved`), `created_at`, `resolved_at`
- `status_page_incident_monitors` — join table (`incident_id`, `monitor_id`, `monitor_type`) linking one incident to any number of monitors across cron/uptime/SSL/domain/port types (US-2401)
- `status_page_incident_updates` — append-only timestamped entries (`incident_id`, `message`, `created_at`) driving the reverse-chronological update feed (US-2402)

This table is entirely separate from the automatic `cron_incidents`/`uptime_incidents` worker-driven rows — declaring, updating, or resolving a manual incident never touches monitor `status` or the automatic incident tables (US-2401's independence requirement). The status page merges both signals when rendering the overall banner (US-2403): automatic up/down state and active manual incidents each can escalate it independently.

---

## Consequences

- Enables a manual incident to span multiple monitors of different types in one record, matching how EP-24's acceptance criteria are written (US-2401, US-2405).
- Keeps the existing `cron_incidents`/`uptime_incidents` tables and worker logic untouched — no risk to the automatic alerting path ([ADR-016](016-alert-debounce.md)) from this change.
- Status page rendering (US-2403) now reads two independent incident sources instead of one; the "overall status" banner logic has to merge them rather than reflect a single table's state.
- Unblocks [EP-24](../stories/ep-24-incident-management.md) — the schema question was the only thing gating US-2401.
