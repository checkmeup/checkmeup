# Search indexing and URL form of the public site

Date: 2026-10-13

Release notes stay at `/blog/<slug>` but carry `noindex, follow` in their prerendered HTML, so dozens of thin per-version pages don't drag down the whole site's ranking. `/changelog` is the one indexed page for them: every release, newest first, each with its own anchor and a link to the full notes. Articles stay indexed and are the only posts on `/blog` and in `sitemap.xml`. `robots.txt` doesn't disallow release notes, because a crawler blocked from a page never sees its `noindex`.

Every public URL has one form, without a trailing slash (except `/`). The server answers a trailing-slash URL with a 301 to the bare form, keeping the query string. The canonical and `og:url` are built from the path alone, without query, hash or trailing slash.

## Considered options

- A 301 from each release note to its `/changelog` anchor: offered as the alternative to `noindex`. With `noindex`, the full notes stay readable at their old URLs.
