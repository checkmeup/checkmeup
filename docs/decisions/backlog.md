# Decision backlog

Open questions that need an answer before (or early in) the relevant implementation. When decided, create a numbered ADR and remove the entry here.

---

- **Multi-user orgs / invite conflict** ([EP-12](../stories/ep-12-team-management.md)): `users.email` is globally unique and a user belongs to exactly one org via `users.org_id` — no `org_members` table. What happens when someone invites an email that already has an account under a *different* org? Reject the invite outright, or support one user belonging to multiple orgs (bigger schema change: needs an `org_members` join table + role column, away from the current 1:1 `users.org_id`)? Needed before US-1201 can be implemented.

- **WhatsApp provider** ([EP-15](../stories/ep-15-whatsapp-alerts.md)): Meta Cloud API directly, or a BSP (Twilio, MessageBird)? Affects setup complexity, per-message cost, and who handles template approval. Message template content (down alert, recovery alert) also needs drafting and submitting to Meta for approval before US-1502/US-1503 can ship — that approval lead time should be factored into scheduling, not assumed instant.

- **Signal: go/no-go** ([EP-16](../stories/ep-16-signal-alerts.md)): Signal has no official bot/business API — the only path is self-hosting `signal-cli` and registering a phone number as a bot account, which Signal's anti-abuse systems could flag or ban, and which adds an always-on process this repo has otherwise avoided ([ADR-001](001-worker-model.md): no Redis/job queue/external broker). Decide whether this is worth building at all before any EP-16 story starts, ideally backed by actual user demand rather than channel completeness for its own sake.

- **Teams webhook mechanism** ([EP-18](../stories/ep-18-teams-alerts.md)): Microsoft has been retiring the legacy Office 365 Connector "Incoming Webhook" in favor of Power Automate workflow webhooks (Adaptive Card payload). Re-verify the current retirement status and recommended mechanism immediately before implementing US-1801 — this is the kind of platform detail that can shift between when this was written and when the epic is actually picked up.

- **SMS provider + compliance** ([EP-19](../stories/ep-19-sms-alerts.md)): which provider (Twilio, Vonage, AWS SNS) — affects cost and setup. Also confirm the opt-in flow in US-1901 actually satisfies TCPA-style anti-spam regulation in the markets checkmeup's users are in before sending any automated text; this is a legal compliance question, not just a UX one.

- **Manual incident schema** ([EP-24](../stories/ep-24-incident-management.md)): new `status_page_incidents` table decoupled from monitors/monitor-type, or extend the existing `cron_incidents`/`uptime_incidents` tables? The new-table approach looks like the better fit (a manual incident can span multiple monitors of different types, unlike the automatic incidents which are 1:1 with a single monitor's check transitions) but needs a real ADR before US-2401 starts.

- **2FA secret encryption** ([EP-25](../stories/ep-25-two-factor-auth.md)): the TOTP secret needs reversible storage, unlike the one-way `password_hash`/`refresh_tokens.token_hash` pattern already in use. Pick an approach — `pgcrypto`, application-level encryption with a key from env/secrets, or a KMS — before US-2501 starts. First reversible-encryption-at-rest requirement in the codebase, so this sets the pattern for anything similar later.

- **Amend ADR-003 for API keys** ([EP-26](../stories/ep-26-public-api-keys.md)): ADR-003's "no `Authorization` header" rule (also a hard "Don't" in CLAUDE.md) was written for browser session auth; a public API for non-browser clients needs its own mechanism. EP-26 proposes a dedicated `X-API-Key` header, not `Authorization`, to stay clearly outside ADR-003's existing rule rather than quietly contradicting it. Write the actual ADR (or amend ADR-003 directly) confirming this scope before US-2601 starts — and update the CLAUDE.md "Don't" line once decided, since it currently reads as an unconditional ban.
