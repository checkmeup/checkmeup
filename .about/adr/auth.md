# Authentication: browser sessions, API keys, two-factor and org membership

Date: 2026-10-13

A browser session is a short-lived JWT (15 min, HS256, claims `sub` user and `org`) in the `access_token` httpOnly cookie, with a refresh token in the `refresh_token` cookie, stored hashed in the database and rotated on use (7 days). Both cookies are `HttpOnly`, `SameSite=Strict`, `Secure` in production and `Path=/`; `SameSite=Strict` is the CSRF protection, so there is no CSRF token. Browser sessions never use the `Authorization` header. The frontend retries a `401` once through `POST /auth/refresh`; `auth.init()` uses plain `fetch` instead, because a `401` on `/me` at startup just means "not signed in".

The public API is for scripts and third-party servers and authenticates with an `X-API-Key` header, never `Authorization`, so the two mechanisms stay visibly apart in code and logs. A key belongs to one org, is shown once, stored hashed, starts with `cmu_live_`, can be revoked at any time and is rate limited on its own (ADR limits). Read-only scoping is agreed but not built; it must exist before the first write endpoint under `/api/v1/public` (see the Inbox).

TOTP secrets must be recoverable, so they are encrypted in the application with AES-256-GCM under `TOTP_ENCRYPTION_KEY` (32 bytes, base64), the nonce stored with the ciphertext. Backup codes are one-way hashed.

A user belongs to exactly one org. Inviting an email that already has an account, in any org, is refused with "This email is already registered".

## Considered options

- A stateless JWT in the `Authorization` header, or a server-side session table: the cookie JWT gives XSS protection with a revocable refresh token.
- `pgcrypto` for TOTP secrets: the key still comes from the application, and it adds a Postgres extension and puts security logic in SQL.
- A cloud KMS for TOTP secrets: the first cloud dependency, for one column.
- Users in several orgs: needs a membership table, an active org in the session and every org-scoped query rechecked; no demand from small teams yet.

## Consequences

- Access tokens stay valid until they expire even after logout; the 15-minute lifetime bounds that.
- Losing or rotating `TOTP_ENCRYPTION_KEY` breaks every user's 2FA; recovery is disabling 2FA by hand and re-enrolling. Back it up like `JWT_SECRET`.
- Someone invited with an email registered elsewhere needs another email address.
