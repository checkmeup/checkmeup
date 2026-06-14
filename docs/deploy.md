# Deployment Guide

Kamal 2 on Hetzner CX23 · kamal-proxy SSL · GHCR image registry

---

## Prerequisites (one-time)

### Secrets file

Create `.env` in the repo root (gitignored — never commit this):

```
KAMAL_REGISTRY_PASSWORD=ghp_...   # GitHub PAT with write:packages scope
DATABASE_URL=postgres://checkmeup:<url-encoded-password>@checkmeup-db:5432/checkmeup?sslmode=disable
JWT_SECRET=<32+ random bytes, base64>
POSTGRES_PASSWORD=<strong random password>
```

**URL-encode the database password** — special characters in a PostgreSQL URL userinfo
segment must be percent-encoded (`/` → `%2F`, `=` → `%3D`, `+` → `%2B`).  
Generate a safe password to avoid this: `openssl rand -hex 32`

### GitHub PAT

The PAT needs `write:packages` scope to push images to `ghcr.io/checkmeup/checkmeup`.
Create it at: GitHub → Settings → Developer settings → Personal access tokens → Classic.

### Server (done once)

```bash
# On the Hetzner server as root
apt-get install -y docker.io
usermod -aG docker deploy   # or create deploy user first
ufw allow 22 && ufw allow 80 && ufw allow 443 && ufw enable
```

### Devcontainer Docker socket

The devcontainer mounts `/var/run/docker.sock` from the host. After rebuilding the
devcontainer, the socket is owned by `root:root` (mode 660). Fix it once per session
from your **Mac terminal** (not the devcontainer):

```bash
docker exec -u root <container-id> chmod 666 /var/run/docker.sock
```

After rebuilding the devcontainer with the updated Dockerfile, `postStartCommand` does
this automatically via a sudoers rule.

---

## Deploy commands

Always source `.env` first so Kamal picks up the secrets:

```bash
set -a; source .env; set +a
```

Or use the Makefile shortcut which does this for you:

```bash
make deploy cmd=setup    # first-time server bootstrap
make deploy cmd=deploy   # rolling deploy
make deploy cmd=rollback # roll back to previous version
```

### First deploy

```bash
set -a; source .env; set +a
kamal setup      # installs kamal-proxy on server, starts db accessory, deploys app
```

### Subsequent deploys

```bash
git push         # ensure commits are on main
set -a; source .env; set +a
kamal deploy     # builds image, pushes to GHCR, rolls out on server
```

### Useful Kamal commands

```bash
kamal app logs            # tail app logs
kamal app exec --reuse -- /api --version
kamal accessory logs db   # PostgreSQL logs
kamal lock release        # clear a stuck deploy lock (after Ctrl+C)
kamal app details         # running container info
```

---

## Architecture on the server

```
Internet → kamal-proxy (:80/:443) → checkmeup-web (Docker, :8080)
                                            ↓ (kamal network)
                                    checkmeup-db (Postgres 17)
```

- `kamal-proxy` handles TLS via Let's Encrypt (certificate stored on host)
- App container and db accessory share the `kamal` Docker bridge network
- App reaches db at `checkmeup-db:5432` (container name resolution)
- PostgreSQL port is bound to `127.0.0.1:5432` on the host (not publicly exposed; UFW blocks 5432 anyway)
- Go binary runs goose migrations on every startup before accepting traffic

---

## Known gotchas

### `host.docker.internal` doesn't exist on Linux Docker

On macOS, Docker Desktop auto-registers `host.docker.internal`. On Linux (the server),
it doesn't unless you pass `--add-host`. Since both app and db are Kamal-managed containers
on the `kamal` network, use the **container name** as the hostname: `checkmeup-db`.

### Cross-platform Docker builds (arm64 dev → amd64 server)

Kamal targets `linux/amd64` for the server. Running JavaScript tooling (Bun/Vite/Node.js)
under QEMU emulation is 10-50× slower and can hang indefinitely.

**Fix:** use `--platform=${BUILDPLATFORM}` on the build stages so they run natively on
the build machine (arm64), and cross-compile the Go binary with `GOOS=linux GOARCH=amd64`.
The Vite output (JS/CSS) is platform-independent and works in the amd64 runtime image.

### Bun lockfile is platform-specific

`bun install --frozen-lockfile` fails when the lockfile was generated on arm64 but bun
runs under amd64 emulation — optional platform-native packages differ. Drop the flag in Docker.

### `bun --cwd <path> run <script>` flag parsing

Bun parses `--cwd path` (space-separated) differently in some contexts. Use
`cd path && bun run script` or `--cwd=path` (equals form) in Dockerfiles.

### `vue-tsc` fails on cold cache in Docker

`vue-tsc -b` (the TypeScript type-check step) behaves inconsistently without the
`.tsbuildinfo` incremental cache that exists in the devcontainer. In Docker, call
`vite build` directly — type-checking is a CI responsibility, not a build step.

### Deploy lock stuck after Ctrl+C

Interrupting `kamal setup` or `kamal deploy` leaves a lock file on the server.
Release it with: `kamal lock release`

### Stray line in `.kamal/secrets`

The secrets file must contain only `KEY=$ENV_VAR` lines. Any extra content (e.g., a raw
token on its own line) causes Kamal to pass an empty value as a CLI flag (`-p` with no
argument), which fails with `flag needs an argument`.

---

## Secrets management

| Secret | Where | How to rotate |
|--------|-------|---------------|
| `KAMAL_REGISTRY_PASSWORD` | `.env` only | Regenerate PAT on GitHub; update `.env` |
| `JWT_SECRET` | `.env` only | Update `.env`; existing sessions invalidated on next deploy |
| `POSTGRES_PASSWORD` | `.env` + server db container | Must match what PostgreSQL was initialized with; changing requires manual db steps |
| `DATABASE_URL` | `.env` only | Update if server, port, user, or password changes |

The `.kamal/secrets` file just re-exports from the environment:
```
KAMAL_REGISTRY_PASSWORD=$KAMAL_REGISTRY_PASSWORD
DATABASE_URL=$DATABASE_URL
JWT_SECRET=$JWT_SECRET
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
```

Never commit `.env` or `.kamal/secrets`. Both are gitignored. Verify with:
```bash
git check-ignore -v .env .kamal/secrets
```
