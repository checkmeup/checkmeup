---
name: rate-limit-audit
description: Audit apps/api/internal/server/server.go for routes with no rate-limit coverage — neither their own httprate wrapper nor an enclosing r.Use(httprate...) on a parent route group. Use when asked to "check rate limiting", "audit rate limits", "did this add a route without rate limiting", or after adding a new endpoint to server.go.
---

# Rate-limit audit

EP-08 established that every route gets rate-limit coverage — either its
own `.With(httprate...)` wrapper, or an enclosing `r.Group`/`r.Route`
block's blanket `r.Use(httprate.Limit(...))` (chi propagates middleware
down through nested route blocks even though `Route()` mounts a
sub-router, so a parent's `r.Use()` still wraps everything nested inside
it). It's easy to add a new endpoint and forget this.

## Steps

**1. Run the audit.**

```bash
python3 .claude/skills/rate-limit-audit/audit.py apps/api/internal/server/server.go
```

Exit code `0` and "All routes covered" means clean. Exit code `1` lists
every route with zero rate-limit coverage, plus a separate "Known
exceptions" section for the two routes that are deliberately unlimited
(see below) — those are reported for visibility, not flagged as failures.

**2. For each real finding**, add rate limiting following the existing
patterns in `server.go`:

- Public, unauthenticated route → `r.With(httprate.LimitByIP(N,
  duration)).Verb(...)`, tuned to the endpoint's abuse potential (compare
  neighboring routes — auth endpoints are tight, e.g. sign-up at 5/hour;
  status-page badges are generous at 300/min since they're embed-friendly)
- Authenticated route inside the `RequireAuth` group → usually needs
  nothing extra, it inherits the blanket 300/min-per-org limit at
  server.go:128; add a tighter `.With(httprate.Limit(N, duration,
  httprate.WithKeyFuncs(authOrgKey)))` only if the action is unusually
  expensive (real third-party API calls, like `/billing/checkout`) or
  spammable (like notification-channel test-sends)
- New top-level route group (like `/public` for the API-key-authed
  routes) → needs its own `r.Use(httprate.Limit(...))`, since it isn't
  nested under the session-auth group's blanket limit

**3. Verify** by re-running the audit script — should return to "All
routes covered."

## Known exceptions (won't be flagged as failures)

- `GET /api/v1/health` — load-balancer/monitoring health check, not
  user-reachable data
- `GET /*` (SPA catch-all, `handleSPA`) — static file serving, no
  per-request cost

If a new genuinely-public, cheap, static-ish route is added and you want
it treated the same way, add it to `KNOWN_EXCEPTIONS` in `audit.py` with
a one-line justification — don't silently ignore a real finding instead.

## Local verification of audit.py itself

If you edit `audit.py`, Codacy runs `Lizard`/`Prospector`/`Bandit` on it
in CI the same as any other file — including this repo's own complexity
limits. `.devcontainer/Dockerfile` installs `lizard`, `prospector`, and
`bandit` via `pipx` (rebuild the devcontainer to pick up a Dockerfile
change), so check complexity locally before pushing instead of guessing:

```bash
lizard .claude/skills/rate-limit-audit/audit.py
prospector .claude/skills/rate-limit-audit/audit.py
```

## Scope

This checks **presence** of rate-limit coverage, not whether the
specific limit value is well-chosen — that's a judgment call the PR
author/reviewer makes by comparing against similar existing routes (see
step 2). It also only covers `server.go`'s route tree; it doesn't check
worker-side loops or other rate-limiting outside the HTTP layer.
