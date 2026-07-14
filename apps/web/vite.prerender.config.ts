import { defineConfig, mergeConfig } from 'vite'
import baseConfig from './vite.config'

// Used only by `vite-node` to run scripts/prerender.mts (see package.json's
// build script) — never by `vite dev`/`vite build`. Without `ssr.noExternal`,
// Vite resolves `vue-router` (and friends) two different ways in this
// context: once via its client-oriented dep-optimizer cache for the .vue
// components rendered through the app, and once via plain node_modules
// resolution for the prerender script's own top-level imports. That split
// produces two separate vue-router module instances, so RouterView's
// provide/inject symbols don't match and rendering throws
// ("injection 'router view location' not found"). Forcing everything
// through Vite's SSR transform pipeline keeps it to a single instance.
export default defineConfig(
  mergeConfig(baseConfig, {
    ssr: {
      noExternal: true,
    },
    // Prevents Vite from redirecting some `vue-router` imports to its
    // pre-bundled `.vite/deps/vue-router.js` copy while others resolve
    // straight from node_modules — that split is what caused the dual
    // module instance above. A one-shot build script doesn't need the
    // dep-optimizer's cold-start speedup anyway.
    optimizeDeps: {
      disabled: true,
    },
  }),
)
