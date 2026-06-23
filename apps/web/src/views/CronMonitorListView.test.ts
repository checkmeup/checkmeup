import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import CronMonitorListView from './CronMonitorListView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {} }),
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const refetchMock = vi.fn()
const listData = ref<unknown[] | null>(null)
const pending = ref(false)
const queryError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useCronMonitors', () => ({
  useCronMonitors: () => ({
    data: listData,
    isPending: pending,
    error: queryError,
    refetch: refetchMock,
  }),
}))

const monitors = [
  {
    id: 'c1',
    name: 'Nightly backup',
    schedule: '0 0 * * *',
    status: 'up' as const,
    lastPingAt: new Date(Date.now() - 5 * 60000).toISOString(),
    nextPingAt: new Date(Date.now() + 55 * 60000).toISOString(),
  },
  {
    id: 'c2',
    name: 'Weekly cleanup',
    schedule: 'every 7d',
    status: 'paused' as const,
    lastPingAt: null,
    nextPingAt: null,
  },
]

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  listData.value = null
  pending.value = false
  queryError.value = null
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('CronMonitorListView', () => {
  it('shows a loading state while pending', () => {
    pending.value = true
    const wrapper = mount(CronMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message and retries on click', async () => {
    queryError.value = { message: 'Failed to load monitors' }
    const wrapper = mount(CronMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Failed to load monitors')

    await findButtonByText(wrapper, 'Try again')!.trigger('click')
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an empty state with a call to action when there are no monitors', async () => {
    listData.value = []
    const wrapper = mount(CronMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('No cron monitors yet')

    await findButtonByText(wrapper, 'Add your first monitor')!.trigger('click')
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'cron-monitor-create' })
  })

  it('renders a row per monitor with name, status, and schedule', () => {
    listData.value = [...monitors]
    const wrapper = mount(CronMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Nightly backup')
    expect(wrapper.text()).toContain('Up')
    expect(wrapper.text()).toContain('0 0 * * *')
    expect(wrapper.text()).toContain('Weekly cleanup')
    expect(wrapper.text()).toContain('Paused')
    expect(wrapper.text()).toContain('every 7d')
  })

  it('navigates to the create view when Add monitor is clicked', async () => {
    listData.value = [...monitors]
    const wrapper = mount(CronMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Add monitor')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'cron-monitor-create' })
  })

  it('navigates to the detail view when a desktop row is clicked', async () => {
    listData.value = [...monitors]
    const wrapper = mount(CronMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const row = wrapper.findAll('table tbody tr').find((r) => r.text().includes('Nightly backup'))
    await row!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'cron-monitor-detail',
      params: { id: 'c1' },
    })
  })

  it('navigates to the detail view when a mobile card is clicked', async () => {
    listData.value = [...monitors]
    const wrapper = mount(CronMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const card = wrapper
      .findAll('.md\\:hidden > div')
      .find((c) => c.text().includes('Weekly cleanup'))
    await card!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'cron-monitor-detail',
      params: { id: 'c2' },
    })
  })

  it('renders a dash for monitors with no last ping', () => {
    listData.value = [monitors[1]]
    const wrapper = mount(CronMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Last: —')
  })
})
