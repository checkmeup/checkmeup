# checkmeup.net — developer monitoring platform

Cron, uptime, SSL, domain expiry, and port (TCP) monitors with execution logs, Telegram alerts, and white-label status pages for agencies.

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
**Infra:** Hetzner CX23 · Kamal 2 · kamal-proxy  
**Test tooling:** golangci-lint · gcov2lcov (Go coverage → lcov) · Vitest

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

Reusable Claude Code skills live in [`.claude/skills/`](.claude/skills/) — PR merging, release notes, hours logging, and Codacy/security/architecture audits, each self-documenting via its own `SKILL.md`.

**Hours:** [`docs/hours.md`](docs/hours.md) is the raw daily log (`Date | Day | Epic/Story | Hours`). When asked to log/update hours for a day, check `git log --since/--until` for that day's commits (exclude stash artifacts — `git log --all` surfaces `refs/stash` entries as fake commits titled "On <branch>:"/"index on <branch>:"/"untracked files on <branch>:"), group related commits into one line per logical task, and estimate effort from diff size and complexity — minimum 1h per task/line, combine small related commits rather than one line per commit. Non-commit work in the same session (launch/marketing copy, doc corrections, etc.) gets its own line too. Roll the new daily total into that month's `docs/reports/YYYY-MM.md` Notes section.

---

## Conventions

- **Commits:** conventional commits enforced by commitlint — `feat:`, `fix:`, `chore:`, etc.
- **Merging to `main`:** always via PR, never a direct push — and PRs merge via rebase only (no merge commits, no squash) to keep `main`'s log a straight line. Merge with plain git (`git merge --ff-only` after rebasing on `main`, then `git push origin main`), not `gh pr merge` — the GitHub PAT in use here lacks merge permissions on the repo
- **JS/TS lint+format:** oxlint + oxfmt (run automatically on commit via lint-staged)
- **Go lint:** golangci-lint
- **Multi-tenancy:** every tenant-scoped query **must** filter by `org_id`
- **Colors/logo:** never hardcode hex values — use tokens from [`docs/reference/design.md`](docs/reference/design.md)
- **Theme:** the app supports light/dark via `data-theme` on `<html>` — design tokens (`--bg`, `--text`, etc.) already flip per theme, so styling with tokens is theme-safe by default; hardcoding a token's current value isn't (see [EP-10](docs/stories/ep-10-theme.md))

---

## Code quality (Codacy)

CI uploads coverage to Codacy after every push. `CODACY_API_TOKEN` (account-level) is in `apps/api/.env`.

**Fetch current issues before starting a fix session:**

```bash
source apps/api/.env
curl -s -X POST \
  -H "api-token: $CODACY_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"filters":{"categories":[],"levels":[],"languages":[]}}' \
  "https://app.codacy.com/api/v3/analysis/organizations/gh/checkmeup/repositories/checkmeup/issues/search?limit=100" \
  | python3 -c "
import json, sys
issues = json.load(sys.stdin)['data']
priority = [i for i in issues if i['patternInfo']['level'] in ('Error','High','Warning')]
for i in sorted(priority, key=lambda x: x['patternInfo']['level']):
    print(f\"[{i['toolInfo']['name']}][{i['patternInfo']['level']}] {i['filePath']}:{i['lineNumber']}: {i['message'][:100]}\")
"
```

**Triage guide:**

- **TSQLLint** — always ignore, SQL Server rules applied to PostgreSQL migrations
- **Opengrep cookies in `*_test.go`** — ignore, synthetic request cookies in tests intentionally lack HttpOnly/Secure
- **Trivy on `go.mod`** — real; upgrade the flagged dependency or pin a patched version
- **Opengrep/ESLint in production code** — investigate before dismissing
- **Bandit subprocess findings (B603/B404) on `.claude/skills/**/*.py`** — usually fine if the script only ever invokes a literal argv list (never `shell=True`, no externally-supplied input); suppress with `# nosec <code>` **on the exact flagged line**, not a comment above it, plus a one-line rationale
- **Prospector docstring D212/D213 on `.py` files** — these two rules are mutually exclusive (one wants the summary on line 1, the other on line 2); don't chase them back and forth — use a single-line module docstring instead, which sidesteps the multi-line-summary rule pair entirely
- **Lizard/Prospector complexity on `.py` files** — same threshold philosophy as Go handlers; split into small single-purpose functions rather than suppressing

---

## Don't

- Use Playwright (or `npx playwright`) to drive a browser for UI verification — not wanted in this environment. If a UI change needs visual/browser verification and no project run-skill covers it, say so explicitly instead of reaching for Playwright.
- Run `make deploy`, `make ghcr-clean`, `kamal <anything>`, or `docker build`/`buildx` against `config/deploy.yml` — these touch the real production server (`checkmeup.net`) and push to the real GHCR registry. Only the human operator runs these, deliberately, from their own machine. This isn't hypothetical: `make -n deploy` (meant as a dry run) once triggered a real deploy anyway, because the `deploy` target's recipe references `$(MAKE)` (for the post-deploy `ghcr-clean` step) — GNU Make always executes a recipe line that contains a `$(MAKE)`/`${MAKE}` reference, even under `-n`, so it can show what a recursive sub-make would print. `-n` is not a safe way to preview this target.
- Add Redis, a job queue, or any external broker — goroutine workers are intentional ([ADR-001](docs/decisions/001-worker-model.md))
- Use an ORM — sqlc only ([ADR-004](docs/decisions/004-sqlc-over-orm.md))
- Skip `org_id` filters in DB queries — silent data leak across tenants
- Add subdomains for status pages — `/status/:slug` path is intentional ([ADR-005](docs/decisions/005-status-page-same-domain.md))
- Use `Authorization` header for **browser session** auth — the `access_token` httpOnly cookie is the only mechanism there ([ADR-003](docs/decisions/003-auth-jwt-httponly-cookie.md)). The public API (EP-26) is a separate, non-browser mechanism and authenticates via `X-API-Key` instead — never `Authorization`, to keep the two visibly distinct ([ADR-028](docs/decisions/028-api-key-auth-scope.md))
- Use `api.get/post/…` in `auth.init()` — use plain `fetch` there to bypass the 401 interceptor; a 401 on `/me` during init means "not logged in", not a session error
- Switch payment providers — Paddle is the MoR (handles global tax); see [ADR-026](docs/decisions/026-billing-paddle-mor.md) (supersedes [ADR-018](docs/decisions/018-billing-lemonsqueezy-mor.md), which chose LemonSqueezy)
- Write Tailwind classes like `bg-[--token]` or `text-[--token]` for a CSS variable — this Tailwind v4 setup compiles that to invalid CSS (`background-color: --token`, missing `var()`), silently dropping the style. Always write `bg-[var(--token)]`. Found broken across `Button.vue`/`Input.vue`/`Label.vue`/`LandingLayout.vue` during EP-10 — e.g. the sign-in button had no background in *either* theme until fixed
- Build a public feature-request board, voting system, or ticketing queue — feedback goes straight to the founder via the in-app form (Settings, EP-23), email, and GitHub Issues; that's a deliberate choice, not a gap (see `DocsView.vue`'s "Need help?" section)
- Commit directly onto local `main` — always create/switch to a feature branch first, even for doc-only or copy-only changes. `main` tracks `origin/main` and branch protection rejects a direct push anyway, so committing there first just means redoing the work on a branch afterward (create the branch at the existing commits, then `git reset --hard origin/main` to move `main` back before pushing the branch and opening the PR)
