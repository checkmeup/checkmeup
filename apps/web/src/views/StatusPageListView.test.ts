import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import StatusPageListView from './StatusPageListView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const refetchMock = vi.fn()
const pagesData = ref<unknown>(null)
const pagesPending = ref(false)
const pagesError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useStatusPages', () => ({
  useStatusPages: () => ({
    data: pagesData,
    isPending: pagesPending,
    error: pagesError,
    refetch: refetchMock,
  }),
}))

const pages = [
  {
    id: 'sp1',
    slug: 'acme',
    title: 'Acme Status',
    description: '',
    logoUrl: '',
    publicUrl: 'https://checkmeup.net/status/acme',
    createdAt: '2026-06-01T00:00:00Z',
  },
  {
    id: 'sp2',
    slug: 'beta',
    title: 'Beta Status',
    description: '',
    logoUrl: '',
    publicUrl: 'https://checkmeup.net/status/beta',
    createdAt: '2026-06-02T00:00:00Z',
  },
]

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  pagesData.value = null
  pagesPending.value = false
  pagesError.value = null
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('StatusPageListView', () => {
  it('shows a loading state while pending', () => {
    pagesPending.value = true
    const wrapper = mount(StatusPageListView)

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message and retries on click', async () => {
    pagesError.value = { message: 'Failed to load status pages' }
    const wrapper = mount(StatusPageListView)

    expect(wrapper.text()).toContain('Failed to load status pages')
    await findButtonByText(wrapper, 'Try again')!.trigger('click')

    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an empty state when there are no pages', () => {
    pagesData.value = []
    const wrapper = mount(StatusPageListView)

    expect(wrapper.text()).toContain('No status pages yet')
    expect(findButtonByText(wrapper, 'Create your first page')).toBeTruthy()
  })

  it('navigates to the create view from the empty state', async () => {
    pagesData.value = []
    const wrapper = mount(StatusPageListView)

    await findButtonByText(wrapper, 'Create your first page')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'status-page-create' })
  })

  it('renders the list of pages with their slugs', () => {
    pagesData.value = [...pages]
    const wrapper = mount(StatusPageListView)

    expect(wrapper.text()).toContain('Acme Status')
    expect(wrapper.text()).toContain('/status/acme')
    expect(wrapper.text()).toContain('Beta Status')
    expect(wrapper.text()).toContain('/status/beta')
  })

  it('navigates to the create view via the header button', async () => {
    pagesData.value = [...pages]
    const wrapper = mount(StatusPageListView)

    await findButtonByText(wrapper, 'Create page')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'status-page-create' })
  })

  it('navigates to the detail view when a row is clicked', async () => {
    pagesData.value = [...pages]
    const wrapper = mount(StatusPageListView)

    const row = wrapper.findAll('tbody tr').find((r) => r.text().includes('Acme Status'))
    await row!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'status-page-detail',
      params: { id: 'sp1' },
    })
  })
})
