# Search indexing and URL form of the public site

Date: 2026-11-05

Release notes stay at `/blog/<slug>` but carry `noindex, follow` in their prerendered HTML, so dozens of thin per-version pages don't drag down the whole site's ranking. `/changelog` is the one indexed page for them: every release, newest first, each with its own anchor and a link to the full notes. Articles stay indexed and are the only posts on `/blog` and in `sitemap.xml`. `robots.txt` doesn't disallow release notes, because a crawler blocked from a page never sees its `noindex`.

Every public URL has one form, without a trailing slash (except `/`). The server answers a trailing-slash URL with a 301 to the bare form, keeping the query string. The canonical and `og:url` are built from the path alone, without query, hash or trailing slash.

Public routes (`/`, `/pricing`, `/docs`, `/faq`, `/about`, `/terms`, `/privacy`, `/refund`, `/blog`, `/changelog`, the free checkers under `/tools/`, the monitor type pages such as `/cron-monitoring`, the `/alternatives/` pages and every post) are prerendered to static HTML at build time, so crawlers that don't run JavaScript (Bingbot among them) see the content and the head tags. `apps/web/scripts/prerender.mts` renders the real components with `vue/server-renderer` under `vite-node` and writes `dist/<route>/index.html`, which the Go server serves as is. The browser then renders the app fresh over it rather than hydrating. Signed-in app routes stay client-rendered.

A path no route of the web app matches gets a 404 with a prerendered `404.html` (`noindex`), not the app shell with a 200: a soft 404 gets crawled as a duplicate page, and an uptime monitor on a missing page reads it as up. The prerender step writes the app's route patterns to `dist/spa-routes.json` from the same `routes.ts` the router uses, so the server never keeps its own copy; `/blog/:slug` is left out because every real post has its own prerendered file.

Each deploy tells the IndexNow engines (Bing, Yandex and others; Google doesn't take part) which public pages it changed, so they recrawl them within hours rather than weeks. The release workflow hashes the live HTML of every URL in `sitemap.xml` before `kamal deploy` and again after it, with asset file names and the footer's version number taken out, and `scripts/indexnow.mjs` sends the URLs whose hash changed, or that are new, to `api.indexnow.org`. The key is the one in `apps/web/public/b76076cc360a44c49982ee5d5143ea9b.txt`, which proves the site owns it. Both steps run with `continue-on-error`, so a failed ping never fails a deploy.

## Considered options

- A 301 from each release note to its `/changelog` anchor: offered as the alternative to `noindex`. With `noindex`, the full notes stay readable at their old URLs.
- `vite-plugin-ssg` or a server-side rendering server: more machinery than a dozen static routes need.
- Hand-written fallback HTML in `index.html`: fixes the homepage only.
- IndexNow with every sitemap URL on each deploy: simpler, but most deploys change no public page, and engines are asked to be sent only changed URLs.
