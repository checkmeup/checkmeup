declare global {
  interface Window {
    dataLayer?: unknown[]
  }
}

let loaded = false

// Injects the GTM container script — only call this after consent is granted.
// Skips silently (same posture as the missing-Paddle-token case) if the
// container ID isn't configured, e.g. in local dev.
export function loadGtm() {
  if (loaded) return
  const id = import.meta.env.VITE_GTM_ID
  if (!id) return
  loaded = true

  window.dataLayer = window.dataLayer ?? []
  window.dataLayer.push({ 'gtm.start': Date.now(), event: 'gtm.js' })

  const script = document.createElement('script')
  script.async = true
  script.src = `https://www.googletagmanager.com/gtm.js?id=${id}`
  document.head.appendChild(script)
}

// GTM's own pageview trigger only fires on the initial hard load; SPA route
// changes need an explicit push so GA4 sees them as pageviews too.
export function trackPageview(path: string) {
  if (!loaded) return
  window.dataLayer?.push({ event: 'page_view', page_path: path })
}
