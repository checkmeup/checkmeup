// Renders a public page's Open Graph image at build time (PRD social-images):
// its <h1>, a label of its kind, the logo and checkmeup.net, in the light
// theme's colors, so a shared link previews as the page rather than as the
// homepage card. scripts/prerender.mts calls it for every page but `/`.
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
const LOGO_ASPECT = 1686 / 276 // public/img/logo-light.svg's viewBox

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

// The light theme's colors, read from the tokens in src/style.css so the
// picture follows them (docs/reference/design.md).
function lightThemeColors() {
  // A path fixed in this script, not input.
  // eslint-disable-next-line security/detect-non-literal-fs-filename
  const css = readFileSync(new URL('../src/style.css', import.meta.url), 'utf-8')
  const tokens = (block: string | undefined) =>
    new Map(
      [...(block ?? '').matchAll(/--([\w-]+):\s*([^;]+);/g)].map((m) => [m[1], m[2].trim()]),
    )
  const theme = tokens(/@theme\s*\{([\s\S]*?)\n\}/.exec(css)?.[1])
  const light = tokens(/:root\[data-theme='light'\]\s*\{([\s\S]*?)\n\}/.exec(css)?.[1])
  const color = (name: string): string => {
    const value = light.get(name)
    if (!value) throw new Error(`no light-theme token --${name} in src/style.css`)
    const ref = /^var\(--([\w-]+)\)$/.exec(value)?.[1]
    return ref ? (theme.get(ref) ?? light.get(ref) ?? value) : value
  }
  return {
    bg: color('bg'),
    border: color('border'),
    text: color('text-strong'),
    dim: color('text-dim'),
    accent: color('accent'),
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
      { name: 'Geist', data: fontFile('geist-sans', 'Geist-SemiBold.ttf'), weight: 600 },
      { name: 'Geist Mono', data: fontFile('geist-mono', 'GeistMono-Regular.ttf'), weight: 400 },
    )
  }
  return { width: SOCIAL_IMAGE_WIDTH, height: SOCIAL_IMAGE_HEIGHT, fonts }
}

interface Node {
  type: string
  props: { style?: Record<string, unknown>; children?: string | Node | (Node | null)[]; src?: string }
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
export async function fitTitle(
  title: string,
): Promise<{ fontSize: number; lines: number } | null> {
  for (const fontSize of TITLE_SIZES) {
    const svg = await satori(titleNode(title, fontSize, '#000') as never, {
      ...satoriOptions(),
      height: SOCIAL_IMAGE_HEIGHT * 2,
      embedFont: false,
    })
    const words = [...svg.matchAll(/<text x="([\d.]+)" y="([\d.]+)" width="([\d.]+)"[^>]*>([^<]*)<\/text>/g)]
      .filter((m) => m[4].trim() !== '')
      .map((m) => ({ y: Number(m[2]), right: Number(m[1]) + Number(m[3]) }))
    const lines = new Set(words.map((w) => w.y)).size
    if (lines <= MAX_TITLE_LINES && words.every((w) => w.right <= TITLE_WIDTH + 0.5)) {
      return { fontSize, lines }
    }
  }
  return null
}

export function socialImageNode(title: string, label: string | null, fontSize: number): Node {
  const colors = lightThemeColors()
  // A path fixed in this script, not input.
  // eslint-disable-next-line security/detect-non-literal-fs-filename
  const logo = readFileSync(new URL('../public/img/logo-light.svg', import.meta.url))
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
        padding: `64px ${PADDING_X}px`,
        backgroundColor: colors.bg,
      },
      children: [
        {
          type: 'img',
          props: {
            src: `data:image/svg+xml;base64,${logo.toString('base64')}`,
            style: { width: LOGO_HEIGHT * LOGO_ASPECT, height: LOGO_HEIGHT },
          },
        },
        {
          type: 'div',
          props: {
            style: { display: 'flex', flexDirection: 'column', gap: 20 },
            children: [
              label
                ? { type: 'div', props: { style: { ...mono, fontSize: 28, color: colors.accent }, children: label } }
                : null,
              titleNode(title, fontSize, colors.text),
            ],
          },
        },
        {
          type: 'div',
          props: {
            style: {
              ...mono,
              display: 'flex',
              paddingTop: 22,
              borderTop: `2px solid ${colors.border}`,
              fontSize: 26,
              color: colors.dim,
            },
            children: 'checkmeup.net',
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
