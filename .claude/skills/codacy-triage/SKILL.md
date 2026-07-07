---
name: codacy-triage
description: Fetch current Codacy issues for this repo via the Codacy API and triage them using this project's known-noise rules (ignore TSQLLint, ignore cookie findings in *_test.go, treat Trivy/go.mod hits as real, investigate ESLint/Opengrep in production code). Use when asked to "check Codacy", "triage Codacy issues", "fix Codacy findings", or before starting a code-quality fix session.
---

# Codacy triage

Fetches the account-level Codacy issue list and separates real problems from
known noise, per CLAUDE.md's "Code quality (Codacy)" section.

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
| **Opengrep / ESLint** in production code (not `*_test.go`) | Investigate before dismissing | Not pre-classified as noise — read the finding on its merits |

**3. For each "real" finding**, read the file at the reported line, fix it,
and re-run the relevant linter locally before committing:

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
