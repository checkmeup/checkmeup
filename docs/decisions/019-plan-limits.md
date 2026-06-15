# ADR-019: Plan limits

**Date:** 2026-06-15  
**Status:** Accepted  
**Revised:** 2026-06-15 — limits loosened after competitor review

---

## Context

Four pricing tiers: Hobbyist ($0) / Indie ($12) / Studio ($39) / Agency ($99). Limits must be enforced server-side at the API level; the UI reflects them with inline upgrade prompts.

Monitors are counted in aggregate across all types (cron + uptime + SSL) because users care about "how many things am I watching", not the internal type split.

---

## Limits

| | Hobbyist | Indie | Studio | Agency |
|---|---|---|---|---|
| Total monitors (cron + uptime + SSL) | 10 | 30 | 100 | unlimited |
| Status pages | 1 | 3 | 10 | unlimited |
| Min uptime check interval | 5 min | 1 min | 1 min | 1 min |

### Competitor reference (Jun 2026)

| Product | Free monitors | $10–12/mo | $38–39/mo | Min interval (paid) |
|---|---|---|---|---|
| Healthchecks.io (cron only) | 20 | — | 100 ($20) | n/a |
| Cronitor (cron + uptime) | 5 | $2/monitor | $2/monitor | 30 sec |
| UptimeRobot (uptime only) | 50 | 10 ($10) | 100 ($38) | 60 sec |
| Better Stack (uptime) | 10 | pay-per-monitor | pay-per-monitor | 30 sec |
| **checkmeup (cron + uptime + SSL)** | **10** | **30 ($12)** | **100 ($39)** | **1 min** |

checkmeup bundles three monitor types in one product, which justifies slightly lower raw monitor counts than single-purpose tools. The revised limits bring the free tier and Studio tier in line with the market.

### 30-second interval — deferred

Competitors offer 30-second check intervals at the $10–38/mo tier. Supporting this requires changing `interval_mins INTEGER` → `interval_secs INTEGER` in the DB schema and updating the worker + all queries. Deferred to a future migration; current minimum is 1 minute on all paid plans.

---

## Enforcement

- **API**: returns `402 Payment Required` with `{"code": "plan_limit_reached", "message": "..."}` when a limit is exceeded
- **UI**: catches `402`, shows an inline upgrade prompt (not a page redirect)
- **Interval clamping**: if the requested interval is below the plan minimum, the API rejects with `402` and explains the minimum allowed
- **Downgrade**: monitors and pages already over the new limit are NOT deleted — they stay but the user cannot create new ones until under the limit

## Implementation

Limits are defined as Go constants in `internal/billing/plans.go`. Each create handler calls `billing.CheckMonitorLimit` / `billing.CheckStatusPageLimit` / `billing.ClampInterval` before inserting.
