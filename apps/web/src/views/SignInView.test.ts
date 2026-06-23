import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import SignInView from './SignInView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {}, query: {} }),
}))

vi.mock('@/layouts/AuthLayout.vue', () => ({
  default: { name: 'AuthLayout', template: '<div><slot /></div>' },
}))

const { postMock, ApiError } = vi.hoisted(() => {
  class ApiError extends Error {
    status: number
    code: string
    constructor(status: number, message: string, code = '') {
      super(message)
      this.name = 'ApiError'
      this.status = status
      this.code = code
    }
  }
  return { postMock: vi.fn(), ApiError }
})

vi.mock('@/api/client', () => ({
  api: { get: vi.fn(), post: postMock, put: vi.fn(), patch: vi.fn(), delete: vi.fn() },
  ApiError,
}))

const authStoreMock = {
  user: null as unknown,
  setUser: vi.fn(),
  clear: vi.fn(),
}

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreMock,
}))

async function fillAndSubmit(wrapper: ReturnType<typeof mount>, email: string, password: string) {
  await wrapper.find('#email').setValue(email)
  await wrapper.find('#password').setValue(password)
  await wrapper.find('form').trigger('submit.prevent')
  await flushPromises()
}

beforeEach(() => {
  authStoreMock.user = null
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('SignInView', () => {
  it('renders the sign-in form', () => {
    const wrapper = mount(SignInView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Sign in')
    expect(wrapper.find('#email').exists()).toBe(true)
    expect(wrapper.find('#password').exists()).toBe(true)
  })

  it('submits credentials, sets the user, and navigates to the dashboard on success', async () => {
    const user = {
      id: 'u1',
      email: 'a@b.com',
      orgId: 'o1',
      termsVersion: null,
      termsAcceptedAt: null,
      needsTermsAcceptance: false,
    }
    postMock.mockResolvedValueOnce(user)
    const wrapper = mount(SignInView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await fillAndSubmit(wrapper, 'a@b.com', 'password123')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/auth/sign-in', {
      email: 'a@b.com',
      password: 'password123',
    })
    expect(authStoreMock.setUser).toHaveBeenCalledExactlyOnceWith(user)
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'dashboard' })
  })

  it('shows an incorrect credentials message on a 401 error', async () => {
    postMock.mockRejectedValueOnce(new ApiError(401, 'unauthorized'))
    const wrapper = mount(SignInView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await fillAndSubmit(wrapper, 'a@b.com', 'wrongpass')

    expect(wrapper.text()).toContain('Incorrect email or password.')
    expect(authStoreMock.setUser).not.toHaveBeenCalled()
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic error message on a non-ApiError failure', async () => {
    postMock.mockRejectedValueOnce(new Error('network down'))
    const wrapper = mount(SignInView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await fillAndSubmit(wrapper, 'a@b.com', 'password123')

    expect(wrapper.text()).toContain('Something went wrong. Please try again.')
  })

  it('shows a generic error message on a non-401 ApiError', async () => {
    postMock.mockRejectedValueOnce(new ApiError(500, 'server error'))
    const wrapper = mount(SignInView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await fillAndSubmit(wrapper, 'a@b.com', 'password123')

    expect(wrapper.text()).toContain('Something went wrong. Please try again.')
  })

  it('disables the submit button and shows a loading label while submitting', async () => {
    let resolvePost: (value: unknown) => void = () => {}
    postMock.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolvePost = resolve
        }),
    )
    const wrapper = mount(SignInView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#email').setValue('a@b.com')
    await wrapper.find('#password').setValue('password123')
    await wrapper.find('form').trigger('submit.prevent')

    const button = wrapper.find('button[type="submit"]')
    expect(button.text()).toBe('Signing in…')
    expect(button.attributes('disabled')).toBeDefined()

    resolvePost({
      id: 'u1',
      email: 'a@b.com',
      orgId: 'o1',
      termsVersion: null,
      termsAcceptedAt: null,
      needsTermsAcceptance: false,
    })
    await flushPromises()
  })
})
