---
title: Deployment Guide
type: reference
status: active
updated: 2026-08-05
tags: [ops, deployment, kamal, hetzner]
---

# Deployment Guide

Kamal 2 on Hetzner CX23 · kamal-proxy SSL · GHCR image registry

---

## How the dev workflow changed

**Nothing changed for day-to-day development.** You still use:

```bash
make dev    # hot-reload API + Vue dev server
make test   # lint + test + coverage
```

What is new:

| Before | After |
|--------|-------|
| No production setup | `make deploy cmd=deploy` ships to prod |
| Single `.env` (dev only) | Two env files: `apps/api/.env` (dev) and `.env` (Kamal/prod secrets) |
| Devcontainer had no Docker CLI | Devcontainer has Docker CLI + buildx + Kamal gem |
| No image registry | Images pushed to `ghcr.io/checkmeup/checkmeup` on every deploy |

The two `.env` files are intentionally separate:

- **`apps/api/.env`** — local dev only; `DATABASE_URL` points to `db:5432` (devcontainer service); committed as `.env.example` template, real file gitignored
- **`.env`** (repo root) — Kamal/prod secrets only; never touches the running dev server; always gitignored

---

## How to deploy

**You don't — CI does.** Since 2026-07-18 (`5bb3292`), merging to `main` deploys
production automatically: CI runs on `main`, then
[`.github/workflows/release.yml`](../../.github/workflows/release.yml) cuts a
semantic-release version and runs `kamal deploy` with secrets held in GitHub.
Check it with `gh run list --workflow=release.yml`.

Everything below is the **manual/break-glass path** — first-ever setup, a
rollback, or deploying when CI can't. It is also the only path this doc
described before auto-deploy existed, so treat any "you must run this to ship"
framing below as historical.

### Every deploy (after `kamal setup` is done)

```bash
git push                              # get commits on main
set -a; source .env; set +a          # load prod secrets into shell
kamal deploy                          # build → push to GHCR → roll out on server
```

Or via Make (sources `.env` automatically):

```bash
make deploy cmd=deploy
```

Kamal will:

1. Clone the repo at HEAD into a temp dir
2. Build the Docker image locally (arm64 native, cross-compiles Go to amd64)
3. Push the image to GHCR tagged with the git SHA
4. SSH to the server, pull the image, start a new container
5. Wait for the health check (`/api/v1/health`) to pass
6. Register the container with kamal-proxy (zero-downtime swap)
7. Stop the old container

### First-ever deploy (`kamal setup`)

```bash
set -a; source .env; set +a
kamal setup
```

This runs `kamal deploy` plus bootstrapping: installs kamal-proxy on the server and starts the PostgreSQL accessory container.

### Other useful commands

```bash
kamal app logs              # tail live logs
kamal accessory logs db     # PostgreSQL logs
kamal app exec --reuse -- /bin/sh   # shell in running container
kamal rollback              # roll back to previous image
kamal lock release          # clear stuck deploy lock (after Ctrl+C)
kamal app details           # show running containers and image SHA
```

---

## Cost breakdown

### What you pay

| Service | Cost | Notes |
|---------|------|-------|
| **Hetzner CX23** | ~€5.99/mo | 2 vCPU, 4 GB RAM, 80 GB SSD, 20 TB/mo traffic. Price varies slightly by region/VAT. |
| **Cloudflare DNS** | €0 | DNS management is free on all Cloudflare plans. |
| **Let's Encrypt SSL** | €0 | kamal-proxy provisions and renews TLS certificates automatically. |
| **Kamal** | €0 | Open-source MIT tool, no SaaS fees. |
| **GHCR (GitHub Container Registry)** | €0 – small | See below. |
| **Domain name** | ~$12/yr | `checkmeup.net` renewal cost. |

**Total recurring: ~€6/mo + domain**

### GHCR storage cost detail

GitHub Container Registry charges for private packages only:

| GitHub plan | Free storage | Free transfer | Overage |
|-------------|-------------|---------------|---------|
| Free | 500 MB | 1 GB/mo | $0.008/GB/day storage, $0.50/GB transfer |
| Pro ($4/mo) | 2 GB | 10 GB/mo | same rates |

Our image is ~150–250 MB compressed. Each deploy creates a new tagged image.
**Kamal does not prune old images automatically** — storage accumulates over time.

Prune old images periodically:

```bash
# On the server (SSH in as deploy):
docker image prune -a --filter "until=720h"   # remove images older than 30 days

# Or via Kamal:
kamal app images              # list images on server
```

In practice: 10 deploys/week × 200 MB = 2 GB/month — stays within free tier unless deploying very frequently or the image grows large.

If the `ghcr.io/checkmeup/checkmeup` package is set to **public** (no auth needed to pull), storage and transfer are completely free. Go to GitHub → Packages → checkmeup → Package settings → Change visibility → Public.

### Hetzner traffic

CX23 includes **20 TB/month** outbound. At MVP scale this is effectively unlimited. Inbound traffic is always free.

### Hidden cost: build time

Every `kamal deploy` builds the Docker image on your **local machine** (devcontainer). The build takes 1–3 minutes. This is CPU and energy on your Mac, not a monetary cost, but worth knowing — especially since:

- Go dependency layer is cached between deploys (only rebuilt when `go.mod` changes)
- Bun dependency layer is cached between deploys (only rebuilt when `package.json` changes)
- Source changes only rebuild the final compile/bundle step (~30s)

---

## Prerequisites (one-time)

### Secrets file

Create `.env` in the repo root (gitignored — never commit this):

```bash
KAMAL_REGISTRY_PASSWORD=ghp_...   # GitHub PAT with write:packages scope
DATABASE_URL=postgres://checkmeup:<url-encoded-password>@checkmeup-db:5432/checkmeup?sslmode=disable
JWT_SECRET=<32+ random bytes, base64>
POSTGRES_PASSWORD=<strong random password>
```

**Tip — avoid URL encoding headaches:** generate passwords without special characters:

```bash
openssl rand -hex 32   # produces only [0-9a-f], safe in URLs
```

If the password does contain `/`, `=`, `+`: percent-encode them (`/`→`%2F`, `=`→`%3D`, `+`→`%2B`) in the `DATABASE_URL` only. `POSTGRES_PASSWORD` is always the raw value.

### GitHub PAT

The PAT needs `write:packages` scope to push to `ghcr.io/checkmeup/checkmeup`.
Create it at: GitHub → Settings → Developer settings → Personal access tokens → Classic.

### Server (done once)

```bash
# On the Hetzner server as root (Debian 12)
apt-get update && apt-get install -y docker.io
useradd -m -s /bin/bash deploy
usermod -aG docker deploy
ufw allow 22 && ufw allow 80 && ufw allow 443 && ufw enable
```

Add your SSH public key to `/home/deploy/.ssh/authorized_keys`.

### Devcontainer Docker socket

The devcontainer mounts `/var/run/docker.sock` from the Mac host (`docker-compose.yml`).
Docker Desktop (re)creates that socket as `root:root 660` — not root:docker, since its VM
has no group that maps to a container's non-root user — so `dev` gets `permission denied`
running `docker`/`kamal` commands until it's chmod'd wide open.

This isn't a one-time fix: the socket resets to `660` any time Docker Desktop restarts,
the Mac reboots, or the devcontainer is rebuilt. `devcontainer.json`'s
`postStartCommand: sudo chmod 666 /var/run/docker.sock` re-applies the fix on every
container start (the `Dockerfile` pre-authorizes this exact `sudo chmod` via
`/etc/sudoers.d/dev-docker` so it needs no password). If it's still broken, run that same
command by hand inside the container: `sudo chmod 666 /var/run/docker.sock`.

**It's a shared, host-wide resource.** The socket is one file on the host — every
devcontainer across every project bind-mounts the *same* inode. Whichever container chmods
it first fixes access for all of them until the next Docker Desktop restart. So this isn't
checkmeup-specific: any other project's devcontainer hitting the same error needs its own
copy of both pieces (the sudoers rule in its `Dockerfile`, the `postStartCommand` in its
`devcontainer.json`) to self-heal independently, rather than depending on checkmeup's
container happening to start first.

---

## Architecture on the server

```text
Internet
    │
    ▼
kamal-proxy  :80 (redirect) / :443 (TLS)
    │  Host: checkmeup.net
    ▼
checkmeup-web  (Docker, :8080)
    │  kamal bridge network
    ▼
checkmeup-db  (Postgres 17 Alpine, :5432)
```

- `kamal-proxy` provisions and renews TLS via Let's Encrypt; stores cert on the host
- App and db share the `kamal` Docker bridge — app connects to `checkmeup-db:5432`
- PostgreSQL binds to `127.0.0.1:5432` on the host (not reachable from internet; UFW blocks 5432)
- The Go binary runs `goose` migrations on every startup before accepting traffic
- kamal-proxy health-checks `/api/v1/health` before cutting traffic to a new container

---

## Known gotchas

### `host.docker.internal` doesn't exist on Linux Docker

On macOS, Docker Desktop auto-registers `host.docker.internal` pointing to the host.
On Linux (the Hetzner server), it doesn't exist unless you pass `--add-host` explicitly.

Since both app and db are Kamal-managed containers on the `kamal` network, connect using
the **accessory container name**: `checkmeup-db`. This is `<service>-<accessory-name>`.

### Cross-platform Docker builds (arm64 Mac → amd64 server)

Kamal targets `linux/amd64` for the server. Running Bun/Vite/Node.js under QEMU emulation
(to produce amd64 output) is 10-50× slower and can hang indefinitely.

**Fix:** `FROM --platform=${BUILDPLATFORM}` on build stages — they run natively on the
build machine (arm64). Go cross-compiles to amd64 via `GOOS=linux GOARCH=amd64`; Vite
output is platform-independent JS/CSS. Only the final runtime `FROM debian:bookworm-slim`
is amd64.

### Bun lockfile is platform-specific

`bun install --frozen-lockfile` fails when the lockfile was generated on arm64 but bun
runs on a different arch — optional native packages differ. Drop the flag in Docker.

### `vue-tsc` fails on cold cache in Docker

`vue-tsc -b` uses `.tsbuildinfo` incremental cache. Without it (fresh Docker build), it
can fail or behave inconsistently. In Docker, call `./node_modules/.bin/vite build`
directly — TypeScript type-checking belongs in CI, not in the image build step.

### Deploy lock stuck after Ctrl+C

Interrupting `kamal setup` or `kamal deploy` leaves a lock on the server.
Clear it: `kamal lock release`

### Stray content in `.kamal/secrets`

The file must contain only `KEY=$ENV_VAR` lines. Any extra line (raw token, blank line with
content) causes Kamal to pass an empty `-p` flag to `docker login`, which fails.

---

## Secrets management

| Secret | Location | How to rotate |
|--------|----------|---------------|
| `KAMAL_REGISTRY_PASSWORD` | `.env` only | Regenerate PAT on GitHub → update `.env` |
| `JWT_SECRET` | `.env` only | Update `.env`; all active sessions expire on next deploy |
| `POSTGRES_PASSWORD` | `.env` + server | Changing requires: stop app, `docker exec` into db container, `ALTER USER`, update `.env`, redeploy |
| `DATABASE_URL` | `.env` only | Update if host/port/user/password changes |

`.kamal/secrets` re-exports from the shell environment — it never contains literal values:

```bash
KAMAL_REGISTRY_PASSWORD=$KAMAL_REGISTRY_PASSWORD
DATABASE_URL=$DATABASE_URL
JWT_SECRET=$JWT_SECRET
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
```

Verify nothing sensitive is tracked:

```bash
git check-ignore -v .env .kamal/secrets   # both should appear as gitignored
git log --all -- .env .kamal/secrets      # should be empty (no history)
```
