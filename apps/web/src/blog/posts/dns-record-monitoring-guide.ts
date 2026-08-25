import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "An attacker who hijacks your DNS and points your domain at their own server doesn't need to take your site down — they can serve a convincing fake login page that returns a perfectly healthy 200. Your uptime monitor stays green the whole time, because it's checking whether *something* answers at the resolved address, not whether that address is still the one you set. The DNS record itself — the thing an uptime check never looks at — is where that kind of hijack actually shows up first.",
  },
  {
    type: 'h2',
    text: 'What DNS record monitoring actually does',
  },
  {
    type: 'p',
    text: 'A DNS monitor resolves a hostname\'s record on a fixed interval — A, AAAA, CNAME, MX, TXT, or NS — and compares what comes back against either a value you supply up front or a baseline it captured on its first successful check. Multi-value answers (several A records, several MX hosts) are sorted before comparing, so a resolver returning them in a different order on two consecutive checks never reads as a change. A lookup that fails outright — NXDOMAIN, SERVFAIL, timeout — is recorded as a distinct failure reason, kept separate from a mismatch, so the alert can tell you "this domain stopped resolving" instead of lumping it in with "this domain now resolves somewhere else."',
  },
  {
    type: 'h2',
    text: "One vantage point, not the whole internet's view",
  },
  {
    type: 'p',
    text: "A DNS check only sees what the monitoring server's own resolver returns at the moment it asks. That's not necessarily what a visitor in a different region sees — split-horizon DNS, geo-targeted answers, and ordinary caching mean two different resolvers can legitimately return different values for the same record at the same time. A DNS monitor built on a single lookup, from a single location, is telling you what one vantage point saw, not confirming the record is consistent everywhere it's queried. It's also comparing resolved values, not validating them: a TXT or MX record can change to something syntactically valid but operationally wrong, and the monitor has no way to know that a new SPF string is broken or a new mail host doesn't actually accept mail — only that the text is different from what it expected.",
  },
  {
    type: 'h2',
    text: 'Setting up DNS monitoring in three steps',
  },
  {
    type: 'h3',
    text: '1. Decide between an expected value and a baseline',
  },
  {
    type: 'p',
    text: "If you already know the record's correct value — your CDN's documented IPs, your verified MX host — enter it as the expected value; the monitor alerts on any mismatch from the moment you create it, which is the right mode for a security-sensitive record you want pinned. If you don't have that value memorized, leave it blank: the first successful lookup becomes the baseline, and the monitor alerts the first time a later check disagrees with it. Baseline mode is the easier default, but it means the very first check locks in whatever the record happens to be at that instant — create the monitor when you're confident the record is already correct, not mid-migration.",
  },
  {
    type: 'h3',
    text: '2. Pick the record type the risk actually lives on',
  },
  {
    type: 'p',
    text: "An A or AAAA record monitor catches your domain resolving to an IP you didn't put there — the classic hijack. An MX monitor catches mail silently rerouting through a host you don't control, which an uptime check would never notice since nothing about your website changes. A CNAME or NS monitor catches a broken or stolen delegation upstream of both. Monitor the record type where an unnoticed change would actually hurt, not just the one that's easiest to set up.",
  },
  {
    type: 'h3',
    text: '3. Match the interval and failure threshold to how fast a change matters',
  },
  {
    type: 'p',
    text: "A hijacked A record is worth catching within minutes, not hours — pick the shortest interval your plan allows for anything security-sensitive. Leave the alert-after-N-failures threshold at zero for that same reason: a single mismatched check on a record monitor is meaningful in a way a single dropped ping isn't, since DNS answers don't flap the way network connections do.",
  },
  {
    type: 'h2',
    text: 'What to look for in a DNS monitoring tool',
  },
  {
    type: 'ul',
    items: [
      "Both an expected-value mode and a baseline mode, not just one — you need the first when you know the correct value, the second when you only know it shouldn't move",
      'A distinct failure state for "record didn\'t resolve" versus "record resolved to something else" — a hijack and an outage need different responses',
      "Multi-value answers compared as a set, so answer-order jitter on round-robin A or MX records doesn't generate false alerts",
      'Support for the record types that actually carry risk for you — at minimum A/AAAA and MX, not just A',
      'A change log showing old value and new value, not just a live status badge that forgets what changed once it goes green',
    ],
  },
  {
    type: 'h2',
    text: 'The mistake worth avoiding',
  },
  {
    type: 'p',
    text: "A DNS monitor in baseline or expected-value mode stays down after a legitimate change until you go back and update it — migrating hosts, switching mail providers, or moving to a new CDN will all trip the alert, and it won't clear on its own. The mistake is treating that as noise: muting the monitor, or getting used to seeing it red after every planned migration, instead of updating the expected value the same day you make the change. Once you've trained yourself to ignore this particular monitor because it's \"probably just last week's migration again,\" an actual hijack looks identical to that noise — which defeats the one thing DNS monitoring exists to catch.",
  },
  {
    type: 'divider',
  },
  {
    type: 'p',
    text: "Checkmeup's DNS monitor supports both expected-value and baseline modes across A, AAAA, CNAME, MX, TXT, and NS records, with old-value/new-value change history and the same consecutive-failure and per-incident alert caps as its other monitor types. The Hobby plan is free for up to 10 monitors of any type, no credit card required.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
