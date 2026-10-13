# Search indexing and URL form of the public site

Date: 2026-10-13

Release notes stay at `/blog/<slug>` but carry `noindex, follow` in their prerendered HTML, so dozens of thin per-version pages don't drag down the whole site's ranking. `/changelog` is the one indexed page for them: every release, newest first, each with its own anchor and a link to the full notes. Articles stay indexed and are the only posts on `/blog` and in `sitemap.xml`. `robots.txt` doesn't disallow release notes, because a crawler blocked from a page never sees its `noindex`.

Every public URL has one form, without a trailing slash (except `/`). The server answers a trailing-slash URL with a 301 to the bare form, keeping the query string. The canonical and `og:url` are built from the path alone, without query, hash or trailing slash.

Public routes (`/`, `/pricing`, `/docs`, `/faq`, `/about`, `/terms`, `/privacy`, `/refund`, `/blog`, `/changelog` and every post) are prerendered to static HTML at build time, so crawlers that don't run JavaScript (Bingbot among them) see the content and the head tags. `scripts/prerender.mts` renders the real components with `vue/server-renderer` under `vite-node` and writes `dist/<route>/index.html`, which the Go server serves as is. The browser then renders the app fresh over it rather than hydrating. Signed-in app routes stay client-rendered.

## Considered options

- A 301 from each release note to its `/changelog` anchor: offered as the alternative to `noindex`. With `noindex`, the full notes stay readable at their old URLs.
- `vite-plugin-ssg` or a server-side rendering server: more machinery than a dozen static routes need.
- Hand-written fallback HTML in `index.html`: fixes the homepage only.
