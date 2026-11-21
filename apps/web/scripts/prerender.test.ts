// @vitest-environment node
// Runs the real scripts/prerender.mts with node:fs stubbed
// (scripts/prerender-fs-stub.ts), twice, the second time from fresh modules,
// to check the pages it writes and that the same source writes them the same.
import { beforeAll, describe, expect, it, vi } from 'vitest'
import { Window } from 'happy-dom'
import { articlesMeta, postsMeta, releasesMeta } from '@/blog/posts'
import { written } from './prerender-fs-stub'

vi.mock('node:fs', async (importOriginal) => {
  const { stubNodeFs } = await import('./prerender-fs-stub')
  return stubNodeFs(await importOriginal<typeof import('node:fs')>())
})

// What the first run wrote, to compare with the second.
let firstRun = new Map<string, string | Buffer>()

beforeAll(async () => {
  vi.spyOn(console, 'log').mockImplementation(() => {})
  await import('./generate-sitemap.mjs')
  await import('./prerender.mts')
  firstRun = new Map(written)
  written.clear()
  vi.resetModules()
  await import('./generate-sitemap.mjs')
  await import('./prerender.mts')
  // Parsed up front, within this hook's timeout rather than each test's.
  for (const [file, html] of written) {
    if (file.endsWith('.html')) parsed.set(file, parseHtml(String(html)))
  }
}, 240_000)

const SITE_URL = 'https://checkmeup.net'

// Parses like a client that doesn't run JavaScript or fetch subresources.
const domWindow = new Window({
  settings: {
    disableJavaScriptEvaluation: true,
    disableJavaScriptFileLoading: true,
    disableCSSFileLoading: true,
    disableIframePageLoading: true,
  },
})

function parseHtml(html: string): Document {
  return new domWindow.DOMParser().parseFromString(html, 'text/html') as unknown as Document
}

// Each prerendered file, by its path on disk.
const parsed = new Map<string, Document>()

// The prerendered document served for a URL path.
function page(path: string): Document {
  const file = path === '/' ? '/dist/index.html' : `/dist${path}/index.html`
  const doc = [...parsed.entries()].find(([p]) => p.endsWith(file))?.[1]
  expect(doc, `no prerendered file for ${path}`).toBeDefined()
  return doc!
}

function meta(doc: Document, selector: string): string | null {
  return doc.querySelector(`meta[${selector}]`)?.getAttribute('content') ?? null
}

// A written file by the end of its path, such as /dist/img/og/pricing.png.
function file(suffix: string): string | Buffer | undefined {
  return [...written.entries()].find(([path]) => path.endsWith(suffix))?.[1]
}

// Every prerendered page's URL path, from the index.html files written.
function pagePaths(): string[] {
  return [...parsed.keys()]
    .filter((path) => path.endsWith('/index.html'))
    .map((path) => path.slice(path.lastIndexOf('/dist/') + 5, -'/index.html'.length) || '/')
}

const HOMEPAGE_CARD = `${SITE_URL}/img/checkmeup-og.png`

const LANDING_PATHS = [
  '/',
  '/pricing',
  '/docs',
  '/faq',
  '/about',
  '/terms',
  '/privacy',
  '/refund',
  '/blog',
  '/changelog',
]
const TOOL_PATHS = [
  '/tools/ssl-checker',
  '/tools/domain-expiry-checker',
  '/tools/dns-lookup',
  '/tools/cron-expression',
]

describe('prerender', () => {
  it('serves every landing page with an <h1> and its own title and Open Graph tags', () => {
    const titles = new Set<string>()
    for (const path of LANDING_PATHS) {
      const doc = page(path)
      expect(doc.querySelector('#app h1'), path).not.toBeNull()
      expect(doc.querySelectorAll('title'), path).toHaveLength(1)
      const title = doc.title
      titles.add(title)
      expect(meta(doc, 'property="og:title"'), path).toBe(title)
      expect(meta(doc, 'property="og:description"'), path).toBe(meta(doc, 'name="description"'))
      expect(meta(doc, 'property="og:url"'), path).toBe(`${SITE_URL}${path}`)
      expect(meta(doc, 'property="og:image"'), path).toBeTruthy()
    }
    expect(titles.size).toBe(LANDING_PATHS.length)
  })

  it('serves every blog post with its title, description and og:title, og:description, og:image', () => {
    for (const post of postsMeta) {
      const path = `/blog/${post.slug}`
      const doc = page(path)
      const description = post.description ?? post.excerpt
      expect(doc.title, path).toContain(post.title)
      expect(meta(doc, 'name="description"'), path).toBe(description)
      expect(meta(doc, 'property="og:title"'), path).toContain(post.title)
      expect(meta(doc, 'property="og:description"'), path).toBe(description)
      expect(meta(doc, 'property="og:image"'), path).toMatch(/^https:\/\//)
    }
  })

  it('serves every free checker with their <h1>, title, description and Open Graph tags', () => {
    const home = page('/')
    const titles = new Set<string>()
    for (const path of TOOL_PATHS) {
      const doc = page(path)
      expect(doc.querySelector('#app h1'), path).not.toBeNull()
      titles.add(doc.title)
      expect(doc.title, path).not.toBe(home.title)
      expect(meta(doc, 'name="description"'), path).not.toBe(meta(home, 'name="description"'))
      expect(meta(doc, 'property="og:title"'), path).toBe(doc.title)
      expect(meta(doc, 'property="og:description"'), path).toBe(meta(doc, 'name="description"'))
      expect(meta(doc, 'property="og:url"'), path).toBe(`${SITE_URL}${path}`)
    }
    expect(titles.size).toBe(TOOL_PATHS.length)
  })

  it('puts noindex, follow in every release note and index, follow in every article', () => {
    expect(releasesMeta.length).toBeGreaterThan(0)
    for (const release of releasesMeta) {
      expect(meta(page(`/blog/${release.slug}`), 'name="robots"'), release.slug).toBe(
        'noindex, follow',
      )
    }
    for (const article of articlesMeta) {
      expect(meta(page(`/blog/${article.slug}`), 'name="robots"'), article.slug).toBe(
        'index, follow',
      )
    }
  })

  it('has a prerendered page for every URL in sitemap.xml', () => {
    const sitemap = [...written.entries()].find(([path]) => path.endsWith('/public/sitemap.xml'))
    expect(sitemap).toBeDefined()
    const locs = [...String(sitemap![1]).matchAll(/<loc>([^<]+)<\/loc>/g)].map((m) => m[1])
    expect(locs.length).toBeGreaterThan(0)
    for (const loc of locs) {
      const doc = page(loc.slice(SITE_URL.length) || '/')
      expect(meta(doc, 'property="og:url"'), loc).toBe(loc)
    }
  })

  it('gives every page but / its own 1200×630 social image under 300 KB, described by its <h1>', () => {
    const paths = pagePaths().filter((path) => path !== '/')
    expect(paths.length).toBeGreaterThan(LANDING_PATHS.length + postsMeta.length)
    for (const path of paths) {
      const doc = page(path)
      const url = `${SITE_URL}/img/og${path}.png`
      expect(meta(doc, 'property="og:image"'), path).toBe(url)
      expect(meta(doc, 'name="twitter:image"'), path).toBe(url)
      expect(meta(doc, 'property="og:image:width"'), path).toBe('1200')
      expect(meta(doc, 'property="og:image:height"'), path).toBe('630')
      const h1 = doc.querySelector('#app h1')?.textContent?.replace(/\s+/g, ' ').trim()
      expect(h1, path).toBeTruthy()
      expect(meta(doc, 'property="og:image:alt"'), path).toBe(h1)
      expect(meta(doc, 'name="twitter:image:alt"'), path).toBe(h1)

      const png = file(`/dist/img/og${path}.png`)
      expect(png, path).toBeInstanceOf(Buffer)
      const bytes = png as Buffer
      expect(bytes.subarray(1, 4).toString(), path).toBe('PNG')
      expect([bytes.readUInt32BE(16), bytes.readUInt32BE(20)], path).toEqual([1200, 630])
      expect(bytes.length, path).toBeLessThan(300 * 1024)
    }
  })

  it('keeps the hand-made card on / and the 404 page', () => {
    const notFound = [...parsed.entries()].find(([path]) => path.endsWith('/dist/404.html'))?.[1]
    for (const doc of [page('/'), notFound!]) {
      expect(meta(doc, 'property="og:image"')).toBe(HOMEPAGE_CARD)
      expect(meta(doc, 'name="twitter:image"')).toBe(HOMEPAGE_CARD)
      expect(meta(doc, 'property="og:image:alt"')).toBeNull()
    }
  })

  it("names each post's own social image in its Article JSON-LD", () => {
    for (const post of postsMeta) {
      const doc = page(`/blog/${post.slug}`)
      const schemas = [...doc.querySelectorAll('script[type="application/ld+json"]')].map(
        (script) => JSON.parse(script.textContent ?? '') as { '@type': string; image?: string },
      )
      const article = schemas.find((schema) => schema['@type'] === 'Article')
      expect(article?.image, post.slug).toBe(meta(doc, 'property="og:image"'))
    }
  })

  it('writes the same pages and social images when built again from the same source', () => {
    expect([...written.keys()].sort()).toEqual([...firstRun.keys()].sort())
    for (const [path, data] of written) {
      const before = firstRun.get(path)!
      const same = Buffer.isBuffer(data) ? data.equals(before as Buffer) : data === before
      expect(same, path).toBe(true)
    }
  })

  it('links Terms and Privacy from the footer of the landing pages and the blog', () => {
    const paths = [...LANDING_PATHS, ...postsMeta.map((post) => `/blog/${post.slug}`)]
    for (const path of paths) {
      const hrefs = [...page(path).querySelectorAll('footer a')].map((a) => a.getAttribute('href'))
      expect(hrefs, path).toContain('/terms')
      expect(hrefs, path).toContain('/privacy')
    }
  })
})
