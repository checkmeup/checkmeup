# EP-31: Assertion-based API checks

Extends the uptime monitor's keyword check ([EP-11](ep-11-keyword-monitoring.md)) with structured pass/fail assertions beyond a body substring match: a JSON response field's value, a response-time threshold treated as a first-class failure condition (not just informational), and — as a larger follow-on — a short chain of dependent requests. This is Cronitor's headline differentiator over checkmeup today (see `docs/proposals/bucket-list.md`); building it on EP-11's existing check-evaluation pipeline avoids introducing a new monitor type.

---

### US-3101: Assert on a JSON response field value

**As a** user, **I want** to assert that a field in my API's JSON response has a specific value **so that** I can catch a `200` response that's actually unhealthy (e.g. `{"status":"degraded"}`).

**Estimate:** 2.5 h

**Acceptance criteria:**

- [x] New assertion type alongside the existing keyword check (EP-11): JSON path (e.g. `$.status` or `data.healthy`), comparator (`equals` / `not_equals` / `contains` / `greater_than` / `less_than`), expected value
- [x] Multiple JSON assertions on one monitor are ANDed — all must pass
- [x] Response body that isn't valid JSON, when a JSON assertion is configured, fails the check with a distinct reason ("response is not valid JSON") rather than silently passing
- [x] Reuses EP-11's existing 512 KB body cap — JSON assertions only see content within that cap

---

### US-3102: Response-time threshold as pass/fail

**As a** user, **I want** to set a maximum acceptable response time **so that** a slow-but-200 response is treated as a failure, not just shown as a number.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Optional field on the uptime monitor: max response time (ms)
- [x] Exceeding it fails the check with reason "response time exceeded" — distinct from a connection-level timeout, which already fails today
- [x] Independent of keyword/JSON assertions — any single failing condition fails the whole check

---

### US-3103: Combine and evaluate all assertions

**As a** user, **I want** all my configured conditions evaluated together on every check **so that** a single check call covers status code, content, and timing.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [x] One HTTP check run evaluates, in order: status code, keyword (EP-11), JSON assertions (US-3101), response-time threshold (US-3102) — the first failing condition is the recorded failure reason
- [x] Same 2-consecutive-failures / alert / recovery state machine as existing checks ([US-0303](ep-03-uptime-monitor.md)) — no new alerting model
- [x] Opt-in — a monitor with no assertions configured behaves exactly as it does today

---

### US-3104: View assertion configuration and results

**As a** user, **I want** to see my configured assertions and which one failed **so that** I can tell a content problem from a timing problem from a connectivity problem.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Monitor list/detail shows configured assertions (JSON path + comparator + expected value, response-time threshold) alongside the existing keyword display
- [x] Check log failure reason distinguishes status-code vs keyword vs JSON-assertion vs response-time failures
- [x] Raw response body is still never stored or displayed, consistent with EP-11 US-1104

---

### US-3105: Multi-step (chained) API checks

**As a** user, **I want** to chain a short sequence of requests in one check **so that** I can verify a flow (e.g. log in, then call an authenticated endpoint) rather than a single isolated request.

**Estimate:** 5 h — significantly larger than the rest of this epic; re-scope and consider splitting into its own epic during grooming, since request-chaining and value-templating aren't needed by US-3101–3104.

**Acceptance criteria:**

- [ ] A monitor can optionally define an ordered sequence of 2–5 HTTP requests, each with its own assertions (US-3101–3102)
- [ ] A later step can reference a value extracted from an earlier step's JSON response (e.g. an auth token) using the same JSON path syntax as US-3101
- [ ] The whole chain counts as one check; the first failing step stops the chain and is the recorded failure reason
- [ ] Single-request monitors (the common case) are entirely unaffected — this is opt-in and additive
