import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: '9:14 AM. A Slack message from a client, no greeting: "is the site down for you too??" Then a screenshot — not a 500 page, not a blank white screen, but a full-bleed red warning. NET::ERR_CERT_DATE_INVALID. Your connection is not private. Every browser on earth was refusing to even load the page, and had been since roughly 3 AM, which meant six hours of checkout traffic had been quietly bounced before a single human noticed.',
  },
  {
    type: 'h2',
    text: "The part that stings isn't the outage",
  },
  {
    type: 'p',
    text: "A server crash is embarrassing but forgivable — things break, that's infrastructure. An expired TLS certificate is different, and it feels different, because it was never a mystery. Certificate lifetimes aren't random; they're printed in the cert itself the day it's issued. Nothing failed unpredictably. A date on a calendar just arrived, and nobody was watching the calendar.",
  },
  {
    type: 'p',
    text: "That's the part that actually costs you sleep afterward — not the six hours of downtime, but replaying how avoidable it was. No exotic infrastructure failure, no DDoS, no upstream provider outage to point at. Just a 90-day Let's Encrypt cert, an auto-renewal cron job that had been silently failing since a server migration two months earlier, and nobody positioned to notice until a customer did.",
  },
  {
    type: 'h2',
    text: "Why the failure is invisible until it isn't",
  },
  {
    type: 'p',
    text: "Right up until the expiry date, an SSL certificate gives you zero signal that anything's wrong. The site loads fine. The padlock shows fine. curl works fine. There's no gradual degradation to notice, no error rate creeping up in a dashboard somewhere — because the certificate itself doesn't care what day it is, right up until the exact second it does. Then every browser, all at once, independently decides your site is unsafe and refuses to render it. Not a slowdown. A cliff.",
  },
  {
    type: 'p',
    text: "And the automation that's supposed to prevent this is exactly the kind of thing that fails silently, too. certbot renew running as a cron job with no monitoring on the cron job itself is a single point of failure disguised as a safety net — it works until a permissions change, a path that moved during a migration, a webhook that silently started returning a non-200, or a renewal rate limit nobody expected. When it fails, it doesn't send an email. It just quietly stops, and the countdown to that cliff keeps running.",
  },
  {
    type: 'h2',
    text: 'What it actually costs',
  },
  {
    type: 'p',
    text: 'Six hours doesn\'t sound like much until you multiply it by every visitor who hit a full-screen security warning instead of a product page. Some closed the tab immediately — a security warning reads as "this site got hacked," not "routine cert hiccup," to anyone who isn\'t in this industry. Some searched the client\'s name plus "down" and found nothing reassuring. A few emailed support, which is how the client found out before the fix was even in motion. The revenue lost that morning was real, but the trust cost was worse and lasted longer — the client\'s next question, reasonably, was why they\'d found out from their own customers instead of from the agency managing their infrastructure.',
  },
  {
    type: 'h2',
    text: 'What actually catches this',
  },
  {
    type: 'p',
    text: "Not the renewal cron job's own success — that's the exact mechanism that already failed silently. What catches it is something external, checking the live certificate itself rather than trusting that the process meant to renew it worked. That's the whole idea behind Checkmeup's SSL monitor: it connects to your domain daily, reads the certificate that's actually being served, and alerts at 30, 14, and 7 days before expiry — plus a final alert if it lapses anyway. It doesn't care whether certbot ran, whether the renewal webhook fired, or whether anyone remembered to check. It only cares what's actually on the wire, which is the one thing that matters at 9:14 AM.",
  },
  {
    type: 'h2',
    text: "A short checklist, if you're carrying this risk today",
  },
  {
    type: 'ul',
    items: [
      "Monitor the certificate independently of the renewal process — a passing cron job and a valid certificate are not the same fact, and only one of them is what your visitors' browsers check.",
      'Set alert thresholds with real runway — 7 days before expiry is enough time to fix a broken renewal script; 7 hours is not.',
      "Confirm renewal actually happened after every server migration, DNS change, or webhook update — it's exactly the kind of automation that keeps working in tests and quietly stops in production.",
      "Make sure the alert reaches a person, not just a log file nobody tails — Telegram, Slack, email, SMS, whatever's actually checked at 3 AM.",
    ],
  },
  {
    type: 'p',
    text: "None of this is exotic — it's the same logic as watching a cron job's heartbeat instead of trusting the job ran, applied to the certificate underneath every request your site serves. The difference between a non-event and a 9:14 AM Slack message is usually just: something was watching before the browser had to.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
