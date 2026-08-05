---
name: mutation-testing
description: Run Stryker against apps/web's pure logic modules to find tests that execute code without actually checking it — line coverage says a branch ran, a surviving mutant says nothing would have noticed if it were wrong. Use when asked to "check test quality", "run mutation testing", "find weak/bad tests", "is this actually tested", or after adding a module whose only coverage is indirect.
---

# Mutation testing

Coverage answers "did this line run". Mutation testing answers "would anything
have failed if this line were wrong". Stryker rewrites each expression —
flipping `>` to `>=`, replacing a condition with `false`, emptying an array
literal — reruns the tests, and reports every mutant that **survived**: a
change to production code that no test objected to.

A survivor is not automatically a bug. It marks code whose behavior nothing
asserts, which is where bugs live undisturbed. The value is in reading the
survivors, not in the score.

## Scope

Mutates `apps/web/src/lib/` only — pure, framework-free logic where a
surviving mutant is unambiguous.

**Not `.vue` files.** Stryker cannot mutate inside `<template>`, so an SFC's
score reflects only its `<script>` block and reads misleadingly low. Test
quality for components is better judged by whether their logic lives in
`lib/` at all.

**Not the Go backend.** Stryker is JS/TS only. `apps/api` would need a
separate tool (gremlins, go-mutesting); nothing here covers it.

## Steps

**1. Run it.**

```bash
.claude/skills/mutation-testing/run.sh
```

Takes ~3 minutes. `run.sh` handles the environment problems documented below;
don't invoke `stryker` directly and don't use `npx stryker`.

Narrow it while iterating:

```bash
.claude/skills/mutation-testing/run.sh --mutate src/lib/uptimeMonitorForm.ts
```

**2. Separate behavioral survivors from cosmetic ones.** Most survivors in
this repo are `StringLiteral`/`ObjectLiteral` mutants inside the exported
option tables (`{ label: '10 minutes', value: 10 }` → `{ label: '', ... }`).
Nothing asserts those labels and nothing should — they are data. Ignore them.

The mutator names worth reading are `ConditionalExpression`,
`EqualityOperator`, `LogicalOperator`, `BooleanLiteral`, `ArrayDeclaration`,
`ArithmeticOperator`, and `BlockStatement`:

```bash
python3 - <<'PY'
import json, pathlib
d = json.loads(pathlib.Path("apps/web/reports/mutation/mutation.json").read_text())
KEEP = {"ConditionalExpression","EqualityOperator","LogicalOperator","BooleanLiteral",
        "ArithmeticOperator","ArrayDeclaration","BlockStatement","UpdateOperator"}
for name, f in d["files"].items():
    src = f["source"].splitlines()
    for m in f["mutants"]:
        if m["status"] == "Survived" and m["mutatorName"] in KEEP:
            ln = m["location"]["start"]["line"]
            print(f'{pathlib.Path(name).name}:{ln} [{m["mutatorName"]}]')
            print(f'   {src[ln-1].strip()[:90]}')
            print(f'-> {str(m.get("replacement"))[:90]}')
PY
```

**3. Read each behavioral survivor and classify it.** Three outcomes, and the
distinction matters more than the count:

- **Under-tested but correct.** The code is right; no test pins it. Add the
  test. Boundary survivors (`> 500` → `>= 500`) almost always mean the test
  used a value far from the boundary — 501 where 500 was the interesting case.
- **A real defect.** A survivor in code nothing verifies is where to look
  hardest. If a whole assignment can be replaced (`[...xs]` → `[]`) and
  nothing fails, *nothing checks that field is populated at all* — read that
  code properly before adding a test to it.
- **Not worth testing.** Data tables, log strings, defensive branches that
  cannot be reached. Leave them; don't contort a test to kill a mutant.

Never add a test whose only purpose is to raise the score. A test that pins
behavior nobody depends on is worse than the survivor it killed.

**4. Re-run after fixing** — including against tests you just wrote. This
catches assertions that are weaker than they look: a test can prove two
objects are *distinct* while never checking the copy has the right *contents*,
which shows up as an `ObjectLiteral` survivor on the copy expression.

## What this found the first time it ran

Recorded because it calibrates what to expect, not as a score to beat.

The five `lib/*Form.ts` modules had no direct tests — only indirect coverage
through view tests. Compared against `notificationChannelTypes.ts`, which has
its own test file:

| File | Score | Tests |
| --- | --- | --- |
| `notificationChannelTypes.ts` | 92.19% | direct |
| `maintenanceWindowForm.ts` | 87.50% | indirect |
| `cronMonitorForm.ts` | 76.92% | indirect |
| `portMonitorForm.ts` | 65.45% | indirect |
| `dnsMonitorForm.ts` | 64.44% | indirect |
| `uptimeMonitorForm.ts` | 60.00% | indirect |

**A real bug, found via two survivors.** `jsonAssertions: [...(m.jsonAssertions ?? [])]`
→ `[]` and `channelIds: m.channelIds ?? []` → `m.channelIds && []` both
survived, because nothing verified an existing monitor's channels or
assertions load into the edit form. Reading that gap showed `channelIds` was
not copied at all and `jsonAssertions` was copied only one level deep — and
the form binds `v-model` straight to those objects, so **editing a JSON
assertion mutated the monitor object held in TanStack Query's cache.** Fixed
by cloning per element.

**A test that looked thorough and was not.** `PortMonitorCreateView.test.ts`'s
"port is out of range" case sets port to `70000` and asserts the error
message. Every boundary mutant survived: `< 1` → `<= 1`, `> 65535` → `>= 65535`,
and the empty-port clause. One far-from-boundary value satisfied the assertion
while leaving the actual edges untested. The validation was correct — a bad
test does not imply a bug.

## Environment problems this works around

All four cost real time; `run.sh` handles them, but they resurface if anyone
runs Stryker by hand.

- **`npx stryker` installs the wrong package.** An abandoned 2019 `stryker`
  package still occupies that name on npm and npx fetches it, failing with
  `Cannot find module 'rx'`. Always `./node_modules/.bin/stryker`.
- **Plugins do not auto-discover under bun.** bun installs into a hoisted
  `node_modules/.bun/` layout that defeats Stryker's `@stryker-mutator/*`
  glob, giving "no TestRunner plugins were loaded". `stryker.config.json`
  lists `plugins` explicitly.
- **No `ps` in the devcontainer, and no sudo to install procps.** `run.sh`
  puts `ps_shim.py` on PATH as `ps`. The shim **must exit non-zero when
  nothing matches** — that is real `ps` behavior and `tree-kill` depends on
  it (`code != 0` is its "no more children" branch). Exiting 0 with empty
  output sends it into `"".match(/\d+/g).forEach` and kills the run.
- **Leftover sandboxes corrupt the next run.** Stryker copies the whole app
  into `.stryker-tmp/sandbox-*/` per worker. A sandbox left behind by an
  interrupted run gets collected as real tests (683 became 3,422 once) and
  breaks the following dry run. `run.sh` removes it before and after;
  `vitest.config.ts` excludes it; `.gitignore` covers it.

## Why `ignoreStatic` is on

41% of mutants here are static — module-level `const` option tables, whose
mutants force a full suite rerun each. A run without `ignoreStatic` did not
finish in 20 minutes; with it, 2m46s. They also measure the least: a surviving
mutant in a label array says nothing about test quality. If you ever need
them, expect a very long run and budget accordingly.

## Interpreting the score

Don't treat it as a target, and don't put it in a report as a quality grade —
same rule as `architecture-guardrails`. `thresholds.break` is deliberately
`null` so this never fails CI: it is a tool for finding weak tests when you go
looking, not a gate. The number moves for uninteresting reasons — adding a
correct-but-untested line lowers it, deleting a data table raises it.
