import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import StatusPageEditView from './StatusPageEditView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'sp1' } }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { updateMock } = vi.hoisted(() => ({
  updateMock: vi.fn(),
}))

vi.mock('@/api/statusPages', () => ({
  statusPagesApi: { update: updateMock },
}))

const pageData = ref<unknown>(null)
const pagePending = ref(false)
const pageError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useStatusPages', () => ({
  useStatusPage: () => ({
    data: pageData,
    isPending: pagePending,
    error: pageError,
  }),
}))

const billingData = ref<{ plan: string } | null>({ plan: 'solo' })

vi.mock('@/composables/useBilling', () => ({
  useBilling: () => ({ data: billingData }),
}))

const page = {
  id: 'sp1',
  slug: 'acme',
  title: 'Acme Status',
  description: 'Live status of Acme services',
  logoUrl: 'https://example.com/logo.png',
  hideBranding: false,
  publicUrl: 'https://checkmeup.net/status/acme',
  createdAt: '2026-06-01T00:00:00Z',
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  pageData.value = null
  pagePending.value = false
  pageError.value = null
  billingData.value = { plan: 'solo' }
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('StatusPageEditView', () => {
  it('shows a loading state while the page is pending', () => {
    pagePending.value = true
    const wrapper = mount(StatusPageEditView)

    expect(wrapper.text()).toContain('Loading…')
    expect(wrapper.find('form').exists()).toBe(false)
  })

  it('populates the form fields from the loaded page', () => {
    pageData.value = { ...page }
    const wrapper = mount(StatusPageEditView)

    expect((wrapper.get('#title').element as HTMLInputElement).value).toBe('Acme Status')
    expect((wrapper.get('#description').element as HTMLInputElement).value).toBe(
      'Live status of Acme services',
    )
    expect((wrapper.get('#logoUrl').element as HTMLInputElement).value).toBe(
      'https://example.com/logo.png',
    )
    expect(wrapper.text()).toContain('/status/acme')
  })

  it('surfaces the load error in the form once data resolves to null', async () => {
    const wrapper = mount(StatusPageEditView)
    pageError.value = { message: 'Status page not found' }
    await flushPromises()

    expect(wrapper.text()).toContain('Status page not found')
  })

  it('requires a title before submitting', async () => {
    pageData.value = { ...page, title: '' }
    const wrapper = mount(StatusPageEditView)

    await wrapper.get('#title').setValue('')
    await wrapper.get('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Title is required')
    expect(updateMock).not.toHaveBeenCalled()
  })

  it('updates the page and navigates to the detail view on success', async () => {
    pageData.value = { ...page }
    updateMock.mockResolvedValueOnce({ ...page })
    const wrapper = mount(StatusPageEditView)

    await wrapper.get('#title').setValue('Acme Status Updated')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(updateMock).toHaveBeenCalledExactlyOnceWith('sp1', {
      title: 'Acme Status Updated',
      description: 'Live status of Acme services',
      logoUrl: 'https://example.com/logo.png',
      hideBranding: false,
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'status-page-detail',
      params: { id: 'sp1' },
    })
  })

  it('shows an error message and does not navigate when update fails', async () => {
    pageData.value = { ...page }
    updateMock.mockRejectedValueOnce(new Error('Failed to update page'))
    const wrapper = mount(StatusPageEditView)

    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to update page')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the detail view when cancel is clicked', async () => {
    pageData.value = { ...page }
    const wrapper = mount(StatusPageEditView)

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'status-page-detail',
      params: { id: 'sp1' },
    })
  })

  it('enables the hide-branding checkbox on a paid plan and submits it checked', async () => {
    billingData.value = { plan: 'solo' }
    pageData.value = { ...page }
    updateMock.mockResolvedValueOnce({ ...page, hideBranding: true })
    const wrapper = mount(StatusPageEditView)

    const checkbox = wrapper.find('input[type="checkbox"]')
    expect((checkbox.element as HTMLInputElement).disabled).toBe(false)
    await checkbox.setValue(true)
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(updateMock).toHaveBeenCalledExactlyOnceWith('sp1', {
      title: 'Acme Status',
      description: 'Live status of Acme services',
      logoUrl: 'https://example.com/logo.png',
      hideBranding: true,
    })
  })

  it('disables the hide-branding checkbox on Hobby and shows an upgrade hint', () => {
    billingData.value = { plan: 'hobby' }
    pageData.value = { ...page }
    const wrapper = mount(StatusPageEditView)

    const checkbox = wrapper.find('input[type="checkbox"]')
    expect((checkbox.element as HTMLInputElement).disabled).toBe(true)
    expect(wrapper.text()).toContain('Hiding branding requires a paid plan')
  })
})
