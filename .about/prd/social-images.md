# A social image per page

## Problem

A link to any Checkmeup page, shared in Slack, X, LinkedIn or a chat, previews with the same picture: the homepage's "Know before your client does." card. A reader can't tell a cron guide from the SSL checker from the picture, and the link reads as an ad for the product rather than as the article they were sent. The founder is sharing links now: cross-posts on dev.to and Hashnode, directory listings, and Show HN, Reddit and Indie Hackers next (Inbox "Launch and directory listings").

## Users

People who see a shared link to a Checkmeup page and decide whether to open it, and the founder, who shares them. Not search engines' result pages, which show no Open Graph image.

## Success signal

Qualitative: pasting the URL of any Article or free tool into a link-preview check (LinkedIn Post Inspector, a Slack message) shows a picture with that page's title, where it showed the homepage card before. Clicks from social referrers are too few today (1–23 visitors a day, ADR frontend) to show a change.

## Scope

- Every prerendered public page other than `/` and `404.html` gets its own Open Graph and Twitter image, Release notes and the legal pages included, rendered at build time with the prerendered pages (ADR seo) from the page as it is in that build.
- The picture shows the page's `<h1>` (an Article's or Release note's title), not its `<title>`: shorter, and it doesn't repeat the title the preview prints under it.
- Above the title, a label of the page's kind: "Article" on an Article, "Release notes" on a Release note, "Free tool" under `/tools/`, "Comparison" under `/alternatives/`, "Monitoring" on a monitor type page such as `/cron-monitoring`, and none on the rest (`/pricing`, `/docs`, `/faq`, `/about`, `/blog`, `/changelog`, the legal pages). The label is only in the picture; the page's own eyebrow stays as it is.
- One design for every page: the Checkmeup logo on a light tile, its type, green accent and check strip, on the brand's deep green rather than either theme's background. Previews don't follow the reader's theme, so the picture is in neither: it stands out in a light feed and a dark one alike, where a light card fades into a light feed and a dark one into a dark feed.
- A title too long for one line wraps, to at most 3 lines; past what fits in 3 lines at full size, the type gets smaller until it fits.
- The homepage card (`/img/checkmeup-og.png`) is in the same design, with the landing hero's headline, cron line, missed-ping alert and a 90-day strip with one miss. A script draws it (`bun scripts/homepage-card.mts`, run by hand when the hero changes), and the PNG is committed rather than built: it is the card a page falls back to when its own image fails to render, so it can't depend on the renderer. The 300 KB limit is the per-page images'; the card is drawn at 2400×1260 and is larger.
- Each page's image has a URL of its own that stays the same from build to build (`/img/og/<page path>.png`), so a deploy that doesn't change a page doesn't change its HTML either, and IndexNow isn't sent every page (ADR seo).
- The page also states the image's width, height and alt text (its `<h1>`).
- Every post's `Article` JSON-LD names its image; the publisher's logo stays `/img/checkmeup-og.png`.
- A page whose image fails to render keeps the homepage card and the build goes on, logging which page; the test of the build output then fails in CI, so the failure doesn't go unnoticed.

## Non-goals

- Images for signed-in pages, status pages or status badges.
- Setting the image as a cover on dev.to or Hashnode: the founder uploads it there by hand.
- Refreshing a preview a site already cached before the image changed (LinkedIn, Slack and X keep their own copy).
- An image editor, or a different design per page.

## Anti-goals

- A title cut off, or running outside the picture, in the preview: worse than the shared card.
- A picture whose title differs from the page's, because it was made once by hand and the title changed since.
- A build that fails, or a deploy held up, because an image couldn't be rendered.

## Acceptance criteria

- Fetching any prerendered public page other than `/` and `404.html` without JavaScript returns `og:image` and `twitter:image` of `https://checkmeup.net/img/og/<page path>.png`, and that file is in the build output.
  → test: apps/web/scripts/prerender.test.ts "gives every page but / its own 1200×630 social image under 300 KB, described by its <h1>"
- Fetching `/` without JavaScript returns `og:image` and `twitter:image` of `https://checkmeup.net/img/checkmeup-og.png`.
  → test: apps/web/scripts/prerender.test.ts "keeps the hand-made card on / and the 404 page"
- Each page's image is a 1200×630 PNG under 300 KB, the size LinkedIn, X and Slack show uncropped in a large card, and the page says so in `og:image:width` and `og:image:height`.
  → test: apps/web/scripts/prerender.test.ts "gives every page but / its own 1200×630 social image under 300 KB, described by its <h1>"
- Each page's `og:image:alt` and `twitter:image:alt` are its `<h1>` text.
  → test: apps/web/scripts/prerender.test.ts "gives every page but / its own 1200×630 social image under 300 KB, described by its <h1>"
- Each page's image is rendered from its `<h1>` text, the label of its kind (or none, by the list in Scope), the Checkmeup logo and `checkmeup.net`.
  → test: apps/web/scripts/social-image.test.ts "is drawn from the page's <h1>, its label, the logo and checkmeup.net"
  → test: apps/web/scripts/social-image.test.ts "labels %s %s"
  → test: apps/web/scripts/prerender.test.ts "gives every page but / its own 1200×630 social image under 300 KB, described by its <h1>"
- Every page's `<h1>` in the build, and a 100-character title of ordinary English words, lay out in at most 3 lines inside the picture's margins, with nothing cut off.
  → test: apps/web/scripts/social-image.test.ts "fits a 100-character title in at most 3 lines"
  → test: apps/web/scripts/social-image.test.ts "refuses a title that doesn't fit 3 lines at the smallest size, rather than cut it"
  → test: apps/web/scripts/prerender.test.ts "gives every page but / its own 1200×630 social image under 300 KB, described by its <h1>"
- Building twice from the same source gives every page the same image URL and the same HTML.
  → test: apps/web/scripts/prerender.test.ts "writes the same pages and social images when built again from the same source"
- Every post's `Article` JSON-LD has `image` set to the same URL as its `og:image`.
  → test: apps/web/scripts/prerender.test.ts "names each post's own social image in its Article JSON-LD"
- A page whose image fails to render is still written, with the homepage card as its `og:image`, and the build finishes.
  → test: apps/web/scripts/prerender-fallback.test.ts "keeps the homepage card on a page whose social image fails to render, and writes every page"
