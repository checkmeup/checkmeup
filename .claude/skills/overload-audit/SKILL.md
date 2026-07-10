---
name: overload-audit
description: Re-verify the concrete, greppable claims in docs/reference/limits.md (bounded check-loop concurrency, incident-list query caps, old-data pruning, rate limits, request body cap, plan-limited monitor creation) so the DoS/overload audit doc can't silently drift stale again. Use when asked to "check for DoS/overload regressions", "audit unbounded operations", "verify limits.md", or periodically as a health check independent of any single PR.
---

# Overload audit

[`docs/reference/limits.md`](../../../docs/reference/limits.md) is a DoS/
overload security audit — a checklist of "things that are fine" (bounded
concurrency, capped queries, rate limits, plan limits), each with a
file/line citation. It already drifted stale once: four findings were
fixed in code across earlier commits without the doc being updated, so it
read as if open vulnerabilities were still open when they weren't. This
skill re-verifies the concrete, mechanically-checkable subset of that
checklist so a *regression* (someone removing the semaphore, the query
cap, or a rate limit while touching nearby code) gets caught the same way
`rate-limit-audit`/`org-id-audit` catch their respective invariants —
this skill exists specifically because those two don't cover this doc's
claims.

## Steps

**1. Run the audit.**

```bash
python3 .claude/skills/overload-audit/audit.py
```

Exit `0` and "All checkable claims... still hold" means clean. Exit `1`
lists every claim that no longer matches the code, each naming the file
and what's missing.

**2. For each finding, decide which side drifted:**

- **Code regressed** (the protection was removed/weakened) → restore it,
  following the pattern the finding names (e.g. re-add the
  `checkConcurrency` semaphore to a check loop, restore the `LIMIT 200`
  on an incident-list query).
- **Doc is stale** (the change was deliberate — e.g. a limit was
  intentionally raised, or a query was intentionally restructured) →
  update `docs/reference/limits.md`'s claim and citation to match, and
  update `audit.py`'s corresponding check if the new shape needs a
  different pattern.

**3. Verify** by re-running the audit — should return to "All checkable
claims... still hold."

## What's checked

- All five check loops (cron/uptime/SSL/domain/port) still bound
  concurrent checks via the `checkConcurrency` semaphore
  (`sem := make(chan struct{}, checkConcurrency)` in each `worker_*.go`)
- `cron_incidents`/`uptime_incidents`/`port_incidents` list queries still
  cap at `LIMIT 200`
- `pruneOldPings` (`worker.go`) still calls all three retention-cleanup
  queries (`DeleteOldCronPings`/`DeleteOldUptimeChecks`/`DeleteOldPortChecks`)
- The public status page + its two badge endpoints are still IP-rate-limited
  at 300/min (`server.go`)
- The `RequireAuth` group still carries its blanket 300/min-per-org
  `httprate.Limit(...)`
- The global 64 KB `http.MaxBytesReader` request-body cap is still wired
- Every plan in `internal/billing/plans.go` still has a finite (non `-1`)
  `MonitorTotal`

## What's deliberately NOT checked

Claims in `limits.md` without an unambiguous, single-pattern mechanical
signal are left to a human/LLM read rather than a fragile regex:

- "Cron pings and uptime check reads are paginated" — the `LIMIT $2
  OFFSET $3` pattern exists but isn't distinctive enough to assert
  absence-means-regression without false positives on unrelated queries.
- "All queries use sqlc parameterized statements (no SQL injection)" —
  a structural guarantee from [ADR-004](../../../docs/decisions/004-sqlc-over-orm.md)
  (sqlc-generated, no ORM/raw string concatenation), not something a
  per-commit grep meaningfully re-verifies.
- Auth/ping endpoint rate limiting generally — already covered by the
  `rate-limit-audit` skill; this skill only checks the two limits
  `limits.md` calls out by exact value (status page, blanket per-org).

If `docs/reference/limits.md` gets a new checkable claim, add a matching
`check_*` function to `audit.py`'s `CHECKS` list rather than letting it
join the "not checked" pile by default.

## Local verification of audit.py itself

Same as other skills' scripts — Codacy runs Lizard/Prospector/Bandit on
it in CI:

```bash
lizard .claude/skills/overload-audit/audit.py
prospector .claude/skills/overload-audit/audit.py
```

## Scope

Only covers the specific files `limits.md` cites
(`internal/worker/worker_*.go`, `queries/{monitors,uptime,port}.sql`,
`internal/server/server.go`, `internal/billing/plans.go`). If a check
loop, query, or limit moves to a new file, update the corresponding path
constant in `audit.py` alongside the move — same discipline
`org-id-audit`'s `KNOWN_EXCEPTIONS` update already requires when code
moves.
