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
    host: '0.0.0.0',
    port: 5173,
    proxy: {
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
