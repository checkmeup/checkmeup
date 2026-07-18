# EP-37: Configurable uptime check parameters

Today an uptime monitor's HTTP method (always `GET`) and "up" condition (exact status `200`, nothing else — [EP-03](ep-03-uptime-monitor.md), [ADR-014](../decisions/014-uptime-check-mechanics.md)) are hardcoded. ADR-014 flagged the strict 200-only rule as an acceptable MVP simplification to revisit if users hit it — a `201` from a create endpoint or a `204` from a delete endpoint is a legitimate healthy response that today gets alerted on as down. This epic makes both configurable per monitor, defaulting to today's exact behavior so existing monitors see zero change.

The request timeout is a third configurable parameter, but it isn't a new field — `uptime_monitors.max_response_time_ms` already exists ([EP-31](ep-31-assertion-checks.md), migration `020_assertion_checks.sql`) as an optional, nullable, per-monitor value. Today it's enforced *after* a successful response as a post-hoc SLA check (`elapsed > maxResponseTimeMs` → "response time exceeded"), completely separate from the actual network-level timeout, which is a shared, hardcoded 10s constant (`deliver.Timeout`, also reused by unrelated outbound HTTP like alert delivery). That's two knobs doing overlapping jobs — a monitor with a 3s max response time still lets the request run the full 10s before failing it retroactively. This epic collapses them into one: `max_response_time_ms` becomes **required**, default `10000` (10s, today's exact default), and is used directly as the HTTP client's request timeout — a slow response now aborts the connection at the configured ceiling instead of completing and then being judged. The old post-hoc check becomes dead code and is removed.

Builds on the same per-monitor check model as keyword monitoring ([EP-11](ep-11-keyword-monitoring.md)) and JSON assertion checks ([EP-31](ep-31-assertion-checks.md)) — those still run on top of the status-code check, unaffected by this epic.

---

### US-3701: Configure check method, timeout, and accepted status codes on an uptime monitor

**As a** user, **I want** to customize the HTTP method, timeout, and which status codes count as "up" for my uptime monitor, **so that** the check matches how my endpoint actually behaves instead of assuming every healthy response is an exact `200` from a `GET`.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] HTTP method field on create/edit (`GET` / `HEAD` / `POST`, default `GET`) — genuinely new field
- [ ] `max_response_time_ms` ("Max response time") becomes **required**, default `10000` (10s) — drop the "(optional)" label, migration backfills existing NULLs to `10000` and adds `NOT NULL DEFAULT 10000`
- [ ] Accepted status codes field (multiselect, default `[200]`) — genuinely new field
- [ ] Existing monitors get today's exact defaults for all three — zero behavior change for any monitor that doesn't touch these fields
- [ ] At least one status code must be selected — cannot save with an empty set
- [ ] Fields shown collapsed under an "advanced" section on the create/edit form, not competing with name/URL/interval for attention

---

### US-3702: Worker performs the check using the monitor's configured parameters

**As a** user, **I want** the actual check to respect my configuration, **so that** the monitor reflects my endpoint's real behavior instead of a hardcoded assumption.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] Worker issues the configured HTTP method instead of always `GET`
- [ ] Worker uses `max_response_time_ms` as the actual per-request HTTP client timeout, replacing the shared `deliver.Timeout` constant for uptime checks specifically (a per-monitor client/context deadline, not the process-wide shared client — `deliver.Timeout` stays as-is for alert delivery, which this epic doesn't touch)
- [ ] A response counts as "up" when its status code is in the configured accepted set, "down" otherwise — replaces the current `resp.StatusCode != http.StatusOK` check
- [ ] The old post-hoc `elapsed > MaxResponseTimeMs` → "response time exceeded" check is removed — now unreachable, since a response that would have exceeded it already errors out as a timeout before the check would run
- [ ] Redirect handling unchanged (still followed automatically per ADR-014); the accepted-status-codes check applies to the final response in the chain
- [ ] Keyword ([EP-11](ep-11-keyword-monitoring.md)) and JSON-body assertion ([EP-31](ep-31-assertion-checks.md)) checks still run on top of the status-code check exactly as today — an accepted status code alone doesn't skip them

---

### US-3703: Validate method, timeout, and status code inputs

**As a** user, **I want** invalid configuration rejected clearly, **so that** I can't accidentally break a monitor with an unsupported method or a nonsensical timeout.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] HTTP method restricted to a fixed whitelist (`GET`, `HEAD`, `POST`) — anything else rejected with a clear error, both create and edit
- [ ] `max_response_time_ms` bounded to 1,000–30,000 (1–30 seconds) and required (replaces today's `<= 0` check, which allowed unbounded-above and nil) — below 1s isn't a meaningful check, above 30s risks piling up against the worker's per-cycle check budget
- [ ] Status codes restricted to the valid HTTP range (100–599)
- [ ] Validation enforced server-side as the authoritative check, not just client-side on the form
