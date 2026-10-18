# Local development: toolchain and tasks in mise, services in Docker

Date: 2026-10-18

Development runs on the host; there is no devcontainer. Languages and command-line tools (Go, Bun, Node, golangci-lint, sqlc, air, goose, gcov2lcov, the Codacy linters) are pinned in `mise.toml` at the versions CI pins, so local tools can't quietly drift from CI's. The project's commands are mise tasks in the same file (`mise tasks` lists them); `mise run setup` installs the tools. CI runs `mise run test` too, with mise as the task runner only: its tools come from its own setup steps, and `MISE_AUTO_INSTALL` is off there.

External services run in Docker from the root `docker-compose.yml`, each with its own mise task. Today that is only Postgres (`mise run db`), on CI's `postgres:17-alpine` image with CI's credentials and port, so one `DATABASE_URL` on `localhost:5432` works locally, in CI and as the tests' fallback. A service added later (a mail catcher, an S3 stand-in) gets a compose service and a mise task the same way.

Git hooks are the baloo plugin's, configured in `.claude/baloo.yml`: on commit they check Conventional Commits (lines up to 140), no AI co-author, no secrets, no conflict markers, no large files and no stale ADR Date, then run lint-staged (oxlint and oxfmt); on push they refuse merge commits. A new session writes them into `.git/hooks`.

Kamal stays out of `mise.toml`: only the operator runs it, on the break-glass deploy path, and installs it with `gem install kamal`.

## Considered options

- husky and commitlint (until 2026-10-13): replaced by baloo's hooks, which also cover secrets and AI attribution.
- A push-time secrets scan in `.claude/hooks/pre_push_guard.py` and the `secrets-scan` skill (until 2026-10-13): replaced by baloo's commit-time secrets check; `pre_push_guard.py` keeps only the guard against force-pushing `main`.
- A Makefile for the commands (until 2026-10-17): mise tasks keep tools and commands in one file, take named arguments (`mise run migrate-create <name>`) and list themselves with descriptions. Make's `-n` also runs any recipe line that names `$(MAKE)`, which once deployed for real (docs/incidents/2026-07-02-make-dry-run-triggered-deploy.md).
- Devcontainer (until 2026-10-13): dropped for a plain host environment. Its Dockerfile installed several tools at `@latest`, so local versions could differ from CI's.
- Native Postgres from mise's `postgres` tool: rejected. Docker runs the exact image CI uses and keeps the data in a disposable volume.
- Homebrew `postgresql@17`: rejected. It is one machine-wide server outside the repo's version pinning.

## Consequences

- A Docker daemon (OrbStack on macOS, Docker Engine or Docker Desktop on Linux and WSL) must be running before the database is usable.
- `mise run db` takes host port 5432, so a Postgres installed directly on the host conflicts with it.
