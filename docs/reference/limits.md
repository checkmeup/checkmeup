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

- All six check loops (cron overdue, uptime, SSL, domain, port, dns) bounded via a 50-goroutine semaphore (`checkConcurrency`, `worker.go:33`) ✓ — fixed 2026-07-04, cron's `checkOverdue` was the last one still processing sequentially; dns added 2026-07-29 (EP-39) following the same pattern
- Incident list queries capped at `LIMIT 200` (`queries/monitors.sql:93`, `queries/uptime.sql:107`, `queries/port.sql:103`, `queries/dns.sql`'s `ListDNSIncidents`) ✓
- Manual (status-page) incident list queries also capped at `LIMIT 200` (`queries/incidents.sql`'s `ListStatusPageIncidents` and `ListActiveStatusPageIncidentsForPage`) ✓ — fixed 2026-07-12; the public status page's resolved-incidents section was already paginated (`LIMIT $2 OFFSET $3`), but the private dashboard list and the public page's active-incidents section had no cap until this pass
- `uptime_checks`, `port_checks`, and `dns_checks` pruned after 90 days by `DeleteOldUptimeChecks`/`DeleteOldPortChecks`/`DeleteOldDNSChecks`, alongside the existing `cron_pings` cleanup (`worker.go`'s `pruneOldPings`) ✓ — dns added 2026-07-29 (EP-39)
- Resolved manual incidents pruned after 90 days by `DeleteOldStatusPageIncidents`, same daily cleanup pass as the checks above — uniform across every plan, no per-plan creation limit needed ✓ — added 2026-07-12; still-active incidents are exempt regardless of age, only resolved ones age out
- Active (non-resolved) manual incidents capped at 100 per org (`maxActiveIncidents`, `incidents.go`'s `checkActiveIncidentCap`, called from `CreateIncident`) ✓ — added 2026-07-12, closing the remaining gap the 90-day retention above doesn't cover on its own: an org that never resolves anything could otherwise grow `status_page_incidents` unboundedly, since only resolved rows age out. `409 too_many_active_incidents` on the 101st; resolving one frees a slot. Uniform across every plan, same as the retention window.
- Updates on a single incident capped at 100 (`maxUpdatesPerIncident`, `incidents.go`'s `checkUpdateCap`, called from `PostIncidentUpdate`; `ListStatusPageIncidentUpdates` also `LIMIT 200` as defense in depth) ✓ — added 2026-07-12. The more serious of the two remaining incident gaps found in the same pass: `status_public.go`'s `loadActiveIncidents` renders an incident's full update timeline on every unauthenticated visitor's page load, so unlimited updates on one incident would have grown unbounded on the *public* page, not just a private dashboard list. `409 too_many_incident_updates` on the 101st.
- Maintenance windows capped at 100 per org total (`maxMaintenanceWindows`, `maintenance.go`'s `CreateMaintenanceWindow`; `ListMaintenanceWindows` also `LIMIT 200`) ✓ — added 2026-07-12. Unlike incidents, maintenance windows have no retention/pruning of old ones, so this caps cumulative creation rather than a concurrently-active count. `409 too_many_maintenance_windows` on the 101st.
- API keys capped at 100 active per org (`maxAPIKeys`, `api_keys.go`'s `CreateAPIKey`; `ListAPIKeys` also `LIMIT 200`) ✓ — added 2026-07-12. `409 too_many_api_keys` on the 101st; revoking one frees a slot.
- `loadActiveIncidents` (`status_public.go`) fetches every active incident's update timeline in one batched query (`ListStatusPageIncidentUpdatesForIncidents`, `WHERE incident_id = ANY(...)`), not one query per incident ✓ — fixed 2026-07-12. The per-resource caps above bound each piece individually, but a page with many active incidents was still issuing that many separate DB round-trips on every unauthenticated page load before this; now it's always 2 queries regardless of how many active incidents apply to the page.
- Public status page rate-limited at 300/min by IP, same as its badge endpoints (`server.go:89`) ✓
- Every authenticated route carries a blanket 300 req/min-per-org limit on top of any tighter per-route limit (`server.go:128`) ✓
- Monitor creation is plan-limited (10–1000 per plan) ✓
- Cron pings and uptime check reads are paginated ✓
- Request body capped at 64 KB globally ✓
- Auth and ping endpoints are rate-limited ✓
- All queries use sqlc parameterized statements (no SQL injection) ✓
