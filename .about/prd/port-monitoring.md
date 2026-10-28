# Port (TCP) monitoring

## Problem

Mail servers, databases and other non-HTTP services can't be watched by an HTTP check; and a port that should be closed (a database exposed to the internet) is a security risk nobody notices.

## Users

Developers running non-HTTP services, and anyone verifying a port stays closed.

## Success signal

A service that stops accepting connections, or a port that unexpectedly opens, is reported within two check intervals.

## Scope

- A port monitor with a host, a port (1–65535), an interval within the plan minimum and an expected state: open (default) or closed.
- A raw TCP connect each interval, closed straight away; connect time recorded.
- Down and recovery alerts; "unexpectedly open" wording for closed-state monitors.
- List, detail, edit (host and port included), pause and resume, delete.

## Non-goals

- Protocol handshakes or sending data.
- Telling a DNS failure apart from a refused connection for closed-state monitors.

## Acceptance criteria

- A port outside 1–65535 or a missing host is refused.
  → test: apps/api/internal/handler/port_monitors_test.go "TestValidatePort"
  → test: apps/web/src/views/PortMonitorCreateView.test.ts "shows a validation error when port is out of range"
- An interval below the plan minimum is refused.
  → test: apps/api/internal/handler/port_monitors_test.go "TestCreatePortMonitor/an interval below the plan minimum is rejected, not silently clamped"
- With expected state "open", a reachable port is up and an unreachable one fails; with "closed", the reverse.
  → test: apps/api/internal/worker/worker_test.go "TestPerformTCPCheck/expected open + reachable port is up"
  → test: apps/api/internal/worker/worker_test.go "TestPerformTCPCheck/expected open + unreachable port is down"
  → test: apps/api/internal/worker/worker_test.go "TestPerformTCPCheck/expected closed + reachable port is down (unexpectedly open)"
  → test: apps/api/internal/worker/worker_test.go "TestPerformTCPCheck/expected closed + unreachable port is up (matches expectation)"
- Each check is recorded; a failed check is rechecked about 30 seconds later, and only a failed recheck marks the monitor down and alerts (ADR alerting); a passing check recovers it with the downtime.
  → test: apps/api/internal/worker/worker_test.go "TestCheckPortMonitors/a failure that passes its recheck sends nothing and stays up"
  → test: apps/api/internal/worker/worker_test.go "TestCheckPortMonitors/records a check row for every poll"
  → test: apps/api/internal/worker/worker_test.go "TestCheckPortMonitors/escalates to down after two consecutive failures, then recovers"
- A closed-state monitor's alert says the port is unexpectedly open, not that it is down.
  → test: apps/api/internal/worker/worker_test.go "TestCheckPortMonitors/a closed-state monitor alerts when the port unexpectedly opens"
  → test: apps/api/internal/worker/worker_test.go "TestBuildPortDownAlert/closed state reads as unexpectedly open, not down"
- The detail page shows host and port with the expected state, connect-time chart, uptime and incidents.
  → test: apps/web/src/views/PortMonitorDetailView.test.ts "renders the closed expected-state line"
  → test: apps/api/internal/handler/port_monitors_test.go "TestGetPortMonitor/success returns checks/incidents newest first and computed stats"
- Editing changes the name, host, port, interval and expected state; pausing stops checks.
  → test: apps/api/internal/handler/port_monitors_test.go "TestUpdatePortMonitor/success applies changes and defaults, including host/port (unlike SSL/domain)"
  → test: apps/api/internal/handler/port_monitors_test.go "TestPauseResumePortMonitor/pause then resume round-trips status"
- Deleting removes the monitor and its history; another org's monitor can't be read.
  → test: apps/api/internal/handler/port_monitors_test.go "TestDeletePortMonitor/cross-tenant delete is a no-op, owner delete succeeds and cascades"
  → test: apps/api/internal/handler/port_monitors_test.go "TestGetPortMonitor/cross-tenant monitor is not found (tenant isolation)"
