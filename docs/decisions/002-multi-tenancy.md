# ADR-002: Single database with org_id columns, no schema-per-tenant

**Status:** Accepted  
**Date:** 2026-06-13

## Context

Multi-tenant SaaS has three common isolation models:
1. Separate database per tenant
2. Separate schema per tenant (PostgreSQL schemas)
3. Shared schema with a `tenant_id` / `org_id` discriminator column on every table

## Decision

Single PostgreSQL database, shared schema, `org_id` column on every tenant-scoped table. All queries must filter by `org_id`.

## Consequences

- **Simpler ops:** one database to backup, migrate, and monitor
- **Lower cost:** no per-tenant DB overhead on MVP
- **Row-level isolation risk:** a missing `org_id` filter in a query leaks data across tenants — mitigated by code review and sqlc's typed queries
- **Migration path:** schema-per-tenant or DB-per-tenant can be adopted later if a compliance requirement or large enterprise customer demands it; migrating out of this model is non-trivial but possible
