import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import DNSMonitorEditView from './DNSMonitorEditView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'd1' } }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { updateDnsMock } = vi.hoisted(() => ({
  updateDnsMock: vi.fn(),
}))

vi.mock('@/api/monitors', async () => {
  const actual = await vi.importActual<typeof import('@/api/monitors')>('@/api/monitors')
  return {
    ...actual,
    monitorsApi: { updateDns: updateDnsMock },
  }
})

const detailData = ref<unknown>(null)
const monitorPending = ref(false)
const loadError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useDNSMonitors', () => ({
  useDNSMonitor: () => ({ data: detailData, isPending: monitorPending, error: loadError }),
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
  id: 'd1',
  name: 'Apex A record',
  hostname: 'example.com',
  recordType: 'A' as const,
  expectedValue: '1.2.3.4',
  baselineCaptured: false,
  lastResolvedValue: '1.2.3.4',
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

describe('DNSMonitorEditView', () => {
  it('shows a loading state while the monitor or billing info is pending', () => {
    monitorPending.value = true
    const wrapper = mount(DNSMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Loading…')
  })

  it('populates the form from the loaded monitor', () => {
    detailData.value = {
      ...detail,
      monitor: { ...monitor, recordType: 'MX' as const, expectedValue: 'mail.example.com' },
    }
    const wrapper = mount(DNSMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect((wrapper.find('#name').element as HTMLInputElement).value).toBe('Apex A record')
    expect((wrapper.find('#hostname').element as HTMLInputElement).value).toBe('example.com')
    expect((wrapper.find('#recordType').element as HTMLSelectElement).value).toBe('MX')
    expect((wrapper.find('#expectedValue').element as HTMLInputElement).value).toBe(
      'mail.example.com',
    )
  })

  it('shows an inline error when loading the monitor fails', async () => {
    const wrapper = mount(DNSMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    loadError.value = { message: 'Monitor not found' }
    await flushPromises()

    expect(wrapper.text()).toContain('Monitor not found')
  })

  it('shows a validation error when name is cleared', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(DNSMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Name is required')
    expect(updateDnsMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when hostname is cleared', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(DNSMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#hostname').setValue('')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Hostname is required')
    expect(updateDnsMock).not.toHaveBeenCalled()
  })

  it('updates the monitor and navigates to its detail page on success', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    updateDnsMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(DNSMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Renamed A record')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(updateDnsMock).toHaveBeenCalledExactlyOnceWith('d1', {
      name: 'Renamed A record',
      hostname: 'example.com',
      recordType: 'A',
      expectedValue: '1.2.3.4',
      intervalMins: 5,
      alertsEnabled: true,
      maxAlertsPerIncident: 3,
      alertAfterNFailures: 0,
      channelIds: [],
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'dns-monitor-detail',
      params: { id: 'd1' },
    })
  })

  it('clearing the expected value submits an empty string, re-arming baseline mode', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    updateDnsMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(DNSMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#expectedValue').setValue('')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(updateDnsMock).toHaveBeenCalledExactlyOnceWith(
      'd1',
      expect.objectContaining({ expectedValue: '' }),
    )
  })

  it('changing the record type clears the stale expected value', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(DNSMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect((wrapper.find('#expectedValue').element as HTMLInputElement).value).toBe('1.2.3.4')
    await wrapper.find('#recordType').setValue('TXT')

    expect((wrapper.find('#expectedValue').element as HTMLInputElement).value).toBe('')
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const { ApiError } = await import('@/api/client')
    updateDnsMock.mockRejectedValueOnce(
      new ApiError(402, 'Upgrade to enable a faster interval', 'plan_limit_reached'),
    )
    const wrapper = mount(DNSMonitorEditView, {
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
    updateDnsMock.mockRejectedValueOnce(new Error('Save failed'))
    const wrapper = mount(DNSMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Save failed')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the detail page when cancel is clicked', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(DNSMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'dns-monitor-detail',
      params: { id: 'd1' },
    })
  })
})
