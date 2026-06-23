import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import SSLMonitorEditView from './SSLMonitorEditView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 's1' } }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { updateSSLMock } = vi.hoisted(() => ({
  updateSSLMock: vi.fn(),
}))

vi.mock('@/api/monitors', () => ({
  monitorsApi: { updateSSL: updateSSLMock },
}))

const detailData = ref<unknown>(null)
const monitorPending = ref(false)
const loadError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useSSLMonitors', () => ({
  useSSLMonitor: () => ({ data: detailData, isPending: monitorPending, error: loadError }),
}))

const channelsData = ref<{ id: string; name: string; type: string; enabled: boolean }[]>([])
const channelsPending = ref(false)

vi.mock('@/composables/useNotificationChannels', () => ({
  useNotificationChannels: () => ({ data: channelsData, isPending: channelsPending }),
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
  channelIds: [] as string[],
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  detailData.value = null
  monitorPending.value = false
  loadError.value = null
  channelsData.value = []
  channelsPending.value = false
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('SSLMonitorEditView', () => {
  it('shows a loading state while the monitor is pending', () => {
    monitorPending.value = true
    const wrapper = mount(SSLMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Loading…')
  })

  it('populates the form from the loaded monitor', () => {
    detailData.value = { ...monitor }
    const wrapper = mount(SSLMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect((wrapper.find('#name').element as HTMLInputElement).value).toBe('Production API')
    expect(wrapper.text()).toContain('api.example.com')
    expect(wrapper.text()).toContain(
      'To change the domain, delete this monitor and create a new one.',
    )
  })

  it('shows an inline error when loading the monitor fails', async () => {
    const wrapper = mount(SSLMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })
    loadError.value = { message: 'Monitor not found' }
    await flushPromises()

    expect(wrapper.text()).toContain('Monitor not found')
  })

  it('shows a validation error when name is cleared', async () => {
    detailData.value = { ...monitor }
    const wrapper = mount(SSLMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Name is required')
    expect(updateSSLMock).not.toHaveBeenCalled()
  })

  it('updates the monitor and navigates to its detail page on success', async () => {
    detailData.value = { ...monitor }
    updateSSLMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(SSLMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Renamed API')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(updateSSLMock).toHaveBeenCalledExactlyOnceWith('s1', {
      name: 'Renamed API',
      hostname: 'api.example.com',
      alertsEnabled: true,
      channelIds: [],
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'ssl-monitor-detail',
      params: { id: 's1' },
    })
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    detailData.value = { ...monitor }
    const { ApiError } = await import('@/api/client')
    updateSSLMock.mockRejectedValueOnce(
      new ApiError(402, 'Upgrade to enable alerts', 'plan_limit_reached'),
    )
    const wrapper = mount(SSLMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to enable alerts')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic inline error when saving fails for another reason', async () => {
    detailData.value = { ...monitor }
    updateSSLMock.mockRejectedValueOnce(new Error('Save failed'))
    const wrapper = mount(SSLMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Save failed')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the detail page when cancel is clicked', async () => {
    detailData.value = { ...monitor }
    const wrapper = mount(SSLMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'ssl-monitor-detail',
      params: { id: 's1' },
    })
  })

  it('navigates back to the detail page when the back link is clicked', async () => {
    detailData.value = { ...monitor }
    const wrapper = mount(SSLMonitorEditView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, '← Back')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'ssl-monitor-detail',
      params: { id: 's1' },
    })
  })
})
