# EP-19: SMS alerts

An eighth alert channel — unlike Slack/Teams ([EP-17](ep-17-slack-alerts.md)/[EP-18](ep-18-teams-alerts.md)), this needs a real SMS provider (Twilio, Vonage, AWS SNS) and a per-message cost, and is subject to anti-spam regulation (TCPA-style rules in the US and equivalents elsewhere) that requires explicit opt-in before sending automated texts — not just collecting a phone number.

Also builds on the multi-channel model in [EP-28](ep-28-notification-channels.md) ([ADR-023](../decisions/023-notification-channels.md)) — adds an `sms` value to `notification_channel_type`. "Off at the org level" below should read "channel disabled or not attached to that monitor" once EP-28 lands.

Provider and opt-in flow decided in [ADR-029](../decisions/029-sms-alerts-twilio.md): Twilio, with an explicit consent checkbox + timestamp recorded at connect time.

Pricing/cost-control model decided in [ADR-032](../decisions/032-sms-credit-quotas.md): plan-bundled monthly credit quotas (Hobby 0 / Solo 10 / Startup 30 / Enterprise 100 — see [ADR-019](../decisions/019-plan-limits.md)), not metered pass-through billing. US-1906–US-1908 below cover this; it's a distinct concern from US-1904/US-1905's per-incident alert cap, which bounds *volume per incident* regardless of cost, not *monthly spend*.

**Shipped 2026-07-04.** US-1901–US-1908 are all live. One simplification vs. the original design: US-1906's destination-weighted cost bands weren't built — credits are a flat 1-per-send today, since the weighting needs a hand-built per-country pricing table (real data entry, not code) that was deliberately deferred. See ADR-032's "Implementation note" for the full rationale; the storage shape and consumption path were built so weighting can be added later without a redesign. Everything else (opt-in flow, delivery-failure logging, shared alert cap, monthly reset, exhaustion → email fallback) matches the acceptance criteria below as written.

Founder-side Twilio account setup: done except US/Canada Toll-Free/10DLC sender registration, deliberately postponed until the first US-based paid customer (real lead time, no cost to defer). Non-US/Canada destinations use an alphanumeric sender (`CHECKMEUP`), live today. Full checklist: [`docs/twilio-setup.md`](../twilio-setup.md).

---

### US-1901: Connect a phone number for SMS alerts

**As a** user, **I want** to provide a phone number for SMS alerts **so that** I'm notified even without an internet connection.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [x] Org setting: phone number, validated in E.164 format
- [x] Explicit opt-in confirmation required before alerts start (regulatory requirement for automated texts, not just providing a number)
- [x] "Send test SMS" button verifies delivery before saving
- [x] Multiple phone numbers can be added per org, each its own `notification_channels` row (EP-28) — no longer limited to one

---

### US-1902: Receive a down alert via SMS

**As a** user, **I want** a text message when a monitor goes down **so that** I'm notified even with no data connection.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Message kept within a single SMS segment (160 chars) where possible: monitor name, status, short reason
- [x] Sent within one check cycle of the transition to "down"
- [x] Not sent if alerts are disabled for that monitor or SMS alerts are off at the org level

---

### US-1903: Receive a recovery alert via SMS

**As a** user, **I want** a text message when a monitor recovers **so that** I know the incident is resolved.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] Message: monitor name, downtime duration, kept within a single SMS segment where possible
- [x] Always sent on genuine recovery regardless of the per-incident alert cap ([ADR-016](../decisions/016-alert-debounce.md))
- [x] Not sent if alerts are disabled for that monitor or SMS alerts are off at the org level

---

### US-1904: Handle SMS delivery failures and cost control

**As a** user, **I want** SMS failures handled gracefully and my SMS volume kept predictable **so that** a flapping monitor doesn't generate a surprise bill.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Non-success delivery status (carrier rejection, invalid number) logged and visible in Settings — same pattern as webhook delivery status ([US-1404](ep-14-webhook-alerts.md))
- [x] No automatic retries on MVP, consistent with webhook (US-1404)
- [x] SMS strictly respects `max_alerts_per_incident` — no segment-splitting or other workaround that effectively sends more notifications than the cap allows

---

### US-1905: Shared alert cap across all channels

**As a** user, **I want** the per-incident alert cap to include SMS **so that** enabling every channel doesn't multiply my alert volume (or my SMS bill).

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] `max_alerts_per_incident` ([ADR-016](../decisions/016-alert-debounce.md)) counts a triggered alert once per incident-event across all enabled channels together, consistent with [US-1305](ep-13-email-alerts.md), [US-1405](ep-14-webhook-alerts.md), [US-1505](ep-15-whatsapp-alerts.md), [US-1605](ep-16-signal-alerts.md), [US-1705](ep-17-slack-alerts.md), and [US-1805](ep-18-teams-alerts.md)
- [x] Recovery event is exempt from the cap on every enabled channel, including SMS

---

### US-1906: Plan-based SMS credit quota with destination weighting

**As a** platform operator, **I want** each org's SMS usage bounded by a plan-based credit quota that weights expensive destinations more heavily **so that** SMS's variable, destination-dependent cost can't blow past what a flat plan price supports.

**Estimate:** 1.5 h

**Shipped 2026-07-04, simplified** — see ADR-032's "Implementation note": flat 1-credit-per-send, no destination weighting yet (needs a hand-built per-country pricing table, deferred as a data-entry task separate from this code).

**Acceptance criteria:**

- [x] `orgs.sms_credits_used_this_month INT NOT NULL DEFAULT 0` tracks consumption; plan quota (Hobby 0 / Solo 10 / Startup 30 / Enterprise 100, [ADR-032](../decisions/032-sms-credit-quotas.md)) is a Go constant, same pattern as `internal/billing/plans.go`'s existing monitor/status-page/channel limits
- [ ] Destination cost-band lookup (Band 1 = 1 credit, Band 2 = 3 credits) is a static Go map keyed by E.164 calling code — **not built**; every send costs a flat 1 credit regardless of destination for now (`ConsumeSMSCredit`'s `credit_cost` param exists and is ready for this, just always called with `1` today)
- [x] A send that would exceed the remaining monthly balance is not sent — see US-1908 for exactly what happens instead
- [x] `GET /api/v1/billing` returns `smsCreditsUsed` / `smsCreditsLimit` alongside the existing monitor/status-page/notification-channel usage fields

---

### US-1907: Monthly credit reset

**As a** user, **I want** my SMS credit balance to reset every month **so that** unused credits don't accumulate and my quota is predictable.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] `orgs.sms_credits_reset_at DATE` tracks the next reset; reset is applied lazily at send-time (if `now() >= sms_credits_reset_at`, zero the used counter and advance the reset date by one month) rather than via a separate scheduled job, consistent with [ADR-001](../decisions/001-worker-model.md)
- [x] Reset date is the 1st of each calendar month for every org, deliberately independent of that org's Paddle billing-cycle anchor ([ADR-032](../decisions/032-sms-credit-quotas.md))
- [x] Unused credits do not roll over — the counter resets to zero, not to a carried-forward balance

---

### US-1908: Handle SMS credit exhaustion without blocking other channels

**As a** user, **I want** a monitor's other alert channels to still fire even if my SMS credits run out **so that** running out of SMS quota never means missing an alert entirely.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] When the org's remaining monthly SMS credit balance is insufficient for a send, that send is skipped and logged as a delivery failure — same visibility pattern as [US-1804](ep-18-teams-alerts.md)/US-1904 — not a `402`-style hard error on the whole alert dispatch
- [x] Every other enabled, non-SMS channel attached to that monitor still fires normally for the same incident
- [x] If SMS was the only channel attached (or every other attached channel also failed for the same incident), the org's account email is used as a fallback — same mechanism `DispatchAlert` already uses for monitors with zero attached channels ([ADR-032](../decisions/032-sms-credit-quotas.md)), generalized to trigger whenever nothing delivered rather than only when nothing was configured
- [x] Settings shows remaining SMS credits for the current month (Billing page usage bar) — no separate dedicated "upgrade for more" pop-up beyond the existing per-page upgrade CTAs, since exhaustion happens silently in the background rather than as a user-initiated action that can 402
- [x] No mid-month credit purchase/top-up on this first cut — exhausted means wait for next month's reset or upgrade plan tier ([ADR-032](../decisions/032-sms-credit-quotas.md))
