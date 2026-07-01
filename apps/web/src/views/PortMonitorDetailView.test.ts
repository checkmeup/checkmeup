import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PortMonitorDetailView from './PortMonitorDetailView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'p1' } }),
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { pausePortMock, resumePortMock, deletePortMock } = vi.hoisted(() => ({
  pausePortMock: vi.fn(),
  resumePortMock: vi.fn(),
  deletePortMock: vi.fn(),
}))

vi.mock('@/api/monitors', () => ({
  monitorsApi: {
    pausePort: pausePortMock,
    resumePort: resumePortMock,
    deletePort: deletePortMock,
  },
}))

const refetchMock = vi.fn()
const detailData = ref<unknown>(null)
const pending = ref(false)
const queryError = ref<{ message: string } | null>(null)

vi.mock('@/composables/usePortMonitors', () => ({
  usePortMonitor: () => ({
    data: detailData,
    isPending: pending,
    error: queryError,
    refetch: refetchMock,
  }),
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
  lastCheckedAt: new Date(Date.now() - 5 * 60000).toISOString(),
  createdAt: '2026-01-01T00:00:00Z',
  uptime24h: 99.9,
}

const detail = {
  monitor,
  chartData: [
    {
      id: 'c1',
      checkedAt: new Date(Date.now() - 2 * 3600000).toISOString(),
      responseTimeMs: 20,
      isUp: true,
      failureReason: null,
    },
    {
      id: 'c2',
      checkedAt: new Date(Date.now() - 1 * 3600000).toISOString(),
      responseTimeMs: 24,
      isUp: true,
      failureReason: null,
    },
  ],
  checks: [
    {
      id: 'c1',
      checkedAt: new Date(Date.now() - 5 * 60000).toISOString(),
      responseTimeMs: 20,
      isUp: true,
      failureReason: null,
    },
    {
      id: 'c2',
      checkedAt: new Date(Date.now() - 3000).toISOString(),
      responseTimeMs: 10000,
      isUp: false,
      failureReason: 'connection refused / timeout',
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

describe('PortMonitorDetailView', () => {
  it('shows a loading state while pending', () => {
    pending.value = true
    const wrapper = mount(PortMonitorDetailView)

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message when the query fails', () => {
    queryError.value = { message: 'Monitor not found' }
    const wrapper = mount(PortMonitorDetailView)

    expect(wrapper.text()).toContain('Monitor not found')
  })

  it('renders monitor header details', () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(PortMonitorDetailView)

    expect(wrapper.text()).toContain('SMTP')
    expect(wrapper.text()).toContain('Up')
    expect(wrapper.text()).toContain('mail.example.com:25')
    expect(wrapper.text()).toContain('Every 5 min')
    expect(wrapper.text()).toContain('Last checked 5m ago')
    expect(wrapper.text()).toContain('Expected open')
  })

  it('renders the closed expected-state line', () => {
    detailData.value = { ...detail, monitor: { ...monitor, expectedState: 'closed' as const } }
    const wrapper = mount(PortMonitorDetailView)

    expect(wrapper.text()).toContain('Expected closed')
  })

  it('pauses an active monitor and refetches', async () => {
    detailData.value = { ...detail, monitor: { ...monitor, status: 'up' } }
    pausePortMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(PortMonitorDetailView)

    await findButtonByText(wrapper, 'Pause')!.trigger('click')
    await flushPromises()

    expect(pausePortMock).toHaveBeenCalledExactlyOnceWith('p1')
    expect(resumePortMock).not.toHaveBeenCalled()
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('resumes a paused monitor and refetches', async () => {
    detailData.value = { ...detail, monitor: { ...monitor, status: 'paused' } }
    resumePortMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(PortMonitorDetailView)

    await findButtonByText(wrapper, 'Resume')!.trigger('click')
    await flushPromises()

    expect(resumePortMock).toHaveBeenCalledExactlyOnceWith('p1')
    expect(pausePortMock).not.toHaveBeenCalled()
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an error and does not refetch when toggling pause fails', async () => {
    detailData.value = { ...detail, monitor: { ...monitor, status: 'up' } }
    pausePortMock.mockRejectedValueOnce(new Error('Action failed'))
    const wrapper = mount(PortMonitorDetailView)

    await findButtonByText(wrapper, 'Pause')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Action failed')
    expect(refetchMock).not.toHaveBeenCalled()
  })

  it('cancels the delete confirmation', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(PortMonitorDetailView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    expect(findButtonByText(wrapper, 'Confirm delete')).toBeTruthy()

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')
    expect(findButtonByText(wrapper, 'Delete')).toBeTruthy()
    expect(deletePortMock).not.toHaveBeenCalled()
  })

  it('deletes the monitor and navigates back to the list on confirm', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    deletePortMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(PortMonitorDetailView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(deletePortMock).toHaveBeenCalledExactlyOnceWith('p1')
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'port-monitors' })
  })

  it('shows an error and resets confirmation state when delete fails', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    deletePortMock.mockRejectedValueOnce(new Error('Delete failed'))
    const wrapper = mount(PortMonitorDetailView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Delete failed')
    expect(pushMock).not.toHaveBeenCalled()
    expect(findButtonByText(wrapper, 'Delete')).toBeTruthy()
  })

  it('renders uptime stats, formatting a missing 30d value as a dash', () => {
    detailData.value = { ...detail }
    const wrapper = mount(PortMonitorDetailView)

    expect(wrapper.text()).toContain('99.90%')
    expect(wrapper.text()).toContain('99.50%')
    expect(wrapper.text()).toContain('Uptime 30d')
    const cells = wrapper.findAll('.grid.grid-cols-3 > div')
    const uptime30d = cells.find((c) => c.text().includes('Uptime 30d'))
    expect(uptime30d!.text()).toContain('—')
  })

  it('renders the connect time chart when there are at least two chart points', () => {
    detailData.value = { ...detail }
    const wrapper = mount(PortMonitorDetailView)

    expect(wrapper.find('svg polyline').exists()).toBe(true)
    expect(wrapper.findAll('svg circle')).toHaveLength(2)
  })

  it('shows an empty state when there is no chart data', () => {
    detailData.value = { ...detail, chartData: [] }
    const wrapper = mount(PortMonitorDetailView)

    expect(wrapper.find('svg').exists()).toBe(false)
    expect(wrapper.text()).toContain(
      'No checks yet — first check runs within the configured interval.',
    )
  })

  it('renders incidents with resolved duration and an ongoing badge', () => {
    detailData.value = { ...detail }
    const wrapper = mount(PortMonitorDetailView)

    const rows = wrapper.findAll('table').at(0)!.findAll('tbody tr')
    expect(rows).toHaveLength(2)
    expect(rows[0]!.text()).toContain('1d')
    expect(rows[1]!.text()).toContain('Ongoing')
  })

  it('shows an empty state when there are no incidents', () => {
    detailData.value = { ...detail, incidents: [] }
    const wrapper = mount(PortMonitorDetailView)

    expect(wrapper.text()).toContain('No incidents recorded.')
  })

  it('renders the check log with up/down status and connect time', () => {
    detailData.value = { ...detail }
    const wrapper = mount(PortMonitorDetailView)

    const tables = wrapper.findAll('table')
    const checkRows = tables.at(-1)!.findAll('tbody tr')
    expect(checkRows).toHaveLength(2)
    expect(checkRows[0]!.text()).toContain('✓ Up')
    expect(checkRows[1]!.text()).toContain('✗ Down')
    expect(checkRows[1]!.text()).toContain('connection refused / timeout')
  })

  it('shows an empty state when there are no checks', () => {
    detailData.value = { ...detail, checks: [] }
    const wrapper = mount(PortMonitorDetailView)

    expect(wrapper.text()).toContain('No checks yet.')
  })
})
