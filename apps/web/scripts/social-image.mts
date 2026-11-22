// Renders a public page's Open Graph image at build time (PRD social-images):
// its <h1>, a label of its kind, the logo and checkmeup.net, on the brand's
// green, so a shared link previews as the page rather than as the homepage
// card. scripts/prerender.mts calls it for every page but `/`.
import { readFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { dirname, join } from 'node:path'
import satori, { type SatoriOptions } from 'satori'
import { postsMeta } from '../src/blog/posts'

export const SOCIAL_IMAGE_WIDTH = 1200
export const SOCIAL_IMAGE_HEIGHT = 630

const PADDING_X = 72
const TITLE_WIDTH = SOCIAL_IMAGE_WIDTH - 2 * PADDING_X
const MAX_TITLE_LINES = 3
// Largest first; a title that doesn't fit 3 lines at one size tries the next.
const TITLE_SIZES = [72, 66, 60, 54, 48]
const LOGO_HEIGHT = 44
const LOGO_TILE_INSET = 9
// public/img/logo-dark.svg is 276 units high: its mark the first 276 wide,
// its wordmark from 300 to the end at 1686.
const LOGO_UNITS = 276
const WORDMARK_FROM = 300
const WORDMARK_UNITS = 1686 - WORDMARK_FROM

// The label above the title: what kind of page a shared link opens.
export function socialImageLabel(path: string): string | null {
  const slug = /^\/blog\/(.+)$/.exec(path)?.[1]
  if (slug) {
    const post = postsMeta.find((meta) => meta.slug === slug)
    return post?.kind === 'release' ? 'Release notes' : 'Article'
  }
  if (path.startsWith('/tools/')) return 'Free tool'
  if (path.startsWith('/alternatives/')) return 'Comparison'
  if (/^\/[a-z-]+-monitoring$/.test(path)) return 'Monitoring'
  return null
}

// The picture's colors, read from the tokens in src/style.css so it follows
// them (docs/reference/design.md): the dark theme's text on the brand's deep
// green, a theme of neither, so a preview stands out in a light feed and a
// dark one alike.
function palette() {
  // A path fixed in this script, not input.
  // eslint-disable-next-line security/detect-non-literal-fs-filename
  const css = readFileSync(new URL('../src/style.css', import.meta.url), 'utf-8')
  const tokens = (block: string | undefined) =>
    new Map([...(block ?? '').matchAll(/--([\w-]+):\s*([^;]+);/g)].map((m) => [m[1], m[2].trim()]))
  const scale = tokens(/@theme\s*\{([\s\S]*?)\n\}/.exec(css)?.[1])
  const dark = tokens(/:root\s*\{([\s\S]*?)\n\}/.exec(css)?.[1])
  const color = (name: string): string => {
    const value = scale.get(name) ?? dark.get(name)
    if (!value) throw new Error(`no token --${name} in src/style.css`)
    const ref = /^var\(--([\w-]+)\)$/.exec(value)?.[1]
    return ref ? color(ref) : value
  }
  return {
    background: `linear-gradient(135deg, ${color('color-green-700')} -30%, ${color('color-green-900')} 30%, ${color('surface-raised')} 100%)`,
    text: color('text-strong'),
    dim: color('text-dim'),
    accent: color('color-green-300'),
    accentWash: color('accent-wash'),
    statusUp: color('status-up'),
    tickUp: color('color-green-700'),
    // The logo's mark is drawn in green-700 and green-500, which vanish at the
    // faint watermark's opacity: each a step lighter there, keeping the two
    // halves as close. The logo itself keeps its own colors.
    markGreens: [color('color-green-500'), color('color-green-300')],
  }
}

function fontFile(dir: string, file: string): Buffer {
  const geist = dirname(createRequire(import.meta.url).resolve('geist/font/sans'))
  // A font file of the geist package this script names, not input.
  // eslint-disable-next-line security/detect-non-literal-fs-filename
  return readFileSync(join(geist, 'fonts', dir, file))
}

// Read once, on the first image.
const fonts: SatoriOptions['fonts'] = []
function satoriOptions(): SatoriOptions {
  if (fonts.length === 0) {
    fonts.push(
      {
        name: 'Geist',
        data: fontFile('geist-sans', 'Geist-SemiBold.ttf'),
        weight: 600,
      },
      {
        name: 'Geist Mono',
        data: fontFile('geist-mono', 'GeistMono-Regular.ttf'),
        weight: 400,
      },
    )
  }
  return { width: SOCIAL_IMAGE_WIDTH, height: SOCIAL_IMAGE_HEIGHT, fonts }
}

interface Node {
  type: string
  props: {
    style?: Record<string, unknown>
    children?: string | Node | (Node | null)[]
    src?: string
  }
}

function titleNode(title: string, fontSize: number, color: string): Node {
  return {
    type: 'div',
    props: {
      style: {
        display: 'flex',
        width: TITLE_WIDTH,
        fontFamily: 'Geist',
        fontWeight: 600,
        fontSize,
        lineHeight: 1.12,
        letterSpacing: -0.02 * fontSize,
        color,
      },
      children: title,
    },
  }
}

// The largest size at which the title wraps to at most 3 lines with no word
// past the right edge, measured from where satori lays out each word; null
// when it doesn't fit even at the smallest size, rather than cut it.
export async function fitTitle(title: string): Promise<{ fontSize: number; lines: number } | null> {
  for (const fontSize of TITLE_SIZES) {
    const svg = await satori(titleNode(title, fontSize, '#000') as never, {
      ...satoriOptions(),
      height: SOCIAL_IMAGE_HEIGHT * 2,
      embedFont: false,
    })
    const words = [
      ...svg.matchAll(/<text x="([\d.]+)" y="([\d.]+)" width="([\d.]+)"[^>]*>([^<]*)<\/text>/g),
    ]
      .filter((m) => m[4].trim() !== '')
      .map((m) => ({ y: Number(m[2]), right: Number(m[1]) + Number(m[3]) }))
    const lines = new Set(words.map((w) => w.y)).size
    if (lines <= MAX_TITLE_LINES && words.every((w) => w.right <= TITLE_WIDTH + 0.5)) {
      return { fontSize, lines }
    }
  }
  return null
}

// The 90-day check strip of the homepage card (CheckStrip.vue), every check
// passing: the product's look, the same on every page.
const STRIP_TICKS = 60
const STRIP_GAP = 6

function checkStrip(color: string): Node {
  const tick = (TITLE_WIDTH - (STRIP_TICKS - 1) * STRIP_GAP) / STRIP_TICKS
  return {
    type: 'div',
    props: {
      style: { display: 'flex', gap: STRIP_GAP },
      children: Array.from({ length: STRIP_TICKS }, () => ({
        type: 'div',
        props: {
          style: {
            width: tick,
            height: 26,
            borderRadius: 3,
            backgroundColor: color,
          },
        },
      })),
    },
  }
}

export function socialImageNode(title: string, label: string | null, fontSize: number): Node {
  const colors = palette()
  // A path fixed in this script, not input.
  // eslint-disable-next-line security/detect-non-literal-fs-filename
  const logo = readFileSync(new URL('../public/img/logo-dark.svg', import.meta.url), 'utf-8')
  // A part of the logo, cropped by the viewBox.
  const crop = (from: number, units: number) =>
    logo.replace(
      /width="[^"]*" height="[^"]*" viewBox="[^"]*"/,
      `width="${units}" height="${LOGO_UNITS}" viewBox="${from} 0 ${units} ${LOGO_UNITS}"`,
    )
  const mark = crop(0, LOGO_UNITS)
  const watermark = mark
    .replaceAll('#0F6E56', colors.markGreens[0])
    .replaceAll('#1D9E75', colors.markGreens[1])
  const svgSrc = (svg: string) => `data:image/svg+xml;base64,${Buffer.from(svg).toString('base64')}`
  const mono = { fontFamily: 'Geist Mono', fontWeight: 400 }
  return {
    type: 'div',
    props: {
      style: {
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between',
        // Satori sizes boxes as content-box unless told otherwise, which
        // would put the padding outside the canvas and the footer off it.
        boxSizing: 'border-box',
        width: SOCIAL_IMAGE_WIDTH,
        height: SOCIAL_IMAGE_HEIGHT,
        padding: `60px ${PADDING_X}px 56px`,
        backgroundImage: colors.background,
        borderTop: `8px solid ${colors.accent}`,
      },
      children: [
        // Drawn first, so the title is drawn over it.
        {
          type: 'img',
          props: {
            src: svgSrc(watermark),
            style: {
              position: 'absolute',
              right: -90,
              top: 60,
              width: 520,
              height: 520,
              opacity: 0.12,
            },
          },
        },
        {
          type: 'div',
          props: {
            style: { display: 'flex', alignItems: 'center', gap: 14 },
            children: [
              // The mark on a light tile: its own green-700 half would vanish
              // on the green background.
              {
                type: 'div',
                props: {
                  style: {
                    display: 'flex',
                    padding: LOGO_TILE_INSET,
                    borderRadius: 14,
                    backgroundColor: colors.text,
                  },
                  children: {
                    type: 'img',
                    props: {
                      src: svgSrc(mark),
                      style: { width: LOGO_HEIGHT, height: LOGO_HEIGHT },
                    },
                  },
                },
              },
              {
                type: 'img',
                props: {
                  src: svgSrc(crop(WORDMARK_FROM, WORDMARK_UNITS)),
                  style: {
                    width: (LOGO_HEIGHT * WORDMARK_UNITS) / LOGO_UNITS,
                    height: LOGO_HEIGHT,
                  },
                },
              },
            ],
          },
        },
        {
          type: 'div',
          props: {
            style: {
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'flex-start',
              gap: 22,
            },
            children: [
              label
                ? {
                    type: 'div',
                    props: {
                      style: {
                        ...mono,
                        display: 'flex',
                        alignItems: 'center',
                        gap: 12,
                        padding: '8px 18px',
                        borderRadius: 999,
                        backgroundColor: colors.accentWash,
                        fontSize: 24,
                        color: colors.accent,
                      },
                      children: [
                        {
                          type: 'div',
                          props: {
                            style: {
                              width: 12,
                              height: 12,
                              borderRadius: 6,
                              backgroundColor: colors.statusUp,
                            },
                          },
                        },
                        label,
                      ],
                    },
                  }
                : null,
              titleNode(title, fontSize, colors.text),
            ],
          },
        },
        {
          type: 'div',
          props: {
            style: { display: 'flex', flexDirection: 'column', gap: 20 },
            children: [
              checkStrip(colors.tickUp),
              {
                type: 'div',
                props: {
                  style: {
                    ...mono,
                    display: 'flex',
                    justifyContent: 'space-between',
                    fontSize: 24,
                    color: colors.dim,
                  },
                  children: [
                    {
                      type: 'div',
                      props: {
                        style: { color: colors.text },
                        children: 'checkmeup.net',
                      },
                    },
                    {
                      type: 'div',
                      props: {
                        children: 'cron · uptime · ssl · domain · port · dns',
                      },
                    },
                  ],
                },
              },
            ],
          },
        },
      ],
    },
  }
}

// The page's image as a PNG, or null when its title doesn't fit.
export async function renderSocialImage(
  title: string,
  label: string | null,
): Promise<Buffer | null> {
  const fit = await fitTitle(title)
  if (!fit) return null
  const svg = await satori(socialImageNode(title, label, fit.fontSize) as never, satoriOptions())
  // Imported here, so a native binary missing on the build's platform fails
  // this page's image, which prerender.mts falls back from, not the build.
  const { Resvg } = await import('@resvg/resvg-js')
  return new Resvg(svg).render().asPng()
}
