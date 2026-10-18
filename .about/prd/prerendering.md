# Prerendered public pages

## Problem

The marketing site and blog render only in the browser, so crawlers that don't run JavaScript (Bingbot) and link-preview bots see an empty page: no `<h1>`, no title, no preview card.

## Users

Search engines and social or chat link previews, on behalf of visitors finding Checkmeup.

## Success signal

Every public page's raw HTML has its own title, description, Open Graph tags and `<h1>`; Bing's "H1 missing" notice is gone.

## Scope

- Every public route prerendered to static HTML at build time (ADR seo), in both build paths (the web app's build script and the Dockerfile).
- The server serving those files directly, without a redirect.

## Non-goals

- Server-side rendering at request time; signed-in pages.

## Acceptance criteria

- A prerendered route is served from its own `index.html` without a redirect.
  → test: apps/api/internal/server/server_test.go "TestHandleSPA/prerendered route directory serves its index.html without a redirect"
- Fetching any blog post's URL without running JavaScript returns its title, description and `og:title`, `og:description`, `og:image`.
  → test: apps/web/scripts/prerender.test.ts "serves every blog post with its title, description and og:title, og:description, og:image"
- Fetching `/`, `/pricing`, `/docs`, `/faq`, `/about`, `/terms`, `/privacy`, `/refund`, `/blog` and `/changelog` without JavaScript returns a page with an `<h1>` and its own title and Open Graph tags.
  → test: apps/web/scripts/prerender.test.ts "serves every landing page with an <h1> and its own title and Open Graph tags"
- `sitemap.xml` and the `noindex` tags keep working on the prerendered pages.
  → test: apps/web/scripts/prerender.test.ts "has a prerendered page for every URL in sitemap.xml"
  → test: apps/web/scripts/prerender.test.ts "puts noindex, follow in every release note and index, follow in every article"
- Every Article's meta description is its own `description` of 155 characters or fewer, so search results don't cut it off, not the longer excerpt shown on the page.
  → test: apps/web/src/views/BlogPostView.test.ts "gives every article a search description of 155 characters or fewer"
  → test: apps/web/src/views/BlogPostView.test.ts "uses the article's description, not its excerpt, as the meta description"
- A `HEAD` request to a public page answers like `GET`, not with a `405`.
  → test: apps/api/internal/server/server_test.go "TestRoutes_HEADOnAPageAnswersLikeGET"
