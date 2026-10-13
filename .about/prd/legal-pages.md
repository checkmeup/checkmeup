# Terms of Service and Privacy Policy

## Problem

The product collects emails and payments and shares data with processors (Paddle, Resend, Telegram, Twilio, Hetzner); users need to know the terms and agree to them, and the business needs a record that they did.

## Users

Every visitor and user.

## Success signal

Every account has a recorded acceptance of the current terms version.

## Scope

- Public `/terms` and `/privacy` pages with a version and effective date, linked from every public page's footer.
- A required acceptance checkbox at sign-up, recorded with the version and time.
- A blocking re-accept screen after sign-in when the terms changed.
- The accepted version and date shown in Settings.

## Non-goals

- A cookie consent manager beyond the analytics banner (ADR frontend).
- Per-country variants of the documents.

## Acceptance criteria

- The Privacy Policy shows its effective date, names the data controller and lists every subprocessor.
  → test: apps/web/src/views/PrivacyView.test.ts "renders the heading and effective date"
  → test: apps/web/src/views/PrivacyView.test.ts "names the data controller"
  → test: apps/web/src/views/PrivacyView.test.ts "lists all subprocessors in the table"
- Sign-up is refused, in the form and by the API, until the terms box is checked.
  → test: apps/web/src/views/SignUpView.test.ts "shows a validation error when terms are not accepted"
  → test: apps/api/internal/handler/auth_test.go "TestSignUp/validation"
- A user whose accepted version is older than the current one is sent to the re-accept screen before any signed-in page.
  → test: apps/web/src/router/index.test.ts "redirects an authenticated user who still needs terms acceptance to accept-terms"
  → test: apps/web/src/views/AcceptTermsView.test.ts "shows the re-accept copy when the user already has a terms version"
- Accepting there records the new version and opens the dashboard; declining offers signing out instead.
  → test: apps/api/internal/handler/auth_test.go "TestAcceptTerms/re-accepting a stale version clears NeedsTermsAcceptance"
  → test: apps/web/src/views/AcceptTermsView.test.ts "accepts terms, sets the user, and navigates to the dashboard on success"
  → test: apps/web/src/views/AcceptTermsView.test.ts "signs out and navigates home when \"Sign out instead\" is clicked"
- Settings shows the accepted version and date, read-only.
  → test: apps/web/src/views/SettingsView.test.ts "shows the formatted terms acceptance date when accepted"
- Terms and Privacy are linked from the footer of the landing pages, the blog and status pages.
