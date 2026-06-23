import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import SignUpView from './SignUpView.vue'

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

async function fillForm(
  wrapper: ReturnType<typeof mount>,
  email: string,
  password: string,
  accept = true,
) {
  await wrapper.find('#email').setValue(email)
  await wrapper.find('#password').setValue(password)
  if (accept) {
    await wrapper.find('input[type="checkbox"]').setValue(true)
  }
}

beforeEach(() => {
  authStoreMock.user = null
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('SignUpView', () => {
  it('renders the sign-up form', () => {
    const wrapper = mount(SignUpView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Create account')
    expect(wrapper.find('#email').exists()).toBe(true)
    expect(wrapper.find('#password').exists()).toBe(true)
    expect(wrapper.find('input[type="checkbox"]').exists()).toBe(true)
  })

  it('disables the submit button until terms are accepted', async () => {
    const wrapper = mount(SignUpView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    const button = wrapper.find('button[type="submit"]')
    expect(button.attributes('disabled')).toBeDefined()

    await wrapper.find('input[type="checkbox"]').setValue(true)
    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeUndefined()
  })

  it('shows a validation error when the password is shorter than 8 characters', async () => {
    const wrapper = mount(SignUpView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await fillForm(wrapper, 'a@b.com', 'short')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Password must be at least 8 characters.')
    expect(postMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when terms are not accepted', async () => {
    const wrapper = mount(SignUpView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await fillForm(wrapper, 'a@b.com', 'password123', false)
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('You must accept the Terms of Service and Privacy Policy.')
    expect(postMock).not.toHaveBeenCalled()
  })

  it('submits the form, sets the user, and navigates to the dashboard on success', async () => {
    const user = {
      id: 'u1',
      email: 'a@b.com',
      orgId: 'o1',
      termsVersion: '1',
      termsAcceptedAt: '2026-06-01',
      needsTermsAcceptance: false,
    }
    postMock.mockResolvedValueOnce(user)
    const wrapper = mount(SignUpView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await fillForm(wrapper, 'a@b.com', 'password123')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/auth/sign-up', {
      email: 'a@b.com',
      password: 'password123',
      acceptedTerms: true,
    })
    expect(authStoreMock.setUser).toHaveBeenCalledExactlyOnceWith(user)
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'dashboard' })
  })

  it('shows an account-exists message on a 409 error', async () => {
    postMock.mockRejectedValueOnce(new ApiError(409, 'conflict'))
    const wrapper = mount(SignUpView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await fillForm(wrapper, 'a@b.com', 'password123')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('An account with this email already exists.')
    expect(authStoreMock.setUser).not.toHaveBeenCalled()
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic error message on a non-409 failure', async () => {
    postMock.mockRejectedValueOnce(new Error('network down'))
    const wrapper = mount(SignUpView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await fillForm(wrapper, 'a@b.com', 'password123')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Something went wrong. Please try again.')
  })

  it('shows a loading label while submitting', async () => {
    let resolvePost: (value: unknown) => void = () => {}
    postMock.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolvePost = resolve
        }),
    )
    const wrapper = mount(SignUpView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await fillForm(wrapper, 'a@b.com', 'password123')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.find('button[type="submit"]').text()).toBe('Creating account…')

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
