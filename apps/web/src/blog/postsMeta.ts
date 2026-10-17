// Single source of truth for every post's list-page metadata — kept
// separate from posts/*.ts's `content` exports specifically so Vite/Rollup
// can code-split each post's (much heavier) prose into its own lazily-loaded
// chunk. A file that eagerly imports from posts/*.ts at all — even just one
// named export — forces Rollup to bundle that file's *entire* module
// (content included) into the eager chunk; see the ineffective-dynamic-import
// warning this replaced. 'file' is the basename under ./posts/ to dynamically
// import for that slug — usually equal to slug, but not always (a post's
// slug can outlive the filename it was drafted under).
import type { BlogPostMeta } from './types'

export interface BlogPostMetaEntry extends BlogPostMeta {
  file: string
}

export const postsMeta: BlogPostMetaEntry[] = [
  {
    file: 'v1-54-to-v1-56-free-tools',
    slug: 'v1-54-to-v1-56-free-tools',
    kind: 'release',
    title: 'v1.54–v1.56: Free SSL and Domain Expiry Checkers',
    date: 'October 17, 2026',
    readTime: '4 min read',
    excerpt:
      "Check when a certificate or a domain expires without an account, using the same checks a monitor runs. Also: per-IP rate limits that were shared by every visitor now apply per visitor, and one failing certificate no longer turns a status page's whole 90-day strip red.",
  },
  {
    file: 'monitoring-many-clients',
    slug: 'monitoring-many-clients',
    kind: 'article',
    title: 'Monitoring Many Clients Without Losing Track',
    date: 'October 15, 2026',
    readTime: '5 min read',
    excerpt:
      "Five or six checks per client turns into fifty monitors in one list by the tenth client, and an alert at 2 AM that doesn't say whose site it is. A practical setup for freelancers and agencies: one baseline per client, grouping by client instead of check type, and alerts that name the client on every channel.",
    description:
      'How freelancers and agencies keep fifty monitors manageable: one baseline per client, grouping by client, and alerts that name the client on every channel.',
  },
  {
    file: 'rdap-vs-whois-domain-expiry',
    slug: 'rdap-vs-whois-domain-expiry',
    kind: 'article',
    title: 'RDAP vs WHOIS: Why Some Domains Hide Their Expiry',
    date: 'October 15, 2026',
    readTime: '6 min read',
    excerpt:
      'Look up when a .com expires and you get a date in a second; try an .io, .co.il or .de domain and you may get an error or a false "not found". Here\'s where expiry dates come from, why many country-code registries have no RDAP, how their WHOIS answers differ, and how to check your own domain from the command line.',
    description:
      'Why a .com expiry date shows up instantly but .io, .de or .co.il lookups fail: where expiry data comes from, RDAP vs WHOIS, and how to check it yourself.',
  },
  {
    file: 'v1-53-domain-whois',
    slug: 'v1-53-domain-whois',
    kind: 'release',
    title: 'v1.53: Domain Monitoring Beyond RDAP',
    date: 'October 15, 2026',
    readTime: '3 min read',
    excerpt:
      "Domains under .il, .io, .co, .me, .us, .it, .ru, .jp, .cn and .se used to sit in Error because their registries don't run RDAP. They're now read over WHOIS, with a parser per registry; TLDs whose expiry dates can't be read at all are refused with the reason.",
  },
  {
    file: 'v1-52-projects',
    slug: 'v1-52-projects',
    kind: 'release',
    title: 'v1.52: Projects, One Group per Client',
    date: 'October 15, 2026',
    readTime: '4 min read',
    excerpt:
      'Monitors can now be grouped into projects, typically one per client. The monitor list shows each project as its own section, and every alert for a project\'s monitor starts with its name, like "[Acme] nightly-backup missed its ping", so you know whose problem it is before opening the app.',
  },
  {
    file: 'v1-49-to-v1-51-landing-page',
    slug: 'v1-49-to-v1-51-landing-page',
    kind: 'release',
    title: 'v1.49–v1.51: A Landing Page That Shows Its Work',
    date: 'October 15, 2026',
    readTime: '4 min read',
    excerpt:
      "An outside review of the landing page found its second half generic. The testimonials section is gone; the page now shows the real Telegram message and webhook payload a missed cron job sends, and checkmeup.net's own live 90-day status with per-day uptime and response times.",
  },
  {
    file: 'v1-48-redesign',
    slug: 'v1-48-redesign',
    kind: 'release',
    title: 'v1.48: A Calmer Checkmeup, Light and Dark',
    date: 'October 15, 2026',
    readTime: '3 min read',
    excerpt:
      'A redesign of every screen in both themes: a top navigation bar, one list for all six monitor types, a dashboard that opens with a plain sentence about what needs you, 48-hour check strips in every list, and a rebuilt public status page.',
  },
  {
    file: 'v1-46-to-v1-47-alert-channels',
    slug: 'v1-46-to-v1-47-alert-channels',
    kind: 'release',
    title: 'v1.46–v1.47: SMS Failures You Can Actually See',
    date: 'October 14, 2026',
    readTime: '4 min read',
    excerpt:
      "Alert SMS that a carrier dropped after Twilio accepted them used to vanish silently. Now each failure is recorded on the channel, and if SMS was the alert's only delivery, a fallback email goes out. Also: masked Slack and webhook URLs, real 404s for missing pages, and interval changes that take effect right away.",
  },
  {
    file: 'v1-42-to-v1-45-monitor-forms',
    slug: 'v1-42-to-v1-45-monitor-forms',
    kind: 'release',
    title: 'v1.42–v1.45: One Form Instead of Fourteen',
    date: 'September 10, 2026',
    readTime: '4 min read',
    excerpt:
      'Four releases that mostly moved code around: the 14 monitor create/edit screens each carried their own copy of the same form, and they now share five components instead — 3,821 lines of views down to 1,206. Pulling that logic into the open made it testable, which is how a bug surfaced where editing a JSON assertion and walking away without saving left the app showing the unsaved edit as if it were real.',
  },
  {
    file: 'dns-record-monitoring-guide',
    slug: 'dns-record-monitoring-guide',
    kind: 'article',
    title: 'DNS Record Monitoring: Beyond the Uptime Check',
    date: 'August 25, 2026',
    readTime: '5 min read',
    excerpt:
      "A hijacked DNS record can serve a convincing fake site while your uptime check stays green — it never looks at the record itself. Here's how DNS record monitoring works, why a single lookup isn't the whole internet's view, and the one habit that turns it back into noise.",
    description:
      'A hijacked DNS record can serve a fake site while uptime checks stay green. How DNS record monitoring works, and how to keep it from turning into noise.',
  },
  {
    file: 'why-monitoring-matters',
    slug: 'why-monitoring-matters',
    kind: 'article',
    title: 'Why Monitoring Matters: Finding Out Before Users Do',
    date: 'August 3, 2026',
    readTime: '5 min read',
    excerpt:
      "Monitoring doesn't keep anything running — what it buys you is the interval between something breaking and you finding out, which is where almost all of an outage's cost lives. A look at the failure modes that produce no signal at all, and why a green dashboard proves less than most people assume.",
    description:
      "Monitoring doesn't keep anything running; it shortens the time between a failure and you knowing. The failures that make no noise, and what green proves.",
  },
  {
    file: 'v1-41-zombie-job-detection',
    slug: 'v1-41-zombie-job-detection',
    kind: 'release',
    title: 'v1.41: Zombie (Stuck) Job Detection',
    date: 'July 30, 2026',
    readTime: '3 min read',
    excerpt:
      'Cron monitors can now report when a job starts, not just when it finishes — set a max run duration and get a distinct alert if a job gets stuck instead of waiting on the normal missed-ping schedule. Fully opt-in: monitors that never send a start ping behave exactly as they always have.',
  },
  {
    file: 'v1-40-status-page-grid-layout',
    slug: 'v1-40-status-page-grid-layout',
    kind: 'release',
    title: 'v1.40: A Second Status Page Layout',
    date: 'July 29, 2026',
    readTime: '3 min read',
    excerpt:
      'A "grid" layout joins the original single-column status page: monitors two-per-row with active incidents and history moved into a sidebar, better suited to pages watching a lot of monitors. Pick it per page from the edit form — every existing page keeps the classic layout by default.',
  },
  {
    file: 'v1-37-to-v1-39-dns-monitoring',
    slug: 'v1-37-to-v1-39-dns-monitoring',
    kind: 'release',
    title: 'v1.37–v1.39: DNS Record Monitoring',
    date: 'July 29, 2026',
    readTime: '3 min read',
    excerpt:
      "A sixth monitor type: watch a hostname's DNS record and get alerted the moment it changes or stops resolving — pin an expected value, or let the first check capture a baseline and alert on any later drift. Plus a couple of small internal-workflow and Codacy fixes from the two versions since the last post.",
  },
  {
    file: 'v1-33-to-v1-36-uptime-form-restyle',
    slug: 'v1-33-to-v1-36-uptime-form-restyle',
    kind: 'release',
    title: 'v1.33–v1.36: The Design Reaches the Uptime Form',
    date: 'July 27, 2026',
    readTime: '3 min read',
    excerpt:
      'Four versions since the last release notes, and the quietest stretch yet: the uptime monitor form catches up to the Claude Design rollout, two new posts landed on the blog, and a chunk of time went into how I work on this codebase rather than what it does for you.',
  },
  {
    file: 'port-tcp-monitoring-guide',
    slug: 'port-tcp-monitoring-guide',
    kind: 'article',
    title: 'Port (TCP) Monitoring: Beyond HTTP Checks',
    date: 'July 27, 2026',
    readTime: '5 min read',
    excerpt:
      "A database or mail server has no URL to poll, so an outage there is invisible to an uptime check. Here's how port (TCP) monitoring works, why a successful connection isn't proof the service behind it is healthy, and how to use it to catch a port that's exposed when it shouldn't be.",
    description:
      "Databases and mail servers have no URL to poll. How port (TCP) monitoring works, what an open port doesn't prove, and how to catch one exposed by mistake.",
  },
  {
    file: 'ssl-certificate-expired-story',
    slug: 'ssl-certificate-expired-story',
    kind: 'article',
    title: 'The 3 AM SSL Cert That Cost Us a Client',
    date: 'July 24, 2026',
    readTime: '5 min read',
    excerpt:
      "A red NET::ERR_CERT_DATE_INVALID screen, six hours of bounced checkout traffic, and a client who found out from their own customers before we did. The renewal cron job had been silently failing for two months — here's why that failure is invisible until the exact second it isn't, and what actually catches it.",
    description:
      'A renewal cron failed silently for two months, then the cert expired at 3 AM and checkout broke for six hours. Why that failure hides, and what catches it.',
  },
  {
    file: 'v1-13-release-notes',
    slug: 'v1-13-release-notes',
    kind: 'release',
    title: 'v1.13: New Homepage, New License',
    date: 'July 1, 2026',
    readTime: '3 min read',
    excerpt:
      "Two unrelated changes landed together: the homepage got a redesign — live product mockups instead of static screenshots, a proper status-page card, a cleaned-up footer — and Checkmeup's license changed from MIT to the Business Source License.",
  },
  {
    file: 'v1-18-release-notes',
    slug: 'v1-18-release-notes',
    kind: 'release',
    title: "v1.18: A Public API, So Status Isn't Dashboard-Only",
    date: 'July 3, 2026',
    readTime: '4 min read',
    excerpt:
      "Every monitor's status has lived exclusively behind the dashboard's session cookie until now. v1.18 adds a read-only public API — generate a key, send it as X-API-Key, and pull a monitor's status into a CI pipeline, an internal ops dashboard, or a physical status LED. Cron pings can now also carry their own tags (a build number, a pass/fail state), which show up in both the dashboard and the API.",
  },
  {
    file: 'v1-26-release-notes',
    slug: 'v1-26-release-notes',
    kind: 'release',
    title: 'v1.26: A Real 404 Page, and a Name Fixed',
    date: 'July 10, 2026',
    readTime: '4 min read',
    excerpt:
      "This release closes out the SEO thread from v1.24–v1.25 with Article schema for richer blog search results, gives Checkmeup an actual 404 page instead of a blank screen for any dead link, and fixes something that's been quietly inconsistent since launch: how the product's own name gets capitalized.",
  },
  {
    file: 'v1-7-release-notes',
    slug: 'v1-7-release-notes',
    kind: 'release',
    title: 'v1.7: Webhook Alerts',
    date: 'June 22, 2026',
    readTime: '3 min read',
    excerpt:
      "A third alert channel alongside Telegram and email: webhooks. Point a monitor at any HTTPS endpoint and get a signed POST when it goes down or recovers — Slack incoming webhooks, PagerDuty, your own scripts, whatever's listening. Plus the SSRF hardening that comes with letting users hand you an arbitrary URL to request.",
  },
  {
    file: 'v1-3-release-notes',
    slug: 'v1-3-release-notes',
    kind: 'release',
    title: 'v1.3: Email Alerts and Keyword Monitoring',
    date: 'June 19, 2026',
    readTime: '2 min read',
    excerpt:
      "A second alert channel alongside Telegram, and uptime monitors that can now look inside the response body. Here's what shipped in v1.3.",
  },
  {
    file: 'v1-8-release-notes',
    slug: 'v1-8-release-notes',
    kind: 'release',
    title: 'v1.8: Domain Expiry Monitoring',
    date: 'June 23, 2026',
    readTime: '3 min read',
    excerpt:
      'A fourth monitor type alongside cron, uptime, and SSL: domain registration expiry. Checkmeup already told you when a certificate was about to lapse — now it tells you when the domain underneath it is about to disappear entirely, which is rarer but a lot worse when it happens.',
  },
  {
    file: 'pomodoro-and-checkmeup',
    slug: 'pomodoro-and-checkmeup',
    kind: 'article',
    title: 'The Cube on My Desk: Pomodoro and Building Checkmeup',
    date: 'June 18, 2026',
    readTime: '3 min read',
    excerpt:
      "A small black cube sits next to my keyboard and ticks down in 25-minute blocks. Here's what the Pomodoro Technique is, and how it shaped the pace Checkmeup got built at.",
    description:
      'A small cube on my desk ticks down in 25-minute blocks. What the Pomodoro Technique is, and how it set the pace Checkmeup was built at.',
  },
  {
    file: 'v1-29-release-notes',
    slug: 'v1-29-release-notes',
    kind: 'release',
    title: 'v1.29: Incidents Get Limits, Finally',
    date: 'July 12, 2026',
    readTime: '4 min read',
    excerpt:
      "A follow-up to last release's incident management feature: resolved incidents now age out after 90 days, and every resource that could previously grow without limit — active incidents, incident updates, maintenance windows, API keys — now has a flat, plan-independent ceiling.",
  },
  {
    file: 'v1-17-release-notes',
    slug: 'v1-17-release-notes',
    kind: 'release',
    title: 'v1.17: Analytics, With a Consent Banner That Counts',
    date: 'July 3, 2026',
    readTime: '3 min read',
    excerpt:
      'Checkmeup.net has had zero visibility into its own landing page and signup funnel since launch. This release adds Google Analytics 4 behind a Google Tag Manager container — but nothing loads, and nothing tracks, until a visitor actually clicks Accept. Also: a Paddle webhook race that could leave a Hobby org stuck with no way to see upgrade options.',
  },
  {
    file: 'v1-12-release-notes',
    slug: 'v1-12-release-notes',
    kind: 'release',
    title: 'v1.12: Port (TCP) Monitoring',
    date: 'July 1, 2026',
    readTime: '3 min read',
    excerpt:
      'A fifth monitor type: raw TCP connect checks for non-HTTP services like mail servers and databases. It ships with a twist beyond the usual up/down check — an expected-state toggle that turns it into a security check for ports that should stay closed.',
  },
  {
    file: 'v1-6-release-notes',
    slug: 'v1-6-release-notes',
    kind: 'release',
    title: 'v1.6: Multi-Channel Notifications',
    date: 'June 21, 2026',
    readTime: '3 min read',
    excerpt:
      'The headline: monitors no longer alert through a single org-wide Telegram chat and email address — you can connect multiple channels and choose which ones each monitor uses. Plus keyword monitoring is free on every plan now, and a couple of security hardening fixes.',
  },
  {
    file: 'v1-20-to-v1-22-redesign',
    slug: 'v1-20-to-v1-22-redesign',
    kind: 'release',
    title: 'v1.20–v1.22: The Redesign Reaches Everything Else',
    date: 'July 6, 2026',
    readTime: '6 min read',
    excerpt:
      'The marketing site got a visual redesign a while back. This week that same look — near-black/near-white neutrals, translucent surfaces, a lower-contrast palette — finally reached the parts people actually use every day: the dashboard, the public status page, and Settings/Status Pages Admin/Monitors/Maintenance.',
  },
  {
    file: 'v1-19-release-notes',
    slug: 'v1-19-release-notes',
    kind: 'release',
    title: 'v1.19: SMS Alerts, and a Real Downgrade',
    date: 'July 4, 2026',
    readTime: '5 min read',
    excerpt:
      "Checkmeup can text you now — an eighth alert channel via Twilio, with a monthly credit quota so a flapping monitor can't turn into a surprise bill. And a smaller but overdue fix: downgrading your plan used to leave every over-limit monitor running forever. Now it actually enforces the new limit — pausing the newest monitors first and keeping your oldest ones active.",
  },
  {
    file: 'v1-27-release-notes',
    slug: 'v1-27-release-notes',
    kind: 'release',
    title: 'v1.27: Checks That Reuse Connections',
    date: 'July 10, 2026',
    readTime: '3 min read',
    excerpt:
      "No new feature in this one — it's a quieter release focused on how the monitoring engine itself runs: reusing connections instead of rebuilding them on every check, giving the database pool an explicit ceiling instead of an implicit default, and cutting duplicated code out of the alert-delivery and notification-settings paths.",
  },
  {
    file: 'v1-23-release-notes',
    slug: 'v1-23-release-notes',
    kind: 'release',
    title: "v1.23: Turning the Audit on Checkmeup's Own Codebase",
    date: 'July 9, 2026',
    readTime: '4 min read',
    excerpt:
      "This release is small on the surface — you'll now see a version number in the dashboard sidebar and the site footer. Most of the week actually went somewhere you won't see directly: an audit of Checkmeup's own codebase that found six real maintainability problems, and fixed all of them.",
  },
  {
    file: 'v1-2-release-notes',
    slug: 'v1-2-release-notes',
    kind: 'release',
    title: 'v1.2: Annual Billing and Real Upgrade Buttons',
    date: 'June 18, 2026',
    readTime: '2 min read',
    excerpt:
      'Annual plans at two months free, and the Billing page finally does something other than say "coming soon." Plus: why you still can\'t pay me yet.',
  },
  {
    file: 'three-days-to-mvp',
    slug: 'three-days-to-mvp',
    kind: 'article',
    title: 'Three Days to MVP: How We Built Checkmeup',
    date: 'June 16, 2026',
    readTime: '3 min read',
    excerpt:
      "A cron monitor, uptime checks, SSL alerts, status pages, and billing — shipped in roughly 20 hours over three days. Here's what actually happened.",
    description:
      'A cron monitor, uptime checks, SSL alerts, status pages and billing, shipped in about 20 hours over three days. What actually happened.',
  },
  {
    file: 'v1-28-release-notes',
    slug: 'v1-28-release-notes',
    kind: 'release',
    title: 'v1.28: Manual Incidents Arrive, Plus a Real Toggle',
    date: 'July 12, 2026',
    readTime: '4 min read',
    excerpt:
      'Two things landed on status pages today: you can now declare and narrate an incident yourself instead of only ever showing raw monitor up/down, and paid plans can finally turn off the "Powered by Checkmeup" footer — a feature the pricing page has quietly promised for a while.',
  },
  {
    file: 'v1-16-release-notes',
    slug: 'v1-16-release-notes',
    kind: 'release',
    title: 'v1.16: Change Plan Without Leaving the Page',
    date: 'July 2, 2026',
    readTime: '3 min read',
    excerpt:
      'Upgrading, downgrading, and cancelling an existing subscription now happens directly on the Billing page — no re-running checkout to switch tiers. Paddle rejections surface with the actual reason instead of a generic error, and a scheduled cancellation is tracked properly instead of just hoping the UI catches up.',
  },
  {
    file: 'v1-9-release-notes',
    slug: 'v1-9-release-notes',
    kind: 'release',
    title: 'v1.9: Status Badges and Assertion Checks',
    date: 'June 23, 2026',
    readTime: '4 min read',
    excerpt:
      'Two additions this release: embeddable SVG status badges you can drop into a README or site footer to show live monitor status, and structured assertion checks for uptime monitors — JSON field assertions and a response-time threshold that actually fail the check, not just annotate it.',
  },
  {
    file: 'v1-5-release-notes',
    slug: 'v1-5-release-notes',
    kind: 'release',
    title: 'v1.5: Cleanup, a Plan-Limit Bug, and More Tests',
    date: 'June 20, 2026',
    readTime: '2 min read',
    excerpt:
      'Another quiet release: one real bug in how uptime monitor intervals get enforced on edit, a dark-mode color fix, a data-fetching cleanup across the dashboard, and more backend test coverage.',
  },
  {
    file: 'domain-expiry-monitoring',
    slug: 'domain-expiry-monitoring',
    kind: 'article',
    title: 'Why Domains Expire Silently, Not From Negligence',
    date: 'July 16, 2026',
    readTime: '5 min read',
    excerpt:
      "A lapsed domain is rarely negligence — it's usually a boring, invisible failure chain: an expired card, a reminder email nobody reads, a WHOIS contact that hasn't been checked in years. Here's why it happens, what it actually costs when it does, and how to catch it before a client does.",
    description:
      'Domains lapse through boring failures: an expired card, an unread reminder, a stale WHOIS contact. Why it happens, what it costs, and how to catch it.',
  },
  {
    file: 'checkmeup-vs-competitors',
    slug: 'checkmeup-vs-healthchecks-uptimerobot-cronitor',
    kind: 'article',
    title: 'Checkmeup vs. Healthchecks, UptimeRobot, Cronitor',
    date: 'July 4, 2026',
    readTime: '6 min read',
    excerpt:
      "Three tools get named in every 'what should I use to monitor my cron jobs' thread: Healthchecks.io, UptimeRobot, and Cronitor. Here's a feature-by-feature, dollar-by-dollar look at where Checkmeup fits — and where it honestly doesn't win yet.",
    description:
      "Checkmeup vs Healthchecks.io, UptimeRobot and Cronitor for cron and uptime monitoring: features and prices side by side, and where Checkmeup doesn't win.",
  },
  {
    file: 'v1-30-release-notes',
    slug: 'v1-30-release-notes',
    kind: 'release',
    title: 'v1.30: Teaching Bing That Checkmeup Has Content',
    date: 'July 14, 2026',
    readTime: '4 min read',
    excerpt:
      "Bing Webmaster Tools flagged the homepage for a missing h1 heading. It wasn't a markup bug — every page really does render one. The actual problem was that checkmeup is a client-rendered app, and Bingbot doesn't run JavaScript. This release fixes that properly, and closes an older, related gap along the way.",
  },
  {
    file: 'v1-11-release-notes',
    slug: 'v1-11-release-notes',
    kind: 'release',
    title: 'v1.11: Alert Noise Reduction',
    date: 'June 25, 2026',
    readTime: '3 min read',
    excerpt:
      'Configure how many consecutive failures must occur before the first alert fires — across all four monitor types. Pair it with the existing per-incident alert cap to control both the start and the volume of notifications for any incident.',
  },
  {
    file: 'v1-24-to-v1-25-seo-foundations',
    slug: 'v1-24-to-v1-25-seo-foundations',
    kind: 'release',
    title: 'v1.24–v1.25: Teaching Search Engines Checkmeup',
    date: 'July 10, 2026',
    readTime: '5 min read',
    excerpt:
      "Checkmeup is a single-page app, which meant every route — the homepage, pricing, every blog post — served the exact same page title and description to search engines. These two releases fix that: real per-page metadata, a sitemap, a robots.txt, and FAQ rich-result markup, plus one honest gap that's still open.",
  },
  {
    file: 'v1-15-release-notes',
    slug: 'v1-15-release-notes',
    kind: 'release',
    title: 'v1.15: Billing Moved to Paddle',
    date: 'July 2, 2026',
    readTime: '3 min read',
    excerpt:
      "Checkmeup's payment provider changed from LemonSqueezy to Paddle — full replacement, not a dual-provider setup. Nothing changes for existing subscribers except where the invoice comes from; the trigger was a payment processor's domain-verification checklist that surfaced Paddle as the better long-term fit before real revenue depended on either provider.",
  },
  {
    file: 'v1-1-release-notes',
    slug: 'v1-1-release-notes',
    kind: 'release',
    title: 'v1.1: Themes, Legal Pages, FAQ, and Feedback',
    date: 'June 18, 2026',
    readTime: '2 min read',
    excerpt:
      "Theme toggle, Terms & Privacy (finally), a proper FAQ page, and an in-app feature-suggestion form. Here's what changed in v1.1.",
  },
  {
    file: 'v1-4-release-notes',
    slug: 'v1-4-release-notes',
    kind: 'release',
    title: 'v1.4: A Testing Pass (and the Bugs It Found)',
    date: 'June 19, 2026',
    readTime: '2 min read',
    excerpt:
      'No new features this time. Just a deliberate pass at backend test coverage — and the handful of real bugs, including one security issue, that it turned up along the way.',
  },
  {
    file: 'v1-10-release-notes',
    slug: 'v1-10-release-notes',
    kind: 'release',
    title: 'v1.10: Slack Alerts',
    date: 'June 24, 2026',
    readTime: '3 min read',
    excerpt:
      'Slack joins Telegram, email, and webhooks as a native alert channel. Paste an Incoming Webhook URL, pick which monitors it covers, and get formatted Block Kit messages when something goes down and when it recovers.',
  },
  {
    file: 'v1-31-release-notes',
    slug: 'v1-31-release-notes',
    kind: 'release',
    title: 'v1.31: The Sitemap Was Lying, Plus Structured Data',
    date: 'July 15, 2026',
    readTime: '4 min read',
    excerpt:
      "v1.30 shipped build-time prerendering to fix a Bing crawl notice. Rather than trust the release notes, I actually curled the live site afterward — and found two more crawler-facing bugs hiding right behind it: the sitemap pointed a real blog post at a URL that didn't exist, and every marketing page cost crawlers an unnecessary redirect. Both fixed, plus Organization and SoftwareApplication structured data on the homepage.",
  },
  {
    file: 'v1-14-release-notes',
    slug: 'v1-14-release-notes',
    kind: 'release',
    title: 'v1.14: Legal Housekeeping',
    date: 'July 2, 2026',
    readTime: '3 min read',
    excerpt:
      'No new monitor type this time. A domain-verification checklist for a payment processor turned into a real audit of the Terms of Service and Privacy Policy — both had gaps worth fixing, plus a standalone Refund Policy page. A couple of Docker/registry housekeeping items rode along.',
  },
  {
    file: 'v1-0-release-notes',
    slug: 'v1-0-release-notes',
    kind: 'release',
    title: "v1.0 — First Release: What Shipped and What's Next",
    date: 'June 16, 2026',
    readTime: '2 min read',
    excerpt:
      "Checkmeup v1.0 is live. Here's the complete changelog, what we deliberately left out, and a foggy look at the road ahead.",
  },
  {
    file: 'cron-job-monitoring-guide',
    slug: 'cron-job-monitoring-guide',
    kind: 'article',
    title: 'Cron Job Monitoring: Catch Silent Failures',
    date: 'July 17, 2026',
    readTime: '6 min read',
    excerpt:
      "Cron jobs fail silently, not loudly. Here's how cron job monitoring — the dead man's switch — works, how to set it up in three steps, and what to check before trusting a tool with it.",
    description:
      "Cron jobs fail silently. How cron job monitoring (a dead man's switch) works, how to set it up in three steps, and what to check before trusting a tool.",
  },
  {
    file: 'uptime-monitoring-guide',
    slug: 'uptime-monitoring-guide',
    kind: 'article',
    title: 'Uptime Monitoring: Beyond the Status Code',
    date: 'July 18, 2026',
    readTime: '6 min read',
    excerpt:
      "A 200 response isn't proof a page actually works. Here's how uptime monitoring works, why content checks beat status codes alone, and what to look for in a tool before you trust it.",
    description:
      "A 200 response doesn't prove a page works. How uptime monitoring works, why content checks beat status codes alone, and what to look for in a tool.",
  },
  {
    file: 'v1-32-release-notes',
    slug: 'v1-32-release-notes',
    kind: 'release',
    title: 'v1.32: Configurable Uptime Checks',
    date: 'July 18, 2026',
    readTime: '3 min read',
    excerpt:
      'Uptime monitors were always a GET request, a 10-second timeout, and exactly HTTP 200. v1.32 makes all three configurable per monitor, and collapses the existing optional response-time SLA field into the actual request timeout instead of leaving two overlapping knobs.',
  },
]
