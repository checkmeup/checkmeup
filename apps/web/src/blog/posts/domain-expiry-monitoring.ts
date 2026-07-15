import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "If you manage other people's sites, you've either had this happen or you know someone it happened to: the domain just stops working. No warning, no error in your deploy pipeline, no monitoring alert — because most monitoring watches the site, not the thing the site's address depends on. The first sign is usually a client's Slack message, and it's rarely polite.",
  },
  {
    type: 'h2',
    text: 'Why domains actually expire',
  },
  {
    type: 'p',
    text: "It's almost never someone deciding not to renew. It's a chain of small, boring failures, any one of which is enough on its own: the card on file at the registrar expired eight months ago and nobody noticed because autorenew failed silently instead of loudly. The WHOIS/registrant contact email is an old employee's inbox, or the client's own address that they never check, or — worse — an email address that itself lives on the domain that's about to expire, so the renewal notice can't even be delivered. ICANN requires registrars to send expiration notices, but a required email and a read email are different things; spam filters exist, inboxes get abandoned, forwarding rules break.",
  },
  {
    type: 'p',
    text: "None of this is a competence problem. It's a coverage gap — the renewal reminder depends entirely on channels the registrar controls and you don't: their email deliverability, their timing, whichever contact address happened to be on file when the domain was registered, possibly years before you ever touched the project.",
  },
  {
    type: 'h2',
    text: "The math nobody does until it's too late",
  },
  {
    type: 'p',
    text: "A domain costs somewhere around $10–15 a year. That's such a small number that losing track of it feels absurd in hindsight — but the cost was never the renewal fee, it's what happens in the gap. DNS resolution for a suspended domain can break within hours of expiry, taking the site, and if it's hosted there, email, down with it. If nobody catches it during the registrar's grace or redemption period, the domain becomes available again — and anyone, including a squatter who specifically watches for expiring domains with existing backlinks and traffic, can register it out from under you. Getting it back at that point isn't a renewal anymore; it's a negotiation, if it's possible at all.",
  },
  {
    type: 'h2',
    text: 'What actually catches this',
  },
  {
    type: 'p',
    text: "Not the registrar's own reminder — that's the exact channel that already failed to reach anyone, or this wouldn't be a problem. Not a calendar reminder either, past two or three domains; it doesn't scale, and it's yet another thing that depends on someone remembering to set it up correctly in the first place. What actually works is checking the registration data independently, on a schedule, from a source that doesn't route through the registrant's own inbox.",
  },
  {
    type: 'p',
    text: "That's what a domain monitor is for. Checkmeup's does a straightforward RDAP lookup — the structured, standardized successor to WHOIS — once a day for every domain you add, and compares the registration's actual expiry date against three thresholds: 30 days out, 14 days out, and 7 days out, plus a final alert if it actually expires. Each threshold only fires once, so renewing resets it cleanly instead of leaving a stale warning around or requiring you to manually dismiss anything. Alerts go out over whichever channels you've already got wired up for everything else — Telegram, Slack, email, and SMS on paid plans — so a domain nobody's watching becomes a domain that pings you the same way a downed cron job or an expiring SSL cert does.",
  },
  {
    type: 'h2',
    text: "A short checklist, if you're doing this by hand today",
  },
  {
    type: 'ul',
    items: [
      "Put your own email as the technical/admin contact on client domains, not the client's — you're the one who needs the notice, and their inbox is one more point of failure you don't control.",
      'Keep one password-manager entry per domain with the registrar, expiry date, and renewal method — a five-minute setup cost that turns "which of our 40 domains needs attention" into a search.',
      "Turn on the registrar's autorenew and monitor independently — autorenew fails silently often enough (expired card, changed billing email) that it shouldn't be the only layer.",
      "Review the full domain list on a quarterly cadence, not just when something breaks — cheap insurance against the ones nobody's actively thinking about.",
    ],
  },
  {
    type: 'p',
    text: "None of this is exotic. It's the same instinct that makes uptime monitoring obvious — you wouldn't wait for a client to tell you their site is down — applied to the one dependency that sits underneath the site itself.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
