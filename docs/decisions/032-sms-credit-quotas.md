# ADR-032: SMS credit quotas — destination-weighted, plan-bundled, not metered pass-through

**Date:** 2026-07-04
**Status:** Accepted — refines [ADR-029](029-sms-alerts-twilio.md)

---

## Context

[ADR-029](029-sms-alerts-twilio.md) picked Twilio as the SMS provider for [EP-19](../stories/ep-19-sms-alerts.md) but left the cost model unaddressed. Twilio's per-message price varies by destination country and carrier — a US segment is roughly $0.0079, but many countries and specific carriers price meaningfully higher, and messages over the GSM-7 160-character (or 70-character for Unicode) limit split into multiple billed segments. SMS is the first alert channel with a real, variable, per-message marginal cost — Telegram, email, Slack, and webhook are all free to send. Every existing plan limit (monitors, status pages, notification channels — [ADR-019](019-plan-limits.md)) bounds a *count*, not a *variable cost*, so none of that machinery protects margin here on its own.

Two shapes were considered:

- **Metered pass-through** — charge each org the actual Twilio cost (+ margin) per message sent, via Paddle usage-based billing. Most "fair," but requires building real usage-metering and invoicing plumbing that doesn't exist anywhere in this codebase, and cuts against checkmeup's own positioning — the flat, predictable, non-per-monitor pricing pitch used directly against Cronitor's $2/monitor model (see the [competitor comparison post](../../apps/web/src/blog/posts/checkmeup-vs-competitors.ts)).
- **Bundled credit quota** — each plan includes a fixed number of SMS credits per month, matching the precedent both UptimeRobot and Healthchecks.io already use for SMS/voice.

## Decision

**Bundled, destination-weighted SMS credit quotas per plan. No metered billing, no credit purchases/top-ups on this first cut.**

**Quotas** (add to [ADR-019](019-plan-limits.md)'s limits table):

| | Hobby | Solo | Startup | Enterprise |
|---|---|---|---|---|
| SMS credits / month | 0 | 10 | 30 | 100 |

Hobby gets none, consistent with SMS being a paid-tier-only channel on every competitor reviewed (UptimeRobot, Healthchecks.io both gate SMS/voice to paid plans).

**Destination weighting** — a credit is not "one message," it's "one domestic/low-cost-band segment." Destinations are grouped into two cost bands via a static, hand-maintained lookup table keyed by E.164 calling code (a Go map, same shape as `internal/billing/plans.go`'s existing constants — not a live per-send call to Twilio's pricing API, which would add latency and an external dependency to every alert dispatch):

- **Band 1** (US/Canada + the small set of countries checkmeup's actual user base concentrates in, revisited as the user base grows): **1 credit** per segment.
- **Band 2** (everywhere else): **3 credits** per segment.

This protects margin against the exact variance that motivated this ADR without building real metered billing — a user texting expensive destinations burns through their quota faster, but the bill to checkmeup stays predictable because the quota (not the dollar cost) is what's fixed per plan.

**Segment cap stays at 1** — US-1902/US-1903's existing acceptance criteria (keep the message within a single 160-character GSM-7 segment) remain as-is and are now also the mechanism that keeps credit consumption predictable per message. Alert templates must stay plain-ASCII; a Unicode character anywhere in the message drops the single-segment limit to 70 characters and risks a silent 2-segment (2-credit) charge instead of 1.

**Exhaustion behavior** — when an org's monthly credit balance is insufficient for a given send, the SMS channel is skipped for that alert (logged as a delivery failure, same visibility pattern as [US-1804](../stories/ep-18-teams-alerts.md)/[US-1904](../stories/ep-19-sms-alerts.md)), while every other enabled channel for that monitor still fires normally. This is **not** a `402`-style hard block ([ADR-019](019-plan-limits.md)'s existing pattern for monitors/status pages/channels) — a monitor going down must never fail to alert entirely just because one channel ran out of quota for the month.

If SMS was the *only* enabled channel attached to the monitor (or every other attached channel also failed to deliver for the same incident), fall back to the org's account email(s) — the same `dispatchFallbackEmail` mechanism `worker.go`'s `DispatchAlert` already uses when a monitor has zero attached channels at all. Concretely: `DispatchAlert` should fall back to email whenever nothing it tried actually delivered (`sent == false` after the loop), not only when `len(channels) == 0` as today — credit exhaustion is just one more reason a channel can fail to deliver. This keeps the existing "a monitor always has somewhere to send an alert" guarantee intact even when its one configured channel silently can't fire that month.

**Reset** — credits reset on the 1st of each calendar month, independent of each org's Paddle billing-cycle anchor. Deliberately decoupled from billing-cycle dates for simplicity (no new integration surface with Paddle's subscription-period data); revisit only if this becomes a real support complaint (e.g. a plan purchased mid-month getting a "short" first partial period feels off). Unused credits do not roll over.

**Not built in this first cut** — buying additional credits mid-month, or metered overage billing past the quota. If real demand for either shows up, it's a follow-on, not assumed here.

**Implementation note (simplified first pass):** what actually shipped is a flat **1 credit per send**, with no destination weighting yet — the Band 1/Band 2 cost-band table above needs hand-built per-country pricing data (Twilio's public pricing pages), which is a data-entry task deferred separately from the code. The quota numbers, column names, `smsCreditsUsed`/`smsCreditsLimit` field names, lazy-reset-at-send-time design, and skip-only-this-channel exhaustion/email-fallback behavior described above are all implemented exactly as specified — only the per-destination weighting is pending. `internal/worker.ConsumeSMSCredit` (query) and `sendSMSAlert`'s `consumeSMSCredit` call are written so that adding weighting later is a matter of passing a computed credit cost instead of a hardcoded `1`, not a redesign.

## Consequences

- New columns: `orgs.sms_credits_used_this_month INT NOT NULL DEFAULT 0`, `orgs.sms_credits_reset_at DATE NOT NULL`. Reset is lazy — checked and applied at send-time (if `now() >= sms_credits_reset_at`, zero the counter and advance the reset date) rather than via a new scheduled job, consistent with [ADR-001](001-worker-model.md)'s no-broker/no-extra-scheduler stance.
- The destination cost-band table is static and hand-maintained, not sourced live from Twilio — it will drift from Twilio's actual pricing over time and needs periodic manual review (no automated freshness check exists or is planned for MVP).
- `GET /api/v1/billing` gains `smsCreditsUsed` / `smsCreditsLimit`, alongside the existing monitor/status-page/notification-channel usage fields, so Settings can show remaining SMS quota the same way it shows the others.
- EP-19's story list gains credit-quota, weighting, and exhaustion-handling stories (US-1906–US-1908) on top of the original US-1901–US-1905.
- Shipped 2026-07-04: `PricingView.vue` and `HomeView.vue` now list SMS credits per plan tier. The `apps/web/src/blog/posts/checkmeup-vs-competitors.ts` comparison post is deliberately left as-is — not updated as part of this change.
- **"Send test SMS" doesn't consume monthly credits.** A security review flagged this the same day: `TestNotificationChannel`'s sms branch checks the plan gate (`SMSCredits > 0`) but never calls `ConsumeSMSCredit`, so test sends aren't bounded by the monthly quota the way real alerts are. Rather than route test sends through the credit ledger (which would let a user burn their real monthly SMS budget just verifying a channel works), the stopgap shipped 2026-07-04 is a dedicated multi-tier rate limit on the test-send endpoint alone: **10/minute, 10/hour, and 20/day per org** (`smsTestLimiter`/`smsTestHourlyLimiter`/`smsTestDailyLimiter` in `internal/handler/notification_channels.go`). This bounds worst-case Twilio spend from test sends to a fixed daily ceiling independent of the credit system, without eating into the quota a user pays for. Revisit if real abuse is observed — routing test sends through `ConsumeSMSCredit` (at a smaller cost, e.g. no charge or a fractional credit) remains an option for a future pass.
