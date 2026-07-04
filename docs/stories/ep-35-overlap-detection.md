# EP-35: Overlap detection

Nothing today tracks whether a previous run finished before a new one started — a cron ping is a single instantaneous event with no run-in-progress state (same gap [EP-34](ep-34-zombie-job-detection.md) is built to close). Overlap detection alerts when a job's next scheduled run begins while the prior run is still going — the race-condition/resource-exhaustion risk called out alongside zombie detection as one of the differentiators that actually matters, versus another plain heartbeat check.

**Depends on [EP-34](ep-34-zombie-job-detection.md) US-3401 shipping first** — overlap has nothing to compare against without the start-of-run ping that epic introduces. Covered by the same ping-model design decision noted there.

---

### US-3501: Detect a start ping while a run is still in progress

**As a** user, **I want** to be notified if my job starts again before the previous run finished **so that** I can catch race conditions or a job silently piling up concurrent runs.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] When a start ping (EP-34 US-3401) arrives and the monitor's last run has no completion ping yet, flag it as an overlap
- [ ] Each overlapping run tracked independently — a job starting a third time while two are already overlapping is still detected
- [ ] Monitors that never use the start ping are unaffected — no false overlaps from single-ping jobs

---

### US-3502: Alert on overlap

**As a** user, **I want** an alert when an overlap is detected **so that** I know before it causes a bigger problem.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Alert is distinct from a "down"/missed-ping alert — names it as an overlap and states when the still-running run started
- [ ] Routed through the same per-monitor notification channels as other cron alerts ([EP-28](ep-28-notification-channels.md))
- [ ] Respects `max_alerts_per_incident` (ADR-016) so a job with pathological overlapping doesn't spam

---

### US-3503: View overlap history

**As a** user, **I want** to see past overlap incidents in the monitor's log **so that** I can tell a one-off from a systemic problem with a specific job.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Overlap events appear in the incident log, distinct from missed-ping incidents
- [ ] Shows both runs' start times, and durations for any that have since completed
