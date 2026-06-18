# EP-11: Keyword monitoring

Extends the uptime monitor's HTTP check ([EP-03](ep-03-uptime-monitor.md)) with an optional text search on the response body — catches failures a `200` status code alone would miss (a maintenance page served with `200`, an error embedded in a JSON payload, a silently-broken page).

Gated to paid plans (Solo and above) — Hobby is excluded ([ADR-019](../decisions/019-plan-limits.md)).

---

### US-1101: Add a keyword check to a monitor

**As a** user, **I want** to specify text that must (or must not) appear in the response body **so that** I can detect content-level failures, not just connectivity failures.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Optional fields on uptime monitor create/edit: keyword text, mode (`Contains` / `Does not contain`)
- [x] Opt-in — existing and new uptime monitors behave exactly as before if the keyword field is left blank
- [x] Case-sensitive toggle, default off (case-insensitive)
- [x] Keyword length validated client- and server-side (1–500 chars)
- [x] Hobby plan cannot set or change a keyword (`402 plan_limit_reached`, server-enforced on create and update); a keyword already set from a previous paid plan survives a downgrade and can still be cleared, just not changed to new text

---

### US-1102: Perform the keyword check

**As a** user, **I want** the platform to search the response body for my keyword on every check **so that** I don't have to check manually.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [x] Runs as part of the existing HTTP check (US-0302) — no second request
- [x] Response body read capped at 512 KB; search runs only within that cap, regardless of `Content-Length`
- [x] Plain substring search (no regex for MVP) — keeps the worker simple and avoids a ReDoS surface
- [x] Body treated as raw text for the search — works the same whether the response is HTML, JSON, or plain text, no content-type-specific parsing

---

### US-1103: Detect a keyword failure and alert

**As a** user, **I want** a keyword mismatch treated the same as downtime **so that** I get the same alerting behavior I already rely on.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] `Contains` mode: check fails if the keyword is absent from the (capped) body
- [x] `Does not contain` mode: check fails if the keyword is present
- [x] Keyword failures follow the same 2-consecutive-failures / alert / recovery rules as status-code failures ([US-0303](ep-03-uptime-monitor.md))
- [x] Alert message distinguishes a keyword failure from an HTTP failure (e.g. "Keyword not found" vs "HTTP 500")

---

### US-1104: View keyword check status

**As a** user, **I want** to see why a check failed **so that** I can tell a content problem from a connectivity problem.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Check log shows failure reason: status-code mismatch vs keyword mismatch
- [x] Monitor list/detail shows the configured keyword and mode when set
- [x] Raw response body is never stored or displayed — only pass/fail + reason, to avoid retaining arbitrary third-party content

---

### US-1105: Edit or remove a keyword check

**As a** user, **I want** to change or clear my keyword check **so that** it stays accurate as my page's content changes.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] Keyword, mode, and case-sensitivity editable independently of the monitor's other fields
- [x] Clearing the keyword field disables the check — monitor reverts to status-code-only checking
- [x] Change takes effect on the next scheduled check; no forced immediate re-check
