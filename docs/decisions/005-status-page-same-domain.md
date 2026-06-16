# ADR-005: Status pages served at /status/:slug, no subdomain

**Status:** Accepted  
**Date:** 2026-06-13

## Context

Status pages could be served at:

- A separate subdomain per org: `acme.checkmeup.net`
- A fixed path on the main domain: `checkmeup.net/status/acme`
- A fully custom domain: `status.acme.com` (white-label)

## Decision

MVP serves status pages at `checkmeup.net/status/:slug`. Custom domain mapping (white-label) is deferred to a post-MVP feature for the Enterprise tier.

## Consequences

- **Simpler DNS/TLS:** no wildcard certificate or per-tenant subdomain provisioning needed on MVP
- **Simpler routing:** a single route pattern handles all status pages
- **Branding limitation:** the URL shows `checkmeup.net`, which is acceptable for Hobby/Solo tiers but a gap for Enterprise customers — addressed by the deferred custom domain feature
- **Custom domain path:** when implemented, Traefik can route `status.acme.com` to the same handler with a `Host` header match; a `domains` table maps custom hostnames to org slugs
