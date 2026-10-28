// A deploy replaces the container, and with it every built chunk: a tab
// opened before the deploy still asks for the old hashed files, which now
// 404, so a lazy route fails to load. Vite fires vite:preloadError when a
// dynamic import fails; reloading fetches the new index.html and its chunks.
// The timestamp stops a reload loop when a chunk is missing for another
// reason (an outage, a broken build).
const STORAGE_KEY = 'stale_chunk_reload_at'
const MIN_INTERVAL_MS = 10_000

export function reloadOnStaleChunk(win: Window = window) {
  win.addEventListener('vite:preloadError', (event) => {
    let last = 0
    try {
      last = Number(win.sessionStorage.getItem(STORAGE_KEY)) || 0
    } catch {
      // Storage blocked: reload anyway, the interval guard is best effort.
    }
    if (Date.now() - last < MIN_INTERVAL_MS) return
    try {
      win.sessionStorage.setItem(STORAGE_KEY, String(Date.now()))
    } catch {
      // See above.
    }
    event.preventDefault()
    win.location.reload()
  })
}
