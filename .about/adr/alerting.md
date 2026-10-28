# Alerting: notification channels, delivery providers and alert limits

Date: 2026-10-28

A notification channel is one configured destination (a Telegram chat, an email address, a webhook, a Slack webhook, a phone number), owned by the org and attached to any number of monitors through a join table. There is no group of channels; a monitor attaches to channels directly. A monitor's `alerts_enabled` stays a master switch, so muting doesn't mean detaching channels. A Slack or webhook channel's URL is bearer-like, since whoever holds it can post to it, so after creation the API returns it masked (host plus the last few characters), and an update that leaves it empty keeps the stored one. A webhook channel's signing secret stays visible in Settings, so the user can verify `X-Checkmeup-Signature` on their side. A Telegram chat ID, an email address and a phone number are returned as stored: none of them lets anyone send on its own.

A monitor always has somewhere to alert: when it has no enabled channel attached, or nothing it tried was delivered, the alert goes to the email of every user in the org. This applied to existing orgs too, on purpose: a monitor going down with nobody told is worse than an unasked-for email.

Each channel gets one delivery attempt per alert, never a retry: retrying against a receiver that is already failing adds load when it hurts most, and the fallback email covers a failed send. Webhook, Slack and SMS record the outcome on the channel, so Settings shows the last delivery.

An SMS Twilio accepts can still fail at the carrier (an unreachable number, a STOP opt-out), so its outcome is final only when its status callback says so: a failure is recorded on the channel then, and when the SMS was the alert's only delivery, the fallback email goes out then. A dispatch sends SMS last, so that is known when the SMS is sent and stored with it; an SMS waits in `sms_messages` for its callback for up to 7 days. Twilio signs callbacks only with the account's primary Auth Token, which this app doesn't hold, so the callback URL carries a secret token derived from `JWT_SECRET` instead.

An alert for a monitor in a project starts its headline with the project name in brackets, `[Acme] nightly-backup missed its ping`, on Telegram, email (body and subject), Slack and SMS; the webhook event adds `projectId` and `projectName` and keeps every existing field, so receivers built before projects keep working (ADR projects). Without a project the alert is unchanged.

An uptime, port or DNS check that fails right after a passing one doesn't count yet: the monitor is checked again on the next tick, about 30 seconds later, and only a failed recheck counts as a failure toward `alert_after_n_failures`. All checks run from one server, so a single request failing between it and the target (as on 2026-10-26, when a check got a 502 while Cloudflare logged no error for any of this server's requests to the target) looked like an outage and alerted, then recovered at the next check 10 minutes later. Both checks stay in the check log. SSL and domain checks run once a day, and an error there showed as Error, on the status page too, and blanked the expiry date until the next day; an error right after a working check is rechecked the same way, and since these monitors keep no check history, a `recheck_pending` flag on the monitor marks the recheck.

An SSL check error that survives its recheck alerts, whatever the cause: an expired or untrusted certificate, a wrong name, or nothing answering on port 443. The TLS handshake verifies the certificate, so an expired one fails as an error and never reaches the expiry thresholds; until 2026-10-28 an error only set the status to Error, so a certificate that expired, or was already invalid when the monitor was added, alerted nobody. An error starts a new incident: it alerts again at each daily check, after `alert_after_n_failures` and up to `max_alerts_per_incident`, as the thresholds do, and the first working check after an alerted error sends a recovery. A domain lookup error doesn't alert: it means RDAP or WHOIS failed, not the user's domain.

Each monitor's `max_alerts_per_incident` caps alerts per incident (default 3; 0 means every failed check alerts); the count lives on the incident and starts at zero with the next one. Uptime, port and DNS monitors send one down alert per incident, so the cap only comes into play for cron, SSL and domain monitors. A recovery alert is always sent. The cap is per monitor because urgency differs per monitor.

Email goes through Resend, from `noreply@checkmeup.net`; human mailboxes stay on Zoho. SMS goes through Twilio, authenticated with a scoped API key. A phone number is saved only with an explicit consent checkbox, recorded with its time; opting out is disabling the channel in the app. SMS templates stay plain ASCII and within one 160-character segment, since one Unicode character halves the limit and doubles the cost. A send the org's SMS credits can't cover skips that channel only; other channels still fire. A test SMS doesn't use credits; a separate per-org limit (10 a minute, 10 an hour, 20 a day) bounds its cost instead.

A missing provider key (Resend, Telegram, Twilio) turns that channel off with a log warning; the app still runs.

## Considered options

- One Telegram chat and one email per org (until 2026-06-20): replaced by channels so a monitor can route to several destinations of any type.
- Returning Slack and webhook URLs as stored (until 2026-10-14): any leak of an API response (XSS, a shared screen, a future team member) would hand out a working way to post to the user's channel.
- SSL alerting only on expiry thresholds (until 2026-10-28): an expired or invalid certificate fails the handshake, so it alerted nobody, while the product pages promised an alert at once.
- Alerting only on SSL certificate errors, not on an unreachable port 443: someone with only an SSL monitor would hear nothing while the site is down.
- Reading an expired certificate's date with a second, unverified handshake so it shows as Expired: more code for the same alert; the Error message already says the certificate has expired.
- Retrying inside the same check: the check's goroutine would wait the 30 seconds, and the tick waits for every check, so one failing monitor would delay all the others.
- A channel group entity: neither healthchecks.io nor UptimeRobot needs one; no demand.
- Viber: since February 2024 Viber bots require a commercial agreement with Rakuten, so there is no self-serve way to get a bot token. Reconsider only if self-serve bots return.
- Signal (2026-10-14): Signal has no bot or business API; the only path is self-hosting `signal-cli` on a phone number registered as a bot, which Signal's anti-abuse systems may ban, and an always-on process ADR workers avoids. Demand is unconfirmed. Reconsider if Signal opens an official API or users keep asking.
- Microsoft Teams (2026-10-14): building and testing it needs a Microsoft 365 business tenant, and Microsoft is retiring the Office 365 Connector incoming webhook for Power Automate workflows, so the mechanism itself is moving. Revisit if users ask for it.
- Verifying `X-Twilio-Signature` on SMS status callbacks: it needs the primary Auth Token in production, which the scoped API key exists to avoid.
- AWS SNS for SMS: needs an AWS account and IAM for one call, unlike every other integration here (one API key, one HTTP call).
- Vonage for SMS: similar pricing, weaker 10DLC registration tooling.
- Postmark, AWS SES or Zoho SMTP for email: Postmark's free tier is too small, SES needs AWS, Zoho's shared IP hurts deliverability.

## Consequences

- A real outage alerts about 30 seconds later than before; a blip shorter than that never alerts.
- Loading a monitor's channels is one more query per alert. Revisit if monitor counts make it matter.
- SMS is the only channel that costs money per message; the per-incident cap is what bounds it.
- Editing a Slack or webhook channel can't send a test until the URL is pasted again: the test needs the real URL, and the client only holds the masked one.
- Rotating `JWT_SECRET` changes the callback token, so callbacks for SMS already sent get a 404 and those messages age out of `sms_messages` unresolved.
- US SMS traffic needs a one-time 10DLC / toll-free registration in the Twilio console (`docs/reference/twilio-setup.md`).
