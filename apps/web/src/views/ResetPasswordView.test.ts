import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ResetPasswordView from './ResetPasswordView.vue'

const pushMock = vi.fn()
const { routeQuery } = vi.hoisted(() => ({
  routeQuery: { token: 'valid-token' } as { token?: string },
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {}, query: routeQuery }),
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

beforeEach(() => {
  routeQuery.token = 'valid-token'
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('ResetPasswordView', () => {
  it('renders the reset-password form when a token is present', async () => {
    const wrapper = mount(ResetPasswordView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Set new password')
    expect(wrapper.find('#password').exists()).toBe(true)
    expect(wrapper.find('#password').attributes('disabled')).toBeFalsy()
    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeFalsy()
  })

  it('shows an error and disables the form when the token is missing', async () => {
    routeQuery.token = undefined
    const wrapper = mount(ResetPasswordView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Invalid or missing reset link.')
    expect(wrapper.find('#password').attributes('disabled')).toBeDefined()
    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('shows a validation error when the password is shorter than 8 characters', async () => {
    const wrapper = mount(ResetPasswordView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    await flushPromises()

    await wrapper.find('#password').setValue('short')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Password must be at least 8 characters.')
    expect(postMock).not.toHaveBeenCalled()
  })

  it('submits the token and password, then navigates to sign-in on success', async () => {
    postMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(ResetPasswordView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    await flushPromises()

    await wrapper.find('#password').setValue('newpassword123')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/auth/reset-password', {
      token: 'valid-token',
      password: 'newpassword123',
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'sign-in' })
  })

  it('shows an expired-link message on a 400 error', async () => {
    postMock.mockRejectedValueOnce(new ApiError(400, 'bad request'))
    const wrapper = mount(ResetPasswordView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    await flushPromises()

    await wrapper.find('#password').setValue('newpassword123')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('This reset link is invalid or has expired.')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic error message on a non-400 failure', async () => {
    postMock.mockRejectedValueOnce(new Error('network down'))
    const wrapper = mount(ResetPasswordView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    await flushPromises()

    await wrapper.find('#password').setValue('newpassword123')
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
    const wrapper = mount(ResetPasswordView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    await flushPromises()

    await wrapper.find('#password').setValue('newpassword123')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.find('button[type="submit"]').text()).toBe('Saving…')

    resolvePost(undefined)
    await flushPromises()
  })
})
