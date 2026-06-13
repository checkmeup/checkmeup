# ADR-010: Testing strategy

**Status:** Accepted  
**Date:** 2026-06-13

## Context

Needed a testing approach for both Go (backend) and Vue (frontend) before writing any code.

## Decision

**Go:** stdlib `testing` package with table-driven tests. `testify/assert` for readable assertions. Integration tests hit a real PostgreSQL instance — no DB mocks (see [ADR-002](002-multi-tenancy.md): mocks risk hiding `org_id` filter bugs).

**Frontend:** Vitest (already in the Turborepo pipeline). Unit tests for pure logic (stores, utils). No component tests on MVP — the UI surface changes too fast to justify the maintenance cost yet.

**Coverage:** collected per-package in Go, per-app in Vitest; merged by `tools/merge-coverage.mjs`.

## Consequences

- **No DB mocks in Go:** tests require a running PostgreSQL — handled by `docker-compose.yml` in local dev and a service container in CI
- **Table-driven tests:** idiomatic Go, easy to add cases without new test functions
- **No E2E on MVP:** Playwright or Cypress can be added post-launch once the UI stabilises
- **testify only for assertions:** not testify/suite or testify/mock — keeps tests simple and stdlib-compatible
