import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "Until v1.61, an SSL monitor watching a certificate that had already expired told nobody. The check opens a TLS connection, and the handshake verifies the certificate, so an expired one fails as an error before its expiry date is ever read. An error set the monitor to Error and stopped there: no alert, and none of the 30, 14 or 7-day reminders, since those come from the date. The product page and the docs promised an alert the moment a certificate stops working, and that promise wasn't kept.",
  },
  {
    type: 'h2',
    text: 'An SSL error now alerts',
  },
  {
    type: 'p',
    text: 'An SSL check error that survives its recheck about 30 seconds later now alerts, whatever the cause: an expired or untrusted certificate, a name that doesn\'t match, or nothing answering on port 443 at all. The alert gives the error as the reason. An error starts an incident of its own, so it alerts again at each daily check, after "alert after N failures" and up to the monitor\'s alerts-per-incident cap, the same way the expiry reminders do. The first working check after that sends a recovery.',
  },
  {
    type: 'p',
    text: 'Nothing answering on 443 alerts too, on purpose. Someone whose only monitor for a site is an SSL one would otherwise hear nothing while that site is down.',
  },
  {
    type: 'p',
    text: "Domain monitors don't get the same treatment. A domain check reads the expiry date from the registry over RDAP, or WHOIS where a registry has no RDAP, so an error there means the registry's service failed, not that anything happened to your domain. It still shows on the monitor, but it doesn't alert.",
  },
  {
    type: 'h2',
    text: 'Alerts no longer wait for a slow endpoint',
  },
  {
    type: 'p',
    text: "Checks run in batches every 30 seconds, and a batch finishes when its slowest check does. Alerts used to be sent inside the check, one channel after another, each allowed 10 seconds, so a single slow webhook receiver held up every other monitor's next check. A cron job's ping had the same problem: its recovery alert went out inside the request, so the job's curl waited for every channel, and a job that gave up after 5 seconds cancelled the alert on the channels not yet tried.",
  },
  {
    type: 'p',
    text: "Now each alert goes out on its own, apart from the check or the ping that raised it, and a ping is answered at once. The trade-off: alerts for one monitor aren't queued in order, so with three or more slow channels in a row, a recovery can arrive before the down alert it follows. For an outage shorter than the time your channels take to answer, I'd rather have both alerts quickly than both in order and late.",
  },
  {
    type: 'p',
    text: 'Every merge to the main branch deploys Checkmeup, often several times a day, and each deploy used to stop the old server mid-flight: requests cut off, checks abandoned halfway, alerts lost. The server now finishes the requests it is answering, the batch of checks it is running and the alerts it is sending, waiting up to 15 seconds for the checks and alerts before it stops. An alert still sending after that is lost.',
  },
  {
    type: 'h2',
    text: 'Two more free tools',
  },
  {
    type: 'p',
    text: 'Next to the SSL and domain expiry checkers, two tools need no account. The DNS lookup at /tools/dns-lookup shows a hostname\'s A, AAAA, CNAME, MX, TXT and NS records at once, through the same resolver a DNS monitor uses. The cron expression explainer at /tools/cron-expression turns an expression such as "*/15 9-17 * * 1-5" into plain English, with the next five runs in UTC and in your own time zone, as you type.',
  },
  {
    type: 'p',
    text: 'The explainer reads expressions with the same parser cron monitors use, and building it turned up a bug: a cron monitor accepted any five fields, so "60 * * * *" was saved and then quietly expected a ping every hour. A schedule the parser can\'t read is now refused, with the field at fault named ("Minute: 60 is above the maximum, 59").',
  },
  {
    type: 'h2',
    text: 'If Checkmeup ever shuts down',
  },
  {
    type: 'p',
    text: "Checkmeup is one person's product, and the obvious question about one is what happens if that person stops. So it is now in the Terms of Service, section 11: if Checkmeup is ever discontinued, every account holder gets an email at least 90 days before it stops, any prepaid time past that date is refunded, and the source stays public, so you can run it yourself. The homepage and the About page say who builds it.",
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'ul',
    items: [
      'A monitor with alerts turned off shows a Muted label in the monitor list and on the dashboard. Its checks still run; nothing is sent.',
      "A new monitor's form starts with every enabled notification channel selected. Clearing them all now attaches none, so alerts go to your account email, as the form says. Before, a monitor saved with none selected got every channel anyway.",
      "Settings shows Node.js code that verifies a webhook's X-Checkmeup-Signature with its signing secret.",
      "A link to any article or free tool, shared in Slack, LinkedIn or a chat, previews with a picture of that page's title instead of the same homepage card.",
      'A request body over 64 KB is refused with 413 up front, instead of a 400 partway through reading it.',
      'The SMS consent checkbox states how often messages come, that rates may apply, how to stop them, and links the Terms and Privacy Policy.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'What comes next depends on the people who start using Checkmeup. Team management, two-factor authentication and checks from more than one region are written up and waiting, and what users ask for will decide the order. If something you need is missing, the feedback form in the app, email or a GitHub issue reaches me directly. The full history, including the decision records behind each change here, is on GitHub.',
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
