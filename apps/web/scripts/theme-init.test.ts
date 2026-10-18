// public/theme-init.js, loaded by index.html, sets the theme before the
// first paint. It's run here as the browser runs it: a classic script
// against the page's globals.
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const webDir = join(dirname(fileURLToPath(import.meta.url)), '..')
const indexHtml = readFileSync(join(webDir, 'index.html'), 'utf-8')
const themeInit = readFileSync(join(webDir, 'public/theme-init.js'), 'utf-8')

function boot(osPrefersLight: boolean) {
  vi.stubGlobal('matchMedia', (query: string) => ({
    matches: query === '(prefers-color-scheme: light)' ? osPrefersLight : !osPrefersLight,
  }))
  // eslint-disable-next-line no-new-func -- runs the served script as the browser does
  new Function(themeInit)()
  return document.documentElement.dataset.theme
}

beforeEach(() => {
  localStorage.clear()
  delete document.documentElement.dataset.theme
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('theme-init.js', () => {
  it('runs as a blocking script in <head>, before the body is painted', () => {
    // Read as text: parsing it into a document would fetch its stylesheets.
    const head = /<head>([\s\S]*)<\/head>/.exec(indexHtml)?.[1] ?? ''
    const tag = /<script\b([^>]*)>\s*<\/script>/g
    const attributes = [...head.matchAll(tag)]
      .map((m) => m[1])
      .find((attrs) => attrs.includes('src="/theme-init.js"'))

    expect(attributes).toBeDefined()
    expect(attributes).not.toMatch(/\b(async|defer|type)\b/)
  })

  it("gives a returning visitor their saved theme, whatever the OS's preference", () => {
    localStorage.setItem('theme', 'light')
    expect(boot(false)).toBe('light')

    localStorage.setItem('theme', 'dark')
    expect(boot(true)).toBe('dark')
  })

  it("gives a first-time visitor the theme matching the OS's prefers-color-scheme", () => {
    expect(boot(true)).toBe('light')
    expect(boot(false)).toBe('dark')
  })
})
