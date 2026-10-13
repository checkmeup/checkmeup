# Annual billing

## Problem

Some customers prefer paying once a year, and a discount for it improves retention.

## Users

Paying customers on Solo, Startup and Enterprise.

## Success signal

Customers can pick annual at checkout and pay ten months' price for twelve.

## Scope

- A monthly/annual switch on the pricing and billing pages; annual prices of ten times the monthly price ($90, $290, $990 a year).
- Checkout for the chosen cycle; the org's cycle tracked from Paddle's webhook.

## Non-goals

- Other cycles or custom terms.

## Acceptance criteria

- The pricing page shows monthly prices by default; switching to annual shows the yearly total and the effective monthly price; Hobby stays $0.
  → test: apps/web/src/views/PricingView.test.ts "does not show annual prices or effective monthly cost by default"
  → test: apps/web/src/views/PricingView.test.ts "switches to annual prices when the Annual toggle is clicked"
  → test: apps/web/src/views/PricingView.test.ts "shows the effective monthly cost for annual billing"
- The billing page shows annual pricing with the effective monthly cost for an annual subscriber.
  → test: apps/web/src/views/BillingView.test.ts "renders annual pricing with the effective monthly cost"
- Checkout passes the chosen cycle, monthly by default; an unknown cycle is refused.
  → test: apps/web/src/api/billing.test.ts "passes an explicit annual billing cycle through"
  → test: apps/web/src/api/billing.test.ts "defaults to a monthly billing cycle"
  → test: apps/api/internal/handler/billing_test.go "TestCreateCheckout/invalid cycle"
- An annual subscriber's renewal date on the billing page is a year after purchase.
