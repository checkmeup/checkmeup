import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "On October 26 and again on October 28, Checkmeup told me one of my own sites had been down for 11 minutes. It hadn't. The site sits behind Cloudflare, and for the 26th Cloudflare's logs show no error for any request from Checkmeup's server, even though the check got a 502. What failed was a single request each time, somewhere between that server and the site. The monitor checks every 10 minutes and, by default, alerted on the first failed check, so one lost request became a down alert, and the next check 10 minutes later became the recovery.",
  },
  {
    type: 'h2',
    text: 'A failed check now gets a second look',
  },
  {
    type: 'p',
    text: 'When an uptime, port or DNS check fails right after a passing one, it no longer counts straight away. The monitor is checked again about 30 seconds later. If the recheck passes, nothing happens: no incident, no alert. If it fails too, that is the first failure counted, and everything works as before from there, including "alert after N failures" if you set it. Both checks stay in the check log, so a blip is still visible on the detail page, it just doesn\'t wake anyone.',
  },
  {
    type: 'p',
    text: "The recheck isn't a retry inside the same check. Checks run in batches every 30 seconds, and a batch waits for all of its checks to finish, so a check that slept 30 seconds before retrying would delay every other monitor in the batch. Instead the monitor is simply due again 30 seconds later and joins the next batch.",
  },
  {
    type: 'p',
    text: 'The trade-off is deliberate: a real outage now alerts about 30 seconds later than before, and a blip shorter than that never alerts at all. For a check that runs every 1 to 30 minutes, I think that is the right side to err on.',
  },
  {
    type: 'p',
    text: "What this doesn't fix: every check still runs from one server, so a longer network problem near that server can still look like your site going down. Confirming a failure from a second region is the real answer, and it is on the roadmap, waiting on a decision about where the extra servers run.",
  },
  {
    type: 'h2',
    text: 'SSL and domain errors no longer stick for a day',
  },
  {
    type: 'p',
    text: "SSL and domain monitors check once a day and alert on expiry thresholds, not on errors, but a single failed TLS handshake or RDAP lookup used to set the monitor to Error, on your public status page too, and blank its expiry date until the next day's check. Now an error right after a working check gets the same recheck about 30 seconds later, and the monitor shows Error only if that fails too.",
  },
  {
    type: 'h2',
    text: 'A page per monitor type, and honest comparisons',
  },
  {
    type: 'p',
    text: "Each of the six monitor types now has its own page, such as /uptime-monitoring and /dns-monitoring, saying what it checks, when it alerts and what it can't do. Three comparison pages, for Healthchecks.io, UptimeRobot and Cronitor, put the plans side by side, with the date the competitor facts were checked and links to the vendors' own pages. Each one also says when the other tool is the better choice.",
  },
  {
    type: 'p',
    text: 'Writing this release turned up a mistake there: the uptime and port pages promised that two failed checks in a row mark a monitor down, while the default was actually one. They now describe the recheck, which is what the default does.',
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'ul',
    items: [
      'A tab left open across a deploy used to break on its next page change, because the files it asked for belonged to the previous version. It now reloads once and carries on.',
      'The footer groups its links under Product, Monitoring, Resources and Company.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'Next up are team management and two-factor authentication. The full history, including the decision records behind the recheck, is on GitHub.',
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
