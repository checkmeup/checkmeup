---
title: Billing & Plan-Limit Enforcement
type: knowledge
status: current
updated: 2026-07-12
tags: [architecture, billing, paddle, backend]
scope: apps/api/internal/billing, internal/handler/billing.go
superseded_by:
---

# Billing & plan-limit enforcement

**Investigated:** 2026-07-10
**Scope:** how a plan's limits are defined, checked at write time, and re-enforced on downgrade; the Paddle checkout/webhook flow that changes an org's plan.

## Summary

Plan limits are one static Go map (`internal/billing/plans.go`), checked two ways: **at creation time** (reject a new monitor/status page/channel that would exceed the limit) and **at downgrade time** (`internal/billing/enforce.go` pauses/disables the newest excess resources so an org is never running more than its new plan allows, without deleting anything). Paddle is the merchant of record ([ADR-026](../decisions/026-billing-paddle-mor.md)) — `internal/handler/billing.go` is the only place that talks to Paddle's API or verifies its webhook signature.

## Findings

1. **One map, six limit dimensions, one escape hatch.** `planLimits` (`plans.go:17-22`) is a `map[db.Plan]billing.Limits` with `MonitorTotal`, `StatusPages`, `NotificationChannels`, `MinIntervalMins`, `SMSCredits`, `HideBrandingAllowed` per plan (Hobby/Solo/Startup/Enterprise — [ADR-019](../decisions/019-plan-limits.md)). `-1` means unlimited (not used by any current plan). `HideBrandingAllowed` is the one boolean gate in the set — see finding 7 — everything else is a count. `GetLimits` falls back to Hobby's limits for an unrecognized `db.Plan` value rather than erroring — the safe-by-default direction if plan data is ever malformed.

2. **Four `Check*Limit` functions, same shape**: `CheckMonitorLimit`/`CheckStatusPageLimit`/`CheckNotificationChannelLimit` all take `(plan, current int)` and return a sentinel error (`ErrMonitorLimit` etc., "...upgrade to add more") if `current >= limit` and the limit isn't `-1`. Called from every monitor-type create handler (`monitors.go`, `uptime_monitors.go`, `ssl_monitors.go`, `port_monitors.go`, `domain_monitors.go`), `status_pages.go`, and `notification_channels.go` (twice — once for creating a channel, once for re-enabling a disabled one) right after counting the org's current resources, before the DB insert. `notification_channels.go` also gates SMS specifically: `GetLimits(plan).SMSCredits <= 0` blocks creating or testing an SMS channel on a plan with no SMS allotment (Hobby), separately from the channel-count limit.

3. **`ClampInterval` is the one non-boolean limit** — instead of rejecting, `uptime_monitors.go`/`port_monitors.go` call `billing.ClampInterval(plan, requestedMins)` to silently round a too-frequent check interval up to the plan's `MinIntervalMins`, returning an error only alongside the clamped value (caller can choose to surface it or just use the clamped interval).

4. **Downgrade enforcement is separate from creation-time checks and only runs from the Paddle webhook.** `EnforceMonitorLimit`/`EnforceNotificationChannelLimit`/`EnforceHideBrandingLimit` (`enforce.go`) don't delete anything — the first two *pause* the newest active monitors (across all 5 types, ordered by `created_at DESC`, so the oldest `limit` stay active) or *disable* the newest enabled channels, leaving the org's actual data intact for if they upgrade again; the third just clears `hide_branding` back to `false` (see finding 7). All three are idempotent (no-op if already at or under the limit / already false) and only called from `BillingHandler.Webhook` (`billing.go:446,449,452`) right after a plan change is persisted — there's no separate cron sweep; enforcement only fires at the moment Paddle confirms the downgrade.

5. **Paddle webhook signature verification gates everything.** `verifyPaddleSignature` (`billing.go:642`) computes HMAC-SHA256 over `"{timestamp}:{rawBody}"` using `PADDLE_WEBHOOK_SECRET` and compares with `hmac.Equal` (constant-time); `Webhook` rejects with no secret configured or a bad signature before touching the payload. `resolveOrgPlanUpdate` (`billing.go:334`) turns the verified webhook body into the plan/cycle/status/subscriptionID/customerID/renewsAt tuple written to the org row — this is the only code path that changes `orgs.plan` outside of manual DB access.

6. **Checkout vs. change-plan are two different Paddle flows**, both requiring `PADDLE_API_KEY` configured (checked before resolving a price ID, so an unconfigured account gets "not configured" rather than a misleading "invalid plan"):
   - `CreateCheckout` (Hobby → first paid plan): resolves a price ID, calls `createPaddleTransaction`, returns a `transactionId` for the frontend's Paddle.js overlay to render. No subscription exists yet, so there's nothing to modify server-side.
   - `ChangePlan` (upgrade/downgrade/cancel between paid tiers): requires an existing `PaddleSubscriptionID` on the org; downgrading to Hobby calls `cancelPaddleSubscription` (Paddle's `subscription.canceled` webhook fires at period end, not immediately — see the comment at `billing.go:588`), anything else calls `updatePaddleSubscription` with the new price ID. The actual `orgs.plan` change (and downgrade enforcement) happens later, when the resulting webhook arrives — `ChangePlan` itself never writes the plan directly.
   - `priceIDForPlan` (`billing.go:456`) resolves plan+cycle to one of the `PADDLE_*_PRICE_ID` env vars ([`docs/reference/billing-setup.md`](../reference/billing-setup.md)); an unset price ID for a valid plan/cycle combination returns `""`, surfaced as "this plan isn't available yet" rather than a hard error.

7. **`HideBrandingAllowed` is a boolean gate, not a count, so it gets its own shape rather than reusing `Check*Limit`.** `status_pages.go`'s `UpdateStatusPage` checks `billing.GetLimits(plan).HideBrandingAllowed` inline (only when the request tries to set `hideBranding: true`) rather than calling a `Check*Limit` function, since there's no "current count" to compare against — just allowed or not, per [ADR-035](../decisions/035-status-page-hide-branding.md). There's no separate "re-enable" endpoint the way monitors/channels have (`Resume<Type>Monitor`, `UpdateNotificationChannel`) — the same `UpdateStatusPage` handler is the only write path, so one inline check covers it.

## Follow-ups

- SMS credit *consumption* (as opposed to the channel-creation gate here) lives in `internal/worker.consumeSMSCredit` — see [notification-channels.md](notification-channels.md) finding 5.
