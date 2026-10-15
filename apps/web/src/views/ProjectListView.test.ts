import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import ProjectListView from './ProjectListView.vue'

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { createMock, renameMock, deleteMock, setMonitorsMock, invalidateMock } = vi.hoisted(() => ({
  createMock: vi.fn(),
  renameMock: vi.fn(),
  deleteMock: vi.fn(),
  setMonitorsMock: vi.fn(),
  invalidateMock: vi.fn(),
}))

vi.mock('@/api/projects', () => ({
  projectsApi: {
    create: createMock,
    rename: renameMock,
    delete: deleteMock,
    setMonitors: setMonitorsMock,
  },
}))

vi.mock('@tanstack/vue-query', () => ({
  useQueryClient: () => ({ invalidateQueries: invalidateMock }),
}))

const projectsData = ref<unknown[] | null>([])
const projectsPending = ref(false)
const projectsError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useProjects', () => ({
  useProjects: () => ({
    data: projectsData,
    isPending: projectsPending,
    error: projectsError,
    refetch: vi.fn(),
  }),
}))

const rowsData = ref<unknown[]>([])
vi.mock('@/composables/useMonitorRows', () => ({
  useMonitorRows: () => ({ rows: rowsData }),
}))

const acme = { id: 'p1', name: 'Acme', monitorCount: 2, createdAt: '' }
const globex = { id: 'p2', name: 'Globex', monitorCount: 0, createdAt: '' }

// The picker loads every monitor type itself; here it only has to show and
// change its selection.
const Picker = {
  name: 'MaintenanceMonitorPicker',
  props: ['modelValue'],
  emits: ['update:modelValue'],
  template: '<div class="picker">{{ modelValue.map((m) => m.monitorId).join(",") }}</div>',
}

function mountView() {
  return mount(ProjectListView, {
    global: {
      stubs: {
        RouterLink: { props: ['to'], template: '<a :data-to="JSON.stringify(to)"><slot /></a>' },
        MaintenanceMonitorPicker: Picker,
      },
    },
  })
}

function button(wrapper: ReturnType<typeof mountView>, text: string, index = 0) {
  return wrapper.findAll('button').filter((b) => b.text() === text)[index]!
}

beforeEach(() => {
  projectsData.value = [acme, globex]
  projectsPending.value = false
  projectsError.value = null
  rowsData.value = [
    { type: 'cron', id: 'c1', name: 'Backup', projectId: 'p1' },
    { type: 'uptime', id: 'u1', name: 'Site', projectId: 'p1' },
    { type: 'ssl', id: 's1', name: 'Cert', projectId: null },
  ]
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('ProjectListView (PRD projects)', () => {
  it('lists projects with their monitor counts, each linking to its monitors', () => {
    const wrapper = mountView()
    expect(wrapper.text()).toContain('Acme')
    expect(wrapper.text()).toContain('2 monitors')
    expect(wrapper.text()).toContain('Globex')
    expect(wrapper.text()).toContain('0 monitors')
    const link = wrapper.findAll('a').find((a) => a.text() === '2 monitors')!
    expect(JSON.parse(link.attributes('data-to')!)).toEqual({
      name: 'monitors',
      query: { project: 'p1' },
    })
  })

  it('shows an empty state without projects', () => {
    projectsData.value = []
    expect(mountView().text()).toContain('No projects yet')
  })

  it('creates a project', async () => {
    createMock.mockResolvedValueOnce({ id: 'p3', name: 'Initech', monitorCount: 0, createdAt: '' })
    const wrapper = mountView()
    await wrapper.find('#new-project').setValue(' Initech ')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(createMock).toHaveBeenCalledExactlyOnceWith('Initech')
    expect(invalidateMock).toHaveBeenCalledWith({ queryKey: ['projects'] })
    expect((wrapper.find('#new-project').element as HTMLInputElement).value).toBe('')
  })

  it('shows why a project was refused', async () => {
    createMock.mockRejectedValueOnce(new Error('a project with this name already exists'))
    const wrapper = mountView()
    await wrapper.find('#new-project').setValue('Acme')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    expect(wrapper.text()).toContain('a project with this name already exists')
  })

  it('renames a project', async () => {
    renameMock.mockResolvedValueOnce({ ...acme, name: 'Acme Corp' })
    const wrapper = mountView()
    await button(wrapper, 'Rename').trigger('click')
    await wrapper.find('#rename-p1').setValue('Acme Corp')
    // The second form on the page, after "Add project".
    await wrapper.findAll('form')[1]!.trigger('submit')
    await flushPromises()

    expect(renameMock).toHaveBeenCalledExactlyOnceWith('p1', 'Acme Corp')
    expect(wrapper.find('#rename-p1').exists()).toBe(false)
  })

  it('deletes a project after confirming, saying its monitors stay', async () => {
    deleteMock.mockResolvedValueOnce(undefined)
    const wrapper = mountView()
    await button(wrapper, 'Delete', 1).trigger('click')
    expect(wrapper.text()).toContain('Delete Globex? Its monitors keep running, without a project.')
    expect(deleteMock).not.toHaveBeenCalled()

    await button(wrapper, 'Confirm delete').trigger('click')
    await flushPromises()
    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('p2')
    expect(invalidateMock).toHaveBeenCalledWith({ queryKey: ['cron-monitors'] })
  })

  it("opens the monitor picker with the project's monitors and saves the new selection", async () => {
    setMonitorsMock.mockResolvedValueOnce({ ...acme, monitorCount: 2 })
    const wrapper = mountView()
    await button(wrapper, 'Monitors').trigger('click')
    const picker = wrapper.findComponent(Picker)
    expect(picker.text()).toBe('c1,u1')

    picker.vm.$emit('update:modelValue', [
      { monitorType: 'uptime', monitorId: 'u1', name: 'Site' },
      { monitorType: 'ssl', monitorId: 's1', name: 'Cert' },
    ])
    await flushPromises()
    await button(wrapper, 'Save monitors').trigger('click')
    await flushPromises()

    expect(setMonitorsMock).toHaveBeenCalledExactlyOnceWith('p1', [
      { type: 'uptime', id: 'u1' },
      { type: 'ssl', id: 's1' },
    ])
    expect(invalidateMock).toHaveBeenCalledWith({ queryKey: ['projects'] })
    expect(invalidateMock).toHaveBeenCalledWith({ queryKey: ['ssl-monitors'] })
    expect(wrapper.findComponent(Picker).exists()).toBe(false)
  })
})
