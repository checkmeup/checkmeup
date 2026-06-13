# ADR-009: log/slog for Go logging, console for frontend

**Status:** Accepted  
**Date:** 2026-06-13

## Context

Go options: `log/slog` (stdlib, Go 1.21+), zerolog (zero-alloc), zap (high-performance).  
Frontend options: logging library (pino, loglevel), or browser `console`.

## Decision

**Go:** `log/slog` from the standard library. Structured JSON output in production (`slog.New(slog.NewJSONHandler(os.Stdout, nil))`), text output in development.

**Frontend:** `console.log / .warn / .error` only. No logging library.

## Consequences

- **Zero dependencies:** slog ships with Go 1.21+ — no `go get`, no version to pin
- **Structured output:** JSON logs are readable by Traefik and any future log aggregator without a parsing step
- **Performance:** slog is fast enough for an MVP API; zerolog/zap only matter at high throughput (10k+ req/s)
- **Frontend:** DevTools is sufficient during MVP; a production error tracker (Sentry) can be added post-launch if needed
