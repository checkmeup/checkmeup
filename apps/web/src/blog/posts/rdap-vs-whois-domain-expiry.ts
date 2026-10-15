import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "Ask a domain monitoring tool when example.com expires and you get a date in a second. Ask about an .io, .co.il or .de domain and, depending on the tool, you might get an error, a blank field, or a confident \"domain not found\" for a domain that's clearly live. That isn't the tool being flaky. Registration data for different top-level domains lives behind different protocols, and some registries don't publish an expiry date at all.",
  },
  {
    type: 'h2',
    text: "Where a domain's expiry date comes from",
  },
  {
    type: 'p',
    text: 'Every top-level domain is run by a registry: Verisign for .com, ISOC-IL for .il, DENIC for .de. The registry is the source of truth for when a registration ends. There are two ways to ask it.',
  },
  {
    type: 'p',
    text: 'RDAP (Registration Data Access Protocol) is the modern one: an HTTPS API that returns JSON, with the expiry as an event named "expiration". IANA keeps a bootstrap file that maps each TLD to its registry\'s RDAP server, and services like rdap.org use it to forward a query to the right place. Registries for generic TLDs like .com, .net, .org and the newer ones such as .app or .dev are required by their contracts with ICANN to run RDAP, so for those the answer is always one request away.',
  },
  {
    type: 'p',
    text: "WHOIS is the old one: you open a TCP connection to port 43, send the domain name, and read back plain text. There's no standard format. Every registry writes its own.",
  },
  {
    type: 'h2',
    text: 'Why some TLDs have no RDAP',
  },
  {
    type: 'p',
    text: "Country-code TLDs (.il, .it, .jp, .de and the rest) aren't bound by ICANN's registry contracts, so whether they run RDAP is up to each registry. Many do: .uk, .nl, .fr, .ca and .au all answer RDAP. A good number don't, including some you'd expect to, like .io, .co and .me, which are country codes even though most people use them as generic names.",
  },
  {
    type: 'p',
    text: 'For a domain under one of those, an RDAP lookup through a bootstrap redirector fails with a 404, exactly the response it gives for a domain nobody registered. A tool that reads that 404 as "not registered" will tell you your live domain doesn\'t exist.',
  },
  {
    type: 'h2',
    text: 'What WHOIS looks like when there is no RDAP',
  },
  {
    type: 'p',
    text: 'The registries without RDAP still answer WHOIS, each in its own shape. The expiry for an Israeli domain looks like this:',
  },
  {
    type: 'code',
    lang: 'text',
    text: 'domain:       isoc.org.il\nassigned:     11-01-1996\nvalidity:     11-01-2029\nregistrar name: Israel Internet Association ISOC-IL',
  },
  {
    type: 'p',
    text: 'Day first, labeled "validity". Russia\'s registry says "paid-till: 2027-09-30T21:00:00Z". Italy\'s says "Expire Date: 2026-12-31". Japan\'s answers in Japanese unless you append "/e" to the query, and second-level names like .co.jp put the date inside a status line: "[State] Connected (2027/03/31)". The .io, .co, .me and .us registries happen to use the same layout as gTLD WHOIS, "Registry Expiry Date: 2027-05-01T00:00:02Z".',
  },
  {
    type: 'p',
    text: 'So reading an expiry date over WHOIS means knowing each registry\'s format. A parser that just looks for anything date-shaped near the word "expir" will eventually read the wrong line, or read 01-11 as January 11 when the registry meant November 1. For a monitor that\'s worse than failing: a wrong date can tell you a domain is fine weeks after it lapsed.',
  },
  {
    type: 'h2',
    text: 'The registries that publish no expiry at all',
  },
  {
    type: 'p',
    text: "Some registries don't put an expiry date in WHOIS. DENIC (.de), EURid (.eu) and nic.at answer WHOIS but leave the date out; DENIC's whole answer for denic.de is two lines, the domain and \"Status: connect\". For .de in particular there's nothing to leave out: a registration runs until it's cancelled, and whether it continues depends on your registrar being paid, not on a date at the registry. Others don't answer WHOIS for just anyone: SWITCH (.ch and .li) replies \"Requests of this client are not permitted\" and points to a web form, and Spain's .es registry only answers addresses on an allow list.",
  },
  {
    type: 'p',
    text: "For those domains, no external monitor can warn you about an expiry date, because there isn't one to read. The risk moves to your registrar account: the card on file, the renewal or billing emails, and who receives them.",
  },
  {
    type: 'h2',
    text: "How to check your own domain's expiry",
  },
  {
    type: 'p',
    text: 'For a TLD with RDAP, one request does it; look for the "expiration" event in the response:',
  },
  {
    type: 'code',
    lang: 'bash',
    text: 'curl -sL https://rdap.org/domain/example.com | jq \'.events[] | select(.eventAction == "expiration")\'',
  },
  {
    type: 'p',
    text: "For one without, ask its registry's WHOIS server directly. whois.iana.org tells you which server that is:",
  },
  {
    type: 'code',
    lang: 'bash',
    text: 'whois -h whois.iana.org il | grep ^whois     # → whois.isoc.org.il\nwhois -h whois.isoc.org.il example.co.il | grep validity\nwhois -h whois.jprs.jp "example.jp/e" | grep "Expires on"',
  },
  {
    type: 'table',
    headers: ['TLD', 'Where the expiry comes from'],
    rows: [
      ['.com, .net, .org, .app, .dev and other gTLDs', 'RDAP'],
      ['.uk, .nl, .fr, .ca, .au, .in, .br and many ccTLDs', 'RDAP'],
      ['.io, .co, .me, .us', 'WHOIS, "Registry Expiry Date"'],
      ['.il', 'WHOIS, "validity" (DD-MM-YYYY)'],
      ['.it', 'WHOIS, "Expire Date"'],
      ['.ru', 'WHOIS, "paid-till"'],
      ['.jp', 'WHOIS with /e, "[Expires on]" or "[State]"'],
      ['.cn', 'WHOIS, "Expiration Time"'],
      ['.se', 'WHOIS, "expires"'],
      ['.de, .eu, .at', 'Not published'],
      ['.ch, .li, .es', 'WHOIS only for allow-listed clients'],
    ],
  },
  {
    type: 'p',
    text: "That table is a snapshot from October 2026. Registries add RDAP over time, and IANA's bootstrap file at data.iana.org/rdap/dns.json is the current list.",
  },
  {
    type: 'h2',
    text: 'How Checkmeup handles it',
  },
  {
    type: 'p',
    text: "Checkmeup's domain monitor uses RDAP wherever a registry runs it, and WHOIS for the TLDs in the table above that publish an expiry date there, with a separate parser for each registry, written against that registry's real answers. A domain the registry doesn't know shows as an error, never as expired. Adding a domain under a TLD whose expiry can't be read is refused with the reason, rather than accepted as a monitor that can only ever show an error. The reasoning behind those choices is in a decision record on GitHub, and the full story of why domains lapse in the first place is in an earlier post on this blog.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
