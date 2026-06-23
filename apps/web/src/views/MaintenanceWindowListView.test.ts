import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import MaintenanceWindowListView from './MaintenanceWindowListView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {} }),
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const setQueryDataMock = vi.fn()

vi.mock('@tanstack/vue-query', async () => {
  const actual = await vi.importActual<typeof import('@tanstack/vue-query')>('@tanstack/vue-query')
  return {
    ...actual,
    useQueryClient: () => ({ setQueryData: setQueryDataMock }),
  }
})

const { endNowMock } = vi.hoisted(() => ({
  endNowMock: vi.fn(),
}))

vi.mock('@/api/maintenance', () => ({
  maintenanceApi: { endNow: endNowMock },
}))

const refetchMock = vi.fn()
const listData = ref<unknown[]>([])
const listPending = ref(false)
const listError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useMaintenanceWindows', () => ({
  useMaintenanceWindows: () => ({
    data: listData,
    isPending: listPending,
    error: listError,
    refetch: refetchMock,
  }),
}))

const windows = [
  {
    id: 'mw1',
    title: 'DB migration',
    message: '',
    startsAt: '2026-07-01T10:00:00.000Z',
    endsAt: '2026-07-01T12:00:00.000Z',
    status: 'upcoming' as const,
    monitorCount: 2,
    createdAt: '2026-06-01T00:00:00Z',
  },
  {
    id: 'mw2',
    title: 'Already ended',
    message: '',
    startsAt: '2026-05-01T10:00:00.000Z',
    endsAt: null,
    status: 'ended' as const,
    monitorCount: 1,
    createdAt: '2026-04-01T00:00:00Z',
  },
]

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  listData.value = []
  listPending.value = false
  listError.value = null
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('MaintenanceWindowListView', () => {
  it('shows a loading state while pending', () => {
    listPending.value = true
    const wrapper = mount(MaintenanceWindowListView)

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message and retries on click', async () => {
    listError.value = { message: 'Failed to load maintenance windows' }
    const wrapper = mount(MaintenanceWindowListView)

    expect(wrapper.text()).toContain('Failed to load maintenance windows')
    await findButtonByText(wrapper, 'Try again')!.trigger('click')

    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an empty state when there are no windows', () => {
    const wrapper = mount(MaintenanceWindowListView)

    expect(wrapper.text()).toContain(
      'No maintenance windows yet. Schedule one to suppress alerts during planned downtime.',
    )
  })

  it('navigates to the create view when "Schedule your first window" is clicked', async () => {
    const wrapper = mount(MaintenanceWindowListView)

    await findButtonByText(wrapper, 'Schedule your first window')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'maintenance-create' })
  })

  it('navigates to the create view when the header "Schedule maintenance" button is clicked', async () => {
    const wrapper = mount(MaintenanceWindowListView)

    await findButtonByText(wrapper, 'Schedule maintenance')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'maintenance-create' })
  })

  it('renders window rows with status, dates, and monitor counts', () => {
    listData.value = windows
    const wrapper = mount(MaintenanceWindowListView)

    expect(wrapper.text()).toContain('DB migration')
    expect(wrapper.text()).toContain('Upcoming')
    expect(wrapper.text()).toContain('Already ended')
    expect(wrapper.text()).toContain('Ended')
    expect(wrapper.text()).toContain('until ended manually')
  })

  it('hides the "End now" button for ended windows', () => {
    listData.value = windows
    const wrapper = mount(MaintenanceWindowListView)

    const desktopRows = wrapper.find('table').findAll('tbody tr')
    const endedRow = desktopRows.find((r) => r.text().includes('Already ended'))!
    expect(endedRow.findAll('button')).toHaveLength(0)

    const upcomingRow = desktopRows.find((r) => r.text().includes('DB migration'))!
    expect(upcomingRow.findAll('button')).toHaveLength(1)
  })

  it('navigates to the edit view when a row is clicked', async () => {
    listData.value = windows
    const wrapper = mount(MaintenanceWindowListView)

    const row = wrapper.find('table').findAll('tbody tr').at(0)!
    await row.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'maintenance-edit',
      params: { id: 'mw1' },
    })
  })

  it('ends a window now and updates the cached list on success', async () => {
    listData.value = windows
    const updated = { ...windows[0], status: 'ended' as const }
    endNowMock.mockResolvedValueOnce(updated)
    const wrapper = mount(MaintenanceWindowListView)

    const row = wrapper.find('table').findAll('tbody tr').at(0)!
    await row.find('button').trigger('click')
    await flushPromises()

    expect(endNowMock).toHaveBeenCalledExactlyOnceWith('mw1')
    expect(setQueryDataMock).toHaveBeenCalledExactlyOnceWith(
      ['maintenance-windows'],
      expect.any(Function),
    )
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('does not navigate to the edit view when "End now" is clicked', async () => {
    listData.value = windows
    endNowMock.mockResolvedValueOnce({ ...windows[0] })
    const wrapper = mount(MaintenanceWindowListView)

    const row = wrapper.find('table').findAll('tbody tr').at(0)!
    await row.find('button').trigger('click')
    await flushPromises()

    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows an inline error when ending a window fails', async () => {
    listData.value = windows
    endNowMock.mockRejectedValueOnce(new Error('Failed to end maintenance window'))
    const wrapper = mount(MaintenanceWindowListView)

    const row = wrapper.find('table').findAll('tbody tr').at(0)!
    await row.find('button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to end maintenance window')
  })
})
