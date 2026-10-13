# Overlapping run detection

## Problem

A job that starts again before its previous run finished (a slow run plus a short schedule) can double-process data or deadlock, and nothing reports it.

## Users

Cron-monitor users who send start pings (PRD zombie-job-detection).

## Success signal

Every overlapping start is reported, naming when the still-running run began.

## Scope

- Flagging a start ping that arrives while an earlier run is still open (ADR workers).
- An overlap alert through the monitor's channels, within its alert cap.
- Overlaps in the incident log.

## Non-goals

- Preventing the overlap (locking the job).

## Acceptance criteria

- A start ping while another run of the same monitor is open is flagged as an overlap; a third start while two are open is flagged too.
- A monitor that never sends start pings never gets an overlap.
- An overlap alert says it is an overlap and when the still-running run started, and goes to the monitor's channels.
- Overlap alerts count against `max_alerts_per_incident`.
- The incident log lists overlaps apart from missed-ping incidents, with both runs' start times and the durations of any finished ones.

## Open questions

- Whether an overlap becomes a `cron_incidents` row with a kind column or has its own model (ADR workers left this open).
