import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import UptimeMonitorDetailView from './UptimeMonitorDetailView.vue'
import { ApiError } from '@/api/client'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'u1' } }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { pauseUptimeMock, resumeUptimeMock, deleteUptimeMock } = vi.hoisted(() => ({
  pauseUptimeMock: vi.fn(),
  resumeUptimeMock: vi.fn(),
  deleteUptimeMock: vi.fn(),
}))

vi.mock('@/api/monitors', () => ({
  monitorsApi: {
    pauseUptime: pauseUptimeMock,
    resumeUptime: resumeUptimeMock,
    deleteUptime: deleteUptimeMock,
  },
}))

const refetchMock = vi.fn()
const detailData = ref<unknown>(null)
const pending = ref(false)
const queryError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useUptimeMonitors', () => ({
  useUptimeMonitor: () => ({
    data: detailData,
    isPending: pending,
    error: queryError,
    refetch: refetchMock,
  }),
}))

const monitor = {
  id: 'u1',
  name: 'API uptime',
  url: 'https://api.example.com/health',
  intervalMins: 5,
  status: 'up' as const,
  alertsEnabled: true,
  maxAlertsPerIncident: 3,
  lastCheckedAt: new Date(Date.now() - 5 * 60000).toISOString(),
  createdAt: '2026-01-01T00:00:00Z',
  uptime24h: 99.9,
  keyword: null,
  keywordMode: 'contains' as const,
  keywordCaseSensitive: false,
  jsonAssertions: [],
  maxResponseTimeMs: null,
}

const detail = {
  monitor,
  chartData: [
    {
      id: 'c1',
      checkedAt: new Date(Date.now() - 2 * 3600000).toISOString(),
      statusCode: 200,
      responseTimeMs: 120,
      isUp: true,
      failureReason: null,
    },
    {
      id: 'c2',
      checkedAt: new Date(Date.now() - 1 * 3600000).toISOString(),
      statusCode: 200,
      responseTimeMs: 140,
      isUp: true,
      failureReason: null,
    },
  ],
  checks: [
    {
      id: 'c1',
      checkedAt: new Date(Date.now() - 5 * 60000).toISOString(),
      statusCode: 200,
      responseTimeMs: 120,
      isUp: true,
      failureReason: null,
    },
    {
      id: 'c2',
      checkedAt: new Date(Date.now() - 3000).toISOString(),
      statusCode: 500,
      responseTimeMs: 80,
      isUp: false,
      failureReason: 'Server error',
    },
  ],
  incidents: [
    {
      id: 'i1',
      startedAt: new Date(Date.now() - 2 * 86400000).toISOString(),
      resolvedAt: new Date(Date.now() - 1 * 86400000).toISOString(),
    },
    { id: 'i2', startedAt: new Date(Date.now() - 3600000).toISOString(), resolvedAt: null },
  ],
  stats: { uptime24h: 99.9, uptime7d: 99.5, uptime30d: null },
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  detailData.value = null
  pending.value = false
  queryError.value = null
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('UptimeMonitorDetailView', () => {
  it('shows a loading state while pending', () => {
    pending.value = true
    const wrapper = mount(UptimeMonitorDetailView)

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message when the query fails', () => {
    queryError.value = { message: 'Monitor not found' }
    const wrapper = mount(UptimeMonitorDetailView)

    expect(wrapper.text()).toContain('Monitor not found')
  })

  it('renders monitor header details', () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(UptimeMonitorDetailView)

    expect(wrapper.text()).toContain('API uptime')
    expect(wrapper.text()).toContain('Up')
    expect(wrapper.text()).toContain('https://api.example.com/health')
    expect(wrapper.text()).toContain('Every 5 min')
    expect(wrapper.text()).toContain('Last checked 5m ago')
  })

  it('renders the keyword line only when a keyword is configured', () => {
    detailData.value = {
      ...detail,
      monitor: {
        ...monitor,
        keyword: 'healthy',
        keywordMode: 'not_contains',
        keywordCaseSensitive: true,
      },
    }
    const wrapper = mount(UptimeMonitorDetailView)

    expect(wrapper.text()).toContain('Does not contain')
    expect(wrapper.text()).toContain('"healthy"')
    expect(wrapper.text()).toContain('(case-sensitive)')
  })

  it('pauses an active monitor and refetches', async () => {
    detailData.value = { ...detail, monitor: { ...monitor, status: 'up' } }
    pauseUptimeMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(UptimeMonitorDetailView)

    await findButtonByText(wrapper, 'Pause')!.trigger('click')
    await flushPromises()

    expect(pauseUptimeMock).toHaveBeenCalledExactlyOnceWith('u1')
    expect(resumeUptimeMock).not.toHaveBeenCalled()
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('resumes a paused monitor and refetches', async () => {
    detailData.value = { ...detail, monitor: { ...monitor, status: 'paused' } }
    resumeUptimeMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(UptimeMonitorDetailView)

    await findButtonByText(wrapper, 'Resume')!.trigger('click')
    await flushPromises()

    expect(resumeUptimeMock).toHaveBeenCalledExactlyOnceWith('u1')
    expect(pauseUptimeMock).not.toHaveBeenCalled()
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an error and does not refetch when toggling pause fails', async () => {
    detailData.value = { ...detail, monitor: { ...monitor, status: 'up' } }
    pauseUptimeMock.mockRejectedValueOnce(new Error('Action failed'))
    const wrapper = mount(UptimeMonitorDetailView)

    await findButtonByText(wrapper, 'Pause')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Action failed')
    expect(refetchMock).not.toHaveBeenCalled()
  })

  it('shows an upgrade prompt instead of a plain error when resume is blocked by the plan limit', async () => {
    detailData.value = { ...detail, monitor: { ...monitor, status: 'paused' } }
    resumeUptimeMock.mockRejectedValueOnce(
      new ApiError(
        402,
        'monitor limit reached for your plan — upgrade to add more',
        'plan_limit_reached',
      ),
    )
    const wrapper = mount(UptimeMonitorDetailView)

    await findButtonByText(wrapper, 'Resume')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('monitor limit reached for your plan')
    expect(wrapper.text()).toContain('View plans')
    expect(refetchMock).not.toHaveBeenCalled()
  })

  it('cancels the delete confirmation', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(UptimeMonitorDetailView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    expect(findButtonByText(wrapper, 'Confirm delete')).toBeTruthy()

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')
    expect(findButtonByText(wrapper, 'Delete')).toBeTruthy()
    expect(deleteUptimeMock).not.toHaveBeenCalled()
  })

  it('deletes the monitor and navigates back to the list on confirm', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    deleteUptimeMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(UptimeMonitorDetailView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(deleteUptimeMock).toHaveBeenCalledExactlyOnceWith('u1')
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'uptime-monitors' })
  })

  it('shows an error and resets confirmation state when delete fails', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    deleteUptimeMock.mockRejectedValueOnce(new Error('Delete failed'))
    const wrapper = mount(UptimeMonitorDetailView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Delete failed')
    expect(pushMock).not.toHaveBeenCalled()
    expect(findButtonByText(wrapper, 'Delete')).toBeTruthy()
  })

  it('renders uptime stats, formatting a missing 30d value as a dash', () => {
    detailData.value = { ...detail }
    const wrapper = mount(UptimeMonitorDetailView)

    expect(wrapper.text()).toContain('99.90%')
    expect(wrapper.text()).toContain('99.50%')
    expect(wrapper.text()).toContain('Uptime 30d')
    const cells = wrapper.findAll('.grid.grid-cols-3 > div')
    const uptime30d = cells.find((c) => c.text().includes('Uptime 30d'))
    expect(uptime30d!.text()).toContain('—')
  })

  it('renders the response time chart when there are at least two chart points', () => {
    detailData.value = { ...detail }
    const wrapper = mount(UptimeMonitorDetailView)

    expect(wrapper.find('svg polyline').exists()).toBe(true)
    expect(wrapper.findAll('svg circle')).toHaveLength(2)
  })

  it('shows an empty state when there is no chart data', () => {
    detailData.value = { ...detail, chartData: [] }
    const wrapper = mount(UptimeMonitorDetailView)

    expect(wrapper.find('svg').exists()).toBe(false)
    expect(wrapper.text()).toContain(
      'No checks yet — first check runs within the configured interval.',
    )
  })

  it('renders incidents with resolved duration and an ongoing badge', () => {
    detailData.value = { ...detail }
    const wrapper = mount(UptimeMonitorDetailView)

    const rows = wrapper.findAll('table').at(0)!.findAll('tbody tr')
    expect(rows).toHaveLength(2)
    expect(rows[0]!.text()).toContain('1d')
    expect(rows[1]!.text()).toContain('Ongoing')
  })

  it('shows an empty state when there are no incidents', () => {
    detailData.value = { ...detail, incidents: [] }
    const wrapper = mount(UptimeMonitorDetailView)

    expect(wrapper.text()).toContain('No incidents recorded.')
  })

  it('renders the check log with up/down status and a dash for a missing status code', () => {
    detailData.value = {
      ...detail,
      checks: [
        {
          id: 'c1',
          checkedAt: new Date().toISOString(),
          statusCode: null,
          responseTimeMs: 50,
          isUp: true,
          failureReason: null,
        },
        {
          id: 'c2',
          checkedAt: new Date().toISOString(),
          statusCode: 500,
          responseTimeMs: 80,
          isUp: false,
          failureReason: 'Timeout',
        },
      ],
    }
    const wrapper = mount(UptimeMonitorDetailView)

    const tables = wrapper.findAll('table')
    const checkRows = tables.at(-1)!.findAll('tbody tr')
    expect(checkRows).toHaveLength(2)
    expect(checkRows[0]!.text()).toContain('✓ Up')
    expect(checkRows[0]!.text()).toContain('—')
    expect(checkRows[1]!.text()).toContain('✗ Down')
    expect(checkRows[1]!.text()).toContain('Timeout')
  })

  it('shows an empty state when there are no checks', () => {
    detailData.value = { ...detail, checks: [] }
    const wrapper = mount(UptimeMonitorDetailView)

    expect(wrapper.text()).toContain('No checks yet.')
  })
})
