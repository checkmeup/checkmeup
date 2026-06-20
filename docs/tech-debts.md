# Tech debt

Known architecture/code smells that aren't worth an ADR or an immediate fix, but shouldn't be forgotten. Add an entry when you spot something during other work rather than stopping to fix it; remove an entry once it's addressed (reference the commit/PR in the removal, not here — `git log` is the record of what was fixed and when).

---

## Backend (Go)

### Maintainability

- **Unbounded goroutine fan-out per worker tick** — `internal/worker/worker.go:157-167` and `:333-343`. `checkUptimeMonitors`/`checkSSLMonitors` spawn one goroutine per due monitor every tick, no semaphore/pool. Fine at current scale; becomes a real concern as monitor count grows and due-times cluster on the same tick.
  → Bound concurrency with a worker pool sized to a sane ceiling.

- **`orgIDFrom` helper ownership is unclear** — defined in `internal/handler/monitors.go:88-93` but used repo-wide (`settings.go`, `status_pages.go`, `uptime_monitors.go`, `ssl_monitors.go`, `billing.go`). `suggestions.go:41-49` reimplements the same `uuid.Parse(claims.Subject/.OrgID)` pattern inline instead of reusing it.
  → Move `orgIDFrom`/`userIDFrom` into a shared `handler/context.go`; update `suggestions.go` to use it.

- **Duplicated alert-message string building** across `checkOverdue`, `checkOneUptimeMonitor`, `checkOneSSLMonitor` in `worker.go` — near-identical `fmt.Sprintf` pairs (Telegram/email subject/HTML) repeated per monitor type, e.g. lines 128-138, 202-206, 242-248, 380-407. The SSL threshold-alert block (374-417) is four near-duplicate `case` arms differing only by day count.
  → Factor into a small templated helper shared across monitor types.
