# EP-09: Maintenance windows

Suppress alerts and checks for one or more monitors during planned (or unplanned) downtime, without affecting uptime stats. Promised on the landing page and blog ahead of implementation; built post-MVP. See [ADR-020](../decisions/020-maintenance-windows.md).

---

### US-0901: Schedule a maintenance window

**As a** user, **I want** to schedule a maintenance window with a start and end time **so that** I don't get paged for planned downtime.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Fields: title, optional message, start time, end time
- [x] End time can be left unset ("no end date") for unplanned/ongoing maintenance — closed manually later
- [x] End time, if set, must be after start time

---

### US-0902: Cover multiple monitors in one window

**As a** user, **I want** to select any combination of cron, uptime, and SSL monitors for a single window **so that** one deploy or maintenance task covers everything it affects.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Multi-select across all monitor types, reusing the same picker pattern as status pages
- [x] At least one monitor required per window
- [x] Each selected monitor validated as belonging to the org

---

### US-0903: Suppress checks, incidents, and alerts during the window

**As a** user, **I want** covered monitors to be left alone while a window is active **so that** no alerts fire and my uptime stats aren't penalized for planned downtime.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [x] Monitors under an active window are excluded from the worker's due/overdue queries — no check runs, no incident is created, no alert is sent
- [x] 90-day uptime bar and uptime % are unaffected, since no incident is ever recorded for the window's duration
- [x] Normal checking resumes automatically once the window ends

---

### US-0904: View "Under maintenance" on the public status page

**As a** visitor, **I want** to see when a service is under planned maintenance **so that** I don't mistake it for an outage.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Covered monitors show an "Under maintenance" chip instead of up/down/paused
- [x] Optional message from the window is shown alongside the chip
- [x] Does not trigger the page's "Partial outage" / "Major outage" overall banner

---

### US-0905: End or delete a maintenance window

**As a** user, **I want** to end a window early or remove it entirely **so that** I can react when planned work finishes ahead of schedule or didn't need to happen.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] "End now" sets the end time to the current moment for any window that hasn't already ended
- [x] Delete removes the window (and its monitor associations) entirely
- [x] List view distinguishes upcoming / active / ended windows
