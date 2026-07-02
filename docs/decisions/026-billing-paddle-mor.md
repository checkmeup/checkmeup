# ADR-026: Billing via Paddle (Merchant of Record) — replaces LemonSqueezy

**Date:** 2026-07-03
**Status:** Accepted — supersedes [ADR-018](018-billing-lemonsqueezy-mor.md)

---

## Context

[ADR-018](018-billing-lemonsqueezy-mor.md) chose LemonSqueezy as the payment processor and Merchant of Record, explicitly noting Paddle as a close second — "Good option; LemonSqueezy simpler for MVP stage" — and flagged in its own Consequences section that a future move to Paddle was plausible once billing needed to grow up. That trigger arrived from an unrelated direction: working through a payment processor's domain-verification checklist (which led to the Terms of Service / Privacy Policy / Refund Policy overhaul earlier this cycle) surfaced Paddle as the better fit going forward, and the decision was made to switch before real revenue depends on either provider.

This ADR replaces LemonSqueezy with Paddle entirely — not a dual-provider setup, not an abstraction layer over both. Same constraints as ADR-018 still apply: the founder operates as an Israeli sole proprietor (עוסק פטור) with no registered company, so global tax handling by the payment provider remains a hard requirement.

---

## Decision

Use **Paddle** (Paddle Billing) as the payment processor and Merchant of Record, replacing LemonSqueezy completely. Paddle is the legal seller on every transaction, same as LemonSqueezy was — customers receive invoices from Paddle, not checkmeup, and Paddle collects and remits all applicable taxes globally.

**Business type on Paddle:** Individual / sole trader — matches the עוסק פטור status, no registered company.

**Compliance check performed before committing:** reviewed [Paddle's prohibited/restricted seller categories](https://www.paddle.com/help/start/intro-to-paddle/what-am-i-not-allowed-to-sell-on-paddle) — checkmeup is explicitly within Paddle's accepted "B2B SaaS" category, with none of the prohibited categories (financial services, unauthorized resale, human/consulting services, device repair/optimization tools, VPN/proxy) applicable.

---

## Israeli tax implications

Unchanged from ADR-018 — Paddle is a foreign MoR company exactly as LemonSqueezy was:

- Revenue from Paddle = export of services to a foreign company → zero-rated for Israeli VAT (פטור).
- עוסק פטור threshold applies only to business turnover; doesn't overlap with salary from primary employment.
- Business profit from Paddle payouts is added to total annual income, taxed via the annual return (דוח שנתי).
- Bookkeeping continues via iCount/GreenInvoice.

---

## Integration approach — what actually changes technically

Paddle Billing's architecture differs from LemonSqueezy's in one structural way: **checkout is client-side (Paddle.js overlay), not a server-generated hosted redirect URL.**

| | LemonSqueezy (before) | Paddle (now) |
|---|---|---|
| Catalog | 6 variants (3 plans × 2 cycles) | 3 products, 6 prices (3 plans × 2 cycles) |
| Checkout | Backend calls `POST /v1/checkouts`, returns a hosted URL, frontend redirects (`window.location.href`) | Backend calls `POST /transactions` (price ID + `custom_data.org_id` from the authenticated session), returns a `transactionId`; frontend loads `@paddle/paddle-js`, calls `Paddle.Checkout.open({ transactionId, settings: { theme, successUrl } })` — an inline overlay, no page navigation |
| Webhook | `X-Signature` header, HMAC-SHA256 hex digest of raw body | `Paddle-Signature` header, `ts=...;h1=...` format, HMAC-SHA256 over `ts:body` |
| Webhook events | Generic `data.attributes.status` branching, `event_name` present but unused | Explicit `event_type` (`subscription.created`/`.updated`/`.canceled`, `transaction.completed`) |
| Customer portal | Static URL (`https://app.lemonsqueezy.com/my-orders/`) | Per-customer session URL, generated server-side via `POST /customers/{id}/portal-sessions` |
| DB columns | `orgs.ls_customer_id`, `orgs.ls_subscription_id` | `orgs.paddle_customer_id`, `orgs.paddle_subscription_id` (migration 026) |
| Env vars | `LS_API_KEY`, `LS_STORE_ID`, `LS_WEBHOOK_SECRET`, 6× `LS_*_VARIANT_ID` | `PADDLE_API_KEY`, `PADDLE_WEBHOOK_SECRET`, 6× `PADDLE_*_PRICE_ID` (no store-ID equivalent — the API key is already seller-scoped); frontend also needs `VITE_PADDLE_CLIENT_TOKEN` (public, safe to expose — same role as Stripe's publishable key) |
| Required Paddle API key scopes | — | `Transactions` (read + write), `Customer portal sessions` (write) — least-privilege, no `Products`/`Prices`/`Customers`/`Subscriptions` access needed since price IDs are config-driven and customer/subscription IDs arrive via webhook |

The `custom_data.org_id` trust boundary is preserved exactly as it worked under LemonSqueezy: the backend reads `org_id` from the authenticated session (JWT cookie), not from client input, so a tampered frontend request can't attribute a purchase to the wrong org.

Frontend gains a real dependency it didn't need before (`@paddle/paddle-js`) — this is the one place where "no frontend changes needed" (true for the LemonSqueezy→Paddle swap in general) doesn't hold, because Paddle's checkout model is client-side by design.

---

## Consequences

- Same tax/MoR posture as before — no change to zero tax registration burden outside Israel.
- Comparable fees (~5% + $0.50, same order of magnitude as LemonSqueezy).
- Webhook signature verification, checkout creation, and customer-portal-link generation are full rewrites, not renames — different request/response shapes end to end.
- Checkout becomes an in-page overlay instead of a redirect, which is arguably better UX (no full navigation away from the app) but is a real behavior change worth noting if anything downstream assumed a redirect (nothing in this codebase did).
- LemonSqueezy integration is removed outright, not kept as a fallback — matches how this ADR is scoped (full replacement, not a dual-provider abstraction).
