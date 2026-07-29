---
title: "EP-39: DNS record monitoring"
type: story
status: planned            # planned | in-progress | shipped | cancelled
updated: 2026-07-29
tags: [monitor, dns]
---

# EP-39: DNS record monitoring

A DNS monitor watches a hostname's resolved DNS record on a fixed interval and alerts when it changes or resolves unexpectedly — e.g. an A record silently pointing somewhere it shouldn't (DNS hijacking, an expired NS delegation, an accidental change during a migration). Distinct from the uptime monitor ([EP-03](ep-03-uptime-monitor.md)), which only cares whether the *current* resolved address answers HTTP requests, not whether the record itself is the one the owner expects. Identified as a gap vs. UptimeRobot in `docs/proposals/bucket-list.md`'s "New monitor types" section.

Two modes, both covering the bucket-list wording ("changes *or* resolves unexpectedly"): if the user supplies an **expected value**, the monitor alerts on any mismatch from creation onward (a security check — confirming a record stays pinned to a known-good value). If left blank, the first successful check establishes a **baseline** value, and the monitor alerts the first time a later check resolves to something different (a drift/tamper check when the "correct" value isn't known in advance, only that it shouldn't move).

Counts toward the org's aggregate monitor limit alongside cron/uptime/SSL/domain/port ([ADR-019](../decisions/019-plan-limits.md)) — implementing this epic requires updating that ADR's limits table to read "cron + uptime + SSL + domain + port + DNS".

---

### US-3901: Create a DNS record monitor

**As a** user, **I want** to add a hostname and record type to monitoring **so that** I'm warned if its DNS record changes or stops resolving as expected.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Fields: name, hostname (e.g. `example.com` or `www.example.com`), record type (A / AAAA / CNAME / MX / TXT / NS — dropdown), expected value (optional free text; comma-separated for multiple expected values, e.g. multiple A or MX records)
- [ ] Interval options match uptime's plan-gated minimums (ADR-019: 5 min Hobby, 1 min paid) — a security-relevant check like port monitoring's "closed" mode, not just an uptime signal
- [ ] First check runs within one interval of creation; if no expected value was given, that first successful lookup becomes the stored baseline
- [ ] Counts toward the org's aggregate monitor limit, enforced the same way as cron/uptime/SSL/domain/port creation (ADR-019)

---

### US-3902: Perform DNS record check

**As a** user, **I want** the platform to look up my DNS record automatically **so that** I don't have to run `dig`/`nslookup` by hand to notice a change.

**Estimate:** 2 h

**Acceptance criteria:**

- [ ] Lookup via Go's `net` package, method selected by record type (`net.LookupHost` for A/AAAA, `net.LookupCNAME`, `net.LookupMX`, `net.LookupTXT`, `net.LookupNS`)
- [ ] 10-second lookup timeout, same as EP-03/EP-33's timeout (ADR-014)
- [ ] Resolved value(s) sorted before comparison, so multi-value answers (multiple A records, multiple MX hosts) in a different order aren't reported as a change
- [ ] **Expected-value mode**: current value(s) compared against the stored expected value; match = up, mismatch = down
- [ ] **Baseline mode**: current value(s) compared against the stored baseline; first differing check = down (baseline is *not* silently updated to the new value — stays down until acknowledged by editing the monitor)
- [ ] NXDOMAIN, SERVFAIL, or lookup timeout recorded as an error state, distinct from a value mismatch — never reported as "changed" when the real record is simply unreachable
- [ ] Lookup time recorded on every check

---

### US-3903: Detect record change and alert

**As a** user, **I want** to be alerted the moment my DNS record changes or fails to resolve as expected **so that** I can respond before it's a bigger problem (a hijack, a broken migration, a squatted domain).

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Down transition follows the same `alert_after_n_failures` filter as other monitor types (default 0 = alert on the first mismatched/failed check)
- [ ] Alert message states old value → new value for a mismatch/drift; states the lookup error (NXDOMAIN/SERVFAIL/timeout) for a resolution failure
- [ ] Alert sent on recovery (value returns to the expected value, or the monitor is edited to accept the new value as the baseline/expected value)
- [ ] Alert cap via `max_alerts_per_incident` (0=always, default 3 — see [ADR-016](../decisions/016-alert-debounce.md))

---

### US-3904: View DNS monitor list and detail

**As a** user, **I want** to see the status and current value of all my DNS monitors **so that** I can spot an unexpected change at a glance.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] List: name, hostname, record type, status, current resolved value, last checked time — status badges consistent with other monitor types
- [ ] Detail: expected value or baseline (labeled which mode it's in), current value, lookup-time chart for the last 24 hours
- [ ] Change log: when the resolved value changed, old value, new value — paginated, latest first
- [ ] Check log: timestamp, up/down, resolved value, error (if any) — paginated

---

### US-3905: Edit, pause, and delete a DNS monitor

**As a** user, **I want** to manage a DNS monitor's settings **so that** I can accept an intentional change without it staying flagged as down.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Editable: name, hostname, record type, expected value, check interval
- [ ] Saving a new expected value (or clearing it back to baseline mode) re-arms the monitor as up against the new value on the next check, same as acknowledging an intentional change
- [ ] Pause stops checks and suppresses alerts; excluded from the due-check query entirely (same as EP-04/EP-33's shipped behavior)
- [ ] Resume restarts checks immediately
- [ ] Delete requires confirmation; all history deleted
