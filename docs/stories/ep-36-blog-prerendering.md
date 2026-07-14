# EP-36: Prerender public routes for crawlers and social link previews

**Status: Done** — resolved 2026-07-14 by [ADR-037](../decisions/037-prerender-public-routes.md), scope expanded from blog-only (as originally written below) to all public routes.

`apps/web` is a pure client-rendered Vue SPA with no server-side rendering. The `useSeo()` composable (shipped as part of the SEO technical-fixes work, PRs #67/#68/#70) sets each route's `<title>`/`<meta description>`/OG tags via `@unhead/vue`, but that's a DOM mutation that only happens once a browser actually runs the page's JavaScript. This story originally assumed Google's crawler executing JS meant marketing-page indexing was already fine, and scoped the fix to just social link-preview bots (Slack, Twitter/X, LinkedIn, Discord, iMessage) missing per-post OG tags on `/blog/:slug`. That assumption didn't hold: Bing flagged the homepage itself for a missing `<h1>` in its raw (non-JS-executed) crawl, showing the same empty-shell problem affects marketing pages too, not just blog social cards.

Fixed by prerendering every public/indexable route — not just blog posts — to static HTML at build time via a custom `vite-node` + `vue/server-renderer` script (`apps/web/scripts/prerender.mts`). See [ADR-037](../decisions/037-prerender-public-routes.md) for the full mechanism.

---

### US-3601: Prerender public pages to static HTML at build time

**As a** person sharing a Checkmeup link, or a search crawler indexing the site, **I want** the raw HTML response to already contain the real title, description, OG image, and heading content, **so that** social previews and non-JS-executing crawlers see the real page instead of an empty shell or generic site-wide card.

**Acceptance criteria:**

- [x] Every `/blog/:slug` route's raw HTML response (fetched without executing any JS) contains that post's real `<title>`, `<meta name="description">`, and OG tags (`og:title`, `og:description`, `og:image`)
- [x] `/blog` (the list page) also serves a real title/description in the raw response, not just individual posts
- [x] Marketing pages (home, pricing, docs, faq, about, terms, privacy, refund) are now in scope too — each serves a real `<h1>` and per-route title/OG tags in the raw HTML response, closing the Bing "H1 missing" notice as well as the original social-preview gap
- [x] Build/deploy pipeline change is documented the same way the sitemap-generation step was — `apps/web/package.json`'s `build` script and the root `Dockerfile`'s frontend build stage were updated together, in the same commit, specifically to avoid the PR #69-style bug where a step added to one but not the other silently breaks
- [x] `sitemap.xml` generation and the `robots.txt`/`noindex` behavior from the earlier SEO work keep working unchanged
