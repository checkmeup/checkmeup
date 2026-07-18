---
title: "EP-38: Split status_public.go by responsibility"
type: story
status: planned            # planned | in-progress | shipped | cancelled
updated: 2026-07-18
tags: [architecture, tech-debt]
---

# EP-38: Split status_public.go by responsibility

`apps/api/internal/handler/status_public.go` is the largest handler file
in the codebase at 824 non-comment lines — flagged by the
`architecture-guardrails` skill's 700-line threshold, with the
next-largest handler file at 562. It bundles four distinct
responsibilities in one file: HTTP request handling for the public
status page and badges, SVG badge rendering, status/uptime computation,
and display formatting helpers. `worker.go` hit the same threshold
earlier and was split by monitor type into `worker_cron.go`,
`worker_uptime.go`, `worker_ssl.go`, etc. in the same package — this
epic applies the same pattern here, same package, no behavior change.

---

### US-3801: Extract SVG badge rendering into its own file

**As a** maintainer, **I want** badge SVG generation isolated from HTTP
handling **so that** badge markup changes don't require reading through
unrelated request-handling code.

**Estimate:** 2h

**Acceptance criteria:**

- [ ] `renderBadgeSVG`, `badgeTextWidth`, `writeBadge`, and
      `badgeStatusWord` move to `status_public_badges.go` in the same
      package
- [ ] `ServePageBadge` and `ServeMonitorBadge` (the HTTP handlers) stay in
      `status_public.go` and call into the moved functions unchanged
- [ ] `go build ./...` and existing tests in
      `status_public_test.go` (or equivalent) pass unchanged

---

### US-3802: Extract status/uptime computation into its own file

**As a** maintainer, **I want** the status-page aggregation logic
(overall status, 90-day bar, per-monitor-type row filling) separated
from HTTP handling **so that** the computation can be read and tested
independently of request/response plumbing.

**Estimate:** 3h

**Acceptance criteria:**

- [ ] `computeOverallStatus`, `applySeverity`, `build90DayBar`,
      `buildRow`, and the `fill*Row` family (`fillUptimeRow`,
      `fillCronRow`, `fillSSLRow`, `fillDomainRow`, `fillPortRow`) move
      to `status_public_status.go` in the same package
- [ ] No behavior change — same function signatures, same call sites
- [ ] `go build ./...` and existing tests pass unchanged

---

### US-3803: Extract display formatting helpers into their own file

**As a** maintainer, **I want** the incident/expiry/monitor label and
color formatting helpers grouped separately from HTTP handling **so
that** display-string changes are easy to find without reading through
the handler.

**Estimate:** 2h

**Acceptance criteria:**

- [ ] `relativeTime`, `formatDuration`, `initials`,
      `incidentSeverityLabel`, `incidentSeverityColor`,
      `incidentStatusLabel`, `expiryStatusDisplay`, `solidColorBar`, and
      `monitorStatusDisplay` move to `status_public_format.go` in the
      same package
- [ ] No behavior change — same function signatures, same call sites
- [ ] `go build ./...` and existing tests pass unchanged

---

### US-3804: Verify the split brings status_public.go under threshold

**As a** maintainer, **I want** confirmation the split actually resolves
the guardrail finding **so that** this epic closes the gap it was opened
for, not just relocates code without checking.

**Estimate:** 1h

**Acceptance criteria:**

- [ ] `python3 .claude/skills/architecture-guardrails/audit.py` reports
      no Go handler/worker file over the 700-line threshold
- [ ] `make test` passes (Go + Vue suites, lint)
- [ ] Remaining `status_public.go` contains only HTTP handlers
      (`ServeHTTP`, `ServePageBadge`, `ServeMonitorBadge`,
      `NewStatusPublicHandler`, `loadRows`, `loadActiveIncidents`,
      `loadResolvedIncidents`, `toPublicIncidentRow`,
      `parseIncidentsPageParam`) and request-scoped types
