import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import StatusPageCreateView from './StatusPageCreateView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { checkSlugMock, createMock } = vi.hoisted(() => ({
  checkSlugMock: vi.fn(),
  createMock: vi.fn(),
}))

vi.mock('@/api/statusPages', () => ({
  statusPagesApi: { checkSlug: checkSlugMock, create: createMock },
}))

const { ApiError } = vi.hoisted(() => ({
  ApiError: class ApiError extends Error {
    status: number
    code: string
    constructor(status: number, message: string, code = '') {
      super(message)
      this.status = status
      this.code = code
    }
  },
}))

vi.mock('@/api/client', () => ({ ApiError }))

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  vi.useFakeTimers()
  checkSlugMock.mockResolvedValue({ available: true, reason: '' })
})

afterEach(() => {
  vi.useRealTimers()
  vi.clearAllMocks()
})

describe('StatusPageCreateView', () => {
  it('renders the form', () => {
    const wrapper = mount(StatusPageCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('New status page')
    expect(wrapper.find('#title').exists()).toBe(true)
    expect(wrapper.find('#slug').exists()).toBe(true)
  })

  it('auto-derives the slug from the title while the user has not edited it', async () => {
    const wrapper = mount(StatusPageCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const titleInput = wrapper.get('#title')
    await titleInput.setValue('Acme Status Page')

    const slugInput = wrapper.get('#slug').element as HTMLInputElement
    expect(slugInput.value).toBe('acme-status-page')
  })

  it('checks slug availability after the debounce and shows "Available"', async () => {
    const wrapper = mount(StatusPageCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.get('#slug').setValue('acme')
    expect(wrapper.text()).toContain('Checking…')

    await vi.advanceTimersByTimeAsync(400)
    await flushPromises()

    expect(checkSlugMock).toHaveBeenCalledExactlyOnceWith('acme')
    expect(wrapper.text()).toContain('Available')
  })

  it('shows the taken reason when the slug is unavailable', async () => {
    checkSlugMock.mockResolvedValueOnce({ available: false, reason: 'Slug already in use' })
    const wrapper = mount(StatusPageCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.get('#slug').setValue('taken')
    await vi.advanceTimersByTimeAsync(400)
    await flushPromises()

    expect(wrapper.text()).toContain('Slug already in use')
  })

  it('requires a title before submitting', async () => {
    const wrapper = mount(StatusPageCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.get('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Title is required')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('requires an available slug before submitting', async () => {
    const wrapper = mount(StatusPageCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.get('#title').setValue('Acme')
    await wrapper.get('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Please enter a valid, available slug')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('creates the page and navigates to the detail view on success', async () => {
    createMock.mockResolvedValueOnce({ id: 'sp1' })
    const wrapper = mount(StatusPageCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.get('#title').setValue('Acme Status')
    await vi.advanceTimersByTimeAsync(400)
    await flushPromises()

    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(createMock).toHaveBeenCalledExactlyOnceWith({
      slug: 'acme-status',
      title: 'Acme Status',
      description: '',
      logoUrl: '',
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'status-page-detail',
      params: { id: 'sp1' },
    })
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    createMock.mockRejectedValueOnce(
      new ApiError(402, 'Upgrade to add more status pages', 'plan_limit_reached'),
    )
    const wrapper = mount(StatusPageCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.get('#title').setValue('Acme Status')
    await vi.advanceTimersByTimeAsync(400)
    await flushPromises()

    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to add more status pages')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalledWith(
      expect.objectContaining({ name: 'status-page-detail' }),
    )
  })

  it('shows a generic error message when creation fails for another reason', async () => {
    createMock.mockRejectedValueOnce(new Error('Network error'))
    const wrapper = mount(StatusPageCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.get('#title').setValue('Acme Status')
    await vi.advanceTimersByTimeAsync(400)
    await flushPromises()

    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Network error')
  })

  it('navigates back to the list when cancel is clicked', async () => {
    const wrapper = mount(StatusPageCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'status-pages' })
  })
})
