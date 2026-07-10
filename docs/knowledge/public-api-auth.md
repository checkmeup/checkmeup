---
title: Public API & API-Key Authentication
type: knowledge
status: current
updated: 2026-07-10
tags: [architecture, auth, security, backend]
scope: apps/api/internal/middleware/apikey.go, internal/handler/api_keys.go, internal/handler/public_status.go, /api/v1/public route group
superseded_by:
---

# Public API & API-key authentication

**Investigated:** 2026-07-10
**Scope:** the `X-API-Key`-authenticated `/api/v1/public` route group (EP-26) — a second, independent auth mechanism alongside the browser session cookie.

## Summary

`X-API-Key` is a deliberately separate mechanism from the `access_token` httpOnly cookie used everywhere else ([ADR-028](../decisions/028-api-key-auth-scope.md), amending [ADR-003](../decisions/003-auth-jwt-httponly-cookie.md)'s "no `Authorization` header" rule to scope it to browser sessions only). The two are kept visibly and mechanically distinct — different header name (never `Authorization`), different middleware, different route group, so a request can't be accidentally validated against the wrong one. As of 2026-07-10 the entire public API surface is five read-only `GET .../status` endpoints; there is no write endpoint yet.

## Findings

1. **Key generation is one-way, same pattern as passwords.** `APIKeyHandler.CreateAPIKey` (`api_keys.go:61`) generates a random key via `generateAPIKey`, prefixed `cmu_live_` for log/secret-scanner recognition (`apiKeyPrefixLen = 16` = 9-char prefix + 7 hex chars, stored separately as `KeyPrefix` for display in the UI's key list). Only `hashToken(raw)` (SHA-256, same helper as `refresh_tokens.token_hash`) is persisted; the raw key is returned once in the `CreateAPIKey` response body and never retrievable again — `ListAPIKeys` only ever returns `KeyPrefix`, not the full key.

2. **`RequireAPIKey` (`internal/middleware/apikey.go:27`) is the only place a raw key is validated.** It reads `X-API-Key`, hashes it, and looks up `GetActiveAPIKeyByHash` — a query that only returns non-revoked keys, so a revoked key fails the same way an unknown one does (`401 invalid_api_key`), not distinguished in the response. The resolved `key.OrgID` (from the DB row, never from the request) is injected into context via `OrgIDFromAPIKey`, mirroring `ClaimsFrom`/session auth's shape closely enough that a handler can't easily tell which mechanism authenticated it just by looking at the context value's presence — every current public-API handler is written specifically against `OrgIDFromAPIKey`, not a shared "any auth" accessor, so there's no accidental cross-acceptance today.

3. **`last_used_at` is updated off the request path.** `touchLastUsed` (`apikey.go:67`) runs in its own goroutine with a 5s-timeout background context, explicitly so a slow write never delays the caller's actual request — errors are silently dropped (`_ = queries.TouchAPIKeyLastUsed(...)`), since this is a display-only timestamp, not something request handling depends on.

4. **The route group, not the middleware, is what makes API-key auth impossible to reach with a session cookie or vice versa.** `server.go:252`'s `/public` route group is a sibling of the `RequireAuth`-protected group, not nested inside it — `r.Use(apimiddleware.RequireAPIKey(...))` is the *only* auth middleware in that subtree. A session-cookie-only request has no `X-API-Key` header and gets `401`; an API-key request never passes through `RequireAuth` at all. Rate limiting is also key-scoped, not IP- or org-scoped: `apiKeyRateKey` (`server.go:293`) keys the limiter on the raw `X-API-Key` header value itself (60/min), so a single leaked or misbehaving key is bounded regardless of how many keys the org has or what network it's called from — deliberately different from `authOrgKey`'s org-scoped limiting used under `RequireAuth`.

5. **Scope (read-only vs. read-write) is designed but not implemented.** ADR-028 describes a `scope` column and `403` enforcement for a read-only key hitting a non-`GET` route — the first shipped slice (EP-26 US-2601/2602/2604) skipped both because every current public-API route is a `GET`, so there was nothing to enforce yet. **This is a known, explicitly-flagged gap**: `RequireAPIKey` does not check scope at all right now, so if a write endpoint is ever added under `/api/v1/public` without first adding the `scope` column and enforcement, every existing key — including ones a user expects to be read-only — would be able to call it. The ADR's own "Implementation status" section and [`docs/decisions/backlog.md`](../decisions/backlog.md) both flag this must land *before*, not after, the first write endpoint.

6. **Every public-API handler is a thin read**: `GetCronStatus`/`GetUptimeStatus`/`GetSSLStatus`/`GetDomainStatus`/`GetPortStatus` (`public_status.go`) each resolve the monitor by `id` + the API-key-derived `org_id` (same tenant-isolation pattern as everywhere else — `org-id-audit` skill would catch a regression here same as any other handler) and return its current status. No mutation, no list/enumerate-all-monitors endpoint exists yet — a caller needs to already know a monitor's ID.

## Follow-ups

- The `scope` column + read-only/`403` enforcement (finding 5) is real, tracked debt — not addressed here since it requires a migration and isn't triggered by anything in the current codebase (no write endpoint exists yet to expose the gap). Should land before EP-26's next slice adds one.
