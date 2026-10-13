# Assertions on uptime checks

## Problem

An API can answer `200` and still be unhealthy (`{"status":"degraded"}`) or too slow to use; a keyword search is too blunt to catch that reliably.

## Users

Uptime-monitor users checking JSON APIs.

## Success signal

A JSON health endpoint reporting itself unhealthy is marked down, with the failed assertion as the reason.

## Scope

- JSON assertions on an uptime monitor: a path (`$.status`, `data.healthy`), a comparator (equals, not equals, contains, greater than, less than) and an expected value; all must pass.
- Evaluation within the same request and the first 512 KB of the body (PRD keyword-monitoring).
- A maximum response time, past which the check fails (PRD configurable-uptime-checks).
- The configured assertions and the failure reason shown on the monitor.

## Non-goals

- Multi-step (chained) checks that pass values between requests: agreed but not built; it needs its own PRD.
- Storing or showing the response body.

## Acceptance criteria

- An assertion compares the value at its JSON path with each comparator, including nested paths, booleans, null and numbers.
  → test: apps/api/internal/worker/worker_test.go "TestEvaluateJsonAssertion"
- A missing path, or a path through a non-object, fails with a "not found" reason.
  → test: apps/api/internal/worker/worker_test.go "TestEvaluateJsonAssertion"
- A body that isn't valid JSON fails the check with its own reason rather than passing.
  → test: apps/api/internal/worker/worker_test.go "TestEvaluateJsonAssertion"
- A numeric comparison against a non-numeric value fails rather than passing.
  → test: apps/api/internal/worker/worker_test.go "TestEvaluateJsonAssertion"
- A failing assertion marks the check down with the assertion failure as the reason; a passing one leaves it up.
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/a failing JSON assertion is down with the assertion failure as the reason"
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/a passing JSON assertion is up"
- A response slower than the monitor's maximum response time fails the check.
  → test: apps/api/internal/worker/worker_test.go "TestPerformHTTPCheck/a response slower than the configured timeout is down"
- A maximum response time outside 1,000–30,000 ms is refused.
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestCreateUptimeMonitor/validation"
- With several assertions, the check fails when any one fails.
- The detail page lists the configured assertions and the response-time limit.
