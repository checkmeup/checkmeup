---
name: codacy-triage
description: Fetch current Codacy issues for this repo via the Codacy API and triage them using this project's known-noise rules (ignore TSQLLint, ignore cookie findings in *_test.go, treat Trivy/go.mod hits as real, investigate ESLint/Opengrep in production code). Use when asked to "check Codacy", "triage Codacy issues", "fix Codacy findings", or before starting a code-quality fix session.
---

# Codacy triage

Fetches the account-level Codacy issue list and separates real problems from
known noise. CLAUDE.md's "Code quality (Codacy)" section just points here.

## Steps

**1. Fetch current issues.**

```bash
source apps/api/.env
curl -s -X POST \
  -H "api-token: $CODACY_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"filters":{"categories":[],"levels":[],"languages":[]}}' \
  "https://app.codacy.com/api/v3/analysis/organizations/gh/checkmeup/repositories/checkmeup/issues/search?limit=100" \
  | python3 -c "
import json, sys
issues = json.load(sys.stdin)['data']
priority = [i for i in issues if i['patternInfo']['level'] in ('Error','High','Warning')]
for i in sorted(priority, key=lambda x: x['patternInfo']['level']):
    print(f\"[{i['toolInfo']['name']}][{i['patternInfo']['level']}] {i['filePath']}:{i['lineNumber']}: {i['message'][:100]}\")
"
```

If `CODACY_API_TOKEN` is unset, `apps/api/.env` is missing it — tell the
user rather than guessing a value.

**2. Triage each hit against the known rules:**

| Tool / pattern | Verdict | Why |
|---|---|---|
| **TSQLLint** (any finding) | Always ignore | SQL Server rules applied to PostgreSQL migrations — not applicable |
| **Opengrep**: cookies in `*_test.go` | Ignore | Synthetic request cookies in tests intentionally lack `HttpOnly`/`Secure` |
| **Trivy** on `go.mod` | Real — act on it | Upgrade the flagged dependency, or pin a patched version if no upgrade path exists yet |
| **Opengrep / ESLint** in production code (not `*_test.go`) | Investigate before dismissing | Not pre-classified as noise — read the finding on its merits (see ESLint note below) |
| **CodeQL / any line-number finding** | Diagnose against the exact commit SHA it was analyzed at | `git show <sha>:<path> \| sed -n '<n>p'`, or the raw GitHub blob URL — never diagnose from memory of "what's around that line," line numbers shift and guessing wrong means fixing the wrong code and re-triggering CI for nothing |
| **Bandit B603/B404** on `.claude/skills/**/*.py` | Usually fine — suppress | If the script only ever invokes a literal argv list (never `shell=True`, no externally-supplied input), suppress with `# nosec <code>` on the exact flagged line + a one-line rationale |
| **Opengrep `open-redirect`** on Go code | Often a false positive — suppress | If the `http.Redirect` target is provably confined to a fixed, config-derived host with only path/query copied from the request (never scheme/host), suppress with `// nosemgrep: <rule-id>` trailing on the exact flagged line (same-line only — a comment above is silently ignored, unlike Bandit's `# nosec`) + rationale |
| **Prospector docstring D212/D213** on `.py` files | Don't chase — use a single-line module docstring | The two rules are mutually exclusive (one wants the summary on line 1, the other on line 2) |
| **Lizard/Prospector complexity** on `.py` files | Real — split, don't suppress | Same threshold philosophy as Go handlers: small single-purpose functions |

**ESLint note:** Codacy's ESLint (`@typescript-eslint`, type-aware rules) is
**stricter than and different from** this repo's local `bun run lint`
(oxlint) — oxlint passing locally does not guarantee Codacy's ESLint pass.
Rules seen catching things oxlint doesn't, or getting them wrong:
`security/detect-object-injection` (fires on any `obj[variable]` bracket
access, even when the variable is statically known-safe — a closed union's
own keys, never external input); `@typescript-eslint/no-redundant-type-constituents`
(fired on a plain `SomeInterface | undefined` param type, likely because a
cross-module `@/` path-aliased type import doesn't resolve in Codacy's
isolated lint environment and silently degrades to `any`; fix by narrowing
the param to a minimal inline structural type instead of importing the
interface, if only 1-2 fields are actually used); `@typescript-eslint/no-unnecessary-condition`
(fired on a `Partial<Record<K, V>>`-typed dictionary lookup's `if (!value)`
guard, claiming the value is "always falsy" — Codacy's isolated environment
isn't resolving the `Partial<...>` cast the same way this project's own
`tsc` does). **When a type-aware rule's verdict is disputed, verify by
running `tsc --strict` directly on a minimal repro of the exact
expression** — if `tsc` genuinely errors without the guard, that's a
confirmed Codacy-environment false positive; suppress with
`// eslint-disable-next-line <rule-id>` plus the verification rationale
rather than contorting the code further to satisfy a tool that's
demonstrably wrong.

**3. Verify a fix locally before pushing**, rather than round-tripping
through CI: `.codacy/cli.sh analyze <path>` runs the same engines (incl.
Opengrep) CI does. It needs a UTF-8 locale forced or Opengrep's config
loader crashes decoding a non-ASCII byte in this environment's default
locale — `LC_ALL=C.UTF-8 LANG=C.UTF-8 .codacy/cli.sh analyze <path>`.

```bash
cd apps/api && golangci-lint run ./...      # Go findings
cd apps/web && bunx oxlint .                # JS/TS findings
```

**4. Group fixes into one branch/PR** rather than one PR per finding, unless
the user asks otherwise — small independent lint fixes are natural to batch.
Use the pr-merge skill once CI is green.

## Notes

- This is an **account-level** token (`CODACY_API_TOKEN`), not project-scoped
  — the same command works from any checkout with the token in `.env`.
- CI already re-runs Codacy analysis on every PR (`Codacy Static Code
  Analysis`, `Codacy Diff Coverage`, `Codacy Coverage Variation` checks) —
  this skill is for proactively triaging the backlog, not something CI needs.
