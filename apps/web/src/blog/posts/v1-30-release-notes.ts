import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-30-release-notes',
  title: "v1.30: Teaching Bing (and Every Non-JS Crawler) That Checkmeup's Pages Have Content",
  date: 'July 14, 2026',
  readTime: '4 min read',
  excerpt:
    "Bing Webmaster Tools flagged the homepage for a missing h1 heading. It wasn't a markup bug — every page really does render one. The actual problem was that checkmeup is a client-rendered app, and Bingbot doesn't run JavaScript. This release fixes that properly, and closes an older, related gap along the way.",
  content: [
    {
      type: 'p',
      text: "A Bing crawl report came back with a \"missing H1\" notice on the homepage. My first instinct was to go check HeroSection.vue — and there it was, a perfectly normal h1 heading reading \"Know before your client does.\", right where it should be. The component was fine. What wasn't fine is that checkmeup ships index.html to every visitor as an empty shell: just an empty root element and a script tag, nothing else. The real markup only exists after the browser runs Vue. Google's crawler executes JavaScript, so it's never had a problem with this. Bingbot apparently doesn't, reliably — so it requested the homepage, got back an empty div, and correctly reported no H1, because there genuinely wasn't one in what it received.",
    },
    {
      type: 'h3',
      text: 'The actual fix: static HTML at build time',
    },
    {
      type: 'p',
      text: "The honest fix for a single-page app serving empty HTML to crawlers is to stop serving empty HTML — render the real content into the file before it ever leaves the build. checkmeup now does that for every public page: the homepage, pricing, docs, FAQ, about, terms, privacy, refund policy, the blog list, and every individual blog post. A small script runs after the normal production build, spins up the real Vue app for each of those routes in Node (not a browser), renders it to a string, and writes the result to disk as that route's actual index.html.",
    },
    {
      type: 'p',
      text: "The nice part is that the server side of this needed zero changes. checkmeup's Go backend already serves static files by checking whether the requested path exists on disk before falling back to the SPA shell — that logic was written for asset caching, not for this, but it does exactly the right thing for a prerendered page too. Write real HTML to the right path, and it's served as-is.",
    },
    {
      type: 'p',
      text: "Getting there took more wiring than the idea sounds like it should: rendering real .vue components outside a browser meant pulling the route table apart from the parts of the router that assume a real browser exists (session history, an authenticated fetch call on every navigation), and finding the two places — a theme preference, a cookie-consent flag — that read from window/localStorage the instant their modules loaded, not just when actually used. Both now check whether those globals exist first and fall back sensibly when they don't.",
    },
    {
      type: 'h3',
      text: 'This also closes a two-week-old gap',
    },
    {
      type: 'p',
      text: 'The v1.24–v1.25 release notes ended with an honest "what this does not fix": per-page titles and descriptions were live for Google, but link-preview bots — Slack, Twitter/X, Discord — never run JavaScript either, so sharing a specific blog post still showed the generic site-wide card. That was already tracked as its own future story, scoped narrowly to just the blog. Once I was rebuilding this mechanism for Bing\'s sake anyway, there was no reason to solve it twice — the same prerendering step captures each page\'s real title, description, and social preview image at build time, blog posts included. One fix, two crawlers.',
    },
    {
      type: 'h2',
      text: 'What this does not change',
    },
    {
      type: 'p',
      text: "Everything behind login — the dashboard, monitor pages, status page admin, billing, settings — stays exactly as it was: rendered entirely in the browser, after sign-in, never crawled by anything. There was never a reason to prerender those, and there still isn't. A real user's browser still runs the normal app the instant it loads; the prerendered HTML is what a crawler sees on the way in, not a replacement for the interactive app itself.",
    },
    {
      type: 'h2',
      text: 'Also this release',
    },
    {
      type: 'ul',
      items: [
        'While wiring the new build step into the production build, discovered that the full build had been silently broken for a while — 24 type errors that never showed up in CI because the test workflow never actually ran a full production build, only lint and tests. Fixed all of them, which also caught a real bug: a couple of numeric form fields (port number, max response time) were using a modifier that looked like it converted text input to a number but silently did nothing, so those fields were sending text where the API expected numbers.',
        "Split the router's route table into its own file so the new build script can build a router instance from the real route definitions without dragging in the parts that only make sense in an actual browser.",
      ],
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: 'Next up is content rather than more technical plumbing, same as last time — the blog is still mostly release notes. Full commit history and the reasoning behind this and every other architecture decision are public on GitHub, including the write-up of exactly why this needed a Node-side rendering context instead of just a build flag.',
    },
    { type: 'signature', text: '— Andrew' },
  ],
}
