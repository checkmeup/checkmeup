# Tech debt

Known architecture/code smells that aren't worth an ADR or an immediate fix, but shouldn't be forgotten. Add an entry when you spot something during other work rather than stopping to fix it; remove an entry once it's addressed (reference the commit/PR in the removal, not here).

This list came out of an architecture review on 2026-06-19. Real bugs found while writing tests are fixed directly and not tracked here — a billing webhook silently swallowing a DB error, a fail-open webhook signature check, a SignUp orphaned-org row on a failed user creation, `ping.go`'s cron recovery only resolving the open incident when `AlertsEnabled` was true (now matches the uptime-monitor worker's pattern: resolution always happens, only the alert send is gated), a CI gap where Turborepo 2.x's default env-var filtering silently dropped `DATABASE_URL` from the `test` task (fixed via `passThroughEnv: ["DATABASE_URL"]` in `turbo.json`), a real **cross-tenant security bug** in `SetStatusPageMonitors` (`status_pages.go`) — it validated a monitor's type and UUID format but never checked org ownership, so any signed-in user could attach another org's monitor ID to their own status page and have that org's live monitor status/90-day uptime bar render on their public page without consent (fixed by routing through `resolveMonitorName`, shared with `maintenance.go`, which scopes the lookup by `org_id`; this also fixed a smaller spec bug where an empty display name fell back to the raw monitor UUID instead of its name, EP-06 US-0602), and `UpdateUptimeMonitor` flooring any sub-10-minute interval to exactly 10 regardless of plan instead of calling `billing.ClampInterval` like `CreateUptimeMonitor` does — a Solo-plan org (1-minute minimum) setting a 1-minute interval via the edit screen got silently bumped to 10, denying them the interval they were paying for; fixed to clamp/reject the same way create does.

---

## Backend (Go)

### Risk

- **`worker.go` is now the only untested file in `internal/`.** Every handler file in `internal/handler/` has its own `*_test.go` as of 2026-06-19 (auth, billing, maintenance windows, cron/uptime/SSL monitors, ping, settings, status pages + the public status page, suggestions) — 11 test files total, plus `internal/middleware/auth_test.go`. `settings_test.go`/`status_pages_test.go` document a routing quirk worth knowing: `TestTelegram`/`TestEmail`/`CheckSlug` don't check auth internally — they're protected only by `RequireAuth` middleware in `server.go`, unlike every other handler in the package which checks `orgIDFrom` itself. Not a bug (the route is correctly gated), but a footgun if any of them is ever called directly from new code without going through the router. Cross-tenant isolation is covered for maintenance windows, cron/uptime/SSL monitors, and status pages (including the monitor-attachment ownership check fixed above). `worker.go` (the goroutine-per-monitor scheduler that actually marks monitors down/overdue, creates incidents, and dispatches alerts — the part `ping_test.go`'s recovery tests and `status_public_test.go`'s status rendering sit downstream of) remains completely untested. [ADR-002](decisions/002-multi-tenancy.md)'s "every tenant query filters by `org_id`" rule is still enforced by code-review convention only — there's no DB-level RLS, which is exactly how the status-page monitor-ownership bug shipped unnoticed.
  → Priority: worker.go — it's where `CreateCronIncident`/`CreateUptimeIncident`/`MarkUptimeMonitorDown`/`UpdateCronMonitorDown` and alert dispatch actually happen, the one part of the request-or-tick lifecycle still with zero coverage.

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
