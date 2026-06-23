import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import AcceptTermsView from './AcceptTermsView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {}, query: {} }),
}))

vi.mock('@/layouts/AuthLayout.vue', () => ({
  default: { name: 'AuthLayout', template: '<div><slot /></div>' },
}))

const { postMock } = vi.hoisted(() => ({
  postMock: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  api: { get: vi.fn(), post: postMock, put: vi.fn(), patch: vi.fn(), delete: vi.fn() },
  ApiError: class ApiError extends Error {},
}))

const authStoreMock = {
  user: null as { termsVersion: string | null } | null,
  setUser: vi.fn(),
  clear: vi.fn(),
}

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreMock,
}))

beforeEach(() => {
  authStoreMock.user = null
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('AcceptTermsView', () => {
  it('shows the first-time copy when the user has no prior terms version', () => {
    authStoreMock.user = { termsVersion: null }
    const wrapper = mount(AcceptTermsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('One more thing')
  })

  it('shows the re-accept copy when the user already has a terms version', () => {
    authStoreMock.user = { termsVersion: '1' }
    const wrapper = mount(AcceptTermsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain("We've updated our Terms and Privacy Policy")
  })

  it('disables the accept button until the checkbox is checked', async () => {
    const wrapper = mount(AcceptTermsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    const acceptButton = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Accept and continue'))

    expect(acceptButton!.attributes('disabled')).toBeDefined()

    await wrapper.find('input[type="checkbox"]').setValue(true)

    const acceptButtonAfter = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Accept and continue'))
    expect(acceptButtonAfter!.attributes('disabled')).toBeUndefined()
  })

  it('accepts terms, sets the user, and navigates to the dashboard on success', async () => {
    const user = {
      id: 'u1',
      email: 'a@b.com',
      orgId: 'o1',
      termsVersion: '2',
      termsAcceptedAt: '2026-06-23',
      needsTermsAcceptance: false,
    }
    postMock.mockResolvedValueOnce(user)
    const wrapper = mount(AcceptTermsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('input[type="checkbox"]').setValue(true)
    await wrapper
      .findAll('button')
      .find((b) => b.text().includes('Accept and continue'))!
      .trigger('click')
    await flushPromises()

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/auth/accept-terms', {})
    expect(authStoreMock.setUser).toHaveBeenCalledExactlyOnceWith(user)
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'dashboard' })
  })

  it('does not submit when the checkbox is unchecked', async () => {
    const wrapper = mount(AcceptTermsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper
      .findAll('button')
      .find((b) => b.text().includes('Accept and continue'))!
      .trigger('click')
    await flushPromises()

    expect(postMock).not.toHaveBeenCalled()
  })

  it('shows an error message when accepting terms fails', async () => {
    postMock.mockRejectedValueOnce(new Error('network down'))
    const wrapper = mount(AcceptTermsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('input[type="checkbox"]').setValue(true)
    await wrapper
      .findAll('button')
      .find((b) => b.text().includes('Accept and continue'))!
      .trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Something went wrong. Please try again.')
    expect(authStoreMock.setUser).not.toHaveBeenCalled()
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a loading label while accepting', async () => {
    let resolvePost: (value: unknown) => void = () => {}
    postMock.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolvePost = resolve
        }),
    )
    const wrapper = mount(AcceptTermsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('input[type="checkbox"]').setValue(true)
    await wrapper
      .findAll('button')
      .find((b) => b.text().includes('Accept and continue'))!
      .trigger('click')

    const acceptButton = wrapper.findAll('button').find((b) => b.text().includes('Saving…'))
    expect(acceptButton).toBeTruthy()

    resolvePost({
      id: 'u1',
      email: 'a@b.com',
      orgId: 'o1',
      termsVersion: '2',
      termsAcceptedAt: '2026-06-23',
      needsTermsAcceptance: false,
    })
    await flushPromises()
  })

  it('signs out and navigates home when "Sign out instead" is clicked', async () => {
    postMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(AcceptTermsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Sign out instead')!
      .trigger('click')
    await flushPromises()

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/auth/sign-out', {})
    expect(authStoreMock.clear).toHaveBeenCalledOnce()
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'home' })
  })

  it('still clears auth and navigates home when sign-out request fails', async () => {
    const onUnhandledRejection = () => {}
    process.on('unhandledRejection', onUnhandledRejection)
    postMock.mockRejectedValueOnce(new Error('network down'))
    const wrapper = mount(AcceptTermsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Sign out instead')!
      .trigger('click')
    await flushPromises()

    expect(authStoreMock.clear).toHaveBeenCalledOnce()
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'home' })
    process.off('unhandledRejection', onUnhandledRejection)
  })
})
