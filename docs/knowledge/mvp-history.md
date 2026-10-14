---
title: MVP Build History
type: knowledge
status: current
updated: 2026-07-05
tags: [architecture, history, planning]
scope: full project
---

# MVP history

Frozen record of how the MVP was actually built. See [roadmap.md](../roadmap.md) for current Now / Next / Later work, and the [changelog](https://checkmeup.net/changelog) for what's shipped since.

**Work schedule (Israel)** — the basis for the phase hour budgets below (~28–30 h/week):

- Sun–Thu: 3–4 h/day (~3.5 h avg)
- Fri–Sat: 5–6 h/day (~5.5 h avg)
- **Weekly capacity: ~28–30 h/week**

---

## Overview

| Phase | Dates | Days | ~Hours | Epics | Status | Milestone |
|-------|-------|------|--------|-------|--------|-----------|
| [0 — Foundation](#phase-0--foundation) | Jun 14–20 | 7 | 28 h | — | ✅ Done | Dev stack + scaffold running |
| [1 — Auth](#phase-1--auth) | Jun 21–27 | 7 | 28 h | PRD accounts | ✅ Done | Sign up · sign in · sessions |
| [2 — Cron + Alerts + Security](#phase-2--cron-monitor--telegram-alerts--security-hardening) | Jun 14–15 | 2 | 4 h | PRD cron-monitoring, PRD telegram-alerts, PRD security-hardening | ✅ Done | **First useful version** |
| [3 — Uptime](#phase-3--uptime-monitor) | Jun 15 | 1 | 1 h | PRD uptime-monitoring | ✅ Done | URL uptime live |
| [4 — SSL](#phase-4--ssl-monitor) | Jun 15 | 1 | 1 h | PRD ssl-monitoring | ✅ Done | Full monitoring suite |
| [5 — Status page](#phase-5--status-page) | Jun 15 | 1 | 1 h | PRD status-pages | ✅ Done | Public status pages |
| [6 — Billing + polish](#phase-6--billing--polish) | Jun 15–16 | 2 | 4 h | PRD billing | ✅ Done | **🚀 MVP live** |

---

## Phase 0 — Foundation

**Sun Jun 14 → Sat Jun 20 · ~28 h**

Close all open questions before writing any feature code. A bad foundation compounds — every phase depends on this.

### Goals

- [x] Resolve [decision backlog](../../.about/inbox.md) — ADRs 001–011 written (worker model, multi-tenancy, auth, sqlc, infra, status pages, API versioning, error format, logging, testing, Chi)
- [x] Backend scaffold — Go module, Chi router, DB connection pool, `GET /health`, CORS middleware, JWT auth middleware, request logging, first goose migration (users + orgs + refresh_tokens)
- [x] Frontend scaffold — Vue 3 + Vite, Pinia, Vue Router, TanStack Query, Radix Vue, base layout, auth guard, design tokens
- [x] Local dev — `docker-compose.yml` for PostgreSQL so `make dev` starts the full stack
- [x] CI — GitHub Actions: lint + test + Codacy coverage upload on push and PR

### Deliverable

`make dev` starts API on `:8080` and Vue on `:5173`. Health check returns 200. CI is green on a blank commit.

---

## Phase 1 — Auth

**Sun Jun 21 → Sat Jun 27 · ~28 h**

[PRD accounts](../../.about/prd/accounts.md) — 6 stories

Nothing else starts until auth is solid. Invest time here — a shaky auth layer causes pain in every phase that follows.

### Goals

- [x] DB schema — users, orgs, refresh_tokens tables (migration 001_initial.sql)
- [x] PRD accounts Sign up — registration endpoint, sign-up page
- [x] PRD accounts Sign in — JWT + httpOnly cookie, sign-in page
- [x] PRD accounts Sign out — revoke token, clear cookie
- [x] PRD accounts Silent token refresh — refresh endpoint, frontend interceptor
- [x] Protected route guard in Vue Router + auth Pinia store
- [x] Dashboard shell (empty) behind auth

> **PRD accounts (password reset)** implemented with Resend ([ADR alerting](../../.about/adr/alerting.md)). Requires `RESEND_API_KEY` + `APP_URL` in env.

> **PRD accounts (abuse protection)** added 2026-06-15: `go-chi/httprate` rate limiting on sign-up, sign-in, forgot-password (see [ADR limits](../../.about/adr/limits.md)).

### Deliverable

Full auth loop working: register → dashboard → refresh → sign out → redirect to sign-in.

---

## Phase 2 — Cron monitor + Telegram alerts + Security hardening

**Jun 14–15 · ~4 h (completed ahead of schedule)**

[PRD cron-monitoring](../../.about/prd/cron-monitoring.md) — 8 stories
[PRD telegram-alerts](../../.about/prd/telegram-alerts.md) — 4 stories
[PRD security-hardening](../../.about/prd/security-hardening.md) — 4 stories

- [x] DB schema — monitors, pings, incidents (migration 003)
- [x] PRD telegram-alerts Connect Telegram — org settings page, chat ID input, test message button; bot webhook so `/start` replies with chat ID
- [x] PRD telegram-alerts Per-monitor alert toggle
- [x] PRD cron-monitoring Create cron monitor — token generation, CRUD API + create form
- [x] PRD cron-monitoring Monitor list — status badges (up / down / waiting / paused)
- [x] PRD cron-monitoring Monitor detail — config + paginated execution log
- [x] PRD cron-monitoring–0208 Edit, pause/resume, delete
- [x] PRD cron-monitoring Receive a ping — `GET /ping/{token}`, log to DB, return 200 always
- [x] PRD cron-monitoring Missed ping detection — goroutine worker, 30-second ticker, grace period logic
- [x] PRD telegram-alerts Down alert — Telegram message on status → down
- [x] PRD telegram-alerts Recovery alert — Telegram message on status → up

### Milestone 🟢 First useful version — Jun 14

Point a real cron job at checkmeup. Watch pings arrive. Get a Telegram message when the job stops. The product can be used for real at this point.

### Security hardening — Jun 15

- [x] PRD accounts Rate limiting — `go-chi/httprate` on sign-up, sign-in, forgot-password ([ADR limits](../../.about/adr/limits.md))
- [x] PRD security-hardening Rate limit `GET /ping/{token}` — 60 req/min per token (prevents DB flooding)
- [x] PRD security-hardening Rate limit Telegram endpoints — webhook 60 req/min, test-message 5 req/min per IP
- [x] PRD security-hardening Global 64 KB request body cap — `http.MaxBytesReader` middleware (prevents OOM on 4 GB server)
- [x] PRD security-hardening Telegram webhook secret token — `sha256(TELEGRAM_BOT_TOKEN)` validated on every incoming update
- [x] ADR limits written; `cron_pings` retention added to decision backlog

---

## Phase 3 — Uptime monitor

**Jun 15 · ~3 h (completed ahead of schedule)**

[PRD uptime-monitoring](../../.about/prd/uptime-monitoring.md) — 6 stories

Reused patterns from Phase 2 (DB shape, worker model, alert wiring). Completed same day as Phase 4 & 5.

- [x] DB schema — uptime_checks, uptime_incidents (migration 006)
- [x] PRD uptime-monitoring Create uptime monitor — CRUD + form, interval selector (10 min / 30 min)
- [x] PRD uptime-monitoring HTTP health check worker — GET always, 10 s timeout, redirects followed, response time recorded (see ADR workers)
- [x] PRD uptime-monitoring Downtime detection — 2-consecutive-failures rule, alert on transition, recovery alert, alert cap via `max_alerts_per_incident`
- [x] PRD uptime-monitoring Monitor list — uptime %, last checked, status badges
- [x] PRD uptime-monitoring Monitor detail — response time chart (24h), uptime % (24h/7d/30d), incident log, check log
- [x] PRD uptime-monitoring Edit, pause, delete

### Milestone 🟢 URL uptime live — Jun 15

---

## Phase 4 — SSL monitor

**Jun 15 · ~3 h (completed ahead of schedule)**

[PRD ssl-monitoring](../../.about/prd/ssl-monitoring.md) — 5 stories

Simplest monitor type — daily checks, no interval config, fixed alert thresholds at 30/14/7 days.

- [x] DB schema — ssl_monitors table, ssl_monitor_status enum
- [x] PRD ssl-monitoring Create SSL monitor — CRUD + form
- [x] PRD ssl-monitoring Daily cert check worker — TLS dial, record issuer + expiry
- [x] PRD ssl-monitoring Threshold alerts — one alert per threshold crossing (30d/14d/7d + expired)
- [x] PRD ssl-monitoring List + detail views — expiry date, days remaining, status chip
- [x] PRD ssl-monitoring Pause, resume, delete

### Milestone 🟢 Full monitoring suite — Jun 15

All three monitor types live. Cron, uptime, and SSL are all running and alerting.

---

## Phase 5 — Status page

**Mon Jun 15 · ~4 h (completed ahead of schedule)**

[PRD status-pages](../../.about/prd/status-pages.md) — 5 stories

- [x] DB schema — status_pages, status_page_monitors (migration 008)
- [x] ADR status-pages — public page served via Go html/template (no JS required)
- [x] PRD status-pages Create status page — slug validation with real-time availability check
- [x] PRD status-pages Add monitors — multi-select, custom display names, ordering (up/down)
- [x] PRD status-pages Public page — overall status banner, 90-day uptime bars, no-login required, SSR
- [x] PRD status-pages Customize — title, description, logo URL
- [x] PRD status-pages Delete — confirmation dialog, page removed immediately

### Milestone 🟢 Public status pages — Jun 15

A live public `/status/:slug` page that anyone can visit and bookmark.

---

## Phase 6 — Billing + polish

**Jun 15–16 · ~4 h (completed ahead of schedule)**

[PRD billing](../../.about/prd/billing.md) — 3 stories + cross-cutting polish

LemonSqueezy (MoR) handles payments and all global tax. Keep the second week for polish — launch on a clean product.

**Plan limits** (revised Jun 2026 after competitor review — see [ADR billing](../../.about/adr/billing.md)):

| | Hobby | Solo ($9) | Startup ($29) | Enterprise ($99) |
|---|---|---|---|---|
| Monitors | 10 | 30 | 100 | 1000 |
| Status pages | 1 | 3 | 10 | 100 |
| Min uptime interval | 5 min | 1 min | 1 min | 1 min |

### Billing (done)

- [x] PRD billing Plan limit enforcement — 402 API responses on monitor/status-page create
- [x] PRD billing Billing page — current plan, usage bars, upgrade CTA, manage subscription link
- [x] PRD billing LemonSqueezy Checkout — session endpoint, redirect to LS, webhook → update org plan

### Polish (done — Jun 16)

- [x] Empty states on all list views
- [x] Error states + retry buttons on all list views
- [x] Mobile responsiveness — hamburger nav, card layout on small screens
- [x] PRD accounts Password reset (Resend)

### Launch (done — Jun 16)

- [x] Add 402 plan-limit log line — `slog.InfoContext` on every 402, fields: `org_id`, `plan`, `resource`
- [x] End-to-end smoke test: each monitor type + status page
- [x] Production deploy to Hetzner via Kamal (`kamal deploy`)
- [x] Smoke test on production

### Milestone 🚀 MVP live — Jun 16

---

## Hours: estimated vs. logged

The phase/estimate practice ended with the MVP — post-MVP work has no hour budgets (see [ADR docs](../../.about/adr/docs.md)). Phase 7 (maintenance windows, the first post-MVP feature) is included here since it was the last phase planned this way.

| Phase                        | Estimated | Logged   | Delta      |
| ----------------------------- | --------- | -------- | ---------- |
| 0 — Foundation                | 28 h      | 11 h     | -17 h      |
| 1 — Auth                      | 28 h      | 2 h      | -26 h      |
| 2 — Cron + Alerts + Security  | 58 h      | 4 h      | -54 h      |
| 3 — Uptime                    | 28 h      | 1 h      | -27 h      |
| 4 — SSL                       | 22 h      | 1 h      | -21 h      |
| 5 — Status page               | 30 h      | 1 h      | -29 h      |
| 6 — Billing + polish          | 40 h      | 4 h      | -36 h      |
| 7 — Maintenance windows       | 25 h      | 5 h      | 0 h        |
| **Total**                     | **259 h** | **29 h** | **-230 h** |

---

## Buffer notes

- Phase 2 was the riskiest (first real workers). Completed in 4 h instead of estimated 58 h.
- Billing (LemonSqueezy webhook handling) is implemented but activation deferred until first real upgrade intent is observed.
- Shabbat (Sat evening) is a natural reset point each week.
