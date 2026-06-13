# EP-02: Cron monitor

A cron monitor expects a periodic HTTP ping from a scheduled job. If no ping arrives within the grace period after the expected time, the monitor goes down and an alert fires.

---

### US-0201: Create a cron monitor

**As a** user, **I want** to create a cron monitor **so that** I can verify my scheduled jobs are running.

**Acceptance criteria:**

- [ ] Fields: name, schedule (cron expression or plain interval — e.g. "every 1h"), grace period
- [ ] A unique ping URL with an unguessable token is generated on creation
- [ ] Ping URL displayed prominently after creation with a copy button
- [ ] Monitor starts in "waiting for first ping" state (not "down")

---

### US-0202: Receive a ping

**As a** user, **I want** my scheduled job to call a URL to confirm it ran **so that** the monitor knows it's healthy.

**Acceptance criteria:**

- [ ] Endpoint: `GET /ping/{token}` — no auth header required; the token is the secret
- [ ] Any response from the endpoint is 200 OK (the job shouldn't fail because monitoring is down)
- [ ] Ping logged with timestamp and source IP
- [ ] Monitor status set to "up" and next expected ping calculated

---

### US-0203: Detect a missed ping and alert

**As a** user, **I want** to be alerted when my job misses its schedule **so that** I can investigate before data is lost or work goes undone.

**Acceptance criteria:**

- [ ] Background worker checks overdue monitors on a ticker (every 30s)
- [ ] Monitor status transitions to "down" when grace period expires without a ping
- [ ] Alert sent via configured channel on transition to "down" (see EP-05)
- [ ] "Expected at" and "missed by" shown on the detail page

---

### US-0204: View cron monitor list

**As a** user, **I want** to see all my cron monitors and their status at a glance **so that** I can spot problems quickly.

**Acceptance criteria:**

- [ ] Shows: name, status, last ping time, next expected ping
- [ ] Status badges: up (green), down (red), waiting (grey), paused (slate)
- [ ] Empty state with a prompt to create the first monitor

---

### US-0205: View cron monitor detail

**As a** user, **I want** to see the full history and configuration of a monitor **so that** I can diagnose missed pings.

**Acceptance criteria:**

- [ ] Shows ping URL, schedule, grace period, current status
- [ ] Execution log: timestamp and status per ping, paginated, latest first
- [ ] Incident log: when it went down, when it recovered, duration

---

### US-0206: Edit a cron monitor

**As a** user, **I want** to update a monitor's name, schedule, or grace period **so that** it stays in sync with my job.

**Acceptance criteria:**

- [ ] Name, schedule, and grace period are editable
- [ ] Ping URL / token is never changed (would break existing jobs)
- [ ] Worker picks up the new schedule immediately without restart

---

### US-0207: Pause and resume a cron monitor

**As a** user, **I want** to pause a monitor during planned maintenance **so that** I don't receive false alerts.

**Acceptance criteria:**

- [ ] Paused monitor accepts pings but does not alert on missed ones
- [ ] Status shown as "paused" in list and detail
- [ ] Resume restores normal schedule checking from that moment

---

### US-0208: Delete a cron monitor

**As a** user, **I want** to delete a monitor I no longer need **so that** my dashboard stays clean.

**Acceptance criteria:**

- [ ] Confirmation dialog required before deletion
- [ ] Ping URL invalidated immediately — further pings return 404
- [ ] All logs deleted
