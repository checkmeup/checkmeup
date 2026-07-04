# ADR-030: TOTP secret encryption — application-level AES-256-GCM

**Date:** 2026-07-04
**Status:** Accepted

---

## Context

[EP-25](../stories/ep-25-two-factor-auth.md) needs the TOTP secret stored reversibly — the server must recover the original secret to compute the current valid code — unlike every other sensitive value in the schema today (`password_hash`, `refresh_tokens.token_hash`), which are one-way hashes ([`001_initial.sql`](../../apps/api/migrations/001_initial.sql)). This is the first reversible-encryption-at-rest requirement in the codebase. Per the [decision backlog](backlog.md), this needed a real choice before US-2501 starts: `pgcrypto`, application-level encryption, or a KMS.

## Decision

**Application-level AES-256-GCM, in Go, keyed by a new env var (`TOTP_ENCRYPTION_KEY`).**

- A KMS (AWS KMS, GCP KMS, etc.) is rejected as disproportionate — it would be the first cloud-provider dependency in a codebase that otherwise runs entirely on one Hetzner VPS plus a handful of API-key-based SaaS integrations ([ADR-001](001-worker-model.md), [ADR-006](006-infrastructure-hetzner-kamal-traefik.md)), solely to encrypt one column.
- `pgcrypto` (`pgp_sym_encrypt`/`pgp_sym_decrypt`) was considered. It avoids writing crypto code in Go, but the encryption key still has to be passed into the SQL call from the application layer at query time — Postgres doesn't manage the key itself — so it doesn't actually reduce key-management burden versus doing it in Go, it just relocates the encrypt/decrypt operation into SQL. It would also require enabling and depending on a Postgres extension (`CREATE EXTENSION pgcrypto`) that nothing else in this schema needs, and it puts security-sensitive logic in migrations/SQL rather than testable Go code — inconsistent with this codebase's existing split of business logic and validation living in Go, with SQL staying declarative ([ADR-004](004-sqlc-over-orm.md)).
- Application-level Go encryption keeps the key in the same trust boundary `JWT_SECRET` already lives in (`apps/api/.env`, read once at startup), is unit-testable like any other Go function, and requires no new infrastructure or Postgres extension.

**Shape:**

- New column `two_factor_secrets.secret_encrypted BYTEA` — a 12-byte nonce prepended to the AES-256-GCM ciphertext in a single column. Decrypted only inside the TOTP verification code path (US-2502); never returned to the client after initial setup, per US-2501's existing acceptance criterion.
- `TOTP_ENCRYPTION_KEY` is a 32-byte, base64-encoded random value, generated once and stored in `apps/api/.env` — same generation and handling pattern as `JWT_SECRET`.
- Backup codes (US-2503) remain one-way hashed (bcrypt or equivalent), same as `password_hash` — they're single-use secrets checked for a match and never displayed again, so they don't need the reversible property TOTP secrets do.

## Consequences

- `TOTP_ENCRYPTION_KEY` becomes a new required-once-2FA-is-used env var — add it to CLAUDE.md's env var table alongside `JWT_SECRET` when EP-25 ships.
- **Key rotation is out of scope for MVP.** If `TOTP_ENCRYPTION_KEY` is ever rotated or lost, every existing user's stored TOTP secret becomes undecryptable, silently breaking their 2FA sign-in. Documented here explicitly as a known operational risk rather than solved now, in keeping with this project's pattern of naming a real gap instead of quietly assuming it away (see [ADR-028](028-api-key-auth-scope.md)'s own "implementation status" section for the same kind of honesty). If it happens: 2FA must be disabled server-side via manual DB/support intervention, and the user re-enrolls.
- `TOTP_ENCRYPTION_KEY` is now a backup-worthy secret, same tier as `JWT_SECRET` and `DATABASE_URL` — should be covered by whatever secret-backup practice already applies to those (defining that practice is outside this ADR's scope).
- Removes the "2FA secret encryption" entry from the [decision backlog](backlog.md). EP-25 is unblocked.
