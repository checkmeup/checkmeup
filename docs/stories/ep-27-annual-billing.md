# EP-27: Annual billing

> LemonSqueezy was replaced by Paddle as the payment provider/MoR on 2026-07-03 — see [ADR-026](../decisions/026-billing-paddle-mor.md). The LemonSqueezy references below (including the `LS_*` env var names) describe this epic as originally shipped; not reflective of the current integration.

Today, paid plans are monthly-only (Solo $9, Startup $29, Enterprise $99 — see [ADR-019](../decisions/019-plan-limits.md)). This epic adds an annual option at roughly 2 months free — Solo $90/yr, Startup $290/yr, Enterprise $990/yr (each is exactly 10× the monthly price) — reusing the existing LemonSqueezy MoR integration ([ADR-018](../decisions/018-billing-lemonsqueezy-mor.md)).

All 3 stories are done in code. The only remaining step is LemonSqueezy dashboard setup (3 more variants, one annual per paid plan, alongside the 3 monthly ones) — see [`docs/billing-setup.md`](../billing-setup.md). That's account/business setup only the account holder can do, not engineering work.

---

### US-2701: Choose annual or monthly at checkout

**As a** user, **I want** to pick annual or monthly billing when upgrading **so that** I can save money if I commit for a year.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [x] Pricing page shows a monthly/annual toggle; annual prices display as `$X/yr` with a "2 months free" callout
- [x] Billing page's upgrade flow offers the same toggle before creating a checkout — built alongside EP-07's previously-deferred checkout buttons
- [x] Checkout request includes the chosen cycle, mapped server-side to the correct LemonSqueezy variant (monthly or annual) for that plan — new env vars `LS_SOLO_ANNUAL_VARIANT_ID`, `LS_STARTUP_ANNUAL_VARIANT_ID`, `LS_ENTERPRISE_ANNUAL_VARIANT_ID`

---

### US-2702: Track billing cycle per org

**As a** platform operator, **I want** to know whether an org is on monthly or annual billing **so that** the billing page and renewal date reflect the real cycle.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] New `orgs.billing_cycle` column (`monthly` / `annual`), set from the LemonSqueezy webhook's variant ID — same mechanism already used to set `plan`
- [x] Billing page shows the current cycle and renewal date correctly for annual subscribers (yearly, not monthly)
- [x] Existing subscribers default to `monthly` — no backfill needed, matches reality for every org so far

---

### US-2703: Show annual pricing on the Pricing page

**As a** visitor, **I want** to see annual prices clearly **so that** I can compare the savings before signing up.

**Estimate:** 1 h

**Acceptance criteria:**

- [x]\* Toggle switches all plan cards between monthly and annual pricing, including the feature comparison table
- [x] Annual price shown as both the total (`$90/yr`) and the effective monthly rate (`$7.50/mo`) for easy comparison
- [x] Hobby stays `$0` regardless of toggle position — it has no billing cycle

\* The feature comparison table doesn't display any price figures (just monitor/interval/page counts and checkmarks), so there's nothing in it for the toggle to switch — only the plan cards above it show prices.
