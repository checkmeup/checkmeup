# EP-03: Uptime monitor

An uptime monitor polls a URL on a fixed interval. Two consecutive failures change the status to "down" and trigger an alert. Recovery triggers a second alert.

---

### US-0301: Create an uptime monitor

**As a** user, **I want** to add a URL to uptime monitoring **so that** I know immediately when it goes down.

**Acceptance criteria:**

- [ ] Fields: name, URL, check interval
- [ ] URL validated — must be `http://` or `https://`
- [ ] Interval options: 1, 3, 5, 10, 30 min (Hobbyist: 5 min minimum — enforced by plan limit)
- [ ] First check runs within one interval of creation

---

### US-0302: Perform HTTP health check

**As a** user, **I want** the platform to check my URL automatically **so that** I don't have to.

**Acceptance criteria:**

- [ ] HEAD request sent first; falls back to GET if HEAD returns 405
- [ ] 10-second request timeout
- [ ] 2xx response = up; non-2xx or timeout = failed check
- [ ] Redirects followed up to 5 hops (redirect chain itself is not considered a failure)
- [ ] Response time recorded on every check

---

### US-0303: Detect downtime and alert

**As a** user, **I want** to be alerted when my service goes down **so that** I can respond quickly.

**Acceptance criteria:**

- [ ] 2 consecutive failed checks before status changes to "down" (avoids flapping on transient errors)
- [ ] Alert sent on transition to "down" (see EP-05)
- [ ] Alert sent on transition back to "up" (recovery)
- [ ] No repeat alerts while status stays "down"

---

### US-0304: View uptime monitor list

**As a** user, **I want** to see all my uptime monitors and their status **so that** I can spot problems at a glance.

**Acceptance criteria:**

- [ ] Shows: name, URL, status, uptime % (last 24h), last checked time
- [ ] Status badges consistent with cron monitors
- [ ] Empty state with prompt to create first monitor

---

### US-0305: View uptime monitor detail

**As a** user, **I want** to see response time history and the incident log **so that** I can understand patterns and diagnose outages.

**Acceptance criteria:**

- [ ] Response time chart for the last 24 hours
- [ ] Uptime % for last 24h / 7d / 30d
- [ ] Incident log: started at, resolved at, duration — paginated, latest first
- [ ] Check log: timestamp, status code, response time — paginated

---

### US-0306: Edit, pause, and delete an uptime monitor

**As a** user, **I want** to manage an uptime monitor's settings **so that** it stays accurate and doesn't generate noise during maintenance.

**Acceptance criteria:**

- [ ] Editable: name, URL, check interval
- [ ] Pause stops checks and suppresses alerts; status shown as "paused"
- [ ] Resume restarts checks immediately
- [ ] Delete requires confirmation; all history deleted
