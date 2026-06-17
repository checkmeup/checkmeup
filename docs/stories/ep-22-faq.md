# EP-22: FAQ

A billing-specific FAQ already exists, hardcoded as a local array in `PricingView.vue` (`faqs`, 6 entries — credit card, monitor definition, plan changes, limits, payment methods, refunds). This epic generalizes it into a dedicated `/faq` page covering the whole product, with the existing entries folded in rather than duplicated.

---

### US-2201: Publish a general FAQ page

**As a** visitor, **I want** a dedicated FAQ page covering the whole product, not just pricing, **so that** I can find answers without digging through marketing pages.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] Public, unauthenticated `/faq` page
- [ ] Covers areas beyond billing: monitor types and check intervals, status pages, alert channels, data retention
- [ ] The existing 6 billing entries move here as the single source of truth — `PricingView.vue` either links to `/faq#billing` or imports the same entries, not a second copy
- [ ] Linked from the site footer and nav, same placement as Terms/Privacy ([EP-21](ep-21-terms-and-privacy.md))

---

### US-2202: Organize FAQ by category

**As a** visitor, **I want** FAQ entries grouped by topic **so that** I can scan to the section I care about.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Categories: Getting started, Billing & plans, Monitors & alerts, Status pages, Privacy & security
- [ ] The existing 6 billing entries slot into "Billing & plans" unchanged
- [ ] Anchor links per category (e.g. `/faq#billing`) so other pages can deep-link into a specific section

---

### US-2203: Link FAQ from the rest of the product

**As a** user, **I want** FAQ links where I'm likely to have a question **so that** I don't have to go searching for the page.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] `PricingView.vue`'s FAQ section replaced with a link to `/faq#billing` (or the shared entries, per US-2201) instead of its own copy
- [ ] Footer link present on every public page (landing, blog, status pages), alongside Terms/Privacy
- [ ] Sign-up page links to the "Getting started" category
