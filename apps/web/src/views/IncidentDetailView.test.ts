import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import IncidentDetailView from './IncidentDetailView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'i1' } }),
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

const { updateTitleMock, postUpdateMock, updateUpdateMessageMock, deleteMock } = vi.hoisted(() => ({
  updateTitleMock: vi.fn(),
  postUpdateMock: vi.fn(),
  updateUpdateMessageMock: vi.fn(),
  deleteMock: vi.fn(),
}))

vi.mock('@/api/incidents', () => ({
  incidentsApi: {
    updateTitle: updateTitleMock,
    postUpdate: postUpdateMock,
    updateUpdateMessage: updateUpdateMessageMock,
    delete: deleteMock,
  },
}))

const incidentData = ref<unknown>(null)
const incidentPending = ref(false)
const loadError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useIncidents', () => ({
  useIncident: () => ({ data: incidentData, isPending: incidentPending, error: loadError }),
}))

const incident = {
  id: 'i1',
  title: 'Elevated latency',
  severity: 'major' as const,
  status: 'investigating' as const,
  monitors: [{ monitorType: 'uptime' as const, monitorId: 'u1', name: 'API uptime' }],
  monitorCount: 1,
  updates: [
    {
      id: 'up1',
      message: 'Initial report',
      status: 'investigating' as const,
      createdAt: '2026-06-01T00:00:00Z',
    },
  ],
  createdAt: '2026-06-01T00:00:00Z',
  updatedAt: '2026-06-01T00:00:00Z',
  resolvedAt: null,
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  incidentData.value = null
  incidentPending.value = false
  loadError.value = null
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('IncidentDetailView', () => {
  it('shows a loading state while pending', () => {
    incidentPending.value = true
    const wrapper = mount(IncidentDetailView)

    expect(wrapper.text()).toContain('Loading…')
  })

  it('renders just the back link when loading fails (incident stays unset)', async () => {
    const wrapper = mount(IncidentDetailView)
    loadError.value = { message: 'Incident not found' }
    await flushPromises()

    expect(wrapper.text()).toContain('← Back')
    expect(wrapper.text()).not.toContain('Loading…')
  })

  it('renders incident details, monitors, and the updates feed', () => {
    incidentData.value = { ...incident }
    const wrapper = mount(IncidentDetailView)

    expect(wrapper.text()).toContain('Elevated latency')
    expect(wrapper.text()).toContain('Major')
    expect(wrapper.text()).toContain('Investigating')
    expect(wrapper.text()).toContain('API uptime')
    expect(wrapper.text()).toContain('Initial report')
  })

  it('hides the post-update form once the incident is resolved', () => {
    incidentData.value = { ...incident, status: 'resolved' }
    const wrapper = mount(IncidentDetailView)

    expect(wrapper.find('#newMessage').exists()).toBe(false)
  })

  it('edits and saves the title', async () => {
    incidentData.value = { ...incident }
    updateTitleMock.mockResolvedValueOnce({ ...incident, title: 'Renamed' })
    const wrapper = mount(IncidentDetailView)

    await findButtonByText(wrapper, 'Edit')!.trigger('click')
    const titleInput = wrapper.find('.flex-1 input')
    await titleInput.setValue('Renamed')
    await findButtonByText(wrapper, 'Save')!.trigger('click')
    await flushPromises()

    expect(updateTitleMock).toHaveBeenCalledExactlyOnceWith('i1', 'Renamed')
    expect(setQueryDataMock).toHaveBeenCalledExactlyOnceWith(['incident', 'i1'], {
      ...incident,
      title: 'Renamed',
    })
  })

  it('cancels title editing without saving', async () => {
    incidentData.value = { ...incident }
    const wrapper = mount(IncidentDetailView)

    await findButtonByText(wrapper, 'Edit')!.trigger('click')
    expect(wrapper.find('.flex-1 input').exists()).toBe(true)

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(updateTitleMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Elevated latency')
  })

  it('requires a message before posting an update', async () => {
    incidentData.value = { ...incident }
    const wrapper = mount(IncidentDetailView)

    await findButtonByText(wrapper, 'Post update')!.trigger('click')

    expect(wrapper.text()).toContain('Update message is required')
    expect(postUpdateMock).not.toHaveBeenCalled()
  })

  it('posts an update defaulting to the next status in sequence', async () => {
    incidentData.value = { ...incident }
    const afterUpdate = { ...incident, status: 'identified' as const }
    postUpdateMock.mockResolvedValueOnce(afterUpdate)
    const wrapper = mount(IncidentDetailView)

    await wrapper.find('#newMessage').setValue('Found the cause')
    await findButtonByText(wrapper, 'Post update')!.trigger('click')
    await flushPromises()

    expect(postUpdateMock).toHaveBeenCalledExactlyOnceWith('i1', 'Found the cause', 'identified')
    expect(setQueryDataMock).toHaveBeenCalledExactlyOnceWith(['incident', 'i1'], afterUpdate)
  })

  it('edits an existing update message', async () => {
    incidentData.value = { ...incident }
    const afterEdit = {
      ...incident,
      updates: [{ ...incident.updates[0], message: 'Corrected' }],
    }
    updateUpdateMessageMock.mockResolvedValueOnce(afterEdit)
    const wrapper = mount(IncidentDetailView)

    // Two "Edit" buttons exist by default: the title's, then this one
    // update's. The update's is the second in DOM order.
    const editButtons = wrapper.findAll('button').filter((b) => b.text() === 'Edit')
    await editButtons[1]!.trigger('click')

    const draftInput = wrapper.find('input:not(#newMessage)')
    await draftInput.setValue('Corrected')
    await findButtonByText(wrapper, 'Save')!.trigger('click')
    await flushPromises()

    expect(updateUpdateMessageMock).toHaveBeenCalledExactlyOnceWith('i1', 'up1', 'Corrected')
    expect(setQueryDataMock).toHaveBeenCalledExactlyOnceWith(['incident', 'i1'], afterEdit)
  })

  it('deletes the incident on confirm and navigates to the list', async () => {
    incidentData.value = { ...incident }
    deleteMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(IncidentDetailView)

    await findButtonByText(wrapper, 'Delete incident')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('i1')
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'incidents' })
  })

  it('shows an error and resets confirmation state when delete fails', async () => {
    incidentData.value = { ...incident }
    deleteMock.mockRejectedValueOnce(new Error('Delete failed'))
    const wrapper = mount(IncidentDetailView)

    await findButtonByText(wrapper, 'Delete incident')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Delete failed')
    expect(pushMock).not.toHaveBeenCalled()
    expect(findButtonByText(wrapper, 'Delete incident')).toBeTruthy()
  })
})
