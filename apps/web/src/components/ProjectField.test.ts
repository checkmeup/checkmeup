import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import ProjectField from './ProjectField.vue'

const projectsData = ref([
  { id: 'p1', name: 'Acme', monitorCount: 1, createdAt: '' },
  { id: 'p2', name: 'Globex', monitorCount: 0, createdAt: '' },
])
const addProjectMock = vi.fn()

vi.mock('@/composables/useProjects', () => ({
  useProjects: () => ({ data: projectsData }),
  useAddProject: () => addProjectMock,
}))

afterEach(() => {
  vi.clearAllMocks()
})

describe('ProjectField (PRD projects)', () => {
  it('offers "No project", every project and "New project…"', () => {
    const wrapper = mount(ProjectField, { props: { modelValue: null } })
    const options = wrapper.findAll('option').map((o) => o.text())
    expect(options).toEqual(['No project', 'Acme', 'Globex', 'New project…'])
    expect((wrapper.find('select').element as HTMLSelectElement).value).toBe('')
  })

  it('selects a project, and "No project" clears it', async () => {
    const wrapper = mount(ProjectField, { props: { modelValue: 'p1' } })
    expect((wrapper.find('select').element as HTMLSelectElement).value).toBe('p1')

    await wrapper.find('select').setValue('p2')
    await wrapper.find('select').setValue('')
    expect(wrapper.emitted('update:modelValue')).toEqual([['p2'], [null]])
  })

  it('creates a new project from the form and selects it', async () => {
    addProjectMock.mockResolvedValueOnce({ id: 'p9', name: 'Initech', monitorCount: 0, createdAt: '' })
    const wrapper = mount(ProjectField, { props: { modelValue: null } })

    await wrapper.find('select').setValue('__new__')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    await wrapper.find('input').setValue('  Initech ')
    await wrapper.findAll('button').find((b) => b.text() === 'Create')!.trigger('click')
    await flushPromises()

    expect(addProjectMock).toHaveBeenCalledExactlyOnceWith('Initech')
    expect(wrapper.emitted('update:modelValue')).toEqual([['p9']])
    expect(wrapper.find('select').exists()).toBe(true)
  })

  it('shows why a new project was refused and keeps the name field open', async () => {
    addProjectMock.mockRejectedValueOnce(new Error('a project with this name already exists'))
    const wrapper = mount(ProjectField, { props: { modelValue: null } })

    await wrapper.find('select').setValue('__new__')
    await wrapper.find('input').setValue('Acme')
    await wrapper.findAll('button').find((b) => b.text() === 'Create')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('a project with this name already exists')
    expect(wrapper.find('input').exists()).toBe(true)
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('needs a name, and Cancel goes back to the list', async () => {
    const wrapper = mount(ProjectField, { props: { modelValue: null } })
    await wrapper.find('select').setValue('__new__')
    await wrapper.findAll('button').find((b) => b.text() === 'Create')!.trigger('click')
    expect(wrapper.text()).toContain('Name is required')
    expect(addProjectMock).not.toHaveBeenCalled()

    await wrapper.findAll('button').find((b) => b.text() === 'Cancel')!.trigger('click')
    expect(wrapper.find('select').exists()).toBe(true)
  })
})
