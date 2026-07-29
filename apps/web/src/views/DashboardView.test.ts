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
const portData = ref<unknown[] | null>(null)
const dnsData = ref<unknown[] | null>(null)

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
vi.mock('@/composables/usePortMonitors', () => ({
  usePortMonitors: () => ({ data: portData }),
}))
vi.mock('@/composables/useDNSMonitors', () => ({
  useDNSMonitors: () => ({ data: dnsData }),
}))

const billingData = ref<{ smsCreditsUsed: number; smsCreditsLimit: number } | null>(null)

vi.mock('@/composables/useBilling', () => ({
  useBilling: () => ({ data: billingData }),
}))

function uptimeMonitor(overrides: Record<string, unknown> = {}) {
  return {
    id: 'u1',
    name: 'checkmeup.net',
    url: 'https://checkmeup.net',
    status: 'up',
    uptime24h: 99.9,
    lastCheckedAt: '2026-07-06T12:00:00Z',
    ...overrides,
  }
}

function sslMonitor(overrides: Record<string, unknown> = {}) {
  return {
    id: 's1',
    name: 'checkmeup.net',
    hostname: 'checkmeup.net',
    status: 'up',
    daysUntilExpiry: 84,
    lastCheckedAt: '2026-07-06T12:00:00Z',
    ...overrides,
  }
}

function domainMonitor(overrides: Record<string, unknown> = {}) {
  return {
    id: 'd1',
    name: 'checkmeup.net',
    domain: 'checkmeup.net',
    status: 'up',
    daysUntilExpiry: 200,
    lastCheckedAt: '2026-07-06T12:00:00Z',
    ...overrides,
  }
}

function cronMonitor(overrides: Record<string, unknown> = {}) {
  return {
    id: 'c1',
    name: 'Nightly Backup',
    schedule: '0 2 * * *',
    status: 'up',
    nextPingAt: null,
    lastPingAt: null,
    ...overrides,
  }
}

function portMonitor(overrides: Record<string, unknown> = {}) {
  return {
    id: 'p1',
    name: 'db-primary',
    host: '10.0.4.12',
    port: 5432,
    status: 'up',
    uptime24h: 100,
    lastCheckedAt: '2026-07-06T12:00:00Z',
    ...overrides,
  }
}

function dnsMonitor(overrides: Record<string, unknown> = {}) {
  return {
    id: 'd1',
    name: 'Apex A record',
    hostname: 'checkmeup.net',
    recordType: 'A',
    status: 'up',
    uptime24h: 100,
    lastCheckedAt: '2026-07-06T12:00:00Z',
    ...overrides,
  }
}

beforeEach(() => {
  authStoreMock.user = null
  cronData.value = null
  uptimeData.value = null
  sslData.value = null
  domainData.value = null
  portData.value = null
  dnsData.value = null
  billingData.value = null
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

  it('shows the empty state when there are no monitors of any type', () => {
    const wrapper = mount(DashboardView)

    expect(wrapper.text()).toContain('No monitors yet')
    expect(wrapper.text()).toContain('0/ 0')
  })

  it('computes the healthy/total hero stat across monitor types', () => {
    uptimeData.value = [
      uptimeMonitor({ status: 'up' }),
      uptimeMonitor({ id: 'u2', status: 'down' }),
    ]
    cronData.value = [cronMonitor({ status: 'up' })]
    const wrapper = mount(DashboardView)

    expect(wrapper.text()).toContain('2/ 3')
    expect(wrapper.text()).toContain('66.7% healthy')
  })

  it('shows SMS credits used/limit when the plan has a limit', () => {
    billingData.value = { smsCreditsUsed: 3, smsCreditsLimit: 10 }
    const wrapper = mount(DashboardView)

    expect(wrapper.text()).toContain('3')
    expect(wrapper.text()).toContain('/ 10')
  })

  it('shows "Not available on your plan" for SMS credits on a 0-limit plan (Hobby)', () => {
    billingData.value = { smsCreditsUsed: 10, smsCreditsLimit: 0 }
    const wrapper = mount(DashboardView)

    expect(wrapper.text()).toContain('Not available on your plan')
    expect(wrapper.text()).not.toContain('10 / 0')
  })

  it('surfaces a down monitor in the needs-attention banner', () => {
    uptimeData.value = [uptimeMonitor({ status: 'down', name: 'api.acmecorp.com' })]
    const wrapper = mount(DashboardView)

    expect(wrapper.text()).toContain('Needs attention')
    expect(wrapper.text()).toContain('api.acmecorp.com is down')
  })

  it('surfaces an expiring-soon SSL monitor in the needs-attention banner', () => {
    sslData.value = [
      sslMonitor({ status: 'expiring_soon', daysUntilExpiry: 6, name: 'client-shop.com' }),
    ]
    const wrapper = mount(DashboardView)

    expect(wrapper.text()).toContain('client-shop.com')
    expect(wrapper.text()).toContain('Renew')
  })

  it('shows no attention banner when every monitor is healthy', () => {
    uptimeData.value = [uptimeMonitor({ status: 'up' })]
    const wrapper = mount(DashboardView)

    // "Needs attention" also labels the always-present hero stat card, so
    // absence of the banner is asserted via its distinguishing content
    // (an item's action link) rather than that shared heading text.
    expect(wrapper.text()).not.toContain('Investigate →')
    expect(wrapper.text()).toContain('All clear')
  })

  it('renders a row per monitor across all types with its type badge', () => {
    uptimeData.value = [uptimeMonitor()]
    cronData.value = [cronMonitor()]
    sslData.value = [sslMonitor()]
    domainData.value = [domainMonitor()]
    portData.value = [portMonitor()]
    dnsData.value = [dnsMonitor()]
    const wrapper = mount(DashboardView)

    expect(wrapper.text()).toContain('checkmeup.net')
    expect(wrapper.text()).toContain('Nightly Backup')
    expect(wrapper.text()).toContain('db-primary')
    expect(wrapper.text()).toContain('Apex A record')
    expect(wrapper.findAll('tbody tr').length).toBe(6)
  })

  it('filters the monitors table when a type chip is clicked', async () => {
    uptimeData.value = [uptimeMonitor({ name: 'uptime-one' })]
    cronData.value = [cronMonitor({ name: 'cron-one' })]
    const wrapper = mount(DashboardView)

    expect(wrapper.text()).toContain('uptime-one')
    expect(wrapper.text()).toContain('cron-one')

    const chips = wrapper.findAll('button').filter((b) => b.text().startsWith('Cron'))
    await chips[0]!.trigger('click')

    expect(wrapper.text()).toContain('cron-one')
    expect(wrapper.text()).not.toContain('uptime-one')
  })

  it('navigates to a monitor detail route when its table row is clicked', async () => {
    uptimeData.value = [uptimeMonitor({ id: 'u42', name: 'Primary Website Uptime' })]
    const wrapper = mount(DashboardView)

    const row = wrapper.findAll('tbody tr').find((r) => r.text().includes('Primary Website Uptime'))
    await row!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'uptime-monitor-detail',
      params: { id: 'u42' },
    })
  })

  it('opens the add-monitor menu and navigates to the chosen monitor type', async () => {
    const wrapper = mount(DashboardView)

    const addButton = wrapper.findAll('button').find((b) => b.text().includes('Add monitor'))
    await addButton!.trigger('click')

    const option = wrapper.findAll('div').find((d) => d.text().trim() === 'Uptime monitor')
    await option!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'uptime-monitor-create' })
  })

  it('prompts to add an SSL/domain monitor when neither type exists yet', () => {
    const wrapper = mount(DashboardView)

    expect(wrapper.text()).toContain('Add an SSL or domain monitor')
  })

  it('lists upcoming expirations sorted soonest-first', () => {
    sslData.value = [sslMonitor({ name: 'far-out', daysUntilExpiry: 300 })]
    domainData.value = [domainMonitor({ name: 'soon', daysUntilExpiry: 5 })]
    const wrapper = mount(DashboardView)

    // Both names also appear in the (unsorted) monitors table above the
    // panel, so scope the ordering check to the panel's own text.
    const panelText = wrapper.text().split('Upcoming expirations')[1]!
    expect(panelText.indexOf('soon')).toBeLessThan(panelText.indexOf('far-out'))
  })
})
