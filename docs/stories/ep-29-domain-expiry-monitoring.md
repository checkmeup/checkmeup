# EP-29: Domain expiry monitoring

A domain monitor checks a domain's *registration* expiry on a daily interval and alerts at multiple thresholds before it lapses — distinct from the TLS certificate expiry already covered by the SSL monitor ([EP-04](ep-04-ssl-monitor.md)), and rarer but more catastrophic when it happens (the site disappears entirely, and the domain is sometimes squatted before the owner notices). Same alerting shape as EP-04, different data source.

Counts toward the org's aggregate monitor limit alongside cron/uptime/SSL ([ADR-019](../decisions/019-plan-limits.md)) — implementing this epic requires updating that ADR's limits table to read "cron + uptime + SSL + domain".

**Shipped 2026-06-23**, with two narrower-than-planned items: the lookup uses RDAP only (via the [rdap.org](https://rdap.org) bootstrap redirector) — the WHOIS text-parsing fallback in US-2902 for TLDs without RDAP support was not built, since it needs a separate per-registry text parser and felt like a half-finished addition to rush; and the "alert immediately on registry hold/pending-delete status" half of US-2903 was dropped since the RDAP client doesn't parse the domain `status` array yet, only the expiration event and registrar entity. Both are good candidates for a follow-up if a real domain hits either gap.

---

### US-2901: Create a domain expiry monitor

**As a** user, **I want** to add a domain to expiry monitoring **so that** I'm warned before its registration lapses.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Field: domain (e.g. `example.com` — apex domain only, same validation pattern as EP-04's `hostname` field)
- [x] First check runs immediately on creation
- [x] Shows registrar, expiry date, and days remaining after first check
- [x] Counts toward the org's aggregate monitor limit, enforced the same way as cron/uptime/SSL creation (ADR-019)

---

### US-2902: Check domain registration daily

**As a** user, **I want** the platform to check my domain's registration automatically **so that** I always have an up-to-date picture without running WHOIS by hand.

**Estimate:** 2 h

**Acceptance criteria:**

- [x] Looked up via RDAP (RFC 9082/9083 — structured JSON, no text parsing) where the TLD's registry supports it
- [ ] Falls back to WHOIS text parsing for TLDs without RDAP support — not built, see shipped note above
- [x] Checked once per day (same cadence as EP-04's SSL check)
- [x] Records: registrar, expiry date, days remaining, valid/error flag
- [x] Lookup failure (registry timeout, unsupported TLD, unparsable response) recorded as an error state — never reported as "expired" when the real status is unknown

---

### US-2903: Alert on approaching expiry

**As a** user, **I want** to receive alerts as my domain nears expiry **so that** I have time to renew it before it's lost.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Alerts sent at 30 days, 14 days, and 7 days before expiry — one alert per threshold, same pattern as EP-04 ([ADR-016](../decisions/016-alert-debounce.md))
- [x] Alert sent immediately if the domain is already expired
- [ ] Alert sent immediately for a registry hold/pending-delete status — not built, RDAP client doesn't parse the domain `status` array yet, see shipped note above
- [x] No repeated alerts at the same threshold

---

### US-2904: View domain monitor list and detail

**As a** user, **I want** to see the status and expiry of all my domains **so that** I can plan renewals.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] List: domain, status (valid / expiring soon / expired / error), expiry date, days remaining
- [x] "Expiring soon" shown when ≤ 30 days remaining, same threshold as EP-04
- [x] Detail: registrar, expiry date, days remaining, last checked time, error message if any

---

### US-2905: Pause and delete a domain monitor

**As a** user, **I want** to pause or remove a domain monitor **so that** I don't get alerts for domains I no longer own.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] Pause suppresses alerts and checks, same as EP-04's shipped SSL monitor behavior (a paused monitor is excluded from the due-check query entirely, not just alert-suppressed)
- [x] Delete requires confirmation; all history deleted
- [x] Domain is editable only by deleting and recreating, same as EP-04's hostname field
