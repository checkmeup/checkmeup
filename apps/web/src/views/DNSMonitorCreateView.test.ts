import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import DNSMonitorCreateView from './DNSMonitorCreateView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {} }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { createDnsMock } = vi.hoisted(() => ({
  createDnsMock: vi.fn(),
}))

vi.mock('@/api/monitors', async () => {
  const actual = await vi.importActual<typeof import('@/api/monitors')>('@/api/monitors')
  return {
    ...actual,
    monitorsApi: { createDns: createDnsMock },
  }
})

const billingData = ref<{ minIntervalMins: number } | null>(null)
const billingPending = ref(false)

vi.mock('@/composables/useBilling', () => ({
  useBilling: () => ({ data: billingData, isPending: billingPending }),
}))

const channelsData = ref<{ id: string; name: string; type: string; enabled: boolean }[]>([])
const channelsPending = ref(false)

vi.mock('@/composables/useNotificationChannels', () => ({
  useNotificationChannels: () => ({ data: channelsData, isPending: channelsPending }),
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

describe('DNSMonitorCreateView', () => {
  it('renders the form with default values', () => {
    const wrapper = mount(DNSMonitorCreateView)

    expect(wrapper.text()).toContain('New DNS monitor')
    expect(wrapper.find('#name').exists()).toBe(true)
    expect(wrapper.find('#hostname').exists()).toBe(true)
    expect(wrapper.find('#recordType').exists()).toBe(true)
    expect(wrapper.find('#expectedValue').exists()).toBe(true)
  })

  it('adds the 1-minute interval option when billing allows it', async () => {
    billingData.value = { minIntervalMins: 1 }
    const wrapper = mount(DNSMonitorCreateView)
    await flushPromises()

    const options = wrapper.find('#interval').findAll('option')
    expect(options.map((o) => o.text())).toContain('1 minute')
  })

  it('shows a validation error when name is missing', async () => {
    const wrapper = mount(DNSMonitorCreateView)

    await wrapper.find('#hostname').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Name is required')
    expect(createDnsMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when hostname is missing', async () => {
    const wrapper = mount(DNSMonitorCreateView)

    await wrapper.find('#name').setValue('Apex A record')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Hostname is required')
    expect(createDnsMock).not.toHaveBeenCalled()
  })

  it('creates the monitor in baseline mode (no expected value) and navigates to its detail page', async () => {
    createDnsMock.mockResolvedValueOnce({ id: 'd1' })
    const wrapper = mount(DNSMonitorCreateView)

    await wrapper.find('#name').setValue('Apex A record')
    await wrapper.find('#hostname').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createDnsMock).toHaveBeenCalledExactlyOnceWith({
      name: 'Apex A record',
      hostname: 'example.com',
      recordType: 'A',
      expectedValue: undefined,
      intervalMins: 10,
      maxAlertsPerIncident: 3,
      alertAfterNFailures: 0,
      channelIds: [],
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'dns-monitor-detail',
      params: { id: 'd1' },
    })
  })

  it('submits a pinned expected value and non-default record type when set', async () => {
    createDnsMock.mockResolvedValueOnce({ id: 'd1' })
    const wrapper = mount(DNSMonitorCreateView)

    await wrapper.find('#name').setValue('Mail MX')
    await wrapper.find('#hostname').setValue('example.com')
    await wrapper.find('#recordType').setValue('MX')
    await wrapper.find('#expectedValue').setValue('mail.example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createDnsMock).toHaveBeenCalledExactlyOnceWith(
      expect.objectContaining({ recordType: 'MX', expectedValue: 'mail.example.com' }),
    )
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    const { ApiError } = await import('@/api/client')
    createDnsMock.mockRejectedValueOnce(
      new ApiError(402, 'Upgrade to add more monitors', 'plan_limit_reached'),
    )
    const wrapper = mount(DNSMonitorCreateView)

    await wrapper.find('#name').setValue('Apex A record')
    await wrapper.find('#hostname').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to add more monitors')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic inline error when creation fails for another reason', async () => {
    createDnsMock.mockRejectedValueOnce(new Error('Network error'))
    const wrapper = mount(DNSMonitorCreateView)

    await wrapper.find('#name').setValue('Apex A record')
    await wrapper.find('#hostname').setValue('example.com')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Network error')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the list when cancel is clicked', async () => {
    const wrapper = mount(DNSMonitorCreateView)

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'dns-monitors' })
  })
})
