// Public status page theme handling. Shares the 'theme' localStorage key
// with src/lib/theme.ts so a visitor's preference carries over across the
// whole site (same origin — ADR-005's same-domain /status/:slug). Runs
// before paint (script is in <head>, unhashed, same as /theme-init.js) to
// avoid a flash of the wrong theme, then wires up the toggle button once
// the DOM is ready.
;(function () {
  var STORAGE_KEY = 'theme'

  function systemPreference() {
    return matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark'
  }

  var stored = localStorage.getItem(STORAGE_KEY)
  document.documentElement.dataset.theme = stored || systemPreference()

  document.addEventListener('DOMContentLoaded', function () {
    var btn = document.getElementById('theme-toggle')
    if (!btn) return

    function render() {
      btn.textContent = document.documentElement.dataset.theme === 'dark' ? '☀' : '☾'
    }

    render()
    btn.addEventListener('click', function () {
      var next = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark'
      document.documentElement.dataset.theme = next
      localStorage.setItem(STORAGE_KEY, next)
      render()
    })
  })
})()
