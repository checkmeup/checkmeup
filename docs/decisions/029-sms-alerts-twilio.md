# ADR-029: SMS alerts — Twilio, opt-in via explicit checkbox + consent record

**Date:** 2026-07-04
**Status:** Accepted

---

## Context

[EP-19](../stories/ep-19-sms-alerts.md) adds SMS as an eighth alert channel. Per the [decision backlog](backlog.md), this needed picking a provider (Twilio, Vonage, AWS SNS) and confirming an opt-in flow that satisfies TCPA-style anti-spam regulation before US-1901 starts.

Every existing external integration in this codebase is a simple, synchronous REST API authenticated by a single API key — Resend (email, [ADR-012](012-email-resend.md)), the Telegram Bot API, Paddle ([ADR-026](026-billing-paddle-mor.md)). None require a cloud infrastructure account beyond the API key itself, consistent with the single-Hetzner-VPS, no-broker architecture ([ADR-001](001-worker-model.md), [ADR-006](006-infrastructure-hetzner-kamal-traefik.md)).

## Decision

**Provider: Twilio.**

- Fits the existing "one API key, one HTTP call" integration pattern. AWS SNS would be the odd one out — it requires an AWS account, IAM credentials, and region configuration, none of which anything else in this codebase touches, purely to send a text message.
- Twilio's Messaging Service still honors STOP/START/HELP keywords automatically at the carrier layer when the sender is a real two-way number (Advanced Opt-Out) — but checkmeup doesn't rely on this for compliance, since alphanumeric sender IDs (usable in many non-US destinations, see [`docs/reference/twilio-setup.md`](../reference/twilio-setup.md)) can't receive replies at all. The actual opt-out mechanism is the existing in-app notification-channel toggle ([EP-28](../stories/ep-28-notification-channels.md)) — works identically regardless of sender type or destination.
- US 10DLC / Toll-Free Verification campaign registration (required for US traffic) is a one-time, founder-done setup step in the Twilio console — an ops task for a future `docs/billing-setup.md`-style checklist, not a code dependency, and not part of this ADR.
- Vonage was the closest alternative — similar pricing and API shape — but has less mature compliance tooling (10DLC registration flow) and a smaller ecosystem. No codebase-specific reason favors it over Twilio.

**Compliance / opt-in flow (US-1901):**

- A dedicated checkbox — "I agree to receive automated SMS alerts from checkmeup at this number" — must be checked before a phone number is saved. Providing a number is not itself consent; TCPA-style rules require prior express consent for automated texts, even informational ones.
- Consent is recorded with a timestamp (`sms_consent_at TIMESTAMPTZ`) next to the phone number — evidence of consent, not just a boolean flag.
- Opt-out is handled entirely in-app: disabling the `sms` notification channel (or the org-level SMS toggle) stops sends immediately, same as any other channel — this is the sole opt-out mechanism checkmeup relies on for compliance. If a recipient replies STOP anyway on a two-way-capable sender, Twilio still honors that automatically at the carrier level (see above); checkmeup doesn't parse or store that event specifically — a subsequent send to that number just comes back as an ordinary delivery failure through the existing status-callback webhook (US-1904), same as a carrier rejection or invalid number, so checkmeup still stops attempting (and paying for) blocked sends without any opt-out-specific code path.
- This is transactional/informational alerting ("your monitor is down"), not marketing, which carries a lower regulatory bar than promotional SMS — but the explicit-checkbox-plus-timestamp pattern above is deliberately conservative rather than assuming the lower bar applies in every market checkmeup has users in.

## Consequences

- New env vars: `TWILIO_ACCOUNT_SID`, `TWILIO_API_KEY_SID`, `TWILIO_API_KEY_SECRET`, `TWILIO_MESSAGING_SERVICE_SID` — same pattern as `RESEND_API_KEY` / `TELEGRAM_BOT_TOKEN` in `apps/api/.env`. Twilio recommends authenticating with a scoped API Key (SID + Secret pair) rather than the primary Account Auth Token, so both halves of the key need their own var — the Account SID is still required separately since it's used in the REST resource URL, not for authentication.
- `notification_channels.config` JSONB for an `sms` channel: `{"phone_number": "+972...", "consent_at": "..."}`, consistent with the existing polymorphic shape from [ADR-023](023-notification-channels.md).
- A Twilio account (with billing enabled) must exist before EP-19 ships — a founder-only setup step, same category as the Paddle dashboard checklist in [`docs/reference/billing-setup.md`](../reference/billing-setup.md).
- Per-message cost is real and variable, unlike every other channel shipped so far (Telegram/email/webhook/Slack are free to send). US-1904's `max_alerts_per_incident` cap is what bounds this — same mechanism as every other channel, no separate SMS-specific budget cap needed.
- Removes the "SMS provider + compliance" entry from the [decision backlog](backlog.md). EP-19 is unblocked.
