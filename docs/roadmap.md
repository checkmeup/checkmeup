# MVP roadmap

**Work schedule (Israel):**
- Sun–Thu: 3–4 h/day (~3.5 h avg)
- Fri–Sat: 5–6 h/day (~5.5 h avg)
- **Weekly capacity: ~28–30 h/week**

**🚀 MVP launched Jun 16 2026 — 9 weeks ahead of the original Aug 14 target.**

---

## Overview

| Phase | Dates | Days | ~Hours | Epics | Status | Milestone |
|-------|-------|------|--------|-------|--------|-----------|
| [0 — Foundation](#phase-0--foundation) | Jun 14–20 | 7 | 28 h | — | ✅ Done | Dev stack + scaffold running |
| [1 — Auth](#phase-1--auth) | Jun 21–27 | 7 | 28 h | EP-01 | ✅ Done | Sign up · sign in · sessions |
| [2 — Cron + Alerts + Security](#phase-2--cron-monitor--telegram-alerts--security-hardening) | Jun 14–15 | 2 | 4 h | EP-02, EP-05, EP-08 | ✅ Done | **First useful version** |
| [3 — Uptime](#phase-3--uptime-monitor) | Jun 15 | 1 | 1 h | EP-03 | ✅ Done | URL uptime live |
| [4 — SSL](#phase-4--ssl-monitor) | Jun 15 | 1 | 1 h | EP-04 | ✅ Done | Full monitoring suite |
| [5 — Status page](#phase-5--status-page) | Jun 15 | 1 | 1 h | EP-06 | ✅ Done | Public status pages |
| [6 — Billing + polish](#phase-6--billing--polish) | Jun 15–16 | 2 | 4 h | EP-07 | ✅ Done | **🚀 MVP live** |

---

## Phase 0 — Foundation

**Sun Jun 14 → Sat Jun 20 · ~28 h**

Close all open questions before writing any feature code. A bad foundation compounds — every phase depends on this.

### Goals

- [x] Resolve [decision backlog](decisions/backlog.md) — ADRs 001–011 written (worker model, multi-tenancy, auth, sqlc, infra, status pages, API versioning, error format, logging, testing, Chi)
- [x] Backend scaffold — Go module, Chi router, DB connection pool, `GET /health`, CORS middleware, JWT auth middleware, request logging, first goose migration (users + orgs + refresh_tokens)
- [x] Frontend scaffold — Vue 3 + Vite, Pinia, Vue Router, TanStack Query, Radix Vue, base layout, auth guard, design tokens
- [x] Local dev — `docker-compose.yml` for PostgreSQL so `make dev` starts the full stack
- [x] CI — GitHub Actions: lint + test + Codacy coverage upload on push and PR

### Deliverable

`make dev` starts API on `:8080` and Vue on `:5173`. Health check returns 200. CI is green on a blank commit.

---

## Phase 1 — Auth

**Sun Jun 21 → Sat Jun 27 · ~28 h**

[EP-01](stories/ep-01-auth.md) — 6 stories

Nothing else starts until auth is solid. Invest time here — a shaky auth layer causes pain in every phase that follows.

### Goals

- [x] DB schema — users, orgs, refresh_tokens tables (migration 001_initial.sql)
- [x] US-0101 Sign up — registration endpoint, sign-up page
- [x] US-0102 Sign in — JWT + httpOnly cookie, sign-in page
- [x] US-0103 Sign out — revoke token, clear cookie
- [x] US-0104 Silent token refresh — refresh endpoint, frontend interceptor
- [x] Protected route guard in Vue Router + auth Pinia store
- [x] Dashboard shell (empty) behind auth

> **US-0105 (password reset)** implemented with Resend ([ADR-012](decisions/012-email-resend.md)). Requires `RESEND_API_KEY` + `APP_URL` in env.

> **US-0106 (abuse protection)** added 2026-06-15: `go-chi/httprate` rate limiting on sign-up, sign-in, forgot-password (see [ADR-013](decisions/013-rate-limiting.md)).

### Deliverable

Full auth loop working: register → dashboard → refresh → sign out → redirect to sign-in.

---

## Phase 2 — Cron monitor + Telegram alerts + Security hardening

**Jun 14–15 · ~4 h (completed ahead of schedule)**

[EP-02](stories/ep-02-cron-monitor.md) — 8 stories  
[EP-05](stories/ep-05-telegram-alerts.md) — 4 stories  
[EP-08](stories/ep-08-security-hardening.md) — 4 stories

- [x] DB schema — monitors, pings, incidents (migration 003)
- [x] US-0501 Connect Telegram — org settings page, chat ID input, test message button; bot webhook so `/start` replies with chat ID
- [x] US-0504 Per-monitor alert toggle
- [x] US-0201 Create cron monitor — token generation, CRUD API + create form
- [x] US-0204 Monitor list — status badges (up / down / waiting / paused)
- [x] US-0205 Monitor detail — config + paginated execution log
- [x] US-0206–0208 Edit, pause/resume, delete
- [x] US-0202 Receive a ping — `GET /ping/{token}`, log to DB, return 200 always
- [x] US-0203 Missed ping detection — goroutine worker, 30-second ticker, grace period logic
- [x] US-0502 Down alert — Telegram message on status → down
- [x] US-0503 Recovery alert — Telegram message on status → up

### Milestone 🟢 First useful version — Jun 14

Point a real cron job at checkmeup. Watch pings arrive. Get a Telegram message when the job stops. The product can be used for real at this point.

### Security hardening — Jun 15

- [x] US-0106 Rate limiting — `go-chi/httprate` on sign-up, sign-in, forgot-password ([ADR-013](decisions/013-rate-limiting.md))
- [x] US-0801 Rate limit `GET /ping/{token}` — 60 req/min per token (prevents DB flooding)
- [x] US-0802 Rate limit Telegram endpoints — webhook 60 req/min, test-message 5 req/min per IP
- [x] US-0803 Global 64 KB request body cap — `http.MaxBytesReader` middleware (prevents OOM on 4 GB server)
- [x] US-0804 Telegram webhook secret token — `sha256(TELEGRAM_BOT_TOKEN)` validated on every incoming update
- [x] ADR-013 written; `cron_pings` retention added to decision backlog

---

## Phase 3 — Uptime monitor

**Jun 15 · ~3 h (completed ahead of schedule)**

[EP-03](stories/ep-03-uptime-monitor.md) — 6 stories

Reused patterns from Phase 2 (DB shape, worker model, alert wiring). Completed same day as Phase 4 & 5.

- [x] DB schema — uptime_checks, uptime_incidents (migration 006)
- [x] US-0301 Create uptime monitor — CRUD + form, interval selector (10 min / 30 min)
- [x] US-0302 HTTP health check worker — GET always, 10 s timeout, redirects followed, response time recorded (see ADR-014)
- [x] US-0303 Downtime detection — 2-consecutive-failures rule, alert on transition, recovery alert, alert cap via `max_alerts_per_incident`
- [x] US-0304 Monitor list — uptime %, last checked, status badges
- [x] US-0305 Monitor detail — response time chart (24h), uptime % (24h/7d/30d), incident log, check log
- [x] US-0306 Edit, pause, delete

### Milestone 🟢 URL uptime live — Jun 15

---

## Phase 4 — SSL monitor

**Jun 15 · ~3 h (completed ahead of schedule)**

[EP-04](stories/ep-04-ssl-monitor.md) — 5 stories

Simplest monitor type — daily checks, no interval config, fixed alert thresholds at 30/14/7 days.

- [x] DB schema — ssl_monitors table, ssl_monitor_status enum
- [x] US-0401 Create SSL monitor — CRUD + form
- [x] US-0402 Daily cert check worker — TLS dial, record issuer + expiry
- [x] US-0403 Threshold alerts — one alert per threshold crossing (30d/14d/7d + expired)
- [x] US-0404 List + detail views — expiry date, days remaining, status chip
- [x] US-0405 Pause, resume, delete

### Milestone 🟢 Full monitoring suite — Jun 15

All three monitor types live. Cron, uptime, and SSL are all running and alerting.

---

## Phase 5 — Status page

**Mon Jun 15 · ~4 h (completed ahead of schedule)**

[EP-06](stories/ep-06-status-page.md) — 5 stories

- [x] DB schema — status_pages, status_page_monitors (migration 008)
- [x] ADR-017 — public page served via Go html/template (no JS required)
- [x] US-0601 Create status page — slug validation with real-time availability check
- [x] US-0602 Add monitors — multi-select, custom display names, ordering (up/down)
- [x] US-0603 Public page — overall status banner, 90-day uptime bars, no-login required, SSR
- [x] US-0604 Customize — title, description, logo URL
- [x] US-0605 Delete — confirmation dialog, page removed immediately

### Milestone 🟢 Public status pages — Jun 15

A live public `/status/:slug` page that anyone can visit and bookmark.

---

## Phase 6 — Billing + polish

**Jun 15–16 · ~4 h (completed ahead of schedule)**

[EP-07](stories/ep-07-billing.md) — 3 stories + cross-cutting polish

LemonSqueezy (MoR) handles payments and all global tax. Keep the second week for polish — launch on a clean product.

**Plan limits** (revised Jun 2026 after competitor review — see [ADR-019](decisions/019-plan-limits.md)):

| | Hobbyist | Indie ($9) | Studio ($29) | Agency ($79) |
|---|---|---|---|---|
| Monitors | 10 | 30 | 100 | unlimited |
| Status pages | 1 | 3 | 10 | unlimited |
| Min uptime interval | 5 min | 1 min | 1 min | 1 min |

### Billing (done)

- [x] US-0702 Plan limit enforcement — 402 API responses on monitor/status-page create
- [x] US-0701 Billing page — current plan, usage bars, upgrade CTA, manage subscription link
- [x] US-0703 LemonSqueezy Checkout — session endpoint, redirect to LS, webhook → update org plan

### Polish (done — Jun 16)

- [x] Empty states on all list views
- [x] Error states + retry buttons on all list views
- [x] Mobile responsiveness — hamburger nav, card layout on small screens
- [x] US-0105 Password reset (Resend)

### Launch (done — Jun 16)

- [x] Add 402 plan-limit log line — `slog.InfoContext` on every 402, fields: `org_id`, `plan`, `resource`
- [x] End-to-end smoke test: each monitor type + status page
- [x] Production deploy to Hetzner via Kamal (`kamal deploy`)
- [x] Smoke test on production

### Milestone 🚀 MVP live — Jun 16

---

## Deferred to post-MVP (trigger: first 402 hit in production or user asks about paid plan)

- [ ] Configure LemonSqueezy products + variants in the LS dashboard; add `LS_*` env vars
- [ ] US-0702 UI inline upgrade prompt on 402 (create form shows upgrade CTA instead of generic error)
- [ ] Verify failed-payment redirect URL in LemonSqueezy checkout settings

---

## Buffer notes

- Phase 2 was the riskiest (first real workers). Completed in 4 h instead of estimated 58 h.
- Billing (LemonSqueezy webhook handling) is implemented but activation deferred until first real upgrade intent is observed.
- Shabbat (Sat evening) is a natural reset point each week.
