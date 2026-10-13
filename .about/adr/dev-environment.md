# Local development: toolchain in mise, services in Docker

Date: 2026-10-13

Development runs on the host; there is no devcontainer. Languages and command-line tools (Go, Bun, Node, golangci-lint, sqlc, air, goose, gcov2lcov, the Codacy linters) are pinned in `mise.toml` at the versions CI pins, so local tools can't quietly drift from CI's. `make setup` installs them.

External services run in Docker from the root `docker-compose.yml`, each with its own `make` target. Today that is only Postgres (`make db`), on CI's `postgres:17-alpine` image with CI's credentials and port, so one `DATABASE_URL` on `localhost:5432` works locally, in CI and as the tests' fallback. A service added later (a mail catcher, an S3 stand-in) gets a compose service and a `make` target the same way.

Kamal stays out of `mise.toml`: only the operator runs it, on the break-glass deploy path, and installs it with `gem install kamal`.

## Considered options

- Devcontainer (until 2026-10-13): dropped for a plain host environment. Its Dockerfile installed several tools at `@latest`, so local versions could differ from CI's.
- Native Postgres from mise's `postgres` tool: rejected. Docker runs the exact image CI uses and keeps the data in a disposable volume.
- Homebrew `postgresql@17`: rejected. It is one machine-wide server outside the repo's version pinning.

## Consequences

- Docker (OrbStack) must be running before the database is usable.
- `make db` takes host port 5432, so a Postgres installed directly on the host conflicts with it.
