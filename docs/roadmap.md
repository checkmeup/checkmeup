# MVP roadmap

**Work schedule (Israel):**
- Sun–Thu: 3–4 h/day (~3.5 h avg)
- Fri–Sat: 5–6 h/day (~5.5 h avg)
- **Weekly capacity: ~28–30 h/week**

**Target: MVP live by ~14 Aug 2026 (~9 weeks from Jun 14)**

---

## Overview

| Phase | Dates | Days | ~Hours | Epics | Status | Milestone |
|-------|-------|------|--------|-------|--------|-----------|
| [0 — Foundation](#phase-0--foundation) | Jun 14–20 | 7 | 28 h | — | ✅ Done | Dev stack + scaffold running |
| [1 — Auth](#phase-1--auth) | Jun 21–27 | 7 | 28 h | EP-01 | ✅ Done | Sign up · sign in · sessions |
| [2 — Cron + Alerts](#phase-2--cron-monitor--telegram-alerts) | Jun 28–Jul 11 | 14 | 57 h | EP-02, EP-05 | ⬜ Not started | **First useful version** |
| [3 — Uptime](#phase-3--uptime-monitor) | Jul 12–18 | 7 | 28 h | EP-03 | ⬜ Not started | URL uptime live |
| [4 — SSL](#phase-4--ssl-monitor) | Jul 19–24 | 6 | 22 h | EP-04 | ⬜ Not started | Full monitoring suite |
| [5 — Status page](#phase-5--status-page) | Jul 25–Aug 1 | 8 | 30 h | EP-06 | ⬜ Not started | Public status pages |
| [6 — Billing + polish](#phase-6--billing--polish) | Aug 2–14 | 13 | 40 h | EP-07 | ⬜ Not started | **MVP launch-ready** |

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

[EP-01](stories/ep-01-auth.md) — 5 stories

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

### Deliverable

Full auth loop working: register → dashboard → refresh → sign out → redirect to sign-in.

---

## Phase 2 — Cron monitor + Telegram alerts

**Sun Jun 28 → Sat Jul 11 · ~57 h (2 weeks)**

[EP-02](stories/ep-02-cron-monitor.md) — 8 stories  
[EP-05](stories/ep-05-telegram-alerts.md) — 4 stories

This is the riskiest phase — first real background workers, most stories. Build Telegram first so every alert integration can be tested as you go.

### Week 1 — Sun Jun 28 → Sat Jul 4 (~28 h)

**Cron CRUD + Telegram setup**

- [ ] DB schema — monitors, pings, incidents
- [ ] US-0501 Connect Telegram — org settings page, chat ID input, test message button
- [ ] US-0504 Per-monitor alert toggle
- [ ] US-0201 Create cron monitor — token generation, CRUD API + create form
- [ ] US-0204 Monitor list — status badges (up / down / waiting / paused)
- [ ] US-0205 Monitor detail — config + paginated execution log
- [ ] US-0206–0208 Edit, pause/resume, delete

### Week 2 — Sun Jul 5 → Sat Jul 11 (~28 h)

**Ping endpoint + worker + live alerts**

- [ ] US-0202 Receive a ping — `GET /ping/{token}`, log to DB, return 200 always
- [ ] US-0203 Missed ping detection — goroutine worker, 30-second ticker, grace period logic
- [ ] US-0502 Down alert — Telegram message on status → down
- [ ] US-0503 Recovery alert — Telegram message on status → up

### Milestone 🟢 First useful version — Sat Jul 11

Point a real cron job at checkmeup. Watch pings arrive. Get a Telegram message when the job stops. The product can be used for real at this point.

---

## Phase 3 — Uptime monitor

**Sun Jul 12 → Sat Jul 18 · ~28 h**

[EP-03](stories/ep-03-uptime-monitor.md) — 6 stories

Reuses patterns from Phase 2 (DB shape, worker model, alert wiring). Should move faster.

- [ ] DB schema — uptime_checks, uptime_incidents
- [ ] US-0301 Create uptime monitor — CRUD + form, interval selector
- [ ] US-0302 HTTP health check worker — HEAD→GET fallback, 10 s timeout, redirect follow
- [ ] US-0303 Downtime detection — 2-consecutive-failures rule, alert on transition
- [ ] US-0304 Monitor list — uptime %, last checked
- [ ] US-0305 Monitor detail — response time chart, uptime % (24h/7d/30d), incident log
- [ ] US-0306 Edit, pause, delete

### Deliverable

URL uptime monitoring is live end-to-end with Telegram alerts.

---

## Phase 4 — SSL monitor

**Sun Jul 19 → Fri Jul 24 · ~22 h**

[EP-04](stories/ep-04-ssl-monitor.md) — 5 stories

Simplest monitor type — daily checks, no interval config, fixed alert thresholds at 30/14/7 days.

- [ ] DB schema — ssl_checks
- [ ] US-0401 Create SSL monitor — CRUD + form
- [ ] US-0402 Daily cert check worker — TLS dial, record issuer + expiry
- [ ] US-0403 Threshold alerts — one alert per threshold crossing
- [ ] US-0404 List + detail views — expiry date, days remaining, status chip
- [ ] US-0405 Pause, delete

### Milestone 🟢 Full monitoring suite — Fri Jul 24

All three monitor types live. Cron, uptime, and SSL are all running and alerting.

---

## Phase 5 — Status page

**Sat Jul 25 → Sat Aug 1 · ~30 h**

[EP-06](stories/ep-06-status-page.md) — 5 stories

First public-facing surface. Needs to look polished — visitors will judge the product by it.

> **Decide SSR vs static render** for the public page before writing US-0603. Add to decision backlog if not resolved by Phase 4.

- [ ] DB schema — status_pages, status_page_monitors
- [ ] US-0601 Create status page — slug validation with real-time availability check
- [ ] US-0602 Add monitors — multi-select, custom display names, ordering
- [ ] US-0603 Public page — overall status banner, 90-day uptime bars, no-login required
- [ ] US-0604 Customise — title, description, logo URL
- [ ] US-0605 Delete

### Deliverable

A live public `/status/:slug` page that anyone can visit and bookmark.

---

## Phase 6 — Billing + polish

**Sun Aug 2 → Fri Aug 14 · ~40 h**

[EP-07](stories/ep-07-billing.md) — 3 stories + cross-cutting polish

Stripe Checkout is straightforward but webhooks need careful handling. Keep the second week for polish — launch on a clean product.

### Week 1 — Sun Aug 2 → Sat Aug 8 (~28 h)

**Billing**

- [ ] Configure Stripe products + prices in the Stripe dashboard
- [ ] US-0702 Plan limit enforcement — API middleware, inline upgrade prompts in UI
- [ ] US-0701 Billing page — current plan, usage bars, upgrade CTA
- [ ] US-0703 Stripe Checkout — session endpoint, success/cancel redirect, webhook → update org plan

### Week 2 — Sun Aug 9 → Fri Aug 14 (~18 h)

**Polish + launch**

- [ ] Empty states on all list views
- [ ] Error states and network failure handling in the frontend
- [ ] Mobile responsiveness pass
- [ ] End-to-end smoke test: each monitor type + billing upgrade + status page
- [ ] US-0105 Password reset — if an email provider is wired up by now
- [ ] Production deploy to Hetzner via Kamal (`kamal deploy`)
- [ ] Smoke test on production

### Milestone 🚀 MVP launch — ~Fri Aug 14

---

## Buffer notes

- Phase 2 is the riskiest (first real workers, 12 stories). If it slips into the weekend of Jul 11–12 that's fine — the following phases are lighter.
- Phase 4 finishes Thursday–Friday, giving a natural buffer before Phase 5 starts Saturday.
- Billing (Stripe webhook handling) often takes longer than expected. If Phase 6 week 1 runs long, cut polish scope — not billing.
- Shabbat (Sat evening) is a natural reset point each week.
