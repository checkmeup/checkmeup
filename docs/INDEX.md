# Docs Index

<!-- AI LOAD ORDER: INDEX.md → reference/ → decisions/ → knowledge/ → proposals/ → investigations/ (active only) -->

Navigation cache for `docs/`. Always load this file first. Every other doc carries its own
frontmatter (`type`, `status`, `updated`, `tags`) — see each folder's `_template.md` for the schema.

**Folder roles:**

| Folder | Question answered | Mutable? |
|---|---|---|
| `reference/` | How does the system work right now? | Yes — update when code or ops change |
| `decisions/` | Why was it built this way? | No — append only |
| `knowledge/` | What do we understand about it? | No — write new doc, set `superseded_by` |
| `investigations/` | What is broken or needs analysis? | No — permanent once written |
| `proposals/` | What might we add or change? | Yes — until decided |
| `stories/` | What are we building and what's the scope? | Yes — update as epics progress |
| `incidents/` | What broke in production? | No — permanent once written |
| `reports/` | What shipped this month? | No — append only per month |

**Archive rule:** Set `status: archived` in frontmatter. Remove from the tables below.
Do not move files — lifecycle is metadata, not filesystem position.

**Controlled tag vocabulary** (reuse these, don't invent new ones): `billing`, `paddle`, `twilio`,
`ops`, `deployment`, `kamal`, `hetzner`, `design`, `tokens`, `colors`, `logo`, `security`, `dos`,
`audit`, `architecture`, `maintainability`, `backend`, `frontend`, `vue3`, `auth`, `monitoring`,
`cron`, `uptime`, `ssl`, `domain`, `port`, `alerts`, `telegram`, `sms`, `email`, `webhook`,
`status-page`, `features`, `competitors`, `backlog`, `history`, `planning`.

## reference/ — living, update-on-change

| Doc | Status | Hook |
| --- | --- | --- |
| [billing-setup.md](reference/billing-setup.md) | active | Paddle dashboard checklist to activate billing |
| [deploy.md](reference/deploy.md) | active | Kamal deployment guide, cost breakdown, gotchas |
| [design.md](reference/design.md) | active | Color tokens, logo assets, theme conventions |
| [limits.md](reference/limits.md) | active | DoS / overload vulnerability audit (all findings resolved) |
| [tech-debts.md](reference/tech-debts.md) | active | Known architecture smells and code quality debts |
| [twilio-setup.md](reference/twilio-setup.md) | active | Twilio account setup checklist for SMS alerts |

## decisions/ — immutable, append-only

| Doc | Status | Hook |
| --- | --- | --- |
| [001-worker-model.md](decisions/001-worker-model.md) | accepted | Goroutine/poll-tick workers, no external queue |
| [002-multi-tenancy.md](decisions/002-multi-tenancy.md) | accepted | org_id on every tenant-scoped table |
| [003-auth-jwt-httponly-cookie.md](decisions/003-auth-jwt-httponly-cookie.md) | accepted | JWT in httpOnly cookie for browser sessions |
| [004-sqlc-over-orm.md](decisions/004-sqlc-over-orm.md) | accepted | sqlc only, no ORM |
| [005-status-page-same-domain.md](decisions/005-status-page-same-domain.md) | accepted | /status/:slug — no subdomains or custom domains |
| [006-infrastructure-hetzner-kamal-traefik.md](decisions/006-infrastructure-hetzner-kamal-traefik.md) | accepted | Single Hetzner CX23, Kamal 2, kamal-proxy |
| [007-api-versioning.md](decisions/007-api-versioning.md) | accepted | /api/v1/ URL prefix |
| [008-api-error-format.md](decisions/008-api-error-format.md) | accepted | {error, code} JSON shape |
| [009-logging.md](decisions/009-logging.md) | accepted | slog structured logging |
| [010-testing-strategy.md](decisions/010-testing-strategy.md) | accepted | Integration tests against real DB, no mocks |
| [011-chi-router.md](decisions/011-chi-router.md) | accepted | Chi over Gin/Echo/stdlib |
| [012-email-resend.md](decisions/012-email-resend.md) | accepted | Resend for transactional email |
| [013-rate-limiting.md](decisions/013-rate-limiting.md) | accepted | go-chi/httprate per-endpoint limits |
| [014-uptime-check-mechanics.md](decisions/014-uptime-check-mechanics.md) | accepted | HTTP check behavior (redirects, timeout, response time) |
| [015-cron-pings-retention.md](decisions/015-cron-pings-retention.md) | accepted | Cron ping retention policy |
| [016-alert-debounce.md](decisions/016-alert-debounce.md) | accepted | Alert debounce / max_alerts_per_incident |
| [017-status-page-ssr.md](decisions/017-status-page-ssr.md) | accepted | Public status page rendered via Go html/template (no JS) |
| [018-billing-lemonsqueezy-mor.md](decisions/018-billing-lemonsqueezy-mor.md) | superseded | Original LemonSqueezy choice — superseded by ADR-026 |
| [019-plan-limits.md](decisions/019-plan-limits.md) | accepted | Plan limits table (monitors, status pages, channels, SMS credits) |
| [020-maintenance-windows.md](decisions/020-maintenance-windows.md) | accepted | Maintenance windows exclude monitors from check loops |
| [021-versioning.md](decisions/021-versioning.md) | accepted | Semantic versioning / release notes |
| [022-post-mvp-docs-organization.md](decisions/022-post-mvp-docs-organization.md) | accepted | Post-MVP docs structure (original; current structure extended by INDEX.md) |
| [023-notification-channels.md](decisions/023-notification-channels.md) | accepted | Polymorphic notification_channels table |
| [024-no-viber-commercial-model.md](decisions/024-no-viber-commercial-model.md) | accepted | Viber deferred — no commercial API without registered business |
| [025-license-busl.md](decisions/025-license-busl.md) | accepted | BUSL 1.1 license (→ Apache 2.0 on 2030-07-01) |
| [026-billing-paddle-mor.md](decisions/026-billing-paddle-mor.md) | accepted | Paddle as MoR, replacing LemonSqueezy |
| [027-web-analytics-ga4.md](decisions/027-web-analytics-ga4.md) | accepted | GA4 for web analytics |
| [028-api-key-auth-scope.md](decisions/028-api-key-auth-scope.md) | accepted | X-API-Key for public API, distinct from browser cookie auth |
| [029-sms-alerts-twilio.md](decisions/029-sms-alerts-twilio.md) | accepted | Twilio for SMS, opt-in checkbox + consent record |
| [030-totp-secret-encryption.md](decisions/030-totp-secret-encryption.md) | accepted | TOTP secret encryption at rest |
| [031-team-invite-conflict.md](decisions/031-team-invite-conflict.md) | accepted | Team invite conflict resolution |
| [032-sms-credit-quotas.md](decisions/032-sms-credit-quotas.md) | accepted | Plan-bundled SMS credit quotas |
| [033-target-customer-freelancers.md](decisions/033-target-customer-freelancers.md) | accepted | Freelance web devs as primary ICP |
| [034-manual-incident-schema.md](decisions/034-manual-incident-schema.md) | accepted | Manual incidents in a new `status_page_incidents` table |
| [035-status-page-hide-branding.md](decisions/035-status-page-hide-branding.md) | accepted | Per-page "hide branding" toggle, gated to paid plans |
| [036-flat-safety-caps.md](decisions/036-flat-safety-caps.md) | accepted | Flat 100-per-org caps on incident updates, maintenance windows, API keys |
| [backlog.md](decisions/backlog.md) | — | Open questions not yet resolved into ADRs |

## knowledge/ — architectural snapshots, versioned not frozen

| Doc | Status | Hook |
| --- | --- | --- |
| [mvp-history.md](knowledge/mvp-history.md) | current | Frozen record of the MVP build: phases, hours, decisions |
| [worker-architecture.md](knowledge/worker-architecture.md) | current | As-built poll-tick/semaphore worker model (corrects ADR-001) |
| [notification-channels.md](knowledge/notification-channels.md) | current | Five alert channels, shared delivery/SSRF plumbing, dispatch fallback |
| [billing-plan-enforcement.md](knowledge/billing-plan-enforcement.md) | current | Plan-limit map, creation-time checks, downgrade enforcement, Paddle webhook |
| [public-api-auth.md](knowledge/public-api-auth.md) | current | X-API-Key auth, key hashing, rate limiting, unshipped scope enforcement gap |

## investigations/ — technical root cause analyses

| Doc | Status | Review after | Hook |
| --- | --- | --- | --- |
| _(none yet)_ | | | |

## proposals/ — undecided options only

| Doc | Status | Hook |
| --- | --- | --- |
| [bucket-list.md](proposals/bucket-list.md) | proposed | Competitor gap analysis — unprioritized feature ideas |

## stories/ — epics and acceptance criteria

See [stories/backlog.md](stories/backlog.md) for the full epic catalog.

Active epics: EP-01 through EP-36 — see `stories/` directory.

## incidents/ — production incidents

See [incidents/README.md](incidents/README.md) for the template. One file per incident.

## reports/ — monthly shipping logs

| Doc |
| --- |
| [2026-06.md](reports/2026-06.md) |
| [2026-07.md](reports/2026-07.md) |

## Root-level docs

| Doc | Hook |
| --- | --- |
| [roadmap.md](roadmap.md) | Now / Next / Later product work |
| [hours.md](hours.md) | Raw daily work-hours log |

## Conventions

- **Naming:** kebab-case, descriptive noun phrase, no dates in the filename (dates live in `updated`).
- **One topic per file:** if a doc grows past a few hundred lines and covers more than one thing
  someone would search for independently, split it.
- **Status drives visibility:** only `active` / `current` / `proposed` docs appear in the tables above.
  Archived docs stay in place and are excluded from active reasoning.
- **decisions/ uniqueness:** no two ADRs cover the same decision — if a decision changes, write a new ADR
  that supersedes the old one (set `superseded_by` in the old one's frontmatter).
- **investigations/ decay:** check `review_after` dates — if the context is no longer relevant,
  set `status: archived` and remove from the table above.
- **Standing "don't do X" rules** belong in `CLAUDE.md`, not here.
