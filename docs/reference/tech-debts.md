---
title: Tech Debt
type: reference
status: active
updated: 2026-07-05
tags: [architecture, maintainability, backend]
---

# Tech debt

Known architecture/code smells that aren't worth an ADR or an immediate fix, but shouldn't be forgotten. Add an entry when you spot something during other work rather than stopping to fix it; remove an entry once it's addressed (reference the commit/PR in the removal, not here — `git log` is the record of what was fixed and when).

---

## Backend (Go)

### Maintainability

- **pgxpool has no explicit connection limit** — `apps/api/cmd/api/main.go:35`, `pgxpool.New(ctx, cfg.DatabaseURL)` takes no `Config`, so it falls back to pgx's default (`max(4, NumCPU)`) — likely 4-8 connections on the current Hetzner CX23. Every check (cron/uptime/SSL/domain/port) writes its result through this same pool alongside all live dashboard/API/status-page traffic. Capacity-planning discussion (2026-07-04) flagged this as the most likely actual ceiling on monitor/customer capacity — probably binding well before the worker's semaphore/timeout math does.
  → Set `MaxConns` explicitly (e.g. 20-25) via `pgxpool.ParseConfig`.

- **No shared HTTP client/dialer across checks** — a fresh `http.Client{Timeout: 10*time.Second}` (`worker.go:651`), `net.Dialer` (`worker.go:1008`), and raw `net.DialTimeout` (`worker.go:1373`) are constructed per check instead of reused, losing keep-alive/connection pooling and adding socket/FD churn that grows with monitor count.
  → Share one `http.Client`/`Transport` (with a sane `MaxIdleConnsPerHost`) across checks of the same type.

- **ADR-001 describes a different worker model than what's implemented** — [ADR-001](../decisions/001-worker-model.md) describes goroutine-per-monitor, each with its own `time.Ticker`. The actual code is a single shared 30s poll tick (`worker.go:54`) that queries for due monitors and dispatches a bounded semaphore of goroutines per check type — a materially different scaling profile (poll-tick degrades gracefully by delaying checks under load; goroutine-per-monitor would instead accumulate long-lived goroutines). Caught during the 2026-07-04 capacity-planning discussion.
  → Either update ADR-001's text to match the as-built poll-tick model, or treat this as a deliberate pivot worth its own "Updated" note.

- **`orgIDFrom` helper ownership is unclear** — defined in `internal/handler/monitors.go:88-93` but used repo-wide (`settings.go`, `status_pages.go`, `uptime_monitors.go`, `ssl_monitors.go`, `billing.go`). `suggestions.go:41-49` reimplements the same `uuid.Parse(claims.Subject/.OrgID)` pattern inline instead of reusing it.
  → Move `orgIDFrom`/`userIDFrom` into a shared `handler/context.go`; update `suggestions.go` to use it.

- **Duplicated alert-message string building** across `checkOverdue`, `checkOneUptimeMonitor`, `checkOneSSLMonitor` in `worker.go` — near-identical `fmt.Sprintf` pairs (Telegram/email subject/HTML) repeated per monitor type, e.g. lines 128-138, 202-206, 242-248, 380-407. The SSL threshold-alert block (374-417) is four near-duplicate `case` arms differing only by day count.
  → Factor into a small templated helper shared across monitor types.
