# Billing: Paddle as Merchant of Record, plans and their limits

Date: 2026-10-13

Paddle Billing is the payment processor and the Merchant of Record: it is the seller on every invoice and collects and remits tax worldwide. The founder sells as an Israeli sole trader (עוסק פטור) with no company, so handing global tax to the processor is a hard requirement; Paddle payouts count as an export of services, zero-rated for Israeli VAT. Checkout is Paddle.js's in-page overlay opened with a transaction the backend creates, with the org taken from the session, never from client input; plan changes arrive by webhook. The customer portal is a per-customer session URL made server-side.

Four plans, monthly or annual: Hobby $0, Solo $9, Startup $29, Enterprise $99.

| | Hobby | Solo | Startup | Enterprise |
|---|---|---|---|---|
| Monitors, all types together | 10 | 30 | 100 | 1000 |
| Status pages | 1 | 3 | 10 | 100 |
| Notification channels | 5 | 20 | 50 | 100 |
| Minimum check interval | 5 min | 1 min | 1 min | 1 min |
| SMS credits a month | 0 | 10 | 30 | 100 |
| Hide status-page branding | — | ✓ | ✓ | ✓ |

Monitors are counted across every type, because users think in "how many things am I watching". Keyword checks are on every plan. Limits are enforced by the API (`402 plan_limit_reached`); the UI shows an inline upgrade prompt.

Nothing is deleted on a downgrade. Every plan webhook re-applies the limits: the newest active monitors and notification channels beyond the new limit are paused or disabled (the oldest is what the user set up first), and `hide_branding` is cleared where the plan no longer allows it. Resuming or re-enabling past the limit is refused with the same `402`. Status pages over the limit stay live; only creating more is blocked.

SMS has a real per-message cost, so each plan bundles a monthly credit quota rather than metering it: one credit a send, reset on the 1st of the month, unused credits lost, no top-ups. Running out skips only the SMS channel for that alert (ADR alerting).

## Considered options

- LemonSqueezy (until 2026-07-03): the first Merchant of Record; replaced by Paddle before revenue depended on either.
- Stripe: not a Merchant of Record; the founder would file tax in every jurisdiction.
- Metered SMS billing through Paddle: needs usage metering that doesn't exist, and cuts against the flat, predictable pricing the product sells on.
- Weighting SMS credits by destination (1 credit for cheap countries, 3 for the rest): agreed but not built; it needs a hand-kept per-country price table. The send path takes a credit cost, so adding it is data, not a redesign.
- 30-second check intervals: need seconds instead of minutes in the schema and the worker; deferred.

## Consequences

- Invoices come from Paddle, not Checkmeup.
- Between a plan lapsing and its webhook, an org can briefly run over its limits.
- Credits reset on the calendar month, not the billing anchor, so a plan bought mid-month gets a short first month. Revisit if it becomes a complaint.
