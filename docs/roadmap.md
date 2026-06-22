# Roadmap

**🚀 MVP launched Jun 16 2026 — 9 weeks ahead of the original Aug 14 target.**

Post-MVP work is tracked as **Now / Next / Later** below instead of dated phases — see [ADR-022](decisions/022-post-mvp-docs-organization.md) for why. The dated phase plan used to build the MVP itself is archived in [mvp-history.md](mvp-history.md).

---

## Now / Next / Later

### Now

(empty)

### Next

1. [EP-17](stories/ep-17-slack-alerts.md) — Slack alerts. Built on [EP-14](stories/ep-14-webhook-alerts.md) (shipped v1.7)
2. [EP-18](stories/ep-18-teams-alerts.md) — Microsoft Teams alerts. Also built on EP-14
3. [EP-20](stories/ep-20-viber-alerts.md) — Viber alerts

### Later

Ordered by how easy the blocker is to clear — internal decisions first, external dependencies (providers, approval processes) last.

- Activate billing — all code is done (inline upgrade prompt on 402, Billing page checkout buttons with monthly/annual toggle, explicit checkout redirect URL). What's left is account setup in the LemonSqueezy dashboard, which only the account holder can do — see [`docs/billing-setup.md`](billing-setup.md) for the exact checklist. Trigger: first 402 hit in production, or a user asks about a paid plan.
- [EP-27](stories/ep-27-annual-billing.md) — Annual billing. All 3 stories landed in code; same LemonSqueezy dashboard setup above also activates this.
- [EP-25](stories/ep-25-two-factor-auth.md) — Two-factor authentication. Blocked on the secret-encryption decision in [decision backlog](decisions/backlog.md) — an internal technical choice, quick to resolve.
- [EP-26](stories/ep-26-public-api-keys.md) — Public API and API keys. Blocked on amending ADR-003 in [decision backlog](decisions/backlog.md) — a policy call, not an external dependency.
- [EP-24](stories/ep-24-incident-management.md) — Incident management. Blocked on the manual-incident schema decision in [decision backlog](decisions/backlog.md) — internal schema design.
- [EP-12](stories/ep-12-team-management.md) — Team management. Blocked on resolving the multi-user-org design question in [decision backlog](decisions/backlog.md) — a bigger architectural call than the items above.
- [EP-15](stories/ep-15-whatsapp-alerts.md) — WhatsApp alerts. Blocked on the provider decision in [decision backlog](decisions/backlog.md) — external: provider choice, cost, Meta template approval lead time.
- [EP-19](stories/ep-19-sms-alerts.md) — SMS alerts. Blocked on the provider + compliance decision in [decision backlog](decisions/backlog.md) — external: provider choice plus a legal compliance question.
- [EP-16](stories/ep-16-signal-alerts.md) — Signal alerts. Blocked on a go/no-go decision in [decision backlog](decisions/backlog.md) — least certain to proceed at all; no official API, self-hosting `signal-cli` would be a real infra/operational commitment.
