// @vitest-environment node
// Runs scripts/prerender.mts (node:fs stubbed, scripts/prerender-fs-stub.ts)
// with every social image failing to render, as a missing font would.
import { expect, it, vi } from 'vitest'
import { written } from './prerender-fs-stub'

vi.mock('node:fs', async (importOriginal) => {
  const { stubNodeFs } = await import('./prerender-fs-stub')
  return stubNodeFs(await importOriginal<typeof import('node:fs')>())
})

vi.mock('./social-image.mts', async (importOriginal) => ({
  ...(await importOriginal<typeof import('./social-image.mts')>()),
  renderSocialImage: () => Promise.reject(new Error('no font')),
}))

it('keeps the homepage card on a page whose social image fails to render, and writes every page', async () => {
  vi.spyOn(console, 'log').mockImplementation(() => {})
  const warn = vi.spyOn(console, 'warn').mockImplementation(() => {})
  await import('./prerender.mts')

  const pages = [...written.entries()].filter(([path]) => path.endsWith('.html'))
  expect(pages.map(([path]) => path)).toContainEqual(
    expect.stringMatching(/\/dist\/pricing\/index\.html$/),
  )
  expect([...written.keys()].filter((path) => path.endsWith('.png'))).toEqual([])
  for (const [path, html] of pages) {
    expect(String(html), path).toContain(
      '<meta property="og:image" content="https://checkmeup.net/img/checkmeup-og.png">',
    )
    expect(String(html), path).not.toContain('og:image:alt')
  }
  expect(warn).toHaveBeenCalledWith(expect.stringContaining('No social image for /pricing'))
}, 120_000)
