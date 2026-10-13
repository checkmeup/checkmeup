# Database: one shared PostgreSQL schema, hand-written SQL through sqlc

Date: 2026-10-13

All tenants share one PostgreSQL database and schema, with an `org_id` column on every tenant-scoped table; every query on such a table filters by `org_id`, since a missing filter silently leaks data across tenants. Queries are hand-written SQL in `apps/api/queries/`; sqlc generates the typed Go code (never edited by hand), and goose runs the migrations. Tests run against a real PostgreSQL, not mocks, so a missing `org_id` filter fails a test (ADR testing).

## Considered options

- A database or a schema per tenant: one database is one thing to back up, migrate and watch. Revisit when a compliance requirement or a large customer demands isolation; moving out is costly but possible.
- An ORM (gorm, ent) or a query builder: hides SQL, risks N+1 and lazy loads; sqlc catches column and type mismatches at compile time instead.

## Consequences

- A new query means writing SQL and running `sqlc generate`.
- Polymorphic join tables (`status_page_monitors`, `maintenance_window_monitors`, `monitor_notification_channels`) hold `monitor_type` + `monitor_id` without a foreign key, since monitors live in one table per type.
