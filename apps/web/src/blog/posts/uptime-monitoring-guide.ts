import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "The first sign a site is down is rarely a dashboard — it's a client's message asking why their checkout page is blank, or a support ticket from someone who tried three times before giving up. By the time that message arrives, the outage has already cost something: a sale, a support hour, a little trust. Uptime monitoring exists to close that gap — to be the thing that notices before the person who matters does.",
  },
  {
    type: 'h2',
    text: 'What uptime monitoring actually does',
  },
  {
    type: 'p',
    text: 'An uptime monitor polls a URL on a fixed interval and records what comes back — status code, response time, how long the request took to fail if it failed at all. Two things happen from there: the check gets logged, building a history you can look back on, and a status change gets compared against what the monitor already believed. Go from up to down, or back, and that transition is what triggers an alert — not every check, just the ones that change the story.',
  },
  {
    type: 'h2',
    text: '"The server responded" isn\'t the same as "the site works"',
  },
  {
    type: 'p',
    text: "A status-code-only check has a real blind spot: a server that's badly broken can still return 200. A misconfigured reverse proxy serving a generic error page, a WordPress site stuck on its own maintenance screen, an API returning an empty JSON body where a real payload should be — all of these look perfectly healthy to a monitor that only checks whether something answered. The fix is checking for a signal that only shows up when the page is actually right: a string that should always appear in the real response (an account balance, a product name, a piece of your own markup), or, for an API, a specific field in the JSON body holding the value you expect. A monitor that can assert on content, not just status code, catches the failures a plain ping never will.",
  },
  {
    type: 'h2',
    text: 'Avoiding false alarms without missing real ones',
  },
  {
    type: 'p',
    text: "Two knobs matter most, and they pull against each other. The check interval decides how fast you find out — 1 minute catches an outage almost as it starts, 30 minutes might let one run most of its length before anyone's told. The failure threshold decides how sure the monitor needs to be before it wakes you up: a single slow response or one dropped packet on an otherwise healthy host shouldn't page anyone, so a real monitor waits for a second consecutive failure before calling it down. Tune the interval to how much downtime is actually tolerable for that specific URL — a marketing page can live with 30 minutes, a checkout flow can't — and trust the consecutive-failure check to absorb the noise in between.",
  },
  {
    type: 'h2',
    text: 'Setting up uptime monitoring in three steps',
  },
  {
    type: 'h3',
    text: '1. Monitor the URL that actually matters, not just the homepage',
  },
  {
    type: 'p',
    text: 'A homepage returning 200 tells you almost nothing about whether checkout, login, or the API your app depends on are working. Add a monitor per critical path — the pages or endpoints where an outage actually costs something — instead of one monitor standing in for the whole site.',
  },
  {
    type: 'h3',
    text: '2. Add a content check, not just a status code',
  },
  {
    type: 'p',
    text: "Pick a string that only appears when the page rendered correctly, or a JSON field with a value you can assert on, and set the monitor to check for it. This is the single highest-leverage step in the whole setup — it's the difference between catching a real outage and catching a technically-200 page that's actually broken.",
  },
  {
    type: 'h3',
    text: "3. Route the alert somewhere you'll see it inside the interval you picked",
  },
  {
    type: 'p',
    text: "A 1-minute check interval is wasted if the alert lands in an inbox you check twice a day. Match the channel to how fast you need to react — Telegram or SMS for anything client-facing, email for something you can afford to notice later — and make sure recovery alerts are on too, so you know when it's actually safe to stop worrying, not just when it broke.",
  },
  {
    type: 'h2',
    text: 'What to look for in an uptime monitoring tool',
  },
  {
    type: 'ul',
    items: [
      'Content or keyword assertions, not just a status-code check — a 200 on a broken page should still count as down',
      'A configurable check interval, short enough for the paths that actually matter to you',
      'A consecutive-failure threshold before alerting, so one flaky check does not turn into constant noise',
      'An incident log with start time, resolution time, and duration — not just a live status badge that forgets everything once it flips back to green',
      'A public status page, if anyone outside your own team — a client, a customer — needs to see it too, without them having to ask you first',
    ],
  },
  {
    type: 'h2',
    text: 'The mistake worth avoiding',
  },
  {
    type: 'p',
    text: 'Setting up a monitor and never checking what "up" actually means to it is the easiest way for this to quietly fail you. Confirm what counts as success for your specific check — plenty of legitimate responses (a 201 from a create endpoint, a 204 from a delete) aren\'t a plain 200, and a tool that only accepts an exact 200 will happily alarm on a working API. Know the rule your monitor is actually applying before you trust the green badge.',
  },
  {
    type: 'divider',
  },
  {
    type: 'p',
    text: "Checkmeup's uptime monitor covers all of this — configurable check intervals down to 1 minute, keyword and JSON-body assertions so a technically-200 page doesn't slip through, a two-strike failure threshold to avoid flapping, full incident and response-time history, and a public status page if clients need to see it too. The Hobby plan is free for up to 10 monitors, no credit card required.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
