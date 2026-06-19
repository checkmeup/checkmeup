# Tech debt

Known architecture/code smells that aren't worth an ADR or an immediate fix, but shouldn't be forgotten. Add an entry when you spot something during other work rather than stopping to fix it; remove an entry once it's addressed (reference the commit/PR in the removal, not here).

This list came out of an architecture review on 2026-06-19. Real bugs found while writing tests are fixed directly and not tracked here — a billing webhook silently swallowing a DB error, a fail-open webhook signature check, a SignUp orphaned-org row on a failed user creation, and `ping.go`'s cron recovery only resolving the open incident when `AlertsEnabled` was true (now matches the uptime-monitor worker's pattern: resolution always happens, only the alert send is gated).

---

## Backend (Go)

### Risk

- **Near-zero test coverage outside auth, billing, maintenance windows, cron monitors, and ping.** `internal/middleware/auth_test.go`, `internal/handler/auth_test.go`, `internal/handler/billing_test.go`, `internal/handler/maintenance_test.go`, `internal/handler/monitors_test.go`, and `internal/handler/ping_test.go` (all added 2026-06-19) are the only test files in the whole `internal/` tree. Cross-tenant isolation is now covered for both maintenance windows and cron monitors (org B can't read/update/pause/delete org A's monitor, or attach org A's monitor to its own maintenance window) — `uptime_monitors.go`, `ssl_monitors.go`, and `status_pages.go` still need the same pattern applied directly (they're only incidentally exercised today as test fixtures via their Create handlers, called from monitors_test.go/maintenance_test.go). `worker.go` (the goroutine-per-monitor scheduler that actually marks monitors down/overdue and triggers the alert paths ping_test.go's recovery tests sit downstream of) remains completely untested. [ADR-002](decisions/002-multi-tenancy.md)'s "every tenant query filters by `org_id`" rule is still enforced by code-review convention only — there's no DB-level RLS.
  → Priority: extend the same cross-tenant-isolation pattern from monitors_test.go/maintenance_test.go to uptime_monitors.go, ssl_monitors.go, and status_pages.go. worker.go is the next big untested surface — it's where incidents actually get created (CreateCronIncident/CreateUptimeIncident) and alerts dispatched.

### Maintainability

- **Unbounded goroutine fan-out per worker tick** — `internal/worker/worker.go:157-167` and `:333-343`. `checkUptimeMonitors`/`checkSSLMonitors` spawn one goroutine per due monitor every tick, no semaphore/pool. Fine at current scale; becomes a real concern as monitor count grows and due-times cluster on the same tick.
  → Bound concurrency with a worker pool sized to a sane ceiling.

- **`orgIDFrom` helper ownership is unclear** — defined in `internal/handler/monitors.go:88-93` but used repo-wide (`settings.go`, `status_pages.go`, `uptime_monitors.go`, `ssl_monitors.go`, `billing.go`). `suggestions.go:41-49` reimplements the same `uuid.Parse(claims.Subject/.OrgID)` pattern inline instead of reusing it.
  → Move `orgIDFrom`/`userIDFrom` into a shared `handler/context.go`; update `suggestions.go` to use it.

- **Duplicated alert-message string building** across `checkOverdue`, `checkOneUptimeMonitor`, `checkOneSSLMonitor` in `worker.go` — near-identical `fmt.Sprintf` pairs (Telegram/email subject/HTML) repeated per monitor type, e.g. lines 128-138, 202-206, 242-248, 380-407. The SSL threshold-alert block (374-417) is four near-duplicate `case` arms differing only by day count.
  → Factor into a small templated helper shared across monitor types.

---

## Frontend (Vue)

### Risk

- **TanStack Query installed and globally registered but unused** — `apps/web/src/main.ts:3,8`, `package.json:14`. Zero `useQuery`/`useMutation` call sites anywhere. All 19 data-fetching views hand-roll `loading`/`error` `ref()` + `onMounted` + try/catch instead. Biggest gap between the stated stack (CLAUDE.md) and reality — currently dead weight in `package.json`.
  → Either wire views onto `useQuery`/`useMutation` for caching/dedup, or drop the dependency. Decide deliberately rather than letting it drift further.

### Maintainability

- **No `composables/` directory** — all data-fetching/business logic lives inline in view `<script setup>` blocks, the same `loading`/`error`/`onMounted` pattern repeated ~19 times. Same root cause as the TanStack Query gap above; a `useApiResource`-style composable (or adopting TanStack Query) would remove most of the duplication.

- **Only one Pinia store (`stores/auth.ts`, 48 lines)** — no store for monitors/billing/settings; all state is view-local. Billing/plan info is fetched independently in `BillingView.vue`, `UptimeMonitorCreateView.vue`, and `UptimeMonitorEditView.vue` — three separate round-trips for the same data in a typical session, no shared cache.

- **Hardcoded `#fff` / raw hex colors instead of design tokens** — `color: #fff` paired with `var(--color-green-500)` across 9 files (`LandingLayout.vue:51,114`, `AboutView.vue:267`, `PricingView.vue:156,279,389`, `HomeView.vue:72,415,484,549`, `BlogPostView.vue:192`, `MaintenanceMonitorPicker.vue:83`, `StatusPageDetailView.vue:240`), plus raw hex dots in `HomeView.vue:94-96` (`#ef4444`, `#f59e0b`, `#1d9e75`) that likely duplicate existing tokens. Not the documented `bg-[--token]` bug (already fixed, doesn't reappear anywhere) — these are valid CSS, just not theme-token-driven.
  → Low risk since white-on-green badges are probably theme-invariant by design, but a `--on-accent` token would be more consistent and future-proof against a new theme.
