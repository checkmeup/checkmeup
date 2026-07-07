---
name: org-id-audit
description: Audit apps/api/queries/*.sql for primary-key lookups on org-scoped tables missing an org_id filter — this repo's single most emphasized invariant ("every tenant-scoped query must filter by org_id — silent data leak across tenants"). Use when asked to "check for tenant isolation gaps", "audit org_id filtering", "did this query leak across orgs", or after adding a new sqlc query.
---

# org_id audit

CLAUDE.md's Conventions section states this as a hard rule, not a
preference: **every tenant-scoped query must filter by `org_id`** —
skipping it is a silent cross-tenant data leak, not just a style issue.

## Steps

**1. Run the audit.**

```bash
python3 .claude/skills/org-id-audit/audit.py
```

Org-scoped tables are found by parsing `apps/api/migrations/*.sql`
directly (any `CREATE TABLE` whose body has an `org_id` column) — not
hardcoded, so a new migration adding `org_id` to a table is picked up
automatically without editing the script. A query is checked if it does
a primary-key lookup (`id = $N`) against one of those tables.

Exit `0` = every primary-key lookup on an org-scoped table either
filters by `org_id` or is a documented exception. Exit `1` lists real
findings separately from the "Known exceptions" section.

**2. For each real finding**, add `AND org_id = $N` to the query's
`WHERE` clause and thread an `orgID` parameter through the handler that
calls it — see any of the `Get*`/`Update*`/`Delete*` queries in
`apps/api/queries/monitors.sql` keyed by both `id` and `org_id` for the
existing pattern.

**3. If a finding is a true exception** (the id is genuinely not
attacker-controlled — see categories below), add it to
`KNOWN_EXCEPTIONS` in `audit.py` with which category applies and a
pointer to the verifying call site. Don't add an exception without
tracing the actual Go call site first — that's exactly the check this
skill exists to enforce.

## Known exception categories

Every current exception (verified against its call site when this audit
was authored) falls into one of three categories — a new exception
should fit one of these, not invent a fourth without strong justification:

1. **Worker-internal writes** — `UpdateCronMonitorDown`,
   `MarkUptimeMonitorDown`, `RecordPortCheckUp`, etc. The id comes from
   a monitor row the background worker already fetched via an
   org-agnostic-but-not-request-facing listing query (e.g.
   `ListOverdueCronMonitors`), not from any HTTP request. See
   `apps/api/internal/worker/worker.go`'s call sites.
2. **Self-lookup by authenticated identity** — `GetUserByID`,
   `UpdateUserPassword`, `AcceptUserTerms`, `TouchAPIKeyLastUsed`. The id
   is the caller's *own* ID from verified JWT claims or an
   already-authenticated API key (`apimiddleware.RequireAuth`/
   `RequireAPIKey`), never a caller-supplied arbitrary ID.
3. **Pre-scoped via a prior joined query** — `GetUptimeMonitorPublic`,
   `GetSSLMonitorPublic`, etc. Called from
   `apps/api/internal/handler/status_public.go`, where the monitor ID
   was already validated as belonging to the requested status page via
   a `status_page_monitors` join in an earlier query in the same
   request — not the raw URL.

If a finding doesn't clearly fit one of these, it's real — fix it,
don't force-fit an exception.

## Scope

Only covers `apps/api/queries/*.sql` — sqlc-generated raw SQL, per
ADR-004 (no ORM). Doesn't check that handlers actually pass the
authenticated org's ID into a query that does filter by `org_id`
correctly (e.g., a handler bug that passes the wrong variable) — that's
a code-review concern, not something greppable from the SQL alone.
