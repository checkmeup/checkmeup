import { beforeEach, describe, expect, it, vi } from 'vitest'
import { reloadOnStaleChunk } from './staleChunk'

function fakeWindow() {
  const target = new EventTarget()
  const reload = vi.fn()
  const win = {
    addEventListener: target.addEventListener.bind(target),
    sessionStorage,
    location: { reload },
  } as unknown as Window
  reloadOnStaleChunk(win)
  const fire = () => {
    const event = new Event('vite:preloadError', { cancelable: true })
    target.dispatchEvent(event)
    return event
  }
  return { reload, fire }
}

describe('reloadOnStaleChunk', () => {
  beforeEach(() => {
    sessionStorage.clear()
    vi.useRealTimers()
  })

  it('reloads the page when a chunk fails to load, and stops Vite from throwing', () => {
    const { reload, fire } = fakeWindow()
    const event = fire()
    expect(reload).toHaveBeenCalledOnce()
    expect(event.defaultPrevented).toBe(true)
  })

  it('does not reload again within 10 seconds, so a missing chunk cannot loop', () => {
    vi.useFakeTimers()
    const { reload, fire } = fakeWindow()
    fire()
    vi.advanceTimersByTime(9_000)
    const second = fire()
    expect(reload).toHaveBeenCalledOnce()
    expect(second.defaultPrevented).toBe(false)
    vi.advanceTimersByTime(2_000)
    fire()
    expect(reload).toHaveBeenCalledTimes(2)
  })
})
