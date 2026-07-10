---
title: Tech Debt
type: reference
status: active
updated: 2026-07-10
tags: [architecture, maintainability, backend]
---

# Tech debt

Known architecture/code smells that aren't worth an ADR or an immediate fix, but shouldn't be forgotten. Add an entry when you spot something during other work rather than stopping to fix it; remove an entry once it's addressed (reference the commit/PR in the removal, not here — `git log` is the record of what was fixed and when).

---

## Backend (Go)

### Maintainability

- **Duplicated alert-message string building** across `checkOverdue` (`worker_cron.go`), `checkOneUptimeMonitor` (`worker_uptime.go`), `checkOneSSLMonitor` (`worker_ssl.go`) — near-identical `fmt.Sprintf` pairs (Telegram/email subject/HTML) repeated per monitor type. The SSL threshold-alert block (`worker_ssl.go`'s `sslExpiredMessages`/`sslExpiringSoonMessages`) is a near-duplicate of the domain one (`worker_domain.go`'s `domainExpiredMessages`/`domainExpiringSoonMessages`), differing only by field names (issuer vs registrar).
  → Factor into a small templated helper shared across monitor types.
