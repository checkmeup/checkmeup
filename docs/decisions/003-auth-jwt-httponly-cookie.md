# ADR-003: JWT in httpOnly cookie with DB-backed refresh tokens

**Status:** Accepted — scope narrowed to browser session auth by [ADR-028](028-api-key-auth-scope.md) (2026-07-03); the "no `Authorization` header" rule below does not apply to the separate `X-API-Key` public-API mechanism
**Date:** 2026-06-13

## Context

Auth options considered:

- Stateless JWT in `Authorization` header (no server-side session state)
- Session ID in httpOnly cookie backed by a sessions table
- Short-lived JWT in httpOnly cookie + long-lived refresh token stored in DB

## Decision

Short-lived JWT in an httpOnly cookie for request auth. Refresh tokens stored in the database, rotated on use. No `Authorization` header — the cookie is sent automatically by the browser.

## Implementation details

**Cookie name:** `access_token` (access JWT), `refresh_token` (refresh token)

**Cookie attributes (both cookies):**
- `HttpOnly` — not readable by JavaScript
- `SameSite=Strict` — provides CSRF protection for same-origin flows; no separate CSRF token needed
- `Secure` — HTTPS only in production; omitted in development
- `Path=/` — sent on all API routes
- `Max-Age` — access cookie: 15 min; refresh cookie: 7 days (matches `JWT_REFRESH_TTL`)

**JWT claims (`access_token`):**
```json
{ "sub": "<user UUID>", "org": "<org UUID>", "exp": <unix timestamp> }
```
`sub` is the standard JWT subject (user ID). `org` is a custom claim for the org ID. Both are UUIDs.

**Signing algorithm:** HS256 (HMAC-SHA256) using `JWT_SECRET` from env.

**401 interceptor:** The frontend `api` client automatically retries any 401 response by calling `POST /auth/refresh`. If the refresh also fails, it clears auth state and redirects to sign-in. **Exception:** `auth.init()` (the initial session check on page load) uses plain `fetch` and bypasses the interceptor — a 401 on `/me` during init simply means "not logged in", not a session error requiring a redirect.

## Consequences

- **XSS protection:** httpOnly cookie means JavaScript cannot read the token
- **CSRF protection:** SameSite=Strict prevents cross-site request forgery without a CSRF token
- **Revocability:** refresh tokens in the DB can be revoked immediately (logout, compromised session); JWTs are still valid until expiry, so keep access TTL short (15 min)
- **Stateful refresh:** requires a DB round-trip on token refresh, unlike fully stateless JWT
- **Authorization header not used:** the cookie is sent automatically; no `Authorization: Bearer …` header is issued or accepted
