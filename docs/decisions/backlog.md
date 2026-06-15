# Decision backlog

Open questions that need an answer before (or early in) the relevant implementation. When decided, create a numbered ADR and remove the entry here.

---

## cron_pings retention / TTL

**Question:** How long do we keep individual ping records, and who deletes old ones?

Specifics to decide:
- Retention policy: keep last N pings per monitor (e.g. 1,000) or keep pings newer than T days (e.g. 90 days)?
- Implementation: background worker ticker or a DB-level `DELETE ... WHERE received_at < NOW() - INTERVAL '90 days'`?
- Does the UI need pings older than 90 days? (Current detail view shows last 50; pagination goes further.)
- Incidents have no FK to pings — safe to prune freely.

**Why it matters:** a monitor pinging every minute generates ~525,000 rows/year. At Agency plan (300 monitors) that is 157 M rows/year per customer. The Hetzner CX23 has an 80 GB SSD — unbounded growth is an unexpected infrastructure expense.

**Needed before:** Phase 3 (first paid customers could have many monitors running for months).

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

