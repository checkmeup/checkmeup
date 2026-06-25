import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { reactive, ref } from 'vue'
import DashboardView from './DashboardView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const authStoreMock = reactive({
  user: null as { email: string } | null,
})

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreMock,
}))

const cronData = ref<unknown[] | null>(null)
const uptimeData = ref<unknown[] | null>(null)
const sslData = ref<unknown[] | null>(null)
const domainData = ref<unknown[] | null>(null)
const statusPageData = ref<unknown[] | null>(null)
const channelData = ref<unknown[] | null>(null)

vi.mock('@/composables/useCronMonitors', () => ({
  useCronMonitors: () => ({ data: cronData }),
}))
vi.mock('@/composables/useUptimeMonitors', () => ({
  useUptimeMonitors: () => ({ data: uptimeData }),
}))
vi.mock('@/composables/useSSLMonitors', () => ({
  useSSLMonitors: () => ({ data: sslData }),
}))
vi.mock('@/composables/useDomainMonitors', () => ({
  useDomainMonitors: () => ({ data: domainData }),
}))
vi.mock('@/composables/useStatusPages', () => ({
  useStatusPages: () => ({ data: statusPageData }),
}))
vi.mock('@/composables/useNotificationChannels', () => ({
  useNotificationChannels: () => ({ data: channelData }),
}))

function findCard(wrapper: ReturnType<typeof mount>, label: string) {
  return wrapper.findAll('.rounded-xl.border').find((c) => c.text().includes(label))
}

beforeEach(() => {
  authStoreMock.user = null
  cronData.value = null
  uptimeData.value = null
  sslData.value = null
  domainData.value = null
  statusPageData.value = null
  channelData.value = null
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('DashboardView', () => {
  it('greets the user with their email when available', () => {
    authStoreMock.user = { email: 'andrew@checkmeup.net' }
    const wrapper = mount(DashboardView)

    expect(wrapper.text()).toContain('Welcome back, andrew@checkmeup.net.')
  })

  it('greets without an email when the user is not loaded', () => {
    authStoreMock.user = null
    const wrapper = mount(DashboardView)

    expect(wrapper.text()).toContain('Welcome back.')
  })

  it('shows a dash placeholder for counts that have not resolved yet', () => {
    const wrapper = mount(DashboardView)

    const cronCard = findCard(wrapper, 'Cron monitors')!
    expect(cronCard.text()).toContain('—')
  })

  it('renders independent counts for each monitor type and status pages', () => {
    cronData.value = [{ id: 'c1' }, { id: 'c2' }]
    uptimeData.value = [{ id: 'u1' }]
    sslData.value = []
    domainData.value = [{ id: 'd1' }, { id: 'd2' }, { id: 'd3' }]
    statusPageData.value = [{ id: 'sp1' }]
    const wrapper = mount(DashboardView)

    expect(findCard(wrapper, 'Cron monitors')!.text()).toContain('2')
    expect(findCard(wrapper, 'Uptime monitors')!.text()).toContain('1')
    expect(findCard(wrapper, 'SSL monitors')!.text()).toContain('0')
    expect(findCard(wrapper, 'Domain monitors')!.text()).toContain('3')
    expect(findCard(wrapper, 'Status pages')!.text()).toContain('1')
  })

  it('shows a dash for a card whose query has not resolved while others have', () => {
    cronData.value = [{ id: 'c1' }]
    uptimeData.value = null
    const wrapper = mount(DashboardView)

    expect(findCard(wrapper, 'Cron monitors')!.text()).toContain('1')
    expect(findCard(wrapper, 'Uptime monitors')!.text()).toContain('—')
  })

  it('navigates to the cron monitors list when its card is clicked', async () => {
    const wrapper = mount(DashboardView)

    await findCard(wrapper, 'Cron monitors')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'cron-monitors' })
  })

  it('navigates to the cron monitor create view from the card button without bubbling to the card click', async () => {
    const wrapper = mount(DashboardView)

    const button = findCard(wrapper, 'Cron monitors')!
      .findAll('button')
      .find((b) => b.text() === 'Add cron monitor')
    await button!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'cron-monitor-create' })
  })

  it('navigates to the status pages list when its card is clicked', async () => {
    const wrapper = mount(DashboardView)

    await findCard(wrapper, 'Status pages')!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'status-pages' })
  })

  it('navigates to the status page create view from the card button', async () => {
    const wrapper = mount(DashboardView)

    const button = findCard(wrapper, 'Status pages')!
      .findAll('button')
      .find((b) => b.text() === 'Create status page')
    await button!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'status-page-create' })
  })

  it('renders the getting started checklist', () => {
    const wrapper = mount(DashboardView)

    expect(wrapper.text()).toContain('Getting started')
    expect(wrapper.text()).toContain('Add a monitor — cron, uptime, SSL, or domain expiry')
  })
})
