# Stuck (zombie) job detection

## Problem

A job stuck in a loop or hung on a lock eventually pings, or never fails cleanly, so missed-ping detection doesn't catch it while it burns resources for hours.

## Users

Cron-monitor users whose jobs can hang.

## Success signal

A run that goes past its maximum duration is reported once, as stuck, while it is still running.

## Scope

- An optional start ping, `GET /ping/{token}/start` (ADR workers).
- An optional maximum run duration per monitor.
- One "stuck run" alert per run, and a recovery when it completes or a new run starts.
- Run durations in the ping log.

## Non-goals

- Killing or restarting the user's job.
- Overlapping runs (PRD overlap-detection).

## Acceptance criteria

- A start ping records a run without changing the monitor's last ping, next expected ping or status.
  → test: apps/api/internal/handler/ping_test.go "TestReceivePingStart/records a run without touching monitor ping state (ADR-039)"
- A completion ping closes the open run; with no open run it is accepted exactly as before.
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/completion ping closes the open run and stamps run_started_at (US-3401/US-3404)"
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/completion ping with no open run is accepted as-is (US-3401)"
- A maximum duration can be set on create and edit; without one, runs are never checked for being stuck.
  → test: apps/api/internal/handler/monitors_test.go "TestCreateCronMonitor/success with maxDurationMins set (US-3402)"
  → test: apps/api/internal/worker/worker_test.go "TestCheckStuckCronRuns/a monitor without max_duration_mins set is never selected"
- A run past its maximum duration sends one alert, never repeated for that run; a run within it is left alone.
  → test: apps/api/internal/worker/worker_test.go "TestCheckStuckCronRuns/alerts once on a run that exceeded max_duration_mins, then marks it alerted"
  → test: apps/api/internal/worker/worker_test.go "TestCheckStuckCronRuns/does not re-alert a run that was already flagged"
  → test: apps/api/internal/worker/worker_test.go "TestCheckStuckCronRuns/a run still within its max_duration_mins budget is left alone"
- A stuck run's completion, or a new start ping replacing it, sends a recovery alert; a run never flagged sends none.
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/recovery alert fires on completion when the closed run had already been flagged stuck (US-3403)"
  → test: apps/api/internal/handler/ping_test.go "TestReceivePingStart/recovery alert fires when a new start ping supersedes an already-alerted open run (US-3403)"
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/no recovery alert when the open run was never flagged stuck"
- Under maintenance, or with alerts off, a stuck run sends nothing.
  → test: apps/api/internal/worker/worker_test.go "TestCheckStuckCronRuns/a monitor under an active maintenance window is excluded"
  → test: apps/api/internal/worker/worker_test.go "TestCheckStuckCronRuns/alerts_enabled=false still marks the run alerted (so it isn't re-evaluated) but sends no alert"
- The ping log shows each completed run's duration, and "—" for a ping without a start.
  → test: apps/web/src/views/CronMonitorDetailView.test.ts "shows a run duration for a completion ping with a matching start ping (US-3404)"
  → test: apps/web/src/views/CronMonitorDetailView.test.ts "shows '—' for a ping with no matching start ping"
