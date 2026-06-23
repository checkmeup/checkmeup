# ADR-013: In-process rate limiting with go-chi/httprate

**Status:** Accepted  
**Date:** 2026-06-15

## Context

Several public endpoints have no request-frequency guard:

| Endpoint | Abuse scenario |
|---|---|
| `POST /auth/sign-up` | Account farming; bcrypt is CPU-heavy, so mass sign-ups burn CPU |
| `POST /auth/sign-in` | Credential brute force |
| `POST /auth/forgot-password` | Email bombing a victim; burns Resend monthly quota |
| `GET /ping/{token}` | Ping flooding a monitor token; inflates the `cron_pings` table unboundedly |

Options considered:

- **Traefik middleware** — rate limiting is available via the `InFlightReq` and `RateLimit` plugins. It works at the reverse-proxy layer and requires no code changes. Downside: limits apply per-route only with paid Traefik Enterprise; free Traefik applies them globally or per-service, not per-IP/per-token at the application route level.
- **Redis + token bucket** — precise distributed rate limiting, survives restarts, supports per-user limits. Overkill for a single-server MVP and contradicts ADR-001 (no external broker).
- **`go-chi/httprate`** — in-process sliding window counter, per-IP or custom key, drop-in Chi middleware. Resets on process restart (acceptable: limits are short-window, 1–60 minutes; a restart clears them but that's a minor leak). No external dependency.

## Decision

Use **`go-chi/httprate`** applied per-route in `buildRouter()`. Limits are enforced at the application layer. Each route gets the narrowest key that makes sense (IP for auth endpoints, ping token for the ping endpoint).

Limits chosen:

| Endpoint | Limit | Window | Key |
|---|---|---|---|
| `POST /sign-up` | 5 | 1 hour | IP |
| `POST /sign-in` | 10 | 10 minutes | IP |
| `POST /forgot-password` | 3 | 10 minutes | IP |
| `GET /ping/{token}` | 60 | 1 minute | token |
| `GET /status/:slug/badge.svg` | 300 | 1 minute | IP |
| `GET /status/:slug/badge/:monitor_id.svg` | 300 | 1 minute | IP |

The ping limit (60/min) is deliberately generous — a real cron job might hit the endpoint once per minute at most, but we want to leave headroom for retries and misconfigured jobs before blocking. The real protection is against floods, not occasional bursts.

Badges (EP-30, [story](../stories/ep-30-status-badges.md)) are keyed by IP rather than slug: unlike the ping token, a badge URL isn't a secret, so keying by slug would let one popular badge's many embedders collectively exhaust the limit for everyone viewing that same page. 300/min is generous because badges are meant to be embedded in READMEs and external sites; `Cache-Control: max-age=60` on the response (US-3004) is the actual defense against repeated hits, the rate limit is just a backstop.

## Consequences

- Rate-limited requests return **HTTP 429** with a `Retry-After` header (provided by `httprate` automatically).
- Limits reset on API process restart — acceptable for MVP; a restart is a brief window, not a sustained bypass.
- IP-based limits can be bypassed from distributed IPs, but this raises the cost of abuse significantly for a bootstrapped product.
- No changes to the database schema or Telegram alert logic — this is purely a middleware concern.
- Alert debounce (flapping monitor spam) is a separate concern tracked in `backlog.md`.
