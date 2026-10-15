import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "The first client is easy to monitor. You add an uptime check for their site, maybe the SSL certificate, and you'd notice anything odd because there's so little to look at. By the eighth client you have forty monitors in one list, an alert at 2 AM says \"backup missed its ping\", and the first thing you have to work out is whose backup it is. Monitoring for many clients is less about which checks to run than about keeping each client's checks, alerts and status in one recognizable place.",
  },
  {
    type: 'h2',
    text: 'Start with the same set of checks for every client',
  },
  {
    type: 'p',
    text: 'Most client sites fail in the same handful of ways, so most clients need the same baseline. Writing it down once means a new client gets complete coverage on day one instead of whatever you remembered to add.',
  },
  {
    type: 'ul',
    items: [
      "An uptime check on the main site, ideally with a keyword from the page, so a blank page or a hosting provider's error page that returns 200 still counts as down.",
      'The SSL certificate, with warnings at 30, 14 and 7 days. Auto-renewal fails quietly more often than it should.',
      'The domain registration. A lapsed domain takes the site, the email and sometimes the name itself with it, and the renewal emails usually go to someone who left.',
      "Every scheduled job you set up for them: backups, imports, sitemap builds. A cron monitor expects a ping each run and alerts when one doesn't arrive, which catches the job that stopped running weeks ago and never errored.",
      'The MX record, if you manage their DNS. Mail rerouted by a bad change is invisible to every website check.',
    ],
  },
  {
    type: 'p',
    text: "Five or six monitors per client is normal. It's also exactly why a flat list stops working: the count grows with clients times checks, not with clients.",
  },
  {
    type: 'h2',
    text: 'Group monitors by client, not by type',
  },
  {
    type: 'p',
    text: "Monitoring tools tend to organize by check type: all uptime checks together, all certificates together. That matches how the tool works, not how you think about the work. When a client calls, you want their uptime, certificate, domain and backups on one screen. When something breaks, you want to see at a glance whether it's one client with three red monitors or three clients with one each, because those are very different mornings.",
  },
  {
    type: 'p',
    text: 'Put each monitor in exactly one group per client. Tags look more flexible, but a monitor tagged "acme", "production" and "wordpress" has no single home, and the grouped view stops answering "what\'s wrong with Acme?" cleanly. If you also need cross-cutting views like all production sites, a consistent naming scheme handles most of it.',
  },
  {
    type: 'h2',
    text: 'Name monitors so they read well out of context',
  },
  {
    type: 'p',
    text: 'An alert arrives on a phone lock screen with one line of text. "Uptime check 3 is down" tells you nothing. Name every monitor by what it watches, in words the client would recognize: "acme.com homepage", "acme shop checkout", "nightly database backup". The client\'s name doesn\'t need to be in every monitor\'s name if your tool puts the group in the alert; if it doesn\'t, put it there yourself.',
  },
  {
    type: 'h2',
    text: 'Make every alert say whose it is',
  },
  {
    type: 'p',
    text: 'This is the part that matters at 2 AM. An alert should tell you which client it concerns before you open anything, on every channel you use: the Telegram message, the email subject line, the Slack post, the SMS. If you route alerts into your own tooling through a webhook, the client should arrive as a field you can route on, not as text you have to parse.',
  },
  {
    type: 'p',
    text: 'Also decide who receives what. Usually you want every alert yourself, and some clients want to hear about their own outages directly. Sending a client their own alerts works well for urgent ones, like the site being down, and badly for noise like a certificate that renews automatically in 29 days.',
  },
  {
    type: 'h2',
    text: 'Give each client a status page of their own',
  },
  {
    type: 'p',
    text: 'A public status page with just that client\'s services answers "is it down or is it me?" before they email you, and it shows the work you do when nothing is wrong. Keep it to what the client cares about, usually the site, the shop and their email. Your internal checks like backups and DNS stay off it.',
  },
  {
    type: 'h2',
    text: 'Announce planned work before it alerts',
  },
  {
    type: 'p',
    text: "A migration or a hosting move will trip the uptime check. A maintenance window on that client's monitors silences the alerts for that period and shows the maintenance on their status page, so nobody, you included, gets woken up by work you scheduled yourself.",
  },
  {
    type: 'h2',
    text: 'Setting this up in Checkmeup',
  },
  {
    type: 'p',
    text: 'Checkmeup has projects for exactly this: one per client, each monitor in at most one. The monitor list shows each project as its own section with how many of its monitors need attention, and /monitors?project= gives every client a link of their own. Every alert for a project\'s monitor starts with the project in brackets, "[Acme] nightly database backup missed its ping", on Telegram, email, Slack and SMS, and webhook events carry projectId and projectName. Projects don\'t count against your plan.',
  },
  {
    type: 'p',
    text: "Status pages and maintenance windows each take a hand-picked list of monitors, so a per-client page or a per-client maintenance window is a matter of picking that client's monitors. What isn't there yet: defaults per project, such as a new Acme monitor alerting Acme's Slack channel automatically. For now, choose the channels on each monitor.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
