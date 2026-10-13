# Domain expiry monitoring

## Problem

A lapsed domain registration takes down a site, its email and everything else on the domain, and can lose the domain itself; renewal reminders from registrars are easy to miss.

## Users

Developers and freelancers who own or manage domains for themselves or clients.

## Success signal

Every monitored domain is renewed before it expires, with warnings at 30, 14 and 7 days.

## Scope

- A domain monitor for an apex domain, looked up daily through RDAP: registrar, expiry, days left.
- One alert at each of 30, 14 and 7 days before expiry, and at once when expired.
- List, detail, edit, pause and resume, delete; counted in the plan's monitor total.

## Non-goals

- A WHOIS fallback for TLDs without RDAP.
- Alerting on registry statuses such as hold or pending-delete.
- Renewing domains.

## Acceptance criteria

- A full URL entered as the domain is reduced to the domain; a new monitor starts as "waiting".
  → test: apps/api/internal/handler/domain_monitors_test.go "TestCreateDomainMonitor/domain is normalised from a full URL"
  → test: apps/api/internal/handler/domain_monitors_test.go "TestCreateDomainMonitor/success defaults to waiting status with alerts enabled"
- A domain monitor counts toward the plan's shared monitor limit.
  → test: apps/api/internal/handler/domain_monitors_test.go "TestCreateDomainMonitor/counts toward the shared aggregate limit alongside other monitor types"
- A successful lookup records registrar and expiry; the detail page shows them with the days left.
  → test: apps/api/internal/worker/worker_test.go "TestCheckDomainMonitors/a successful lookup records registrar and expiry"
  → test: apps/api/internal/handler/domain_monitors_test.go "TestGetDomainMonitor/success reflects the latest check result"
- A failed lookup (timeout, unsupported TLD, unreadable answer) is recorded as an error, never as expired.
  → test: apps/api/internal/worker/worker_test.go "TestCheckDomainMonitors/a lookup failure is recorded as an error status"
  → test: apps/api/internal/handler/domain_monitors_test.go "TestGetDomainMonitor/an error result is reflected without an expiry"
- Crossing 30, 14 or 7 days left sends one alert per threshold; an expired domain is alerted on at once.
  → test: apps/api/internal/worker/worker_test.go "TestCheckDomainMonitors/crossing the 30-day threshold sends one alert via the fallback email"
  → test: apps/api/internal/worker/worker_test.go "TestDomainExpiredMessages"
  → test: apps/api/internal/worker/worker_test.go "TestDomainExpiringSoonMessages"
- Editing changes the name and domain; pausing stops checks and alerts.
  → test: apps/api/internal/handler/domain_monitors_test.go "TestUpdateDomainMonitor/success applies changes and normalises the domain"
  → test: apps/api/internal/handler/domain_monitors_test.go "TestPauseResumeDomainMonitor/pause then resume round-trips status"
- Deleting removes the monitor and its history; another org's monitor can't be read or changed.
  → test: apps/api/internal/handler/domain_monitors_test.go "TestDeleteDomainMonitor/cross-tenant delete is a no-op, owner delete succeeds"
  → test: apps/api/internal/handler/domain_monitors_test.go "TestGetDomainMonitor/cross-tenant monitor is not found (tenant isolation)"
- Each domain is checked once a day.
