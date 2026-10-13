# Keyword checks on uptime monitors

## Problem

A `200` doesn't prove a page works: a maintenance page, an error inside a JSON body or a half-broken page all answer `200`.

## Users

Uptime-monitor users on every plan.

## Success signal

A page that answers `200` but lost its expected content is reported as down.

## Scope

- An optional keyword on an uptime monitor, with "contains" or "does not contain" and a case-sensitivity switch.
- A plain substring search over the first 512 KB of the response, in the same request as the status check.
- Keyword failures shown and alerted on with their own reason.

## Non-goals

- Regular expressions.
- Storing or showing the response body.
- JSON-path assertions (PRD assertion-checks).

## Acceptance criteria

- A monitor without a keyword behaves exactly as a plain uptime check.
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/200 with no keyword is up"
- In "contains" mode the check fails when the keyword is missing; in "does not contain" mode when it is present.
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/keyword absent (contains mode) is down"
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/keyword present (not_contains mode) is down"
- Matching is case-insensitive unless the switch is on.
  → test: apps/api/internal/worker/worker_test.go "TestKeywordCheckPasses"
- A non-`200` response fails on its status before the keyword is checked.
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/non-200 short-circuits before the keyword check"
- A keyword failure's reason ("Keyword not found") differs from an HTTP failure's ("HTTP 500").
  → test: apps/api/internal/worker/worker_test.go "TestKeywordFailureReason"
- A keyword must be 1–500 characters, checked in the form and by the API.
  → test: apps/web/src/lib/uptimeMonitorForm.test.ts "rejects a keyword of 501 characters"
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestValidateKeyword"
- Keywords work on the Hobby plan.
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestCreateUptimeMonitor/keyword monitoring works on the Hobby (free) plan"
- The list and detail page show the keyword and its mode when one is set.
  → test: apps/web/src/views/UptimeMonitorListView.test.ts "renders monitor rows with status, uptime, and keyword label"
  → test: apps/web/src/views/UptimeMonitorDetailView.test.ts "renders the keyword line only when a keyword is configured"
- Only the first 512 KB of the body is searched, whatever its `Content-Length`.
