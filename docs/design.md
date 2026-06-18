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

**Brand greens**
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
| `--status-paused` | `#94A3B8` | Paused / maintenance |

**Neutrals — dark** (default; bare `:root` in `style.css`)
| Token | Hex | Use |
|---|---|---|
| `--bg` | `#0D1117` | Page background |
| `--surface` | `#161B22` | Cards, panels |
| `--surface-raised` | `#1E2530` | Dropdowns, tooltips |
| `--border` | `#2D3748` | Dividers, input borders |
| `--text-muted` | `#718096` | Placeholders, captions |
| `--text-dim` | `#A0AEC0` | Secondary text |
| `--text` | `#E2E8F0` | Body text |
| `--text-strong` | `#FFFFFF` | Headings, emphasis |

**Neutrals — light** (`:root[data-theme='light']` override, see [EP-10](stories/ep-10-theme.md))
| Token | Hex | Use |
|---|---|---|
| `--bg` | `#F8FAFC` | Page background |
| `--surface` | `#FFFFFF` | Cards, panels |
| `--surface-raised` | `#F1F5F9` | Dropdowns, tooltips |
| `--border` | `#E2E8F0` | Dividers, input borders |
| `--text-muted` | `#94A3B8` | Placeholders, captions |
| `--text-dim` | `#475569` | Secondary text |
| `--text` | `#1E293B` | Body text |
| `--text-strong` | `#0F172A` | Headings, emphasis |

Brand greens and status colors are unchanged in both themes — never swap those per-theme.

**Accent**
| Token | Value | Use |
|---|---|---|
| `--accent` | `var(--color-green-500)` | Selected/active state (checkboxes, progress bars) — alias, not a new color |

**CTA banner** (`HomeView`, `PricingView`, `AboutView` — the "Start monitoring..." gradient box; see [EP-10](stories/ep-10-theme.md))
| Token | Dark | Light | Use |
|---|---|---|---|
| `--cta-gradient-start` | `--color-green-900` | `--color-green-100` | Gradient start |
| `--cta-gradient-end` | `--color-green-700` | `--surface` | Gradient end |
| `--cta-border` | `--color-green-700` | `--color-green-300` | Box border |
| `--cta-text` | `#FFFFFF` | `--text-strong` | Heading |
| `--cta-text-dim` | `rgba(255,255,255,0.8)` | `--text-dim` | Subtitle |

This box is intentionally always a green wash (dark gradient in dark mode, light mint in light mode) rather than using `--bg`/`--surface` directly — those two tokens flip from dark to white between themes, which previously broke this banner's contrast in light mode (see `reports/2026-06.md`, EP-10 entry).

Note: the public status page (`apps/api/internal/handler/status_public.go`) is a separate, server-rendered template with its own fixed light palette (hardcoded, not these tokens) — it predates and is out of scope for EP-10's theme toggle. See EP-10's epic note for why.
