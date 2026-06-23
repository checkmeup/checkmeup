import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import DomainMonitorEditView from './DomainMonitorEditView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'd1' } }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { updateDomainMock } = vi.hoisted(() => ({
  updateDomainMock: vi.fn(),
}))

vi.mock('@/api/monitors', () => ({
  monitorsApi: { updateDomain: updateDomainMock },
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

const detailData = ref<unknown>(null)
const pending = ref(false)
const loadError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useDomainMonitors', () => ({
  useDomainMonitor: () => ({
    data: detailData,
    isPending: pending,
    error: loadError,
  }),
}))

const channelsData = ref<{ id: string; enabled: boolean }[]>([])
const channelsPending = ref(false)

vi.mock('@/composables/useNotificationChannels', () => ({
  useNotificationChannels: () => ({ data: channelsData, isPending: channelsPending }),
}))

const monitor = {
  id: 'd1',
  name: 'Production domain',
  domain: 'example.com',
  status: 'up' as const,
  alertsEnabled: true,
  expiresAt: '2027-01-01T00:00:00Z',
  registrar: 'Namecheap',
  errorMsg: null,
  daysUntilExpiry: 190,
  lastCheckedAt: '2026-06-20T00:00:00Z',
  createdAt: '2026-01-01T00:00:00Z',
  channelIds: ['ch1'],
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  detailData.value = null
  pending.value = false
  loadError.value = null
  channelsData.value = []
  channelsPending.value = false
  updateDomainMock.mockReset()
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('DomainMonitorEditView', () => {
  it('shows a loading state while the monitor is pending', () => {
    pending.value = true
    const wrapper = mount(DomainMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an inline error message when loading the monitor fails', async () => {
    const wrapper = mount(DomainMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    loadError.value = { message: 'Monitor not found' }
    await flushPromises()

    expect(wrapper.text()).toContain('Monitor not found')
    expect(wrapper.find('form').exists()).toBe(true)
  })

  it('prefills the form fields from the loaded monitor', () => {
    detailData.value = { ...monitor }
    const wrapper = mount(DomainMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect((wrapper.find('#name').element as HTMLInputElement).value).toBe('Production domain')
    expect(wrapper.text()).toContain('example.com')
    expect((wrapper.find('#alerts').element as HTMLInputElement).checked).toBe(true)
  })

  it('shows the domain as read-only with a note about recreating to change it', () => {
    detailData.value = { ...monitor }
    const wrapper = mount(DomainMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.find('#domain').exists()).toBe(false)
    expect(wrapper.text()).toContain(
      'To change the domain, delete this monitor and create a new one.',
    )
  })

  it('shows a validation error when name is cleared', async () => {
    detailData.value = { ...monitor }
    const wrapper = mount(DomainMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('  ')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Name is required')
    expect(updateDomainMock).not.toHaveBeenCalled()
  })

  it('updates the monitor and navigates to the detail view on success', async () => {
    detailData.value = { ...monitor }
    updateDomainMock.mockResolvedValueOnce({ ...monitor, name: 'Renamed domain' })
    const wrapper = mount(DomainMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Renamed domain')
    await wrapper.find('#alerts').setValue(false)
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(updateDomainMock).toHaveBeenCalledExactlyOnceWith('d1', {
      name: 'Renamed domain',
      domain: 'example.com',
      alertsEnabled: false,
      channelIds: ['ch1'],
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'domain-monitor-detail',
      params: { id: 'd1' },
    })
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    detailData.value = { ...monitor }
    updateDomainMock.mockRejectedValueOnce(
      new ApiErrorMock(402, 'Upgrade to add more monitors', 'plan_limit_reached'),
    )
    const wrapper = mount(DomainMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to add more monitors')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic error message when updating fails for another reason', async () => {
    detailData.value = { ...monitor }
    updateDomainMock.mockRejectedValueOnce(new Error('Network error'))
    const wrapper = mount(DomainMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Network error')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the detail view when back or cancel is clicked', async () => {
    detailData.value = { ...monitor }
    const wrapper = mount(DomainMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, '← Back')!.trigger('click')
    expect(pushMock).toHaveBeenCalledWith({
      name: 'domain-monitor-detail',
      params: { id: 'd1' },
    })

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')
    expect(pushMock).toHaveBeenCalledWith({
      name: 'domain-monitor-detail',
      params: { id: 'd1' },
    })
  })
})
