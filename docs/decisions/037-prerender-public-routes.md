# ADR-037: Build-time prerender of public routes via vite-node + vue/server-renderer

**Date:** 2026-07-14
**Status:** Accepted

---

## Context

`apps/web` is a pure client-rendered Vue SPA — `index.html` ships `<div id="app"></div>` empty, with content only appearing once the browser runs the app's JS. Bing's crawler flagged the homepage for a missing `<h1>`, which turned out not to be a markup bug (every route's component does render a real `<h1>`) but this architectural gap: Bingbot doesn't reliably execute JS, so it sees nothing.

This is the same root cause already tracked as an open decision ("Blog prerendering approach" in `docs/decisions/backlog.md`, blocking [EP-36](../stories/ep-36-blog-prerendering.md)), which was scoped narrowly to blog posts only (for social-preview unfurls) on the assumption that Google's JS-executing crawler already indexed marketing pages correctly. The Bing finding shows that assumption doesn't hold for every crawler, so this decision covers all public/indexable routes at once: `/`, `/pricing`, `/docs`, `/faq`, `/about`, `/terms`, `/privacy`, `/refund`, `/blog`, and `/blog/:slug` for every post. Auth-gated app routes (dashboard, monitors, etc.) are never crawled and stay pure CSR.

## Decision

Prerender the routes above to static HTML at build time with a small custom script (`apps/web/scripts/prerender.mts`), run via `vite-node` after `vite build`:

- `vue/server-renderer`'s `renderToString` renders each route's real component tree (through a guard-free `createRouter({ history: createMemoryHistory(), routes })` built from the app's own route table, now split into a side-effect-free `src/router/routes.ts`).
- `@unhead/vue`'s already-installed `/server` subpath (`createHead`, `transformHtmlTemplate`) captures each route's `useSeo()` output (title/description/OG/canonical) and merges it into the built `dist/index.html` template.
- Output is written as `dist/index.html` (for `/`) or `dist/<route>/index.html` for everything else — matching the directory + `index.html` shape Go's existing `handleSPA` (`apps/api/internal/server/server.go`) already serves for free: it `os.Stat`s the request path and serves it directly if present, only falling back to the SPA shell on a miss. No backend code changes needed.
- `vite-node` (already transitively present via `vitest`, added as an explicit devDependency) runs the script against the real Vite module graph so it can import actual `.vue`/`.ts` source directly — a plain `node` script can't parse SFCs or strip TypeScript. A dedicated `vite.prerender.config.ts` (merged from the app's real `vite.config.ts`) sets `ssr.noExternal: true` and `optimizeDeps.disabled: true`, required to avoid Vite resolving `vue-router` two different ways in this context (one via its client-oriented dep-optimizer cache for `.vue` components, one via plain `node_modules` resolution for the script's own imports) — that split produced two separate `vue-router` module instances and broke `RouterView`'s provide/inject.
- `src/lib/theme.ts` and `src/lib/consent.ts` read `window`/`document`/`localStorage` eagerly at module-import time; both are unconditionally pulled into the render path (`App.vue` calls `useConsent()` directly, `LandingLayout.vue` calls `useTheme()`) and now guard that access with `typeof window === 'undefined'` checks, falling back to a fixed default.
- Wired into both build entry points in lockstep — `apps/web/package.json`'s `build` script and the root `Dockerfile`'s frontend build stage — since a prerender step added to one but not the other previously broke exactly this way (PR #69, sitemap generation).

Rejected `vite-plugin-ssg` and full SSR as unnecessary weight for ~10 static content routes with no live data dependencies, consistent with this repo's bias against adding infra beyond what's needed ([ADR-001](001-worker-model.md), [ADR-004](004-sqlc-over-orm.md)).

## Consequences

- Closes the Bing "H1 missing" notice and [EP-36](../stories/ep-36-blog-prerendering.md)'s social-preview gap in one change — both were the same root cause.
- Zero new runtime dependencies; `vite-node` is a devDependency-only build tool.
- Real users' browsers still run the existing `createApp(App).mount('#app')` in `main.ts` unchanged — a client render, not a hydration, so it briefly replaces the prerendered markup rather than hydrating it. Standard tradeoff for this lightweight prerender-for-crawlers pattern; not a regression from today's already-existing blank-then-render CSR behavior.
- The prerender step adds a few seconds to the frontend build; acceptable for a one-shot production build (not `vite dev`, which is untouched).

## Alternatives considered

1. `vite-plugin-ssg` — a full static-site-generation framework; more moving parts (its own build integration, hydration model, routing conventions) than ~10 static routes justify.
2. Full SSR (a persistent Node render server) — meaningfully bigger architectural change (new runtime process, request-time rendering) for content that doesn't change per-request; nothing here needs live data at request time.
3. A minimal static `<h1>`/fallback paragraph hand-written directly into `index.html` — cheap, but only fixes the homepage; every other public route would still serve the same empty shell to a non-JS crawler, and it doesn't address EP-36's per-post OG tag gap at all.
