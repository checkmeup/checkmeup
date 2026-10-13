# Roadmap

**🚀 MVP launched Jun 16 2026 — 9 weeks ahead of the original Aug 14 target.**

Post-MVP work is tracked as **Now / Next / Later** below instead of dated phases — see [ADR docs](../.about/adr/docs.md) for why. The dated phase plan used to build the MVP itself is archived in [knowledge/mvp-history.md](knowledge/mvp-history.md).

---

## Now / Next / Later

### Now

(empty)

### Next

- [PRD team-management](../.about/prd/team-management.md) — Team management — **6 h**. Invite-conflict handling decided ([ADR auth](../.about/adr/auth.md)) — ready to implement.
- [PRD two-factor-auth](../.about/prd/two-factor-auth.md) — Two-factor authentication — **6 h**. Encryption approach decided ([ADR auth](../.about/adr/auth.md)) — ready to implement.

### Later

Ordered by how easy the blocker is to clear — internal decisions first, external dependencies (providers, approval processes) last.

- Multi-step (chained) API checks — **5 h**. A non-goal of [PRD assertion-checks](../.about/prd/assertion-checks.md), which shipped in v1.9; needs its own PRD first.
- [PRD overlap-detection](../.about/prd/overlap-detection.md) — Overlap detection — **3.5 h**. Ping-model decision made ([ADR workers](../.about/adr/workers.md)), and PRD zombie-job-detection's start-ping endpoint + `cron_runs` table have shipped — ready to implement.
- [PRD ms-teams-alerts](../.about/prd/ms-teams-alerts.md) — Microsoft Teams alerts — **3.5 h**. Postponed: needs a Microsoft 365 work/school (business) account to build and test against, which doesn't exist yet — no cost to the API itself, just a founder-side tenant setup step (e.g. the free Microsoft 365 Developer Program) before work can resume.
- [PRD whatsapp-alerts](../.about/prd/whatsapp-alerts.md) — WhatsApp alerts — **4 h**. Blocked on the provider decision in [the Inbox](../.about/inbox.md) — external: provider choice, cost, Meta template approval lead time.
- [PRD signal-alerts](../.about/prd/signal-alerts.md) — Signal alerts — **5 h**. Blocked on a go/no-go decision in [the Inbox](../.about/inbox.md) — least certain to proceed at all; no official API, self-hosting `signal-cli` would be a real infra/operational commitment.
- [PRD multi-region-checks](../.about/prd/multi-region-checks.md) — Multi-region checking — **9 h**. Blocked on the multi-region infra decision in [the Inbox](../.about/inbox.md) — the current single-Hetzner-VPS model has no compute outside one region; highest user-trust value of the `proposals/bucket-list.md` items (kills false positives from checkmeup's own network) but also the most expensive to unblock.
