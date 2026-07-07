---
name: new-migration
description: Create a new goose migration, write the SQL, regenerate sqlc code, and run the multi-tenancy/org_id checklist before wiring it into queries and handlers. Use when asked to "add a migration", "add a column/table", or "create a new migration for X".
---

# New migration

Schema changes go through goose migrations by hand; sqlc then generates
the Go query layer from `apps/api/queries/*.sql` against that schema
(ADR-004 — no ORM). The two must be kept in sync manually.

## Steps

**1. Create the migration file.**

```bash
make migrate-create name=<snake_case_description>
```

Creates `apps/api/migrations/NNN_<name>.sql` with `+goose Up`/`+goose
Down` stubs, numbered sequentially after the highest existing migration.

**2. Write the SQL.** Both directions required — `Down` must cleanly
reverse `Up` (see `apps/api/migrations/030_sms_credit_quotas.sql` for a
minimal example: two `ALTER TABLE ADD COLUMN`s up, matching `DROP
COLUMN`s down, in reverse order). If the migration encodes a decision
worth remembering (why a default, why this shape), put a comment at the
top of the `Up` block referencing the ADR, not just an inline aside:

```sql
-- +goose Up

-- ADR-NNN (link the reasoning, don't restate it in full)
ALTER TABLE orgs ADD COLUMN ...;

-- +goose Down
ALTER TABLE orgs DROP COLUMN ...;
```

**3. Run the migration locally.**

```bash
make migrate
```

Reads `DATABASE_URL` from `apps/api/.env` — requires Postgres running
(`docker-compose up db` or inside the devcontainer).

**4. Write/update the SQL queries** in `apps/api/queries/*.sql`, then
regenerate:

```bash
cd apps/api && sqlc generate
```

This writes generated code into `apps/api/internal/db/` — never hand-edit
that directory.

**5. Multi-tenancy checklist** — before wiring the new query into a
handler:

- Every tenant-scoped query **must** filter by `org_id`. Silent
  cross-tenant data leaks are the failure mode if this is skipped.
- If the new table/column feeds into an existing plan-limit check (monitor
  count, channel count, etc. — see `docs/reference/limits.md` and
  ADR-019), confirm whether the limit logic needs updating too.

**6. Verify.**

```bash
cd apps/api && go test -count=1 ./...
```

`apps/api/package.json`'s `test` script excludes `internal/db` from
coverage (generated code) — don't add manual coverage exclusions
elsewhere to compensate.
