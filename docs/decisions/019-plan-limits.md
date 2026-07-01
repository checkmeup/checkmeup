# ADR-019: Plan limits

**Date:** 2026-06-15  
**Status:** Accepted  
**Revised:** 2026-06-15 — limits loosened + prices reduced after competitor review
**Revised:** 2026-06-16 — plans renamed Hobbyist/Indie/Studio/Agency → Hobby/Solo/Startup/Enterprise
**Revised:** 2026-06-16 — Enterprise capped at 1000 monitors / 100 status pages, price raised to $99
**Revised:** 2026-06-18 — keyword monitoring ([EP-11](../stories/ep-11-keyword-monitoring.md)) gated to paid plans (Solo and above); Hobby excluded
**Revised:** 2026-06-21 — keyword monitoring gate removed; available on every plan, including Hobby
**Revised:** 2026-06-23 — domain expiry monitoring ([EP-29](../stories/ep-29-domain-expiry-monitoring.md)) added to the aggregate monitor count
**Revised:** 2026-06-25 — notification channel limits added (Hobby 5 / Solo 20 / Startup 50 / Enterprise 100)
**Revised:** 2026-07-01 — port monitoring ([EP-33](../stories/ep-33-port-monitoring.md)) added to the aggregate monitor count

---

## Context

Four pricing tiers: Hobby ($0) / Solo ($9) / Startup ($29) / Enterprise ($99). Limits must be enforced server-side at the API level; the UI reflects them with inline upgrade prompts.

Monitors are counted in aggregate across all types (cron + uptime + SSL + domain + port) because users care about "how many things am I watching", not the internal type split.

---

## Limits

| | Hobby | Solo | Startup | Enterprise |
|---|---|---|---|---|
| Total monitors (cron + uptime + SSL + domain + port) | 10 | 30 | 100 | 1000 |
| Status pages | 1 | 3 | 10 | 100 |
| Notification channels | 5 | 20 | 50 | 100 |
| Min uptime check interval | 5 min | 1 min | 1 min | 1 min |

Keyword monitoring (uptime) is available on every plan, including Hobby — not a plan limit.

### Competitor reference (Jun 2026)

| Product | Free monitors | $10–12/mo | $38–39/mo | Min interval (paid) |
|---|---|---|---|---|
| Healthchecks.io (cron only) | 20 | — | 100 ($20) | n/a |
| Cronitor (cron + uptime) | 5 | $2/monitor | $2/monitor | 30 sec |
| UptimeRobot (uptime only) | 50 | 10 ($10) | 100 ($38) | 60 sec |
| Better Stack (uptime) | 10 | pay-per-monitor | pay-per-monitor | 30 sec |
| **checkmeup (cron + uptime + SSL)** | **10** | **30 ($9)** | **100 ($29)** | **1 min** |

checkmeup bundles three monitor types in one product, which justifies slightly lower raw monitor counts than single-purpose tools. The revised limits bring the free tier and Startup tier in line with the market.

### 30-second interval — deferred

Competitors offer 30-second check intervals at the $10–38/mo tier. Supporting this requires changing `interval_mins INTEGER` → `interval_secs INTEGER` in the DB schema and updating the worker + all queries. Deferred to a future migration; current minimum is 1 minute on all paid plans.

---

## Enforcement

- **API**: returns `402 Payment Required` with `{"code": "plan_limit_reached", "message": "..."}` when a limit is exceeded
- **UI**: catches `402`, shows an inline upgrade prompt (not a page redirect)
- **Interval clamping**: if the requested interval is below the plan minimum, the API rejects with `402` and explains the minimum allowed
- **Downgrade**: monitors and pages already over the new limit are NOT deleted — they stay but the user cannot create new ones until under the limit

## Implementation

Limits are defined as Go constants in `internal/billing/plans.go`. Each create handler calls `billing.CheckMonitorLimit` / `billing.CheckStatusPageLimit` / `billing.CheckNotificationChannelLimit` / `billing.ClampInterval` before inserting. `GET /api/v1/billing` returns `notificationChannelCount` and `notificationChannelLimit` alongside the existing monitor and status page counts so the billing dashboard can display usage. Keyword monitoring (`uptime_monitors.keyword`) was previously gated the same way via `billing.CheckKeywordMonitoringAllowed` and a `keywordMonitoringEnabled` field on `GET /api/v1/billing` — both removed 2026-06-21; the field is now unconditionally available, validated only for length (1–500 chars), same as any other monitor field.
