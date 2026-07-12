---
title: DoS / Overload Vulnerability Audit
type: reference
status: active
updated: 2026-07-12
tags: [security, dos, audit]
---

# DoS / Overload Vulnerabilities

Security audit findings — unbounded operations a user or attacker could abuse to overload the system.

The concrete, file/line-cited claims below are now re-checked mechanically by the `overload-audit` skill (`.claude/skills/overload-audit/`) instead of relying on the next person to notice drift by hand — run it periodically or after touching any of the cited files.

---

## Open findings

- **No plan limit on manual incident creation** ([EP-24](../stories/ep-24-incident-management.md), `internal/handler/incidents.go`'s `CreateIncident`) — every other creatable resource in this codebase (monitors, status pages, notification channels) is plan-limited via a `billing.Check*Limit` call; incident creation has none. Bounded only by the blanket 300 req/min-per-org rate limit (`server.go:128`), which still permits unbounded growth over time. Lower practical risk than the check/ping volume that motivated [ADR-015](../decisions/015-cron-pings-retention.md) (incidents are human-declared, not generated per check), but not actually capped. Deferred — needs a product decision on what the per-plan number should be, not just a mechanical fix, unlike the two query caps below.

## Things that are fine

- All five check loops (cron overdue, uptime, SSL, domain, port) bounded via a 50-goroutine semaphore (`checkConcurrency`, `worker.go:33`) ✓ — fixed 2026-07-04, cron's `checkOverdue` was the last one still processing sequentially
- Incident list queries capped at `LIMIT 200` (`queries/monitors.sql:93`, `queries/uptime.sql:107`, `queries/port.sql:103`) ✓
- Manual (status-page) incident list queries also capped at `LIMIT 200` (`queries/incidents.sql`'s `ListStatusPageIncidents` and `ListActiveStatusPageIncidentsForPage`) ✓ — fixed 2026-07-12; the public status page's resolved-incidents section was already paginated (`LIMIT $2 OFFSET $3`), but the private dashboard list and the public page's active-incidents section had no cap until this pass
- `uptime_checks` pruned after 90 days by `DeleteOldUptimeChecks`, alongside the existing `cron_pings` cleanup (`worker.go:397-402`) ✓
- Public status page rate-limited at 300/min by IP, same as its badge endpoints (`server.go:89`) ✓
- Every authenticated route carries a blanket 300 req/min-per-org limit on top of any tighter per-route limit (`server.go:128`) ✓
- Monitor creation is plan-limited (10–1000 per plan) ✓
- Cron pings and uptime check reads are paginated ✓
- Request body capped at 64 KB globally ✓
- Auth and ping endpoints are rate-limited ✓
- All queries use sqlc parameterized statements (no SQL injection) ✓
