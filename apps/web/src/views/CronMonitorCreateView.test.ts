import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import CronMonitorCreateView from './CronMonitorCreateView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: {} }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { createCronMock } = vi.hoisted(() => ({
  createCronMock: vi.fn(),
}))

vi.mock('@/api/monitors', () => ({
  monitorsApi: { createCron: createCronMock },
}))

const { ApiErrorMock } = vi.hoisted(() => ({
  ApiErrorMock: class extends Error {
    status: number
    code: string
    constructor(status: number, message: string, code = '') {
      super(message)
      this.status = status
      this.code = code
    }
  },
}))

vi.mock('@/api/client', () => ({
  ApiError: ApiErrorMock,
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
  createCronMock.mockReset()
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('CronMonitorCreateView', () => {
  it('renders the form with default field values', () => {
    const wrapper = mount(CronMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('New cron monitor')
    expect((wrapper.find('select#grace').element as HTMLSelectElement).value).toBe('5')
    expect((wrapper.find('select#alertLimit').element as HTMLSelectElement).value).toBe('3')
  })

  it('fills the schedule field when a quick example is clicked', async () => {
    const wrapper = mount(CronMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const exampleButton = wrapper.findAll('button').find((b) => b.text() === 'Every hour')
    await exampleButton!.trigger('click')

    expect((wrapper.find('#schedule').element as HTMLInputElement).value).toBe('every 1h')
  })

  it('shows a validation error when name is missing', async () => {
    const wrapper = mount(CronMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#schedule').setValue('every 1h')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Name is required')
    expect(createCronMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when schedule is missing', async () => {
    const wrapper = mount(CronMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Daily backup')
    await wrapper.find('form').trigger('submit.prevent')

    expect(wrapper.text()).toContain('Schedule is required')
    expect(createCronMock).not.toHaveBeenCalled()
  })

  it('creates the monitor and navigates to the detail view on success', async () => {
    createCronMock.mockResolvedValueOnce({ id: 'c1' })
    const wrapper = mount(CronMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Daily backup')
    await wrapper.find('#schedule').setValue('every 1h')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(createCronMock).toHaveBeenCalledExactlyOnceWith({
      name: 'Daily backup',
      schedule: 'every 1h',
      gracePeriodMins: 5,
      maxAlertsPerIncident: 3,
      alertAfterNFailures: 0,
      channelIds: [],
    })
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'cron-monitor-detail',
      params: { id: 'c1' },
    })
  })

  it('shows an upgrade prompt when the plan limit is reached', async () => {
    createCronMock.mockRejectedValueOnce(
      new ApiErrorMock(402, 'Upgrade to add more monitors', 'plan_limit_reached'),
    )
    const wrapper = mount(CronMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Daily backup')
    await wrapper.find('#schedule').setValue('every 1h')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Upgrade to add more monitors')
    expect(wrapper.text()).toContain('View plans')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('shows a generic error message when creation fails for another reason', async () => {
    createCronMock.mockRejectedValueOnce(new Error('Network error'))
    const wrapper = mount(CronMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.find('#name').setValue('Daily backup')
    await wrapper.find('#schedule').setValue('every 1h')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.text()).toContain('Network error')
    expect(pushMock).not.toHaveBeenCalled()
  })

  it('navigates back to the list when back or cancel is clicked', async () => {
    const wrapper = mount(CronMonitorCreateView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, '← Back')!.trigger('click')
    expect(pushMock).toHaveBeenCalledWith({ name: 'cron-monitors' })

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')
    expect(pushMock).toHaveBeenCalledWith({ name: 'cron-monitors' })
  })
})
