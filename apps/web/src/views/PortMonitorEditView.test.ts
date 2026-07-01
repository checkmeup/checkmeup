import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PortMonitorEditView from './PortMonitorEditView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'p1' } }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { updatePortMock } = vi.hoisted(() => ({
  updatePortMock: vi.fn(),
}))

vi.mock('@/api/monitors', async () => {
  const actual = await vi.importActual<typeof import('@/api/monitors')>('@/api/monitors')
  return {
    ...actual,
    monitorsApi: { updatePort: updatePortMock },
  }
})

const detailData = ref<unknown>(null)
const monitorPending = ref(false)
const loadError = ref<{ message: string } | null>(null)

vi.mock('@/composables/usePortMonitors', () => ({
  usePortMonitor: () => ({ data: detailData, isPending: monitorPending, error: loadError }),
}))

const billingData = ref<{ minIntervalMins: number } | null>(null)
const billingPending = ref(false)

vi.mock('@/composables/useBilling', () => ({
  useBilling: () => ({ data: billingData, isPending: billingPending }),
}))

const channelsData = ref<{ id: string; name: string; type: string; enabled: boolean }[]>([])
const channelsPending = ref(false)

vi.mock('@/composables/useNotificationChannels', () => ({
  useNotificationChannels: () => ({ data: channelsData, isPending: channelsPending }),
}))

const monitor = {
  id: 'p1',
  name: 'SMTP',
  host: 'mail.example.com',
  port: 25,
  expectedState: 'open' as const,
  intervalMins: 5,
  status: 'up' as const,
  alertsEnabled: true,
  maxAlertsPerIncident: 3,
  alertAfterNFailures: 0,
  lastCheckedAt: null,
  createdAt: '2026-01-01T00:00:00Z',
  uptime24h: 99.9,
  channelIds: [] as string[],
}

const detail = {
  monitor,
  chartData: [],
  checks: [],
  incidents: [],
  stats: { uptime24h: 99.9, uptime7d: null, uptime30d: null },
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  detailData.value = null
  monitorPending.value = false
  loadError.value = null
  billingData.value = null
  billingPending.value = false
  channelsData.value = []
  channelsPending.value = false
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('PortMonitorEditView', () => {
  it('shows a loading state while the monitor or billing info is pending', () => {
    monitorPending.value = true
    const wrapper = mount(PortMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Loading…')
  })

  it('populates the form from the loaded monitor', () => {
    detailData.value = { ...detail, monitor: { ...monitor, expectedState: 'closed' as const } }
    const wrapper = mount(PortMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect((wrapper.find('#name').element as HTMLInputElement).value).toBe('SMTP')
    expect((wrapper.find('#host').element as HTMLInputElement).value).toBe('mail.example.com')
    expect((wrapper.find('#port').element as HTMLInputElement).value).toBe('25')
    expect((wrapper.find('#expectedState').element as HTMLSelectElement).value).toBe('closed')
  })

  it('shows an inline error when loading the monitor fails', async () => {
    const wrapper = mount(PortMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    loadError.value = { message: 'Monitor not found' }
    await flushPromises()

    expect(wrapper.text()).toContain('Monitor not found')
  })

  it('shows a validation error when name is cleared', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(PortMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Name is required')
    expect(updatePortMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when host is cleared', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(PortMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#host').setValue('')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Host is required')
    expect(updatePortMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when port is out of range', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(PortMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#port').setValue(0)
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Port must be between 1 and 65535')
    expect(updatePortMock).not.toHaveBeenCalled()
  })

  it('updates the monitor and navigates to its detail page on success', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    updatePortMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(PortMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Renamed SMTP')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(updatePortMock).toHaveBeenCalledExactlyOnceWith('p1', {
      name: 'Renamed SMTP',
      host: 'mail.example.com',
      port: 25,
      expectedState: 'open',
      intervalMins: 5,
      alertsEnabled: true,
      maxAlertsPerIncident: 3,
      alertAfterNFailures: 0,
      channelIds: [],
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'port-monitor-detail',
      params: { id: 'p1' },
    })
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const { ApiError } = await import('@/api/client')
    updatePortMock.mockRejectedValueOnce(
      new ApiError(402, 'Upgrade to enable a faster interval', 'plan_limit_reached'),
    )
    const wrapper = mount(PortMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to enable a faster interval')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic inline error when saving fails for another reason', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    updatePortMock.mockRejectedValueOnce(new Error('Save failed'))
    const wrapper = mount(PortMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Save failed')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the detail page when cancel is clicked', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(PortMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'port-monitor-detail',
      params: { id: 'p1' },
    })
  })
})
