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
- Fetching `/`, `/pricing`, `/docs`, `/faq`, `/about`, `/terms`, `/privacy`, `/refund` and `/blog` without JavaScript returns a page with an `<h1>` and its own title and Open Graph tags.
- `sitemap.xml` and the `noindex` tags keep working on the prerendered pages.
