# ADR-022: Post-MVP doc organization — roadmap, backlog, reports

**Status:** Accepted
**Date:** 2026-06-17

## Context

The MVP shipped 2026-06-16 ([roadmap.md](../roadmap.md)). Everything in `docs/` so far was built around a single countdown to launch: `roadmap.md` is a dated, hour-budgeted phase list; `hours.md` logs time per phase; `stories/` epics get built roughly in phase order. That model fit a fixed sprint with a known end. It doesn't fit what comes next — work driven by user feedback, bugs, and incidents, with no fixed end date and no reason to pre-commit to calendar dates months out.

Three things needed a decision:

1. **Roadmap shape** — keep dated phases, or switch to something that doesn't require predicting dates for reactive work.
2. **Reports** — there's no artifact today that answers "what shipped this month" or "what broke and why" without re-reading commit history.
3. **Logs** — `hours.md` is a time log; nothing tracks production incidents, despite the product itself being an uptime/incident tracker.

## Decision

**`docs/stories/`** — unchanged. Epic-per-file + `backlog.md` index continues; EP-09/EP-10 already followed this pattern post-MVP. `stories/backlog.md` is an epic catalog only — no priority ordering (that's `roadmap.md`'s job) and no pricing/plan-limit data (that's [ADR-019](019-plan-limits.md)'s job, to avoid a third copy of the same table). Its index shows an `x/y` stories-done fraction per epic instead of a status label — more information in the same column, and self-explanatory without a separate legend.

**`docs/roadmap.md`** — trimmed to a **Now / Next / Later** list only: three short lists of epic references, no dates, no hour budgets. Move an item left when work starts; drop it once shipped (the report for that month is the permanent record, not the roadmap). Anything genuinely unfinished but not yet prioritized (e.g. the billing-activation follow-ups deferred at MVP launch) lives in **Later**, not buried in an archive, so it doesn't get forgotten.

**`docs/mvp-history.md`** (new) — the full dated phase plan used to build the MVP (Phase 0–6), frozen and append-only, so `roadmap.md` stays short instead of carrying ~250 lines of completed history above the part anyone actually needs to read. Also holds the **work-schedule assumptions** (~28–30 h/week) that the phase hour budgets were derived from — `roadmap.md` has no hour math left to justify keeping that block. Phase 7 (maintenance windows) was the first *post*-MVP feature, so its narrative moved to `reports/2026-06.md` instead of staying in the MVP archive — only its hour numbers stay in `mvp-history.md`'s estimated-vs-logged table, since that table documents the old phase/estimate practice as a whole, not MVP scope specifically.

**`docs/reports/YYYY-MM.md`** (new) — one file per calendar month, written at month-end. Contents: epics/stories shipped that month (links into `stories/`), any ADRs added, production incidents (see below) with root cause and follow-up, and any plan/pricing/metric changes worth a permanent record. This is the answer to "what happened this month" without re-reading `git log`. It does **not** duplicate the changelog — per [ADR-021](021-versioning.md), GitHub Releases remains the single changelog; reports link to the relevant release tags instead of restating commits.

**`docs/incidents/YYYY-MM-DD-slug.md`** (new) — one file per production incident on checkmeup.net's own infrastructure (the irony of an uptime monitor going down undocumented is not worth risking). Minimal format: what broke, detection (did our own monitors catch it?), impact window, root cause, follow-up actions. Linked from the month's report.

**`docs/hours.md`** — keeps the full raw per-day log (nothing removed), but the **Phase** column became **Epic/Story** (`EP-XX` / `US-XXXX`), since "phase" doesn't exist post-MVP. Entries that predate epics or cut across several (repo scaffolding, ADR-writing days, launch-day deploy) keep a short free-text label instead of a forced epic reference. The estimated-vs-logged comparison table moved to `mvp-history.md` instead, since estimates were a phase-based practice that ended with the MVP — post-MVP logging has nothing to compare against.

## Consequences

- `docs/roadmap.md` no longer makes calendar promises for unstarted work — avoids the MVP pattern of estimates being off by 10x+ (see `hours.md`: 259 h estimated vs 29 h logged) for work that's inherently less predictable than a pre-scoped sprint.
- Monthly reports are a discipline, not automation — nothing enforces that one gets written. Acceptable for a solo founder; revisit if a team is added.
- `docs/incidents/` will likely stay empty for a while — that's the goal, not a sign the structure is unused.
- `stories/backlog.md`'s `x/y` fraction needs manual upkeep (no tooling reads or writes it) — same trust model as the rest of `docs/`.
- `mvp-history.md` is write-once: nothing ever edits it again after a phase lands there. If that assumption breaks (e.g. correcting a historical fact), edit it directly — there's no versioning scheme for the archive itself beyond git history.
- `hours.md`'s free-text labels (vs. `EP-XX`/`US-XXXX`) are a judgment call per entry — acceptable for a handful of cross-cutting MVP-era days, but if post-MVP logging starts accumulating non-epic entries regularly, that's a sign work needs an epic/story written for it rather than staying unattributed.
