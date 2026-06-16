# EP-08: Security hardening

Cross-cutting backend hardening. No new user-visible functionality. All stories completed Jun 15 as a single session before Phase 3.

> **US-0106** (rate-limit auth endpoints) is owned by [EP-01](ep-01-auth.md) and is not repeated here.

---

### US-0801: Rate-limit the ping endpoint

**As a** platform operator, **I want** the ping endpoint to be rate-limited **so that** a misconfigured or abusive client cannot flood the database.

**Estimate:** 0.25 h

**Acceptance criteria:**

- [x] `GET /ping/{token}` limited to 60 req/min per token
- [x] Excess requests return 429 Too Many Requests
- [x] Limit is per-token, not per-IP, to avoid penalising NAT'd clients

---

### US-0802: Rate-limit Telegram endpoints

**As a** platform operator, **I want** the Telegram webhook and test-message endpoints rate-limited **so that** attackers cannot abuse the bot or trigger alert spam.

**Estimate:** 0.25 h

**Acceptance criteria:**

- [x] `POST /webhook/telegram` limited to 60 req/min per IP
- [x] `POST /settings/telegram/test` limited to 5 req/min per IP
- [x] Excess requests return 429 Too Many Requests

---

### US-0803: Cap request body size globally

**As a** platform operator, **I want** all incoming request bodies capped at 64 KB **so that** a malicious client cannot send a multi-GB payload and exhaust memory on the 4 GB server.

**Estimate:** 0.25 h

**Acceptance criteria:**

- [x] `http.MaxBytesReader` applied as global middleware before any handler reads the body
- [x] Payloads exceeding 64 KB result in a 413 error and the connection is closed
- [x] Legitimate API payloads (JSON bodies, Telegram updates) are well under the limit

---

### US-0804: Validate Telegram webhook secret

**As a** platform operator, **I want** incoming Telegram webhook calls verified with a secret token **so that** an attacker cannot inject fake updates by directly calling the webhook URL.

**Estimate:** 0.25 h

**Acceptance criteria:**

- [x] Secret token derived as `sha256(TELEGRAM_BOT_TOKEN)` — no separate config value needed
- [x] Secret registered with Telegram via `setWebhook` on startup
- [x] Every incoming update validates the `X-Telegram-Bot-Api-Secret-Token` header; mismatch returns 401
