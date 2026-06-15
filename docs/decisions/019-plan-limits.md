# ADR-019: Plan limits

**Date:** 2026-06-15  
**Status:** Accepted

---

## Context

Four pricing tiers: Hobbyist ($0) / Indie ($12) / Studio ($39) / Agency ($99). Limits must be enforced server-side at the API level; the UI reflects them with inline upgrade prompts.

Monitors are counted in aggregate across all types (cron + uptime + SSL) because users care about "how many things am I watching", not the internal type split.

---

## Limits

| | Hobbyist | Indie | Studio | Agency |
|---|---|---|---|---|
| Total monitors (cron + uptime + SSL) | 5 | 20 | 50 | unlimited |
| Status pages | 1 | 3 | 10 | unlimited |
| Min uptime check interval | 10 min | 5 min | 1 min | 1 min |

## Enforcement

- **API**: returns `402 Payment Required` with `{"code": "plan_limit_reached", "message": "..."}` when a limit is exceeded
- **UI**: catches `402`, shows an inline upgrade prompt (not a page redirect)
- **Interval clamping**: if the requested interval is below the plan minimum, the API rejects with `402` and explains the minimum allowed
- **Downgrade**: monitrs and pages already over the new limit are NOT deleted — they stay but the user cannot create new ones until under the limit

## Implementation

Limits are defined as Go constants in `internal/billing/plans.go`. Each create handler calls `billing.CheckMonitorLimit` / `billing.CheckStatusPageLimit` / `billing.ClampInterval` before inserting.
