// Runs after `vite build` (via `vite-node`, so it can import real .vue/.ts
// source through the same Vite pipeline/aliases as the app itself — a plain
// `node` script can't parse SFCs or strip TypeScript). Renders each public,
// indexable route to static HTML so crawlers that don't execute JS (Bingbot
// flagged the homepage's missing <h1> — see ADR-037) see real content
// instead of the empty `<div id="app"></div>` shell `vite build` produces.
//
// Auth-gated app routes (dashboard, monitors, etc.) are never crawled and
// stay pure client-rendered — only the routes below get a static snapshot.
import { readFileSync, writeFileSync, mkdirSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { createApp } from 'vue'
import { renderToString } from 'vue/server-renderer'
import { createPinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import { createHead, transformHtmlTemplate } from '@unhead/vue/server'
import App from '../src/App.vue'
import { routes } from '../src/router/routes'
import { posts } from '../src/blog/posts'

const __dirname = dirname(fileURLToPath(import.meta.url))
const distDir = join(__dirname, '../dist')
const template = readFileSync(join(distDir, 'index.html'), 'utf-8')

// Keep in sync with scripts/generate-sitemap.mjs's staticPaths — that script
// derives the same list independently (by filesystem-scanning src/blog/posts
// rather than importing the typed module) for the unrelated job of writing
// sitemap.xml; both lists should keep matching the indexable routes below.
const staticPaths = ['/', '/pricing', '/docs', '/faq', '/about', '/blog', '/terms', '/privacy', '/refund']
const blogPostPaths = posts.map((post) => `/blog/${post.slug}`)
const targetPaths = [...staticPaths, ...blogPostPaths]

function outputPathFor(routePath: string): string {
  if (routePath === '/') return join(distDir, 'index.html')
  return join(distDir, routePath.slice(1), 'index.html')
}

async function renderRoute(path: string): Promise<string> {
  // Fresh pinia/router/head per route — no state should leak between
  // renders. A dedicated createMemoryHistory() router (built from the same
  // `routes` array the real app uses, but never installed with the
  // beforeEach/afterEach guards in src/router/index.ts) means auth.init()
  // never fires — there's no backend running during this build step.
  const app = createApp(App)
  app.use(createPinia())

  const router = createRouter({ history: createMemoryHistory(), routes })
  app.use(router)
  await router.push(path)
  await router.isReady()

  const head = createHead()
  app.use(head)

  const bodyHtml = await renderToString(app)
  const withBody = template.replace('<div id="app"></div>', `<div id="app">${bodyHtml}</div>`)
  return transformHtmlTemplate(head, withBody)
}

async function main() {
  for (const path of targetPaths) {
    const html = await renderRoute(path)
    const outPath = outputPathFor(path)
    mkdirSync(dirname(outPath), { recursive: true })
    writeFileSync(outPath, html)
  }
  console.log(`Prerendered ${targetPaths.length} routes`)
}

await main()
