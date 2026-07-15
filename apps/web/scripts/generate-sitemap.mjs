// Regenerated on every `bun run build` (see package.json) so it never drifts
// from the router or the blog post list — do not hand-edit public/sitemap.xml.
import { readFileSync, readdirSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const postsDir = join(__dirname, '../src/blog/posts')
const publicDir = join(__dirname, '../public')

const SITE_URL = 'https://checkmeup.net'

// Keep in sync with the indexable (non auth-gated) routes in src/router/index.ts.
const staticPaths = ['/', '/pricing', '/docs', '/faq', '/about', '/blog', '/terms', '/privacy', '/refund']

function blogPostUrls() {
  return readdirSync(postsDir)
    .filter((file) => file.endsWith('.ts'))
    .map((file) => {
      const contents = readFileSync(join(postsDir, file), 'utf-8')
      // The post's `slug:` field is the source of truth for its URL (it's what
      // the router and prerender.mts both use) — the filename can drift from it,
      // as happened for checkmeup-vs-competitors.ts (slug:
      // checkmeup-vs-healthchecks-uptimerobot-cronitor), which left the sitemap
      // advertising a dead URL while the real post was missing from it entirely.
      const slugMatch = contents.match(/slug:\s*'([^']+)'/)
      if (!slugMatch) {
        throw new Error(`${file}: no slug field found`)
      }
      const dateMatch = contents.match(/date:\s*'([^']+)'/)
      const lastmod = dateMatch ? new Date(dateMatch[1]).toISOString().slice(0, 10) : undefined
      return { path: `/blog/${slugMatch[1]}`, lastmod }
    })
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
