import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useAuthStore, type User } from '@/stores/auth'

const trackPageviewMock = vi.fn()
vi.mock('@/lib/analytics', () => ({
  trackPageview: (path: string) => trackPageviewMock(path),
}))

const consentStatus = { value: 'denied' as 'granted' | 'denied' }
vi.mock('@/lib/consent', () => ({
  useConsent: () => ({ status: consentStatus }),
}))

// Imported after the mocks above (vi.mock calls are hoisted, but the router
// module itself must still be evaluated afterward) since router/index.ts's
// beforeEach/afterEach guards call trackPageview/useConsent directly.
const { router } = await import('./index')

function terms(
  needsTermsAcceptance: boolean,
): Pick<User, 'termsVersion' | 'termsAcceptedAt' | 'needsTermsAcceptance'> {
  return needsTermsAcceptance
    ? { termsVersion: null, termsAcceptedAt: null, needsTermsAcceptance: true }
    : {
        termsVersion: '2026-01-01',
        termsAcceptedAt: '2026-01-01T00:00:00Z',
        needsTermsAcceptance: false,
      }
}

const testUser = (needsTermsAcceptance = false): User => ({
  id: 'user-1',
  email: 'test@example.com',
  orgId: 'org-1',
  ...terms(needsTermsAcceptance),
})

describe('router guards', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    trackPageviewMock.mockClear()
    consentStatus.value = 'denied'
  })

  it('redirects an unauthenticated user away from a requiresAuth route to sign-in', async () => {
    // clear() sets status to 'unauthenticated' (not 'idle'), so the guard's
    // lazy auth.init() call becomes a no-op instead of hitting the network.
    useAuthStore().clear()

    await router.push('/dashboard')

    expect(router.currentRoute.value.name).toBe('sign-in')
  })

  it('redirects an authenticated user who still needs terms acceptance to accept-terms', async () => {
    useAuthStore().setUser(testUser(true))

    await router.push('/dashboard')

    expect(router.currentRoute.value.name).toBe('accept-terms')
  })

  it('lets an authenticated, terms-accepted user reach a requiresAuth route', async () => {
    useAuthStore().setUser(testUser())

    await router.push('/dashboard')

    expect(router.currentRoute.value.name).toBe('dashboard')
  })

  it('redirects an authenticated user away from a guest-only route to dashboard', async () => {
    useAuthStore().setUser(testUser())

    await router.push('/sign-in')

    expect(router.currentRoute.value.name).toBe('dashboard')
  })

  it('redirects an authenticated user away from home to dashboard', async () => {
    useAuthStore().setUser(testUser())

    await router.push('/')

    expect(router.currentRoute.value.name).toBe('dashboard')
  })

  it('lets an unauthenticated user reach a guest-only route', async () => {
    useAuthStore().clear()

    await router.push('/sign-in')

    expect(router.currentRoute.value.name).toBe('sign-in')
  })

  it('lets an unauthenticated user reach home without redirecting', async () => {
    useAuthStore().clear()

    await router.push('/')

    expect(router.currentRoute.value.name).toBe('home')
  })

  it('tracks a pageview after navigation when consent is granted', async () => {
    consentStatus.value = 'granted'
    useAuthStore().clear()

    await router.push('/sign-in')

    expect(trackPageviewMock).toHaveBeenCalledWith('/sign-in')
  })

  it('does not track a pageview when consent is not granted', async () => {
    useAuthStore().clear()

    await router.push('/sign-in')

    expect(trackPageviewMock).not.toHaveBeenCalled()
  })
})
