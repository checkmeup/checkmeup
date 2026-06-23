import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import DomainMonitorListView from './DomainMonitorListView.vue'

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

vi.mock('@/composables/useDomainMonitors', () => ({
  useDomainMonitors: () => ({
    data: listData,
    isPending: pending,
    error: queryError,
    refetch: refetchMock,
  }),
}))

const monitors = [
  {
    id: 'd1',
    name: 'Production domain',
    domain: 'example.com',
    status: 'up' as const,
    alertsEnabled: true,
    expiresAt: '2027-01-01T00:00:00Z',
    registrar: 'Namecheap',
    errorMsg: null,
    daysUntilExpiry: 190,
    lastCheckedAt: new Date(Date.now() - 5 * 60000).toISOString(),
    createdAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 'd2',
    name: 'Marketing site',
    domain: 'example.org',
    status: 'expiring_soon' as const,
    alertsEnabled: true,
    expiresAt: '2026-07-01T00:00:00Z',
    registrar: 'GoDaddy',
    errorMsg: null,
    daysUntilExpiry: 8,
    lastCheckedAt: new Date(Date.now() - 3600000).toISOString(),
    createdAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 'd3',
    name: 'Expired domain',
    domain: 'example.net',
    status: 'expired' as const,
    alertsEnabled: true,
    expiresAt: '2026-01-01T00:00:00Z',
    registrar: null,
    errorMsg: null,
    daysUntilExpiry: -5,
    lastCheckedAt: null,
    createdAt: '2026-01-01T00:00:00Z',
  },
]

beforeEach(() => {
  listData.value = null
  pending.value = false
  queryError.value = null
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('DomainMonitorListView', () => {
  it('shows a loading state while pending', () => {
    pending.value = true
    const wrapper = mount(DomainMonitorListView)

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message and retries on click when the query fails', async () => {
    queryError.value = { message: 'Failed to load monitors' }
    const wrapper = mount(DomainMonitorListView)

    expect(wrapper.text()).toContain('Failed to load monitors')
    const retryButton = wrapper.findAll('button').find((b) => b.text() === 'Try again')
    await retryButton!.trigger('click')

    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an empty state with a call to action when there are no monitors', () => {
    listData.value = []
    const wrapper = mount(DomainMonitorListView)

    expect(wrapper.text()).toContain(
      'No domain monitors yet. Add one to track registration expiry.',
    )
    const addButton = wrapper.findAll('button').find((b) => b.text() === 'Add your first monitor')
    expect(addButton).toBeTruthy()
  })

  it('navigates to the create view when add monitor is clicked', async () => {
    listData.value = []
    const wrapper = mount(DomainMonitorListView)

    const addButton = wrapper.findAll('button').find((b) => b.text() === 'Add monitor')
    await addButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'domain-monitor-create' })
  })

  it('renders monitors with status, expiry, and last-checked formatting', () => {
    listData.value = [...monitors]
    const wrapper = mount(DomainMonitorListView)

    expect(wrapper.text()).toContain('Production domain')
    expect(wrapper.text()).toContain('example.com')
    expect(wrapper.text()).toContain('Valid')
    expect(wrapper.text()).toContain('190d')

    expect(wrapper.text()).toContain('Marketing site')
    expect(wrapper.text()).toContain('Expiring soon')
    expect(wrapper.text()).toContain('8d')

    expect(wrapper.text()).toContain('Expired domain')
    expect(wrapper.text()).toContain('Expired')
  })

  it('shows a dash for last checked when a monitor has never been checked', () => {
    listData.value = [monitors[2]]
    const wrapper = mount(DomainMonitorListView)

    const row = wrapper.findAll('table tbody tr').at(0)!
    expect(row.text()).toContain('—')
  })

  it('navigates to the detail view when a table row is clicked', async () => {
    listData.value = [...monitors]
    const wrapper = mount(DomainMonitorListView)

    const row = wrapper
      .findAll('table tbody tr')
      .find((r) => r.text().includes('Production domain'))
    await row!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'domain-monitor-detail',
      params: { id: 'd1' },
    })
  })

  it('navigates to the detail view when a mobile card is clicked', async () => {
    listData.value = [...monitors]
    const wrapper = mount(DomainMonitorListView)

    const card = wrapper
      .findAll('.md\\:hidden.space-y-2 > div')
      .find((c) => c.text().includes('Marketing site'))
    await card!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'domain-monitor-detail',
      params: { id: 'd2' },
    })
  })
})
