---
title: Tech Debt
type: reference
status: active
updated: 2026-07-10
tags: [architecture, maintainability, backend]
---

# Tech debt

Known architecture/code smells that aren't worth an ADR or an immediate fix, but shouldn't be forgotten. Add an entry when you spot something during other work rather than stopping to fix it; remove an entry once it's addressed (reference the commit/PR in the removal, not here — `git log` is the record of what was fixed and when).

---

## Backend (Go)

### Maintainability

- **pgxpool has no explicit connection limit** — `apps/api/cmd/api/main.go:35`, `pgxpool.New(ctx, cfg.DatabaseURL)` takes no `Config`, so it falls back to pgx's default (`max(4, NumCPU)`) — likely 4-8 connections on the current Hetzner CX23. Every check (cron/uptime/SSL/domain/port) writes its result through this same pool alongside all live dashboard/API/status-page traffic. Capacity-planning discussion (2026-07-04) flagged this as the most likely actual ceiling on monitor/customer capacity — probably binding well before the worker's semaphore/timeout math does.
  → Set `MaxConns` explicitly (e.g. 20-25) via `pgxpool.ParseConfig`.

- **No shared HTTP client/dialer across checks** — a fresh `httpsafe.Dialer(10*time.Second)` is built per call in `worker_uptime.go:187`, `worker_ssl.go:201`, and `worker_port.go:197` (was `http.Client`/`net.Dialer`/`net.DialTimeout` directly in `worker.go` before the SSRF fix and file split on 2026-07-10 — see [knowledge/worker-architecture.md](../knowledge/worker-architecture.md)) instead of reused, losing keep-alive/connection pooling and adding socket/FD churn that grows with monitor count. The SSRF-guard *logic* is now shared (`internal/httpsafe`), but the client/dialer instance itself still isn't.
  → Share one `http.Client`/`Transport` (with a sane `MaxIdleConnsPerHost`) across checks of the same type, still built via `httpsafe.Dialer` for the SSRF guard.

- **`orgIDFrom` helper ownership is unclear** — defined in `internal/handler/monitors.go:88-93` but used repo-wide (`settings.go`, `status_pages.go`, `uptime_monitors.go`, `ssl_monitors.go`, `billing.go`). `suggestions.go:41-49` reimplements the same `uuid.Parse(claims.Subject/.OrgID)` pattern inline instead of reusing it.
  → Move `orgIDFrom`/`userIDFrom` into a shared `handler/context.go`; update `suggestions.go` to use it.

- **Duplicated alert-message string building** across `checkOverdue` (`worker_cron.go`), `checkOneUptimeMonitor` (`worker_uptime.go`), `checkOneSSLMonitor` (`worker_ssl.go`) — near-identical `fmt.Sprintf` pairs (Telegram/email subject/HTML) repeated per monitor type. The SSL threshold-alert block (`worker_ssl.go`'s `sslExpiredMessages`/`sslExpiringSoonMessages`) is a near-duplicate of the domain one (`worker_domain.go`'s `domainExpiredMessages`/`domainExpiringSoonMessages`), differing only by field names (issuer vs registrar).
  → Factor into a small templated helper shared across monitor types.
