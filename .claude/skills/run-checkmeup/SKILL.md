---
name: run-checkmeup
description: Build, launch, and drive checkmeup (apps/api Go backend + apps/web Vue frontend) end-to-end — sign up, get a session cookie, create a monitor, verify the frontend proxy — using the curl-based smoke.sh driver. Use when asked to "run checkmeup", "start the app", "verify this change works", or "smoke test the app".
---

# Run checkmeup

Paths below are relative to the repo root. The driver is
`.claude/skills/run-checkmeup/smoke.sh` — a `curl`-only smoke test, not a
browser. **This project's CLAUDE.md explicitly disallows Playwright/
`npx playwright` for UI verification** — there is deliberately no
browser driver here. If a change needs actual visual verification, say so
explicitly and ask a human to check in a real browser, rather than
reaching for Playwright.

## Prerequisites

`make setup` once (installs `go`, `bun`, `goose`, … from `mise.toml` and
creates `apps/api/.env` from its example), then `make db` to start
Postgres on `localhost:5432` (`docker-compose.yml`). `apps/web/.env` must
also exist (copy from `apps/web/.env.example` if missing).

## Run (agent path)

```bash
.claude/skills/run-checkmeup/smoke.sh start   # migrate db, build+launch api (:8080) and web (:5173), wait for both
.claude/skills/run-checkmeup/smoke.sh check   # sign up, get session cookie, create a cron monitor, list it back, verify web->api proxy
.claude/skills/run-checkmeup/smoke.sh stop    # kill both
```

`start` logs to `/tmp/checkmeup-smoke/{api,web}.log` and tracks PIDs in
`/tmp/checkmeup-smoke/*.pid` so `stop` can clean up reliably. `check` can
be re-run any number of times against an already-started stack — each
run signs up a fresh timestamped email, so it never collides with a
previous run's data.

## Direct invocation (what most PRs actually need)

Most changes touch either a Go handler/query or a Vue component in
isolation — the full `smoke.sh` flow is for confirming the whole stack
still wires together, not the first thing to reach for on every PR.

```bash
cd apps/api && go test -count=1 ./...      # Go backend — needs `make db`
cd apps/web && bun run test                # Vue frontend (Vitest)
```

## Run (human path)

```bash
make dev
```

Runs `turbo run dev` — both apps with hot reload (air for Go, Vite for
Vue). Opens nothing itself; visit `http://localhost:5173`. This is the
normal local dev loop, not agent-drivable without a browser.

## Gotchas

- **An `apps/api/.env` from before the devcontainer was removed still
  points `DATABASE_URL` at host `db`**, which only resolved inside that
  container's network. On the plain host it must be `localhost:5432` —
  copy the line from `apps/api/.env.example`.
- **Don't `pkill -f` a pattern to stop a process you started with `&` in
  Bash.** Each Bash tool call runs through a wrapper shell whose command
  line contains the text you just wrote, so a pattern like `api-test|vite`
  matches the wrapper too, and killing it kills the tool call itself
  (exit 144, silent). Use `smoke.sh stop`, which tracks real PIDs from
  launch time and never greps.
- **No screenshot / browser step.** The run-skill authoring process
  normally wants a screenshot as proof of a working GUI; this project's
  explicit "don't use Playwright" rule means that step is replaced by the
  curl-based `check` flow above (session cookie + real resource creation)
  as the closest available proof of an end-to-end working app.
- The Vite proxy only forwards `/api` and `/status/` (trailing slash
  matters — see the comment in `apps/web/vite.config.ts`) to `:8080`;
  everything else is served by Vite itself.

## Troubleshooting

- `smoke.sh start` fails at the `wait_for` step and dumps the last 40
  log lines — check `/tmp/checkmeup-smoke/api.log` or `web.log` directly
  for the full trace if that's not enough.
- If port 8080 or 5173 is already bound from a previous unclean session,
  `smoke.sh stop` won't find it (no PID file) — find the holder with
  `lsof -ti :8080` (or `:5173`) and stop it by that PID.
