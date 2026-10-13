# Alerting: notification channels, delivery providers and alert limits

Date: 2026-10-13

A notification channel is one configured destination (a Telegram chat, an email address, a webhook, a Slack webhook, a phone number), owned by the org and attached to any number of monitors through a join table. There is no group of channels; a monitor attaches to channels directly. A monitor's `alerts_enabled` stays a master switch, so muting doesn't mean detaching channels. A webhook channel's signing secret stays visible in Settings, so the user can verify `X-Checkmeup-Signature` on their side; whether bearer-like secrets such as a Slack webhook URL should be masked is open (see the Inbox).

A monitor always has somewhere to alert: when it has no enabled channel attached, or nothing it tried was delivered, the alert goes to the email of every user in the org. This applied to existing orgs too, on purpose: a monitor going down with nobody told is worse than an unasked-for email.

Each monitor's `max_alerts_per_incident` caps alerts per incident (default 3; 0 means every failed check alerts); the count lives on the incident and starts at zero with the next one. A recovery alert is always sent. The cap is per monitor because urgency differs per monitor.

Email goes through Resend, from `noreply@checkmeup.net`; human mailboxes stay on Zoho. SMS goes through Twilio, authenticated with a scoped API key. A phone number is saved only with an explicit consent checkbox, recorded with its time; opting out is disabling the channel in the app. SMS templates stay plain ASCII and within one 160-character segment, since one Unicode character halves the limit and doubles the cost. A send the org's SMS credits can't cover skips that channel only; other channels still fire. A test SMS doesn't use credits; a separate per-org limit (10 a minute, 10 an hour, 20 a day) bounds its cost instead.

A missing provider key (Resend, Telegram, Twilio) turns that channel off with a log warning; the app still runs.

## Considered options

- One Telegram chat and one email per org (until 2026-06-20): replaced by channels so a monitor can route to several destinations of any type.
- A channel group entity: neither healthchecks.io nor UptimeRobot needs one; no demand.
- Viber: since February 2024 Viber bots require a commercial agreement with Rakuten, so there is no self-serve way to get a bot token. Reconsider only if self-serve bots return.
- AWS SNS for SMS: needs an AWS account and IAM for one call, unlike every other integration here (one API key, one HTTP call).
- Vonage for SMS: similar pricing, weaker 10DLC registration tooling.
- Postmark, AWS SES or Zoho SMTP for email: Postmark's free tier is too small, SES needs AWS, Zoho's shared IP hurts deliverability.

## Consequences

- Loading a monitor's channels is one more query per alert. Revisit if monitor counts make it matter.
- SMS is the only channel that costs money per message; the per-incident cap is what bounds it.
- US SMS traffic needs a one-time 10DLC / toll-free registration in the Twilio console (`docs/reference/twilio-setup.md`).
