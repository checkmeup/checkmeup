# EP-01: Authentication & account setup

Covers sign-up through to a working session. Prerequisite for every other epic.

---

### US-0101: Sign up

**As a** new user, **I want** to register with email and password **so that** I can start monitoring my services.

**Acceptance criteria:**

- [ ] Email + password form with client-side validation
- [ ] Email must be unique — duplicate returns a clear error
- [ ] Password minimum 8 characters
- [ ] Organization created automatically on sign-up (org name defaults to email prefix)
- [ ] User is signed in immediately after registration
- [ ] JWT access token set in httpOnly cookie; refresh token stored in DB

---

### US-0102: Sign in

**As a** returning user, **I want** to sign in with email and password **so that** I can access my dashboard.

**Acceptance criteria:**

- [ ] Email + password sign-in form
- [ ] Invalid credentials return a generic "incorrect email or password" error (no user enumeration)
- [ ] JWT access token set in httpOnly cookie on success
- [ ] Redirect to dashboard after sign-in

---

### US-0103: Sign out

**As a** signed-in user, **I want** to sign out **so that** my session is terminated on this device.

**Acceptance criteria:**

- [ ] Sign-out clears the auth cookie
- [ ] Refresh token revoked in DB
- [ ] Redirect to sign-in page

---

### US-0104: Stay signed in (silent token refresh)

**As a** signed-in user, **I want** my session to renew automatically **so that** I'm not interrupted while working.

**Acceptance criteria:**

- [ ] Access token silently refreshed using the refresh token before it expires
- [ ] Refresh token rotated on each use
- [ ] Expired or invalid refresh token redirects to sign-in without data loss

---

### US-0105: Reset forgotten password

**As a** user who forgot their password, **I want** to reset it via email **so that** I can regain access.

**Acceptance criteria:**

- [ ] "Forgot password" link on sign-in page
- [ ] Reset email sent with a time-limited link (1 hour TTL)
- [ ] Token is single-use — invalidated immediately after use
- [ ] Success message shown regardless of whether the email exists (no enumeration)
- [ ] After reset, all existing refresh tokens for that user are revoked
