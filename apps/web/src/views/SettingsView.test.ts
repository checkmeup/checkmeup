import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { reactive, ref } from 'vue'
import SettingsView from './SettingsView.vue'

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

vi.mock('@/components/NotificationChannelsCard.vue', () => ({
  default: { name: 'NotificationChannelsCard', template: '<div />' },
}))

const authStoreMock = reactive({
  user: null as {
    email: string
    termsAcceptedAt: string | null
    termsVersion: string | null
  } | null,
})

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreMock,
}))

const { submitMock } = vi.hoisted(() => ({
  submitMock: vi.fn(),
}))

vi.mock('@/api/suggestions', () => ({
  suggestionsApi: { submit: submitMock },
}))

const themeRef = ref<'light' | 'dark'>('dark')
const setThemeMock = vi.fn((value: 'light' | 'dark') => {
  themeRef.value = value
})

vi.mock('@/lib/theme', () => ({
  useTheme: () => ({ theme: themeRef, setTheme: setThemeMock }),
}))

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  authStoreMock.user = { email: 'andrew@checkmeup.net', termsAcceptedAt: null, termsVersion: null }
  themeRef.value = 'dark'
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('SettingsView', () => {
  it('renders the appearance section with the current theme highlighted', () => {
    const wrapper = mount(SettingsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Appearance')
    expect(findButtonByText(wrapper, 'Dark')).toBeTruthy()
    expect(findButtonByText(wrapper, 'Light')).toBeTruthy()
  })

  it('switches the theme when a theme button is clicked', async () => {
    const wrapper = mount(SettingsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await findButtonByText(wrapper, 'Light')!.trigger('click')

    expect(setThemeMock).toHaveBeenCalledExactlyOnceWith('light')
  })

  it('renders the notification channels card', () => {
    const wrapper = mount(SettingsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.findComponent({ name: 'NotificationChannelsCard' }).exists()).toBe(true)
  })

  it('does not show terms acceptance text when not yet accepted', () => {
    authStoreMock.user = {
      email: 'andrew@checkmeup.net',
      termsAcceptedAt: null,
      termsVersion: null,
    }
    const wrapper = mount(SettingsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).not.toContain('You accepted the')
  })

  it('shows the formatted terms acceptance date when accepted', () => {
    authStoreMock.user = {
      email: 'andrew@checkmeup.net',
      termsAcceptedAt: '2026-06-01T00:00:00Z',
      termsVersion: '1.0',
    }
    const wrapper = mount(SettingsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('You accepted the')
    expect(wrapper.text()).toContain('version 1.0')
    expect(wrapper.text()).toContain('Jun 1, 2026')
  })

  it('shows the user email next to the suggestion box', () => {
    const wrapper = mount(SettingsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    expect(wrapper.text()).toContain('Sent as andrew@checkmeup.net')
  })

  it('disables the send button until the suggestion has content', async () => {
    const wrapper = mount(SettingsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    const sendButton = findButtonByText(wrapper, 'Send suggestion')!
    expect(sendButton.attributes('disabled')).toBeDefined()

    await wrapper.get('#suggestion').setValue('Add dark mode scheduling')
    expect(findButtonByText(wrapper, 'Send suggestion')!.attributes('disabled')).toBeUndefined()
  })

  it('submits the suggestion and shows a confirmation on success', async () => {
    submitMock.mockResolvedValueOnce(undefined)
    const wrapper = mount(SettingsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.get('#suggestion').setValue('Add dark mode scheduling')
    await findButtonByText(wrapper, 'Send suggestion')!.trigger('click')
    await flushPromises()

    expect(submitMock).toHaveBeenCalledExactlyOnceWith('Add dark mode scheduling')
    expect(wrapper.text()).toContain('Thanks — this reaches an engineer directly.')
    expect((wrapper.get('#suggestion').element as HTMLTextAreaElement).value).toBe('')
  })

  it('shows an error message when submitting the suggestion fails', async () => {
    submitMock.mockRejectedValueOnce(new Error('Failed to send'))
    const wrapper = mount(SettingsView, {
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
    })

    await wrapper.get('#suggestion').setValue('Add dark mode scheduling')
    await findButtonByText(wrapper, 'Send suggestion')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to send')
    expect(wrapper.text()).not.toContain('Thanks — this reaches an engineer directly.')
  })
})
