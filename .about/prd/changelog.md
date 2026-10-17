# Changelog and release notes out of the search index

## Problem

Every release gets its own blog post, and 35 thin per-version pages sat in the search index next to 10 real articles, dragging down the whole site's rating. Each post was also reachable with and without a trailing slash, and the canonical followed whichever form was visited, so search engines saw duplicates.

## Users

Search engines, on behalf of visitors searching for what Checkmeup monitors; and people who want to read what shipped.

## Success signal

Search Console shows the release notes excluded by `noindex` and no trailing-slash duplicates indexed (Inbox: Check Search Console after the changelog release).

## Scope

- Release notes keep their `/blog/<slug>` URLs, carry `noindex, follow`, and are listed on an indexed `/changelog` (ADR seo).
- `/blog` lists only Articles; `sitemap.xml` lists Articles and `/changelog`, no Release notes.
- One URL form per page: a trailing-slash URL 301s to the bare form; the canonical and `og:url` come from the path alone.

## Non-goals

- A 301 from each release note to its `/changelog` anchor; the full notes stay readable.
- Blocking release notes in `robots.txt`: a blocked crawler never sees the `noindex`.

## Acceptance criteria

- A release note's page sets `noindex, follow`; an Article's sets `index, follow`.
  → test: apps/web/src/views/BlogPostView.test.ts "marks a release-notes post noindex but still followed"
- The `noindex, follow` tag is in each release note's prerendered HTML, as fetched without running JavaScript.
- `/blog` lists no release notes and links to `/changelog`.
  → test: apps/web/src/views/BlogView.test.ts "does not list release-notes posts"
  → test: apps/web/src/views/BlogView.test.ts "links to the changelog"
- `/changelog` lists every release newest first, each under an anchor named by its slug and linked to its full notes, and lists no Articles.
  → test: apps/web/src/views/ChangelogView.test.ts "renders one entry per release, newest first, each anchored by its slug"
  → test: apps/web/src/views/ChangelogView.test.ts "links every release to its full notes"
  → test: apps/web/src/views/ChangelogView.test.ts "does not list articles"
- A release note links back to `/changelog`, not `/blog`.
  → test: apps/web/src/views/BlogPostView.test.ts "links a release-notes post back to the changelog"
- `sitemap.xml` has `/changelog` and every Article, and no release note.
- `robots.txt` doesn't disallow release notes.
- A trailing-slash URL other than `/` answers 301 to the bare form, keeping the query, and never to another host.
  → test: apps/api/internal/server/server_test.go "TestHandleSPA/trailing slash 301s to the slash-less form, keeping the query"
  → test: apps/api/internal/server/server_test.go "TestHandleSPA/root path is served, not redirected"
- A post visited with a trailing slash or a query still emits the slash-less, query-less canonical.
  → test: apps/web/src/composables/useSeo.test.ts "drops a trailing slash so both URL forms share one canonical"
  → test: apps/web/src/views/BlogPostView.test.ts "emits a slash-less canonical URL even when visited with a trailing slash"
