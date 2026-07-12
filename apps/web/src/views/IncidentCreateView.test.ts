import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import IncidentCreateView from './IncidentCreateView.vue'
import { ApiError } from '@/api/client'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {} }),
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { createMock } = vi.hoisted(() => ({
  createMock: vi.fn(),
}))

vi.mock('@/api/incidents', async () => {
  const actual = await vi.importActual<typeof import('@/api/incidents')>('@/api/incidents')
  return { ...actual, incidentsApi: { create: createMock } }
})

const cronData = ref<{ id: string; name: string }[]>([])
const uptimeData = ref<{ id: string; name: string }[]>([])
const sslData = ref<{ id: string; name: string }[]>([])
const domainData = ref<{ id: string; name: string }[]>([])
const portData = ref<{ id: string; name: string }[]>([])
const monitorsPending = ref(false)

vi.mock('@/composables/useCronMonitors', () => ({
  useCronMonitors: () => ({ data: cronData, isPending: monitorsPending }),
}))
vi.mock('@/composables/useUptimeMonitors', () => ({
  useUptimeMonitors: () => ({ data: uptimeData, isPending: monitorsPending }),
}))
vi.mock('@/composables/useSSLMonitors', () => ({
  useSSLMonitors: () => ({ data: sslData, isPending: monitorsPending }),
}))
vi.mock('@/composables/useDomainMonitors', () => ({
  useDomainMonitors: () => ({ data: domainData, isPending: monitorsPending }),
}))
vi.mock('@/composables/usePortMonitors', () => ({
  usePortMonitors: () => ({ data: portData, isPending: monitorsPending }),
}))

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

async function selectMonitor(wrapper: ReturnType<typeof mount>, name: string) {
  const item = wrapper.findAll('li').find((li) => li.text().includes(name))
  await item!.trigger('click')
}

beforeEach(() => {
  monitorsPending.value = false
  cronData.value = []
  uptimeData.value = [{ id: 'u1', name: 'API uptime' }]
  sslData.value = []
  domainData.value = []
  portData.value = []
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('IncidentCreateView', () => {
  it('renders the form with the monitor picker', () => {
    const wrapper = mount(IncidentCreateView)

    expect(wrapper.text()).toContain('Declare incident')
    expect(wrapper.text()).toContain('API uptime')
    expect(wrapper.text()).toContain('Monitors (0 selected)')
  })

  it('shows a validation error when title is missing', async () => {
    const wrapper = mount(IncidentCreateView)

    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Title is required')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when message is missing', async () => {
    const wrapper = mount(IncidentCreateView)

    await wrapper.find('#title').setValue('Elevated latency')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('An initial message is required')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when no monitors are selected', async () => {
    const wrapper = mount(IncidentCreateView)

    await wrapper.find('#title').setValue('Elevated latency')
    await wrapper.find('#message').setValue('Investigating')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Select at least one affected monitor')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('declares the incident with the selected monitors and navigates on success', async () => {
    createMock.mockResolvedValueOnce({ id: 'i1' })
    const wrapper = mount(IncidentCreateView)

    await wrapper.find('#title').setValue('Elevated latency')
    await wrapper.find('#message').setValue('Investigating')
    await wrapper.find('#severity').setValue('major')
    await selectMonitor(wrapper, 'API uptime')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createMock).toHaveBeenCalledExactlyOnceWith({
      title: 'Elevated latency',
      message: 'Investigating',
      severity: 'major',
      monitors: [{ monitorType: 'uptime', monitorId: 'u1' }],
      confirmOverlap: false,
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'incidents' })
  })

  it('shows an inline error when creation fails for a non-overlap reason', async () => {
    createMock.mockRejectedValueOnce(new Error('Failed to declare incident'))
    const wrapper = mount(IncidentCreateView)

    await wrapper.find('#title').setValue('Elevated latency')
    await wrapper.find('#message').setValue('Investigating')
    await selectMonitor(wrapper, 'API uptime')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to declare incident')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a maintenance-overlap warning and declares anyway on confirm', async () => {
    createMock.mockRejectedValueOnce(
      new ApiError(409, 'already under active maintenance: API uptime', 'maintenance_overlap'),
    )
    createMock.mockResolvedValueOnce({ id: 'i1' })
    const wrapper = mount(IncidentCreateView)

    await wrapper.find('#title').setValue('Elevated latency')
    await wrapper.find('#message').setValue('Investigating')
    await selectMonitor(wrapper, 'API uptime')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('already under active maintenance: API uptime')
    expect(pushMock).not.toHaveBeenCalled()

    await findButtonByText(wrapper, 'Declare anyway')!.trigger('click')
    await flushPromises()

    expect(createMock).toHaveBeenLastCalledWith(expect.objectContaining({ confirmOverlap: true }))
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'incidents' })
  })

  it('dismisses the overlap warning on cancel without navigating', async () => {
    createMock.mockRejectedValueOnce(new ApiError(409, 'overlap', 'maintenance_overlap'))
    const wrapper = mount(IncidentCreateView)

    await wrapper.find('#title').setValue('Elevated latency')
    await wrapper.find('#message').setValue('Investigating')
    await selectMonitor(wrapper, 'API uptime')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(wrapper.text()).not.toContain('overlap')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the incidents list when back is clicked', async () => {
    const wrapper = mount(IncidentCreateView)

    await wrapper.find('button').trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'incidents' })
  })
})
