import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-24-to-v1-25-seo-foundations',
  title: 'v1.24–v1.25: Teaching Search Engines Checkmeup Exists',
  date: 'July 10, 2026',
  readTime: '5 min read',
  excerpt:
    "Checkmeup is a single-page app, which meant every route — the homepage, pricing, every blog post — served the exact same page title and description to search engines. These two releases fix that: real per-page metadata, a sitemap, a robots.txt, and FAQ rich-result markup, plus one honest gap that's still open.",
  content: [
    {
      type: 'p',
      text: "Building the product came first, understandably. Nobody had gone back and checked whether Checkmeup could actually be found by the people searching for it. Turns out the answer was mostly no: it's a client-rendered Vue app with no server-side rendering, and every single route — the homepage, /pricing, /docs, twenty-five blog posts — was serving the identical static page title and meta description baked into index.html. A search engine looking at a blog post about SSL monitoring and the homepage saw the same generic tag. No sitemap existed either, so there was no map handed to crawlers of what pages even exist beyond whatever they stumbled onto by following links.",
    },
    {
      type: 'h3',
      text: 'v1.24 — every page gets its own identity',
    },
    {
      type: 'p',
      text: "The fix is a small composable, useSeo(), built on @unhead/vue, called once per page with a title, description, and canonical path. Nine marketing pages (home, pricing, docs, FAQ, about, terms, privacy, refund, blog) now each set their own tags, and the blog post page sets its title and description reactively from whichever post is actually loaded — so a shared link to a specific post shows that post's real title, not the site-wide default.",
    },
    {
      type: 'p',
      text: "Alongside that: a robots.txt that allows the marketing and blog pages while disallowing everything behind login — dashboard, monitors, status page admin, billing, settings, and the auth flow itself. None of those pages have anything to gain from being indexed, and a couple of them would leak internal URL structure for no benefit. And a sitemap.xml, generated automatically at build time from the router's static routes plus every blog post's filename — not hand-maintained, so it can't quietly go stale as new posts get published.",
    },
    {
      type: 'h3',
      text: 'v1.25 — the FAQ page learns to talk to Google directly, and a quiet indexing gap gets closed',
    },
    {
      type: 'p',
      text: "The FAQ page already had well-structured question-and-answer content — it just wasn't marked up as such. v1.25 adds FAQPage JSON-LD schema built from that same data, which is what makes a page eligible for Google's expandable FAQ rich results directly in search listings, instead of just a blue link.",
    },
    {
      type: 'p',
      text: 'The second half is smaller but was worth catching: because the blog post route is client-rendered, visiting an unknown slug — a typo, a dead link, an old URL — still returns HTTP 200 with a "post not found" page, not a real 404. Left alone, that\'s a page search engines could index as thin, low-value content. It now sets a noindex meta tag whenever the requested slug doesn\'t resolve to a real post.',
    },
    {
      type: 'h2',
      text: 'What this does not fix',
    },
    {
      type: 'p',
      text: "Google's crawler executes JavaScript, so it sees the per-page titles and descriptions fine. Slack, Twitter/X, and most link-preview bots don't — they read whatever the server returns before any JavaScript runs, which is still the generic site-wide index.html tags. Share a blog post link in Slack today and the preview card is accurate for the site but not for that post. Fixing that properly means prerendering the blog to static HTML at build time, which is a real architectural change for a pure SPA, not a metadata tweak — deliberately deferred rather than rushed into this pair of releases.",
    },
    {
      type: 'h2',
      text: 'Also this release',
    },
    {
      type: 'ul',
      items: [
        "Added a shared vitest setup that installs @unhead/vue's head plugin globally for component tests, since useHead()/useSeoMeta() throw when called outside an app that has it installed — every existing view test needed this, not just the new SEO code.",
        'Verified both the FAQ schema and the noindex fix by actually mounting the components and inspecting document.head, not just trusting the types — @unhead/vue debounces its DOM writes by a tick, which the first version of that check missed.',
      ],
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: "Next up on this thread is content, not more technical plumbing: the blog is almost entirely release notes, which are useful to existing users but don't target anyone actually searching for a cron or uptime monitoring tool. Full commit history and the architecture decisions behind Checkmeup are public on GitHub if you want the details behind any of this.",
    },
    { type: 'signature', text: '— Andrew' },
  ],
}
