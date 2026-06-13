# Decision backlog

Open questions that need an answer before (or early in) the relevant implementation. When decided, create a numbered ADR and remove the entry here.

---

## Testing strategy

**Question:** What are the testing conventions for Go and for the frontend?

Go options: standard `testing` package with table-driven tests, testify for assertions, integration tests hitting a real DB (see [ADR-002](002-multi-tenancy.md) — mocks risk hiding `org_id` filter bugs).  
Frontend options: Vitest (already in the Turborepo pipeline), component tests with Vue Test Utils, E2E with Playwright.

**Needed before:** writing the first handler or component.

---

## API error response format

**Question:** What shape do error responses take?

Options: plain `{"error": "message"}`, RFC 7807 Problem Details (`application/problem+json`), or a custom envelope.  
Affects: frontend error handling, API docs, and client SDK if we ever ship one.

**Needed before:** writing the first API handler.

---

## API versioning

**Question:** Do we version the API, and if so, how?

Options: `/api/v1/` URL prefix (simplest), `Accept` header versioning, no versioning (break freely while pre-1.0).  
Given MVP pace, "no versioning until a public API is promised" may be the right call — but it should be a deliberate choice.

**Needed before:** writing the first route.

---

## Logging and observability

**Question:** Which logging library for Go, and what structure/level convention?

Options: stdlib `log/slog` (no dependency, structured), zerolog (zero-alloc, popular in Go services), zap.  
Decision affects how errors surface in production and what Kamal/Traefik ship to stdout.

**Needed before:** writing the first handler.

---

## Alert debounce / cooldown

**Question:** What is the logic for sending a Telegram alert vs. staying silent?

Specifics to decide:
- How many consecutive failures before alerting? (e.g. 2-of-3)
- How long to wait before re-alerting on a still-down monitor?
- Alert on recovery? (monitor back up)
- Per-monitor override vs. global setting?

**Needed before:** implementing the alert system.

---

## Uptime check mechanics

**Question:** How exactly does an uptime check work?

Specifics to decide:
- HTTP method: HEAD (lighter) or GET (catches body-level errors)?
- Timeout per request
- What counts as "down": non-2xx, connection timeout, both?
- Check interval granularity (minimum 1 min? 30 s?)
- Follow redirects or flag them?

**Needed before:** implementing the uptime monitor worker.

---

## Chi router rationale

**Question:** Document why Chi over alternatives (gin, echo, fiber, stdlib `net/http`).

Likely reasons: idiomatic stdlib-compatible `http.Handler`, good middleware ecosystem, no magic — but this should be confirmed and recorded as an ADR so the choice isn't relitigated.

**Needed before:** not blocking — but worth capturing when the first route is written.
