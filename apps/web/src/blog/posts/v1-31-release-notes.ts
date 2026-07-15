import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-31-release-notes',
  title: 'v1.31: The Sitemap Was Lying, and Some Structured Data',
  date: 'July 15, 2026',
  readTime: '4 min read',
  excerpt:
    "v1.30 shipped build-time prerendering to fix a Bing crawl notice. Rather than trust the release notes, I actually curled the live site afterward — and found two more crawler-facing bugs hiding right behind it: the sitemap pointed a real blog post at a URL that didn't exist, and every marketing page cost crawlers an unnecessary redirect. Both fixed, plus Organization and SoftwareApplication structured data on the homepage.",
  content: [
    {
      type: 'p',
      text: "v1.30 fixed the big, obvious SEO gap: checkmeup is a client-rendered app, so Bingbot was seeing an empty page. Prerendering fixed that, the release notes said so, and I moved on. What I hadn't done was actually go check — curl the live site the way a crawler would, not just trust that a passing build meant a working production deploy. Doing that turned up two smaller bugs that the prerendering work itself didn't cause, but that were sitting right next to it, undermining it.",
    },
    {
      type: 'h3',
      text: "The sitemap was advertising a page that doesn't exist",
    },
    {
      type: 'p',
      text: "sitemap.xml is generated at build time from two independent places: the prerender script reads each blog post's actual `slug` field (the same one the router uses), while the sitemap generator — written earlier, before prerendering existed — derived each post's URL from its filename instead. Those matched for every post except one: the competitor-comparison article's file is named checkmeup-vs-competitors.ts, but its slug field is checkmeup-vs-healthchecks-uptimerobot-cronitor. The sitemap had been telling Google to index /blog/checkmeup-vs-competitors this whole time — a URL nothing prerenders, so it silently served the homepage's fallback shell instead of a real 404. Meanwhile the actual article, at its real URL, wasn't in the sitemap at all. Fixed by making the sitemap generator read the same slug field the rest of the app already treats as the source of truth, with a hard failure instead of a silent fallback if a post is ever missing one.",
    },
    {
      type: 'h3',
      text: "Every marketing URL paid for a redirect it didn't need",
    },
    {
      type: 'p',
      text: "Separately: every prerendered route ships as a directory containing its own index.html — dist/pricing/index.html, dist/blog/some-post/index.html, and so on. checkmeup's Go server was passing those bare directory paths straight to Go's http.ServeFile, which has a built-in behavior of 301-redirecting any request that doesn't already end in a slash. So a crawler requesting exactly the URL in the sitemap — /pricing, no trailing slash — got a redirect before it ever saw the real page. Harmless for ranking (crawlers follow redirects fine), but an avoidable hop on literally every indexed URL. Fixed by resolving the directory's index.html and serving that file directly instead of handing the bare directory to ServeFile.",
    },
    {
      type: 'h2',
      text: 'Also this release',
    },
    {
      type: 'ul',
      items: [
        "Organization and SoftwareApplication JSON-LD added to the homepage — same pattern as the FAQPage schema on /faq and the Article schema on every blog post, just applied to the front door. The SoftwareApplication entry's pricing offers are generated from the same plans data the pricing page renders, so it can't quietly drift out of sync with what checkmeup actually charges.",
      ],
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: "With this, the technical side of SEO is in a place I'm comfortable calling done for now — proper metadata, a sitemap that tells the truth, prerendered content crawlers can actually read, and structured data on the pages that benefit from it. What's next isn't more markup, it's off-page: getting checkmeup in front of people through the Product Hunt launch, developer directories, and more of the comparison/how-to content that's worked so far. The full commit history and every architectural decision behind it is on GitHub, linked from the About page.",
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
