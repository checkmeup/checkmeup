# Hosting, deploys and versioning

Date: 2026-10-13

Production is one Hetzner CX23 VPS. Kamal 2 deploys the single Docker image (frontend and backend built together) with zero-downtime swaps; kamal-proxy terminates TLS with Let's Encrypt. PostgreSQL runs as a Kamal accessory container on the same host; the app reaches it as `checkmeup-db` on the `kamal` Docker network. Images go to `ghcr.io/checkmeup/checkmeup`.

Merging to `main` deploys: after CI passes on `main`, `.github/workflows/release.yml` runs semantic-release, which reads the Conventional Commits since the last tag and creates a tag and a GitHub release (the GitHub Releases page is the changelog; no `CHANGELOG.md`, no version in `package.json`), then runs `kamal deploy` with the secrets held in GitHub. Running a deploy by hand is the operator's break-glass path only. Every build stamps its version from `git describe --tags --always --dirty` into the Go binary and the Vite build; the API reports it on `/api/v1/health`.

## Considered options

- Managed platforms (ECS/Fargate, Cloud Run, Fly.io, Render): $20–50+ a month against about €6, for redundancy the product doesn't need yet.
- Traefik as the proxy: Kamal 2 ships kamal-proxy.
- Deploying by hand with `kamal deploy` (until 2026-07-18): replaced by deploying from CI on every merge.

## Consequences

- One server is a single point of failure.
- Nothing in front of the server absorbs a volumetric attack; whether to add a CDN/WAF is open (see the Inbox).
- One version covers the whole monorepo; deploying `apps/api` and `apps/web` separately would mean revisiting this.
