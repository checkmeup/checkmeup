import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import NotificationChannelsCard from './NotificationChannelsCard.vue'
import type { NotificationChannel } from '@/api/notificationChannels'

const { createMock, updateMock, deleteMock, testMock, regenerateMock } = vi.hoisted(() => ({
  createMock: vi.fn(),
  updateMock: vi.fn(),
  deleteMock: vi.fn(),
  testMock: vi.fn(),
  regenerateMock: vi.fn(),
}))

vi.mock('@/api/notificationChannels', () => ({
  notificationChannelsApi: {
    create: createMock,
    update: updateMock,
    delete: deleteMock,
    test: testMock,
    regenerateWebhookSecret: regenerateMock,
  },
}))

const channelsData = ref<NotificationChannel[]>([])
const channelsPending = ref(false)
const refetchMock = vi.fn()

vi.mock('@/composables/useNotificationChannels', () => ({
  useNotificationChannels: () => ({
    data: channelsData,
    isPending: channelsPending,
    refetch: refetchMock,
  }),
}))

const telegramChannel: NotificationChannel = {
  id: 'ch1',
  type: 'telegram',
  name: 'Ops Telegram',
  config: { chatId: '-1001234567890' },
  enabled: true,
  createdAt: '2026-01-01T00:00:00Z',
}

const webhookChannel: NotificationChannel = {
  id: 'ch2',
  type: 'webhook',
  name: 'Ops webhook',
  config: { url: 'https://example.com/hooks/checkmeup', secret: 'shh-secret' },
  enabled: true,
  createdAt: '2026-01-01T00:00:00Z',
  lastDeliveryStatus: 'success',
  lastDeliveryDetail: '200 OK',
  lastDeliveryAt: new Date().toISOString(),
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

function mountCard() {
  return mount(NotificationChannelsCard)
}

beforeEach(() => {
  channelsData.value = []
  channelsPending.value = false
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('NotificationChannelsCard', () => {
  it('shows a loading state while channels are pending', () => {
    channelsPending.value = true
    const wrapper = mountCard()

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an empty state when there are no channels', () => {
    const wrapper = mountCard()

    expect(wrapper.text()).toContain('No channels connected yet.')
  })

  it('renders channel rows with type, value, and delivery summary', () => {
    channelsData.value = [telegramChannel, webhookChannel]
    const wrapper = mountCard()

    expect(wrapper.text()).toContain('Ops Telegram')
    expect(wrapper.text()).toContain('Telegram · -1001234567890')
    expect(wrapper.text()).toContain('Ops webhook')
    expect(wrapper.text()).toContain('Last delivery: success, 200 OK, just now')
  })

  it("toggles a channel's enabled state", async () => {
    channelsData.value = [telegramChannel]
    updateMock.mockResolvedValueOnce({ ...telegramChannel, enabled: false })
    const wrapper = mountCard()

    await wrapper.find('input[type="checkbox"]').trigger('change')
    await flushPromises()

    expect(updateMock).toHaveBeenCalledExactlyOnceWith('ch1', {
      type: 'telegram',
      name: 'Ops Telegram',
      config: { chatId: '-1001234567890' },
      enabled: false,
    })
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('opens the add-channel form with telegram defaults', async () => {
    const wrapper = mountCard()

    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')

    expect((wrapper.find('select#channel-type').element as HTMLSelectElement).value).toBe(
      'telegram',
    )
    expect((wrapper.find('#channel-name').element as HTMLInputElement).value).toBe('')
    expect((wrapper.find('#channel-value').element as HTMLInputElement).value).toBe('')
  })

  it('populates the form when editing a webhook channel, including its secret', async () => {
    channelsData.value = [webhookChannel]
    const wrapper = mountCard()

    await findButtonByText(wrapper, 'Edit')?.trigger('click')

    expect((wrapper.find('select#channel-type').element as HTMLSelectElement).value).toBe('webhook')
    expect((wrapper.find('#channel-name').element as HTMLInputElement).value).toBe('Ops webhook')
    expect((wrapper.find('#channel-value').element as HTMLInputElement).value).toBe(
      'https://example.com/hooks/checkmeup',
    )
    expect((wrapper.find('#channel-secret').element as HTMLInputElement).value).toBe('shh-secret')
  })

  it('shows a validation error when the name is empty', async () => {
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await wrapper.find('#channel-value').setValue('-1001234567890')

    await findButtonByText(wrapper, 'Add channel')?.trigger('click')

    expect(wrapper.text()).toContain('Name is required')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('shows a validation error when the value is empty', async () => {
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await wrapper.find('#channel-name').setValue('Ops Telegram')

    await findButtonByText(wrapper, 'Add channel')?.trigger('click')

    expect(wrapper.text()).toContain('Chat ID is required')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('rejects a webhook URL that does not start with https://', async () => {
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await wrapper.find('select#channel-type').setValue('webhook')
    await wrapper.find('#channel-name').setValue('Ops webhook')
    await wrapper.find('#channel-value').setValue('http://example.com/hook')

    await findButtonByText(wrapper, 'Add channel')?.trigger('click')

    expect(wrapper.text()).toContain('Webhook URL must start with https://')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('rejects a Slack URL that is not an Incoming Webhook URL', async () => {
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await wrapper.find('select#channel-type').setValue('slack')
    await wrapper.find('#channel-name').setValue('Ops Slack')
    await wrapper.find('#channel-value').setValue('https://example.com/not-slack')

    await findButtonByText(wrapper, 'Add channel')?.trigger('click')

    expect(wrapper.text()).toContain('Must be a Slack Incoming Webhook URL')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('creates a new channel and hides the form on success', async () => {
    createMock.mockResolvedValueOnce({ ...telegramChannel })
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await wrapper.find('#channel-name').setValue('Ops Telegram')
    await wrapper.find('#channel-value').setValue('-1001234567890')

    await findButtonByText(wrapper, 'Add channel')?.trigger('click')
    await flushPromises()

    expect(createMock).toHaveBeenCalledExactlyOnceWith({
      type: 'telegram',
      name: 'Ops Telegram',
      config: { chatId: '-1001234567890' },
    })
    expect(refetchMock).toHaveBeenCalledOnce()
    expect(wrapper.find('#channel-name').exists()).toBe(false)
  })

  it('updates an existing channel on save', async () => {
    channelsData.value = [telegramChannel]
    updateMock.mockResolvedValueOnce({ ...telegramChannel, name: 'Renamed' })
    const wrapper = mountCard()
    await findButtonByText(wrapper, 'Edit')?.trigger('click')
    await wrapper.find('#channel-name').setValue('Renamed')

    await findButtonByText(wrapper, 'Save changes')?.trigger('click')
    await flushPromises()

    expect(updateMock).toHaveBeenCalledExactlyOnceWith('ch1', {
      type: 'telegram',
      name: 'Renamed',
      config: { chatId: '-1001234567890' },
      enabled: true,
    })
  })

  it('shows a generic error message when saving fails', async () => {
    createMock.mockRejectedValueOnce(new Error('Network error'))
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await wrapper.find('#channel-name').setValue('Ops Telegram')
    await wrapper.find('#channel-value').setValue('-1001234567890')

    await findButtonByText(wrapper, 'Add channel')?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Network error')
  })

  it('disables the test button until a value is entered', async () => {
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')

    expect(findButtonByText(wrapper, 'Send test message')?.attributes('disabled')).toBeDefined()

    await wrapper.find('#channel-value').setValue('-1001234567890')

    expect(findButtonByText(wrapper, 'Send test message')?.attributes('disabled')).toBeUndefined()
  })

  it('sends a test message and shows a success message', async () => {
    testMock.mockResolvedValueOnce({})
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await wrapper.find('#channel-value').setValue('-1001234567890')

    await findButtonByText(wrapper, 'Send test message')?.trigger('click')
    await flushPromises()

    expect(testMock).toHaveBeenCalledExactlyOnceWith({
      type: 'telegram',
      config: { chatId: '-1001234567890' },
    })
    expect(wrapper.text()).toContain('Test message sent!')
  })

  it('shows a test error when the test call fails', async () => {
    testMock.mockRejectedValueOnce(new Error('bot not started'))
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await wrapper.find('#channel-value').setValue('-1001234567890')

    await findButtonByText(wrapper, 'Send test message')?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('bot not started')
  })

  it('removes a channel and shows a pending label while in flight', async () => {
    channelsData.value = [telegramChannel]
    let resolveDelete!: () => void
    deleteMock.mockReturnValueOnce(
      new Promise<void>((resolve) => {
        resolveDelete = resolve
      }),
    )
    const wrapper = mountCard()

    const removeClick = findButtonByText(wrapper, 'Remove')?.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Removing…')

    resolveDelete()
    await removeClick
    await flushPromises()

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('ch1')
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('regenerates the webhook signing secret', async () => {
    channelsData.value = [webhookChannel]
    regenerateMock.mockResolvedValueOnce({
      ...webhookChannel,
      config: { ...webhookChannel.config, secret: 'new-secret' },
    })
    const wrapper = mountCard()
    await findButtonByText(wrapper, 'Edit')?.trigger('click')

    await findButtonByText(wrapper, 'Regenerate')?.trigger('click')
    await flushPromises()

    expect(regenerateMock).toHaveBeenCalledExactlyOnceWith('ch2')
    expect((wrapper.find('#channel-secret').element as HTMLInputElement).value).toBe('new-secret')
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an error when regenerating the secret fails', async () => {
    channelsData.value = [webhookChannel]
    regenerateMock.mockRejectedValueOnce(new Error('regeneration failed'))
    const wrapper = mountCard()
    await findButtonByText(wrapper, 'Edit')?.trigger('click')

    await findButtonByText(wrapper, 'Regenerate')?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('regeneration failed')
  })

  it('cancels the form without saving', async () => {
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await wrapper.find('#channel-name').setValue('Should be discarded')

    await findButtonByText(wrapper, 'Cancel')?.trigger('click')

    expect(wrapper.find('#channel-name').exists()).toBe(false)
    expect(findButtonByText(wrapper, '+ Add channel')).toBeDefined()
    expect(createMock).not.toHaveBeenCalled()
  })
})
