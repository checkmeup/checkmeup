---
name: architecture-guardrails
description: Audit Go handlers/worker and Vue views/components for function complexity and file-size outliers — objective proxies for Single Responsibility violations (a function or component quietly doing more than one job). Use when asked to "do an architecture review", "check for SRP violations", "find god objects/god components", or periodically as a health check independent of any single PR.
---

# Architecture guardrails

Unlike `org-id-audit` or `rate-limit-audit`, this doesn't enforce a rule
CLAUDE.md states outright — "keep functions and files focused" isn't
mechanically checkable the way "every tenant-scoped query needs
`org_id`" is. What *is* checkable are two objective proxies for it:

1. **Go function cyclomatic complexity** (lizard, `-C 15`, its own
   conventional default) in `apps/api/internal/handler` and
   `apps/api/internal/worker` — the two packages carrying most of this
   repo's business logic.
2. **File size**, per area, set at a real gap in this repo's current
   distribution rather than a round number — see thresholds below. A
   big file isn't proof of a violation on its own, but every current
   hit was in fact doing several unrelated things when read.

A clean run doesn't mean the architecture is good — it means nothing
crossed an objective size/complexity line. Genuine SRP judgment (is this
abstraction premature, is this the right boundary) still needs a human
or an LLM read of the actual file; this just tells you where to look
first.

## Steps

**1. Run the audit.**

```bash
python3 .claude/skills/architecture-guardrails/audit.py
```

Prints one section per check; exit `0` means nothing over threshold,
exit `1` lists real findings. Findings are ranked worst-first within
each section.

**2. For each finding, read the file before deciding it's real** — size
alone doesn't prove a violation, but in practice every current hit is:

- **Go function CCN > 15** → extract validation/branching into helper
  functions, same philosophy CLAUDE.md's Codacy triage guide already
  states for Lizard findings on `.py` files: "split into small
  single-purpose functions rather than suppressing." Don't extract
  a helper that's only ever called from one place and doesn't reduce
  branching — that's just moving the complexity, not reducing it.
- **Go handler/worker file > 700 logical lines** → currently only
  `worker.go` (~1070 logical lines, all monitor types' check loops in
  one file). Split by monitor type (`worker_cron.go`, `worker_uptime.go`,
  `worker_ssl.go`, …) in the same package — Go doesn't need one file per
  package, and per-type files make each check loop reviewable on its own.
- **Vue view > 600 lines** → extract subcomponents for reusable chunks
  of template, and move non-trivial logic into a composable in
  `apps/web/src/composables/` rather than leaving it inline in `<script
  setup>`. This repo's composables (`useCronMonitors.ts` etc.) are
  currently thin TanStack Query wrappers — that's the target shape, push
  business logic there, not into the view.
- **Vue component > 250 lines** (excl. `components/ui/` primitives,
  which are supposed to be simple) → same as above: split into
  subcomponents or extract a composable.

**3. Fix or file it.** If it's a quick, contained split, do it in the
same PR as whatever work touched the file. If it's a larger untangling
(e.g. `worker.go`), don't do a drive-by refactor — use `new-story` to
scope it as its own piece of work instead of expanding an unrelated PR.

## Threshold rationale (why these numbers)

Picked by finding the actual gap in this repo's size distribution, not
guessed — re-derive if the codebase grows enough that a threshold starts
flagging routine files:

- Go handler files run 95–668 raw lines; `worker.go` alone jumps to
  ~1070 logical lines with no other file within 2x of it. 700 sits in
  that gap.
- Vue views run 251–494 lines, then jump to 729/746/1158 for
  `DocsView`/`DashboardView`/`HomeView`. 600 sits in that gap.
- Vue components (excl. `ui/`) run 34–170 lines, then jump to 570 for
  `NotificationChannelsCard.vue`. 250 sits in that gap.

## Known exceptions

None yet. `KNOWN_EXCEPTIONS` in `audit.py` is a dict of `file →
rationale` — only add an entry after actually reading the file and
confirming it's large because it's one cohesive job (e.g. a big
generated or declarative table), not because no one has split it yet.
"It's always been this size" is not a rationale.

## Local verification of audit.py itself

Same as other skills' scripts — Codacy runs Lizard/Prospector/Bandit on
it in CI:

```bash
lizard .claude/skills/architecture-guardrails/audit.py
prospector .claude/skills/architecture-guardrails/audit.py
bandit .claude/skills/architecture-guardrails/audit.py
```

The `subprocess` calls import lizard as a subprocess with a fixed argv
(never `shell=True`, no externally-supplied input) — flagged Bandit
findings are suppressed with `# nosec B404`/`# nosec B603` per CLAUDE.md's
triage guide, not blanket-ignored.

## Scope

Only covers `apps/api/internal/handler`, `apps/api/internal/worker`,
`apps/web/src/views`, and `apps/web/src/components` (excl. `ui/`).
Doesn't cover other Go packages (`internal/telegram`, `internal/slack`,
`internal/twilio`, `internal/email`, `internal/webhook`, …), Pinia
stores, or composables — extend `GO_DIRS`/the Vue directory constants
in `audit.py` if those grow enough to need it. Doesn't check duplication,
coupling, or naming — only complexity and size.

For daily/periodic cadence rather than ad hoc, run this via the `loop`
skill or the `schedule` skill's cron routines — this skill itself
doesn't schedule anything.

## Known limitation: Lizard's JS/TS function-length metric is unreliable

Confirmed by bisection on a real `.ts` composable file: Lizard's parser
can lose brace-depth tracking partway through a file and misattribute a
function's line span all the way through to whichever function is
defined *last* before a trailing `return {...}` statement — inflating
that function's reported length by 3-4x (one real case: a genuinely
~24-line function reported as 87 "lines of code"). Root cause not fully
pinned down (bisected past template-literal interpolation, typed catch
clauses, and a `{1,14}`-style regex quantifier without finding the exact
trigger) — treat any Lizard `nloc-medium`/length finding on a `.ts`/`.vue`
file as needing visual confirmation of the function's real size before
acting on it, the same way `go_function_complexity`'s CCN numbers don't
need this caveat (Lizard's Go parsing hasn't shown this issue). A
practical workaround if a real finding turns out to be Lizard-measurement
noise: reorder so a short, simple function sits last in the file,
right before `return`.
