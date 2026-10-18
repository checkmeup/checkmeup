# Configurable uptime checks

## Problem

Health endpoints differ: some answer `204` or a redirect when healthy, some need `HEAD` or `POST`, some are slow by design; a fixed `GET`, `200`-only, 10-second check raised false alarms for them.

## Users

Uptime-monitor users whose endpoints don't fit the defaults.

## Success signal

A healthy endpoint answering `204` (or needing a 20-second timeout) stays green once configured, and monitors left on defaults behave exactly as before.

## Scope

- Per monitor: HTTP method (`GET`, `HEAD`, `POST`; default `GET`), maximum response time (1–30 s; default 10 s) used as the request timeout, and accepted status codes (default `200`).
- The fields in a collapsed "advanced" section of the form.
- Keyword and JSON checks still run on top (PRD keyword-monitoring, PRD assertion-checks).

## Non-goals

- Custom request headers or bodies.
- Methods other than `GET`, `HEAD` and `POST`.

## Acceptance criteria

- The check uses the configured method.
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/HEAD method is issued when configured"
- A status code in the accepted set is up, even if it isn't `200`.
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/a non-default accepted status code is up"
- A response slower than the maximum response time is a failed check.
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/a response slower than the configured timeout is down"
- A method other than `GET`, `HEAD` or `POST`, a maximum response time outside 1,000–30,000 ms, or a status code outside 100–599 is refused by the API.
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestCreateUptimeMonitor/validation"
- An empty set of accepted status codes can't be saved.
  → test: apps/web/src/lib/uptimeMonitorForm.test.ts "rejects an empty accepted-status-code set"
- A monitor created without these fields gets `GET`, 10 seconds and `200`.
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestCreateUptimeMonitor/success at the plan minimum interval, with defaults applied"
- The accepted status codes are checked against the final response after redirects.
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/accepted status codes are checked against the final response after redirects"
- The three fields sit in a collapsed "advanced" section of the create and edit forms.
  → test: apps/web/src/views/UptimeMonitorCreateView.test.ts "keeps the method, max response time and accepted status codes in a collapsed advanced section"
  → test: apps/web/src/views/UptimeMonitorEditView.test.ts "keeps the method, max response time and accepted status codes in a collapsed advanced section"
