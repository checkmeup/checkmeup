import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import DashboardView from './DashboardView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const projectsData = ref<{ id: string; name: string; monitorCount: number; createdAt: string }[]>([])
const addProjectMock = vi.fn()

vi.mock('@/composables/useProjects', () => ({
  useProjects: () => ({ data: projectsData }),
  useAddProject: () => addProjectMock,
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
const historyData = ref<unknown>(null)
vi.mock('@/composables/useCheckHistory', () => ({
  useCheckHistory: () => ({ data: historyData }),
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
  historyData.value = null
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

// RouterLink renders as a plain anchor — the router itself is mocked out.
function mountView() {
  return mount(DashboardView, {
    global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
  })
}

describe('DashboardView', () => {
  it('shows the empty state when there are no monitors of any type', () => {
    const wrapper = mountView()

    expect(wrapper.text()).toContain('No monitors yet')
    expect(wrapper.find('table').exists()).toBe(false)
  })

  it('says every monitor is fine when none needs attention', () => {
    uptimeData.value = [uptimeMonitor({ status: 'up' }), uptimeMonitor({ id: 'u2' })]
    const wrapper = mountView()

    expect(wrapper.find('h1').text()).toBe('All 2 monitors are fine.')
    expect(wrapper.text()).not.toContain('Needs attention')
  })

  it('counts the monitors that need attention in the headline', () => {
    uptimeData.value = [
      uptimeMonitor({ status: 'up' }),
      uptimeMonitor({ id: 'u2', status: 'down' }),
    ]
    sslData.value = [sslMonitor({ status: 'expiring_soon', daysUntilExpiry: 6 })]
    cronData.value = [cronMonitor({ status: 'up' })]
    const wrapper = mountView()

    expect(wrapper.find('h1').text()).toContain('2 monitors need you.')
    expect(wrapper.text()).toContain('The other 2 are fine.')
  })

  it('summarizes average uptime and up count under the headline', () => {
    uptimeData.value = [
      uptimeMonitor({ uptime24h: 99 }),
      uptimeMonitor({ id: 'u2', status: 'down', uptime24h: 97 }),
    ]
    cronData.value = [cronMonitor({ status: 'up' })]
    const wrapper = mountView()

    expect(wrapper.text()).toContain('98.00% average uptime today')
    expect(wrapper.text()).toContain('2 of 3 up')
  })

  it('counts paused monitors on their own, not as fine or in the uptime average', () => {
    uptimeData.value = [
      uptimeMonitor({ status: 'up', uptime24h: 100 }),
      uptimeMonitor({ id: 'u2', status: 'down', uptime24h: 90 }),
      uptimeMonitor({ id: 'u3', status: 'paused', uptime24h: 0 }),
    ]
    const wrapper = mountView()

    expect(wrapper.find('h1').text()).toContain('1 monitor needs you.')
    expect(wrapper.find('h1').text()).toBe('1 monitor needs you.')
    expect(wrapper.text()).toContain('The other one is fine. 1 is paused.')
    expect(wrapper.text()).toContain('95.00% average uptime today')
    expect(wrapper.text()).toContain('1 of 2 up · 1 paused')
  })

  it('says so when every monitor is paused', () => {
    uptimeData.value = [uptimeMonitor({ status: 'paused' })]
    const wrapper = mountView()

    expect(wrapper.find('h1').text()).toBe('All your monitors are paused.')
  })

  it('shows SMS credits used/limit when the plan has a limit', () => {
    uptimeData.value = [uptimeMonitor()]
    billingData.value = { smsCreditsUsed: 3, smsCreditsLimit: 10 }
    const wrapper = mountView()

    expect(wrapper.text()).toContain('3 of 10 SMS credits used')
  })

  it('leaves SMS credits out on a 0-limit plan (Hobby)', () => {
    uptimeData.value = [uptimeMonitor()]
    billingData.value = { smsCreditsUsed: 10, smsCreditsLimit: 0 }
    const wrapper = mountView()

    expect(wrapper.text()).not.toContain('SMS credits')
  })

  it('surfaces a down monitor under needs attention', () => {
    uptimeData.value = [uptimeMonitor({ status: 'down', name: 'api.acmecorp.com' })]
    const wrapper = mountView()

    const panel = wrapper.find('section[aria-labelledby="attention-heading"]')
    expect(panel.text()).toContain('api.acmecorp.com')
    expect(panel.text()).toContain('Down — the last check of https://checkmeup.net failed.')
  })

  it("names a problem monitor's project under needs attention (PRD projects)", () => {
    projectsData.value = [{ id: 'p1', name: 'Acme', monitorCount: 1, createdAt: '' }]
    uptimeData.value = [uptimeMonitor({ status: 'down', name: 'api.acmecorp.com', projectId: 'p1' })]
    const wrapper = mountView()

    const attention = wrapper.find('section[aria-labelledby="attention-heading"]')
    expect(attention.text()).toContain('Acme')
    expect(attention.text()).toContain('api.acmecorp.com')
    projectsData.value = []
  })

  it('surfaces an expiring-soon SSL monitor with a renew action', () => {
    sslData.value = [
      sslMonitor({ status: 'expiring_soon', daysUntilExpiry: 6, name: 'client-shop.com' }),
    ]
    const wrapper = mountView()

    const panel = wrapper.find('section[aria-labelledby="attention-heading"]')
    expect(panel.text()).toContain('client-shop.com')
    expect(panel.text()).toContain('Certificate expires in 6 days.')
    expect(panel.text()).toContain('Renew')
  })

  it('renders a row per monitor across all types with its type', () => {
    uptimeData.value = [uptimeMonitor()]
    cronData.value = [cronMonitor()]
    sslData.value = [sslMonitor()]
    domainData.value = [domainMonitor()]
    portData.value = [portMonitor()]
    dnsData.value = [dnsMonitor()]
    const wrapper = mountView()

    const rows = wrapper.findAll('tbody tr')
    expect(rows.length).toBe(6)
    expect(wrapper.find('tbody').text()).toContain('Nightly Backup')
    expect(wrapper.find('tbody').text()).toContain('db-primary')
    expect(wrapper.find('tbody').text()).toContain('Apex A record')
    expect(wrapper.find('tbody').text()).toContain('DNS')
  })

  it('lists problems first in the monitors table', () => {
    uptimeData.value = [uptimeMonitor({ name: 'a-healthy' })]
    portData.value = [portMonitor({ name: 'z-broken', status: 'down' })]
    const wrapper = mountView()

    const rows = wrapper.findAll('tbody tr')
    expect(rows[0]!.text()).toContain('z-broken')
    expect(rows[1]!.text()).toContain('a-healthy')
  })

  it('filters the monitors table when a type tab is clicked', async () => {
    uptimeData.value = [uptimeMonitor({ name: 'uptime-one' })]
    cronData.value = [cronMonitor({ name: 'cron-one' })]
    const wrapper = mountView()

    const tab = wrapper.findAll('[role="tab"]').find((b) => b.text().startsWith('Cron'))
    await tab!.trigger('click')

    expect(wrapper.find('tbody').text()).toContain('cron-one')
    expect(wrapper.find('tbody').text()).not.toContain('uptime-one')
  })

  it('narrows the monitors table by search text', async () => {
    uptimeData.value = [uptimeMonitor({ name: 'shop', url: 'https://shop.example.com' })]
    cronData.value = [cronMonitor({ name: 'backup' })]
    const wrapper = mountView()

    await wrapper.find('input[type="search"]').setValue('shop')

    expect(wrapper.findAll('tbody tr').length).toBe(1)
    expect(wrapper.find('tbody').text()).toContain('shop')
  })

  it('navigates to a monitor detail route when its table row is clicked', async () => {
    uptimeData.value = [uptimeMonitor({ id: 'u42', name: 'Primary Website Uptime' })]
    const wrapper = mountView()

    const row = wrapper.findAll('tbody tr').find((r) => r.text().includes('Primary Website Uptime'))
    await row!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({
      name: 'uptime-monitor-detail',
      params: { id: 'u42' },
    })
  })

  it('opens the new-monitor menu and navigates to the chosen monitor type', async () => {
    const wrapper = mountView()

    const addButton = wrapper.findAll('button').find((b) => b.text() === 'New monitor')
    await addButton!.trigger('click')

    const option = wrapper
      .findAll('[role="menuitem"]')
      .find((d) => d.text().startsWith('Uptime monitor'))
    await option!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'uptime-monitor-create' })
  })

  it('prompts to add an SSL/domain monitor when neither type exists yet', () => {
    uptimeData.value = [uptimeMonitor()]
    const wrapper = mountView()

    expect(wrapper.text()).toContain('Add an SSL or domain monitor')
  })

  it('lists upcoming expirations sorted soonest-first', () => {
    sslData.value = [sslMonitor({ name: 'far-out', daysUntilExpiry: 300 })]
    domainData.value = [domainMonitor({ name: 'soon', daysUntilExpiry: 5 })]
    const wrapper = mountView()

    const panelText = wrapper.find('section[aria-labelledby="expirations-heading"]').text()
    expect(panelText.indexOf('soon')).toBeLessThan(panelText.indexOf('far-out'))
  })

  it('draws the last 48 hours as a check strip for monitors with history', () => {
    uptimeData.value = [uptimeMonitor({ id: 'u1', name: 'with-history' })]
    sslData.value = [sslMonitor({ name: 'no-history' })]
    historyData.value = {
      hours: 3,
      monitors: [{ type: 'uptime', id: 'u1', ticks: ['up', null, 'down'] }],
    }
    const wrapper = mountView()

    const rows = wrapper.findAll('tbody tr')
    const withHistory = rows.find((r) => r.text().includes('with-history'))!
    expect(withHistory.find('[aria-label="1 of the last 2 checks failed"]').exists()).toBe(true)
    const withoutHistory = rows.find((r) => r.text().includes('no-history'))!
    expect(withoutHistory.find('[role="img"]').exists()).toBe(false)
  })
})
