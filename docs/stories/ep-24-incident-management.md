# EP-24: Incident management

Today, `cron_incidents` and `uptime_incidents` are created and resolved entirely automatically by the worker from up/down transitions ([ADR-016](../decisions/016-alert-debounce.md)) — there's no way for a user to narrate what's happening, post progress updates, or declare something affecting visitors that isn't a hard monitor-down (e.g. degraded performance). This epic adds manually-managed incidents, shown on the public status page ([EP-06](ep-06-status-page.md)) alongside the existing automatic up/down state.

**Needs a schema decision before implementation** (add to [decision backlog](../decisions/backlog.md)): should manual incidents live in a new `status_page_incidents` table (decoupled from monitors and from the existing per-monitor-type `cron_incidents`/`uptime_incidents`), or extend the existing incident tables? A new, monitor-type-agnostic table is the more likely fit, since a manual incident can span multiple monitors of different types and isn't tied to a single check transition the way the automatic ones are.

---

### US-2401: Manually declare an incident

**As a** user, **I want** to manually declare an incident affecting one or more monitors **so that** my status page reflects what's actually happening, not just automatic up/down state.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] Fields: title, initial message, affected monitors (multi-select, any type), severity (Minor / Major / Critical)
- [ ] Initial status: Investigating
- [ ] Independent of automatic monitor status — declaring an incident doesn't change a monitor's own up/down state
- [ ] Visible immediately on the public status page (US-2403)

---

### US-2402: Post incident updates

**As a** user, **I want** to post timestamped updates to an open incident **so that** visitors see progress without me editing the original message.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Updates are append-only entries (timestamp + message), shown reverse-chronological under the incident
- [ ] Status progresses with each update: Investigating → Identified → Monitoring → Resolved
- [ ] Marking "Resolved" sets the incident's end time and removes it from the public page's active list

---

### US-2403: View incidents on the public status page

**As a** visitor, **I want** to see active and recent incidents on the status page **so that** I understand what's currently affecting the service.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] Active (unresolved) incidents shown above the monitor list, with current status and most recent update
- [ ] Resolved incidents listed in a paginated history, separate from the existing per-monitor 90-day uptime bar ([US-0603](ep-06-status-page.md))
- [ ] An active Major/Critical incident escalates the existing overall status banner ("Partial outage" / "Major outage"), the same banner monitor-down already drives today

---

### US-2404: Edit or delete an incident

**As a** user, **I want** to fix a mistake in an incident's title or an update **so that** the public record stays accurate.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Title and any update's message editable after creation
- [ ] Deleting an incident removes it (and its updates) from the public page entirely — for genuine mistakes, not a way to hide real incidents
- [ ] Edits don't trigger alerts on any channel — declaring/updating an incident is a status-page concept for MVP, not a new alert source

---

### US-2405: Don't conflict with maintenance windows

**As a** user, **I want** incidents and maintenance windows to not contradict each other on the status page **so that** visitors aren't shown conflicting information.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Declaring an incident on a monitor already under an active maintenance window ([EP-09](ep-09-maintenance-windows.md)) shows a warning before confirming — both can coexist but it's an explicit choice, not an accident
- [ ] When both apply to the same monitor, the status page clearly distinguishes "planned maintenance" from "unplanned incident" rather than showing one generic state
