# ADR-018: Billing via LemonSqueezy (Merchant of Record)

**Date:** 2026-06-15  
**Status:** Accepted

---

## Context

checkmeup needs a billing system for its subscription plans (Hobbyist $0 / Indie $9 / Studio $29 / Agency $79). The founder is an Israeli resident operating as a self-employed individual (עוסק פטור) without a registered company.

Key constraints:
- Must handle global tax collection and remittance automatically (VAT, GST, US sales tax) — no capacity to manage multi-country tax compliance manually
- Self-serve setup — no sales calls or enterprise contracts
- Minimal ongoing accounting overhead outside Israel

---

## Alternatives considered

| Option | MoR | Fee | Tax handling | Ruled out because |
|---|---|---|---|---|
| Stripe | ❌ | ~2.9% + $0.30 | Stripe Tax collects, founder remits | Founder still responsible for filing in every jurisdiction |
| Paddle | ✅ | ~5% + $0.50 | Full MoR | Good option; LemonSqueezy simpler for MVP stage |
| PayPro Global | ✅ | Unknown | Full MoR | No public pricing, requires sales call |
| LemonSqueezy | ✅ | ~5% + $0.50 | Full MoR | **Chosen** |

---

## Decision

Use **LemonSqueezy** as the payment processor and Merchant of Record.

LemonSqueezy is the legal seller on every transaction. Customers receive invoices from LemonSqueezy, not checkmeup. LemonSqueezy collects and remits all applicable taxes globally. checkmeup receives net payouts.

---

## Israeli tax implications

The founder operates as **עוסק פטור** (VAT-exempt self-employed individual). Key points:

- **Revenue from LemonSqueezy = export of services to a foreign company** → zero-rated for Israeli VAT (פטור). No VAT charged, no VAT remitted.
- **עוסק פטור threshold (~120K NIS/year)** applies only to business turnover, not salary from primary employment. The two do not overlap.
- **Income tax**: business profit from LemonSqueezy payouts is added to total annual income and taxed at the marginal rate via the annual return (דוח שנתי). Quarterly advance payments (מקדמות) apply once the tax authority sets them.
- **Bookkeeping**: iCount or GreenInvoice for income tracking and annual report preparation.

---

## Integration approach

- LemonSqueezy webhook (`order_created`, `subscription_updated`, `subscription_cancelled`) → update `orgs.plan` in DB
- Billing page in app: current plan + usage bars + "Upgrade" CTA linking to LemonSqueezy checkout URL
- Customer Portal: LemonSqueezy's hosted portal for plan changes and cancellation (no custom UI needed)
- No payment UI built in-app — all card handling on LemonSqueezy's side

---

## Consequences

- Zero tax registration burden outside Israel
- Slightly higher transaction fees than Stripe (~2% difference) — acceptable until meaningful revenue
- Customer invoices show "LemonSqueezy" — not an issue for self-serve developer audience at these price points
- If revenue grows significantly or enterprise buyers require branded invoices, migrate to Paddle (which offers white-label invoicing on higher plans)
