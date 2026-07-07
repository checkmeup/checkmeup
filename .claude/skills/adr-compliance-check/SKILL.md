---
name: adr-compliance-check
description: Check the codebase against a handful of concrete, greppable ADR rules from CLAUDE.md's Don't section (no Redis/broker, no ORM, no Stripe/LemonSqueezy, auth.init() must use plain fetch, no Authorization header for browser session auth). Use when asked to "check ADR compliance", "did this reintroduce a banned pattern", or before merging a PR that touches dependencies or auth code.
---

# ADR compliance check

CLAUDE.md's Don't section rejects several architectural patterns
outright — this is a pre-flight for the subset that's reliably
detectable by grepping the right *specific* place, not the whole repo.

## Steps

**1. Run the check.**

```bash
python3 .claude/skills/adr-compliance-check/check.py
```

Exit `0` = all 5 checks pass. Exit `1` = at least one failed, printed
with the offending line(s).

**2. For each failure**, the fix is almost always "revert the change" —
these are things CLAUDE.md explicitly says not to do, not judgment
calls:

| Check | Rule | If it fails |
|---|---|---|
| ADR-001 | No Redis/queue/broker — goroutine workers are intentional | Remove the dependency; solve it with a goroutine worker instead |
| ADR-004 | No ORM — sqlc only | Remove the ORM; write the query by hand in `apps/api/queries/` |
| ADR-026 | Paddle only, no Stripe/LemonSqueezy (ADR-026 supersedes ADR-018) | Remove the dependency; use the existing Paddle integration |
| (Don't list) | `auth.init()` must use plain `fetch`, not `api.*` | A 401 on `/me` during init means "not logged in" — the `api.*` client's interceptor would wrongly treat that as a session error |
| ADR-003 | Browser session auth is the `access_token` httpOnly cookie only, never `Authorization` | Remove the header; rely on `credentials: 'include'`. (The public API's `X-API-Key` header is a separate, intentional mechanism per ADR-028 — this check only scans `apps/web/src`, so it never sees that Go-side code) |

**3. Verify** by re-running the check — should return to all `OK`.

## Why these 5 and not more

Each check here is scoped to the *specific file or dependency manifest*
where a real violation would appear — not a blanket text search. A
blanket search for "Authorization" or "Stripe"/"LemonSqueezy" across the
repo hits constantly on legitimate prose: blog posts narrate the
Stripe-vs-LemonSqueezy-vs-Paddle decision, ADR-history comments mention
LemonSqueezy's old behavior for context, and "Authorization" appears in
explanatory text throughout. Scoping to `go.mod`/`package.json` for
dependency checks, and excluding `apps/web/src/blog/` for the
`Authorization` check, keeps the signal-to-noise ratio high — verified
zero false positives against the current tree when this was authored.

Other Don't-list items (no subdomains for status pages, no public
feature-request board, never commit directly on `main`) were considered
and dropped: they're either not reliably greppable (a subdomain-routing
violation could take unboundedly many shapes) or aren't a code-content
concern at all (git workflow, product-scope decisions) — a flaky or
always-passing check is worse than no check, so they're a manual-review
reminder instead of a false sense of automated coverage.
