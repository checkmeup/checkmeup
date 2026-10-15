import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "When domain monitoring shipped in v1.8, I wrote that RDAP doesn't cover every TLD, that the WHOIS fallback for the rest had been cut, and that it was a reasonable follow-up \"if a real domain hits either gap\". This week one did: a monitor for an Israeli .co.il domain sat in Error, saying the domain wasn't found in the registry. The domain exists. Its registry just doesn't speak RDAP.",
  },
  {
    type: 'h2',
    text: 'What was going wrong',
  },
  {
    type: 'p',
    text: 'Checkmeup reads a domain\'s expiry date over RDAP, the structured successor to WHOIS, through rdap.org, which forwards each query to the right registry. For a TLD whose registry runs no RDAP service, rdap.org answers 404, exactly what it answers for a domain nobody registered. So the monitor reported "not found", which was wrong, and it would have stayed that way forever.',
  },
  {
    type: 'p',
    text: "It wasn't only .il. Going through the common TLDs, .io, .co, .me, .it, .ru, .jp, .cn, .se and .us all had the same gap, and .io and .co are exactly the kind of domains developers and agencies use.",
  },
  {
    type: 'h2',
    text: 'WHOIS for the registries without RDAP',
  },
  {
    type: 'p',
    text: 'Domains under .il, .io, .co, .me, .us, .it, .ru, .jp (including .co.jp), .cn and .se are now looked up over WHOIS, the old plain-text protocol those registries still answer. Everything else goes through RDAP as before.',
  },
  {
    type: 'p',
    text: "WHOIS has no standard format. Israel's registry writes the expiry as \"validity: 11-01-2029\", day first; Russia's says \"paid-till\"; Japan's answers in Japanese unless asked for English. So each supported TLD has its own parser, written against real answers from that registry, both for a registered domain and for one that doesn't exist, and those answers are kept as test fixtures. I deliberately didn't write a generic parser that hunts for anything date-shaped near the word \"expir\". A parser that guesses wrong could tell you a live domain is fine when it's about to lapse, and that's worse than an honest error.",
  },
  {
    type: 'p',
    text: 'If a registry changes its format one day, only that TLD\'s lookups fail, and they fail as an error, never as a wrong date. A domain the registry doesn\'t know is also an error, never "expired".',
  },
  {
    type: 'h2',
    text: 'Registries that publish no expiry date',
  },
  {
    type: 'p',
    text: "Some registries don't publish an expiry date anywhere. .de, .eu, .ch, .li and .at answer WHOIS but leave the date out, and .es only answers WHOIS from an allow list of addresses. There's nothing to monitor there, so adding a domain under one of them is now refused with that reason, instead of accepting a monitor that could only ever show an error. If you already have one, it shows the same explanation.",
  },
  {
    type: 'p',
    text: "And for the TLDs that are left, RDAP's 404 no longer claims the domain wasn't found. It says what's actually known: either the domain isn't registered, or its registry doesn't publish RDAP.",
  },
  {
    type: 'p',
    text: "Existing monitors pick this up on their own at their next daily check; there's nothing to re-create. The docs page now lists which TLDs go through WHOIS and which can't be monitored.",
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'ul',
    items: [
      'The changelog listed releases from the same day oldest first. Posts were sorted by date only, which has no time of day, so same-day ties fell back to whatever order they happened to be written in. They now go by version.',
      'The toggle on each project group in the monitor list is now a proper chevron that turns when the group collapses, instead of a tiny text arrow.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: "If your domain sits under a TLD that's still unsupported and its registry does publish an expiry date somewhere, send me the TLD through the feedback form in Settings and I'll take a look. The decision record for domain lookups, with the options I turned down, is on GitHub with the rest.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
