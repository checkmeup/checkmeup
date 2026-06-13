# EP-07: Billing & plan limits

Stripe handles payment. Plan limits are enforced server-side. Hobbyist ($0) requires no payment flow — it's the default for all new accounts.

Plan limits reference: see [backlog.md](backlog.md).

---

### US-0701: View current plan and usage

**As a** user, **I want** to see my current plan and how much of my quota I've used **so that** I know when to upgrade.

**Acceptance criteria:**

- [ ] Shows: plan name, billing period, next renewal date
- [ ] Usage bars for: monitors used / limit, status pages used / limit
- [ ] Min check interval shown for the current plan
- [ ] "Upgrade" CTA shown for non-Agency plans

---

### US-0702: Enforce plan limits

**As a** user, **I want** a clear message when I hit a plan limit **so that** I understand why I can't add more monitors and what to do.

**Acceptance criteria:**

- [ ] API returns `402 Payment Required` with a descriptive message when a limit is exceeded
- [ ] UI shows an inline upgrade prompt — not a generic error page
- [ ] Creating a monitor over the limit is blocked at both API and UI level
- [ ] Setting a check interval below the plan minimum is blocked with an explanation

---

### US-0703: Upgrade plan via Stripe Checkout

**As a** user, **I want** to upgrade my plan **so that** I can add more monitors.

**Acceptance criteria:**

- [ ] Stripe Checkout session created server-side; user redirected to Stripe
- [ ] On successful payment, Stripe webhook updates the org's plan in DB
- [ ] New limits applied immediately after webhook received
- [ ] Failed payment returns user to the billing page with an error
- [ ] Downgrade takes effect at end of current billing period — no immediate cutoff
