import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore } from './auth'

function jsonResponse(ok: boolean, body: unknown): Response {
  return { ok, json: () => Promise.resolve(body) } as Response
}

describe('auth store', () => {
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    setActivePinia(createPinia())
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('starts idle and unauthenticated', () => {
    const auth = useAuthStore()

    expect(auth.status).toBe('idle')
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
  })

  it('init() authenticates and stores the user on a successful /me response', async () => {
    const user = {
      id: 'user-1',
      email: 'test@example.com',
      orgId: 'org-1',
      termsVersion: '2026-01-01',
      termsAcceptedAt: '2026-01-01T00:00:00Z',
      needsTermsAcceptance: false,
    }
    fetchMock.mockResolvedValueOnce(jsonResponse(true, user))
    const auth = useAuthStore()

    await auth.init()

    expect(auth.isAuthenticated).toBe(true)
    expect(auth.user).toEqual(user)
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/me', { credentials: 'include' })
  })

  it('init() ends up unauthenticated on a non-ok /me response, without throwing', async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(false, {}))
    const auth = useAuthStore()

    await auth.init()

    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
  })

  it('init() ends up unauthenticated when fetch itself rejects (network error)', async () => {
    fetchMock.mockRejectedValueOnce(new Error('network down'))
    const auth = useAuthStore()

    await auth.init()

    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
  })

  it('init() only calls fetch once — a second call is a no-op once no longer idle', async () => {
    fetchMock.mockResolvedValue(jsonResponse(true, { id: 'user-1' }))
    const auth = useAuthStore()

    await auth.init()
    await auth.init()

    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('setUser() authenticates immediately without calling fetch', () => {
    const user = {
      id: 'user-2',
      email: 'set@example.com',
      orgId: 'org-2',
      termsVersion: null,
      termsAcceptedAt: null,
      needsTermsAcceptance: true,
    }
    const auth = useAuthStore()

    auth.setUser(user)

    expect(auth.isAuthenticated).toBe(true)
    expect(auth.user).toEqual(user)
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('clear() logs out an authenticated user', () => {
    const auth = useAuthStore()
    auth.setUser({
      id: 'user-3',
      email: 'clear@example.com',
      orgId: 'org-3',
      termsVersion: '2026-01-01',
      termsAcceptedAt: '2026-01-01T00:00:00Z',
      needsTermsAcceptance: false,
    })

    auth.clear()

    expect(auth.isAuthenticated).toBe(false)
    expect(auth.user).toBeNull()
    expect(auth.status).toBe('unauthenticated')
  })
})
