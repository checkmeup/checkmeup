// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { articlesMeta, releasesMeta } from '@/blog/posts'
import { fitTitle, renderSocialImage, socialImageLabel, socialImageNode } from './social-image.mts'

// 100 characters of ordinary words, longer than any page's <h1> today.
const HUNDRED_CHARACTERS =
  'Why a cron job that stopped running never tells anyone, and how to find out before your clients see.'

// Every string in a node tree, and every image source.
function texts(node: unknown): string[] {
  if (typeof node === 'string') return [node]
  if (!node || typeof node !== 'object') return []
  if (Array.isArray(node)) return node.flatMap(texts)
  const { props } = node as { props: { children?: unknown; src?: string } }
  return [...(props.src ? [props.src] : []), ...texts(props.children)]
}

describe('socialImageLabel', () => {
  it.each([
    [`/blog/${articlesMeta[0].slug}`, 'Article'],
    [`/blog/${releasesMeta[0].slug}`, 'Release notes'],
    ['/tools/cron-expression', 'Free tool'],
    ['/alternatives/cronitor', 'Comparison'],
    ['/cron-monitoring', 'Monitoring'],
    ['/domain-expiry-monitoring', 'Monitoring'],
    ['/pricing', null],
    ['/blog', null],
    ['/changelog', null],
    ['/terms', null],
  ])('labels %s %s', (path, label) => {
    expect(socialImageLabel(path)).toBe(label)
  })
})

describe('fitTitle', () => {
  it('fits a 100-character title in at most 3 lines', async () => {
    expect(HUNDRED_CHARACTERS).toHaveLength(100)
    const { lines } = await fitTitle(HUNDRED_CHARACTERS)
    expect(lines).toBeLessThanOrEqual(3)
  })

  it('keeps a short title on one line at full size', async () => {
    expect(await fitTitle('Pricing')).toEqual({ fontSize: 72, lines: 1 })
  })

  it("refuses a title that doesn't fit 3 lines at the smallest size, rather than cut it", async () => {
    await expect(fitTitle(HUNDRED_CHARACTERS.repeat(3))).rejects.toThrow(/doesn't fit 3 lines/)
  })
})

describe('socialImageNode', () => {
  it("is drawn from the page's <h1>, its label, the logo and checkmeup.net", () => {
    const all = texts(socialImageNode('Cron expression explainer', 'Free tool', 72))
    expect(all).toContain('Cron expression explainer')
    expect(all).toContain('Free tool')
    expect(all).toContain('checkmeup.net')
    expect(all.some((text) => text.startsWith('data:image/svg+xml;base64,'))).toBe(true)
  })

  it('leaves the label out on a page without one', () => {
    expect(texts(socialImageNode('Pricing', null, 72))).toHaveLength(3)
  })
})

describe('renderSocialImage', () => {
  it('renders a 1200×630 PNG', async () => {
    const png = await renderSocialImage('Cron expression explainer', 'Free tool')
    expect(png.subarray(1, 4).toString()).toBe('PNG')
    expect([png.readUInt32BE(16), png.readUInt32BE(20)]).toEqual([1200, 630])
  })
})
