# EP-07: Billing & plan limits

LemonSqueezy (Merchant of Record) handles payment, tax collection, and invoicing. Plan limits are enforced server-side. Hobby ($0) requires no payment flow — it's the default for all new accounts. See [ADR-018](../decisions/018-billing-lemonsqueezy-mor.md) and [ADR-019](../decisions/019-plan-limits.md).

Plan limits:

| | Hobby | Solo ($9) | Startup ($29) | Enterprise ($99) |
|---|---|---|---|---|
| Total monitors | 10 | 30 | 100 | 1000 |
| Status pages | 1 | 3 | 10 | 100 |
| Min uptime interval | 5 min | 1 min | 1 min | 1 min |

---

### US-0701: View current plan and usage

**As a** user, **I want** to see my current plan and how much of my quota I've used **so that** I know when to upgrade.

**Acceptance criteria:**

- [x] Shows: plan name, price, renewal date (if subscribed)
- [x] Usage bars for: monitors used / limit, status pages used / limit
- [x] Min check interval shown for the current plan
- [x] "Upgrade" CTA shown for non-Enterprise plans
- [x] "Manage subscription →" link for active subscribers (LemonSqueezy customer portal)

---

### US-0702: Enforce plan limits

**As a** user, **I want** a clear message when I hit a plan limit **so that** I understand why I can't add more monitors and what to do.

**Acceptance criteria:**

- [x] API returns `402 Payment Required` with `code: "plan_limit_reached"` when a limit is exceeded
- [x] Blocked: creating monitors over the total limit (cron, uptime, SSL)
- [x] Blocked: creating status pages over the limit
- [x] Blocked: setting uptime interval below plan minimum (returns 402 with explanation)
- [ ] UI shows inline upgrade prompt when a 402 is received — not a generic error page *(deferred post-MVP)*

---

### US-0703: Upgrade plan via LemonSqueezy

**As a** user, **I want** to upgrade my plan **so that** I can add more monitors.

**Acceptance criteria:**

- [x] LemonSqueezy Checkout session created server-side; user redirected to LemonSqueezy
- [x] On successful payment, LemonSqueezy webhook updates the org's plan in DB
- [x] New limits applied immediately after webhook received
- [x] Cancellation: plan stays active until `ends_at`, then reverts to Hobby
- [ ] Failed payment: user lands back on billing page (LemonSqueezy handles this natively) *(deferred post-MVP — verify redirect URL when activating billing)*
