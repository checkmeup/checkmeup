---
title: DoS / Overload Vulnerability Audit
type: reference
status: active
updated: 2026-07-12
tags: [security, dos, audit]
---

# DoS / Overload Vulnerabilities

Security audit findings — unbounded operations a user or attacker could abuse to overload the system.

No open findings as of 2026-07-12 — the manual-incident growth gap noted earlier the same day was resolved by time-based retention rather than a plan-count limit (see below).

The concrete, file/line-cited claims below are now re-checked mechanically by the `overload-audit` skill (`.claude/skills/overload-audit/`) instead of relying on the next person to notice drift by hand — run it periodically or after touching any of the cited files.

---

## Things that are fine

- All five check loops (cron overdue, uptime, SSL, domain, port) bounded via a 50-goroutine semaphore (`checkConcurrency`, `worker.go:33`) ✓ — fixed 2026-07-04, cron's `checkOverdue` was the last one still processing sequentially
- Incident list queries capped at `LIMIT 200` (`queries/monitors.sql:93`, `queries/uptime.sql:107`, `queries/port.sql:103`) ✓
- Manual (status-page) incident list queries also capped at `LIMIT 200` (`queries/incidents.sql`'s `ListStatusPageIncidents` and `ListActiveStatusPageIncidentsForPage`) ✓ — fixed 2026-07-12; the public status page's resolved-incidents section was already paginated (`LIMIT $2 OFFSET $3`), but the private dashboard list and the public page's active-incidents section had no cap until this pass
- `uptime_checks` pruned after 90 days by `DeleteOldUptimeChecks`, alongside the existing `cron_pings` cleanup (`worker.go:397-402`) ✓
- Resolved manual incidents pruned after 90 days by `DeleteOldStatusPageIncidents`, same daily cleanup pass as the checks above — uniform across every plan, no per-plan creation limit needed ✓ — added 2026-07-12; still-active incidents are exempt regardless of age, only resolved ones age out
- Active (non-resolved) manual incidents capped at 100 per org (`maxActiveIncidents`, `incidents.go`'s `checkActiveIncidentCap`, called from `CreateIncident`) ✓ — added 2026-07-12, closing the remaining gap the 90-day retention above doesn't cover on its own: an org that never resolves anything could otherwise grow `status_page_incidents` unboundedly, since only resolved rows age out. `409 too_many_active_incidents` on the 101st; resolving one frees a slot. Uniform across every plan, same as the retention window.
- Public status page rate-limited at 300/min by IP, same as its badge endpoints (`server.go:89`) ✓
- Every authenticated route carries a blanket 300 req/min-per-org limit on top of any tighter per-route limit (`server.go:128`) ✓
- Monitor creation is plan-limited (10–1000 per plan) ✓
- Cron pings and uptime check reads are paginated ✓
- Request body capped at 64 KB globally ✓
- Auth and ping endpoints are rate-limited ✓
- All queries use sqlc parameterized statements (no SQL injection) ✓
