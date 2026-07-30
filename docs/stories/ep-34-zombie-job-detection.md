# EP-34: Zombie (stuck) job detection

Today a cron monitor only ever receives one kind of ping — `GET /ping/{token}` (EP-02 US-0202) — a single instantaneous check-in with no concept of "run started" vs "run finished." That's enough to detect a job that never pinged at all (EP-02 US-0203's missed-ping/grace-period alert), but not a job that started, is still running far longer than normal, and hasn't crashed or hung in a way that stops it from eventually pinging — the "zombie job" gap raised repeatedly as a real-world cron-monitoring pain point (silent successes aside, a job stuck in a loop consuming CPU for hours doesn't miss its ping window, it just delays it).

**Ping-model design decided in [ADR-039](../decisions/039-cron-ping-model.md):** detecting a stuck run requires a start-of-run signal distinct from today's single completion ping. `GET /ping/{token}/start` is a new, purely-additive endpoint — jobs that never call it keep exactly today's behavior (single ping = success, EP-02 unaffected) — backed by a new `cron_runs` table (one row per run, not a column on `cron_monitors`) so run history is available for US-3404 without a second table later. [EP-35](ep-35-overlap-detection.md) (overlap detection) shares this same start-ping signal and table.

---

### US-3401: Send a start-of-run ping

**As a** user, **I want** to ping a start URL when my job begins **so that** checkmeup knows when it started and how long it's been running.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] New endpoint distinct from the existing completion ping (e.g. `GET /ping/{token}/start`) records a run-start timestamp
- [ ] Opt-in: a monitor that never receives a start ping behaves exactly as it does today — zero change for existing jobs
- [ ] A run is considered "in progress" from the start ping until the next completion ping for that token
- [ ] A completion ping with no preceding start ping is accepted as-is (today's behavior), not treated as an error

---

### US-3402: Configure max expected run duration

**As a** user, **I want** to set a maximum expected duration for a monitor **so that** I get an early alert if a job is stuck rather than waiting on the normal schedule.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Optional `max_duration_mins` field on create/edit, only meaningful for monitors using the start ping (US-3401)
- [ ] Left unset, zombie detection stays inactive for that monitor — no behavior change
- [ ] Editable independently of schedule and grace period

---

### US-3403: Detect and alert on a stuck run

**As a** user, **I want** an alert when a run exceeds its max expected duration without completing **so that** I find out before it blocks the next scheduled run or exhausts server resources.

**Estimate:** 2 h

**Acceptance criteria:**

- [ ] Existing overdue-check worker ticker (EP-02 US-0203, 30s) also checks in-progress runs against `max_duration_mins`
- [ ] Alert fires once when a run first exceeds its max duration, not repeatedly on every subsequent tick
- [ ] Alert message distinguishes "stuck run" (started at X, still running after Y) from a standard missed-ping alert
- [ ] Recovery: alert clears once the completion ping arrives, or the run is superseded by a new start ping

---

### US-3404: View run duration history

**As a** user, **I want** to see how long recent runs took **so that** I can pick a sensible max-duration threshold instead of guessing.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Execution log (EP-02 US-0205) shows duration per completed run (start ping to completion ping) for monitors using US-3401
- [ ] Runs without a start ping show duration as "n/a", not zero or blank
