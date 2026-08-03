---
title: "EP-40: Extract shared monitor form components"
type: story
status: in-progress        # planned | in-progress | shipped | cancelled
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

**SSL and Domain are out of scope for form extraction.** Both scored
54%, but reading them shows two genuinely different screens: the edit
views add alert settings (`alertsEnabled`, `alertAfterNFailures`,
`maxAlertsPerIncident`), render hostname read-only with a "delete and
recreate to change the domain" note, and add a loading state. Only ~41
markup lines actually coincide, and those are the page shell and the
name field. Extracting a shared form there would abstract over a
difference that is real. They do still get the shared page shell
(US-4002), which applies to all seven pairs.

**Scope correction, added after US-4001.** Extracting the uptime form
took that pair from 87% to 72% — still flagging, because what remained
shared was no longer form fields but the *page shell*: `AppLayout`, back
link, title, `<form>` element, error row, and button row. That shell is
near-identical across every monitor type (each differs from uptime's by
4 lines: route name and title), so every form story would have landed at
~72% and US-4006 could never have passed. US-4002 extracts the shell,
and is sequenced before the remaining form stories so each of those is
done against an already-thin view instead of touching all 14 views
twice.

---

### US-4001: Extract the uptime monitor form

**As a** maintainer, **I want** the uptime form defined once **so that**
adding a field or fixing a label doesn't have to be done twice in two
files that both change often.

**Estimate:** 4h

**Acceptance criteria:**

- [x] `apps/web/src/components/UptimeMonitorForm.vue` holds the shared
      fields: URL, check interval, keyword + keyword mode + case
      sensitivity, JSON assertions, HTTP method, accepted status codes,
      max response time, alert settings, and the
      `NotificationChannelPicker`
- [x] Both views render it and keep their own data loading, submit
      handler, page title, and cancel destination
- [x] Create's plan-limit handling (`UpgradePrompt` on
      `plan_limit_reached`) and Edit's load-error handling both still
      work — neither moves into the shared component
- [x] `minIntervalMins` from `useBilling()` still gates the interval
      options on both views
- [x] No visual or behavioral change to either screen — with one
      deliberate exception, below
- [x] `npx vue-tsc --noEmit` and the existing Vitest suite pass

**Implementation note.** The shared form came to ~290 lines, over the
250-line component threshold `architecture-guardrails` enforces, so it
ships as three components rather than one: `UptimeMonitorForm.vue`
(154), `UptimeAdvancedSettings.vue` (119), and
`UptimeJsonAssertionsField.vue` (56). Static helpers, the form-state
type, and validation live in `lib/uptimeMonitorForm.ts` (147), matching
the existing `lib/notificationChannelTypes.ts` /
`useNotificationChannelForm.ts` split. The two views went from 439 + 488
= 927 lines to 97 + 122 = 219.

**One deliberate behavior change.** The two copies of
`validateUptimeMonitorForm` had already drifted: create returned "URL is
required" then "URL must start with http:// or https://" as separate
messages, while edit collapsed both into the second. The shared version
uses create's more specific pair, so submitting the edit form with an
empty URL now says "URL is required" instead of "URL must start with
http://". This is the duplication's own cost showing up — the fix picks
the better message rather than preserving the drift.

---

### US-4002: Extract the shared monitor form page shell

**As a** maintainer, **I want** the page chrome every monitor
create/edit screen shares defined once **so that** the remaining
duplication in each pair is actually about that monitor's fields.

**Estimate:** 3h

**Acceptance criteria:**

- [x] `apps/web/src/components/MonitorFormPage.vue` owns `AppLayout`,
      the back link, the title, the `<form>` element, the loading state,
      the error / `UpgradePrompt` row, and the submit/cancel buttons
- [x] All 14 create/edit views render it, passing title, back route,
      button labels, and their own error/submitting/loading state
- [x] The component owns no form state and never branches on create vs
      edit — the only mode-shaped input is `loading`, which edit views
      set because they fetch and create views don't
- [x] `MaintenanceWindowEditView`'s delete-with-confirm survives, via an
      `actions` slot rather than a prop
- [x] No visual change: with the `actions` slot empty, `justify-between`
      leaves the single button group left-aligned as before
- [x] `npx vue-tsc --noEmit` and the existing Vitest suite pass

**Implementation note.** `MonitorFormPage.vue` is 82 lines and takes 8
props, which is the upper end of what a layout component should carry —
the deletion test is what justified it: removing the component would
scatter ~45 lines of identical chrome back across 14 views, so it
concentrates complexity rather than moving it. If it ever needs a
per-monitor-type conditional, that's the signal to stop rather than add
a ninth prop.

The delete-with-confirm block was nearly lost: the first generated pass
dropped it, because the extraction assumed every view's footer was just
save + cancel. A content check comparing every non-blank line of each
original against its rewrite caught it before the files were applied.

**Result:** SSL and Domain dropped off the audit's finding list
entirely, falling under the 50% threshold on their own — so US-4006 no
longer needs a `KNOWN_EXCEPTIONS` entry for them. Uptime went from 72%
across 199 lines to 67% across 152.

---

### US-4003: Extract the cron monitor form

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

### US-4004: Extract the port and DNS monitor forms

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

### US-4005: Extract the maintenance window form

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

### US-4006: Verify the extraction closes the guardrail finding

**As a** maintainer, **I want** confirmation the extraction actually
resolves what the audit flagged **so that** this epic closes the gap it
was opened for rather than moving markup around without checking.

**Estimate:** 1h

**Acceptance criteria:**

- [ ] `python3 .claude/skills/architecture-guardrails/audit.py` reports
      no sibling pair over the 50% threshold
- [ ] `KNOWN_EXCEPTIONS` stays empty. US-4002 dropped SSL and Domain
      under the threshold on their own, so the exception this story
      originally planned for them is no longer needed — prefer no
      suppression over a justified one
- [ ] No new finding appears in the Vue component size check — a shared
      form larger than the 250-line component threshold needs splitting
      further, not an exception
- [ ] `make test` passes (Go + Vue suites, lint)
- [ ] Each of the five extracted forms is rendered by exactly two views
