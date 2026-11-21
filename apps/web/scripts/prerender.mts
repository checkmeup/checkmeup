// Runs after `vite build` (via `vite-node`, so it can import real .vue/.ts
// source through the same Vite pipeline/aliases as the app itself — a plain
// `node` script can't parse SFCs or strip TypeScript). Renders each public,
// indexable route to static HTML so crawlers that don't execute JS (Bingbot
// flagged the homepage's missing <h1> — see ADR seo) see real content
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
import { Window } from 'happy-dom'
import App from '../src/App.vue'
import { routes } from '../src/router/routes'
import { postsMeta } from '../src/blog/posts'
import { DEFAULT_OG_IMAGE } from '../src/composables/useSeo'
import {
  SOCIAL_IMAGE_HEIGHT,
  SOCIAL_IMAGE_WIDTH,
  renderSocialImage,
  socialImageLabel,
} from './social-image.mts'

const __dirname = dirname(fileURLToPath(import.meta.url))
const distDir = join(__dirname, '../dist')
const template = readFileSync(join(distDir, 'index.html'), 'utf-8')

// Keep in sync with scripts/generate-sitemap.mjs's staticPaths — that script
// derives the same list independently (by filesystem-scanning src/blog/posts
// rather than importing the typed module) for the unrelated job of writing
// sitemap.xml; both lists should keep matching the indexable routes below.
const staticPaths = ['/', '/pricing', '/docs', '/faq', '/about', '/blog', '/changelog', '/terms', '/privacy', '/refund', '/tools/ssl-checker', '/tools/domain-expiry-checker', '/tools/dns-lookup', '/tools/cron-expression', '/cron-monitoring', '/uptime-monitoring', '/ssl-monitoring', '/domain-expiry-monitoring', '/port-monitoring', '/dns-monitoring', '/alternatives/healthchecks-io', '/alternatives/uptimerobot', '/alternatives/cronitor']
// Every post, release notes included: a release post is noindex but still
// prerendered, since the robots tag has to be in the HTML the server sends —
// a noindex added only by client JS after load can be missed by crawlers.
// (Its absence from sitemap.xml is the one place the two lists differ.)
const blogPostPaths = postsMeta.map((meta) => `/blog/${meta.slug}`)
const targetPaths = [...staticPaths, ...blogPostPaths]

function outputPathFor(routePath: string): string {
  if (routePath === '/') return join(distDir, 'index.html')
  return join(distDir, routePath.slice(1), 'index.html')
}

const NOT_FOUND_PATH = '/404'

// Parses like a client that runs no JavaScript and fetches nothing.
const domWindow = new Window({
  settings: {
    disableJavaScriptEvaluation: true,
    disableJavaScriptFileLoading: true,
    disableCSSFileLoading: true,
    disableIframePageLoading: true,
  },
})

// The text of a rendered page's first <h1>, as a reader sees it.
function h1FromHtml(bodyHtml: string): string | null {
  // This build's own server-rendered markup, parsed only to read its text.
  // eslint-disable-next-line xss/no-mixed-html
  const doc = new domWindow.DOMParser().parseFromString(bodyHtml, 'text/html')
  return doc.querySelector('h1')?.textContent?.replace(/\s+/g, ' ').trim() || null
}

// Writes the page's own Open Graph image (PRD social-images) and returns the
// head tags that describe it. `/` keeps its hand-made card and the 404 page
// gets it too; so does a page whose image can't be rendered, logged here and
// caught in CI by prerender.test.ts rather than failing the deploy.
async function socialImageHead(path: string, title: string | null) {
  const homepageCard = {
    meta: [
      { property: 'og:image', content: DEFAULT_OG_IMAGE },
      { name: 'twitter:image', content: DEFAULT_OG_IMAGE },
    ],
  }
  if (path === '/' || path === NOT_FOUND_PATH) return homepageCard
  const fallBack = (reason: string) => {
    console.warn(`No social image for ${path}, keeping the homepage card: ${reason}`)
    return homepageCard
  }
  if (!title) return fallBack('no <h1>')
  try {
    const png = await renderSocialImage(title, socialImageLabel(path))
    if (!png) return fallBack(`"${title}" doesn't fit 3 lines`)
    const outPath = join(distDir, 'img/og', `${path.slice(1)}.png`)
    mkdirSync(dirname(outPath), { recursive: true })
    // Under dist/, named after one of this script's own routes, not input.
    // eslint-disable-next-line security/detect-non-literal-fs-filename
    writeFileSync(outPath, png)
    return {
      meta: [
        { property: 'og:image:width', content: String(SOCIAL_IMAGE_WIDTH) },
        { property: 'og:image:height', content: String(SOCIAL_IMAGE_HEIGHT) },
        { property: 'og:image:alt', content: title },
        { name: 'twitter:image:alt', content: title },
      ],
    }
  } catch (err) {
    return fallBack((err as Error).message)
  }
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
  // eslint-disable-next-line xss/no-mixed-html -- the page's own markup, read for its <h1> text
  head.push(await socialImageHead(path, path === '/' ? null : h1FromHtml(bodyHtml)))
  const withBody = template.replace('<div id="app"></div>', `<div id="app">${bodyHtml}</div>`)
  return transformHtmlTemplate(head, withBody)
}

// The routes the Go server answers with the app and a 200; any other path
// gets 404.html with a 404, not the app shell with a 200 (ADR seo). The
// catch-all is left out, and so is /blog/:slug: every real post has its own
// prerendered file, so a slug without one doesn't exist.
const spaRoutePaths = routes
  .map((route) => route.path)
  .filter((path) => path !== '/:pathMatch(.*)*' && path !== '/blog/:slug')

async function main() {
  for (const path of targetPaths) {
    const html = await renderRoute(path)
    const outPath = outputPathFor(path)
    mkdirSync(dirname(outPath), { recursive: true })
    writeFileSync(outPath, html)
  }
  // Any path no route matches renders the catch-all's NotFoundView.
  writeFileSync(join(distDir, '404.html'), await renderRoute(NOT_FOUND_PATH))
  writeFileSync(join(distDir, 'spa-routes.json'), JSON.stringify(spaRoutePaths))
  console.log(`Prerendered ${targetPaths.length} routes and 404.html`)
}

await main()
