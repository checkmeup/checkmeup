import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import MaintenanceWindowEditView from './MaintenanceWindowEditView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'mw1' } }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { updateMock, deleteMock } = vi.hoisted(() => ({
  updateMock: vi.fn(),
  deleteMock: vi.fn(),
}))

vi.mock('@/api/maintenance', () => ({
  maintenanceApi: { update: updateMock, delete: deleteMock },
}))

const winData = ref<unknown>(null)
const winPending = ref(false)
const loadError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useMaintenanceWindows', () => ({
  useMaintenanceWindow: () => ({ data: winData, isPending: winPending, error: loadError }),
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

const win = {
  id: 'mw1',
  title: 'DB migration',
  message: 'Expect brief downtime',
  startsAt: '2026-07-01T10:00:00.000Z',
  endsAt: '2026-07-01T12:00:00.000Z',
  status: 'upcoming' as const,
  monitors: [{ monitorType: 'cron' as const, monitorId: 'c1', name: 'Nightly backup' }],
  monitorCount: 1,
  createdAt: '2026-06-01T00:00:00Z',
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  winData.value = null
  winPending.value = false
  loadError.value = null
  monitorsPending.value = false
  cronData.value = [
    { id: 'c1', name: 'Nightly backup' },
    { id: 'c2', name: 'Weekly cleanup' },
  ]
  uptimeData.value = []
  sslData.value = []
  domainData.value = []
  portData.value = []
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('MaintenanceWindowEditView', () => {
  it('shows a loading state while pending', () => {
    winPending.value = true
    const wrapper = mount(MaintenanceWindowEditView)

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an inline error when loading fails', async () => {
    const wrapper = mount(MaintenanceWindowEditView)
    loadError.value = { message: 'Maintenance window not found' }
    await flushPromises()

    expect(wrapper.text()).toContain('Maintenance window not found')
  })

  it('populates the form from the loaded window, including selected monitors', () => {
    winData.value = { ...win }
    const wrapper = mount(MaintenanceWindowEditView)

    expect((wrapper.find('#title').element as HTMLInputElement).value).toBe('DB migration')
    expect((wrapper.find('#message').element as HTMLInputElement).value).toBe(
      'Expect brief downtime',
    )
    expect(wrapper.text()).toContain('Monitors (1 selected)')
  })

  it('checks "no end date" and disables the ends field when endsAt is null', () => {
    winData.value = { ...win, endsAt: null }
    const wrapper = mount(MaintenanceWindowEditView)

    const noEndCheckbox = wrapper.find('input[type="checkbox"]')
    expect((noEndCheckbox.element as HTMLInputElement).checked).toBe(true)
    expect((wrapper.find('#endsAt').element as HTMLInputElement).disabled).toBe(true)
  })

  it('shows a validation error when title is cleared', async () => {
    winData.value = { ...win }
    const wrapper = mount(MaintenanceWindowEditView)

    await wrapper.find('#title').setValue('')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Title is required')
    expect(updateMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when all monitors are deselected', async () => {
    winData.value = { ...win }
    const wrapper = mount(MaintenanceWindowEditView)

    const item = wrapper.findAll('li').find((li) => li.text().includes('Nightly backup'))
    await item!.trigger('click')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Select at least one monitor')
    expect(updateMock).not.toHaveBeenCalled()
  })

  it('updates the window and navigates to the list on success', async () => {
    winData.value = { ...win }
    updateMock.mockResolvedValueOnce({ ...win })
    const wrapper = mount(MaintenanceWindowEditView)

    await wrapper.find('#title').setValue('DB migration v2')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(updateMock).toHaveBeenCalledExactlyOnceWith('mw1', {
      title: 'DB migration v2',
      message: 'Expect brief downtime',
      startsAt: win.startsAt,
      endsAt: win.endsAt,
      monitors: [{ monitorType: 'cron', monitorId: 'c1' }],
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'maintenance' })
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    winData.value = { ...win }
    const { ApiError } = await import('@/api/client')
    updateMock.mockRejectedValueOnce(
      new ApiError(402, 'Upgrade to add more monitors', 'plan_limit_reached'),
    )
    const wrapper = mount(MaintenanceWindowEditView)

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to add more monitors')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic inline error when saving fails for another reason', async () => {
    winData.value = { ...win }
    updateMock.mockRejectedValueOnce(new Error('Save failed'))
    const wrapper = mount(MaintenanceWindowEditView)

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Save failed')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('cancels the delete confirmation', async () => {
    winData.value = { ...win }
    const wrapper = mount(MaintenanceWindowEditView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    expect(findButtonByText(wrapper, 'Confirm delete')).toBeTruthy()

    const cancelButtons = wrapper.findAll('button').filter((b) => b.text() === 'Cancel')
    await cancelButtons.at(-1)!.trigger('click')
    expect(findButtonByText(wrapper, 'Delete')).toBeTruthy()
    expect(deleteMock).not.toHaveBeenCalled()
  })

  it('deletes the window and navigates back to the list on confirm', async () => {
    winData.value = { ...win }
    deleteMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(MaintenanceWindowEditView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('mw1')
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'maintenance' })
  })

  it('shows an error and resets confirmation state when delete fails', async () => {
    winData.value = { ...win }
    deleteMock.mockRejectedValueOnce(new Error('Delete failed'))
    const wrapper = mount(MaintenanceWindowEditView)

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Delete failed')
    expect(pushMock).not.toHaveBeenCalled()
    expect(findButtonByText(wrapper, 'Delete')).toBeTruthy()
  })
})
