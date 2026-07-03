// Set the theme before first paint to avoid a flash of the wrong theme.
// Mirrors the logic in src/lib/theme.ts — keep both in sync if this changes.
// Externalized (rather than inline in index.html) so the CSP script-src can
// stay 'self'-only without an 'unsafe-inline' carve-out.
;(function () {
  var stored = localStorage.getItem('theme')
  var theme = stored || (matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark')
  document.documentElement.dataset.theme = theme
})()
