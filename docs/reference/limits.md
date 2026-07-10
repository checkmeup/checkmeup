---
title: DoS / Overload Vulnerability Audit
type: reference
status: active
updated: 2026-07-10
tags: [security, dos, audit]
---

# DoS / Overload Vulnerabilities

Security audit findings — unbounded operations a user or attacker could abuse to overload the system.

No open findings as of 2026-07-04 — the four below were all confirmed fixed in code while re-auditing this doc (it had drifted stale; each had already been patched in an earlier commit without this file being updated).

The concrete, file/line-cited claims below are now re-checked mechanically by the `overload-audit` skill (`.claude/skills/overload-audit/`) instead of relying on the next person to notice drift by hand — run it periodically or after touching any of the cited files.

---

## Things that are fine

- All five check loops (cron overdue, uptime, SSL, domain, port) bounded via a 50-goroutine semaphore (`checkConcurrency`, `worker.go:33`) ✓ — fixed 2026-07-04, cron's `checkOverdue` was the last one still processing sequentially
- Incident list queries capped at `LIMIT 200` (`queries/monitors.sql:93`, `queries/uptime.sql:107`) ✓
- `uptime_checks` pruned after 90 days by `DeleteOldUptimeChecks`, alongside the existing `cron_pings` cleanup (`worker.go:397-402`) ✓
- Public status page rate-limited at 300/min by IP, same as its badge endpoints (`server.go:89`) ✓
- Every authenticated route carries a blanket 300 req/min-per-org limit on top of any tighter per-route limit (`server.go:128`) ✓
- Monitor creation is plan-limited (10–1000 per plan) ✓
- Cron pings and uptime check reads are paginated ✓
- Request body capped at 64 KB globally ✓
- Auth and ping endpoints are rate-limited ✓
- All queries use sqlc parameterized statements (no SQL injection) ✓
