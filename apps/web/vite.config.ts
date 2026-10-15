import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'node:path'

export default defineConfig({
  plugins: [tailwindcss(), vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
    },
  },
  server: {
    // true listens on every address, IPv6 included: browsers that resolve
    // localhost to ::1 first can't reach a server bound to 0.0.0.0 alone.
    host: true,
    port: 5173,
    proxy: {
      // The landing hero's live status strip and platform totals come from
      // production in dev: a local database has no checkmeup-net status
      // page (ownStatusSlug in src/content/home.ts) and too little traffic
      // to clear the stats thresholds, so the hero would only ever show its
      // example. Both are read-only public endpoints. Listed before '/api'
      // and '/status/' because the first matching entry wins.
      // LANDING_DATA=local sends them to the local API instead, for work on
      // the endpoints themselves.
      ...(process.env.LANDING_DATA === 'local'
        ? {}
        : {
            '/status/checkmeup-net/summary.json': {
              target: 'https://checkmeup.net',
              changeOrigin: true,
            },
            '/api/v1/stats': {
              target: 'https://checkmeup.net',
              changeOrigin: true,
            },
          }),
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      // The public status page is server-rendered by the Go API (no Vue
      // route for it) and references root-relative static assets like
      // /status-theme.js and /favicon.svg. Proxying only the HTML through
      // keeps those requests on this origin, where Vite already serves
      // apps/web/public/ — in production STATIC_DIR makes the API serve
      // them itself, but locally STATIC_DIR is unset (see config/deploy.yml
      // vs apps/api/.env), so hitting :8080/status/:slug directly 404s on
      // every asset. Visit localhost:5173/status/:slug in dev instead.
      // Trailing slash matters: the admin SPA route /status-pages must NOT
      // be caught by this — it starts with "/status" too, just without a
      // "/" right after, so a bare "/status" prefix would wrongly proxy it.
      '/status/': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
