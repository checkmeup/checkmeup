import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import UptimeMonitorCreateView from './UptimeMonitorCreateView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {} }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { createUptimeMock } = vi.hoisted(() => ({
  createUptimeMock: vi.fn(),
}))

vi.mock('@/api/monitors', async () => {
  const actual = await vi.importActual<typeof import('@/api/monitors')>('@/api/monitors')
  return {
    ...actual,
    monitorsApi: { createUptime: createUptimeMock },
  }
})

const billingData = ref<{ minIntervalMins: number } | null>(null)
const billingPending = ref(false)

vi.mock('@/composables/useBilling', () => ({
  useBilling: () => ({ data: billingData, isPending: billingPending }),
}))

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  billingData.value = null
  billingPending.value = false
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('UptimeMonitorCreateView', () => {
  it('renders the form with default values', () => {
    const wrapper = mount(UptimeMonitorCreateView)

    expect(wrapper.text()).toContain('New uptime monitor')
    expect(wrapper.find('#name').exists()).toBe(true)
    expect(wrapper.find('#url').exists()).toBe(true)
    expect(wrapper.find('#keywordMode').exists()).toBe(false)
  })

  it('reveals keyword mode options once a keyword is entered', async () => {
    const wrapper = mount(UptimeMonitorCreateView)

    await wrapper.find('#keyword').setValue('Welcome')

    expect(wrapper.find('#keywordMode').exists()).toBe(true)
    expect(wrapper.find('#keywordCaseSensitive').exists()).toBe(true)
  })

  it('adds the 1-minute interval option when billing allows it', async () => {
    billingData.value = { minIntervalMins: 1 }
    const wrapper = mount(UptimeMonitorCreateView)
    await flushPromises()

    const options = wrapper.find('#interval').findAll('option')
    expect(options.map((o) => o.text())).toContain('1 minute')
  })

  it('shows a validation error when name is missing', async () => {
    const wrapper = mount(UptimeMonitorCreateView)

    await wrapper.find('#url').setValue('https://example.com')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Name is required')
    expect(createUptimeMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when url is missing', async () => {
    const wrapper = mount(UptimeMonitorCreateView)

    await wrapper.find('#name').setValue('Production API')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('URL is required')
    expect(createUptimeMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when url has no protocol', async () => {
    const wrapper = mount(UptimeMonitorCreateView)

    await wrapper.find('#name').setValue('Production API')
    await wrapper.find('#url').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('URL must start with http:// or https://')
    expect(createUptimeMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when keyword exceeds 500 characters', async () => {
    const wrapper = mount(UptimeMonitorCreateView)

    await wrapper.find('#name').setValue('Production API')
    await wrapper.find('#url').setValue('https://example.com')
    await wrapper.find('#keyword').setValue('a'.repeat(501))
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Keyword must be 500 characters or fewer')
    expect(createUptimeMock).not.toHaveBeenCalled()
  })

  it('creates the monitor and navigates to its detail page on success', async () => {
    createUptimeMock.mockResolvedValueOnce({ id: 'u1' })
    const wrapper = mount(UptimeMonitorCreateView)

    await wrapper.find('#name').setValue('Production API')
    await wrapper.find('#url').setValue('https://example.com/health')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createUptimeMock).toHaveBeenCalledExactlyOnceWith({
      name: 'Production API',
      url: 'https://example.com/health',
      intervalMins: 10,
      maxAlertsPerIncident: 3,
      keyword: '',
      keywordMode: 'contains',
      keywordCaseSensitive: false,
      jsonAssertions: [],
      maxResponseTimeMs: null,
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'uptime-monitor-detail',
      params: { id: 'u1' },
    })
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    const { ApiError } = await import('@/api/client')
    createUptimeMock.mockRejectedValueOnce(
      new ApiError(402, 'Upgrade to add more monitors', 'plan_limit_reached'),
    )
    const wrapper = mount(UptimeMonitorCreateView)

    await wrapper.find('#name').setValue('Production API')
    await wrapper.find('#url').setValue('https://example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to add more monitors')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic inline error when creation fails for another reason', async () => {
    createUptimeMock.mockRejectedValueOnce(new Error('Network error'))
    const wrapper = mount(UptimeMonitorCreateView)

    await wrapper.find('#name').setValue('Production API')
    await wrapper.find('#url').setValue('https://example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Network error')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the list when cancel is clicked', async () => {
    const wrapper = mount(UptimeMonitorCreateView)

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'uptime-monitors' })
  })
})
