# HTTP API conventions: routing, versioning, errors and logging

Date: 2026-10-13

The API is built on Chi, a thin router over `net/http`, so any standard middleware (cors, httprate) works without adapters. Every route sits under `/api/v1/`; a breaking change gets `/api/v2/` alongside, so old clients keep working. Errors are `{"error": "<message for people>", "code": "<snake_case_code>"}` with the HTTP status as the category; the frontend branches on `code`, never on the message. The Go side logs with `log/slog`, JSON in production and text in development; the frontend uses `console` only.

## Considered options

- gin, echo or fiber: their own context types (or fasthttp) break the `net/http` middleware ecosystem.
- `Accept`-header versioning: invisible in DevTools and awkward with curl.
- RFC 7807 Problem Details: more fields than a single-client API needs.
- zerolog or zap: only pay off at a throughput this API doesn't have.
