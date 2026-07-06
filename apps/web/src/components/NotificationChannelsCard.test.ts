import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import NotificationChannelsCard from './NotificationChannelsCard.vue'
import type { NotificationChannel } from '@/api/notificationChannels'
import { ApiError } from '@/api/client'

vi.mock('vue-router', () => ({
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

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

const billingData = ref<{ plan: string } | null>(null)

vi.mock('@/composables/useBilling', () => ({
  useBilling: () => ({ data: billingData }),
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

const smsChannel: NotificationChannel = {
  id: 'ch3',
  type: 'sms',
  name: 'Ops SMS',
  config: { phone_number: '+15005550006', consent_at: '2026-01-01T00:00:00Z' },
  enabled: true,
  createdAt: '2026-01-01T00:00:00Z',
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

function mountCard() {
  return mount(NotificationChannelsCard)
}

async function selectChannelType(wrapper: ReturnType<typeof mount>, label: string) {
  const option = wrapper
    .findAll('[data-testid="channel-type-option"]')
    .find((b) => b.text() === label)
  await option?.trigger('click')
}

function selectedChannelType(wrapper: ReturnType<typeof mount>) {
  return wrapper.find('[data-testid="channel-type-option"][aria-pressed="true"]').text()
}

function channelTypeOptionLabels(wrapper: ReturnType<typeof mount>) {
  return wrapper.findAll('[data-testid="channel-type-option"]').map((b) => b.text())
}

beforeEach(() => {
  channelsData.value = []
  channelsPending.value = false
  billingData.value = { plan: 'solo' }
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

  it('shows an upgrade prompt instead of failing silently when re-enabling a channel is blocked by the plan limit', async () => {
    channelsData.value = [{ ...telegramChannel, enabled: false }]
    updateMock.mockRejectedValueOnce(
      new ApiError(
        402,
        'notification channel limit reached for your plan — upgrade to add more',
        'plan_limit_reached',
      ),
    )
    const wrapper = mountCard()

    await wrapper.find('input[type="checkbox"]').trigger('change')
    await flushPromises()

    expect(wrapper.text()).toContain('notification channel limit reached for your plan')
    expect(wrapper.text()).toContain('View plans')
    expect(refetchMock).not.toHaveBeenCalled()
  })

  it('opens the add-channel form with telegram defaults', async () => {
    const wrapper = mountCard()

    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')

    expect(selectedChannelType(wrapper)).toBe('Telegram')
    expect((wrapper.find('#channel-name').element as HTMLInputElement).value).toBe('')
    expect((wrapper.find('#channel-value').element as HTMLInputElement).value).toBe('')
  })

  it('populates the form when editing a webhook channel, including its secret', async () => {
    channelsData.value = [webhookChannel]
    const wrapper = mountCard()

    await findButtonByText(wrapper, 'Edit')?.trigger('click')

    expect(selectedChannelType(wrapper)).toBe('Webhook')
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
    await selectChannelType(wrapper, 'Webhook')
    await wrapper.find('#channel-name').setValue('Ops webhook')
    await wrapper.find('#channel-value').setValue('http://example.com/hook')

    await findButtonByText(wrapper, 'Add channel')?.trigger('click')

    expect(wrapper.text()).toContain('Webhook URL must start with https://')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('rejects a Slack URL that is not an Incoming Webhook URL', async () => {
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await selectChannelType(wrapper, 'Slack')
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

  it('rejects a phone number that is not E.164', async () => {
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await selectChannelType(wrapper, 'SMS')
    await wrapper.find('#channel-name').setValue('Ops SMS')
    await wrapper.find('#channel-value').setValue('0501234567')
    await wrapper.find('input[type="checkbox"]').setValue(true)

    await findButtonByText(wrapper, 'Add channel')?.trigger('click')

    expect(wrapper.text()).toContain('Phone number must be in E.164 format')
    expect(createMock).not.toHaveBeenCalled()
  })

  it('disables save and test until the sms consent checkbox is checked', async () => {
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await selectChannelType(wrapper, 'SMS')
    await wrapper.find('#channel-name').setValue('Ops SMS')
    await wrapper.find('#channel-value').setValue('+15005550006')

    expect(findButtonByText(wrapper, 'Send test SMS')?.attributes('disabled')).toBeDefined()
    expect(findButtonByText(wrapper, 'Add channel')?.attributes('disabled')).toBeDefined()

    await wrapper.find('input[type="checkbox"]').setValue(true)

    expect(findButtonByText(wrapper, 'Send test SMS')?.attributes('disabled')).toBeUndefined()
    expect(findButtonByText(wrapper, 'Add channel')?.attributes('disabled')).toBeUndefined()
  })

  it('creates an sms channel with consent in the config', async () => {
    createMock.mockResolvedValueOnce({ ...smsChannel })
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await selectChannelType(wrapper, 'SMS')
    await wrapper.find('#channel-name').setValue('Ops SMS')
    await wrapper.find('#channel-value').setValue('+15005550006')
    await wrapper.find('input[type="checkbox"]').setValue(true)

    await findButtonByText(wrapper, 'Add channel')?.trigger('click')
    await flushPromises()

    expect(createMock).toHaveBeenCalledExactlyOnceWith({
      type: 'sms',
      name: 'Ops SMS',
      config: { phone_number: '+15005550006', consent: 'true' },
    })
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('shows an upgrade prompt instead of a plain error when sms is blocked on the Hobby plan', async () => {
    createMock.mockRejectedValueOnce(
      new ApiError(
        402,
        'SMS alerts require a paid plan — upgrade to enable this channel',
        'plan_limit_reached',
      ),
    )
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')
    await selectChannelType(wrapper, 'SMS')
    await wrapper.find('#channel-name').setValue('Ops SMS')
    await wrapper.find('#channel-value').setValue('+15005550006')
    await wrapper.find('input[type="checkbox"]').setValue(true)

    await findButtonByText(wrapper, 'Add channel')?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('SMS alerts require a paid plan')
    expect(wrapper.text()).toContain('View plans')
  })

  it('hides the SMS option from the type picker on the Hobby plan', async () => {
    billingData.value = { plan: 'hobby' }
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')

    expect(channelTypeOptionLabels(wrapper)).not.toContain('SMS')
    expect(wrapper.text()).toContain('SMS alerts require a paid plan')
  })

  it('shows the SMS option on a paid plan', async () => {
    billingData.value = { plan: 'solo' }
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Add channel')?.trigger('click')

    expect(channelTypeOptionLabels(wrapper)).toContain('SMS')
  })

  it('still shows sms as an option when editing an existing sms channel on the Hobby plan (e.g. after a downgrade)', async () => {
    billingData.value = { plan: 'hobby' }
    channelsData.value = [smsChannel]
    const wrapper = mountCard()

    await findButtonByText(wrapper, 'Edit')?.trigger('click')

    expect(channelTypeOptionLabels(wrapper)).toContain('SMS')
    expect(selectedChannelType(wrapper)).toBe('SMS')
  })

  it('shows consent-on-file instead of a checkbox when editing an already-consented sms channel', async () => {
    channelsData.value = [smsChannel]
    const wrapper = mountCard()

    await findButtonByText(wrapper, 'Edit')?.trigger('click')

    expect(wrapper.text()).toContain('Consent given on')
    // Scoped to a label, not just any checkbox — the channel-list row above
    // the form also renders an unrelated "enabled" toggle checkbox.
    expect(wrapper.find('label input[type="checkbox"]').exists()).toBe(false)
    // Consent already on file for the unchanged number — save/test aren't blocked.
    expect(findButtonByText(wrapper, 'Save changes')?.attributes('disabled')).toBeUndefined()
  })

  it('re-requires consent when the phone number is edited away from the one on file', async () => {
    channelsData.value = [smsChannel]
    const wrapper = mountCard()
    await findButtonByText(wrapper, 'Edit')?.trigger('click')
    expect(wrapper.text()).toContain('Consent given on')

    await wrapper.find('#channel-value').setValue('+15005550001')

    expect(wrapper.text()).not.toContain('Consent given on')
    expect(wrapper.find('label input[type="checkbox"]').exists()).toBe(true)
    expect(findButtonByText(wrapper, 'Save changes')?.attributes('disabled')).toBeDefined()
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
