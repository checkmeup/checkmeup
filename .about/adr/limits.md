# Rate limits, safety caps and data retention

Date: 2026-10-13

Rate limits are in-process `go-chi/httprate` middleware, per route, keyed by the narrowest thing that fits: IP for auth, sign-up, webhooks and badges, the ping token for pings, the org (plus IP) for feature suggestions, the API key for the public API. A limited request gets `429` with `Retry-After`; counters reset when the process restarts. The current routes and numbers are in `apps/api/internal/server/server.go` and `docs/reference/limits.md`. Badges are keyed by IP, not slug: a badge URL is public, and keying by slug would let one page's many embedders exhaust everyone's limit.

Resources that grow by human action rather than by plan get a flat cap, the same on every plan, with a `LIMIT 200` on their list query: 100 updates per incident, 100 maintenance windows per org, 100 active API keys per org, 100 active manual incidents per org. Hitting one returns `409` with its own code (`too_many_incident_updates`, `too_many_maintenance_windows`, `too_many_api_keys`, `too_many_active_incidents`), never `plan_limit_reached`, because upgrading doesn't lift it; only resolving, deleting or revoking does.

Check history is pruned daily by the worker: `cron_pings` and `cron_runs` after 30 days, `uptime_checks`, `port_checks` and `dns_checks` after 90, resolved manual incidents 90 days after they resolve. An unresolved incident is never pruned. Automatic incidents live in their own tables, so pruning checks doesn't touch incident history.

## Considered options

- Rate limiting in the reverse proxy: can't key per IP or per token at the route level.
- Redis token buckets: precise and restart-proof, but an external broker (ADR workers).
- Keeping the last N pings per monitor: 16 hours of history for a minute job, years for a daily one.
- A per-plan limit on manual incidents: 90-day retention plus the active cap bounds growth without one.

## Consequences

- Per-IP limits can be dodged from many IPs; they raise the cost of abuse rather than prevent it.
- Maintenance windows have a cap but no expiry: an org at 100 deletes old ones by hand. Revisit if anyone reaches it.
- Caps that are each bounded can still multiply into one unbounded page render; the public status page loads all active incidents' updates in one batched query for that reason.
