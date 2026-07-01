import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PortMonitorCreateView from './PortMonitorCreateView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {} }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { createPortMock } = vi.hoisted(() => ({
  createPortMock: vi.fn(),
}))

vi.mock('@/api/monitors', async () => {
  const actual = await vi.importActual<typeof import('@/api/monitors')>('@/api/monitors')
  return {
    ...actual,
    monitorsApi: { createPort: createPortMock },
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

describe('PortMonitorCreateView', () => {
  it('renders the form with default values', () => {
    const wrapper = mount(PortMonitorCreateView)

    expect(wrapper.text()).toContain('New port monitor')
    expect(wrapper.find('#name').exists()).toBe(true)
    expect(wrapper.find('#host').exists()).toBe(true)
    expect(wrapper.find('#port').exists()).toBe(true)
    expect(wrapper.find('#expectedState').exists()).toBe(true)
  })

  it('adds the 1-minute interval option when billing allows it', async () => {
    billingData.value = { minIntervalMins: 1 }
    const wrapper = mount(PortMonitorCreateView)
    await flushPromises()

    const options = wrapper.find('#interval').findAll('option')
    expect(options.map((o) => o.text())).toContain('1 minute')
  })

  it('shows a validation error when name is missing', async () => {
    const wrapper = mount(PortMonitorCreateView)

    await wrapper.find('#host').setValue('example.com')
    await wrapper.find('#port').setValue(443)
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Name is required')
    expect(createPortMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when host is missing', async () => {
    const wrapper = mount(PortMonitorCreateView)

    await wrapper.find('#name').setValue('SMTP')
    await wrapper.find('#port').setValue(25)
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Host is required')
    expect(createPortMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when port is out of range', async () => {
    const wrapper = mount(PortMonitorCreateView)

    await wrapper.find('#name').setValue('SMTP')
    await wrapper.find('#host').setValue('example.com')
    await wrapper.find('#port').setValue(70000)
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Port must be between 1 and 65535')
    expect(createPortMock).not.toHaveBeenCalled()
  })

  it('creates the monitor and navigates to its detail page on success', async () => {
    createPortMock.mockResolvedValueOnce({ id: 'p1' })
    const wrapper = mount(PortMonitorCreateView)

    await wrapper.find('#name').setValue('SMTP')
    await wrapper.find('#host').setValue('mail.example.com')
    await wrapper.find('#port').setValue(25)
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createPortMock).toHaveBeenCalledExactlyOnceWith({
      name: 'SMTP',
      host: 'mail.example.com',
      port: 25,
      expectedState: 'open',
      intervalMins: 10,
      maxAlertsPerIncident: 3,
      alertAfterNFailures: 0,
      channelIds: [],
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'port-monitor-detail',
      params: { id: 'p1' },
    })
  })

  it('submits expectedState closed when selected', async () => {
    createPortMock.mockResolvedValueOnce({ id: 'p1' })
    const wrapper = mount(PortMonitorCreateView)

    await wrapper.find('#name').setValue('Firewalled DB')
    await wrapper.find('#host').setValue('db.internal')
    await wrapper.find('#port').setValue(5432)
    await wrapper.find('#expectedState').setValue('closed')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createPortMock).toHaveBeenCalledExactlyOnceWith(
      expect.objectContaining({ expectedState: 'closed' }),
    )
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    const { ApiError } = await import('@/api/client')
    createPortMock.mockRejectedValueOnce(
      new ApiError(402, 'Upgrade to add more monitors', 'plan_limit_reached'),
    )
    const wrapper = mount(PortMonitorCreateView)

    await wrapper.find('#name').setValue('SMTP')
    await wrapper.find('#host').setValue('example.com')
    await wrapper.find('#port').setValue(25)
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to add more monitors')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic inline error when creation fails for another reason', async () => {
    createPortMock.mockRejectedValueOnce(new Error('Network error'))
    const wrapper = mount(PortMonitorCreateView)

    await wrapper.find('#name').setValue('SMTP')
    await wrapper.find('#host').setValue('example.com')
    await wrapper.find('#port').setValue(25)
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Network error')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the list when cancel is clicked', async () => {
    const wrapper = mount(PortMonitorCreateView)

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'port-monitors' })
  })
})
