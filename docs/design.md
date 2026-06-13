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

**Neutrals** (dark-first UI)
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
