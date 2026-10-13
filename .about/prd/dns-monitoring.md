# DNS record monitoring

## Problem

A DNS record changed by hijacking, an expired delegation or a botched migration can send traffic or email elsewhere while an uptime check stays green, because it never looks at the record itself.

## Users

Developers and freelancers responsible for their own or their clients' DNS.

## Success signal

A changed or unresolvable record is reported within one check interval, saying what it changed from and to.

## Scope

- A DNS monitor with a hostname, a record type (A, AAAA, CNAME, MX, TXT, NS), an interval within the plan minimum, and an optional expected value.
- Without an expected value, the first successful lookup becomes the baseline; any later difference is a change until the user edits the monitor.
- Lookup errors (NXDOMAIN, SERVFAIL, timeout) recorded apart from changes.
- Down and recovery alerts; list, detail, edit, pause and resume, delete.

## Non-goals

- Checking from several resolvers or regions.
- DNSSEC validation.

## Acceptance criteria

- An unknown record type is refused; an interval below the plan minimum is refused.
  → test: apps/api/internal/handler/dns_monitors_test.go "TestValidateRecordType"
  → test: apps/api/internal/handler/dns_monitors_test.go "TestCreateDNSMonitor/an interval below the plan minimum is rejected, not silently clamped"
- A monitor without an expected value starts in baseline mode; one with an expected value is pinned to it.
  → test: apps/api/internal/handler/dns_monitors_test.go "TestCreateDNSMonitor/no expected value: baseline mode, with defaults applied"
  → test: apps/api/internal/handler/dns_monitors_test.go "TestCreateDNSMonitor/an explicit expected value is pinned"
- In baseline mode the first successful lookup is stored as the baseline, and a later different answer marks the monitor down as a change.
  → test: apps/api/internal/worker/worker_dns_test.go "TestCheckDNSMonitors/baseline mode captures the first successful lookup as the comparison value"
  → test: apps/api/internal/worker/worker_dns_test.go "TestCheckDNSMonitors/a later check that differs from the captured baseline goes down as a mismatch, not a lookup error"
- A pinned monitor whose lookup matches is up, and recovers an open incident.
  → test: apps/api/internal/worker/worker_dns_test.go "TestCheckDNSMonitors/pinned expected value: matching lookup is up, recovers an existing incident"
- Several values are compared sorted, so the same records in another order are not a change.
  → test: apps/api/internal/worker/worker_dns_test.go "TestJoinSortedValues"
- NXDOMAIN, SERVFAIL or a timeout is recorded as a lookup error, never as a change.
  → test: apps/api/internal/worker/worker_dns_test.go "TestCheckDNSMonitors/NXDOMAIN is recorded as a lookup error, distinct from a mismatch"
  → test: apps/api/internal/worker/worker_dns_test.go "TestClassifyLookupError"
- A change alert shows the old and the new value; an error alert says the record is unreachable; record text is escaped in alerts.
  → test: apps/api/internal/worker/worker_dns_test.go "TestBuildDNSDownAlert/mismatch reads as a change, old value to new value"
  → test: apps/api/internal/worker/worker_dns_test.go "TestBuildDNSDownAlert/lookup error reads as unreachable, not a change"
  → test: apps/api/internal/worker/worker_dns_test.go "TestBuildDNSDownAlert/HTML in a DNS response is escaped in Telegram/email, not injected raw"
- Each lookup is recorded.
  → test: apps/api/internal/worker/worker_dns_test.go "TestCheckDNSMonitors/records a check row for every poll"
- The detail page labels the comparison value "Baseline" or "Expected value" by mode.
  → test: apps/web/src/views/DNSMonitorDetailView.test.ts "labels an auto-captured value as Baseline instead of Expected value"
- Editing can change hostname and record type; clearing the expected value re-arms baseline mode.
  → test: apps/api/internal/handler/dns_monitors_test.go "TestUpdateDNSMonitor/success applies changes and defaults, including hostname/record type"
  → test: apps/api/internal/handler/dns_monitors_test.go "TestUpdateDNSMonitor/clearing the expected value re-arms baseline mode"
- Pausing stops checks; deleting removes the monitor and its history.
  → test: apps/api/internal/handler/dns_monitors_test.go "TestPauseResumeDNSMonitor/pause then resume round-trips status"
  → test: apps/api/internal/handler/dns_monitors_test.go "TestDeleteDNSMonitor/cross-tenant delete is a no-op, owner delete succeeds and cascades"
- The detail page has a change log of when the value changed, from what to what.
