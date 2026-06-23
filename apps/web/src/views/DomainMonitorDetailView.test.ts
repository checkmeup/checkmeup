import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import DomainMonitorDetailView from './DomainMonitorDetailView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'd1' } }),
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { pauseDomainMock, resumeDomainMock, deleteDomainMock } = vi.hoisted(() => ({
  pauseDomainMock: vi.fn(),
  resumeDomainMock: vi.fn(),
  deleteDomainMock: vi.fn(),
}))

vi.mock('@/api/monitors', () => ({
  monitorsApi: {
    pauseDomain: pauseDomainMock,
    resumeDomain: resumeDomainMock,
    deleteDomain: deleteDomainMock,
  },
}))

const refetchMock = vi.fn()
const detailData = ref<unknown>(null)
const pending = ref(false)
const queryError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useDomainMonitors', () => ({
  useDomainMonitor: () => ({
    data: detailData,
    isPending: pending,
    error: queryError,
    refetch: refetchMock,
  }),
}))

const monitor = {
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
  channelIds: [],
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

describe('DomainMonitorDetailView', () => {
  it('shows a loading state while pending', () => {
    pending.value = true
    const wrapper = mount(DomainMonitorDetailView)

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message when the query fails', () => {
    queryError.value = { message: 'Monitor not found' }
    const wrapper = mount(DomainMonitorDetailView)

    expect(wrapper.text()).toContain('Monitor not found')
  })

  it('renders monitor header and expiry details', () => {
    detailData.value = { ...monitor }
    const wrapper = mount(DomainMonitorDetailView)

    expect(wrapper.text()).toContain('Production domain')
    expect(wrapper.text()).toContain('Valid')
    expect(wrapper.text()).toContain('example.com')
    expect(wrapper.text()).toContain('190 days')
    expect(wrapper.text()).toContain('Namecheap')
    expect(wrapper.text()).toContain('Last checked')
  })

  it('shows an expiring soon status with degraded styling', () => {
    detailData.value = { ...monitor, status: 'expiring_soon', daysUntilExpiry: 20 }
    const wrapper = mount(DomainMonitorDetailView)

    expect(wrapper.text()).toContain('Expiring soon')
    expect(wrapper.text()).toContain('20 days')
  })

  it('shows an expired status', () => {
    detailData.value = { ...monitor, status: 'expired', daysUntilExpiry: -5 }
    const wrapper = mount(DomainMonitorDetailView)

    expect(wrapper.text()).toContain('Expired')
    expect(wrapper.text()).toContain('-5 days')
  })

  it('shows a dash for days remaining when daysUntilExpiry is null', () => {
    detailData.value = { ...monitor, expiresAt: null, daysUntilExpiry: null }
    const wrapper = mount(DomainMonitorDetailView)

    const cells = wrapper.findAll('.grid.grid-cols-2 > div')
    const daysCell = cells.find((c) => c.text().includes('Days remaining'))
    expect(daysCell!.text()).toContain('—')
    const expiresCell = cells.find((c) => c.text().includes('Expires'))
    expect(expiresCell!.text()).toContain('—')
  })

  it('shows a dash for registrar when missing', () => {
    detailData.value = { ...monitor, registrar: null }
    const wrapper = mount(DomainMonitorDetailView)

    const cells = wrapper.findAll('.grid.grid-cols-2 > div')
    const registrarCell = cells.find((c) => c.text().includes('Registrar'))
    expect(registrarCell!.text()).toContain('—')
  })

  it('renders the error message banner when errorMsg is present', () => {
    detailData.value = { ...monitor, errorMsg: 'RDAP lookup failed' }
    const wrapper = mount(DomainMonitorDetailView)

    expect(wrapper.text()).toContain('RDAP lookup failed')
  })

  it('shows a first-check notice when never checked', () => {
    detailData.value = { ...monitor, lastCheckedAt: null }
    const wrapper = mount(DomainMonitorDetailView)

    expect(wrapper.text()).toContain('First check runs within 24 hours of creation.')
  })

  it('navigates to the edit view when Edit is clicked', async () => {
    detailData.value = { ...monitor }
    const wrapper = mount(DomainMonitorDetailView)

    await findButtonByText(wrapper, 'Edit')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'domain-monitor-edit',
      params: { id: 'd1' },
    })
  })

  it('pauses an active monitor and refetches', async () => {
    detailData.value = { ...monitor, status: 'up' }
    pauseDomainMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(DomainMonitorDetailView)

    await findButtonByText(wrapper, 'Pause')!.trigger('click')
    await flushPromises()

    expect(pauseDomainMock).toHaveBeenCalledExactlyOnceWith('d1')
    expect(resumeDomainMock).not.toHaveBeenCalled()
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('resumes a paused monitor and refetches', async () => {
    detailData.value = { ...monitor, status: 'paused' }
    resumeDomainMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(DomainMonitorDetailView)

    await findButtonByText(wrapper, 'Resume')!.trigger('click')
    await flushPromises()

    expect(resumeDomainMock).toHaveBeenCalledExactlyOnceWith('d1')
    expect(pauseDomainMock).not.toHaveBeenCalled()
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an error and does not refetch when toggling pause fails', async () => {
    detailData.value = { ...monitor, status: 'up' }
    pauseDomainMock.mockRejectedValueOnce(new Error('Action failed'))
    const wrapper = mount(DomainMonitorDetailView)

    await findButtonByText(wrapper, 'Pause')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Action failed')
    expect(refetchMock).not.toHaveBeenCalled()
  })

  it('cancels the delete confirmation', async () => {
    detailData.value = { ...monitor }
    const wrapper = mount(DomainMonitorDetailView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    expect(findButtonByText(wrapper, 'Confirm delete')).toBeTruthy()

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')
    expect(findButtonByText(wrapper, 'Delete')).toBeTruthy()
    expect(deleteDomainMock).not.toHaveBeenCalled()
  })

  it('deletes the monitor and navigates back to the list on confirm', async () => {
    detailData.value = { ...monitor }
    deleteDomainMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(DomainMonitorDetailView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(deleteDomainMock).toHaveBeenCalledExactlyOnceWith('d1')
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'domain-monitors' })
  })

  it('shows an error and resets confirmation state when delete fails', async () => {
    detailData.value = { ...monitor }
    deleteDomainMock.mockRejectedValueOnce(new Error('Delete failed'))
    const wrapper = mount(DomainMonitorDetailView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Delete failed')
    expect(pushMock).not.toHaveBeenCalled()
    expect(findButtonByText(wrapper, 'Delete')).toBeTruthy()
  })
})
