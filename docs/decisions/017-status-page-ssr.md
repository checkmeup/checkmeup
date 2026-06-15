# ADR-017: Public status page rendered by Go HTML template

**Status:** Accepted  
**Date:** 2026-06-15

## Context

US-0603 requires the public `/status/:slug` page to render correctly without JavaScript. Options considered:

1. **Vue SPA (client-side only)** — fast to build, but requires JS; crawlers and users with JS disabled see nothing
2. **Nuxt / SSR layer** — full SSR support but adds a second server process and significant complexity
3. **Go `html/template`** — minimal dependencies, renders server-side HTML the Go API already serves, works without JS

## Decision

Serve `/status/:slug` as a Go `html/template` response from the existing API process.

The route is registered _before_ the SPA catch-all so the Go handler takes priority. The admin dashboard (list, create, edit) is still served as Vue SPA routes; only the public-facing page uses a Go template.

## Consequences

- **No JS dependency on the public page** — works in curl, feed readers, status bots
- **Zero additional infrastructure** — same API process, no Node.js SSR server
- **Styling is inline CSS in the template** — no PostCSS/Tailwind pipeline for the public page; manually maintained
- **Admin UI unchanged** — Vue SPA handles the management interface normally
