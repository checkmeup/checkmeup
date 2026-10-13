# Accounts: sign-up, sign-in, sessions and password reset

## Problem

Nothing else in Checkmeup works without an account and a session that lasts while someone works.

## Users

Anyone who signs up; each new account gets its own org.

## Success signal

A new user goes from the sign-up form to the dashboard in one step and stays signed in across a working day.

## Scope

- Sign-up with email, password and accepted terms, creating the user's org.
- Sign-in, sign-out and silent session renewal.
- Password reset by email.
- Rate limits on the sign-up, sign-in and reset endpoints (ADR limits).

## Non-goals

- Signing in with Google, GitHub or SSO.
- One user in several orgs (ADR auth).

## Acceptance criteria

- A visitor signs up with a new email and a password of 8+ characters: an org is created and they land on the dashboard, signed in.
  → test: apps/api/internal/handler/auth_test.go "TestSignUp/success sets cookies and persists user"
  → test: apps/web/src/views/SignUpView.test.ts "submits the form, sets the user, and navigates to the dashboard on success"
- Signing up with an email that already has an account fails with an "account exists" message and creates no org.
  → test: apps/api/internal/handler/auth_test.go "TestSignUp/duplicate email returns 409 and leaves no orphan org"
  → test: apps/web/src/views/SignUpView.test.ts "shows an account-exists message on a 409 error"
- A password under 8 characters is refused on sign-up and on reset.
  → test: apps/web/src/views/SignUpView.test.ts "shows a validation error when the password is shorter than 8 characters"
  → test: apps/api/internal/handler/auth_test.go "TestResetPassword/short password rejected"
- Signing up is impossible until the terms are accepted.
  → test: apps/web/src/views/SignUpView.test.ts "disables the submit button until terms are accepted"
- A wrong password and an unknown email get the same "incorrect email or password" answer.
  → test: apps/api/internal/handler/auth_test.go "TestSignIn/wrong password"
  → test: apps/api/internal/handler/auth_test.go "TestSignIn/unknown email"
  → test: apps/web/src/views/SignInView.test.ts "shows an incorrect credentials message on a 401 error"
- A correct sign-in sets the session cookies and opens the dashboard.
  → test: apps/api/internal/handler/auth_test.go "TestSignIn/success"
  → test: apps/web/src/views/SignInView.test.ts "submits credentials, sets the user, and navigates to the dashboard on success"
- Signing out clears the cookies and revokes the refresh token.
  → test: apps/api/internal/handler/auth_test.go "TestSignOut"
- When the access token expires mid-session, the next request renews the session once and succeeds; a used refresh token is refused.
  → test: apps/web/src/api/client.test.ts "on a 401, refreshes the session and retries the original request once"
  → test: apps/api/internal/handler/auth_test.go "TestRefresh/rotates the refresh token and rejects reuse"
- When the session can't be renewed, the user is sent to sign-in.
  → test: apps/web/src/api/client.test.ts "clears the auth store and redirects to sign-in when the refresh itself fails"
- Requesting a reset for any email shows the same confirmation, whether or not the account exists.
  → test: apps/api/internal/handler/auth_test.go "TestForgotPassword/unknown email still returns 204 (no enumeration)"
  → test: apps/web/src/views/ForgotPasswordView.test.ts "shows the same confirmation message even when the request fails"
- A reset link works once, within an hour; using it sets the new password and signs the user out everywhere.
  → test: apps/api/internal/handler/auth_test.go "TestResetPassword/rotates password and revokes existing sessions"
  → test: apps/api/internal/handler/auth_test.go "TestResetPassword/invalid token"
- An expired or used reset link shows an "expired link" message.
  → test: apps/web/src/views/ResetPasswordView.test.ts "shows an expired-link message on a 400 error"
- A sixth sign-up from one IP within an hour, an eleventh sign-in within 10 minutes, or a fourth reset request within 10 minutes gets `429` with `Retry-After`.
