# How monitors are checked: the worker, check mechanics, maintenance and cron runs

Date: 2026-10-13

The worker runs inside the API process with no Redis, job queue or external broker: one shared 30-second tick queries each monitor type for what is due and fans the checks out over a bounded semaphore of goroutines per type (`checkConcurrency = 50`). Under load this delays checks rather than piling up goroutines, and all state stays in PostgreSQL.

An uptime check is one `GET` with the redirects followed; by default it is up only on a final `200` within 10 seconds. Method, timeout and accepted status codes are per-monitor settings (PRD configurable-uptime-checks); these are the defaults, kept strict because most health endpoints answer `200` and anything else is worth an alert. `GET` rather than `HEAD`, because many servers answer `HEAD` differently or not at all.

A maintenance window hides its monitors from the due/overdue queries, the same way `paused` does, so no check runs, no incident opens, no alert is sent and uptime stats stay clean without a separate exclusion step. Windows are one-off (`starts_at`, optional `ends_at` ended by hand) and cover monitors of any type through a polymorphic join table.

A cron job may send `GET /ping/{token}/start` before it runs, as well as the completion ping `GET /ping/{token}`. Each start opens a row in `cron_runs`; a start while another run is open is marked as an overlap; the completion ping closes the latest open run and is a no-op when there is none. The start ping never touches `last_ping_at`, `next_ping_at` or `status`, so missed-ping detection works exactly as for a monitor that never sends a start ping.

## Considered options

- A goroutine and `time.Ticker` per monitor (until 2026-07-10): the original decision, never built that way; the shared tick degrades by delaying checks instead of accumulating long-lived goroutines.
- A job queue (Redis/BullMQ): an extra process to run and pay for, and a producer/consumer layer, for no need at this scale.
- `HEAD`, or `HEAD` with a `GET` fallback, for uptime checks: cheaper, but servers disagree on `HEAD`.
- Teaching the worker an "in maintenance" branch: every check path would need it; the query exclusion needs none.
- A recurring-schedule (RRULE) engine for maintenance windows: deferred; recurrence can later insert concrete windows into the same tables.
- `?state=start` on the existing ping URL: query parameters there are already stored as user metadata.
- A `run_started_at` column on `cron_monitors`: holds only the current run, and two near-simultaneous starts can both miss each other; `cron_runs` turns that race into a false-positive overlap instead of a missed one.

## Consequences

- The API process is stateful: running more than one instance needs coordination (advisory locks or a queue) so checks don't run twice.
- A maintenance window that starts mid-incident freezes the incident until the window ends; a cron monitor still overdue at that moment can go down again on the next tick.
