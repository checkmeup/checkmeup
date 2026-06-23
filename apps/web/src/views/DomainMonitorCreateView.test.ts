import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import DomainMonitorCreateView from './DomainMonitorCreateView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {} }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { createDomainMock } = vi.hoisted(() => ({
  createDomainMock: vi.fn(),
}))

vi.mock('@/api/monitors', () => ({
  monitorsApi: { createDomain: createDomainMock },
}))

const { ApiErrorMock } = vi.hoisted(() => ({
  ApiErrorMock: class extends Error {
    status: number
    code: string
    constructor(status: number, message: string, code = '') {
      super(message)
      this.status = status
      this.code = code
    }
  },
}))

vi.mock('@/api/client', () => ({
  ApiError: ApiErrorMock,
}))

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  createDomainMock.mockReset()
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('DomainMonitorCreateView', () => {
  it('renders the form with empty default field values', () => {
    const wrapper = mount(DomainMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('New domain monitor')
    expect((wrapper.find('#name').element as HTMLInputElement).value).toBe('')
    expect((wrapper.find('#domain').element as HTMLInputElement).value).toBe('')
  })

  it('shows a validation error when name is missing', async () => {
    const wrapper = mount(DomainMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#domain').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Name is required')
    expect(createDomainMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when domain is missing', async () => {
    const wrapper = mount(DomainMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Production domain')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Domain is required')
    expect(createDomainMock).not.toHaveBeenCalled()
  })

  it('creates the monitor and navigates to the detail view on success', async () => {
    createDomainMock.mockResolvedValueOnce({ id: 'd1' })
    const wrapper = mount(DomainMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Production domain')
    await wrapper.find('#domain').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createDomainMock).toHaveBeenCalledExactlyOnceWith({
      name: 'Production domain',
      domain: 'example.com',
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'domain-monitor-detail',
      params: { id: 'd1' },
    })
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    createDomainMock.mockRejectedValueOnce(
      new ApiErrorMock(402, 'Upgrade to add more monitors', 'plan_limit_reached'),
    )
    const wrapper = mount(DomainMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Production domain')
    await wrapper.find('#domain').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to add more monitors')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic error message when creation fails for another reason', async () => {
    createDomainMock.mockRejectedValueOnce(new Error('Network error'))
    const wrapper = mount(DomainMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Production domain')
    await wrapper.find('#domain').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Network error')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the list when back or cancel is clicked', async () => {
    const wrapper = mount(DomainMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, '← Back')!.trigger('click')
    expect(pushMock).toHaveBeenCalledWith({ name: 'domain-monitors' })

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')
    expect(pushMock).toHaveBeenCalledWith({ name: 'domain-monitors' })
  })
})
