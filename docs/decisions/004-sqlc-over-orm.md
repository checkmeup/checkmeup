# ADR-004: sqlc for database access instead of an ORM

**Status:** Accepted  
**Date:** 2026-06-13

## Context

Go database access options:

- ORM (gorm, ent) — model structs, query builder, auto-migrations
- Query builder (squirrel, jet) — programmatic SQL construction
- Raw `database/sql` — manual scanning, no type safety
- sqlc — write SQL, get generated typed Go code

## Decision

Use sqlc. SQL is written by hand in `.sql` files; sqlc generates type-safe Go functions and structs from it. Migrations handled separately by goose.

## Consequences

- **Full SQL control:** complex queries, CTEs, window functions — no ORM impedance
- **Type safety:** generated code catches column/type mismatches at compile time, not runtime
- **No magic:** the generated code is readable and auditable; no hidden N+1 or lazy-load surprises
- **More upfront work:** adding a new query requires writing SQL + running `sqlc generate`, vs. an ORM's one-liner
- **Schema changes:** migrations must be kept in sync with queries manually; sqlc will catch mismatches if `sqlc generate` is run after each migration
