import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import SSLMonitorCreateView from './SSLMonitorCreateView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {} }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { createSSLMock } = vi.hoisted(() => ({
  createSSLMock: vi.fn(),
}))

vi.mock('@/api/monitors', () => ({
  monitorsApi: { createSSL: createSSLMock },
}))

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

afterEach(() => {
  vi.clearAllMocks()
})

describe('SSLMonitorCreateView', () => {
  it('renders the form with default values', () => {
    const wrapper = mount(SSLMonitorCreateView)

    expect(wrapper.text()).toContain('New SSL monitor')
    expect(wrapper.find('#name').exists()).toBe(true)
    expect(wrapper.find('#hostname').exists()).toBe(true)
  })

  it('shows a validation error when name is missing', async () => {
    const wrapper = mount(SSLMonitorCreateView)

    await wrapper.find('#hostname').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Name is required')
    expect(createSSLMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when hostname is missing', async () => {
    const wrapper = mount(SSLMonitorCreateView)

    await wrapper.find('#name').setValue('Production API')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Hostname is required')
    expect(createSSLMock).not.toHaveBeenCalled()
  })

  it('creates the monitor and navigates to its detail page on success', async () => {
    createSSLMock.mockResolvedValueOnce({ id: 's1' })
    const wrapper = mount(SSLMonitorCreateView)

    await wrapper.find('#name').setValue('Production API')
    await wrapper.find('#hostname').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createSSLMock).toHaveBeenCalledExactlyOnceWith({
      name: 'Production API',
      hostname: 'example.com',
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'ssl-monitor-detail',
      params: { id: 's1' },
    })
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    const { ApiError } = await import('@/api/client')
    createSSLMock.mockRejectedValueOnce(
      new ApiError(402, 'Upgrade to add more monitors', 'plan_limit_reached'),
    )
    const wrapper = mount(SSLMonitorCreateView)

    await wrapper.find('#name').setValue('Production API')
    await wrapper.find('#hostname').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to add more monitors')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic inline error when creation fails for another reason', async () => {
    createSSLMock.mockRejectedValueOnce(new Error('Network error'))
    const wrapper = mount(SSLMonitorCreateView)

    await wrapper.find('#name').setValue('Production API')
    await wrapper.find('#hostname').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Network error')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the list when cancel is clicked', async () => {
    const wrapper = mount(SSLMonitorCreateView)

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'ssl-monitors' })
  })

  it('navigates back to the list when the back link is clicked', async () => {
    const wrapper = mount(SSLMonitorCreateView)

    await findButtonByText(wrapper, '← Back')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'ssl-monitors' })
  })
})
