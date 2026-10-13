---
name: architecture-guardrails
description: Run audit.py's objective size/complexity/duplication metrics over Go handlers/worker and Vue views/components — Go function complexity, file-size outliers, and near-duplicate Create/Edit view pairs, plus churn hot spots. Use as the input to baloo:architecture's Scope step, or when asked to "find god objects/god components", "find duplicated views", or for a periodic size/complexity health check. The review itself (reading the code, deepening candidates, the report) is baloo:architecture's.
---

# Architecture guardrails

The measurable part of an architecture review, which `baloo:architecture`
doesn't do on its own: it finds candidates by reading the hot spots, while
this script points at three objective proxies for a module doing too much:

1. **Go function cyclomatic complexity** over 15 (lizard's conventional
   `-C 15`) in `apps/api/internal/handler` and `apps/api/internal/worker`,
   the two packages carrying most of the business logic.
2. **File size** per area, at a real gap in this repo's distribution
   (see thresholds below).
3. **Sibling near-duplication**: each `*CreateView.vue` against its
   `*EditView.vue`, on normalized lines — the same screen written twice,
   which size limits can't see since each half stays under them.

A non-failing **churn hot spots** section lists the most-touched source
files, to rank the findings: a finding in a hot spot is where a fix pays.

## Steps

**1. Run the audit.**

```bash
python3 .claude/skills/architecture-guardrails/audit.py
```

Exit `0` means nothing over threshold; exit `1` lists findings,
worst-first per section.

**2. Hand the findings to `baloo:architecture`** as its Scope: read each
file before calling it a finding — a threshold hit nobody has read is not
one. Fixes that already worked here, per category:

- **Go function CCN > 15** → extract the validation/branching into helpers;
  a helper with one caller that removes no branching only moves it.
- **Go file > 700 non-comment lines** → split by job within the same
  package, as `worker.go` was split per monitor type (`worker_cron.go`,
  `worker_uptime.go`, …).
- **Vue view > 600 / component > 250 lines** (`components/ui/` excluded) →
  extract subcomponents, and move logic out of `<script setup>` into a
  composable in `apps/web/src/composables/`.
- **Sibling pair over threshold** → one shared form component both views
  render, with what differs (initial values, submit, title) passed as
  props — not one view with an `isEdit` branch everywhere.

**Not a finding:** the `use*Monitors.ts` composables are shallow on
purpose — they keep TanStack Query keys in one place.

## Thresholds (why these numbers)

Set at measured gaps in the distribution when each was chosen, not round
numbers — re-derive if a threshold starts flagging routine files:

- Go handler/worker files ran 21–562 non-comment lines, then
  `status_public.go` at 824; 700 sat in the gap. Measured as lizard's
  whole-file NLOC (`run_lizard_file_totals` in `audit.py`), since summing
  per-function NLOC misses top-level code.
- Vue views ran 251–494 lines, then 729/746/1158
  (`DocsView`/`DashboardView`/`HomeView`); 600 sat in the gap.
- Vue components (excl. `ui/`) ran 34–170, then 570
  (`NotificationChannelsCard.vue`); 250 sat in the gap.
- Sibling duplication needs **both** >50% shared **and** >=100 shared
  lines: two thin views still score ~65% on boilerplate alone. When the monitor form extraction
  opened, the real pairs shared 117–369 lines, the false ones 74, and
  after the extraction every pair fell to 35–57.

## Known exceptions

None yet. `KNOWN_EXCEPTIONS` in `audit.py` maps `file → rationale`; add
one only after reading the file and confirming it's large because it does
one cohesive job. "It's always been this size" is not a rationale.

## Scope

`apps/api/internal/handler`, `apps/api/internal/worker`,
`apps/web/src/views`, `apps/web/src/components` (excl. `ui/`). Extend
`GO_DIRS` and the Vue directory constants in `audit.py` if other packages
grow enough to need it.

## Known limitation: lizard's JS/TS function length

Lizard can lose brace depth partway through a `.ts`/`.vue` file and stretch
a function's span to the last function before a trailing `return {...}`
(a real ~24-line function reported as 87). Confirm any length finding on
`.ts`/`.vue` by eye; Go CCN numbers haven't shown this. If it's noise,
moving a short function to the end of the file works around it.

## Verifying audit.py itself

Codacy runs Lizard/Prospector/Bandit on it in CI:

```bash
lizard .claude/skills/architecture-guardrails/audit.py
prospector .claude/skills/architecture-guardrails/audit.py
bandit .claude/skills/architecture-guardrails/audit.py
```

Its `subprocess` calls use a fixed argv, never `shell=True`; Bandit's
findings on them carry `# nosec B404`/`# nosec B603`.
