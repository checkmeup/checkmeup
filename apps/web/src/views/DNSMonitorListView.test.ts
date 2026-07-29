import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import DNSMonitorListView from './DNSMonitorListView.vue'

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

vi.mock('@/composables/useDNSMonitors', () => ({
  useDNSMonitors: () => ({
    data: listData,
    isPending: listPending,
    error: listError,
    refetch: refetchMock,
  }),
}))

const monitors = [
  {
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
    lastCheckedAt: new Date(Date.now() - 5 * 60000).toISOString(),
    createdAt: '2026-01-01T00:00:00Z',
    uptime24h: 99.9,
  },
  {
    id: 'd2',
    name: 'Mail MX',
    hostname: 'mail.example.com',
    recordType: 'MX' as const,
    expectedValue: null,
    baselineCaptured: false,
    lastResolvedValue: null,
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

describe('DNSMonitorListView', () => {
  it('shows a loading state while pending', () => {
    listPending.value = true
    const wrapper = mount(DNSMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message and retries on click', async () => {
    listError.value = { message: 'Failed to load monitors' }
    const wrapper = mount(DNSMonitorListView, {
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
    const wrapper = mount(DNSMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('No DNS monitors yet.')
  })

  it('navigates to the create view when "Add your first monitor" is clicked', async () => {
    const wrapper = mount(DNSMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Add your first monitor')!
      .trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'dns-monitor-create' })
  })

  it('navigates to the create view when the header "Add monitor" button is clicked', async () => {
    const wrapper = mount(DNSMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Add monitor')!
      .trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'dns-monitor-create' })
  })

  it('renders monitor rows with hostname, record type, current value, and status', () => {
    listData.value = monitors
    const wrapper = mount(DNSMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Apex A record')
    expect(wrapper.text()).toContain('example.com')
    expect(wrapper.text()).toContain('1.2.3.4')
    expect(wrapper.text()).toContain('Up')
    expect(wrapper.text()).toContain('Mail MX')
    expect(wrapper.text()).toContain('Down')
  })

  it('navigates to the detail view when a row is clicked', async () => {
    listData.value = monitors
    const wrapper = mount(DNSMonitorListView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const row = wrapper.find('table').findAll('tbody tr').at(0)!
    await row.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'dns-monitor-detail',
      params: { id: 'd1' },
    })
  })
})
