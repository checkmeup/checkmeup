# EP-23: Suggest a feature

Today, feature requests go through a `mailto:` link and a GitHub Issues link on the Docs page ("There's no support ticket queue — questions, bug reports, and feature requests all reach an engineer directly" — `DocsView.vue`). That's a deliberate choice, not a gap to "fix" with a full public roadmap/voting board (Canny-style) — this epic just makes the existing direct-to-founder flow easier to use from inside the app, without adding a ticketing system, statuses, or public visibility.

---

### US-2301: Submit a feature suggestion from the app

**As a** user, **I want** to suggest a feature without leaving the app **so that** I don't have to switch to email or GitHub.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] "Suggest a feature" entry point inside the app (e.g. Settings)
- [x] Simple form: free-text description, account email pre-filled (read-only)
- [x] Submission stored (`org_id`, `user_id`, text, `created_at`) and emailed via Resend to the founder — no ticket statuses, no public board, no comments/voting
- [x] Confirmation shown after submit, consistent with the existing "reaches an engineer directly" framing

---

### US-2302: Rate-limit suggestions

**As a** platform operator, **I want** suggestion submissions rate-limited **so that** the form can't be used to spam storage or the inbox.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Rate limit per IP and per org, same pattern as existing endpoints ([ADR-013](../decisions/013-rate-limiting.md)) — e.g. 5 per hour
- [ ] Limit-exceeded response shows a clear message, not a generic error

---

### US-2303: Surface the suggestion entry point where users already look for help

**As a** user, **I want** to find the suggestion form where I'd already look for help **so that** I don't have to know it exists in advance.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Docs page "Need help?" section ([DocsView.vue](../../apps/web/src/views/DocsView.vue)) gains the in-app suggestion link alongside the existing email and GitHub links — those stay, this is additive
- [ ] Settings nav includes the same entry point for signed-in users
