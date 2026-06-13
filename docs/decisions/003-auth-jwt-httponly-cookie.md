# ADR-003: JWT in httpOnly cookie with DB-backed refresh tokens

**Status:** Accepted  
**Date:** 2026-06-13

## Context

Auth options considered:

- Stateless JWT in `Authorization` header (no server-side session state)
- Session ID in httpOnly cookie backed by a sessions table
- Short-lived JWT in httpOnly cookie + long-lived refresh token stored in DB

## Decision

Short-lived JWT in an httpOnly cookie for request auth. Refresh tokens stored in the database, rotated on use. No `Authorization` header — the cookie is sent automatically by the browser.

## Consequences

- **XSS protection:** httpOnly cookie means JavaScript cannot read the token
- **CSRF exposure:** cookie-based auth requires CSRF protection on state-changing endpoints (SameSite=Strict or a CSRF token)
- **Revocability:** refresh tokens in the DB can be revoked immediately (logout, compromised session); JWTs are still valid until expiry, so keep JWT TTL short (e.g. 15 min)
- **Stateful refresh:** requires a DB round-trip on token refresh, unlike fully stateless JWT
