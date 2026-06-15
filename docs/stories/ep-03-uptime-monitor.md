# EP-03: Uptime monitor

An uptime monitor polls a URL on a fixed interval. Two consecutive failures change the status to "down" and trigger an alert. Recovery triggers a second alert.

---

### US-0301: Create an uptime monitor

**As a** user, **I want** to add a URL to uptime monitoring **so that** I know immediately when it goes down.

**Acceptance criteria:**

- [x] Fields: name, URL, check interval
- [x] URL validated — must be `http://` or `https://`
- [x] Interval options: 10 min, 30 min (MVP minimum — see ADR-014)
- [x] First check runs within one interval of creation

---

### US-0302: Perform HTTP health check

**As a** user, **I want** the platform to check my URL automatically **so that** I don't have to.

**Acceptance criteria:**

- [x] GET request always (see ADR-014 — HEAD eliminated for simplicity)
- [x] 10-second request timeout
- [x] HTTP 200 = up; any other status code or timeout = failed check (see ADR-014)
- [x] Redirects followed automatically
- [x] Response time recorded on every check

---

### US-0303: Detect downtime and alert

**As a** user, **I want** to be alerted when my service goes down **so that** I can respond quickly.

**Acceptance criteria:**

- [x] 2 consecutive failed checks before status changes to "down" (avoids flapping on transient errors)
- [x] Alert sent on transition to "down" (see EP-05)
- [x] Alert sent on transition back to "up" (recovery)
- [x] Alert cap via `max_alerts_per_incident` (0=always, default 3 — see ADR-016)

---

### US-0304: View uptime monitor list

**As a** user, **I want** to see all my uptime monitors and their status **so that** I can spot problems at a glance.

**Acceptance criteria:**

- [x] Shows: name, URL, status, uptime % (last 24h), last checked time
- [x] Status badges consistent with cron monitors
- [x] Empty state with prompt to create first monitor

---

### US-0305: View uptime monitor detail

**As a** user, **I want** to see response time history and the incident log **so that** I can understand patterns and diagnose outages.

**Acceptance criteria:**

- [x] Response time chart for the last 24 hours
- [x] Uptime % for last 24h / 7d / 30d
- [x] Incident log: started at, resolved at, duration — paginated, latest first
- [x] Check log: timestamp, status code, response time — paginated

---

### US-0306: Edit, pause, and delete an uptime monitor

**As a** user, **I want** to manage an uptime monitor's settings **so that** it stays accurate and doesn't generate noise during maintenance.

**Acceptance criteria:**

- [x] Editable: name, URL, check interval
- [x] Pause stops checks and suppresses alerts; status shown as "paused"
- [x] Resume restarts checks immediately
- [x] Delete requires confirmation; all history deleted
