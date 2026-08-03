# Roadmap

**🚀 MVP launched Jun 16 2026 — 9 weeks ahead of the original Aug 14 target.**

Post-MVP work is tracked as **Now / Next / Later** below instead of dated phases — see [ADR-022](decisions/022-post-mvp-docs-organization.md) for why. The dated phase plan used to build the MVP itself is archived in [knowledge/mvp-history.md](knowledge/mvp-history.md).

---

## Now / Next / Later

### Now

(empty)

### Next

- [EP-12](stories/ep-12-team-management.md) — Team management — **6 h**. Invite-conflict handling decided ([ADR-031](decisions/031-team-invite-conflict.md)) — ready to implement.
- [EP-25](stories/ep-25-two-factor-auth.md) — Two-factor authentication — **6 h**. Encryption approach decided ([ADR-030](decisions/030-totp-secret-encryption.md)) — ready to implement.

### Later

Ordered by how easy the blocker is to clear — internal decisions first, external dependencies (providers, approval processes) last.

- [EP-31](stories/ep-31-assertion-checks.md) US-3105 — Multi-step (chained) API checks — **5 h**. US-3101–3104 shipped in v1.9; this is the request-chaining follow-on — story flags it as a good candidate for its own epic during grooming
- [EP-20](stories/ep-20-viber-alerts.md) — Viber alerts — **4 h**.

- [EP-35](stories/ep-35-overlap-detection.md) — Overlap detection — **3.5 h**. Ping-model decision made ([ADR-039](decisions/039-cron-ping-model.md)), and EP-34's start-ping endpoint + `cron_runs` table have shipped — ready to implement.
- [EP-18](stories/ep-18-teams-alerts.md) — Microsoft Teams alerts — **3.5 h**. Postponed: needs a Microsoft 365 work/school (business) account to build and test against, which doesn't exist yet — no cost to the API itself, just a founder-side tenant setup step (e.g. the free Microsoft 365 Developer Program) before work can resume.
- [EP-15](stories/ep-15-whatsapp-alerts.md) — WhatsApp alerts — **4 h**. Blocked on the provider decision in [decision backlog](decisions/backlog.md) — external: provider choice, cost, Meta template approval lead time.
- [EP-16](stories/ep-16-signal-alerts.md) — Signal alerts — **5 h**. Blocked on a go/no-go decision in [decision backlog](decisions/backlog.md) — least certain to proceed at all; no official API, self-hosting `signal-cli` would be a real infra/operational commitment.
- [EP-32](stories/ep-32-multi-region-checking.md) — Multi-region checking — **9 h**. Blocked on the multi-region infra decision in [decision backlog](decisions/backlog.md) — the current single-Hetzner-VPS model has no compute outside one region; highest user-trust value of the `proposals/bucket-list.md` items (kills false positives from checkmeup's own network) but also the most expensive to unblock.
- [EP-39](stories/ep-39-dns-monitoring.md) — DNS record monitoring — **6 h**. No blocker; not yet prioritized against the rest of this list.
