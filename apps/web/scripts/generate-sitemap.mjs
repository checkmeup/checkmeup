// Regenerated on every `bun run build` (see package.json) so it never drifts
// from the router or the blog post list — do not hand-edit public/sitemap.xml.
import { readFileSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const postsMetaFile = join(__dirname, '../src/blog/postsMeta.ts')
const publicDir = join(__dirname, '../public')

const SITE_URL = 'https://checkmeup.net'

// Keep in sync with the indexable (non auth-gated) routes in src/router/index.ts.
const staticPaths = ['/', '/pricing', '/docs', '/faq', '/about', '/blog', '/changelog', '/terms', '/privacy', '/refund']

// postsMeta.ts (not posts/*.ts, which only holds each post's `content` now —
// see its own file comment) is the single source of truth for slug/date, the
// same fields the router and prerender.mts read. A plain regex scan over its
// one `postsMeta` array literal, rather than importing it, since this script
// runs under plain node (no TS runtime) — each entry's `slug`/`kind`/`date` is
// matched together so a missing field on either side fails loudly instead of
// silently miscounting or misaligning two separate global matches.
//
// Release-notes posts (kind: 'release') are left out on purpose: they are
// noindex (see BlogPostMeta.kind in src/blog/types.ts), and a sitemap should
// only list pages meant to be indexed. /changelog covers them instead. Do NOT
// also Disallow them in robots.txt — a crawler that can't fetch the page
// never sees its noindex, and the URL can stay indexed.
function blogPostUrls() {
  const contents = readFileSync(postsMetaFile, 'utf-8')
  const entryPattern =
    /\{[^{}]*?slug:\s*'([^']+)'[^{}]*?kind:\s*'(article|release)'[^{}]*?date:\s*'([^']+)'[^{}]*?\}/gs
  const entries = [...contents.matchAll(entryPattern)]
  if (entries.length === 0) {
    throw new Error(`${postsMetaFile}: no post entries found`)
  }
  // Every entry has a slug, so a count mismatch means some entry's kind/date
  // didn't match — fail rather than silently drop (or wrongly index) a post.
  const slugCount = [...contents.matchAll(/^\s*slug:\s*'/gm)].length
  if (entries.length !== slugCount) {
    throw new Error(`${postsMetaFile}: parsed ${entries.length} entries but found ${slugCount} slugs`)
  }
  return entries
    .filter(([, , kind]) => kind === 'article')
    .map(([, slug, , date]) => ({
      path: `/blog/${slug}`,
      lastmod: new Date(date).toISOString().slice(0, 10),
    }))
}

const urls = [...staticPaths.map((path) => ({ path })), ...blogPostUrls()]

const xml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urls
  .map(
    ({ path, lastmod }) =>
      `  <url>\n    <loc>${SITE_URL}${path}</loc>${lastmod ? `\n    <lastmod>${lastmod}</lastmod>` : ''}\n  </url>`,
  )
  .join('\n')}
</urlset>
`

writeFileSync(join(publicDir, 'sitemap.xml'), xml)
console.log(`Generated sitemap.xml with ${urls.length} URLs`)
