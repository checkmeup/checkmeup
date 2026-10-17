import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: 'Most developers who need to know when a certificate or a domain expires don\'t want a monitoring account. They want the answer once, right now, and they search for "ssl checker" or "domain expiry check" to get it. Until this release, nothing on checkmeup.net answered that without a sign-up, so those searches never found it.',
  },
  {
    type: 'h2',
    text: 'Two free checkers, no account',
  },
  {
    type: 'p',
    text: "There are now two public tools: /tools/ssl-checker and /tools/domain-expiry-checker. Type a hostname and the SSL checker shows when its certificate expires, how many days are left and who issued it. Type a domain and the domain checker shows its expiry date, days left and registrar. Nothing is stored, and you don't need an account.",
  },
  {
    type: 'p',
    text: "They aren't a separate, simplified implementation. They run the same TLS check and the same RDAP-then-WHOIS lookup a real monitor runs, including the per-registry WHOIS parsers from v1.53. So a domain the checker can read is one a monitor can watch, and a domain it refuses, such as one under .de, is refused with the same reason. Under each result there's an offer to monitor it, because a date you checked once is only useful until you forget it.",
  },
  {
    type: 'p',
    text: "Two limits keep this from turning into a free bulk-lookup API. Each visitor gets 5 checks a minute per tool. The domain checker also has a cap of 30 lookups a minute across everyone, because its WHOIS queries go to the same registries that paying domain monitors depend on, and some of those registries throttle hard. Past either limit, the page asks you to try again in a minute. For scripted lookups there's the public API and real monitors.",
  },
  {
    type: 'p',
    text: "One trade-off I left in on purpose: an expired or untrusted certificate shows the TLS error, not its expiry date and issuer. The checker verifies the chain the way an SSL monitor does, and an SSL monitor treats a broken chain as the problem. I may change that for the tool if it turns out people mostly use it to look at certificates that are already broken. DNS lookup, TCP port and cron expression tools are possible next steps, but they aren't built yet.",
  },
  {
    type: 'h2',
    text: 'Rate limits were shared by everyone',
  },
  {
    type: 'p',
    text: 'Building the tools\' per-visitor limit turned up a real bug. The app runs behind a reverse proxy, and every per-IP rate limit was keyed by the address of the connection it saw, which was always the proxy. So "5 sign-ups an hour per IP" actually meant 5 sign-ups an hour for the whole internet. With today\'s traffic nobody hit it, but it would have shown up at the worst possible moment, the day a launch post brought people in.',
  },
  {
    type: 'p',
    text: "Limits now use the visitor's address from the X-Forwarded-For header, but only when the request comes from the proxy, which always appends the real address itself. A request that connects directly can't forge its way into someone else's bucket.",
  },
  {
    type: 'h2',
    text: 'One bad certificate no longer reddens 90 days',
  },
  {
    type: 'p',
    text: "A status page's 90-day strip combines every monitor on the page. SSL and domain monitors keep no daily history, only their current state, so one failing certificate counted as failing on all 90 days and turned the whole strip red. They now count only for today. The strip shows what actually happened on each day, and today still shows the failing certificate.",
  },
  {
    type: 'h2',
    text: 'Also these releases',
  },
  {
    type: 'ul',
    items: [
      'The footer of a status page now reads "Powered by Checkmeup · Get a free status page like this" instead of a bare link, tagged so I can see whether status pages bring in sign-ups. Hobby includes one status page, so "free" is literal. Pages with branding hidden are unchanged.',
      "HEAD requests used to get a 405 while GET worked, so link checkers, SEO crawlers and other people's uptime monitors saw checkmeup.net as broken. Every page now answers HEAD too.",
      'Articles have their own meta descriptions of 155 characters or fewer, so search results no longer cut them off mid-sentence.',
      'Sign-ups are now counted in analytics, if you accepted cookies. Until now I had no way of knowing where sign-ups came from.',
      'The app shows its version in a footer that links to the changelog, as the landing page does, instead of hiding it in the account menu.',
      'Pricing tables show included features as a labelled icon instead of a tiny check mark, and a Billing usage bar that is nearly full is now amber instead of the red used for "down".',
      'Blog stays highlighted in the nav while you read a post, and the site footer was rearranged: the free tools are in the first column, and Status and GitHub are at the bottom.',
      'The docs, FAQ and pricing page now cover projects, and the FAQ lists which domain TLDs can be monitored.',
      'Two new articles: why some domains hide their expiry date (RDAP vs WHOIS), and how to keep fifty client monitors manageable.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'Next on the roadmap are team management and two-factor authentication, both ready to build. If a free check would save you a sign-up somewhere else, say which one through the feedback form in Settings. The requirements for the free tools, with what I deliberately left out, are on GitHub with the rest.',
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
