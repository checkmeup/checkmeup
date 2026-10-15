import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { reactive, ref } from 'vue'
import MonitorListView from './MonitorListView.vue'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeMock = reactive({ query: {} as Record<string, string> })

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock, replace: replaceMock }),
  useRoute: () => routeMock,
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const projectsData = ref<{ id: string; name: string; monitorCount: number; createdAt: string }[]>(
  [],
)
const addProjectMock = vi.fn()

vi.mock('@/composables/useProjects', () => ({
  useProjects: () => ({ data: projectsData }),
  useAddProject: () => addProjectMock,
}))

const cronData = ref<unknown[] | null>(null)
const uptimeData = ref<unknown[] | null>(null)
const sslData = ref<unknown[] | null>(null)
const pending = ref(false)
const queryError = ref<{ message: string } | null>(null)
const refetchMock = vi.fn()

vi.mock('@/composables/useCronMonitors', () => ({
  useCronMonitors: () => ({
    data: cronData,
    isPending: pending,
    error: queryError,
    refetch: refetchMock,
  }),
}))
vi.mock('@/composables/useUptimeMonitors', () => ({
  useUptimeMonitors: () => ({
    data: uptimeData,
    isPending: pending,
    error: ref(null),
    refetch: vi.fn(),
  }),
}))
vi.mock('@/composables/useSSLMonitors', () => ({
  useSSLMonitors: () => ({ data: sslData, isPending: pending, error: ref(null), refetch: vi.fn() }),
}))
// Domain, port and DNS share one list ref; these tests only need them empty
// or, for the loading case, not yet loaded.
const otherData = await vi.hoisted(async () => (await import('vue')).ref<unknown[] | null>([]))
const empty = vi.hoisted(() => () => ({ data: otherData }))
vi.mock('@/composables/useDomainMonitors', () => ({ useDomainMonitors: empty }))
vi.mock('@/composables/usePortMonitors', () => ({ usePortMonitors: empty }))
const historyData = ref<unknown>(null)
vi.mock('@/composables/useCheckHistory', () => ({
  useCheckHistory: () => ({ data: historyData }),
}))
vi.mock('@/composables/useDNSMonitors', () => ({ useDNSMonitors: empty }))

const cron = {
  id: 'c1',
  name: 'Nightly backup',
  schedule: '0 0 * * *',
  status: 'up',
  lastPingAt: new Date(Date.now() - 5 * 60000).toISOString(),
  nextPingAt: new Date(Date.now() + 55 * 60000).toISOString(),
}

const uptime = {
  id: 'u1',
  name: 'Marketing site',
  url: 'https://example.com',
  status: 'up',
  uptime24h: 99.5,
  lastCheckedAt: new Date().toISOString(),
  keyword: 'Welcome',
  keywordMode: 'contains',
}

const ssl = {
  id: 's1',
  name: 'Production API',
  hostname: 'api.example.com',
  status: 'up',
  expiresAt: '2026-12-13T00:00:00Z',
  daysUntilExpiry: 60,
  issuer: "Let's Encrypt",
  errorMsg: null,
  lastCheckedAt: null,
}

function mountView() {
  return mount(MonitorListView, {
    global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
  })
}

beforeEach(() => {
  historyData.value = null
  cronData.value = []
  uptimeData.value = []
  sslData.value = []
  otherData.value = []
  pending.value = false
  queryError.value = null
  routeMock.query = {}
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('MonitorListView', () => {
  it('shows an empty state with a call to action when there are no monitors', async () => {
    const wrapper = mountView()

    expect(wrapper.text()).toContain('No monitors yet')
    const cta = wrapper.findAll('button').filter((b) => b.text() === 'New monitor')
    await cta[cta.length - 1]!.trigger('click')
    const option = wrapper
      .findAll('[role="menuitem"]')
      .find((b) => b.text().startsWith('Cron monitor'))
    await option!.trigger('click')

    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'cron-monitor-create' })
  })

  it('shows a loading state while no monitor list has loaded', () => {
    cronData.value = null
    uptimeData.value = null
    sslData.value = null
    otherData.value = null
    pending.value = true
    const wrapper = mountView()

    expect(wrapper.text()).toContain('Loading…')
  })

  it('renders a cron row with its name, status, and schedule', () => {
    cronData.value = [cron]
    const wrapper = mountView()

    const row = wrapper.find('tbody tr')
    expect(row.text()).toContain('Nightly backup')
    expect(row.text()).toContain('Up')
    expect(row.text()).toContain('0 0 * * *')
  })

  it('renders an uptime row with status, uptime, and keyword rule', () => {
    uptimeData.value = [uptime]
    const wrapper = mountView()

    const row = wrapper.find('tbody tr')
    expect(row.text()).toContain('Marketing site')
    expect(row.text()).toContain('Up')
    expect(row.text()).toContain('99.50%')
    expect(row.text()).toContain('contains "Welcome"')
  })

  it('renders an SSL row with status, expiry, and issuer', () => {
    sslData.value = [ssl]
    const wrapper = mountView()

    const row = wrapper.find('tbody tr')
    expect(row.text()).toContain('Production API')
    expect(row.text()).toContain('Valid')
    expect(row.text()).toContain('60 days')
    expect(row.text()).toContain(new Date('2026-12-13T00:00:00Z').toLocaleDateString())
    expect(row.text()).toContain("Let's Encrypt")
  })

  it('shows "Expired" for a negative days-until-expiry even with an expiry date', () => {
    sslData.value = [{ ...ssl, status: 'expired', daysUntilExpiry: -3 }]
    const wrapper = mountView()

    expect(wrapper.find('tbody tr').text()).toContain('Expired')
  })

  it('preselects the type tab from ?type= (the old per-type list URLs)', () => {
    routeMock.query = { type: 'cron' }
    cronData.value = [cron]
    uptimeData.value = [uptime]
    const wrapper = mountView()

    expect(wrapper.find('[aria-selected="true"]').text()).toContain('Cron')
    expect(wrapper.find('tbody').text()).toContain('Nightly backup')
    expect(wrapper.find('tbody').text()).not.toContain('Marketing site')
  })

  it('writes the chosen tab into ?type=', async () => {
    cronData.value = [cron]
    uptimeData.value = [uptime]
    const wrapper = mountView()

    const tab = wrapper.findAll('[role="tab"]').find((t) => t.text().startsWith('Uptime'))
    await tab!.trigger('click')

    expect(replaceMock).toHaveBeenCalledExactlyOnceWith({ query: { type: 'uptime' } })
  })

  it('shows a load error with a retry when nothing loaded', async () => {
    queryError.value = { message: 'Network down' }
    const wrapper = mountView()

    expect(wrapper.text()).toContain('Network down')
    await wrapper
      .findAll('button')
      .find((b) => b.text() === 'Try again')!
      .trigger('click')
    expect(refetchMock).toHaveBeenCalled()
  })

  describe('projects (PRD projects)', () => {
    beforeEach(() => {
      projectsData.value = [
        { id: 'p1', name: 'Acme', monitorCount: 2, createdAt: '' },
        { id: 'p2', name: 'Globex', monitorCount: 0, createdAt: '' },
      ]
      cronData.value = [{ ...cron, status: 'down', projectId: 'p1' }]
      uptimeData.value = [{ ...uptime, projectId: 'p1' }]
      sslData.value = [{ ...ssl, projectId: null }]
    })
    afterEach(() => {
      projectsData.value = []
    })

    // The desktop table's group header rows, in order.
    function groupHeaders(wrapper: ReturnType<typeof mountView>) {
      return wrapper.findAll('tbody button[aria-expanded]').map((b) => b.text())
    }

    it('groups monitors by project with "No project" last and a summary of what needs attention', () => {
      const wrapper = mountView()
      const headers = groupHeaders(wrapper)
      expect(headers).toHaveLength(2)
      expect(headers[0]).toContain('Acme')
      expect(headers[0]).toContain('2 monitors')
      expect(headers[0]).toContain('1 need attention')
      expect(headers[1]).toContain('No project')
      expect(headers[1]).not.toContain('need attention')
    })

    it('collapses and expands a project group', async () => {
      const wrapper = mountView()
      const acme = wrapper.find('tbody button[aria-expanded]')
      expect(wrapper.find('table').text()).toContain('Nightly backup')
      await acme.trigger('click')
      expect(acme.attributes('aria-expanded')).toBe('false')
      expect(wrapper.find('table').text()).not.toContain('Nightly backup')
      expect(wrapper.find('table').text()).toContain('Production API')
      await acme.trigger('click')
      expect(wrapper.find('table').text()).toContain('Nightly backup')
    })

    it("shows only one project's monitors, ungrouped, for ?project=", () => {
      routeMock.query = { project: 'p1' }
      const wrapper = mountView()
      const table = wrapper.find('table').text()
      expect(groupHeaders(wrapper)).toHaveLength(0)
      expect(table).toContain('Nightly backup')
      expect(table).toContain('Marketing site')
      expect(table).not.toContain('Production API')
    })

    it('shows the monitors without a project for ?project=none', () => {
      routeMock.query = { project: 'none' }
      const table = mountView().find('table').text()
      expect(table).toContain('Production API')
      expect(table).not.toContain('Nightly backup')
    })

    it('writes the chosen project into ?project=, keeping ?type=', async () => {
      routeMock.query = { type: 'cron' }
      const wrapper = mountView()
      await wrapper.find('select').setValue('p2')
      expect(replaceMock).toHaveBeenCalledExactlyOnceWith({
        query: { type: 'cron', project: 'p2' },
      })
    })

    it('counts the tabs within the chosen project', () => {
      routeMock.query = { project: 'p1' }
      const tabs = mountView()
        .findAll('[role="tab"]')
        .map((t) => t.text())
      expect(tabs).toContain('All 2')
      expect(tabs).toContain('Problems 1')
      expect(tabs.some((t) => t.startsWith('SSL'))).toBe(false)
    })

    it("ignores a ?project= that is not one of the org's projects, rather than hiding every monitor", () => {
      routeMock.query = { project: 'deleted-project' }
      const wrapper = mountView()
      expect(groupHeaders(wrapper)).toHaveLength(2)
      expect((wrapper.find('select').element as HTMLSelectElement).value).toBe('')
    })

    it('counts a monitor whose project is gone as "No project" in the filter too', () => {
      sslData.value = [{ ...ssl, projectId: 'deleted-project' }]
      routeMock.query = { project: 'none' }
      expect(mountView().find('table').text()).toContain('Production API')
    })

    it('stays a flat list with no project filter when the org has no projects', () => {
      projectsData.value = []
      const wrapper = mountView()
      expect(groupHeaders(wrapper)).toHaveLength(0)
      expect(wrapper.find('select').exists()).toBe(false)
    })
  })
})
