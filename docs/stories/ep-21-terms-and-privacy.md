# EP-21: Terms and Privacy

> LemonSqueezy was replaced by Paddle as the payment provider/MoR on 2026-07-03 — see [ADR-026](../decisions/026-billing-paddle-mor.md). The subprocessor list below reflects this epic as originally shipped; the live Privacy Policy lists Paddle today.

The product is live, collecting emails and payment data (via LemonSqueezy) and sharing data with third-party processors (Resend, Telegram, LemonSqueezy, Hetzner), with no published Terms of Service or Privacy Policy and no sign-up acceptance step. This is a real gap today, not a future nice-to-have — worth prioritizing ahead of most other post-MVP work.

---

### US-2101: Publish Terms of Service and Privacy Policy pages

**As a** visitor, **I want** to read the Terms of Service and Privacy Policy before signing up **so that** I understand what I'm agreeing to.

**Estimate:** 2 h

**Acceptance criteria:**

- [x] Two public, unauthenticated pages: `/terms` and `/privacy`
- [x] Privacy Policy lists every third-party processor handling user/visitor data (Resend, Telegram, LemonSqueezy, Hetzner) and what each is used for
- [x] Both pages linked from the site footer on every public page (landing, blog, status pages)
- [x] Static content for MVP, edited directly in the repo — same pattern already used for blog posts (no CMS)

---

### US-2102: Require acceptance at sign-up

**As a** user, **I want** to explicitly accept the Terms and Privacy Policy when I sign up **so that** my consent is on record.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Sign-up form has a required checkbox — "I agree to the Terms of Service and Privacy Policy" with inline links to both — unchecked by default
- [x] Sign-up is blocked until the checkbox is checked, enforced client- and server-side
- [x] Acceptance timestamp and the accepted document version/date stored on the user record

---

### US-2103: Re-prompt on material changes

**As a** user, **I want** to be asked to re-accept if the Terms or Privacy Policy materially change **so that** I'm not silently bound to new terms.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Each document carries a version/effective date in its content
- [x] On sign-in, if the user's stored accepted version is older than the current one, show a blocking re-accept screen before the dashboard loads
- [x] Re-acceptance updates the stored version/timestamp the same way as initial sign-up acceptance (US-2102)

---

### US-2104: Surface the accepted terms in account settings

**As a** user, **I want** to see what I agreed to and when **so that** I can find it later without searching my email.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] Settings shows: "You accepted the Terms of Service and Privacy Policy (version X) on `<date>`" with links to both documents
- [x] Read-only — no re-acceptance action here (that's US-2103's job, triggered automatically on sign-in)
