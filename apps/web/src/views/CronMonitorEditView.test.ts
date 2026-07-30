import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import CronMonitorEditView from './CronMonitorEditView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'c1' } }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { updateCronMock } = vi.hoisted(() => ({
  updateCronMock: vi.fn(),
}))

vi.mock('@/api/monitors', () => ({
  monitorsApi: { updateCron: updateCronMock },
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

vi.mock('@/composables/useCronMonitors', () => ({
  useCronMonitor: () => ({
    data: detailData,
    isPending: pending,
    error: loadError,
  }),
}))

const channelsData = ref<{ id: string; name: string; type: string; enabled: boolean }[]>([])
const channelsPending = ref(false)

vi.mock('@/composables/useNotificationChannels', () => ({
  useNotificationChannels: () => ({ data: channelsData, isPending: channelsPending }),
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
  maxDurationMins: 30,
  lastPingAt: null,
  nextPingAt: null,
  createdAt: '2026-01-01T00:00:00Z',
  channelIds: ['ch1'],
}

const detail = {
  monitor,
  pings: [],
  incidents: [],
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

function mountView() {
  return mount(CronMonitorEditView, {
    global: { stubs: { RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' } } },
  })
}

beforeEach(() => {
  detailData.value = null
  pending.value = false
  loadError.value = null
  channelsData.value = []
  channelsPending.value = false
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('CronMonitorEditView', () => {
  it('shows a loading state while the monitor is pending', () => {
    pending.value = true
    const wrapper = mountView()

    expect(wrapper.text()).toContain('Loading…')
    expect(wrapper.find('form').exists()).toBe(false)
  })

  it('shows a load error inline', async () => {
    const wrapper = mountView()
    loadError.value = { message: 'Monitor not found' }
    await flushPromises()

    expect(wrapper.text()).toContain('Monitor not found')
  })

  it('populates the form fields from the loaded monitor', () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mountView()

    expect((wrapper.find('#name').element as HTMLInputElement).value).toBe('Nightly backup')
    expect((wrapper.find('#schedule').element as HTMLInputElement).value).toBe('0 0 * * *')
    expect((wrapper.find('select#grace').element as HTMLSelectElement).value).toBe('5')
    expect((wrapper.find('#alerts').element as HTMLInputElement).checked).toBe(true)
    const maxDurationSelect = wrapper.find('select#maxDuration').element as HTMLSelectElement
    expect(maxDurationSelect.options[maxDurationSelect.selectedIndex].text).toBe('30 min')
  })

  it('shows a validation error when name is cleared', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mountView()

    await wrapper.find('#name').setValue('')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Name is required')
    expect(updateCronMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when schedule is cleared', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mountView()

    await wrapper.find('#schedule').setValue('')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Schedule is required')
    expect(updateCronMock).not.toHaveBeenCalled()
  })

  it('saves changes and navigates to the detail view on success', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    updateCronMock.mockResolvedValueOnce({ ...monitor, name: 'Renamed backup' })
    const wrapper = mountView()

    await wrapper.find('#name').setValue('Renamed backup')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(updateCronMock).toHaveBeenCalledExactlyOnceWith('c1', {
      name: 'Renamed backup',
      schedule: '0 0 * * *',
      gracePeriodMins: 5,
      alertsEnabled: true,
      maxAlertsPerIncident: 3,
      maxDurationMins: 30,
      channelIds: ['ch1'],
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'cron-monitor-detail',
      params: { id: 'c1' },
    })
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    updateCronMock.mockRejectedValueOnce(
      new ApiErrorMock(402, 'Upgrade to change alert settings', 'plan_limit_reached'),
    )
    const wrapper = mountView()

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to change alert settings')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic error message when saving fails for another reason', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    updateCronMock.mockRejectedValueOnce(new Error('Network error'))
    const wrapper = mountView()

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Network error')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the detail view when back or cancel is clicked', async () => {
    detailData.value = { ...detail, monitor: { ...monitor } }
    const wrapper = mountView()

    await findButtonByText(wrapper, '← Back')!.trigger('click')
    expect(pushMock).toHaveBeenCalledWith({ name: 'cron-monitor-detail', params: { id: 'c1' } })

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')
    expect(pushMock).toHaveBeenCalledWith({ name: 'cron-monitor-detail', params: { id: 'c1' } })
  })
})
