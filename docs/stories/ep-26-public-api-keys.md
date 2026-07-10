# EP-26: Public API and API key management

A programmatic API for scripts, CI pipelines, and third-party integrations — distinct from the existing browser-session auth ([EP-01](ep-01-auth.md)).

**Conflicts with an existing rule — needs an explicit ADR amendment, not a silent workaround** (add to [decision backlog](../decisions/backlog.md)): [ADR-003](../decisions/003-auth-jwt-httponly-cookie.md) states "no `Authorization` header — the cookie is sent automatically by the browser" and CLAUDE.md lists this as a hard "Don't" (`Use Authorization header for auth — the access_token httpOnly cookie is the only auth mechanism`). That rule was written for browser session auth (its CSRF/XSS rationale doesn't apply to non-browser clients, which never auto-send anything and have no cookie jar). A public API needs a *separate* mechanism for non-browser clients — this epic proposes a dedicated `X-API-Key` header (deliberately not `Authorization`, to keep the two mechanisms visibly distinct in code and docs) and ADR-003 should be amended to scope its rule to session auth, not silently reinterpreted or ignored.

---

### US-2601: Generate an API key

**As a** user, **I want** to generate an API key **so that** I can call the checkmeup API from scripts or other tools without a browser session.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [x] Settings: "Create API key" action with an optional label
- [x] Key value shown once at creation, never retrievable again — only its hash is stored, same pattern already used for `password_hash`/`refresh_tokens.token_hash`
- [x] Multiple keys per org supported, each independently revocable
- [x] Key format includes a recognizable prefix (e.g. `cmu_live_...`) for easy identification in logs and secret scanners

---

### US-2602: Authenticate API requests with a key

**As a** developer, **I want** to authenticate programmatic API requests with an API key **so that** I can integrate without a browser session.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [x] API key sent via a dedicated `X-API-Key` header (see the ADR-003 note above) — not `Authorization`
- [x] Valid key resolves to its owning org for every `org_id`-scoped query, same multi-tenancy rules as session auth ([ADR-002](../decisions/002-multi-tenancy.md))
- [x] Invalid or revoked key returns `401` in the existing error format ([ADR-008](../decisions/008-api-error-format.md))

---

### US-2603: Scope what an API key can do

**As a** user, **I want** to limit what an API key is allowed to do **so that** a leaked key from one integration can't do everything my account can.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Each key has a scope: read-only or read-write (full per-endpoint permissions deferred past MVP)
- [ ] Read-only keys reject any non-`GET` request with `403`
- [ ] Scope set at creation and not editable after — revoke and recreate instead, keeps the implementation simple

---

### US-2604: View and revoke API keys

**As a** user, **I want** to see all active API keys and revoke any of them **so that** I can clean up unused or compromised keys.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] List shows: label, scope, created date, last-used date, masked key (e.g. `cmu_live_••••1234`) — not built as specified: no scope column (US-2603 isn't built), and the key is shown as a bare prefix (`cmu_live_…`) rather than a `••••`-masked suffix. See [backlog.md](backlog.md)'s EP-26 footnote.
- [x] Revoke takes effect immediately — the next request with that key returns `401`
- [x] Last-used timestamp updates asynchronously and never blocks the request it's attached to

---

### US-2605: Rate-limit and document the public API

**As a** developer, **I want** documented endpoints and predictable rate limits **so that** I can build a reliable integration.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [x] Public API docs page listing available `/api/v1/` endpoints, request/response shapes, and the `X-API-Key` header
- [x] Per-key rate limit (e.g. 60 req/min), consistent with the existing rate-limiting pattern ([ADR-013](../decisions/013-rate-limiting.md))
- [x] Rate-limited responses return `429` with `Retry-After`, same as other rate-limited endpoints
