# Decision backlog

Open questions that need an answer before (or early in) the relevant implementation. When decided, create a numbered ADR and remove the entry here.

---

## Alert debounce / cooldown

**Question:** What is the logic for sending a Telegram alert vs. staying silent?

Specifics to decide:
- How many consecutive failures before alerting? (e.g. 2-of-3)
- How long to wait before re-alerting on a still-down monitor?
- Alert on recovery?
- Per-monitor override vs. global setting?

**Needed before:** implementing the alert system (Phase 2).

---

## Uptime check mechanics

**Question:** How exactly does an uptime check work?

Specifics to decide:
- HTTP method: HEAD first, fall back to GET — or always GET?
- Timeout per request
- What counts as "down": non-2xx, connection timeout, both?
- Minimum check interval granularity
- Follow redirects or flag them?

**Needed before:** implementing the uptime monitor worker (Phase 3).

---

## Chi router rationale

**Question:** Document why Chi over alternatives (gin, echo, fiber, stdlib `net/http`).

**Needed before:** not blocking — capture when the first route is written.
