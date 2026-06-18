# Activating billing (LemonSqueezy)

All the code is ready (EP-07, EP-27) — checkout creation, the webhook handler, plan-limit enforcement, the inline upgrade prompt, and the Billing page's upgrade buttons all work today and have been verified against simulated LemonSqueezy requests. What's left is account setup in the LemonSqueezy dashboard, which only the account holder can do (business/tax identity, payout banking details — see [ADR-018](decisions/018-billing-lemonsqueezy-mor.md)).

## Status

Store, all 6 variants, API key, and webhook secret are configured — **LemonSqueezy is currently in Test mode** (no real charges). Steps 1–6 below are done; what's left is the test-mode smoke test (step 7) and switching to Live mode before announcing paid plans publicly (step 8).

## Checklist

1. ~~**Create a LemonSqueezy store**~~ — done.
2. ~~**Create 6 variants**~~ — done. One per paid plan × billing cycle:

   | Plan | Monthly | Annual |
   |---|---|---|
   | Solo | $9/mo | $90/yr |
   | Startup | $29/mo | $290/yr |
   | Enterprise | $99/mo | $990/yr |

   Annual prices are exactly 10× monthly (2 months free) — keep them that way unless the pricing model changes (see [ADR-019](decisions/019-plan-limits.md), [EP-27](stories/ep-27-annual-billing.md)).
3. ~~**Get the API key**~~ — done (currently a Test mode key).
4. ~~**Get the Store ID**~~ — done.
5. ~~**Set up the webhook**~~ — done. Confirm it's subscribed to `subscription_created`, `subscription_updated`, `subscription_cancelled`, `subscription_expired`.
6. ~~**Set env vars**~~ — done (`LS_API_KEY`, `LS_STORE_ID`, `LS_WEBHOOK_SECRET`, and all 6 `LS_*_VARIANT_ID`).
7. **Test-mode smoke test** (safe — no real charges while Test mode is on): sign up a throwaway account on prod, hit a plan limit (confirm the inline upgrade prompt appears), click upgrade from the Billing page with a test card, confirm the webhook flips the org's `plan`/`billing_cycle` and the success redirect lands on `/billing?upgraded=true`, then cancel via the customer portal and confirm it reverts to Hobby. Check LemonSqueezy's webhook delivery log for `200`s along the way.
8. **Switch to Live mode** in the LemonSqueezy dashboard once step 7 passes, before announcing paid plans publicly. Live mode likely needs its own variant IDs/API key (test and live data don't share IDs on most platforms) — re-check the env vars in step 6 still match after switching, don't assume they carry over automatically.

## What's already handled in code, so you don't need to configure it manually

- **Success redirect** — checkout requests include an explicit `product_options.redirect_url` back to `/billing?upgraded=true`. No dashboard-level redirect setting needed.
- **Failed payments** — handled natively by LemonSqueezy's hosted checkout page (the user stays there and can retry); nothing for us to configure.
- **Plan downgrades/cancellations** — handled via LemonSqueezy's Customer Portal (the "Manage subscription" link on the Billing page), not a custom UI.
