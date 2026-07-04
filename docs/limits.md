# DoS / Overload Vulnerabilities

Security audit findings — unbounded operations a user or attacker could abuse to overload the system.

---

## HIGH — Incident list queries have no LIMIT

**Files:** `apps/api/queries/monitors.sql:90`, `apps/api/queries/uptime.sql:107`

```sql
SELECT * FROM cron_incidents WHERE monitor_id = $1 ORDER BY started_at DESC;
SELECT * FROM uptime_incidents WHERE monitor_id = $1 ORDER BY started_at DESC;
```

Called in `apps/api/internal/handler/monitors.go:306` and `uptime_monitors.go:427` with no limit. A flapping monitor accumulates thousands of incidents; every detail-page load returns them all.

**Fix:** add `LIMIT 200` to both queries.

---

## MEDIUM — Uptime checks have no retention policy

`cron_pings` are cleaned up every 24h via `DeleteOldCronPings` (worker.go:282). `uptime_checks` have no equivalent cleanup — a 1-minute monitor generates 525K rows/year per monitor and they grow forever.

**Fix:** add a `DeleteOldUptimeChecks` query deleting rows older than 90 days, run in the same daily cleanup ticker.

---

## MEDIUM — Public status page has no rate limit

**File:** `apps/api/internal/server/server.go:79`

```go
r.Get("/status/{slug}", statusPublic.ServeHTTP)  // no rate limit
```

Badge endpoints have 300/min, but the main page (3–4 DB queries per hit) is unprotected.

**Fix:** wrap with `httprate.LimitByIP(300, time.Minute)` like the badge routes.

---

## MEDIUM — Authenticated CRUD endpoints have no rate limit

Auth endpoints (`/sign-in`, `/sign-up`, `/forgot-password`) and suggestions are rate-limited, but every monitor/notification-channel/status-page GET/POST/PATCH/DELETE has no throttle. An authenticated user can hammer list endpoints in a tight loop.

**Fix:** global per-user middleware rate limit (e.g., 120 req/min) applied to the authenticated subrouter in `server.go`.

---

## Things that are fine

- All five check loops (cron overdue, uptime, SSL, domain, port) bounded via a 50-goroutine semaphore (`checkConcurrency`, `worker.go:33`) ✓ — fixed 2026-07-04, cron's `checkOverdue` was the last one still processing sequentially
- Monitor creation is plan-limited (10–1000 per plan) ✓
- Cron pings and uptime check reads are paginated ✓
- Request body capped at 64 KB globally ✓
- Auth and ping endpoints are rate-limited ✓
- All queries use sqlc parameterized statements (no SQL injection) ✓

---

## Priority order

1. Incident query LIMIT
2. Uptime check retention/cleanup
3. Status page rate limit
4. Global authenticated-route rate limit
