# ADR-007: /api/v1/ URL prefix for API versioning

**Status:** Accepted  
**Date:** 2026-06-13

## Context

Options considered:

- No versioning — break freely while pre-1.0
- URL prefix: `/api/v1/`
- `Accept` header versioning: `Accept: application/vnd.checkmeup.v1+json`

## Decision

All API routes are prefixed with `/api/v1/`. The frontend and any future clients always include the version in the path.

## Consequences

- **Zero cost now:** one Chi router group, no extra logic
- **Clean migration path:** `/api/v2/` routes can coexist with v1 when breaking changes are needed — old clients keep working
- **Header versioning rejected:** invisible in browser DevTools, harder to test with curl, no meaningful benefit over URL prefix for a small team
