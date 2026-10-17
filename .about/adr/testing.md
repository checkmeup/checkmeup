# Testing

Date: 2026-10-17

Go tests use the standard `testing` package, table-driven, with `testify/assert` only for assertions (not its suites or mocks), and run against a real PostgreSQL (`mise run db` locally, a service container in CI), never a mocked database, because mocks hide missing `org_id` filters (ADR database). The frontend uses Vitest. Coverage from both is converted to lcov and merged into one `coverage/lcov.info` for Codacy. Mutation testing with Stryker runs by hand on `apps/web/src/lib/`, never in CI.

## Considered options

- Database mocks in Go: faster, but they hide the tenant-isolation bugs that matter most here.
- End-to-end browser tests (Playwright, Cypress): not used.
