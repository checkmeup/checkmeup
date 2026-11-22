// Draws the homepage's social card, public/img/checkmeup-og.png, in the look
// of every other page's (scripts/social-image.mts, PRD social-images): the
// landing hero's headline, its cron line and missed-ping alert, and a 90-day
// strip with the one miss. Run by hand when the hero changes:
//   bun scripts/homepage-card.mts
// The PNG is committed, not built: it is also the card a page falls back to
// when its own image fails to render, so it can't depend on the renderer.
import { writeFileSync } from 'node:fs'
import { Resvg } from '@resvg/resvg-js'
import satori from 'satori'
import {
  type Node,
  SOCIAL_IMAGE_HEIGHT,
  SOCIAL_IMAGE_WIDTH,
  checkStrip,
  logoNode,
  markSrc,
  palette,
  satoriOptions,
} from './social-image.mts'

const colors = palette()
const mono = { fontFamily: 'Geist Mono', fontWeight: 400 }
const sans = (fontWeight: 400 | 600) => ({ fontFamily: 'Geist', fontWeight })

const div = (style: Record<string, unknown>, children?: Node['props']['children']): Node => ({
  type: 'div',
  props: { style: { display: 'flex', ...style }, children },
})

const hero = div({ flexDirection: 'column', width: 500 }, [
  div({ ...mono, fontSize: 18, color: colors.dim }, 'cron · uptime · ssl · domain · port · dns'),
  div(
    {
      ...sans(600),
      flexDirection: 'column',
      marginTop: 16,
      fontSize: 64,
      lineHeight: 1.04,
      letterSpacing: -2.2,
      color: colors.text,
    },
    [div({}, 'Know before'), div({}, 'your client does.')],
  ),
  div(
    { ...sans(400), marginTop: 18, fontSize: 22, lineHeight: 1.45, color: colors.dim },
    'Add one line to a cron job, or paste a URL. When something stops, you hear about it first.',
  ),
  div({ alignItems: 'center', gap: 24, marginTop: 26 }, [
    div(
      {
        ...sans(600),
        padding: '13px 24px',
        borderRadius: 10,
        fontSize: 20,
        backgroundColor: colors.accent,
        color: colors.onAccent,
      },
      'Start free — no credit card',
    ),
    div({ ...mono, fontSize: 20, color: colors.dim }, 'checkmeup.net'),
  ]),
])

const terminal = div(
  {
    ...mono,
    flexDirection: 'column',
    gap: 9,
    padding: '18px 22px',
    borderRadius: 14,
    border: `1px solid ${colors.border}`,
    backgroundColor: colors.terminalBg,
    fontSize: 17,
    color: colors.text,
  },
  [
    div({ justifyContent: 'space-between', marginBottom: 4, color: colors.terminalDim }, [
      div({}, 'crontab -e'),
      div({}, 'server-01'),
    ]),
    div({ color: colors.terminalDim }, '# nightly database backup'),
    div({}, '0 3 * * * /opt/backup.sh'),
    div({ paddingLeft: 20, color: colors.accent }, '&& curl -fsS checkmeup.net/ping/k7f2c9'),
  ],
)

// A light card, as the alert stands out on the hero: the logo's tile again.
const alert = div(
  {
    gap: 16,
    padding: '18px 20px',
    borderRadius: 14,
    backgroundColor: colors.text,
    color: colors.onAccent,
  },
  [
    { type: 'img', props: { src: markSrc(), style: { width: 40, height: 40 } } },
    div({ flexDirection: 'column', flexGrow: 1, gap: 6 }, [
      div({ justifyContent: 'space-between' }, [
        div({ ...sans(600), fontSize: 17 }, 'Checkmeup'),
        div({ ...mono, fontSize: 16, color: colors.terminalDim }, '03:05'),
      ]),
      div({ ...sans(600), alignItems: 'center', gap: 10, fontSize: 20 }, [
        div({ width: 10, height: 10, borderRadius: 5, backgroundColor: colors.statusDown }),
        'nightly-backup missed its 03:00 ping',
      ]),
      div(
        { ...sans(400), fontSize: 14, color: colors.terminalDim },
        'Last ping yesterday at 03:00 · grace period 5 min · Telegram',
      ),
    ]),
  ],
)

const story = div({ flexDirection: 'column', width: 520 }, [
  terminal,
  div(
    { ...mono, margin: '14px 0 14px 20px', fontSize: 16, color: colors.dim },
    '↓ 03:00 passes, no ping arrives',
  ),
  alert,
])

const strip = div({ flexDirection: 'column', gap: 12 }, [
  div({ justifyContent: 'space-between', fontSize: 18, color: colors.dim }, [
    div({ ...sans(400) }, [
      div({ ...sans(600), color: colors.text }, 'nightly-backup'),
      // A no-break space: satori trims a plain one at the start of a text.
      '\u00a0· Cron · last 90 days',
    ]),
    div({ ...sans(400) }, '1 missed ping'),
  ]),
  checkStrip(90, 4, colors.tickUp, { at: 61, color: colors.statusDown }),
])

const card = div(
  {
    flexDirection: 'column',
    justifyContent: 'space-between',
    boxSizing: 'border-box',
    width: SOCIAL_IMAGE_WIDTH,
    height: SOCIAL_IMAGE_HEIGHT,
    padding: '52px 72px 52px',
    backgroundImage: colors.background,
    borderTop: `8px solid ${colors.accent}`,
  },
  [
    logoNode(colors),
    div({ justifyContent: 'space-between', alignItems: 'center' }, [hero, story]),
    strip,
  ],
)

const svg = await satori(card as never, satoriOptions())
// Twice the size, as the hand-made card was, for sharp text on a dense screen.
const png = new Resvg(svg, { fitTo: { mode: 'width', value: SOCIAL_IMAGE_WIDTH * 2 } })
  .render()
  .asPng()
const out = process.argv[2] ?? new URL('../public/img/checkmeup-og.png', import.meta.url).pathname
// The card's own path, or one given on the command line to preview it.
// eslint-disable-next-line security/detect-non-literal-fs-filename
writeFileSync(out, png)
console.log(`${out}: ${png.length} bytes`)
