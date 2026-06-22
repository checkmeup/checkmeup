import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-7-release-notes',
  title: 'v1.7: Webhook Alerts',
  date: 'June 22, 2026',
  readTime: '3 min read',
  excerpt:
    "A third alert channel alongside Telegram and email: webhooks. Point a monitor at any HTTPS endpoint and get a signed POST when it goes down or recovers — Slack incoming webhooks, PagerDuty, your own scripts, whatever's listening. Plus the SSRF hardening that comes with letting users hand you an arbitrary URL to request.",
  content: [
    {
      type: 'p',
      text: 'v1.6 introduced notification channels — the ability to attach more than one Telegram chat or email address to a monitor instead of one fixed pair per org. Webhooks are the first new channel type built on top of that model, and the reason the model needed to exist before this could: a generic outbound webhook is what lets checkmeup plug into Slack, PagerDuty, or anything else you run, without checkmeup building a bespoke integration for each one.',
    },
    {
      type: 'h3',
      text: 'Webhook alerts',
    },
    {
      type: 'p',
      text: "Add a webhook channel in Settings the same way you'd add a Telegram chat or alert email: give it a name, an HTTPS URL, and hit \"Send test webhook\" to confirm it's reachable before you save — same pattern as the existing test buttons. You can add as many webhook channels as you want, and each monitor's channel picker treats them exactly like any other channel: attach the ones you want that monitor to use.",
    },
    {
      type: 'p',
      text: 'A down or recovery event fires a POST within one check cycle of the transition, same timing guarantee as Telegram and email. The JSON body carries the event type, monitor name and type (cron/uptime/SSL), the failure reason on a down event, and downtime duration on recovery — enough to drive an automation without an extra API call back to checkmeup.',
    },
    {
      type: 'h3',
      text: "Verifying it's really checkmeup",
    },
    {
      type: 'p',
      text: "Every webhook request carries an X-Checkmeup-Signature header — an HMAC-SHA256 of the raw request body, signed with a secret generated automatically the first time you save the webhook. Settings shows that secret along with a short snippet for verifying the signature on your end, so whatever's receiving the webhook can confirm it actually came from checkmeup before acting on it. Regenerating the secret only affects future sends — it won't retroactively invalidate anything already delivered.",
    },
    {
      type: 'h3',
      text: 'No retry storms',
    },
    {
      type: 'p',
      text: "Webhook delivery gets one attempt, no retries — a broken endpoint on your end logs a failure instead of triggering a retry loop that could pile up against the worker. Settings shows the last delivery's outcome (status code or error, how long ago) so a silently-failing integration doesn't stay invisible. And the per-incident alert cap now counts across all three channels together, so turning on Telegram, email, and webhook at once doesn't triple your alert volume for the same incident — recovery events still always get through, uncapped, on every enabled channel.",
    },
    {
      type: 'h3',
      text: 'Hardening against where webhook URLs point',
    },
    {
      type: 'p',
      text: "A webhook URL is the one place in checkmeup where a user hands the server an arbitrary destination and asks it to make a request. Left unchecked, that's a standard SSRF vector — someone could point a webhook at internal infrastructure, like a cloud metadata endpoint, and have checkmeup's own server fetch it on their behalf. Outbound webhook requests now refuse to connect to loopback, private, link-local, unspecified, or multicast addresses, checked on the actual resolved IP at connect time rather than on the URL up front — so a hostname that resolves differently between validation and connect can't slip past the check. Redirects aren't followed either, closing off a 3xx response retargeting the request after the original URL passed muster.",
    },
    {
      type: 'h3',
      text: 'Test coverage',
    },
    {
      type: 'p',
      text: 'The web API client modules and several backend packages — legal, respond, server, and Telegram — picked up their first unit tests this release, narrowing the set of code that still relies on manual checking alone.',
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: 'Slack and Teams alerts are next on the roadmap, both built directly on the webhook channel shipped here. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.',
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
