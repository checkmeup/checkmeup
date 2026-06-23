import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ForgotPasswordView from './ForgotPasswordView.vue'

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

beforeEach(() => {
  postMock.mockReset()
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('ForgotPasswordView', () => {
  it('renders the forgot-password form', () => {
    const wrapper = mount(ForgotPasswordView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Reset password')
    expect(wrapper.find('#email').exists()).toBe(true)
  })

  it('submits the email and shows a confirmation message on success', async () => {
    postMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(ForgotPasswordView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#email').setValue('a@b.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/auth/forgot-password', {
      email: 'a@b.com',
    })
    expect(wrapper.text()).toContain("you'll receive a reset link shortly")
    expect(wrapper.text()).toContain('a@b.com')
    expect(wrapper.find('#email').exists()).toBe(false)
  })

  it('shows the same confirmation message even when the request fails', async () => {
    const onUnhandledRejection = () => {}
    process.on('unhandledRejection', onUnhandledRejection)
    postMock.mockRejectedValueOnce(new Error('network down'))
    const wrapper = mount(ForgotPasswordView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#email').setValue('a@b.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain("you'll receive a reset link shortly")
    process.off('unhandledRejection', onUnhandledRejection)
  })

  it('shows a loading label while submitting', async () => {
    let resolvePost: (value: unknown) => void = () => {}
    postMock.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolvePost = resolve
        }),
    )
    const wrapper = mount(ForgotPasswordView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#email').setValue('a@b.com')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.find('button[type="submit"]').text()).toBe('Sending…')

    resolvePost(undefined)
    await flushPromises()
  })
})
