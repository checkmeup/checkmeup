---
name: architecture-guardrails
description: Audit Go handlers/worker and Vue views/components for function complexity, file-size outliers, and near-duplicate sibling views — objective proxies for Single Responsibility violations (a module quietly doing more than one job, or the same screen written twice), ranked against recent churn. Use when asked to "do an architecture review", "check for SRP violations", "find god objects/god components", "find duplicated views", or periodically as a health check independent of any single PR.
---

# Architecture guardrails

Unlike `org-id-audit` or `rate-limit-audit`, this doesn't enforce a rule
CLAUDE.md states outright — "keep functions and files focused" isn't
mechanically checkable the way "every tenant-scoped query needs
`org_id`" is. What *is* checkable are three objective proxies for it:

1. **Go function cyclomatic complexity** (lizard, `-C 15`, its own
   conventional default) in `apps/api/internal/handler` and
   `apps/api/internal/worker` — the two packages carrying most of this
   repo's business logic.
2. **File size**, per area, set at a real gap in this repo's current
   distribution rather than a round number — see thresholds below. A
   big file isn't proof of a violation on its own, but every current
   hit was in fact doing several unrelated things when read.
3. **Sibling near-duplication** — `*CreateView.vue` against its
   `*EditView.vue`, on normalized lines. This catches the opposite
   failure from checks 1 and 2: not one file doing too much, but the
   same screen written twice. Size thresholds are structurally blind to
   it, since each half can sit comfortably under the limit while the
   pair is 80%+ identical.

Plus one **non-failing** section, `churn hot spots`: the most-touched
source files in recent history. Churn is not a defect and never sets the
exit code — it tells you which of the findings above to fix first.
Deepening a file nobody edits buys nothing; a finding that overlaps a
hot spot is where the leverage is.

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
- **Go handler/worker file > 700 non-comment lines** → currently
  `status_public.go` (824 non-comment lines — HTTP handlers, SVG badge
  rendering, and status/uptime computation all in one file). `worker.go`
  was this file's original example (~1070 logical lines, all monitor
  types' check loops in one place) and has since been split by monitor
  type (`worker_cron.go`, `worker_uptime.go`, `worker_ssl.go`, …) in the
  same package — Go doesn't need one file per package, and per-type files
  make each check loop reviewable on its own. Same pattern applies to
  whatever crosses the threshold next.
- **Vue view > 600 lines** → extract subcomponents for reusable chunks
  of template, and move non-trivial logic into a composable in
  `apps/web/src/composables/` rather than leaving it inline in `<script
  setup>`. This repo's composables (`useCronMonitors.ts` etc.) are
  currently thin TanStack Query wrappers — that's the target shape, push
  business logic there, not into the view.
- **Vue component > 250 lines** (excl. `components/ui/` primitives,
  which are supposed to be simple) → same as above: split into
  subcomponents or extract a composable.
- **Sibling pair > 50% shared** → extract the shared form into one
  component both views render, with mode-specific behaviour (initial
  values, submit handler, page title) passed in as props. Don't
  "fix" it by deleting one view and adding an `isEdit` branch
  everywhere — that trades duplication for a conditional maze. Check
  what actually differs first: the current pairs differ mostly in
  data-loading and submit, not in the form body.

**3. Fix or file it.** If it's a quick, contained split, do it in the
same PR as whatever work touched the file. If it's a larger untangling
(e.g. `worker.go`), don't do a drive-by refactor — use `new-story` to
scope it as its own piece of work instead of expanding an unrelated PR.

**4. Report each confirmed finding in this shape**, so the decision in
step 3 is reviewable rather than a bare threshold number:

```text
Issue:          <what the file/function is doing that's more than one job>
Location:       <path:line>
Impact:         <what this costs — review difficulty, test surface, blast radius>
Recommendation: <the specific split, per step 2's per-category guidance>
Confidence:     strong | worth exploring | speculative
Effort:         low | medium | high
```

`Confidence` is the honest calibration CLAUDE.md's principle 2 already
asks for — `strong` means you read the file and the split is obvious,
`speculative` means the threshold fired but the fix isn't clear yet.
Prefer it over inventing a priority ranking the audit can't support.

Only emit this for findings you confirmed by *reading* the file in step
2 — a threshold hit you haven't read is not a finding yet.

**Never attach a numeric quality score** to a file, a package, or the
codebase (no "structure: 8/10", no "coupling: 6/10"). This skill
measures size, complexity, and duplication; anything past those is not
derivable from what it ran. State what was measured, or say the run was
clean.

**Don't re-litigate an ADR.** Several shapes that look like findings are
settled decisions — goroutine workers rather than a queue
([ADR-001](../../../docs/decisions/001-worker-model.md)), sqlc rather than
an ORM ([ADR-004](../../../docs/decisions/004-sqlc-over-orm.md)), and the
rest of CLAUDE.md's Don't list. Read the relevant ADR before proposing a
restructure in its area. If the friction is genuinely bad enough to
reopen one, say so explicitly and name the ADR — don't quietly propose
something it forbids.

### Naming things in a finding

Use the same word for the same idea every time, so findings stay
comparable across runs:

- **module** — a Go package or a Vue component/composable. Not
  "service", not "unit", not "layer".
- **interface** — what callers must know to use a module. Not "API"
  (this repo's public HTTP API is a different thing entirely, see
  [ADR-028](../../../docs/decisions/028-api-key-auth-scope.md)).
- **shallow / deep** — shallow means the interface costs nearly as much
  to understand as the implementation it hides. Deep is the goal: a
  small interface over substantial implementation. The 9-line
  `use*Monitors.ts` composables are deliberately shallow and fine —
  they exist to keep TanStack query keys in one place, and deleting
  them would scatter those literals across every view. Shallow is only
  a finding when removing the module would *reduce* total complexity.
- Say what improves in those terms — "one interface, six call sites",
  "form logic stops being written twice" — rather than "cleaner" or
  "easier to maintain", which don't say what changed.

## Threshold rationale (why these numbers)

Picked by finding the actual gap in this repo's size distribution, not
guessed — re-derive if the codebase grows enough that a threshold starts
flagging routine files:

- Go handler/worker files run 21–562 non-comment lines; `status_public.go`
  alone jumps to 824, with no other file within 1.5x of it. 700 sits in
  that gap. (Measured via lizard's whole-file NLOC — the per-function
  `--csv` sum audit.py used to use undercounts real file size, since it
  misses top-level code outside any function body; see `run_lizard_file_totals`
  in `audit.py`.)
- Vue views run 251–494 lines, then jump to 729/746/1158 for
  `DocsView`/`DashboardView`/`HomeView`. 600 sits in that gap.
- Vue components (excl. `ui/`) run 34–170 lines, then jump to 570 for
  `NotificationChannelsCard.vue`. 250 sits in that gap.
- Sibling duplication is set at **50%** and is *not* a distribution gap
  — unlike the size checks, this one isn't outlier detection. Past half,
  the pair is more the same screen than two screens, and every change
  has to be made twice regardless of how the rest of the repo looks.
  All seven current pairs measure 54–87%, so the number that matters is
  the ranking, not the cutoff.

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
