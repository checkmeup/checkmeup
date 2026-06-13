# checkmeup.net — developer monitoring platform

## Product

Cron, uptime, and SSL monitors with execution logs, Telegram alerts, and white-label status pages for agencies.

**MVP order:** Cron monitor → Uptime monitor → SSL expiry monitor → Status page

**Pricing:** Hobbyist $0 / Indie $12 / Studio $39 / Agency $99

---

## Monorepo structure

```
apps/
  api/          # Go backend
  web/          # Vue 3 frontend
packages/
  ui/           # shadcn-vue components (white-label reuse)
  types/        # shared TypeScript types
```

Turborepo orchestrates all tasks including Go via `package.json` shims in each app.

---

## Backend — `apps/api`

| Concern     | Choice                                                         |
| ----------- | -------------------------------------------------------------- |
| Language    | Go                                                             |
| HTTP router | Chi                                                            |
| DB queries  | sqlc (write SQL → get typed Go)                                |
| Migrations  | goose                                                          |
| DB          | PostgreSQL                                                     |
| Auth        | JWT in httpOnly cookie + refresh token in DB                   |
| Alerts      | Telegram Bot API (direct HTTP)                                 |
| Workers     | goroutines + `time.Ticker` per monitor (no Redis/queue on MVP) |
| Hot reload  | air                                                            |
| Lint        | golangci-lint                                                  |

**Worker model:** each active monitor gets a goroutine with a ticker. Cron monitors wait for inbound pings; a separate ticker checks for missed pings. All state in PostgreSQL — no external queue needed for MVP.

---

## Frontend — `apps/web`

| Concern      | Choice                                 |
| ------------ | -------------------------------------- |
| Framework    | Vue 3 + TypeScript                     |
| Build        | Vite                                   |
| State        | Pinia                                  |
| Router       | Vue Router                             |
| Server state | TanStack Query (`@tanstack/vue-query`) |
| UI           | Radix Vue + Tailwind CSS               |
| Components   | shadcn-vue (in `packages/ui`)          |

---

## Infrastructure

| Concern             | Choice                  |
| ------------------- | ----------------------- |
| Server              | Hetzner CX23 (€3.49/mo) |
| Deploy              | Kamal                   |
| Reverse proxy / SSL | Traefik                 |

---

## Multi-tenancy

Single PostgreSQL database with `org_id` columns — no schema-per-tenant on MVP.

## Status page

Served at `checkmeup.net/status/:slug` — no separate subdomain on MVP. White-label via custom domain mapping later.

---

## Design

### Logo assets

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

### Website palette

**Brand greens**
| Token | Hex |
|---|---|
| `--green-100` | `#D3F5E9` |
| `--green-300` | `#4DC9A0` |
| `--green-500` | `#1D9E75` |
| `--green-700` | `#0F6E56` |
| `--green-900` | `#08392E` |

**Status colors** (semantic, never rebrand these)
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
