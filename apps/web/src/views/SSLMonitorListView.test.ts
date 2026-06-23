import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import SSLMonitorListView from './SSLMonitorListView.vue'

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

vi.mock('@/composables/useSSLMonitors', () => ({
  useSSLMonitors: () => ({
    data: listData,
    isPending: listPending,
    error: listError,
    refetch: refetchMock,
  }),
}))

const monitors = [
  {
    id: 's1',
    name: 'Production API',
    hostname: 'api.example.com',
    status: 'up' as const,
    alertsEnabled: true,
    expiresAt: new Date(Date.now() + 60 * 86400000).toISOString(),
    issuer: "Let's Encrypt",
    errorMsg: null,
    daysUntilExpiry: 60,
    lastCheckedAt: new Date(Date.now() - 5 * 60000).toISOString(),
    createdAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 's2',
    name: 'Marketing site',
    hostname: 'example.com',
    status: 'expired' as const,
    alertsEnabled: true,
    expiresAt: new Date(Date.now() - 5 * 86400000).toISOString(),
    issuer: null,
    errorMsg: null,
    daysUntilExpiry: -5,
    lastCheckedAt: null,
    createdAt: '2026-01-02T00:00:00Z',
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

describe('SSLMonitorListView', () => {
  it('shows a loading state while pending', () => {
    listPending.value = true
    const wrapper = mount(SSLMonitorListView)

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message and retries on click', async () => {
    listError.value = { message: 'Failed to load monitors' }
    const wrapper = mount(SSLMonitorListView)

    expect(wrapper.text()).toContain('Failed to load monitors')
    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Try again')!
      .trigger('click')

    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an empty state when there are no monitors', () => {
    const wrapper = mount(SSLMonitorListView)

    expect(wrapper.text()).toContain('No SSL monitors yet. Add one to track certificate expiry.')
  })

  it('navigates to the create view when "Add your first monitor" is clicked', async () => {
    const wrapper = mount(SSLMonitorListView)

    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Add your first monitor')!
      .trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'ssl-monitor-create' })
  })

  it('navigates to the create view when the header "Add monitor" button is clicked', async () => {
    const wrapper = mount(SSLMonitorListView)

    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Add monitor')!
      .trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'ssl-monitor-create' })
  })

  it('renders monitor rows with status, expiry, and issuer', () => {
    listData.value = monitors
    const wrapper = mount(SSLMonitorListView)

    expect(wrapper.text()).toContain('Production API')
    expect(wrapper.text()).toContain('api.example.com')
    expect(wrapper.text()).toContain('Valid')
    expect(wrapper.text()).toContain('60d')
    expect(wrapper.text()).toContain('Marketing site')
    expect(wrapper.text()).toContain('Expired')
  })

  it('formats a missing last-checked time as a dash', () => {
    listData.value = monitors
    const wrapper = mount(SSLMonitorListView)

    const desktopRows = wrapper.find('table').findAll('tbody tr')
    const expiredRow = desktopRows.find((r) => r.text().includes('Marketing site'))!
    expect(expiredRow.text()).toContain('—')
  })

  it('shows "Expired" for a negative days-until-expiry even with an expiry date', () => {
    listData.value = [monitors[1]]
    const wrapper = mount(SSLMonitorListView)

    const cells = wrapper.find('table').findAll('tbody tr td')
    expect(cells.some((c) => c.text() === 'Expired')).toBe(true)
  })

  it('navigates to the detail view when a row is clicked', async () => {
    listData.value = monitors
    const wrapper = mount(SSLMonitorListView)

    const row = wrapper.find('table').findAll('tbody tr').at(0)!
    await row.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'ssl-monitor-detail',
      params: { id: 's1' },
    })
  })
})
