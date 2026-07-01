# EP-33: Port (TCP) monitoring

A port monitor opens a raw TCP connection to a host:port on a fixed interval — no HTTP request, no response body. Same up/down/alert shape as the uptime monitor ([EP-03](ep-03-uptime-monitor.md)), different transport: this covers non-HTTP services (mail servers, databases, custom daemons) that a URL-based check can't reach. Identified as a gap vs. UptimeRobot in `docs/bucket-list.md`'s "New monitor types" section.

Each monitor has an **expected state** of either "open" (default — alert if the port stops accepting connections) or "closed" (alert if the port unexpectedly *starts* accepting connections). The closed mode is a security check: confirming a port that should be firewalled off (e.g. a database bound to a public interface, an admin panel, a debug port) actually stays unreachable, and catching the moment it doesn't — a misconfiguration or exposure, not an uptime concern.

Counts toward the org's aggregate monitor limit alongside cron/uptime/SSL/domain ([ADR-019](../decisions/019-plan-limits.md)) — implementing this epic requires updating that ADR's limits table to read "cron + uptime + SSL + domain + port".

**Shipped 2026-07-01**, with two adjustments from the original spec: US-3302's "DNS/resolution failure recorded as a distinct error state" for expected-state=closed wasn't built — `net.Dial` doesn't cleanly distinguish DNS failure from connection-refused/timeout without an extra lookup call, and the `monitor_status` enum this reuses (`waiting/up/down/paused`, same as EP-03) has no `error` state to put it in anyway. Any dial failure counts as "matches expectation" when closed. Separately, US-3303's "2 consecutive failed checks" wording (copied from EP-03's original spec) doesn't match either monitor type's current behavior since the `alert_after_n_failures` filter (migration `023_alert_filter.sql`) was added — both now alert based on that filter (default 0, i.e. alert on the first failure), which is what port monitoring implements too.

---

### US-3301: Create a port monitor

**As a** user, **I want** to add a host and port to monitoring **so that** I know when a non-HTTP service stops accepting connections.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Fields: name, host (hostname or IP), port (1–65535), check interval, expected state (open / closed — default open)
- [x] Interval options match uptime's plan-gated minimums (ADR-019: 5 min Hobby, 1 min paid)
- [x] First check runs within one interval of creation
- [x] Counts toward the org's aggregate monitor limit, enforced the same way as cron/uptime/SSL/domain creation (ADR-019)

---

### US-3302: Perform TCP connect check

**As a** user, **I want** the platform to check my host:port automatically **so that** I don't have to run my own connectivity checks.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [x] Raw TCP connect (`net.Dial("tcp", host:port)`) — no data sent or received, no protocol-specific handshake
- [x] 10-second connection timeout, same as EP-03's HTTP timeout (ADR-014)
- [x] **Expected state = open** (default): successful connect = up; connection refused, timeout, or DNS/resolution failure = down
- [x] **Expected state = closed**: connection refused or timeout = up (matches expectation); successful connect = down (the port is unexpectedly reachable)
- [ ] DNS/resolution failure recorded as a distinct error state when closed — not built, see shipped note above; treated the same as connection refused/timeout (i.e. "matches expectation")
- [x] Connection is closed immediately after any successful connect — no data exchanged, regardless of expected state
- [x] Response time (time to establish the connection, or time to receive the refusal/timeout) recorded on every check

---

### US-3303: Detect state change and alert

**As a** user, **I want** to be alerted when my port stops matching its expected state **so that** I can respond quickly — whether that's a service going down or a port being unexpectedly exposed.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Down transition follows the same `alert_after_n_failures` filter as EP-03/EP-02 (default 0 = alert on the first failure) — applies identically to both expected states, since "failed" already means "didn't match expectation" (see shipped note above re: the story's original "2 consecutive failures" wording)
- [x] Alert sent on transition to "down" (see EP-05); for expected-state=closed, the alert message says the port is unexpectedly open, not "down"
- [x] Alert sent on transition back to "up" (recovery)
- [x] Alert cap via `max_alerts_per_incident` (0=always, default 3 — see ADR-016)

---

### US-3304: View port monitor list and detail

**As a** user, **I want** to see the status and history of all my port monitors **so that** I can spot problems at a glance and diagnose outages.

**Acceptance criteria:**

- [x] List: name, host:port, expected state, status, uptime % (last 24h), last checked time — status badges consistent with other monitor types
- [x] Detail: connect-time chart for the last 24 hours, uptime % for last 24h/7d/30d, expected state shown alongside the host:port
- [x] Incident log: started at, resolved at, duration — paginated, latest first
- [x] Check log: timestamp, up/down, connect time, error (if any) — paginated

**Estimate:** 1.5 h

---

### US-3305: Edit, pause, and delete a port monitor

**As a** user, **I want** to manage a port monitor's settings **so that** it stays accurate and doesn't generate noise during maintenance.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] Editable: name, host, port, check interval, expected state
- [x] Pause stops checks and suppresses alerts; excluded from the due-check query entirely (same as EP-04's shipped behavior, not just alert-suppressed)
- [x] Resume restarts checks immediately
- [x] Delete requires confirmation; all history deleted
