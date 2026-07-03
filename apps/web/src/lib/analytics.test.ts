import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

describe('analytics', () => {
  beforeEach(() => {
    vi.resetModules()
    document.head.innerHTML = ''
    window.dataLayer = undefined
  })

  afterEach(() => {
    vi.unstubAllEnvs()
  })

  async function loadAnalytics() {
    return import('./analytics')
  }

  it('loadGtm() does nothing when VITE_GTM_ID is unset', async () => {
    vi.stubEnv('VITE_GTM_ID', '')
    const { loadGtm } = await loadAnalytics()
    loadGtm()
    expect(document.head.querySelector('script')).toBeNull()
    expect(window.dataLayer).toBeUndefined()
  })

  it('loadGtm() injects the GTM script and seeds dataLayer when configured', async () => {
    vi.stubEnv('VITE_GTM_ID', 'GTM-TEST123')
    const { loadGtm } = await loadAnalytics()
    loadGtm()

    const script = document.head.querySelector('script')
    expect(script?.src).toBe('https://www.googletagmanager.com/gtm.js?id=GTM-TEST123')
    expect(script?.async).toBe(true)
    expect(window.dataLayer?.[0]).toMatchObject({ event: 'gtm.js' })
  })

  it('loadGtm() only injects the script once even if called repeatedly', async () => {
    vi.stubEnv('VITE_GTM_ID', 'GTM-TEST123')
    const { loadGtm } = await loadAnalytics()
    loadGtm()
    loadGtm()
    expect(document.head.querySelectorAll('script')).toHaveLength(1)
  })

  it('trackPageview() is a no-op before GTM has loaded', async () => {
    vi.stubEnv('VITE_GTM_ID', 'GTM-TEST123')
    const { trackPageview } = await loadAnalytics()
    trackPageview('/pricing')
    expect(window.dataLayer).toBeUndefined()
  })

  it('trackPageview() pushes a page_view event once GTM has loaded', async () => {
    vi.stubEnv('VITE_GTM_ID', 'GTM-TEST123')
    const { loadGtm, trackPageview } = await loadAnalytics()
    loadGtm()
    trackPageview('/pricing')
    expect(window.dataLayer?.at(-1)).toEqual({ event: 'page_view', page_path: '/pricing' })
  })
})
