import { beforeEach, describe, expect, it, vi } from 'vitest'

describe('useConsent', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.resetModules()
  })

  async function loadConsent() {
    return import('./consent')
  }

  it('starts undefined when nothing is stored', async () => {
    const { useConsent } = await loadConsent()
    expect(useConsent().status.value).toBeUndefined()
  })

  it('grant() sets status to granted and persists it', async () => {
    const { useConsent } = await loadConsent()
    const { status, grant } = useConsent()
    grant()
    expect(status.value).toBe('granted')
    expect(localStorage.getItem('cookie_consent')).toBe('granted')
  })

  it('deny() sets status to denied and persists it', async () => {
    const { useConsent } = await loadConsent()
    const { status, deny } = useConsent()
    deny()
    expect(status.value).toBe('denied')
    expect(localStorage.getItem('cookie_consent')).toBe('denied')
  })

  it('reads a previously stored decision on module load', async () => {
    localStorage.setItem('cookie_consent', 'granted')
    const { useConsent } = await loadConsent()
    expect(useConsent().status.value).toBe('granted')
  })

  it('ignores a garbage stored value', async () => {
    localStorage.setItem('cookie_consent', 'yes-please')
    const { useConsent } = await loadConsent()
    expect(useConsent().status.value).toBeUndefined()
  })

  it('shares state across multiple useConsent() calls (module singleton)', async () => {
    const { useConsent } = await loadConsent()
    const a = useConsent()
    const b = useConsent()
    a.grant()
    expect(b.status.value).toBe('granted')
  })
})
