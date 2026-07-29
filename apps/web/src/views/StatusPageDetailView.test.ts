import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import StatusPageDetailView from './StatusPageDetailView.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
  useRoute: () => ({ params: { id: 'sp1' } }),
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { deleteMock, setMonitorsMock } = vi.hoisted(() => ({
  deleteMock: vi.fn(),
  setMonitorsMock: vi.fn(),
}))

vi.mock('@/api/statusPages', () => ({
  statusPagesApi: { delete: deleteMock, setMonitors: setMonitorsMock },
}))

const refetchMock = vi.fn()
const pageData = ref<unknown>(null)
const pagePending = ref(false)
const pageError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useStatusPages', () => ({
  useStatusPage: () => ({
    data: pageData,
    isPending: pagePending,
    error: pageError,
    refetch: refetchMock,
  }),
}))

const cronData = ref<{ id: string; name: string }[]>([])
const uptimeData = ref<{ id: string; name: string }[]>([])
const sslData = ref<{ id: string; name: string }[]>([])
const domainData = ref<{ id: string; name: string }[]>([])
const portData = ref<{ id: string; name: string }[]>([])
const monitorsPending = ref(false)

vi.mock('@/composables/useCronMonitors', () => ({
  useCronMonitors: () => ({ data: cronData, isPending: monitorsPending }),
}))
vi.mock('@/composables/useUptimeMonitors', () => ({
  useUptimeMonitors: () => ({ data: uptimeData, isPending: monitorsPending }),
}))
vi.mock('@/composables/useSSLMonitors', () => ({
  useSSLMonitors: () => ({ data: sslData, isPending: monitorsPending }),
}))
vi.mock('@/composables/useDomainMonitors', () => ({
  useDomainMonitors: () => ({ data: domainData, isPending: monitorsPending }),
}))
vi.mock('@/composables/usePortMonitors', () => ({
  usePortMonitors: () => ({ data: portData, isPending: monitorsPending }),
}))
vi.mock('@/composables/useDNSMonitors', () => ({
  useDNSMonitors: () => ({ data: portData, isPending: monitorsPending }),
}))

const detail = {
  id: 'sp1',
  slug: 'status',
  title: 'checkmeup status',
  description: 'Live status of our services',
  logoUrl: '',
  publicUrl: 'https://checkmeup.net/status/status',
  createdAt: '2026-06-01T00:00:00Z',
  monitors: [
    {
      id: 'mi1',
      monitorType: 'cron' as const,
      monitorId: 'c1',
      displayName: 'Nightly backup',
      displayOrder: 0,
    },
  ],
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  pageData.value = null
  pagePending.value = false
  pageError.value = null
  monitorsPending.value = false
  cronData.value = [
    { id: 'c1', name: 'Nightly backup' },
    { id: 'c2', name: 'Weekly cleanup' },
  ]
  uptimeData.value = [{ id: 'u1', name: 'API uptime' }]
  sslData.value = []
  domainData.value = []
  portData.value = []
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('StatusPageDetailView', () => {
  it('shows a loading state while the page or monitors are pending', () => {
    pagePending.value = true
    const wrapper = mount(StatusPageDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message when the page fails to load', () => {
    pageError.value = { message: 'Status page not found' }
    const wrapper = mount(StatusPageDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Status page not found')
  })

  it('renders the page details, available monitors, and selected monitors', () => {
    pageData.value = { ...detail, monitors: [...detail.monitors] }
    const wrapper = mount(StatusPageDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('checkmeup status')
    expect(wrapper.text()).toContain('/status/status')
    expect(wrapper.text()).toContain('Live status of our services')
    expect(wrapper.text()).toContain('Nightly backup')
    expect(wrapper.text()).toContain('Weekly cleanup')
    expect(wrapper.text()).toContain('API uptime')
    expect(wrapper.text()).toContain('On this page (1)')
  })

  it('adds a monitor to the selected list when clicked in the available list', async () => {
    pageData.value = { ...detail, monitors: [] }
    const wrapper = mount(StatusPageDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const item = wrapper.findAll('li').find((li) => li.text().includes('Weekly cleanup'))
    await item!.trigger('click')

    expect(wrapper.text()).toContain('On this page (1)')
    const selectedNames = wrapper.findAll('input').map((i) => (i.element as HTMLInputElement).value)
    expect(selectedNames).toEqual(['Weekly cleanup'])
  })

  it('removes a monitor from the selected list when its remove button is clicked', async () => {
    pageData.value = { ...detail, monitors: [...detail.monitors] }
    const wrapper = mount(StatusPageDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('On this page (1)')
    const removeButton = findButtonByText(wrapper, '✕')
    await removeButton!.trigger('click')

    expect(wrapper.text()).toContain('On this page (0)')
    expect(wrapper.text()).toContain('Select monitors from the left.')
  })

  it('reorders selected monitors with the move up/down controls', async () => {
    pageData.value = {
      ...detail,
      monitors: [
        {
          id: 'mi1',
          monitorType: 'cron' as const,
          monitorId: 'c1',
          displayName: 'Nightly backup',
          displayOrder: 0,
        },
        {
          id: 'mi2',
          monitorType: 'uptime' as const,
          monitorId: 'u1',
          displayName: 'API uptime',
          displayOrder: 1,
        },
      ],
    }
    const wrapper = mount(StatusPageDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const inputsBefore = wrapper.findAll('input').map((i) => (i.element as HTMLInputElement).value)
    expect(inputsBefore).toEqual(['Nightly backup', 'API uptime'])

    const downButton = wrapper.findAll('button').find((b) => b.text() === '▼')
    await downButton!.trigger('click')

    const inputsAfter = wrapper.findAll('input').map((i) => (i.element as HTMLInputElement).value)
    expect(inputsAfter).toEqual(['API uptime', 'Nightly backup'])
  })

  it('saves the monitor list and refetches on success', async () => {
    pageData.value = { ...detail, monitors: [...detail.monitors] }
    setMonitorsMock.mockResolvedValueOnce([])
    const wrapper = mount(StatusPageDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const saveButton = wrapper.findAll('button').find((b) => b.text().includes('Save changes'))
    await saveButton!.trigger('click')
    await flushPromises()

    expect(setMonitorsMock).toHaveBeenCalledExactlyOnceWith('sp1', {
      monitors: [
        { monitorType: 'cron', monitorId: 'c1', displayName: 'Nightly backup', displayOrder: 0 },
      ],
    })
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an error and does not refetch when saving the monitor list fails', async () => {
    pageData.value = { ...detail, monitors: [...detail.monitors] }
    setMonitorsMock.mockRejectedValueOnce(new Error('Failed to save monitors'))
    const wrapper = mount(StatusPageDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const saveButton = wrapper.findAll('button').find((b) => b.text().includes('Save changes'))
    await saveButton!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to save monitors')
    expect(refetchMock).not.toHaveBeenCalled()
  })

  it('cancels the delete confirmation', async () => {
    pageData.value = { ...detail, monitors: [...detail.monitors] }
    const wrapper = mount(StatusPageDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    expect(findButtonByText(wrapper, 'Confirm delete')).toBeTruthy()

    await findButtonByText(wrapper, 'Cancel')!.trigger('click')
    expect(findButtonByText(wrapper, 'Delete')).toBeTruthy()
    expect(deleteMock).not.toHaveBeenCalled()
  })

  it('deletes the page and navigates back to the list on confirm', async () => {
    pageData.value = { ...detail, monitors: [...detail.monitors] }
    deleteMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(StatusPageDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('sp1')
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'status-pages' })
  })

  it('shows an error and resets confirmation state when delete fails', async () => {
    pageData.value = { ...detail, monitors: [...detail.monitors] }
    deleteMock.mockRejectedValueOnce(new Error('Delete failed'))
    const wrapper = mount(StatusPageDetailView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Delete')!.trigger('click')
    await findButtonByText(wrapper, 'Confirm delete')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Delete failed')
    expect(pushMock).not.toHaveBeenCalled()
    expect(findButtonByText(wrapper, 'Delete')).toBeTruthy()
  })

  describe('Badges (EP-30)', () => {
    it('renders a page-level badge and one per attached monitor', () => {
      pageData.value = { ...detail, monitors: [...detail.monitors] }
      const wrapper = mount(StatusPageDetailView, {
        global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
      })

      const images = wrapper.findAll('img')
      expect(images).toHaveLength(2)
      expect(images[0]!.attributes('src')).toBe('https://checkmeup.net/status/status/badge.svg')
      expect(images[1]!.attributes('src')).toBe('https://checkmeup.net/status/status/badge/c1.svg')
      expect(wrapper.text()).toContain('Overall status')
      expect(wrapper.text()).toContain('Nightly backup')
    })

    it('copies the Markdown snippet for a badge', async () => {
      pageData.value = { ...detail, monitors: [...detail.monitors] }
      const writeText = vi.fn()
      vi.stubGlobal('navigator', { ...navigator, clipboard: { writeText } })
      const wrapper = mount(StatusPageDetailView, {
        global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
      })

      const copyButtons = wrapper.findAll('button').filter((b) => b.text() === 'Copy Markdown')
      await copyButtons[0]!.trigger('click')

      expect(writeText).toHaveBeenCalledExactlyOnceWith(
        '![checkmeup status](https://checkmeup.net/status/status/badge.svg)',
      )
      expect(copyButtons[0]!.text()).toBe('Copied!')
      vi.unstubAllGlobals()
    })

    it('copies the HTML snippet for a monitor badge', async () => {
      pageData.value = { ...detail, monitors: [...detail.monitors] }
      const writeText = vi.fn()
      vi.stubGlobal('navigator', { ...navigator, clipboard: { writeText } })
      const wrapper = mount(StatusPageDetailView, {
        global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
      })

      const copyButtons = wrapper.findAll('button').filter((b) => b.text() === 'Copy HTML')
      await copyButtons[1]!.trigger('click')

      expect(writeText).toHaveBeenCalledExactlyOnceWith(
        '<a href="https://checkmeup.net/status/status"><img src="https://checkmeup.net/status/status/badge/c1.svg" alt="Nightly backup"></a>',
      )
      vi.unstubAllGlobals()
    })
  })
})
