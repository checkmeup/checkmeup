import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: 'Nobody sets out to find out from a customer. It just happens that way, because the failures that hurt most are the quiet ones — a nightly backup job that stopped running six weeks ago, a certificate that expired at 2am on a Saturday, a domain that lapsed because the renewal email went to someone who left the company. None of these throw an error anyone sees. They just stop working, and the silence looks exactly like everything being fine.',
  },
  {
    type: 'h2',
    text: 'What monitoring actually buys you',
  },
  {
    type: 'p',
    text: "Monitoring doesn't keep anything running. It's worth being blunt about that, because a lot of monitoring copy quietly implies otherwise. A monitor is a second observer: something outside your system that checks, on a schedule, whether a specific fact is still true — this URL returns the right content, this job checked in, this certificate has more than a week left. What you're buying is the interval between when something breaks and when you find out. That's it. But that interval is where almost all of the cost of an outage lives.",
  },
  {
    type: 'p',
    text: 'Consider the difference in practice. A checkout page that breaks at 3am and gets noticed at 9am costs six hours of orders, plus however many customers decided the site was untrustworthy and went elsewhere. The same break, caught in five minutes, costs five minutes. The underlying bug is identical. The only variable is how long it ran unobserved.',
  },
  {
    type: 'h2',
    text: "The failures that don't announce themselves",
  },
  {
    type: 'p',
    text: 'Most teams monitor whether the site is up, and stop there. That covers the loudest failure mode and misses the ones that are actually more common, because a service being hard-down is the one thing you probably would hear about. The quieter categories are worth naming explicitly:',
  },
  {
    type: 'ul',
    items: [
      'Scheduled work that stops. A cron job, backup, or sync that fails silently produces no error page and no angry user — just a gap in your data that gets discovered when someone needs the thing that was supposed to be there.',
      'Expiry. Certificates, domains, and paid registrations all have a fixed date attached and no mechanism to remind you at the moment it matters. These are the most preventable outages there are, and they still happen constantly, because nothing fails until the deadline passes — and then everything does at once.',
      'Responses that are technically fine. A 200 status code on a page rendering an empty template, or an API returning valid JSON with no records in it, looks healthy to anything checking status codes alone.',
      'Infrastructure below HTTP. A database port that stopped accepting connections, or a DNS record that got edited during an unrelated migration, breaks your application without your web server ever knowing something is wrong.',
    ],
  },
  {
    type: 'p',
    text: "The common thread is that none of these produce a signal on their own. Every one of them requires something to go looking, on a schedule, and to notice the absence of an expected thing rather than the presence of an error. That's a genuinely different job from watching for crashes.",
  },
  {
    type: 'h2',
    text: "What a green dashboard doesn't prove",
  },
  {
    type: 'p',
    text: "Here's the failure mode nobody warns you about: monitoring that has quietly stopped meaning anything. A monitor pointed at a URL that got retired two refactors ago will happily report green forever. Alerts routed to a Slack channel everyone muted are alerts you don't have. A check interval of 30 minutes on a checkout flow means a green dashboard is telling you, at best, that things were fine within the last half hour. And a monitoring service that itself goes down produces the most reassuring possible signal — nothing at all.",
  },
  {
    type: 'p',
    text: 'So a green dashboard proves that the checks you configured, at the frequency you set, passed the assertions you wrote. It does not prove your system is healthy. Those overlap only to the degree that you did the setup thoughtfully — which is why the setup is worth more attention than the tool choice.',
  },
  {
    type: 'h2',
    text: 'Setting up monitoring in three steps',
  },
  {
    type: 'h3',
    text: '1. Start from what actually costs you money',
  },
  {
    type: 'p',
    text: "Don't start with a list of servers. Start by asking what has to be working for the business to function — the checkout, the login, the nightly invoice run, the API a client integrated against. Then put a monitor on each of those specifically. Ten monitors on the things that matter beat a hundred on things nobody would notice failing, and they generate far less noise.",
  },
  {
    type: 'h3',
    text: '2. Assert on something meaningful, not just reachability',
  },
  {
    type: 'p',
    text: 'For anything HTTP, check for a string or JSON field that only appears when the page genuinely rendered correctly. For scheduled jobs, have the job itself check in on completion, so silence is the failure signal rather than something you have to infer. This is the single step that separates monitoring that catches real problems from monitoring that catches only total outages.',
  },
  {
    type: 'h3',
    text: '3. Route alerts to somewhere you will actually react',
  },
  {
    type: 'p',
    text: 'A one-minute check interval is wasted if the alert lands in an inbox you read twice a day. Match the channel to the urgency — a push channel for anything customer-facing, email for the things you can address tomorrow — and turn on recovery notifications too, so you find out it is over rather than just that it started.',
  },
  {
    type: 'h2',
    text: 'What to look for in a monitoring tool',
  },
  {
    type: 'ul',
    items: [
      'Coverage of the failure modes you actually have — scheduled jobs and expiry dates included, not only HTTP uptime',
      'Content or keyword assertions, so a technically-successful response on a broken page still counts as a failure',
      'A check interval short enough for your most costly path, and a failure threshold you can tune so one flaky check does not become noise',
      'Advance warning on anything with an expiry date, at more than one interval — a single notice on the day it expires is not a warning',
      'A history you can look back on: when it broke, when it recovered, how long it lasted. A live status badge that forgets everything once it flips back to green cannot tell you whether a problem is recurring',
    ],
  },
  {
    type: 'h2',
    text: 'The mistake worth avoiding',
  },
  {
    type: 'p',
    text: "The mistake isn't having no monitoring — it's having monitoring nobody trusts. Alert on everything, including the things that resolve themselves in forty seconds, and people start ignoring the notifications within a week. From there the tool is worse than nothing, because it produces a documented feeling of safety while the real alert scrolls past unread among the noise. Fewer monitors, on things that genuinely matter, with thresholds tuned so that an alert means something is actually wrong: that's what makes the next real one get read.",
  },
  {
    type: 'divider',
  },
  {
    type: 'p',
    text: 'Checkmeup covers the six failure modes above in one place — cron jobs, uptime, SSL expiry, domain expiry, ports, and DNS records — with keyword and JSON assertions on HTTP checks, tiered advance warnings at 30, 14, and 7 days before a certificate expires, a tunable failure threshold, full incident history, and alerts to email, Telegram, Slack, SMS, or your own webhook. The Hobby plan is free for up to 10 monitors on a 5-minute check interval, no credit card required.',
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
