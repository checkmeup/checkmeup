# Activating billing (LemonSqueezy)

All the code is ready (EP-07, EP-27) — checkout creation, the webhook handler, plan-limit enforcement, the inline upgrade prompt, and the Billing page's upgrade buttons all work today and have been verified against simulated LemonSqueezy requests. What's left is account setup in the LemonSqueezy dashboard, which only the account holder can do (business/tax identity, payout banking details — see [ADR-018](decisions/018-billing-lemonsqueezy-mor.md)).

## Checklist

1. **Create a LemonSqueezy store** (if not already done) at [lemonsqueezy.com](https://lemonsqueezy.com) — business details, payout method, tax info.
2. **Create 6 variants** — one per paid plan × billing cycle. Either as 3 products (Solo/Startup/Enterprise) each with 2 variants (Monthly/Annual), or 6 standalone products — either works, the app only cares about variant IDs:

   | Plan | Monthly | Annual |
   |---|---|---|
   | Solo | $9/mo | $90/yr |
   | Startup | $29/mo | $290/yr |
   | Enterprise | $99/mo | $990/yr |

   Annual prices are already set to exactly 10× monthly (2 months free) — keep them that way unless the pricing model changes (see [ADR-019](decisions/019-plan-limits.md), [EP-27](stories/ep-27-annual-billing.md)).
3. **Get the API key** — Settings → API, create a key with store read/write access.
4. **Get the Store ID** — shown in Settings → Stores.
5. **Set up the webhook** — Settings → Webhooks → add `https://checkmeup.net/webhook/lemonsqueezy`, subscribe to `subscription_created`, `subscription_updated`, `subscription_cancelled`, `subscription_expired`. Copy the signing secret.
6. **Set env vars** in production (`apps/api/.env` / Kamal secrets):

   ```
   LS_API_KEY=...
   LS_STORE_ID=...
   LS_WEBHOOK_SECRET=...
   LS_SOLO_VARIANT_ID=...
   LS_STARTUP_VARIANT_ID=...
   LS_ENTERPRISE_VARIANT_ID=...
   LS_SOLO_ANNUAL_VARIANT_ID=...
   LS_STARTUP_ANNUAL_VARIANT_ID=...
   LS_ENTERPRISE_ANNUAL_VARIANT_ID=...
   ```

7. **Deploy and smoke test**: sign up a throwaway account, hit a plan limit (confirm the inline upgrade prompt appears), click upgrade from the Billing page, complete a real low-value test purchase if LemonSqueezy's test mode is available, confirm the webhook flips the org's plan and `billing_cycle`, then cancel and confirm it reverts to Hobby.

## What's already handled in code, so you don't need to configure it manually

- **Success redirect** — checkout requests include an explicit `product_options.redirect_url` back to `/billing?upgraded=true`. No dashboard-level redirect setting needed.
- **Failed payments** — handled natively by LemonSqueezy's hosted checkout page (the user stays there and can retry); nothing for us to configure.
- **Plan downgrades/cancellations** — handled via LemonSqueezy's Customer Portal (the "Manage subscription" link on the Billing page), not a custom UI.
