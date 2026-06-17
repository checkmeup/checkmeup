# EP-25: Two-factor authentication

TOTP-based 2FA (Google Authenticator-style) for sign-in ([EP-01](ep-01-auth.md)), independent of password reset — resetting a forgotten password ([US-0105](ep-01-auth.md)) does not bypass 2FA; the second factor is still required afterward.

**Needs an encryption decision before implementation** (add to [decision backlog](../decisions/backlog.md)): the TOTP secret must be stored reversibly (the server needs the original value to compute the current valid code), unlike everything else in this schema today — `password_hash` and `refresh_tokens.token_hash` are one-way hashes (`apps/api/migrations/001_initial.sql`). This is the first reversible-encryption-at-rest requirement in the codebase and needs a real choice (e.g. `pgcrypto` vs. application-level encryption with a key from env/secrets), not an assumption.

---

### US-2501: Enable two-factor authentication

**As a** user, **I want** to enable 2FA using an authenticator app **so that** my account is protected even if my password leaks.

**Estimate:** 2 h

**Acceptance criteria:**

- [ ] Settings flow shows a QR code and manual-entry secret; user confirms by entering a generated code before it activates
- [ ] TOTP secret stored encrypted at rest (see decision above), never returned in plaintext after initial setup
- [ ] Settings shows 2FA as enabled/disabled

---

### US-2502: Require a TOTP code at sign-in

**As a** user with 2FA enabled, **I want** to enter a code from my authenticator app after my password **so that** a leaked password alone can't access my account.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] After correct email+password, if 2FA is enabled, a second step prompts for the 6-digit code before the session cookie is issued ([ADR-003](../decisions/003-auth-jwt-httponly-cookie.md))
- [ ] Incorrect code rejected with a generic error, no enumeration, consistent with sign-in ([US-0102](ep-01-auth.md))
- [ ] Code accepted within the standard TOTP time-step window (±1 step) to tolerate clock drift
- [ ] Still required after a password reset ([US-0105](ep-01-auth.md)) — resetting the password never bypasses 2FA

---

### US-2503: Generate and use backup codes

**As a** user, **I want** one-time backup codes **so that** I can still sign in if I lose access to my authenticator app.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] 10 single-use backup codes generated and shown once when 2FA is enabled — not retrievable again, only regenerable
- [ ] A backup code can substitute for the TOTP code at sign-in and is consumed (invalidated) on use
- [ ] Regenerating backup codes invalidates all previous ones

---

### US-2504: Disable two-factor authentication

**As a** user, **I want** to disable 2FA **so that** I can stop using it if I no longer want the extra step.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Disabling requires re-entering the current password and a valid TOTP/backup code — not a single click, since this lowers account security
- [ ] Disabling removes the stored TOTP secret and all backup codes
- [ ] All existing refresh tokens revoked on disable, consistent with other security-sensitive changes ([US-0105](ep-01-auth.md))

---

### US-2505: Rate-limit 2FA verification

**As a** platform operator, **I want** TOTP/backup code verification rate-limited **so that** an attacker can't brute-force a 6-digit code or guess backup codes.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Rate limit per account and per IP on the 2FA verification step, same pattern as sign-in ([ADR-013](../decisions/013-rate-limiting.md))
- [ ] Limit-exceeded response returns `429` with `Retry-After`, consistent with other rate-limited endpoints
