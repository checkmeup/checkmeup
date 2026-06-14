# ADR-006: Hetzner VPS + Kamal 2 + kamal-proxy

**Status:** Accepted  
**Date:** 2026-06-13  
**Updated:** 2026-06-14 (Traefik → kamal-proxy; Kamal 2 uses its own built-in proxy)

## Context

Hosting options considered:

- Managed cloud (AWS ECS/Fargate, GCP Cloud Run, Fly.io, Render) — higher cost, less control
- Self-managed VPS with manual deploy scripts
- Self-managed VPS with a container-native deploy tool (Kamal)

## Decision

Single Hetzner CX23 VPS (€5.99/mo, ARM64). Kamal 2 deploys Docker containers and manages
zero-downtime rolling deploys. **kamal-proxy** (Kamal 2's bundled reverse proxy, not Traefik)
handles SSL termination via Let's Encrypt and proxies traffic to the app container.

The PostgreSQL database runs as a **Kamal accessory** container on the same host, on the
`kamal` Docker network. The app container reaches it using the accessory container name
(`checkmeup-db`), not via the host.

Image registry: GitHub Container Registry (`ghcr.io/checkmeup/checkmeup`), pushed from the
devcontainer during deploy.

## Consequences

- **Cost:** ~€5.99/mo vs. $20–50+/mo for comparable managed options
- **Simplicity:** Kamal wraps `docker run` with zero-downtime deploys and health checks; no Kubernetes complexity
- **Single point of failure:** one server means no built-in redundancy; acceptable for MVP
- **kamal-proxy vs Traefik:** Kamal 2 ships its own proxy; Traefik is not needed and not installed
- **Container networking:** app and db communicate over the `kamal` bridge network by container name — `host.docker.internal` does **not** resolve on Linux Docker without explicit `--add-host`
- **Cross-platform builds:** devcontainer is arm64 (Apple Silicon), server is amd64. Both build stages use `--platform=${BUILDPLATFORM}` so they run natively on the build machine. Go cross-compiles to amd64 via `GOOS=linux GOARCH=amd64`; Vite output is platform-independent JS/CSS.
