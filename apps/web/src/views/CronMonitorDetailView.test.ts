import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import CronMonitorDetailView from './CronMonitorDetailView.vue'
import { ApiError } from '@/api/client'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'c1' } }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { pauseCronMock, resumeCronMock, deleteCronMock } = vi.hoisted(() => ({
  pauseCronMock: vi.fn(),
  resumeCronMock: vi.fn(),
  deleteCronMock: vi.fn(),
}))

vi.mock('@/api/monitors', () => ({
  monitorsApi: {
    pauseCron: pauseCronMock,
    resumeCron: resumeCronMock,
    deleteCron: deleteCronMock,
  },
}))

const refetchMock = vi.fn()
const detailData = ref<unknown>(null)
const pending = ref(false)
const queryError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useCronMonitors', () => ({
  useCronMonitor: () => ({
    data: detailData,
    isPending: pending,
    error: queryError,
    refetch: refetchMock,
  }),
}))

const monitor = {
  id: 'c1',
  name: 'Nightly backup',
  schedule: '0 0 * * *',
  gracePeriodMins: 5,
  pingToken: 'tok123',
  pingUrl: 'https://checkmeup.net/ping/tok123',
  status: 'up' as const,
  alertsEnabled: true,
  maxAlertsPerIncident: 3,
  lastPingAt: new Date(Date.now() - 5 * 60000).toISOString(),
  nextPingAt: new Date(Date.now() + 55 * 60000).toISOString(),
  createdAt: '2026-01-01T00:00:00Z',
}

const detail = {
  monitor,
  pings: [
    { id: 'p1', receivedAt: new Date(Date.now() - 5 * 60000).toISOString(), sourceIp: '1.2.3.4' },
    { id: 'p2', receivedAt: new Date(Date.now() - 65 * 60000).toISOString(), sourceIp: '' },
  ],
  incidents: [
    {
      id: 'i1',
      startedAt: new Date(Date.now() - 2 * 86400000).toISOString(),
      resolvedAt: new Date(Date.now() - 1 * 86400000).toISOString(),
    },
    { id: 'i2', startedAt: new Date(Date.now() - 3600000).toISOString(), resolvedAt: null },
  ],
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

describe('CronMonitorDetailView', () => {
  it('shows a loading state while pending', () => {
    pending.value = true
    const wrapper = mount(CronMonitorDetailView)

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message when the query fails', () => {
    queryError.value = { message: 'Monitor not found' }
    const wrapper = mount(CronMonitorDetailView)

    expect(wrapper.text()).toContain('Monitor not found')
  })

  it('renders monitor header and configuration details', () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(CronMonitorDetailView)

    expect(wrapper.text()).toContain('Nightly backup')
    expect(wrapper.text()).toContain('Up')
    expect(wrapper.text()).toContain('0 0 * * *')
    expect(wrapper.text()).toContain('5 min')
    expect(wrapper.text()).toContain('https://checkmeup.net/ping/tok123')
  })

  it('renders the incident list with resolved duration and an ongoing entry', () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(CronMonitorDetailView)

    expect(wrapper.text()).toContain('Resolved after 1d 0h')
    expect(wrapper.text()).toContain('Ongoing')
  })

  it('does not render the incidents card when there are no incidents', () => {
    detailData.value = { ...detail, monitor: { ...monitor }, incidents: [] }
    const wrapper = mount(CronMonitorDetailView)

    expect(wrapper.text()).not.toContain('Incidents')
  })

  it('renders the ping log', () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(CronMonitorDetailView)

    expect(wrapper.text()).toContain('1.2.3.4')
    expect(wrapper.text()).toContain('Execution log')
  })

  it('shows a run duration for a completion ping with a matching start ping (US-3404)', () => {
    const receivedAt = new Date(Date.now() - 5 * 60000).toISOString()
    const runStartedAt = new Date(Date.now() - 15 * 60000).toISOString()
    detailData.value = {
      ...detail,
      monitor: { ...monitor },
      pings: [{ id: 'p1', receivedAt, sourceIp: '1.2.3.4', runStartedAt }],
    }
    const wrapper = mount(CronMonitorDetailView)

    expect(wrapper.text()).toContain('10m')
  })

  it("shows '—' for a ping with no matching start ping", () => {
    detailData.value = {
      ...detail,
      monitor: { ...monitor },
      pings: [
        { id: 'p1', receivedAt: new Date().toISOString(), sourceIp: '1.2.3.4', runStartedAt: null },
      ],
    }
    const wrapper = mount(CronMonitorDetailView)

    expect(wrapper.text()).toContain('—')
  })

  it("renders a ping's metadata as key: value chips", () => {
    detailData.value = {
      ...detail,
      monitor: { ...monitor },
      pings: [
        {
          id: 'p1',
          receivedAt: new Date().toISOString(),
          sourceIp: '1.2.3.4',
          metadata: { build: '142', state: 'success' },
        },
      ],
    }
    const wrapper = mount(CronMonitorDetailView)

    expect(wrapper.text()).toContain('build: 142')
    expect(wrapper.text()).toContain('state: success')
  })

  it('truncates a long metadata value and exposes the full text via title', () => {
    const longValue = 'x'.repeat(256)
    detailData.value = {
      ...detail,
      monitor: { ...monitor },
      pings: [
        {
          id: 'p1',
          receivedAt: new Date().toISOString(),
          sourceIp: '1.2.3.4',
          metadata: { log: longValue },
        },
      ],
    }
    const wrapper = mount(CronMonitorDetailView)

    const chip = wrapper.find('span.truncate')
    expect(chip.exists()).toBe(true)
    expect(chip.attributes('title')).toBe(`log: ${longValue}`)
  })

  it('shows an empty state when there are no pings', () => {
    detailData.value = { ...detail, monitor: { ...monitor }, pings: [] }
    const wrapper = mount(CronMonitorDetailView)

    expect(wrapper.text()).toContain('No pings received yet. Add the ping URL to your cron job.')
  })

  it('pauses an active monitor and refetches', async () => {
    detailData.value = { ...detail, monitor: { ...monitor, status: 'up' } }
    pauseCronMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(CronMonitorDetailView)

    await findButtonByText(wrapper, 'Pause')!.trigger('click')
    await flushPromises()

    expect(pauseCronMock).toHaveBeenCalledExactlyOnceWith('c1')
    expect(resumeCronMock).not.toHaveBeenCalled()
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('resumes a paused monitor and refetches', async () => {
    detailData.value = { ...detail, monitor: { ...monitor, status: 'paused' } }
    resumeCronMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(CronMonitorDetailView)

    await findButtonByText(wrapper, 'Resume')!.trigger('click')
    await flushPromises()

    expect(resumeCronMock).toHaveBeenCalledExactlyOnceWith('c1')
    expect(pauseCronMock).not.toHaveBeenCalled()
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an upgrade prompt instead of a plain error when resume is blocked by the plan limit', async () => {
    detailData.value = { ...detail, monitor: { ...monitor, status: 'paused' } }
    resumeCronMock.mockRejectedValueOnce(
      new ApiError(
        402,
        'monitor limit reached for your plan — upgrade to add more',
        'plan_limit_reached',
      ),
    )
    const wrapper = mount(CronMonitorDetailView)

    await findButtonByText(wrapper, 'Resume')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('monitor limit reached for your plan')
    expect(wrapper.text()).toContain('View plans')
    expect(refetchMock).not.toHaveBeenCalled()
  })

  it('navigates to the edit view when Edit is clicked', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(CronMonitorDetailView)

    await findButtonByText(wrapper, 'Edit')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'cron-monitor-edit',
      params: { id: 'c1' },
    })
  })

  it('cancels the delete confirmation', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mount(CronMonitorDetailView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    expect(wrapper.text()).toContain('Delete monitor?')

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')
    expect(wrapper.text()).not.toContain('Delete monitor?')
    expect(deleteCronMock).not.toHaveBeenCalled()
  })

  it('deletes the monitor and navigates back to the list on confirm', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    deleteCronMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(CronMonitorDetailView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    const confirmButton = wrapper.findAll('button').filter((b) => b.text() === 'Delete')[1]
    await confirmButton!.trigger('click')
    await flushPromises()

    expect(deleteCronMock).toHaveBeenCalledExactlyOnceWith('c1')
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'cron-monitors' })
  })

  it('does not navigate and resets pending state when delete fails', async () => {
    // confirmDelete has no catch around monitorsApi.deleteCron, so the
    // rejection is intentionally left unhandled by the component; swallow it
    // here so the test can still assert the resulting UI state.
    const onUnhandledRejection = () => {}
    process.on('unhandledRejection', onUnhandledRejection)

    detailData.value = { ...detail, monitor: { ...monitor } }
    deleteCronMock.mockRejectedValueOnce(new Error('Delete failed'))
    const wrapper = mount(CronMonitorDetailView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    const confirmButton = wrapper.findAll('button').filter((b) => b.text() === 'Delete')[1]
    await confirmButton!.trigger('click')
    await flushPromises()

    expect(pushMock).not.toHaveBeenCalled()
    expect(wrapper.text()).not.toContain('Delete monitor?')

    process.off('unhandledRejection', onUnhandledRejection)
  })

  it('copies the ping URL to the clipboard when Copy is clicked', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const writeText = vi.fn()
    vi.stubGlobal('navigator', { ...navigator, clipboard: { writeText } })
    const wrapper = mount(CronMonitorDetailView)

    await findButtonByText(wrapper, 'Copy')!.trigger('click')

    expect(writeText).toHaveBeenCalledExactlyOnceWith('https://checkmeup.net/ping/tok123')
    vi.unstubAllGlobals()
  })
})
