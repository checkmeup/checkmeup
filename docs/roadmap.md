# Roadmap

**🚀 MVP launched Jun 16 2026 — 9 weeks ahead of the original Aug 14 target.**

Post-MVP work is tracked as **Now / Next / Later** below instead of dated phases — see [ADR docs](../.about/adr/docs.md) for why. The dated phase plan used to build the MVP itself is archived in [knowledge/mvp-history.md](knowledge/mvp-history.md).

---

## Now / Next / Later

### Now

(empty)

### Next

- [PRD team-management](../.about/prd/team-management.md) — Team management. Invite-conflict handling decided ([ADR auth](../.about/adr/auth.md)) — ready to implement.
- [PRD two-factor-auth](../.about/prd/two-factor-auth.md) — Two-factor authentication. Encryption approach decided ([ADR auth](../.about/adr/auth.md)) — ready to implement.

### Later

Ordered by how easy the blocker is to clear — internal decisions first, external dependencies (providers, approval processes) last.

- Per-project settings: default notification channels, maintenance windows by project, a status page from a project. A non-goal of [PRD projects](../.about/prd/projects.md), which shipped in v1.52; needs its own PRD first.
- Multi-step (chained) API checks. A non-goal of [PRD assertion-checks](../.about/prd/assertion-checks.md), which shipped in v1.9; needs its own PRD first.
- [PRD overlap-detection](../.about/prd/overlap-detection.md) — Overlap detection. Ping-model decision made ([ADR workers](../.about/adr/workers.md)), and PRD zombie-job-detection's start-ping endpoint + `cron_runs` table have shipped — ready to implement.
- [PRD whatsapp-alerts](../.about/prd/whatsapp-alerts.md) — WhatsApp alerts. Blocked on the provider decision in [the Inbox](../.about/inbox.md) — external: provider choice, cost, Meta template approval lead time.
- [PRD multi-region-checks](../.about/prd/multi-region-checks.md) — Multi-region checking. Blocked on the multi-region infra decision in [the Inbox](../.about/inbox.md) — the current single-Hetzner-VPS model has no compute outside one region; highest user-trust value of the `proposals/bucket-list.md` items (kills false positives from Checkmeup's own network) but also the most expensive to unblock.
