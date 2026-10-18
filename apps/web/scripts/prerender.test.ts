// @vitest-environment node
// Runs the real scripts/prerender.mts with node:fs stubbed: it reads the
// source index.html in place of the built dist/index.html (only the hashed
// asset tags differ) and its writes land in `written` instead of dist/, so
// each page is checked as a crawler that doesn't run JavaScript fetches it,
// without a full `vite build`.
import { beforeAll, describe, expect, it, vi } from 'vitest'
import { Window } from 'happy-dom'
import { articlesMeta, postsMeta, releasesMeta } from '@/blog/posts'

const written = new Map<string, string>()

vi.mock('node:fs', async (importOriginal) => {
  const fs = await importOriginal<typeof import('node:fs')>()
  const readFileSync = ((path: string, ...rest: unknown[]) =>
    String(path).endsWith('dist/index.html')
      ? fs.readFileSync(new URL('../index.html', import.meta.url), 'utf-8')
      : (fs.readFileSync as (...args: unknown[]) => unknown)(
          path,
          ...rest,
        )) as typeof fs.readFileSync
  const writeFileSync = ((path: string, data: string) => {
    written.set(String(path), String(data))
  }) as typeof fs.writeFileSync
  const mkdirSync = (() => undefined) as typeof fs.mkdirSync
  const stubs = { readFileSync, writeFileSync, mkdirSync }
  return { ...fs, ...stubs, default: { ...fs, ...stubs } }
})

beforeAll(async () => {
  vi.spyOn(console, 'log').mockImplementation(() => {})
  await import('./generate-sitemap.mjs')
  await import('./prerender.mts')
  // Parsed up front, within this hook's timeout rather than each test's.
  for (const [file, html] of written) {
    if (file.endsWith('.html')) parsed.set(file, parseHtml(html))
  }
}, 120_000)

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
const TOOL_PATHS = ['/tools/ssl-checker', '/tools/domain-expiry-checker']

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

  it('serves both free checkers with their <h1>, title, description and Open Graph tags', () => {
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
    const locs = [...sitemap![1].matchAll(/<loc>([^<]+)<\/loc>/g)].map((m) => m[1])
    expect(locs.length).toBeGreaterThan(0)
    for (const loc of locs) {
      const doc = page(loc.slice(SITE_URL.length) || '/')
      expect(meta(doc, 'property="og:url"'), loc).toBe(loc)
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
