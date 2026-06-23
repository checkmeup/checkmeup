import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import UptimeMonitorListView from './UptimeMonitorListView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {} }),
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const refetchMock = vi.fn()
const listData = ref<unknown[]>([])
const listPending = ref(false)
const listError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useUptimeMonitors', () => ({
  useUptimeMonitors: () => ({
    data: listData,
    isPending: listPending,
    error: listError,
    refetch: refetchMock,
  }),
}))

const monitors = [
  {
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
  },
  {
    id: 'u2',
    name: 'Marketing site',
    url: 'https://example.com',
    intervalMins: 10,
    status: 'down' as const,
    alertsEnabled: true,
    maxAlertsPerIncident: 3,
    lastCheckedAt: null,
    createdAt: '2026-01-02T00:00:00Z',
    uptime24h: null,
    keyword: 'Welcome back',
    keywordMode: 'not_contains' as const,
    keywordCaseSensitive: false,
  },
]

beforeEach(() => {
  listData.value = []
  listPending.value = false
  listError.value = null
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('UptimeMonitorListView', () => {
  it('shows a loading state while pending', () => {
    listPending.value = true
    const wrapper = mount(UptimeMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message and retries on click', async () => {
    listError.value = { message: 'Failed to load monitors' }
    const wrapper = mount(UptimeMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Failed to load monitors')
    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Try again')!
      .trigger('click')

    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an empty state when there are no monitors', () => {
    const wrapper = mount(UptimeMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('No uptime monitors yet. Add one to start watching your URLs.')
  })

  it('navigates to the create view when "Add your first monitor" is clicked', async () => {
    const wrapper = mount(UptimeMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Add your first monitor')!
      .trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'uptime-monitor-create' })
  })

  it('navigates to the create view when the header "Add monitor" button is clicked', async () => {
    const wrapper = mount(UptimeMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Add monitor')!
      .trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'uptime-monitor-create' })
  })

  it('renders monitor rows with status, uptime, and keyword label', () => {
    listData.value = monitors
    const wrapper = mount(UptimeMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('API uptime')
    expect(wrapper.text()).toContain('Up')
    expect(wrapper.text()).toContain('99.90%')
    expect(wrapper.text()).toContain('Marketing site')
    expect(wrapper.text()).toContain('Down')
    expect(wrapper.text()).toContain('does not contain')
    expect(wrapper.text()).toContain('Welcome back')
  })

  it('formats a missing uptime and last-checked time as a dash', () => {
    listData.value = monitors
    const wrapper = mount(UptimeMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const desktopRows = wrapper.find('table').findAll('tbody tr')
    const downRow = desktopRows.find((r) => r.text().includes('Marketing site'))!
    expect(downRow.text()).toContain('—')
  })

  it('navigates to the detail view when a row is clicked', async () => {
    listData.value = monitors
    const wrapper = mount(UptimeMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const row = wrapper.find('table').findAll('tbody tr').at(0)!
    await row.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'uptime-monitor-detail',
      params: { id: 'u1' },
    })
  })
})
