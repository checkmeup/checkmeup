import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import IncidentListView from './IncidentListView.vue'

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

vi.mock('@/composables/useIncidents', () => ({
  useIncidents: () => ({
    data: listData,
    isPending: listPending,
    error: listError,
    refetch: refetchMock,
  }),
}))

const incidents = [
  {
    id: 'i1',
    title: 'Elevated latency',
    severity: 'major' as const,
    status: 'investigating' as const,
    monitorCount: 2,
    createdAt: '2026-06-01T00:00:00Z',
    updatedAt: '2026-06-01T00:00:00Z',
    resolvedAt: null,
  },
  {
    id: 'i2',
    title: 'Resolved outage',
    severity: 'critical' as const,
    status: 'resolved' as const,
    monitorCount: 1,
    createdAt: '2026-05-01T00:00:00Z',
    updatedAt: '2026-05-01T02:00:00Z',
    resolvedAt: '2026-05-01T02:00:00Z',
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

describe('IncidentListView', () => {
  it('shows a loading state while pending', () => {
    listPending.value = true
    const wrapper = mount(IncidentListView)

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message and retries on click', async () => {
    listError.value = { message: 'Failed to load incidents' }
    const wrapper = mount(IncidentListView)

    expect(wrapper.text()).toContain('Failed to load incidents')
    await findButtonByText(wrapper, 'Try again')!.trigger('click')

    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an empty state when there are no incidents', () => {
    const wrapper = mount(IncidentListView)

    expect(wrapper.text()).toContain('No incidents yet.')
  })

  it('navigates to the create view when "Declare your first incident" is clicked', async () => {
    const wrapper = mount(IncidentListView)

    await findButtonByText(wrapper, 'Declare your first incident')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'incidents-create' })
  })

  it('navigates to the create view when the header "Declare incident" button is clicked', async () => {
    const wrapper = mount(IncidentListView)

    await findButtonByText(wrapper, 'Declare incident')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'incidents-create' })
  })

  it('renders incident rows with title, severity, status, and monitor count', () => {
    listData.value = incidents
    const wrapper = mount(IncidentListView)

    expect(wrapper.text()).toContain('Elevated latency')
    expect(wrapper.text()).toContain('Major')
    expect(wrapper.text()).toContain('Investigating')
    expect(wrapper.text()).toContain('Resolved outage')
    expect(wrapper.text()).toContain('Critical')
    expect(wrapper.text()).toContain('Resolved')
  })

  it('navigates to the detail view when a row is clicked', async () => {
    listData.value = incidents
    const wrapper = mount(IncidentListView)

    const row = wrapper.find('table').findAll('tbody tr').at(0)!
    await row.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'incident-detail',
      params: { id: 'i1' },
    })
  })
})
