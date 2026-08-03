---
title: "EP-40: Extract shared monitor form components"
type: story
status: planned            # planned | in-progress | shipped | cancelled
updated: 2026-08-03
tags: [architecture, tech-debt, frontend]
---

# EP-40: Extract shared monitor form components

Every monitor type has a `*MonitorCreateView.vue` and a
`*MonitorEditView.vue` that each carry a full copy of the same form
markup. The `architecture-guardrails` skill's sibling-duplication check
flagged seven pairs at 54–87% shared lines; reading them narrowed that
to five genuine ones, where the duplication is the form body itself
rather than incidental page shell:

| Pair | Shared markup lines |
| ---- | ------------------- |
| Uptime | ~250 |
| Cron | ~118 |
| DNS | ~115 |
| Port | ~107 |
| Maintenance window | ~75 |

The file-size thresholds are structurally blind to this — every one of
these files sits well under the 600-line view limit, so the audit
reported clean for months while the same form was maintained in two
places. `UptimeMonitorCreateView.vue` and `UptimeMonitorEditView.vue`
are additionally both in the churn top-10 (6 commits each in the last
150), so the cost is already being paid: every field added and every
label reworded to the uptime form has been done twice.

The fix per pair is one `<Type>MonitorForm.vue` in
`apps/web/src/components/` holding the shared fields, following the
existing `NotificationChannelForm.vue` convention. Each view keeps its
own data loading, submit handler, page title, and cancel destination,
and passes initial values in as props. **Not** a single merged view
behind an `isEdit` flag — that trades duplication for a conditional
maze, and the create/edit split is also the route structure.

Same shape as [EP-38](ep-38-status-public-handler-split.md), which
split `status_public.go` after the same skill's size threshold fired:
pure extraction, no behavior change, verified by the audit going quiet.

**SSL and Domain are deliberately out of scope.** Both scored 54%, but
reading them shows two genuinely different screens: the edit views add
alert settings (`alertsEnabled`, `alertAfterNFailures`,
`maxAlertsPerIncident`), render hostname read-only with a "delete and
recreate to change the domain" note, and add a loading state. Only ~41
markup lines actually coincide, and those are the page shell and the
name field. Extracting a shared form there would abstract over a
difference that is real.

---

### US-4001: Extract the uptime monitor form

**As a** maintainer, **I want** the uptime form defined once **so that**
adding a field or fixing a label doesn't have to be done twice in two
files that both change often.

**Estimate:** 4h

**Acceptance criteria:**

- [ ] `apps/web/src/components/UptimeMonitorForm.vue` holds the shared
      fields: URL, check interval, keyword + keyword mode + case
      sensitivity, JSON assertions, HTTP method, accepted status codes,
      max response time, alert settings, and the
      `NotificationChannelPicker`
- [ ] Both views render it and keep their own data loading, submit
      handler, page title, and cancel destination
- [ ] Create's plan-limit handling (`UpgradePrompt` on
      `plan_limit_reached`) and Edit's load-error handling both still
      work — neither moves into the shared component
- [ ] `minIntervalMins` from `useBilling()` still gates the interval
      options on both views
- [ ] No visual or behavioral change to either screen
- [ ] `npx vue-tsc --noEmit` and the existing Vitest suite pass

---

### US-4002: Extract the cron monitor form

**As a** maintainer, **I want** the cron form defined once **so that**
the schedule, grace period, and max-duration fields have one definition.

**Estimate:** 2h

**Acceptance criteria:**

- [ ] `apps/web/src/components/CronMonitorForm.vue` holds the shared
      fields, including the max run duration field added in
      [EP-34](ep-34-zombie-job-detection.md)
- [ ] Both views render it, keeping their own loading/submit/title/cancel
- [ ] No visual or behavioral change to either screen
- [ ] `npx vue-tsc --noEmit` and the existing Vitest suite pass

---

### US-4003: Extract the port and DNS monitor forms

**As a** maintainer, **I want** the port and DNS forms defined once
**so that** the two most recently added monitor types don't accumulate
the same drift as the older ones.

**Estimate:** 3h

**Acceptance criteria:**

- [ ] `apps/web/src/components/PortMonitorForm.vue` and
      `DNSMonitorForm.vue` hold their respective shared fields
- [ ] DNS's record type, expected value, and baseline-mode fields
      ([EP-39](ep-39-dns-monitoring.md)) live in the shared component
- [ ] Port's host/port fields are editable on create and read-only on
      edit if that is the current behavior — verify against the views
      before assuming, and keep whichever behavior ships today
- [ ] No visual or behavioral change to any of the four screens
- [ ] `npx vue-tsc --noEmit` and the existing Vitest suite pass

---

### US-4004: Extract the maintenance window form

**As a** maintainer, **I want** the maintenance window form defined once
**so that** the schedule and monitor-picker fields have one definition.

**Estimate:** 2h

**Acceptance criteria:**

- [ ] `apps/web/src/components/MaintenanceWindowForm.vue` holds the
      shared fields, including the `MaintenanceMonitorPicker`
- [ ] Both views render it, keeping their own loading/submit/title/cancel
- [ ] No visual or behavioral change to either screen
- [ ] `npx vue-tsc --noEmit` and the existing Vitest suite pass

---

### US-4005: Verify the extraction closes the guardrail finding

**As a** maintainer, **I want** confirmation the extraction actually
resolves what the audit flagged **so that** this epic closes the gap it
was opened for rather than moving markup around without checking.

**Estimate:** 1h

**Acceptance criteria:**

- [ ] `python3 .claude/skills/architecture-guardrails/audit.py` reports
      no sibling pair over the 50% threshold except SSL and Domain
- [ ] SSL and Domain are added to `KNOWN_EXCEPTIONS` in `audit.py` with
      the rationale above, so the audit stops re-reporting a pair that
      has been read and deliberately left alone
- [ ] No new finding appears in the Vue component size check — a shared
      form larger than the 250-line component threshold needs splitting
      further, not an exception
- [ ] `make test` passes (Go + Vue suites, lint)
- [ ] Each of the five extracted forms is rendered by exactly two views
