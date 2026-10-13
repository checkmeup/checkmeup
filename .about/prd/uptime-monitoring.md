# Uptime monitoring: alert when a URL stops answering

## Problem

A site or API that goes down is usually reported by its users first; the owner wants to know before they do.

## Users

Developers and freelancers watching their own or their clients' sites and APIs.

## Success signal

An outage is reported within two check intervals, and a single transient error doesn't page anyone.

## Scope

- An uptime monitor with a name, an `http(s)` URL and a check interval no shorter than the plan allows (ADR billing).
- Periodic checks recording status code and response time; down after two consecutive failed checks; down and recovery alerts.
- List, detail (response-time chart, uptime %, incidents, check log), edit, pause and resume, delete.

## Non-goals

- Checks from several regions (PRD multi-region-checks).
- Browser-level (synthetic) checks.

## Acceptance criteria

- A URL that isn't `http://` or `https://` is refused.
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestValidateURL"
- An interval below the plan's minimum is refused, not silently raised; the minimum itself is accepted.
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestCreateUptimeMonitor/an interval below the plan minimum is rejected, not silently clamped"
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestCreateUptimeMonitor/success at the plan minimum interval, with defaults applied"
- Each check is recorded with its status code and response time.
  → test: apps/api/internal/worker/worker_test.go "TestCheckUptimeMonitors/records a check row for every poll"
- With default settings a `200` is up; another status code, a connection error or a timeout is a failed check (ADR workers).
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/200 with no keyword is up"
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/non-200 is down with the status code in the reason"
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/connection error is down with no status code"
- Two consecutive failed checks mark the monitor down and send an alert; the next passing check marks it up and sends a recovery alert with the downtime.
  → test: apps/api/internal/worker/worker_test.go "TestCheckUptimeMonitors/escalates to down after two consecutive failures, then recovers"
  → test: apps/api/internal/worker/worker_test.go "TestCheckUptimeMonitors/delivers a recovery webhook event with a non-zero downtime duration"
- The list shows each monitor's status and uptime, and an empty state with a create button.
  → test: apps/web/src/views/UptimeMonitorListView.test.ts "renders monitor rows with status, uptime, and keyword label"
  → test: apps/web/src/views/UptimeMonitorListView.test.ts "shows an empty state when there are no monitors"
- The detail page shows a 24-hour response-time chart, uptime % for 24 hours, 7 and 30 days, incidents with duration, and a paginated check log newest first.
  → test: apps/web/src/views/UptimeMonitorDetailView.test.ts "renders the response time chart when there are at least two chart points"
  → test: apps/web/src/views/UptimeMonitorDetailView.test.ts "renders uptime stats, formatting a missing 30d value as a dash"
  → test: apps/web/src/views/UptimeMonitorDetailView.test.ts "renders incidents with resolved duration and an ongoing badge"
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestGetUptimeMonitor/success returns checks/incidents newest first and computed stats"
- Editing changes the name, URL and interval.
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestUpdateUptimeMonitor/success applies changes and defaults"
- Pausing stops checks and alerts; resuming starts them again.
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestPauseResumeUptimeMonitor/pause then resume round-trips status"
- Deleting asks for confirmation, then removes the monitor and its history.
  → test: apps/web/src/views/UptimeMonitorDetailView.test.ts "deletes the monitor and navigates back to the list on confirm"
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestDeleteUptimeMonitor/cross-tenant delete is a no-op, owner delete succeeds and cascades"
- Another org's monitor can't be read or changed.
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestGetUptimeMonitor/cross-tenant monitor is not found (tenant isolation)"
