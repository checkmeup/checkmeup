---
title: Worker / Monitor-Check Architecture
type: knowledge
status: current
updated: 2026-07-29
tags: [architecture, monitoring, cron, uptime, ssl, domain, port, dns, backend]
scope: apps/api/internal/worker
superseded_by:
---

# Worker / monitor-check architecture

**Investigated:** 2026-07-10
**Scope:** `apps/api/internal/worker` — the background loop that checks every monitor type and dispatches alerts.

## Summary

One goroutine (`worker.Run`, started from `cmd/api/main.go`) drives everything. It is **not** goroutine-per-monitor — despite [ADR-001](../decisions/001-worker-model.md) describing that model, the as-built design is a single shared poll tick that lists due monitors from Postgres each round and fans work out to a bounded pool of per-check goroutines. See `docs/reference/tech-debts.md` for the open item tracking that ADR/code mismatch; this doc describes what's actually running.

## Findings

1. **One `Run` loop, two tickers** — `worker.go:58-79`. A 30 s ticker drives every check type sequentially (`checkOverdue` → `checkUptimeMonitors` → `checkSSLMonitors` → `checkDomainMonitors` → `checkPortMonitors` → `checkDNSMonitors`); a separate 24 h ticker runs `pruneOldPings` (retention cleanup, [ADR-015](../decisions/015-cron-pings-retention.md)). No per-monitor `time.Ticker` exists anywhere.

2. **Poll-then-fan-out, not goroutine-per-monitor** — each `check*Monitors` function (e.g. `checkUptimeMonitors`, `worker_uptime.go:24-44`) runs one `ListDue*Monitors` query, then spawns one goroutine per due monitor, bounded by a shared semaphore: `sem := make(chan struct{}, checkConcurrency)` where `checkConcurrency = 50` (`worker.go:29`). This caps *concurrent in-flight checks per tick*, not total monitors — a org with thousands of monitors just takes longer per tick rather than spawning thousands of goroutines. This is a materially different scaling/failure profile than ADR-001's description: under load, this model degrades gracefully by delaying checks; goroutine-per-monitor would instead accumulate long-lived goroutines.

3. **One file per monitor type**, split out of a single `worker.go` that had grown to ~1070 logical lines (commit `61ef075`, tracked by the `architecture-guardrails` skill's 700-line threshold):
   - `worker_cron.go` — `checkOverdue`: cron monitors are ping-driven (the monitor's own schedule triggers an inbound `POST /ping`, handled in `internal/handler/ping.go`, not here); this loop only detects *missed* pings past the grace period.
   - `worker_uptime.go` — `checkUptimeMonitors` / `performHTTPCheck`: HTTP request using the monitor's configured method (GET/HEAD/POST, GET by default), bounded by a per-monitor `context.WithTimeout` derived from `max_response_time_ms` rather than a fixed client timeout ([EP-37](../stories/ep-37-configurable-uptime-checks.md), 2026-07-18); evaluates accepted-status-code membership → keyword → JSON assertions in that order, first failure wins. The response-time threshold used to be a fourth post-hoc step here but is now the request timeout itself, enforced before any of the three checks can run.
   - `worker_ssl.go` — `checkSSLMonitors` / `performTLSCheck`: TLS handshake, reads the leaf cert's `NotAfter`; alerts at 30/14/7-day thresholds and on expiry (`sslCrossedThreshold`).
   - `worker_domain.go` — `checkDomainMonitors`: RDAP lookup via `internal/rdap`, same 30/14/7-day threshold pattern as SSL (`domainThresholdAlert` mirrors `sslThresholdAlert`), differing only in data source (RDAP vs TLS handshake) and field names (registrar vs issuer).
   - `worker_port.go` — `checkPortMonitors` / `performTCPCheck`: raw TCP dial, no data sent; a monitor's `ExpectedState` (open/closed) decides whether a successful connect means up or down (US-3302).
   - `worker_dns.go` — `checkDNSMonitors` / `performDNSLookup`: real DNS lookup for the monitor's record type (A/AAAA/CNAME/MX/TXT/NS), sorted-and-joined so multi-value answer reordering doesn't read as a change; compares against `expected_value` (user-pinned, or auto-captured from the first successful lookup via `RecordDNSCheckUp`'s `COALESCE` — EP-39, 2026-07-29). A lookup error (NXDOMAIN/SERVFAIL/timeout) is a distinct `failure_reason` from a value mismatch, so the down alert text says the right thing either way.
   - Shared across all six: `checkConcurrency`, the `Notifiers` struct, `AlertMessage`/`MonitorRef`, and `DispatchAlert` — all still in `worker.go`.

4. **Every user-supplied dial target is SSRF-guarded.** Uptime (`m.Url`), port (`m.Host`), and SSL (`hostname`) checks all go through `internal/httpsafe.Dialer`, which rejects loopback/private/link-local (including the `169.254.169.254` cloud-metadata address)/unspecified/multicast destinations at dial time — added 2026-07-10 after uptime/port/SSL checks were found dialing arbitrary user-supplied hosts with no IP restriction (the Slack/webhook notification clients already had this guard; see [notification-channels.md](notification-channels.md)). Domain checks (RDAP) aren't user-host-controlled in the same way — RDAP servers are resolved from the registry, not the monitored domain's own infrastructure — so they don't need the same guard. DNS checks don't need it either, for a different reason: a DNS lookup never dials the resolved address, only resolves it — there's simply no connect step to guard (`Notifiers.DNSResolver` has no SSRF-hardening wrapper, unlike `HTTPClient`/`TCPDialer`). DNS's own alert-building code has a different injection surface instead — see the 2026-07-29 fix in `git log` for `worker_dns.go` (resolved DNS values are `html.EscapeString`'d before reaching Telegram/email, since that text comes from the monitored domain's own DNS records, not from the alerted org). Uptime/port checks take an injectable client/dialer via `Notifiers.HTTPClient`/`Notifiers.TCPDialer` (nil → hardened default in production) so tests can still reach a local `httptest` server; DNS checks take an injectable `Notifiers.DNSResolver`, tested against an in-memory fake DNS server (`net.Pipe`) rather than a real socket.

5. **`Notifiers` is the one dependency-injection seam** (`worker.go:35-53`): every check/alert-dispatch function takes a `Notifiers` value bundling `*db.Queries` and one client per channel (`Telegram`, `Mailer`, `Webhook`, `Slack`, `SMS`, `RDAP`), plus `Logger`. Production wires real clients in `cmd/api/main.go`; tests substitute test-double clients (`NewClientWithHTTPClient`, `rdap.NewClientWithHTTPClient`, etc.) built against a local `httptest` server. `internal/handler/ping.go` builds its own `Notifiers` value to route a cron-recovery event through the same `DispatchAlert` path the worker uses, rather than duplicating channel-dispatch logic.

6. **Alert dispatch always lands somewhere.** `DispatchAlert` (`worker.go:119-138`) sends to every channel attached to the monitor via `monitor_notification_channels`; if none are attached, or every attached channel fails, it falls back to emailing every user in the org (`dispatchFallbackEmail`) rather than staying silent ([ADR-023](../decisions/023-notification-channels.md)). See [notification-channels.md](notification-channels.md) for the channel-dispatch mechanics themselves.

## Follow-ups

- `docs/reference/tech-debts.md` already tracks the ADR-001/code mismatch and the "no shared HTTP client across checks" debt (partially superseded now that uptime/port/SSL checks share the `httpsafe.Dialer` construction pattern, though each check still builds its own client/dialer rather than reusing one across calls).
- ADR-001 itself hasn't been corrected — see the update note added there 2026-07-10 pointing back to this doc, rather than rewriting the original decision text (decisions/ is append-only).
