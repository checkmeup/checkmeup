# Two-factor authentication

## Problem

A leaked or reused password gives full access to an account that controls clients' monitoring and status pages.

## Users

Users who want a second factor at sign-in; optional for everyone.

## Success signal

An account with 2FA on can't be entered with the password alone, including after a password reset.

## Scope

- Turning on TOTP 2FA from Settings with a QR code, confirmed by a code; the secret encrypted at rest (ADR auth).
- A code step at sign-in for accounts with 2FA, before any session cookie is set.
- Ten single-use backup codes, shown once and regenerable.
- Turning 2FA off with the password and a code.
- Rate limits on the code step.

## Non-goals

- SMS or email codes, WebAuthn and passkeys.
- Requiring 2FA for a whole org.

## Acceptance criteria

- Settings shows a QR code and the secret for manual entry; 2FA turns on only after a valid code is entered, and the secret is never returned again.
- With 2FA on, a correct password leads to a code prompt and no session cookie is set until a valid code is entered.
- A wrong code gets the same generic error as a wrong password.
- A code from the previous or next 30-second step is accepted, to tolerate clock drift.
- After a password reset, sign-in still asks for the code.
- Turning 2FA on shows ten backup codes once; each works once in place of a code; regenerating them invalidates the old ones.
- Turning 2FA off needs the current password and a valid code, deletes the secret and backup codes, and signs the user out everywhere.
- Too many code attempts from one account or IP get `429` with `Retry-After`.
