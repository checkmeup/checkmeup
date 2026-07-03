# Roadmap

**🚀 MVP launched Jun 16 2026 — 9 weeks ahead of the original Aug 14 target.**

Post-MVP work is tracked as **Now / Next / Later** below instead of dated phases — see [ADR-022](decisions/022-post-mvp-docs-organization.md) for why. The dated phase plan used to build the MVP itself is archived in [mvp-history.md](mvp-history.md).

---

## Now / Next / Later

### Now

- [EP-18](stories/ep-18-teams-alerts.md) — Microsoft Teams alerts — **3.5 h**. Built on [EP-14](stories/ep-14-webhook-alerts.md) (shipped v1.7)

### Next

(empty)

### Later

Ordered by how easy the blocker is to clear — internal decisions first, external dependencies (providers, approval processes) last.

- [EP-31](stories/ep-31-assertion-checks.md) US-3105 — Multi-step (chained) API checks — **5 h**. US-3101–3104 shipped in v1.9; this is the request-chaining follow-on — story flags it as a good candidate for its own epic during grooming
- [EP-20](stories/ep-20-viber-alerts.md) — Viber alerts — **4 h**.

- Activate billing — all code is done (inline upgrade prompt on 402, Billing page checkout buttons with monthly/annual toggle, Paddle.js checkout overlay). What's left is account setup in the Paddle dashboard, which only the account holder can do — see [`docs/billing-setup.md`](billing-setup.md) for the exact checklist. Trigger: first 402 hit in production, or a user asks about a paid plan.
- [EP-25](stories/ep-25-two-factor-auth.md) — Two-factor authentication — **6 h**. Blocked on the secret-encryption decision in [decision backlog](decisions/backlog.md) — an internal technical choice, quick to resolve.
- [EP-24](stories/ep-24-incident-management.md) — Incident management — **5.5 h**. Blocked on the manual-incident schema decision in [decision backlog](decisions/backlog.md) — internal schema design.
- [EP-12](stories/ep-12-team-management.md) — Team management — **6 h**. Blocked on resolving the multi-user-org design question in [decision backlog](decisions/backlog.md) — a bigger architectural call than the items above.
- [EP-15](stories/ep-15-whatsapp-alerts.md) — WhatsApp alerts — **4 h**. Blocked on the provider decision in [decision backlog](decisions/backlog.md) — external: provider choice, cost, Meta template approval lead time.
- [EP-19](stories/ep-19-sms-alerts.md) — SMS alerts — **4.5 h**. Blocked on the provider + compliance decision in [decision backlog](decisions/backlog.md) — external: provider choice plus a legal compliance question.
- [EP-16](stories/ep-16-signal-alerts.md) — Signal alerts — **5 h**. Blocked on a go/no-go decision in [decision backlog](decisions/backlog.md) — least certain to proceed at all; no official API, self-hosting `signal-cli` would be a real infra/operational commitment.
- [EP-32](stories/ep-32-multi-region-checking.md) — Multi-region checking — **9 h**. Blocked on the multi-region infra decision in [decision backlog](decisions/backlog.md) — the current single-Hetzner-VPS model has no compute outside one region; highest user-trust value of the `bucket-list.md` items (kills false positives from checkmeup's own network) but also the most expensive to unblock.
