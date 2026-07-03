# ADR-028: API key auth for the public API — amends ADR-003 scope

**Date:** 2026-07-03
**Status:** Accepted — amends [ADR-003](003-auth-jwt-httponly-cookie.md)

---

## Context

[EP-26](../stories/ep-26-public-api-keys.md) needs a programmatic API so third parties — not just the founder's own scripts — can integrate with checkmeup (e.g. pull monitor status) without a browser session.

[ADR-003](003-auth-jwt-httponly-cookie.md) states "No `Authorization` header — the cookie is sent automatically by the browser," and CLAUDE.md repeats this as an unconditional "Don't." That rule's rationale is specific to browser session auth: `HttpOnly` stops JavaScript from reading the token (XSS), and `SameSite=Strict` stops cross-site forgery of an ambient browser credential (CSRF). Non-browser clients — scripts, CI pipelines, third-party servers — have no cookie jar, no ambient credential a malicious page could ride on, and no DOM for an XSS payload to run in. The threats ADR-003 defends against don't apply to them, but its literal wording would still block any header-based token for them, leaving no way to build a public API at all.

Because the audience here is explicitly third parties, blast radius from a leaked key matters more than it would for a purely internal integration — revocation, scoping, and rate limiting aren't optional polish.

## Decision

Scope ADR-003's "no `Authorization` header" rule to **browser session auth only**. Introduce a second, independent auth mechanism for the public API: a dedicated **`X-API-Key`** header — deliberately not `Authorization`, so the two mechanisms stay visually and mechanically distinct in code, logs, and docs, and a request accidentally validated against the wrong mechanism is obvious at a glance rather than silently accepted.

Implementation shape (full detail in [EP-26](../stories/ep-26-public-api-keys.md)):

- Keys are generated per-org, shown once at creation, stored only as a hash — same pattern as `password_hash` / `refresh_tokens.token_hash`
- Key format carries a recognizable prefix (`cmu_live_...`) for log/secret-scanner recognition
- Each key has a scope — read-only or read-write — set at creation and not editable (revoke + recreate instead)
- Keys are independently revocable; revocation takes effect on the next request
- A valid key resolves to its owning org for every `org_id`-scoped query, same multi-tenancy rules as session auth ([ADR-002](002-multi-tenancy.md))
- Invalid/revoked key → `401` in the existing error format ([ADR-008](008-api-error-format.md)); read-only key on a non-`GET` → `403`
- Per-key rate limiting, consistent with the existing rate-limiting pattern ([ADR-013](013-rate-limiting.md)); rate-limited responses → `429` + `Retry-After`
- Public docs page enumerating available endpoints and request/response shapes

## Consequences

- ADR-003 no longer reads as an unconditional ban on non-cookie auth — its own Status line and CLAUDE.md's "Don't" line are updated to reflect the narrower scope (browser session auth only)
- Two parallel auth mechanisms now exist in the codebase; middleware must keep them clearly separated — no endpoint should accept either without that being an explicit design choice
- API keys carry their own security surface (leakage, rotation, scope creep) distinct from session auth; revocation and read-only scoping exist specifically to bound blast radius for third-party use
- Once shipped, the public API is a real external contract — endpoint changes need the same versioning discipline as anything else under [ADR-007](007-api-versioning.md)

## Implementation status (as of the first shipped slice)

The first cut (read-only monitor status, [EP-26](../stories/ep-26-public-api-keys.md) US-2601/US-2602/US-2604 plus part of US-2605) shipped **without** the `scope` column and without the read-only/`403`-on-non-`GET` enforcement described above (US-2603) — every route currently under `/api/v1/public` is a `GET`, so there was nothing to scope yet, and skipping it avoided a schema column with no behavior attached. A security review flagged this explicitly: the gap is harmless today, but this ADR's own text reads as if scope enforcement already exists, which could lead a future author to add a write endpoint under `/api/v1/public` assuming `RequireAPIKey` already blocks non-`GET` requests on read-only keys — it doesn't. See the [decision backlog](backlog.md) — the `scope` column and enforcement must land *before*, not after, the first write endpoint.
