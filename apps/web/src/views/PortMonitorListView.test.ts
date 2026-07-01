import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import PortMonitorListView from './PortMonitorListView.vue'

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

vi.mock('@/composables/usePortMonitors', () => ({
  usePortMonitors: () => ({
    data: listData,
    isPending: listPending,
    error: listError,
    refetch: refetchMock,
  }),
}))

const monitors = [
  {
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
  },
  {
    id: 'p2',
    name: 'Locked-down admin panel',
    host: 'internal.example.com',
    port: 9090,
    expectedState: 'closed' as const,
    intervalMins: 10,
    status: 'down' as const,
    alertsEnabled: true,
    maxAlertsPerIncident: 3,
    alertAfterNFailures: 0,
    lastCheckedAt: null,
    createdAt: '2026-01-02T00:00:00Z',
    uptime24h: null,
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

describe('PortMonitorListView', () => {
  it('shows a loading state while pending', () => {
    listPending.value = true
    const wrapper = mount(PortMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message and retries on click', async () => {
    listError.value = { message: 'Failed to load monitors' }
    const wrapper = mount(PortMonitorListView, {
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
    const wrapper = mount(PortMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('No port monitors yet.')
  })

  it('navigates to the create view when "Add your first monitor" is clicked', async () => {
    const wrapper = mount(PortMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Add your first monitor')!
      .trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'port-monitor-create' })
  })

  it('navigates to the create view when the header "Add monitor" button is clicked', async () => {
    const wrapper = mount(PortMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Add monitor')!
      .trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'port-monitor-create' })
  })

  it('renders monitor rows with host:port, status, uptime, and expected state', () => {
    listData.value = monitors
    const wrapper = mount(PortMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('SMTP')
    expect(wrapper.text()).toContain('mail.example.com:25')
    expect(wrapper.text()).toContain('Up')
    expect(wrapper.text()).toContain('99.90%')
    expect(wrapper.text()).toContain('Locked-down admin panel')
    expect(wrapper.text()).toContain('Down')
    expect(wrapper.text()).toContain('Closed')
  })

  it('navigates to the detail view when a row is clicked', async () => {
    listData.value = monitors
    const wrapper = mount(PortMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const row = wrapper.find('table').findAll('tbody tr').at(0)!
    await row.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'port-monitor-detail',
      params: { id: 'p1' },
    })
  })
})
