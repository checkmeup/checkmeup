---
title: Design System
type: reference
status: active
updated: 2026-10-18
tags: [design, tokens, colors, logo]
---

# Design system

## Logo assets

All logo files live in `apps/web/src/assets/`, except the two theme wordmarks, which live in `apps/web/public/img/` so the prerendered HTML can reference them by a stable URL (an imported asset resolves to a dev-only `/src/assets/…` path during prerender). Do not recreate or modify them:

| File | Use |
|---|---|
| `public/img/logo-light.svg` | On light backgrounds (wordmark `#333333`) |
| `public/img/logo-dark.svg` | On dark backgrounds (wordmark `#DDDDDD`) |
| `assets/logo-icon.svg` | Favicon, app icon, square placements |

The icon is a stylized `C`/bracket that morphs into a checkmark. Two greens in the icon:

| Token | Hex | Use in logo |
|---|---|---|
| `green` | `#1D9E75` | Checkmark / light part |
| `green-dark` | `#0F6E56` | Bracket / shadow part |

## Color palette

Redesigned October 2026 ("Redesign v3", the hybrid direction: a calm, light-first
interface with dense data tables). Opaque neutrals replace v2's translucent ones, and
the status colors now switch per theme like the neutrals.

**Brand greens** (Tailwind theme colors in `style.css`'s `@theme`; fixed hex, both themes)

| Token | Hex |
|---|---|
| `--color-green-100` | `#D3F5E9` |
| `--color-green-300` | `#4DC9A0` |
| `--color-green-500` | `#1D9E75` |
| `--color-green-700` | `#0F6E56` |
| `--color-green-900` | `#08392E` |

**Status colors** (semantic — never rebrand these). Each resolves per theme so it also
reads as text on `--bg` (WCAG AA, 4.5:1), not just as a dot or fill.

| Token | Dark | Light | Meaning |
|---|---|---|---|
| `--status-up` | `#3FBF8F` | `#12805F` | Up / healthy |
| `--status-degraded` | `#EBAE4A` | `#A86200` | Degraded / slow / expiring |
| `--status-down` | `#F0605A` | `#C93A2E` | Down / error |
| `--status-down-wash` | `#3A1A18` | `#FBE6E4` | Down/error tinted background |
| `--status-paused` | `#7A847F` | `#6F7873` | Paused / maintenance |
| `--on-status-down` | `#1A0807` | `#FFFFFF` | Text/icon on a `--status-down` fill (destructive button) |

**Neutrals** (dark is the bare `:root` in `style.css`; light is the
`:root[data-theme='light']` override, see [PRD theme](../../.about/prd/theme.md))

| Token | Dark | Light | Use |
|---|---|---|---|
| `--bg` | `#0E1211` | `#FFFFFF` | Page background |
| `--surface` | `#131816` | `#F7F9F7` | Table headers, nav-adjacent chrome |
| `--surface-raised` | `#1A201E` | `#F1F4F1` | Active nav item, hover backgrounds |
| `--card` | `rgb(236 240 238 / 3%)` | `rgb(17 23 20 / 3%)` | Feature/testimonial/pricing card fill |
| `--border` | `#232A27` | `#E4E8E4` | Dividers, table rules |
| `--border-strong` | `#323B37` | `#D6DCD7` | Input and secondary-button outlines |
| `--text-muted` | `#7A847F` | `#6F7873` | Placeholders, captions |
| `--text-dim` | `#9AA39E` | `#59625D` | Secondary text |
| `--text` | `#ECF0EE` | `#111714` | Body text |
| `--text-strong` | `#ECF0EE` | `#111714` | Headings, emphasis — same value as `--text` |

Every text token meets WCAG AA (4.5:1) on `--bg` in its theme — `--text-muted` included,
unlike v2, where captions fell to 1.9–2.5:1.

**Accent**

| Token | Dark | Light | Use |
|---|---|---|---|
| `--accent` | `--color-green-300` | `--color-green-700` | Buttons, links, selected/active state — the light mint reads on dark, the deep green on white |
| `--accent-hover` | `#6EE7B7` | `#0C5E49` | Hover fill of an `--accent` button |
| `--on-accent` | `#06201A` | `#FFFFFF` | Text/icon on an `--accent` fill — dark ink on dark mode's mint, white on light mode's green |
| `--accent-emphasis` | `#6EE7B7` | `--color-green-700` | Extra-bright headline-highlight text — dark mode only |
| `--accent-deep` | `--color-green-700` | `--color-green-700` | Alias, same both themes |
| `--accent-wash` | `rgb(77 201 160 / 12%)` | `rgb(15 110 86 / 10%)` | Icon badges, pills, progress-track background |
| `--accent-wash-dim` | `rgb(77 201 160 / 6%)` | `rgb(15 110 86 / 5%)` | Larger-area tint — CTA banner, highlighted pricing card |

**Nav**

| Token | Dark | Light | Use |
|---|---|---|---|
| `--nav-bg` | `rgb(14 18 17 / 88%)` | `rgb(255 255 255 / 88%)` | Sticky nav background under `backdrop-filter: blur(...)` |

**CTA banner** (`HomeView`, `PricingView`, `AboutView` — the "Start monitoring..." box; see [PRD theme](../../.about/prd/theme.md))

A flat `--accent-wash-dim` fill with a `--cta-border` outline, same both themes — no
separate CTA-only text tokens needed, since heading/body just use the normal
`--text-strong`/`--text-dim` tokens on top of the wash.

| Token | Value (both themes) | Use |
|---|---|---|
| `--cta-border` | `rgba(29,158,117,0.25)` | Box border |

## App layout

- **Top navigation**, not a sidebar (`AppLayout.vue`): Overview, Monitors, Status pages,
  Incidents, Maintenance; Settings, Billing and sign-out sit in the account menu. On phones
  the same links open from a menu button.
- **One monitor list** at `/monitors` for every type, filtered by `?type=`. The old per-type
  list URLs (`/monitors/cron`, …) redirect there with their type preselected, and keep their
  route names so `router.push({ name: 'cron-monitors' })` still works.
- **Problems first**: every monitor table sorts down, then warnings, then the rest by name
  (`sortRows` in `lib/monitorRows.ts`), so the rows that need action are always on top.
- **No cards for content**: sections are separated by whitespace and hairline `--border`
  rules; tables sit in one bordered box with a `--surface` header row.
- **Check strips** (`CheckStrip.vue`): a row of ticks, oldest on the left — soft green
  (`--tick-up`) for a passing check or hour, `--status-down` for a failure, `--tick-empty`
  for none. Monitor tables show the last 48 hours from `GET /api/v1/monitors/check-history`;
  monitor pages show their recent checks. SSL and domain monitors keep no check history and
  show no strip.
- **Headings** are Geist 600 with tight negative tracking (about −0.035em on page titles,
  −0.02em on section titles); statuses are a colored dot plus a label (`StatusLabel.vue`).
- **Content pages** (About, Blog, Changelog, blog posts, Terms, Privacy, Refund) share the
  1200px page grid with the landing page: a left-aligned `PageHero`, then rows with a 260px
  label column and a text column capped at 680px (`LabeledSection.vue`, the same component
  Settings and Billing use). Docs and FAQ keep their sticky section nav in that left column.
- **Landing hero** (`HeroSection.vue`) shows the product's one moment of value instead of a
  dashboard mockup: a crontab line ending in the ping `curl`, in a terminal panel
  (`--terminal-*` tokens: dark in both themes), and the missed-ping alert it produces. Under it,
  a live 90-day strip of checkmeup.net's own status page (`GET /status/checkmeup-net/summary.json`,
  slug in `content/home.ts`), which falls back to the example job's strip when that page doesn't
  answer, as in local dev. The
  monitor types below it are a spec table (`monitorTypes` in `content/home.ts`), not icon cards.

## Typography

- **Sans:** `Geist` (400/500/600/700/800) for body and headings alike — headings are
  set apart by weight (600) and tight tracking, not by a second typeface. Wired as
  Tailwind's `--font-sans` theme token in `style.css`, loaded via Google Fonts `<link>`
  tags in `apps/web/index.html`. Falls back to the system UI stack.
- **Mono:** `Geist Mono` (400/500), wired as Tailwind's `--font-mono` theme token in
  `style.css` — every `font-mono` utility class (timestamps, ping URLs, uptime stats)
  picks it up automatically. Falls back to `ui-monospace`/`SFMono-Regular`.

Note: the public status page (`apps/api/internal/handler/status_public.go`) is a separate, server-rendered Go template. It carries its own copy of these tokens (light and dark, switched by `status-theme.js`) and loads Geist from Google Fonts; statuses reach it as the badge hex colors and are restyled through `tone-*` classes (`statusTone`), so keep its token block in sync with `style.css`. It follows the app's layout: a top bar with a hairline under it, a 44px sentence headline, active incidents as hairline rows, and a 22px "Monitors" section.

The embeddable status badges (`status_public_badges.go`) can't switch with a theme, since they sit on whatever page embeds them, so they always use the light theme's status colors, which carry white text at AA contrast, with `--text-dim` (light) as the label segment. Keep `badgeFill` in sync with the light status tokens.
