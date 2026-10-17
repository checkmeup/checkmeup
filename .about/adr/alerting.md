# Alerting: notification channels, delivery providers and alert limits

Date: 2026-10-17

A notification channel is one configured destination (a Telegram chat, an email address, a webhook, a Slack webhook, a phone number), owned by the org and attached to any number of monitors through a join table. There is no group of channels; a monitor attaches to channels directly. A monitor's `alerts_enabled` stays a master switch, so muting doesn't mean detaching channels. A Slack or webhook channel's URL is bearer-like, since whoever holds it can post to it, so after creation the API returns it masked (host plus the last few characters), and an update that leaves it empty keeps the stored one. A webhook channel's signing secret stays visible in Settings, so the user can verify `X-Checkmeup-Signature` on their side. A Telegram chat ID, an email address and a phone number are returned as stored: none of them lets anyone send on its own.

A monitor always has somewhere to alert: when it has no enabled channel attached, or nothing it tried was delivered, the alert goes to the email of every user in the org. This applied to existing orgs too, on purpose: a monitor going down with nobody told is worse than an unasked-for email.

Each channel gets one delivery attempt per alert, never a retry: retrying against a receiver that is already failing adds load when it hurts most, and the fallback email covers a failed send. Webhook, Slack and SMS record the outcome on the channel, so Settings shows the last delivery.

An SMS Twilio accepts can still fail at the carrier (an unreachable number, a STOP opt-out), so its outcome is final only when its status callback says so: a failure is recorded on the channel then, and when the SMS was the alert's only delivery, the fallback email goes out then. A dispatch sends SMS last, so that is known when the SMS is sent and stored with it; an SMS waits in `sms_messages` for its callback for up to 7 days. Twilio signs callbacks only with the account's primary Auth Token, which this app doesn't hold, so the callback URL carries a secret token derived from `JWT_SECRET` instead.

An alert for a monitor in a project starts its headline with the project name in brackets, `[Acme] nightly-backup missed its ping`, on Telegram, email (body and subject), Slack and SMS; the webhook event adds `projectId` and `projectName` and keeps every existing field, so receivers built before projects keep working (ADR projects). Without a project the alert is unchanged.

Each monitor's `max_alerts_per_incident` caps alerts per incident (default 3; 0 means every failed check alerts); the count lives on the incident and starts at zero with the next one. A recovery alert is always sent. The cap is per monitor because urgency differs per monitor.

Email goes through Resend, from `noreply@checkmeup.net`; human mailboxes stay on Zoho. SMS goes through Twilio, authenticated with a scoped API key. A phone number is saved only with an explicit consent checkbox, recorded with its time; opting out is disabling the channel in the app. SMS templates stay plain ASCII and within one 160-character segment, since one Unicode character halves the limit and doubles the cost. A send the org's SMS credits can't cover skips that channel only; other channels still fire. A test SMS doesn't use credits; a separate per-org limit (10 a minute, 10 an hour, 20 a day) bounds its cost instead.

A missing provider key (Resend, Telegram, Twilio) turns that channel off with a log warning; the app still runs.

## Considered options

- One Telegram chat and one email per org (until 2026-06-20): replaced by channels so a monitor can route to several destinations of any type.
- Returning Slack and webhook URLs as stored (until 2026-10-14): any leak of an API response (XSS, a shared screen, a future team member) would hand out a working way to post to the user's channel.
- A channel group entity: neither healthchecks.io nor UptimeRobot needs one; no demand.
- Viber: since February 2024 Viber bots require a commercial agreement with Rakuten, so there is no self-serve way to get a bot token. Reconsider only if self-serve bots return.
- Signal (2026-10-14): Signal has no bot or business API; the only path is self-hosting `signal-cli` on a phone number registered as a bot, which Signal's anti-abuse systems may ban, and an always-on process ADR workers avoids. Demand is unconfirmed. Reconsider if Signal opens an official API or users keep asking.
- Microsoft Teams (2026-10-14): building and testing it needs a Microsoft 365 business tenant, and Microsoft is retiring the Office 365 Connector incoming webhook for Power Automate workflows, so the mechanism itself is moving. Revisit if users ask for it.
- Verifying `X-Twilio-Signature` on SMS status callbacks: it needs the primary Auth Token in production, which the scoped API key exists to avoid.
- AWS SNS for SMS: needs an AWS account and IAM for one call, unlike every other integration here (one API key, one HTTP call).
- Vonage for SMS: similar pricing, weaker 10DLC registration tooling.
- Postmark, AWS SES or Zoho SMTP for email: Postmark's free tier is too small, SES needs AWS, Zoho's shared IP hurts deliverability.

## Consequences

- Loading a monitor's channels is one more query per alert. Revisit if monitor counts make it matter.
- SMS is the only channel that costs money per message; the per-incident cap is what bounds it.
- Editing a Slack or webhook channel can't send a test until the URL is pasted again: the test needs the real URL, and the client only holds the masked one.
- Rotating `JWT_SECRET` changes the callback token, so callbacks for SMS already sent get a 404 and those messages age out of `sms_messages` unresolved.
- US SMS traffic needs a one-time 10DLC / toll-free registration in the Twilio console (`docs/reference/twilio-setup.md`).
