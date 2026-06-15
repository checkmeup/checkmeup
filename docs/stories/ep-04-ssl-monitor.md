# EP-04: SSL monitor

An SSL monitor checks a domain's TLS certificate daily and alerts at multiple thresholds before expiry.

---

### US-0401: Create an SSL monitor

**As a** user, **I want** to add a domain to SSL monitoring **so that** I'm warned before a certificate expires.

**Acceptance criteria:**

- [x] Field: domain (e.g. `example.com` — no protocol, no path)
- [x] First check runs immediately on creation
- [x] Shows expiry date and days remaining after first check

---

### US-0402: Check SSL certificate daily

**As a** user, **I want** the platform to check my certificate automatically **so that** I always have an up-to-date picture.

**Acceptance criteria:**

- [x] Certificate checked once per day
- [x] Records: issuer, expiry date, days remaining, valid/invalid flag
- [x] Invalid or unreachable certificate recorded as an error state

---

### US-0403: Alert on approaching expiry

**As a** user, **I want** to receive alerts as my certificate nears expiry **so that** I have time to renew it.

**Acceptance criteria:**

- [x] Alerts sent at 30 days, 14 days, and 7 days before expiry — one alert per threshold
- [x] Alert sent immediately if certificate is already expired or invalid
- [x] No repeated alerts at the same threshold (only once per crossing)

---

### US-0404: View SSL monitor list and detail

**As a** user, **I want** to see the status and expiry of all my certificates **so that** I can plan renewals.

**Acceptance criteria:**

- [x] List: domain, status (valid / expiring soon / expired / error), expiry date, days remaining
- [x] "Expiring soon" shown when ≤ 30 days remaining
- [x] Detail: issuer, expiry date, days remaining, last checked time, error message if any

---

### US-0405: Pause and delete an SSL monitor

**As a** user, **I want** to pause or remove an SSL monitor **so that** I don't get alerts for domains I'm decommissioning.

**Acceptance criteria:**

- [x] Pause suppresses alerts but keeps running daily checks
- [x] Delete requires confirmation; all history deleted
- [x] Domain is editable only by deleting and recreating (changing the domain is a different monitor)
