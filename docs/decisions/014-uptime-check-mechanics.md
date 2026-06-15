# ADR-014: Uptime check mechanics

**Status:** Accepted  
**Date:** 2026-06-15

## Context

The uptime monitor worker needs a concrete definition of how to probe a URL and when to declare it "down". Options to decide:

- HTTP method: GET, HEAD, or HEAD-with-GET-fallback
- What response counts as "up" vs "down"
- Timeout per request
- Minimum check interval
- Redirect handling

## Decision

**Method:** `GET` always. HEAD is cheaper but many servers return different status codes for HEAD vs GET, and some don't support HEAD at all. GET is the authoritative check.

**"Down" condition:** any response code other than `200`, or a request timeout. This is intentionally strict — a `204`, `301`, or `503` all count as down. Most health-check endpoints return `200`; anything else signals a problem worth alerting on.

**Timeout:** 10 seconds per request (matches the existing Telegram client timeout; adequate for any healthy endpoint).

**Minimum check interval:** 10 minutes for MVP. Keeps the outbound HTTP load negligible and avoids hitting rate limits on checked URLs. Lower intervals (1 min) added in a later billing tier.

**Redirects:** followed automatically (Go's `http.Client` default). A redirect chain that ultimately resolves to `200` is considered up. A redirect loop or a redirect to a non-200 final response is down.

## Consequences

- Simple worker: one `http.Get` call with a 10-second timeout, check `resp.StatusCode == 200`.
- No retries before alerting on MVP — a single failure triggers "down". Debounce / consecutive-failures logic deferred to the alert debounce ADR (still in backlog).
- 10-minute minimum interval means the fastest an outage is detected is 10 minutes after it starts (plus the grace period / alert debounce once that's decided).
- Strict 200-only rule may generate false alerts for endpoints that legitimately return `204`. Acceptable on MVP; relax to 2xx if users report it.
