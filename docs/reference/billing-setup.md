---
title: Activating Billing (Paddle)
type: reference
status: active
updated: 2026-07-05
tags: [billing, paddle, ops]
---

# Activating billing (Paddle)

All the code is ready (EP-07, EP-27, [ADR-026](../decisions/026-billing-paddle-mor.md)) — checkout creation, the webhook handler, plan-limit enforcement, the inline upgrade prompt, and the Billing page's upgrade buttons all work today. What's left is account setup in the Paddle dashboard, which only the account holder can do (business/tax identity — see [ADR-026](../decisions/026-billing-paddle-mor.md)).

## Checklist

1. **Create a Paddle account** — business type: Individual/sole trader (matches עוסק פטור, no registered company). Confirm checkmeup fits Paddle's accepted categories (it does — B2B SaaS is explicitly listed; see ADR-026).
2. **Create 3 products** (Solo, Startup, Enterprise) — tax category: SaaS. Product icon: `https://checkmeup.net/favicon.svg`.
3. **Create 6 prices** — one per plan × billing cycle:

   | Plan | Monthly | Annual |
   |---|---|---|
   | Solo | $9/mo | $90/yr |
   | Startup | $29/mo | $290/yr |
   | Enterprise | $99/mo | $990/yr |

   Annual prices are exactly 10× monthly (2 months free) — keep them that way unless the pricing model changes (see [ADR-019](../decisions/019-plan-limits.md), [EP-27](../stories/ep-27-annual-billing.md)). Quantity: min 1, max 1 (not a per-seat product). Custom data per price: `plan`/`cycle` (e.g. `plan: solo`, `cycle: monthly`) — not required by the code, but makes every webhook self-describing.
4. **Set the default payment link** to `https://checkmeup.net/billing` (Checkout > Checkout settings) — this is the URL Paddle Checkout is approved to run from. Submit `checkmeup.net` for domain approval if not already done (the same domain-verification gate that prompted the Terms/Privacy/Refund Policy overhaul — see the v1.14 release notes).
5. **Create an API key** (Authentication > API keys) — non-expiring, scopes: `Transactions` (read + write), `Customer portal sessions` (write), and `Subscriptions` (read + write — needed for in-app plan downgrades/cancellations, which call Paddle's subscription update/cancel APIs directly). This becomes `PADDLE_API_KEY`.
6. **Create a client-side token** (Authentication > Client-side tokens) — this is public, safe to expose in the browser. Becomes `VITE_PADDLE_CLIENT_TOKEN` on the frontend.
7. **Set up the webhook** — destination URL `https://checkmeup.net/webhook/paddle`, subscribed to `subscription.created`, `subscription.updated`, `subscription.canceled`. Grab the endpoint's signing secret → `PADDLE_WEBHOOK_SECRET`.
8. **Set env vars**:
   - Backend (`.env`): `PADDLE_ENVIRONMENT`, `PADDLE_API_KEY`, `PADDLE_WEBHOOK_SECRET`, `PADDLE_SOLO_PRICE_ID`, `PADDLE_STARTUP_PRICE_ID`, `PADDLE_ENTERPRISE_PRICE_ID`, `PADDLE_SOLO_ANNUAL_PRICE_ID`, `PADDLE_STARTUP_ANNUAL_PRICE_ID`, `PADDLE_ENTERPRISE_ANNUAL_PRICE_ID`.
   - Frontend (`apps/web/.env`, build-time): `VITE_PADDLE_CLIENT_TOKEN`, `VITE_PADDLE_ENVIRONMENT`.
9. **Sandbox smoke test** (safe — no real charges in Sandbox mode): sign up a throwaway account on prod, hit a plan limit (confirm the inline upgrade prompt appears), click upgrade from the Billing page with a test card (see below), confirm the Paddle Checkout overlay opens and the webhook flips the org's `plan`/`billing_cycle`, then cancel via the customer portal (Billing page's "Manage subscription" link) and confirm it reverts to Hobby. Check Paddle's webhook delivery log for `200`s along the way.

   Real card numbers don't work against a Sandbox account — use Paddle's test cards instead ([source](https://developer.paddle.com/concepts/payment-methods/card/)). Any cardholder name, any future expiry, any CVC:

   | Card number | Result |
   |---|---|
   | `4242 4242 4242 4242` | Success, no 3DS |
   | `4000 0038 0000 0446` | Success, with 3DS challenge |
   | `4000 0566 5566 5556` | Success, Visa debit |
   | `4000 0000 0000 0002` | Declined |
   | `4000 0027 6000 3184` | Succeeds once, then declines on a later charge — useful for testing renewal/dunning failures |

10. **Switch to Live mode** in the Paddle dashboard once step 9 passes, before announcing paid plans publicly. Live mode needs its own API key/client-side token/price IDs — re-check every env var in step 8 still matches after switching, don't assume sandbox and live share IDs.

## Local dev (sandbox) vs. production (live)

Paddle's sandbox and production are fully separate environments — separate API keys, client-side tokens, price IDs, and API host (`sandbox-api.paddle.com` vs `api.paddle.com`, selected by `PADDLE_ENVIRONMENT`/`VITE_PADDLE_ENVIRONMENT`). A sandbox key against the production host (or vice versa) just fails.

| | `apps/api/.env` (backend, local dev) | `apps/web/.env` (frontend, local dev) | repo-root `.env` (Kamal, production) |
|---|---|---|---|
| Environment | `PADDLE_ENVIRONMENT=sandbox` | `VITE_PADDLE_ENVIRONMENT=sandbox` | omit `PADDLE_ENVIRONMENT`/`VITE_PADDLE_ENVIRONMENT` — both default to `production` |
| Credentials | Sandbox API key, webhook secret, price IDs (from Paddle's Sandbox account) | Sandbox client-side token | Live API key, webhook secret, price IDs, client-side token |

`apps/api/.env.example` and `apps/web/.env.example` already default to sandbox — copy them as-is for local dev. For production, the repo-root `.env` (Kamal secrets, gitignored) should either leave the environment vars out entirely or set them explicitly to `production`; [`config/deploy.yml`](../../config/deploy.yml) reads the Paddle secrets into the running container (`env.secret`) and the client-side token/environment into the frontend build (`builder.args`, since Vite only bakes `VITE_*` vars in at build time).

## What's already handled in code, so you don't need to configure it manually

- **Success redirect** — Paddle.js's `Checkout.open()` call includes an explicit `settings.successUrl` back to `/billing?upgraded=true`. On that redirect, `BillingView.vue` polls `GET /api/v1/billing` for a few seconds waiting for the plan to actually change, since Paddle's webhook that persists the new plan lands asynchronously and can trail the browser redirect by a few seconds.
- **Checkout theme** — the overlay automatically matches whichever light/dark theme the user is currently in.
- **Failed payments** — handled natively by Paddle's checkout overlay (the user stays there and can retry); nothing for us to configure.
- **Plan downgrades between paid tiers, and cancellation down to Hobby** — an in-app "Downgrade" section on the Billing page, backed by `POST /api/v1/billing/change-plan`. Paid-to-paid changes call Paddle's subscription update API directly (`prorated_immediately`); dropping to Hobby schedules a cancellation for the end of the current billing period via Paddle's subscription cancel API — the org keeps paid-tier access until then. The Customer Portal link ("Manage subscription") still exists for anything else (payment method, invoices).
