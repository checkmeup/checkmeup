---
title: Activating SMS Alerts (Twilio)
type: reference
status: active
updated: 2026-10-14
tags: [sms, twilio, ops, billing]
---

# Activating SMS alerts (Twilio)

SMS alerts ([PRD sms-alerts](../../.about/prd/sms-alerts.md), [ADR alerting](../../.about/adr/alerting.md), [ADR billing](../../.about/adr/billing.md)) shipped 2026-07-04. This checklist remains the reference for the account-side setup only the founder can do — same category as [`docs/reference/billing-setup.md`](billing-setup.md) for Paddle. Current status: account funded, API Key + Messaging Service configured, alphanumeric sender live for non-US/Canada destinations. Step 4's US/Canada Toll-Free/10DLC registration is the one item still outstanding, deliberately postponed until the first US-based paid customer (real lead time, no cost to defer).

## Checklist

1. **Create a Twilio account and add billing** — a free trial account prepends "Sent from a Twilio trial account" to every message and can only send to phone numbers you've manually verified in the console. Neither is acceptable for a real alert channel; fund the account before testing beyond the very first message.
2. **Create an API Key** (Account > API keys & tokens) rather than using the primary Account SID + Auth Token pair directly — a scoped key can be revoked independently if it ever leaks, same reasoning as this codebase's own API keys ([ADR auth](../../.about/adr/auth.md)). The API Key SID and Secret become `TWILIO_API_KEY_SID` + `TWILIO_API_KEY_SECRET`; `TWILIO_ACCOUNT_SID` is still needed separately (it's part of the REST resource URL, not the auth pair) — see [ADR alerting](../../.about/adr/alerting.md). Do not store anything under a `TWILIO_AUTH_TOKEN` name — that name is reserved for the primary Account Auth Token, which this setup deliberately avoids using.
3. **Create a Messaging Service** (Messaging > Services) — do not send from a bare phone number directly. The Messaging Service is what enables:
   - A single stable `TWILIO_MESSAGING_SERVICE_SID` to send from, regardless of how many underlying senders are attached later.

   Leave the Messaging Service's own status callback empty: each alert SMS sets its own `StatusCallback` (`BASE_URL` + `/webhook/twilio/sms-status/` + a token derived from `JWT_SECRET`), so a message the carrier fails after Twilio accepted it is recorded as failed ([ADR alerting](../../.about/adr/alerting.md)). It's only set when `BASE_URL` is `https://`, so local dev sends without one.

   Note: opt-out compliance does **not** depend on this — checkmeup relies on its own in-app notification-channel toggle for that (see ADR alerting), since alphanumeric senders can't receive STOP replies at all. If a two-way sender happens to receive one anyway, Twilio's Advanced Opt-Out still honors it automatically at the carrier level, and the message's status callback reports it as failed (below).
4. **Register a sender and attach it to the Messaging Service** — this step has real lead time, unlike everything else here. checkmeup registers **two sender types on the same Messaging Service**, since alphanumeric sender IDs are not available in the US or Canada at all — Twilio routes each send to whichever sender fits the destination automatically:

   **US/Canada — Toll-Free Verification:**
   - Buy a toll-free number (Phone Numbers > Buy a number, filter Toll-Free) if you don't already have one.
   - On that number, go to **Regulatory Information** > **Verify this toll free number**.
   - Fill out the three-step form: business/contact info + address; messaging use case (describe transactional monitoring alerts), estimated monthly volume, and a sample message using actual alert copy (e.g. `"checkmeup: monitor 'api-prod' is DOWN since 14:32 UTC"`); and **Privacy Policy** (`https://checkmeup.net/privacy`) + **Terms & Conditions** (`https://checkmeup.net/terms`) URLs — mandatory fields as of Sept 15, 2026, so include them now to avoid a later rejection/resubmission.
   - Submit; Twilio emails approval/rejection, usually within days.
   - Once approved, add the toll-free number to the Messaging Service's Sender Pool.
   - Fall back to full **A2P 10DLC registration** (brand + campaign registration, Console > Messaging > Regulatory Compliance, "Low-Volume Standard" campaign type fits our volume) only if toll-free is rejected or a higher-throughput number is needed — reviews there run 10–15 days, notably slower than toll-free.
   - Sending US-bound SMS through an unregistered sender gets aggressively filtered or blocked by carriers today, not just "sent slower" — don't skip this for US traffic.

   **Everywhere else — Alphanumeric Sender ID:**
   - On the Messaging Service, go to **Senders** > **Add Senders IDs** > select **Alpha Sender**.
   - Enter a string up to 11 characters (`A-Z a-z 0-9 space - + _ &`, not all-numeric) — `CHECKMEUP` fits.
   - No approval wait in most countries; live as soon as added. A handful of destinations still have their own requirements regardless (India, notably, requires separate DLT registration on top) — check the specific country in Twilio Console before assuming this covers it.
5. **Fund the account with a real balance and consider auto-recharge** — sends silently fail once the balance is exhausted; auto-recharge (Billing > Auto-recharge in the console) avoids alerts quietly going dark because the Twilio balance ran dry, which would be a bad failure mode for a monitoring product specifically.
6. **Set env vars** (`apps/api/.env` locally, GitHub Actions secrets in production): `TWILIO_ACCOUNT_SID`, `TWILIO_API_KEY_SID`, `TWILIO_API_KEY_SECRET`, `TWILIO_MESSAGING_SERVICE_SID` — same pattern as `RESEND_API_KEY`/`TELEGRAM_BOT_TOKEN`.

## Outside the Twilio account

- **Destination cost bands** — agreed but not built ([ADR billing](../../.about/adr/billing.md)). When built, the table comes by hand from Twilio's per-country pricing ([twilio.com/en-us/sms/pricing](https://www.twilio.com/en-us/sms/pricing)) and needs review as prices change.
- **TCPA-style consent copy** for the opt-in checkbox (PRD sms-alerts) — exact wording should be reviewed against applicable regulation in the markets checkmeup actually has users in before launch; this is a legal-text question, not a Twilio account setting.

## Sandbox / test numbers

Twilio has no separate "sandbox mode" the way Paddle does — all testing happens against the real account, real (small) per-message cost, using either your own verified number (pre-funding) or the funded account sending to real numbers post-funding. Budget a small amount of real spend for manual testing.
