import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import SSLMonitorDetailView from './SSLMonitorDetailView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 's1' } }),
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { pauseSSLMock, resumeSSLMock, deleteSSLMock } = vi.hoisted(() => ({
  pauseSSLMock: vi.fn(),
  resumeSSLMock: vi.fn(),
  deleteSSLMock: vi.fn(),
}))

vi.mock('@/api/monitors', () => ({
  monitorsApi: {
    pauseSSL: pauseSSLMock,
    resumeSSL: resumeSSLMock,
    deleteSSL: deleteSSLMock,
  },
}))

const refetchMock = vi.fn()
const detailData = ref<unknown>(null)
const pending = ref(false)
const queryError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useSSLMonitors', () => ({
  useSSLMonitor: () => ({
    data: detailData,
    isPending: pending,
    error: queryError,
    refetch: refetchMock,
  }),
}))

const monitor = {
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

describe('SSLMonitorDetailView', () => {
  it('shows a loading state while pending', () => {
    pending.value = true
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message when the query fails', () => {
    queryError.value = { message: 'Monitor not found' }
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Monitor not found')
  })

  it('renders monitor header details', () => {
    detailData.value = { ...monitor }
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Production API')
    expect(wrapper.text()).toContain('Valid')
    expect(wrapper.text()).toContain('api.example.com')
    expect(wrapper.text()).toContain('60 days')
    expect(wrapper.text()).toContain("Let's Encrypt")
    expect(wrapper.text()).toContain('5m ago')
  })

  it('shows a dash for issuer and days remaining when unavailable', () => {
    detailData.value = { ...monitor, issuer: null, daysUntilExpiry: null, expiresAt: null }
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const cells = wrapper.findAll('.grid.grid-cols-2 > div')
    const issuerCell = cells.find((c) => c.text().includes('Issuer'))
    expect(issuerCell!.text()).toContain('—')
    const daysCell = cells.find((c) => c.text().includes('Days remaining'))
    expect(daysCell!.text()).toContain('—')
  })

  it('shows the error message block when the monitor has an error', () => {
    detailData.value = { ...monitor, status: 'error', errorMsg: 'TLS handshake failed' }
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('TLS handshake failed')
  })

  it('shows the first-check notice when the monitor has never been checked', () => {
    detailData.value = { ...monitor, lastCheckedAt: null }
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('First check runs within 24 hours of creation.')
  })

  it('navigates to the edit view when Edit is clicked', async () => {
    detailData.value = { ...monitor }
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Edit')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'ssl-monitor-edit',
      params: { id: 's1' },
    })
  })

  it('pauses an active monitor and refetches', async () => {
    detailData.value = { ...monitor, status: 'up' }
    pauseSSLMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Pause')!.trigger('click')
    await flushPromises()

    expect(pauseSSLMock).toHaveBeenCalledExactlyOnceWith('s1')
    expect(resumeSSLMock).not.toHaveBeenCalled()
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('resumes a paused monitor and refetches', async () => {
    detailData.value = { ...monitor, status: 'paused' }
    resumeSSLMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Resume')!.trigger('click')
    await flushPromises()

    expect(resumeSSLMock).toHaveBeenCalledExactlyOnceWith('s1')
    expect(pauseSSLMock).not.toHaveBeenCalled()
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an error and does not refetch when toggling pause fails', async () => {
    detailData.value = { ...monitor, status: 'up' }
    pauseSSLMock.mockRejectedValueOnce(new Error('Action failed'))
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Pause')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Action failed')
    expect(refetchMock).not.toHaveBeenCalled()
  })

  it('cancels the delete confirmation', async () => {
    detailData.value = { ...monitor }
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    expect(findButtonByText(wrapper, 'Confirm delete')).toBeTruthy()

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')
    expect(findButtonByText(wrapper, 'Delete')).toBeTruthy()
    expect(deleteSSLMock).not.toHaveBeenCalled()
  })

  it('deletes the monitor and navigates back to the list on confirm', async () => {
    detailData.value = { ...monitor }
    deleteSSLMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(deleteSSLMock).toHaveBeenCalledExactlyOnceWith('s1')
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'ssl-monitors' })
  })

  it('shows an error and resets confirmation state when delete fails', async () => {
    detailData.value = { ...monitor }
    deleteSSLMock.mockRejectedValueOnce(new Error('Delete failed'))
    const wrapper = mount(SSLMonitorDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Delete failed')
    expect(pushMock).not.toHaveBeenCalled()
    expect(findButtonByText(wrapper, 'Delete')).toBeTruthy()
  })
})
