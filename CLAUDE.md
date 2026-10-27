# checkmeup.net — developer monitoring platform

Cron, uptime, SSL expiry, domain expiry, port (TCP) and DNS record monitors, with alerts and public status pages.

Read the relevant ADR in [`.about/adr/`](.about/adr/) before changing what it covers; [`docs/INDEX.md`](docs/INDEX.md) maps the rest of `.about/` and `docs/`. Commands are mise tasks (`mise tasks`): `mise run setup` once, `mise run db` before `dev` or `test`.

## Conventions

- **Commits:** conventional commits, header up to 140 characters.
- **Merging to `main`:** rebase-only with plain git (`git merge --ff-only`, `git push origin main`), never `gh pr merge` (the PAT here can't). Every merge deploys to production from CI ([ADR infrastructure](.about/adr/infrastructure.md)), so a merged change never "still needs a deploy"; watch `gh run list --workflow=release.yml`.
- **PRs and issues in other people's repos** (awesome lists, directories): read their CONTRIBUTING and any AI policy before opening one, and leave the "Generated with Claude Code" line out of it. ivbeg/awesome-status-pages closed PR 252 on 2026-10-27 for that line and bans AI-generated PRs.
- **Capitalization:** "Checkmeup" in prose; the domain `checkmeup.net` and code identifiers (`@checkmeup/web`) stay lowercase.
- **TypeScript stays on 6.x** (`^6.0.3`): `vue-tsc` can't bootstrap against 7 yet, whatever its peerDeps say. If `bun update --latest` floats it to 7, re-pin the latest 6.x.

## Don't

- Run deploy commands (`kamal`, `mise run deploy`, `mise run ghcr-clean`, `docker build` for `config/deploy.yml`), not even as a dry run; hooks block them.
- Skip the `org_id` filter on a tenant-scoped query ([ADR database](.about/adr/database.md)).
- Add Redis, a job queue or broker ([ADR workers](.about/adr/workers.md)), an ORM ([ADR database](.about/adr/database.md)), a payment provider besides Paddle ([ADR billing](.about/adr/billing.md)), or status-page subdomains ([ADR status-pages](.about/adr/status-pages.md)).
- Authenticate requests to this app with the `Authorization` header: browser sessions authenticate only by the `access_token` cookie, the public API by `X-API-Key` ([ADR auth](.about/adr/auth.md)).
- Use `api.get/post/…` in `auth.init()` — use plain `fetch`, past the 401 interceptor: a 401 on `/me` there means "not logged in".
- Hardcode a hex color or a token's current value — use the tokens in [`docs/reference/design.md`](docs/reference/design.md), which flip with the theme. Write `bg-[var(--token)]`, never `bg-[--token]`, which this Tailwind v4 setup compiles to invalid CSS.
- Expect an eagerly imported module (even one export) to also be its own lazy chunk — Rollup bundles all of it eagerly. Keep cheap metadata in a separate file nothing lazy imports from, as `apps/web/src/blog/postsMeta.ts` does for `posts/*.ts`.
- Build a public feature-request board, voting or ticketing — feedback reaches the founder privately (in-app form, email, GitHub Issues) on purpose.
- Commit directly onto local `main` — branch first (a hook blocks it). If it happens anyway: create the branch at those commits, `git reset --hard origin/main`, then push the branch and open the PR.
