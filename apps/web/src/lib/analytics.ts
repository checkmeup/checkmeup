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

// A completed sign-up, so GA4 can count it as a key event. GTM forwards it to
// GA4 through a tag on the `sign_up` custom event. Email is the only sign-up
// method there is.
export function trackSignUp() {
  if (!loaded) return
  window.dataLayer?.push({ event: 'sign_up', method: 'email' })
}

// A run of a free checker (PRD free-tools), so GA4 can count tool use beside
// the sign-ups that follow it.
export function trackToolCheck(tool: string, ok: boolean) {
  if (!loaded) return
  window.dataLayer?.push({ event: 'tool_check', tool, outcome: ok ? 'ok' : 'error' })
}
