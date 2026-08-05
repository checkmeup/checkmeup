# checkmeup.net — developer monitoring platform

Cron, uptime, SSL, domain expiry, port (TCP), and DNS record monitors with execution logs, Telegram alerts, and white-label status pages for agencies.

**MVP order:** Cron monitor → Uptime monitor → SSL expiry monitor → Status page  
**Pricing:** Hobby $0 / Solo $9 / Startup $29 / Enterprise $99

---

## Commands

```bash
make dev             # start all apps (hot reload)
make test            # lint + test + merge coverage
make lint            # lint only
make build           # production build
make clean           # wipe node_modules, dist, coverage
make migrate         # run goose migrations (reads DATABASE_URL from apps/api/.env)
make migrate-create name=foo  # create a new goose migration file
```

---

## Stack

**Backend (`apps/api`):** Go · Chi · sqlc · goose · PostgreSQL · JWT auth · Resend (email) · Telegram alerts · Paddle (billing) · go-chi/httprate (rate limiting) · air (hot reload)  
**Frontend (`apps/web`):** Vue 3 · Vite · Pinia · TanStack Query · Radix Vue · Tailwind  
> `typescript` is pinned to `^6.0.3`, not "latest" — TypeScript 7 is a restructured Go-based rewrite that `vue-tsc` can't bootstrap against yet (`ERR_PACKAGE_PATH_NOT_EXPORTED` on `typescript/lib/tsc`), despite `vue-tsc`'s own peerDeps claiming `>=5.0.0` support. A `bun update --latest` will try to float this to 7.x again — re-pin to the latest 6.x if that happens.  
**Infra:** Hetzner CX23 · Kamal 2 · kamal-proxy  
**Test tooling:** golangci-lint · gcov2lcov (Go coverage → lcov) · Vitest · Stryker (mutation testing on `apps/web/src/lib/`, ad hoc via the `mutation-testing` skill — never in CI)

> `make test` requires PostgreSQL running (`docker-compose up db` or inside the devcontainer).

**Key env vars (`apps/api/.env`):**

| Variable | Required | Notes |
|----------|----------|-------|
| `DATABASE_URL` | ✅ | postgres DSN |
| `JWT_SECRET` | ✅ | random secret |
| `RESEND_API_KEY` | optional | enables password-reset emails |
| `APP_URL` | optional | frontend origin (default `http://localhost:5173`) |
| `BASE_URL` | optional | backend origin (default `http://localhost:8080`) |
| `TELEGRAM_BOT_TOKEN` | optional | enables Telegram alerts |
| `TWILIO_ACCOUNT_SID` | optional | enables SMS alerts ([EP-19](docs/stories/ep-19-sms-alerts.md)) — see [`docs/reference/twilio-setup.md`](docs/reference/twilio-setup.md) |
| `TWILIO_API_KEY_SID` | optional | scoped API Key SID (not the primary Account SID) |
| `TWILIO_API_KEY_SECRET` | optional | scoped API Key secret — never the primary Account Auth Token |
| `TWILIO_MESSAGING_SERVICE_SID` | optional | Messaging Service SID SMS is sent from |
| `PADDLE_ENVIRONMENT` | optional | `production` (default) or `sandbox` — selects `api.paddle.com` vs `sandbox-api.paddle.com`; set to `sandbox` for local dev, leave unset in production |
| `PADDLE_API_KEY` | billing | Paddle API key (server-side, secret) |
| `PADDLE_WEBHOOK_SECRET` | billing | Paddle webhook signing secret |
| `PADDLE_SOLO_PRICE_ID` | billing | price ID for Solo plan (monthly) |
| `PADDLE_STARTUP_PRICE_ID` | billing | price ID for Startup plan (monthly) |
| `PADDLE_ENTERPRISE_PRICE_ID` | billing | price ID for Enterprise plan (monthly) |
| `PADDLE_SOLO_ANNUAL_PRICE_ID` | billing | price ID for Solo plan (annual, [EP-27](docs/stories/ep-27-annual-billing.md)) |
| `PADDLE_STARTUP_ANNUAL_PRICE_ID` | billing | price ID for Startup plan (annual) |
| `PADDLE_ENTERPRISE_ANNUAL_PRICE_ID` | billing | price ID for Enterprise plan (annual) |
| `CODACY_API_TOKEN` | CI only | account-level Codacy token |

> All `PADDLE_*` vars are unset by default — see [`docs/reference/billing-setup.md`](docs/reference/billing-setup.md) for the Paddle dashboard checklist to activate billing. The frontend also needs `VITE_PADDLE_CLIENT_TOKEN` (public, safe to expose) in `apps/web/.env` for Paddle.js — see [`apps/web/.env.example`](apps/web/.env.example).

> `TURBO_TELEMETRY_DISABLED=1` is set in CI via `.github/workflows/ci.yml` env block — do not add a `turbo telemetry disable` step.

Full rationale in [`docs/decisions/`](docs/decisions/). Open questions in [`docs/decisions/backlog.md`](docs/decisions/backlog.md). Post-MVP work is tracked as Now/Next/Later in [`docs/roadmap.md`](docs/roadmap.md), stories in [`docs/stories/`](docs/stories/), monthly snapshots in [`docs/reports/`](docs/reports/), and self-incidents in [`docs/incidents/`](docs/incidents/) ([ADR-022](docs/decisions/022-post-mvp-docs-organization.md)). Docs navigation: [`docs/INDEX.md`](docs/INDEX.md) — living how-to in [`docs/reference/`](docs/reference/), architectural snapshots in [`docs/knowledge/`](docs/knowledge/), feature gap analysis in [`docs/proposals/`](docs/proposals/), technical investigations in [`docs/investigations/`](docs/investigations/).

Reusable Claude Code skills live in [`.claude/skills/`](.claude/skills/) — PR merging, release notes, hours logging, mutation testing, and Codacy/security/architecture audits, each self-documenting via its own `SKILL.md`. `.claude/hooks/` enforces the highest-stakes rules above deterministically rather than relying on Claude remembering an advisory instruction — never commit on local `main`, never run a deploy-adjacent command, never force-push `main`, never hand-edit sqlc-generated code, and a `secrets-scan` gate before every `git push` — wired up in [`.claude/settings.json`](.claude/settings.json).

**Hours:** [`docs/reports/hours/`](docs/reports/hours/README.md) holds the raw daily log, one file per month; [`docs/reports/`](docs/reports/) holds monthly rollups. Use the `log-hours` and `monthly-report` skills for these — they already encode the reconstruction rules (git-history grouping, stash-artifact exclusion, quarter-hour granularity). A `PreToolUse` hook blocks `gh pr create` until hours are logged and committed on the current branch, so they ship in the same PR as the work.

---

## Conventions

- **Commits:** conventional commits enforced by commitlint — `feat:`, `fix:`, `chore:`, etc.
- **Merging to `main`:** always via PR, never a direct push — and PRs merge via rebase only (no merge commits, no squash) to keep `main`'s log a straight line. Merge with plain git (`git merge --ff-only` after rebasing on `main`, then `git push origin main`), not `gh pr merge` — the GitHub PAT in use here lacks merge permissions on the repo
- **JS/TS lint+format:** oxlint + oxfmt (run automatically on commit via lint-staged)
- **Go lint:** golangci-lint
- **Multi-tenancy:** every tenant-scoped query **must** filter by `org_id`
- **Colors/logo:** never hardcode hex values — use tokens from [`docs/reference/design.md`](docs/reference/design.md)
- **Naming/capitalization:** in prose, capitalize the organization/product name — "Checkmeup", not "checkmeup" (e.g. "Checkmeup started as the tool I wished existed"). The domain `checkmeup.net` stays lowercase mid-sentence, and only capitalizes to `Checkmeup.net` when it opens a sentence. Never capitalize `checkmeup` where it's a code identifier rather than prose — GitHub org/repo slugs (`github.com/checkmeup/checkmeup`), the npm scope (`@checkmeup/web`), storage/cache keys (`'checkmeup:billingPendingPlan'`), and file/asset names (`checkmeup-og.png`) all stay lowercase
- **Theme:** the app supports light/dark via `data-theme` on `<html>` — design tokens (`--bg`, `--text`, etc.) already flip per theme, so styling with tokens is theme-safe by default; hardcoding a token's current value isn't (see [EP-10](docs/stories/ep-10-theme.md))

---

## Code quality (Codacy)

CI uploads coverage to Codacy after every push (`CODACY_API_TOKEN`, account-level, in `apps/api/.env`). Use the `codacy-triage` skill to fetch and triage the current issue backlog — don't hand-roll the fetch/triage steps inline, the skill already encodes this repo's known-noise rules and keeps them current.

---

## Don't

- Use Playwright (or `npx playwright`) to drive a browser for UI verification — not wanted in this environment. If a UI change needs visual/browser verification and no project run-skill covers it, say so explicitly instead of reaching for Playwright.
- Run `make deploy`, `make ghcr-clean`, `kamal <anything>`, or `docker build`/`buildx` against `config/deploy.yml` — these touch the real production server (`checkmeup.net`) and push to the real GHCR registry. Only the human operator runs these, deliberately, from their own machine. `-n` (dry run) is not a safe way to preview `make deploy`: it has actually triggered a real deploy before — see [the incident](docs/incidents/2026-07-02-make-dry-run-triggered-deploy.md) for why. A `PreToolUse` hook (`.claude/hooks/block_make_deploy.py`, `block_kamal.py`, `block_docker_deploy_build.py`) now blocks all of these for Claude at the tool-call level, dry-run flags included.
- Add Redis, a job queue, or any external broker — goroutine workers are intentional ([ADR-001](docs/decisions/001-worker-model.md))
- Use an ORM — sqlc only ([ADR-004](docs/decisions/004-sqlc-over-orm.md))
- Skip `org_id` filters in DB queries — silent data leak across tenants
- Add subdomains for status pages — `/status/:slug` path is intentional ([ADR-005](docs/decisions/005-status-page-same-domain.md))
- Use `Authorization` header for **browser session** auth — the `access_token` httpOnly cookie is the only mechanism there ([ADR-003](docs/decisions/003-auth-jwt-httponly-cookie.md)). The public API (EP-26) is a separate, non-browser mechanism and authenticates via `X-API-Key` instead — never `Authorization`, to keep the two visibly distinct ([ADR-028](docs/decisions/028-api-key-auth-scope.md))
- Use `api.get/post/…` in `auth.init()` — use plain `fetch` there to bypass the 401 interceptor; a 401 on `/me` during init means "not logged in", not a session error
- Switch payment providers — Paddle is the MoR (handles global tax); see [ADR-026](docs/decisions/026-billing-paddle-mor.md) (supersedes [ADR-018](docs/decisions/018-billing-lemonsqueezy-mor.md), which chose LemonSqueezy)
- Write Tailwind classes like `bg-[--token]` or `text-[--token]` for a CSS variable — this Tailwind v4 setup compiles that to invalid CSS (`background-color: --token`, missing `var()`), silently dropping the style. Always write `bg-[var(--token)]`. Found broken across `Button.vue`/`Input.vue`/`Label.vue`/`LandingLayout.vue` during [EP-10](docs/stories/ep-10-theme.md) — e.g. the sign-in button had no background in *either* theme until fixed
- Expect a module to be both eagerly imported (even just one named export) *and* code-split into its own lazy chunk — Rollup bundles the whole module (every export) into the eager chunk regardless, logged as `[INEFFECTIVE_DYNAMIC_IMPORT]`. To get a real lazy chunk for something (e.g. a blog post's heavy content), its cheap metadata needs to live in a genuinely separate file that nothing lazy-loaded ever gets eagerly imported from — see `apps/web/src/blog/postsMeta.ts` vs `posts/*.ts` for the pattern
- Build a public feature-request board, voting system, or ticketing queue — feedback goes straight to the founder via the in-app form (Settings, EP-23), email, and GitHub Issues; that's a deliberate choice, not a gap (see `DocsView.vue`'s "Need help?" section)
- Commit directly onto local `main` — always create/switch to a feature branch first, even for doc-only or copy-only changes. A `PreToolUse` hook (`.claude/hooks/block_main_commit.py`) blocks this for Claude, so it should fail fast rather than needing a recovery — but if it ever happens anyway: create the branch at the existing commits, then `git reset --hard origin/main` to move `main` back before pushing the branch and opening the PR
