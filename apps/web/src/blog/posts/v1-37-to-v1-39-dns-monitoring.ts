import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "Three versions since the last release notes, and the headline one is the real story: a sixth monitor type. Every other monitor in Checkmeup watches whether something is reachable. DNS monitoring watches whether it's still the *right thing* being reached — a record that quietly changes (a hijack, a migration gone sideways, a typo in a DNS provider's dashboard) doesn't fail the way a downed server does. Nothing goes offline. Traffic just starts going somewhere else, and the first sign is usually a client noticing before you do. DNS monitors close that gap: watch a hostname's resolved record and alert the moment it's not what it should be.",
  },
  {
    type: 'h3',
    text: 'Adding a DNS monitor',
  },
  {
    type: 'p',
    text: 'Give it a name, a hostname, and a record type (A, AAAA, CNAME, MX, TXT, or NS). An optional expected value lets you pin exactly what the record should resolve to; leave it blank and the first successful check captures whatever it currently resolves to as the baseline instead. Either way, every later check compares against that same value — pinned or captured, the comparison logic doesn\'t care which. DNS monitors count toward the aggregate monitor limit alongside the other five types — "cron + uptime + SSL + domain + port + DNS" — and plug into maintenance windows and status pages the same way.',
  },
  {
    type: 'h3',
    text: 'Two ways to use it',
  },
  {
    type: 'p',
    text: "Pin a value when you already know what's correct — an A record that should always point at one IP, an SPF/TXT record that should never change without you doing it deliberately. You get an alert from the moment the monitor is created if reality doesn't match. Leave it blank when you don't want to hand-type the current value, or when the point is just \"tell me if this ever moves\" rather than asserting a specific answer — a CNAME pointing at a third-party CDN, an NS delegation you don't control directly. Both modes end up comparing against the same stored value; the only difference is whether you typed it or the monitor captured it for you.",
  },
  {
    type: 'h3',
    text: 'How the check works',
  },
  {
    type: 'p',
    text: 'A real DNS lookup for the configured record type, with a multi-value answer (multiple A records, multiple MX hosts) sorted before comparing — so a resolver returning the same set in a different order on two consecutive checks never reads as a change. A lookup failure (NXDOMAIN, SERVFAIL, timeout) is recorded and alerted on separately from a value mismatch, so the alert text says the right thing: "can\'t resolve" when the name itself is broken, "changed from X to Y" when it resolves fine but to something else. Those are different problems with different fixes, and conflating them into one generic "down" would make the alert less useful right when you need it most.',
  },
  {
    type: 'h3',
    text: 'Alerts',
  },
  {
    type: 'p',
    text: 'A change alert states the old value and the new one side by side, not just "something\'s wrong" — the whole point is knowing what it changed to without reaching for `dig` yourself. Editing a monitor to accept a new value (or clearing it back to blank, re-arming baseline mode) is how you tell Checkmeup an intentional change is fine going forward. Recovery alerts, the per-incident cap, and every notification channel work exactly as they do for the other five monitor types.',
  },
  {
    type: 'h3',
    text: "What's not in this release",
  },
  {
    type: 'p',
    text: 'MX records compare on hostname only — the numeric preference isn\'t part of the stored value, so re-ranking two mail servers without changing which hosts are listed won\'t trigger an alert. Full RFC-faithful comparison would need the pair to be part of the compared value, which felt like more precision than the actual use case ("did my mail exchange change") needed for a first release.',
  },
  {
    type: 'h2',
    text: 'Also this cycle',
  },
  {
    type: 'ul',
    items: [
      "Hours logging (an internal workflow thing, not a product feature — how I track my own time building this) now ships in the same PR as the work it's logging, instead of a follow-up docs PR, and rounds to the nearest 15 minutes instead of a whole hour.",
      'A Codacy false positive on the blog: ESLint flagged a post that named an HTML element in its own prose as unsanitized markup. Fixed the trigger, not the check — the rule is still doing its job everywhere it should.',
      'The DNS monitor sidebar link went missing from the initial release — caught within the hour and fixed same-day.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: "That's six monitor types now. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
