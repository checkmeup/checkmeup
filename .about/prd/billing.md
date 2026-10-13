# Billing: plans, usage and upgrades

## Problem

Checkmeup has to earn money without making a solo developer talk to sales or think about tax, and users need to see what their plan allows before they hit a wall.

## Users

Every org; paying ones on Solo, Startup and Enterprise (ADR billing).

## Success signal

A user upgrades from the billing page in one checkout, and their new limits apply as soon as Paddle confirms it.

## Scope

- A billing page with the plan, price, renewal or end date, and usage against each limit.
- Plan limits enforced by the API with an inline upgrade prompt in the UI.
- Upgrading through Paddle's checkout overlay; managing or cancelling through Paddle's portal.
- Applying plan changes from Paddle's webhook, including downgrades.
- Annual billing (PRD annual-billing).

## Non-goals

- Usage-based or metered pricing (ADR billing).
- Coupons, trials, invoices from Checkmeup.

## Acceptance criteria

- The billing page shows the plan, its price, and usage against the monitor, status page and channel limits; Enterprise shows "unlimited" wording and no upgrade buttons.
  → test: apps/web/src/views/BillingView.test.ts "renders the hobby plan as free with upgrade options for all paid plans"
  → test: apps/web/src/views/BillingView.test.ts "shows unlimited usage labels and hides upgrade options for the top enterprise plan"
- A paying org sees "Manage subscription"; a Hobby org doesn't.
  → test: apps/web/src/views/BillingView.test.ts "renders monthly pricing and the manage subscription link for a paid plan"
  → test: apps/web/src/views/BillingView.test.ts "does not show the manage subscription link for the hobby plan"
- Creating a monitor or status page past the plan's limit, or an interval below its minimum, gets `402 plan_limit_reached`, and the form shows an upgrade prompt.
  → test: apps/api/internal/handler/monitors_test.go "TestCreateCronMonitor/plan limit enforced at 10 monitors on Hobby"
  → test: apps/api/internal/handler/status_pages_test.go "TestCreateStatusPage/plan limit enforced at 1 status page on Hobby"
  → test: apps/api/internal/handler/uptime_monitors_test.go "TestCreateUptimeMonitor/an interval below the plan minimum is rejected, not silently clamped"
  → test: apps/web/src/views/CronMonitorCreateView.test.ts "shows an upgrade prompt when the plan limit is reached"
- Choosing a plan opens Paddle's checkout overlay for that plan; with billing not configured, the page says so.
  → test: apps/web/src/views/BillingView.test.ts "starts checkout and opens the Paddle overlay with the returned transaction"
  → test: apps/web/src/views/BillingView.test.ts "shows a not-configured message when billing has not been activated"
- A signed Paddle webhook for a known price moves the org to that plan; an unsigned or wrongly signed one is refused.
  → test: apps/api/internal/handler/billing_test.go "TestWebhook/known price upgrades the org's plan"
  → test: apps/api/internal/handler/billing_test.go "TestWebhook/invalid signature returns 401"
- A cancelled subscription keeps the plan until its period ends ("Access until"), then drops to Hobby.
  → test: apps/api/internal/handler/billing_test.go "TestWebhook/scheduled cancellation marks status cancel_scheduled without downgrading yet"
  → test: apps/api/internal/handler/billing_test.go "TestWebhook/cancellation downgrades to hobby and clears subscription fields"
  → test: apps/web/src/views/BillingView.test.ts "shows \"Access until\" wording when the subscription is cancelled"
- On a downgrade, the newest active monitors and channels beyond the new limit are paused or disabled and the oldest stay on; hidden branding comes back.
  → test: apps/api/internal/handler/billing_test.go "TestWebhook/downgrade to hobby auto-pauses the newest monitors and disables the newest channels beyond the new limit (ADR billing)"
  → test: apps/api/internal/billing/enforce_test.go "TestEnforceMonitorLimit/pauses the newest active monitors beyond limit, leaves the oldest active"
  → test: apps/api/internal/billing/enforce_test.go "TestEnforceHideBrandingLimit/clears hide_branding across every page in the org when not allowed"
- Resuming a paused monitor past the limit is refused with an upgrade prompt.
  → test: apps/api/internal/handler/monitors_test.go "TestPauseResumeCronMonitor/resume is blocked once active monitors are already at the plan limit (ADR billing)"
  → test: apps/web/src/views/CronMonitorDetailView.test.ts "shows an upgrade prompt instead of a plain error when resume is blocked by the plan limit"
