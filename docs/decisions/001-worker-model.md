# ADR-001: Goroutine-per-monitor worker model, no external queue

**Status:** Accepted  
**Date:** 2026-06-13

## Context

Monitors need to fire on a schedule (cron) or be polled on an interval (uptime, SSL). Common approaches are:

- A job queue backed by Redis/BullMQ/etc. with worker processes consuming it
- A goroutine per monitor with a `time.Ticker`
- A single scheduler loop iterating over due monitors

## Decision

Each active monitor gets its own goroutine with a `time.Ticker`. Cron monitors wait for inbound pings; a separate ticker checks for missed pings. All state lives in PostgreSQL — no Redis or external queue.

## Consequences

- **Simpler ops:** no Redis process to run, monitor, or pay for on MVP
- **Simpler code:** no queue producer/consumer abstraction, just goroutines
- **Scaling ceiling:** goroutine-per-monitor works fine up to tens of thousands of monitors on a single node; beyond that, a queue-based model would be needed
- **Statefulness:** the API process is stateful — horizontal scaling requires a coordination layer (etcd, Postgres advisory locks, or migrating to a queue) if we ever need multiple API instances running monitors

## Update (2026-07-10)

The as-built worker is **not** goroutine-per-monitor with a per-monitor `time.Ticker`, despite the Decision text above. It's a single shared 30 s poll tick that queries for due monitors each round and fans out a bounded semaphore of goroutines per check type (`checkConcurrency = 50`) — a materially different scaling profile (poll-tick degrades gracefully by delaying checks under load; the originally-decided model would instead accumulate long-lived goroutines). Still no Redis/external queue, so the core decision (Postgres-only state, no queue) holds. Caught during the 2026-07-04 capacity-planning discussion, tracked in `docs/reference/tech-debts.md`. Left as an addendum rather than rewritten, per `decisions/`'s append-only convention — see [knowledge/worker-architecture.md](../knowledge/worker-architecture.md) for the as-built model.
