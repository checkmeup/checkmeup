# Cron monitoring: alert when a scheduled job doesn't check in

## Problem

A scheduled job that silently stops running (a crashed crontab, a dead worker) loses data or work, and nobody notices until the damage shows.

## Users

Developers running cron jobs, scheduled tasks or queue workers.

## Success signal

A missed run is reported within the job's grace period, without the user writing any monitoring code beyond one `curl` to a ping URL.

## Scope

- A cron monitor with a schedule (cron expression or interval like "every 1h") and a grace period, and a secret ping URL.
- The ping endpoint; missed-ping detection; down and recovery alerts.
- List, detail (ping log, incidents), edit, pause and resume, delete.

## Non-goals

- Running the user's jobs.
- Detecting stuck or overlapping runs (PRD zombie-job-detection, PRD overlap-detection).

## Acceptance criteria

- Creating a monitor with a name, a 5-field cron expression or an interval and a grace period gives an unguessable ping URL; the monitor starts as "waiting", not "down".
  → test: apps/api/internal/handler/monitors_test.go "TestCreateCronMonitor/success with an interval schedule, applying defaults"
  → test: apps/api/internal/handler/monitors_test.go "TestCreateCronMonitor/success with a 5-field cron expression"
- A missing name or schedule is refused with a field error.
  → test: apps/web/src/views/CronMonitorCreateView.test.ts "shows a validation error when name is missing"
  → test: apps/web/src/views/CronMonitorCreateView.test.ts "shows a validation error when schedule is missing"
- The ping URL can be copied from the detail page in one click.
  → test: apps/web/src/views/CronMonitorDetailView.test.ts "copies the ping URL to the clipboard when Copy is clicked"
- `GET /ping/{token}` with a known token, without any auth header, records the ping with its time and source IP, sets the monitor up and schedules the next expected ping.
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/success records a ping and flips status to up"
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/source IP resolution"
- The next expected ping is the schedule's next time plus the grace period; an unparseable schedule falls back to one hour.
  → test: apps/api/internal/handler/ping_test.go "TestComputeNextPing/interval schedule adds the interval plus grace"
  → test: apps/api/internal/handler/ping_test.go "TestComputeNextPing/cron expression uses the next scheduled fire time plus grace"
  → test: apps/api/internal/handler/ping_test.go "TestComputeNextPing/unparseable schedule falls back to 1 hour plus grace"
- A ping to an unknown or deleted monitor's token gets `404`.
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/unknown token returns 404"
- A ping's query string is kept as metadata and shown in the ping log.
  → test: apps/api/internal/handler/monitors_test.go "TestGetCronPings/includes a ping's query-string metadata"
  → test: apps/web/src/views/CronMonitorDetailView.test.ts "renders a ping's metadata as key: value chips"
- Within 30 seconds of the grace period running out without a ping, the monitor goes down, an incident opens and an alert is sent.
  → test: apps/api/internal/worker/worker_test.go "TestCheckOverdue/marks an overdue monitor down, opens an incident, and delivers an alert"
- With alerts disabled, the incident still opens but no alert is sent.
  → test: apps/api/internal/worker/worker_test.go "TestCheckOverdue/an incident still opens when alerts are disabled, but no alert is sent"
- The next ping after a missed one resolves the incident and sends a recovery alert.
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/recovery resolves the open incident when alerts are enabled"
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/recovery delivers a webhook event via the monitor's attached channel"
- The list shows each monitor's name, status and schedule, and an empty state with a create button when there are none.
  → test: apps/web/src/views/MonitorListView.test.ts "renders a cron row with its name, status, and schedule"
  → test: apps/web/src/views/MonitorListView.test.ts "shows an empty state with a call to action when there are no monitors"
- The detail page shows the configuration, a paginated ping log newest first, and incidents with their duration.
  → test: apps/api/internal/handler/monitors_test.go "TestGetCronMonitor/success returns pings and incidents newest first"
  → test: apps/api/internal/handler/monitors_test.go "TestGetCronPings/pagination"
  → test: apps/web/src/views/CronMonitorDetailView.test.ts "renders the incident list with resolved duration and an ongoing entry"
- Editing changes the name, schedule and grace period but never the ping URL; the new schedule applies on the next tick.
  → test: apps/api/internal/handler/monitors_test.go "TestUpdateCronMonitor/success applies changes and defaults"
  → test: apps/web/src/views/CronMonitorEditView.test.ts "saves changes and navigates to the detail view on success"
- A paused monitor keeps accepting pings but is neither marked up nor alerted on; resuming restores checking.
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/a paused monitor is not flipped to up or rescheduled"
  → test: apps/api/internal/handler/monitors_test.go "TestPauseResumeCronMonitor/pause then resume round-trips status"
- Deleting asks for confirmation, then removes the monitor and its history.
  → test: apps/web/src/views/CronMonitorDetailView.test.ts "deletes the monitor and navigates back to the list on confirm"
  → test: apps/api/internal/handler/monitors_test.go "TestDeleteCronMonitor/cross-tenant delete is a no-op, owner delete succeeds and cascades"
- Another org's monitor can't be read, changed, paused or deleted.
  → test: apps/api/internal/handler/monitors_test.go "TestGetCronMonitor/cross-tenant monitor is not found (tenant isolation)"
  → test: apps/api/internal/handler/monitors_test.go "TestUpdateCronMonitor/cross-tenant update is not found (tenant isolation)"
