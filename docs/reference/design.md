---
title: Design System
type: reference
status: active
updated: 2026-07-05
tags: [design, tokens, colors, logo]
---

# Design system

## Logo assets

All logo files live in `assets/` — do not recreate or modify them:

| File | Use |
|---|---|
| `assets/logo-light.svg` | On light backgrounds (wordmark `#333333`) |
| `assets/logo-dark.svg` | On dark backgrounds (wordmark `#DDDDDD`) |
| `assets/logo-grey.svg` | Monochrome / watermark contexts |
| `assets/logo-icon.svg` | Favicon, app icon, square placements |

The icon is a stylized `C`/bracket that morphs into a checkmark. Two greens in the icon:

| Token | Hex | Use in logo |
|---|---|---|
| `green` | `#1D9E75` | Checkmark / light part |
| `green-dark` | `#0F6E56` | Bracket / shadow part |

## Color palette

Redesigned July 2026 ("Redesign v2") to a lower-contrast, near-black/near-white neutral
scale with translucent surfaces, replacing the earlier flat-hex GitHub-dark-style palette.

**Brand greens** (fixed hex, both themes — the raw scale never changes)
| Token | Hex |
|---|---|
| `--green-100` | `#D3F5E9` |
| `--green-300` | `#4DC9A0` |
| `--green-500` | `#1D9E75` |
| `--green-700` | `#0F6E56` |
| `--green-900` | `#08392E` |

**Status colors** (semantic — never rebrand these)
| Token | Hex | Meaning |
|---|---|---|
| `--status-up` | `#1D9E75` | Up / healthy |
| `--status-degraded` | `#F59E0B` | Degraded / slow |
| `--status-down` | `#EF4444` | Down / error |
| `--status-down-wash` | `rgba(239,68,68,0.10)` | Down/error tinted background (e.g. a destructive button) |
| `--status-paused` | `#94A3B8` | Paused / maintenance |

**Neutrals — dark** (default; bare `:root` in `style.css`)
| Token | Value | Use |
|---|---|---|
| `--bg` | `#0A0F0D` | Page background |
| `--surface` | `#101512` | Nav-adjacent chrome, the hero product-screenshot mockup |
| `--surface-raised` | `#1A221D` | Further-elevated chrome (mockup's internal header bar), hover backgrounds |
| `--card` | `rgba(255,255,255,0.035)` | Feature/testimonial/pricing card fill — translucent, sits directly on `--bg` |
| `--border` | `rgba(255,255,255,0.08)` | Dividers, input borders |
| `--text-muted` | `rgba(242,245,243,0.28)` | Placeholders, captions |
| `--text-dim` | `rgba(242,245,243,0.55)` | Secondary text |
| `--text` | `#F2F5F3` | Body text |
| `--text-strong` | `#F2F5F3` | Headings, emphasis — same value as `--text` in this palette |

**Neutrals — light** (`:root[data-theme='light']` override, see [EP-10](../stories/ep-10-theme.md))
| Token | Value | Use |
|---|---|---|
| `--bg` | `#FBFDFC` | Page background |
| `--surface` | `#F2F6F4` | Nav-adjacent chrome, the hero product-screenshot mockup |
| `--surface-raised` | `#E7EDE9` | Further-elevated chrome, hover backgrounds |
| `--card` | `rgba(0,0,0,0.03)` | Feature/testimonial/pricing card fill |
| `--border` | `rgba(0,0,0,0.08)` | Dividers, input borders |
| `--text-muted` | `rgba(11,15,12,0.45)` | Placeholders, captions |
| `--text-dim` | `rgba(11,15,12,0.70)` | Secondary text |
| `--text` | `#0D1512` | Body text |
| `--text-strong` | `#0D1512` | Headings, emphasis |

**Note:** `--text-muted`/`--text-dim`'s light-mode alpha is *not* the same percentage as
dark mode's (dark: 28%/55%; light: 45%/70%) — a straight port of the dark-mode alpha onto a
light background under-shoots contrast badly (measured: `--text-dim` fell from a 7.24:1
ratio in the pre-redesign solid-color palette to 4.08:1, below WCAG AA's 4.5:1 for normal
text; `--text-muted` fell from 2.45:1 to 1.88:1). Dark-on-light and light-on-dark alpha
blending aren't symmetric in sRGB — don't "fix" these back to match dark mode's percentage.

The brand-green scale and status colors above are fixed hex in both themes. The **semantic
accent** below is the one exception: it resolves to a different green step per theme.

**Accent**
| Token | Dark | Light | Use |
|---|---|---|---|
| `--accent` | `--color-green-500` | `--color-green-700` | Buttons, links, selected/active state (checkboxes, progress bars) — darker in light mode for contrast against a near-white background |
| `--accent-emphasis` | `#6EE7B7` | `--color-green-700` (= `--accent`) | Extra-bright headline-highlight text — dark mode only; a bright mint has no contrast on white, so light mode falls back to `--accent` |
| `--accent-deep` | `--color-green-700` | `--color-green-700` | Alias, same both themes |
| `--accent-wash` | `rgba(29,158,117,0.13)` | `rgba(15,110,86,0.10)` | Icon badges, pills, progress-track background |
| `--accent-wash-dim` | `rgba(29,158,117,0.065)` | `rgba(15,110,86,0.05)` | Larger-area tint — CTA banner, status-operational banner, highlighted pricing card |
| `--on-accent` | `#FFFFFF` | `#FFFFFF` | Text/icon color on top of `--accent` — no override needed, white reads on every accent step used |

**Nav**
| Token | Dark | Light | Use |
|---|---|---|---|
| `--nav-bg` | `rgba(10,15,13,0.88)` | `rgba(251,253,252,0.88)` | Sticky nav background under `backdrop-filter: blur(...)` |

**CTA banner** (`HomeView`, `PricingView`, `AboutView` — the "Start monitoring..." box; see [EP-10](../stories/ep-10-theme.md))

A flat `--accent-wash-dim` fill with a `--cta-border` outline, same both themes — no
separate CTA-only text tokens needed, since heading/body just use the normal
`--text-strong`/`--text-dim` tokens on top of the wash.

| Token | Value (both themes) | Use |
|---|---|---|
| `--cta-border` | `rgba(29,158,117,0.25)` | Box border |

This replaces the earlier dark-gradient-in-dark-mode / light-mint-in-light-mode treatment
(see `../reports/2026-06.md`, EP-10 entry, for why a flat `--bg`/`--surface` fill didn't work
there) — the flat wash reads fine in both themes without a per-theme gradient swap.

## Typography

- **Sans:** `Inter` (400/500/600/700/800), loaded via Google Fonts `<link>` tags in
  `apps/web/index.html`. Falls back to the system UI stack.
- **Mono:** `IBM Plex Mono` (400/500), wired as Tailwind's `--font-mono` theme token in
  `style.css` — every `font-mono` utility class (timestamps, ping URLs, uptime stats)
  picks it up automatically. Falls back to `ui-monospace`/`SFMono-Regular`.

Note: the public status page (`apps/api/internal/handler/status_public.go`) is a separate, server-rendered template with its own fixed light palette (hardcoded, not these tokens) — it predates and is out of scope for EP-10's theme toggle. See EP-10's epic note for why.
