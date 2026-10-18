// @vitest-environment node
// Runs the real scripts/generate-sitemap.mjs with writeFileSync stubbed, so
// the sitemap it would write to public/ is checked without touching disk.
import { beforeAll, describe, expect, it, vi } from 'vitest'
import { articlesMeta, releasesMeta } from '@/blog/posts'

const written = new Map<string, string>()

vi.mock('node:fs', async (importOriginal) => {
  const fs = await importOriginal<typeof import('node:fs')>()
  const writeFileSync = ((path: string, data: string) => {
    written.set(String(path), String(data))
  }) as typeof fs.writeFileSync
  return { ...fs, writeFileSync, default: { ...fs, writeFileSync } }
})

let locs: string[] = []

beforeAll(async () => {
  vi.spyOn(console, 'log').mockImplementation(() => {})
  await import('./generate-sitemap.mjs')
  const sitemap = [...written.entries()].find(([path]) => path.endsWith('/public/sitemap.xml'))
  expect(sitemap).toBeDefined()
  locs = [...sitemap![1].matchAll(/<loc>([^<]+)<\/loc>/g)].map((m) => m[1])
})

describe('generate-sitemap', () => {
  it('lists /changelog and every article, and no release note', () => {
    expect(locs).toContain('https://checkmeup.net/changelog')
    expect(articlesMeta.length).toBeGreaterThan(0)
    for (const article of articlesMeta) {
      expect(locs).toContain(`https://checkmeup.net/blog/${article.slug}`)
    }
    expect(releasesMeta.length).toBeGreaterThan(0)
    for (const release of releasesMeta) {
      expect(locs).not.toContain(`https://checkmeup.net/blog/${release.slug}`)
    }
  })

  it('lists both free checkers', () => {
    expect(locs).toContain('https://checkmeup.net/tools/ssl-checker')
    expect(locs).toContain('https://checkmeup.net/tools/domain-expiry-checker')
  })
})
