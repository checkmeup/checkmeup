# Domain expiry monitoring

## Problem

A lapsed domain registration takes down a site, its email and everything else on the domain, and can lose the domain itself; renewal reminders from registrars are easy to miss.

## Users

Developers and freelancers who own or manage domains for themselves or clients.

## Success signal

Every monitored domain is renewed before it expires, with warnings at 30, 14 and 7 days.

## Scope

- A domain monitor for an apex domain, looked up daily through RDAP, or WHOIS for the TLDs whose registries have no RDAP but publish an expiry date there (.il, .io, .co, .me, .us, .it, .ru, .jp, .cn, .se): registrar, expiry, days left. A TLD whose registry publishes no expiry date (.de, .eu, .ch, .li, .at, .es) is refused with the reason ([ADR domain-lookups](../adr/domain-lookups.md)).
- One alert at each of 30, 14 and 7 days before expiry, and at once when expired.
- List, detail, edit, pause and resume, delete; counted in the plan's monitor total.

## Non-goals

- WHOIS for TLDs beyond the listed ones: each registry's answer has its own format and needs its own parser, written against its real answers.
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
- A domain under a TLD without RDAP is looked up over WHOIS, with its expiry and registrar read from that registry's real answer format; a domain the registry doesn't know is an error, not "expired".
  → test: apps/api/internal/whois/whois_test.go "TestParse_RegisteredDomains"
  → test: apps/api/internal/whois/whois_test.go "TestParse_UnregisteredDomains"
  → test: apps/api/internal/worker/worker_test.go "TestCheckDomainMonitors/a TLD without RDAP is looked up over WHOIS (ADR domain-lookups)"
  → test: apps/api/internal/worker/worker_test.go "TestCheckDomainMonitors/a domain the WHOIS registry doesn't know is an error, not expired"
- A domain under a TLD whose registry publishes no expiry date can't be added, and the form says why; an existing one shows that reason as its error.
  → test: apps/api/internal/handler/domain_monitors_test.go "TestCreateDomainMonitor/a TLD without RDAP but with WHOIS is accepted, and a refused one says why"
  → test: apps/api/internal/worker/worker_test.go "TestCheckDomainMonitors/a TLD whose registry publishes no expiry says so instead of asking anyone"
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
