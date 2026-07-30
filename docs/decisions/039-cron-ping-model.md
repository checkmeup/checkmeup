# ADR-039: Cron ping model — start-of-run ping and a `cron_runs` table

**Date:** 2026-07-30
**Status:** Accepted

---

## Context

Today a cron monitor has exactly one signal: `GET /ping/{token}` (EP-02 US-0202), a single instantaneous completion ping. That's enough to detect a job that never checks in (EP-02 US-0203's missed-ping/grace-period alert), but not:

- a job that started, is still running far longer than normal, and hasn't crashed or hung in a way that stops it from eventually pinging ([EP-34](../stories/ep-34-zombie-job-detection.md), "zombie job" detection)
- a job whose next scheduled run begins while the previous run is still going ([EP-35](../stories/ep-35-overlap-detection.md), overlap detection)

Both epics need a start-of-run signal distinct from the completion ping, and somewhere to hold "run in progress" state. [`docs/decisions/backlog.md`](backlog.md) flagged this as needed before EP-34 US-3401 starts, and asked that one decision cover both epics since they depend on the same signal.

The change must be purely additive: a monitor that never sends a start ping keeps today's single-ping behavior, unchanged, forever.

---

## Alternatives considered

**Endpoint shape:**

| Option | Additive | Ruled out because |
|---|---|---|
| Overload `GET /ping/{token}?state=start` | ❌ | The existing handler already treats arbitrary query params as user metadata (`buildPingMetadata` in `apps/api/internal/handler/ping.go`) and stores them verbatim. Reusing the query string as a protocol control channel makes the same mechanism mean two different things depending on context. |
| `GET /ping/{token}/start` | ✅ | **Chosen** — new path, existing completion endpoint untouched, still a bare `curl` URL like every other ping. |

**Where "run in progress" state lives:**

| Option | Supports history (US-3404, US-3503) | Concurrency safety | Ruled out because |
|---|---|---|---|
| New nullable column(s) on `cron_monitors` (e.g. `run_started_at`) | Only the current run — history requires a second table anyway | Read-then-write race: two near-simultaneous start pings can both observe `NULL` before either writes, silently missing an overlap | Ends up as duplicate state once the history table is added; the race's failure mode is a missed alert |
| New `cron_runs` table (one row per run) | ✅ native | Same theoretical race, but the failure mode is a false-positive double-overlap, not a missed one | **Chosen** |

The table option also matches an existing pattern in this codebase — `cron_pings` is already an append-only event table per ping, not a mutable pointer column on `cron_monitors`.

---

## Decision

**New endpoint:** `GET /ping/{token}/start` records a run start. `GET /ping/{token}` keeps meaning completion, exactly as it does today.

**New table `cron_runs`:**

```sql
CREATE TABLE cron_runs (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id   UUID        NOT NULL REFERENCES cron_monitors(id) ON DELETE CASCADE,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    overlap      BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_cron_runs_monitor_open ON cron_runs(monitor_id) WHERE completed_at IS NULL;
```

- `GET /ping/{token}/start` checks for an existing open row (`completed_at IS NULL`) for that monitor before inserting a new one. If one is found, the new row is marked `overlap = TRUE` (EP-35 US-3501) — each start ping is evaluated independently, so a third overlapping start is still detected even while two prior runs remain open.
- `GET /ping/{token}` (completion), in addition to its existing behavior, finds the latest open row for the monitor and stamps `completed_at`. If none exists — because the monitor never sends a start ping, or this completion has no matching start — it's a no-op; today's single-ping behavior is unaffected (EP-34 US-3401's acceptance criteria).
- `max_duration_mins` (EP-34 US-3402) is monitor-level config, not run state — a new nullable column on `cron_monitors`, left unset by default.
- The start ping does **not** touch `cron_monitors.last_ping_at` / `next_ping_at` / `status` — those remain owned entirely by the completion ping, keeping EP-02's missed-ping detection completely independent of this new signal.

Left open for EP-34/EP-35 implementation (not blocking this schema decision): whether a detected stuck run or overlap becomes its own `cron_incidents` row via a discriminator column (reusing the existing alert-count/notification-channel machinery in `apps/api/internal/worker/worker_cron.go`) or a separate incident model.

---

## Consequences

- EP-34 (zombie detection) and EP-35 (overlap detection) share one signal and one table instead of inventing separate state per epic.
- Duration history (EP-34 US-3404) and overlap history (EP-35 US-3503) fall out of `cron_runs` directly — no second history table needed later.
- Monitors that never call the start endpoint are provably unaffected: no new columns on `cron_monitors` change meaning, and the completion-ping no-op path is explicit.
- The `WHERE completed_at IS NULL` partial index keeps the open-run lookup cheap regardless of how large `cron_runs` grows; old completed rows should follow the same pruning pattern as `cron_pings` (`DeleteOldCronPings`) once retention is decided for the new table.
- A genuine two-start race can log a false-positive overlap; accepted as safer than the column-based alternative's failure mode (a missed overlap).
