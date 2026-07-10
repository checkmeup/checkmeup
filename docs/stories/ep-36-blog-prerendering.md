# EP-36: Prerender blog posts for social link previews

`apps/web` is a pure client-rendered Vue SPA with no server-side rendering. The `useSeo()` composable (shipped as part of the SEO technical-fixes work, PRs #67/#68/#70) sets each route's `<title>`/`<meta description>`/OG tags via `@unhead/vue`, but that's a DOM mutation that only happens once a browser actually runs the page's JavaScript. Google's crawler does execute JS, so search indexing and rich results are already correct — but social link-preview bots (Slack, Twitter/X, LinkedIn, Discord, iMessage) fetch the raw HTML response and never run any JS, so they still see the single generic site-wide title/description/image baked into `index.html` for every URL, whichever blog post it actually is. Release-notes and comparison posts get shared directly (Reddit, Twitter, etc.), so this is a real, visible-if-cosmetic gap, not a theoretical one.

Fixing it means the raw HTML response has to differ per `/blog/:slug` route — a build-time prerender step (e.g. `vite-plugin-ssg`, a custom prerender script) or SSR, not another `useHead()` call. Real architectural change to the build/deploy pipeline, deliberately deferred rather than bundled into the rest of the SEO work. Which approach to take is an open question — see [decision backlog](../decisions/backlog.md).

---

### US-3601: Prerender blog post pages to static HTML at build time

**As a** person sharing a Checkmeup blog post link, **I want** the raw HTML response to already contain that post's actual title, description, and OG image, **so that** Slack/Twitter/LinkedIn/Discord previews show the real post instead of the generic site-wide card.

**Estimate:** 3 h (rough — depends on which approach the decision backlog entry lands on)

**Acceptance criteria:**

- [ ] Every `/blog/:slug` route's raw HTML response (fetched without executing any JS — e.g. `curl`, or an actual platform link-unfurl debugger) contains that post's real `<title>`, `<meta name="description">`, and OG tags (`og:title`, `og:description`, `og:image`)
- [ ] `/blog` (the list page) also serves a real title/description in the raw response, not just individual posts
- [ ] Marketing pages already covered by `useSeo()` (home, pricing, docs, FAQ, etc.) are out of scope — Google already indexes those correctly since its crawler executes JS; only the social-preview gap is being closed here
- [ ] Build/deploy pipeline change is documented the same way the sitemap-generation step was (`package.json` build script, and the Dockerfile fix in PR #69 — a prerender step added to one but not the other silently breaks in exactly the way sitemap generation did)
- [ ] `sitemap.xml` generation and the `robots.txt`/`noindex` behavior from the earlier SEO work keep working unchanged
