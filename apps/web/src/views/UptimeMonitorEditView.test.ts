import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import UptimeMonitorEditView from './UptimeMonitorEditView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'u1' } }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { updateUptimeMock } = vi.hoisted(() => ({
  updateUptimeMock: vi.fn(),
}))

vi.mock('@/api/monitors', async () => {
  const actual = await vi.importActual<typeof import('@/api/monitors')>('@/api/monitors')
  return {
    ...actual,
    monitorsApi: { updateUptime: updateUptimeMock },
  }
})

const detailData = ref<unknown>(null)
const monitorPending = ref(false)
const loadError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useUptimeMonitors', () => ({
  useUptimeMonitor: () => ({ data: detailData, isPending: monitorPending, error: loadError }),
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
  id: 'u1',
  name: 'API uptime',
  url: 'https://api.example.com/health',
  intervalMins: 5,
  status: 'up' as const,
  alertsEnabled: true,
  maxAlertsPerIncident: 3,
  lastCheckedAt: null,
  createdAt: '2026-01-01T00:00:00Z',
  uptime24h: 99.9,
  keyword: null,
  keywordMode: 'contains' as const,
  keywordCaseSensitive: false,
  jsonAssertions: [],
  maxResponseTimeMs: 10000,
  httpMethod: 'GET' as const,
  acceptedStatusCodes: [200],
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

describe('UptimeMonitorEditView', () => {
  it('shows a loading state while the monitor or billing info is pending', () => {
    monitorPending.value = true
    const wrapper = mount(UptimeMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows a loading state while billing info is pending even if the monitor loaded', () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    billingPending.value = true
    const wrapper = mount(UptimeMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Loading…')
  })

  it('populates the form from the loaded monitor', () => {
    detailData.value = {
      ...detail,
      monitor: {
        ...monitor,
        keyword: 'healthy',
        keywordMode: 'not_contains',
        keywordCaseSensitive: true,
      },
    }
    const wrapper = mount(UptimeMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect((wrapper.find('#name').element as HTMLInputElement).value).toBe('API uptime')
    expect((wrapper.find('#url').element as HTMLInputElement).value).toBe(
      'https://api.example.com/health',
    )
    expect((wrapper.find('#keyword').element as HTMLInputElement).value).toBe('healthy')
    expect(wrapper.find('#keywordMode').exists()).toBe(true)
  })

  it('shows an inline error when loading the monitor fails', async () => {
    const wrapper = mount(UptimeMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    loadError.value = { message: 'Monitor not found' }
    await flushPromises()

    expect(wrapper.text()).toContain('Monitor not found')
  })

  it('shows a validation error when name is cleared', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(UptimeMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Name is required')
    expect(updateUptimeMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when url is invalid', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(UptimeMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#url').setValue('not-a-url')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('URL must start with http:// or https://')
    expect(updateUptimeMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when keyword exceeds 500 characters', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(UptimeMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#keyword').setValue('a'.repeat(501))
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Keyword must be 500 characters or fewer')
    expect(updateUptimeMock).not.toHaveBeenCalled()
  })

  it('updates the monitor and navigates to its detail page on success', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    updateUptimeMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(UptimeMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Renamed API')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(updateUptimeMock).toHaveBeenCalledExactlyOnceWith('u1', {
      name: 'Renamed API',
      url: 'https://api.example.com/health',
      intervalMins: 5,
      alertsEnabled: true,
      maxAlertsPerIncident: 3,
      keyword: '',
      keywordMode: 'contains',
      keywordCaseSensitive: false,
      jsonAssertions: [],
      maxResponseTimeMs: 10000,
      httpMethod: 'GET',
      acceptedStatusCodes: [200],
      channelIds: [],
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'uptime-monitor-detail',
      params: { id: 'u1' },
    })
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const { ApiError } = await import('@/api/client')
    updateUptimeMock.mockRejectedValueOnce(
      new ApiError(402, 'Upgrade to enable alerts', 'plan_limit_reached'),
    )
    const wrapper = mount(UptimeMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to enable alerts')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic inline error when saving fails for another reason', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    updateUptimeMock.mockRejectedValueOnce(new Error('Save failed'))
    const wrapper = mount(UptimeMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Save failed')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the detail page when cancel is clicked', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(UptimeMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'uptime-monitor-detail',
      params: { id: 'u1' },
    })
  })
})
