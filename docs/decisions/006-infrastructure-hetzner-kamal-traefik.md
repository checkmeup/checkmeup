# ADR-006: Hetzner VPS + Kamal + Traefik instead of managed cloud

**Status:** Accepted  
**Date:** 2026-06-13

## Context

Hosting options considered:

- Managed cloud (AWS ECS/Fargate, GCP Cloud Run, Fly.io, Render) — higher cost, less control
- Self-managed VPS with manual deploy scripts
- Self-managed VPS with a container-native deploy tool (Kamal)

## Decision

Single Hetzner CX23 VPS (€3.49/mo). Kamal deploys the Docker containers. Traefik handles reverse proxying and TLS certificate provisioning via Let's Encrypt.

## Consequences

- **Cost:** ~€3.49/mo vs. $20–50+/mo for comparable managed options — meaningful at zero revenue
- **Simplicity:** Kamal wraps `docker run` with zero-downtime deploys and health checks; no Kubernetes complexity
- **Single point of failure:** one server means no built-in redundancy; acceptable for MVP, revisit when uptime SLA is offered to customers
- **Scaling ceiling:** vertical scaling (Hetzner upgrade) buys significant headroom before horizontal scaling is needed; stateful worker model (see ADR-001) also limits horizontal scaling options
- **Ops ownership:** we manage OS updates, disk, and server health rather than a platform doing it
