import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import MaintenanceWindowCreateView from './MaintenanceWindowCreateView.vue'

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

vi.mock('@/api/maintenance', () => ({
  maintenanceApi: { create: createMock },
}))

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

beforeEach(() => {
  monitorsPending.value = false
  cronData.value = [{ id: 'c1', name: 'Nightly backup' }]
  uptimeData.value = [{ id: 'u1', name: 'API uptime' }]
  sslData.value = []
  domainData.value = []
  portData.value = []
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('MaintenanceWindowCreateView', () => {
  it('renders the form with the monitor picker', () => {
    const wrapper = mount(MaintenanceWindowCreateView)

    expect(wrapper.text()).toContain('Schedule maintenance')
    expect(wrapper.text()).toContain('Nightly backup')
    expect(wrapper.text()).toContain('API uptime')
    expect(wrapper.text()).toContain('Monitors (0 selected)')
  })

  it('shows a validation error when title is missing', async () => {
    const wrapper = mount(MaintenanceWindowCreateView)

    await wrapper.find('#startsAt').setValue('2026-07-01T10:00')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Title is required')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when start time is missing', async () => {
    const wrapper = mount(MaintenanceWindowCreateView)

    await wrapper.find('#title').setValue('DB migration')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Start time is required')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when end time is missing and "no end" is unchecked', async () => {
    const wrapper = mount(MaintenanceWindowCreateView)

    await wrapper.find('#title').setValue('DB migration')
    await wrapper.find('#startsAt').setValue('2026-07-01T10:00')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('End time is required, or check "no end date"')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('does not require an end time when "no end date" is checked', async () => {
    createMock.mockResolvedValueOnce({ id: 'm1' })
    const wrapper = mount(MaintenanceWindowCreateView)

    await wrapper.find('#title').setValue('DB migration')
    await wrapper.find('#startsAt').setValue('2026-07-01T10:00')
    await wrapper.find('input[type="checkbox"]').setValue(true)
    const item = wrapper.findAll('li').find((li) => li.text().includes('Nightly backup'))
    await item!.trigger('click')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createMock).toHaveBeenCalledOnce()
    const payload = createMock.mock.calls[0]![0]
    expect(payload.endsAt).toBeNull()
  })

  it('shows a validation error when no monitors are selected', async () => {
    const wrapper = mount(MaintenanceWindowCreateView)

    await wrapper.find('#title').setValue('DB migration')
    await wrapper.find('#startsAt').setValue('2026-07-01T10:00')
    await wrapper.find('#endsAt').setValue('2026-07-01T12:00')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Select at least one monitor')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('toggles monitor selection and reflects the selected count', async () => {
    const wrapper = mount(MaintenanceWindowCreateView)

    const item = wrapper.findAll('li').find((li) => li.text().includes('Nightly backup'))
    await item!.trigger('click')

    expect(wrapper.text()).toContain('Monitors (1 selected)')

    await item!.trigger('click')

    expect(wrapper.text()).toContain('Monitors (0 selected)')
  })

  it('creates the maintenance window with selected monitors and navigates on success', async () => {
    createMock.mockResolvedValueOnce({ id: 'm1' })
    const wrapper = mount(MaintenanceWindowCreateView)

    await wrapper.find('#title').setValue('DB migration')
    await wrapper.find('#message').setValue('Expect brief downtime')
    await wrapper.find('#startsAt').setValue('2026-07-01T10:00')
    await wrapper.find('#endsAt').setValue('2026-07-01T12:00')
    const item = wrapper.findAll('li').find((li) => li.text().includes('Nightly backup'))
    await item!.trigger('click')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createMock).toHaveBeenCalledExactlyOnceWith({
      title: 'DB migration',
      message: 'Expect brief downtime',
      startsAt: new Date('2026-07-01T10:00').toISOString(),
      endsAt: new Date('2026-07-01T12:00').toISOString(),
      monitors: [{ monitorType: 'cron', monitorId: 'c1' }],
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'maintenance' })
  })

  it('shows an inline error when creation fails', async () => {
    createMock.mockRejectedValueOnce(new Error('Failed to create maintenance window'))
    const wrapper = mount(MaintenanceWindowCreateView)

    await wrapper.find('#title').setValue('DB migration')
    await wrapper.find('#startsAt').setValue('2026-07-01T10:00')
    await wrapper.find('#endsAt').setValue('2026-07-01T12:00')
    const item = wrapper.findAll('li').find((li) => li.text().includes('Nightly backup'))
    await item!.trigger('click')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to create maintenance window')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the maintenance list when cancel or back is clicked', async () => {
    const wrapper = mount(MaintenanceWindowCreateView)

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'maintenance' })
  })
})
