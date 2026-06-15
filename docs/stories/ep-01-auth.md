# EP-01: Authentication & account setup

Covers sign-up through to a working session. Prerequisite for every other epic.

---

### US-0101: Sign up

**As a** new user, **I want** to register with email and password **so that** I can start monitoring my services.

**Acceptance criteria:**

- [x] Email + password form with client-side validation
- [x] Email must be unique — duplicate returns a clear error
- [x] Password minimum 8 characters
- [x] Organization created automatically on sign-up (org name defaults to email prefix)
- [x] User is signed in immediately after registration
- [x] JWT access token set in httpOnly cookie; refresh token stored in DB

---

### US-0102: Sign in

**As a** returning user, **I want** to sign in with email and password **so that** I can access my dashboard.

**Acceptance criteria:**

- [x] Email + password sign-in form
- [x] Invalid credentials return a generic "incorrect email or password" error (no user enumeration)
- [x] JWT access token set in httpOnly cookie on success
- [x] Redirect to dashboard after sign-in

---

### US-0103: Sign out

**As a** signed-in user, **I want** to sign out **so that** my session is terminated on this device.

**Acceptance criteria:**

- [x] Sign-out clears the auth cookie
- [x] Refresh token revoked in DB
- [x] Redirect to sign-in page

---

### US-0104: Stay signed in (silent token refresh)

**As a** signed-in user, **I want** my session to renew automatically **so that** I'm not interrupted while working.

**Acceptance criteria:**

- [x] Access token silently refreshed using the refresh token before it expires
- [x] Refresh token rotated on each use
- [x] Expired or invalid refresh token redirects to sign-in without data loss

---

### US-0105: Reset forgotten password

**As a** user who forgot their password, **I want** to reset it via email **so that** I can regain access.

**Acceptance criteria:**

- [x] "Forgot password?" link on sign-in page
- [x] Reset email sent with a time-limited link (1 hour TTL)
- [x] Token is single-use — invalidated immediately after use
- [x] Success message shown regardless of whether the email exists (no enumeration)
- [x] After reset, all existing refresh tokens for that user are revoked

---

### US-0106: Protect auth endpoints from abuse

**As a** platform operator, **I want** auth endpoints to be rate-limited **so that** brute-force attacks cannot compromise accounts and email-bombing cannot exhaust our Resend quota.

**Acceptance criteria:**

- [x] `POST /sign-up`: max 5 requests per IP per hour — prevents account farming and CPU exhaustion via bcrypt
- [x] `POST /sign-in`: max 10 requests per IP per 10 minutes — raises brute-force cost to an impractical level
- [x] `POST /forgot-password`: max 3 requests per IP per 10 minutes — prevents email bombing and Resend quota drain
- [x] Rate-limited responses return HTTP 429 with a `Retry-After` header
- [x] Limits are enforced in-process (no Redis); reset on restart is acceptable for MVP (see ADR-013)
