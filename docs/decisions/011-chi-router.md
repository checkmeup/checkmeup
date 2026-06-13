# ADR-011: Chi as the HTTP router

**Status:** Accepted  
**Date:** 2026-06-13

## Context

Go HTTP router options considered:

- **stdlib `net/http`** — no routing, only `http.ServeMux`; pattern matching is limited (no path params before Go 1.22, and even the Go 1.22 pattern syntax is minimal)
- **gin** — fast, popular, but opinionated request context (wraps `*http.Request`) and heavier dependency surface
- **echo** — similar to gin; own context type breaks middleware interoperability with stdlib ecosystem
- **fiber** — built on fasthttp instead of `net/http`; fastest benchmarks but incompatible with the entire stdlib `net/http` middleware ecosystem
- **chi** — thin wrapper over `net/http`; uses standard `context` for route params; fully compatible with any `http.Handler` middleware

## Decision

Use Chi (`github.com/go-chi/chi/v5`). It provides URL parameter routing, subrouters (`r.Route`, `r.Group`), and middleware chaining while staying 100% compatible with the `net/http` interface.

## Consequences

- **stdlib compatibility:** any `net/http`-compatible middleware (e.g. `go-chi/cors`, `go-chi/httprate`) works without adapters
- **No context lock-in:** route params are accessed via `chi.URLParam(r, "id")` using standard `context`; handlers stay portable
- **Minimal surface:** Chi adds routing only — no templating, no DI, no ORM; each concern is a separate, swappable package
- **Maturity:** chi v5 is stable and widely used in production Go services
