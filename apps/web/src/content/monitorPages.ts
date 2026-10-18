// The product page of each monitor type, at /<type>-monitoring (PRD
// marketing-pages). Each claim matches the type's PRD and the worker's actual
// behavior — check both before changing one. The spec line (what it watches,
// how often, when it alerts) comes from `monitorTypes` in home.ts, so the
// landing page's table and these pages can't disagree.
export type MonitorTypeId = 'cron' | 'uptime' | 'ssl' | 'domain' | 'port' | 'dns'

export interface MonitorPage {
  id: MonitorTypeId
  path: string
  seoTitle: string
  description: string
  title: string
  lead: string
  // How the check works, one short paragraph each.
  how: string[]
  // What a passing check doesn't prove: said plainly, as the guides do.
  limits: string
  related: { to: string; label: string }[]
}

export const monitorPages: MonitorPage[] = [
  {
    id: 'cron',
    path: '/cron-monitoring',
    seoTitle: 'Cron Job Monitoring — Checkmeup',
    description:
      'Get alerted when a cron job, scheduled task or worker stops running. One curl to a ping URL, no agent. Free for 10 monitors.',
    title: 'Cron job monitoring',
    lead: 'Know the moment a scheduled job stops running. Your job pings a URL when it finishes; if the ping is late, you get an alert.',
    how: [
      'Give the monitor a schedule, a 5-field cron expression or an interval such as every hour, and a grace period. Checkmeup gives you a secret ping URL. Add one curl to the end of your job.',
      "Each ping marks the monitor up and sets when the next one is due: the schedule's next run plus the grace period. If that time passes without a ping, the monitor goes down within 30 seconds and you get an alert. The next ping resolves the incident and sends a recovery alert.",
      "Send a start ping as well and Checkmeup records each run's duration, and with a maximum duration set, alerts you once about a run that is stuck. Query parameters on a ping are kept with it, so a job can report what it did.",
    ],
    limits:
      "A ping says the job reached the line that sends it. It doesn't say the job did the right thing: a backup that wrote an empty file still pings. Put the ping after the step that matters, and have the job fail before it when that step fails.",
    related: [{ to: '/blog/cron-job-monitoring-guide', label: 'Guide: cron job monitoring' }],
  },
  {
    id: 'uptime',
    path: '/uptime-monitoring',
    seoTitle: 'Uptime Monitoring with Keyword and JSON Checks — Checkmeup',
    description:
      'Check a URL every minute and get alerted when it errors, slows down or loses the content you expect. Keyword and JSON checks on every plan.',
    title: 'Uptime monitoring',
    lead: 'Find out a site or API is down before its users tell you. Checkmeup requests your URL on a schedule and alerts you when the answer is wrong.',
    how: [
      'Each check is one request, GET by default, with redirects followed. With default settings a 200 within 10 seconds is up. You can change the method to HEAD or POST, the accepted status codes, and the maximum response time, from 1 to 30 seconds.',
      'A keyword check fails a page that answers 200 but lost its content: "contains" or "does not contain", searched in the first 512 KB. JSON assertions check a health endpoint\'s fields, such as $.status equals "ok", and all of them must pass.',
      "Two failed checks in a row mark the monitor down, so one network blip doesn't page anyone. The next passing check sends a recovery alert with the downtime. The detail page shows response times, uptime over 24 hours, 7 and 30 days, and every incident.",
    ],
    limits:
      'Every check runs from one server, so a problem on the network between that server and your site can look like your site going down. Two failures in a row filter out most blips, not all.',
    related: [{ to: '/blog/uptime-monitoring-guide', label: 'Guide: uptime monitoring' }],
  },
  {
    id: 'ssl',
    path: '/ssl-monitoring',
    seoTitle: 'SSL Certificate Expiry Monitoring — Checkmeup',
    description:
      'Get alerted 30, 14 and 7 days before an SSL certificate expires, and at once when one is invalid. Checked daily. Free for 10 monitors.',
    title: 'SSL certificate monitoring',
    lead: 'Renew every certificate before visitors see a browser warning. Checkmeup reads the certificate once a day and warns you three times before it expires.',
    how: [
      "Enter a hostname, or paste a URL and Checkmeup keeps the hostname. Once a day it connects on port 443 over TLS, the way a browser does, and records the certificate's issuer, expiry date and days left.",
      "You get one alert as the certificate crosses 30, 14 and 7 days left, never two for the same threshold, and an alert at once when it has expired or doesn't validate.",
    ],
    limits:
      "The check verifies the certificate the way most clients do, from one place. It doesn't test every client's trust store, and a host that serves different certificates by region may show you only one of them.",
    related: [
      { to: '/tools/ssl-checker', label: 'Free SSL certificate checker' },
      {
        to: '/blog/ssl-certificate-expired-story',
        label: 'The 3 AM SSL cert that cost us a client',
      },
    ],
  },
  {
    id: 'domain',
    path: '/domain-expiry-monitoring',
    seoTitle: 'Domain Expiry Monitoring — Checkmeup',
    description:
      'Get alerted 30, 14 and 7 days before a domain registration expires. Daily RDAP and WHOIS lookups, including .io and .co. Free for 10 monitors.',
    title: 'Domain expiry monitoring',
    lead: 'A lapsed domain takes the site, its email and everything else on it down with it. Checkmeup reads the registration once a day and warns you before it lapses.',
    how: [
      'Once a day Checkmeup looks the domain up through RDAP, or over WHOIS for registries without RDAP that publish an expiry date there: .il, .io, .co, .me, .us, .it, .ru, .jp, .cn and .se. It records the registrar, the expiry date and the days left.',
      'You get one alert as the registration crosses 30, 14 and 7 days left, and one at once when it has expired. A failed lookup is shown as an error, never as expired.',
    ],
    limits:
      "Some registries publish no expiry date at all (.de, .eu, .at) or answer WHOIS only for listed clients (.ch, .li, .es), so those domains can't be monitored; the form says so instead of accepting a monitor that could never work.",
    related: [{ to: '/tools/domain-expiry-checker', label: 'Free domain expiry checker' }],
  },
  {
    id: 'port',
    path: '/port-monitoring',
    seoTitle: 'TCP Port Monitoring — Checkmeup',
    description:
      'Monitor that a TCP port stays open, or stays closed. For mail servers, databases and other non-HTTP services. Free for 10 monitors.',
    title: 'Port (TCP) monitoring',
    lead: "Watch the services an HTTP check can't reach, and the ports that should never be open. Checkmeup connects to host:port on a schedule and alerts you when the answer changes.",
    how: [
      'Each check is a plain TCP connect to the host and port, closed straight away, with the connect time recorded. Set the expected state to open, the default, or to closed for a port that must stay shut, such as a database that should never face the internet.',
      "Two failed checks in a row mark the monitor down and alert you; a closed-state monitor's alert says the port is unexpectedly open. The next passing check sends a recovery alert with the downtime.",
    ],
    limits:
      'An open port means something accepted the connection, not that the service behind it works: no handshake or data is sent. Pair it with an uptime check where the service speaks HTTP.',
    related: [{ to: '/blog/port-tcp-monitoring-guide', label: 'Guide: port (TCP) monitoring' }],
  },
  {
    id: 'dns',
    path: '/dns-monitoring',
    seoTitle: 'DNS Record Monitoring — Checkmeup',
    description:
      'Get alerted when an A, AAAA, CNAME, MX, TXT or NS record changes or stops resolving, with the old and new value. Free for 10 monitors.',
    title: 'DNS record monitoring',
    lead: 'A changed record can send your traffic or your email somewhere else while every uptime check stays green. Checkmeup looks the record up on a schedule and tells you what changed.',
    how: [
      'Pick a hostname and a record type: A, AAAA, CNAME, MX, TXT or NS. Pin the value you expect, or leave it empty and the first successful lookup becomes the baseline.',
      'A different answer marks the monitor down as a change, and the alert shows the old value and the new one. The same records in another order are not a change. NXDOMAIN, SERVFAIL and timeouts are recorded as lookup errors, apart from changes.',
    ],
    limits:
      "Lookups come from one resolver in one place, so a change that hasn't reached it yet, or a split-horizon answer, isn't seen. DNSSEC isn't validated.",
    related: [{ to: '/blog/dns-record-monitoring-guide', label: 'Guide: DNS record monitoring' }],
  },
]
