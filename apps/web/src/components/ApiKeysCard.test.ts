import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import ApiKeysCard from './ApiKeysCard.vue'
import type { ApiKey } from '@/api/apiKeys'

const { createMock, revokeMock } = vi.hoisted(() => ({
  createMock: vi.fn(),
  revokeMock: vi.fn(),
}))

vi.mock('@/api/apiKeys', () => ({
  apiKeysApi: {
    create: createMock,
    revoke: revokeMock,
  },
}))

const keysData = ref<ApiKey[]>([])
const keysPending = ref(false)
const refetchMock = vi.fn()

vi.mock('@/composables/useApiKeys', () => ({
  useApiKeys: () => ({
    data: keysData,
    isPending: keysPending,
    refetch: refetchMock,
  }),
}))

const key1: ApiKey = {
  id: 'key1',
  label: 'CI integration',
  keyPrefix: 'cmu_live_a1b2c3d',
  createdAt: '2026-01-01T00:00:00Z',
  lastUsedAt: null,
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

function mountCard() {
  return mount(ApiKeysCard)
}

beforeEach(() => {
  keysData.value = []
  keysPending.value = false
})

afterEach(() => {
  vi.clearAllMocks()
})

describe('ApiKeysCard', () => {
  it('shows a loading state while keys are pending', () => {
    keysPending.value = true
    const wrapper = mountCard()

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an empty state when there are no keys', () => {
    const wrapper = mountCard()

    expect(wrapper.text()).toContain('No API keys yet.')
  })

  it('renders key rows with label and masked prefix', () => {
    keysData.value = [key1]
    const wrapper = mountCard()

    expect(wrapper.text()).toContain('CI integration')
    expect(wrapper.text()).toContain('cmu_live_a1b2c3d…')
    expect(wrapper.text()).toContain('Never used')
  })

  it('opens the generate-key form', async () => {
    const wrapper = mountCard()

    await findButtonByText(wrapper, '+ Generate key')?.trigger('click')

    expect(wrapper.find('#key-label').exists()).toBe(true)
  })

  it('creates a key and shows the raw value once', async () => {
    createMock.mockResolvedValueOnce({ ...key1, key: 'cmu_live_a1b2c3d4e5f6' })
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Generate key')?.trigger('click')
    await wrapper.find('#key-label').setValue('CI integration')

    await findButtonByText(wrapper, 'Generate key')?.trigger('click')
    await flushPromises()

    expect(createMock).toHaveBeenCalledExactlyOnceWith('CI integration')
    expect(refetchMock).toHaveBeenCalledOnce()
    expect(wrapper.text()).toContain('cmu_live_a1b2c3d4e5f6')
    expect(wrapper.text()).toContain("Copy this key now — you won't be able to see it again.")
    expect(wrapper.find('#key-label').exists()).toBe(false)
  })

  it('copies the created key to the clipboard', async () => {
    const writeText = vi.fn()
    vi.stubGlobal('navigator', { ...navigator, clipboard: { writeText } })
    createMock.mockResolvedValueOnce({ ...key1, key: 'cmu_live_a1b2c3d4e5f6' })
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Generate key')?.trigger('click')
    await findButtonByText(wrapper, 'Generate key')?.trigger('click')
    await flushPromises()

    await findButtonByText(wrapper, 'Copy')?.trigger('click')

    expect(writeText).toHaveBeenCalledExactlyOnceWith('cmu_live_a1b2c3d4e5f6')
    expect(wrapper.text()).toContain('Copied!')
  })

  it('dismisses the created-key banner', async () => {
    createMock.mockResolvedValueOnce({ ...key1, key: 'cmu_live_a1b2c3d4e5f6' })
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Generate key')?.trigger('click')
    await findButtonByText(wrapper, 'Generate key')?.trigger('click')
    await flushPromises()

    await findButtonByText(wrapper, 'Done')?.trigger('click')

    expect(wrapper.text()).not.toContain('cmu_live_a1b2c3d4e5f6')
  })

  it('shows a generic error message when creation fails', async () => {
    createMock.mockRejectedValueOnce(new Error('Network error'))
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Generate key')?.trigger('click')

    await findButtonByText(wrapper, 'Generate key')?.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Network error')
  })

  it('revokes a key and shows a pending label while in flight', async () => {
    keysData.value = [key1]
    let resolveRevoke!: () => void
    revokeMock.mockReturnValueOnce(
      new Promise<void>((resolve) => {
        resolveRevoke = resolve
      }),
    )
    const wrapper = mountCard()

    const revokeClick = findButtonByText(wrapper, 'Revoke')?.trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Revoking…')

    resolveRevoke()
    await revokeClick
    await flushPromises()

    expect(revokeMock).toHaveBeenCalledExactlyOnceWith('key1')
    expect(refetchMock).toHaveBeenCalledOnce()
  })

  it('cancels the form without creating a key', async () => {
    const wrapper = mountCard()
    await findButtonByText(wrapper, '+ Generate key')?.trigger('click')
    await wrapper.find('#key-label').setValue('Should be discarded')

    await findButtonByText(wrapper, 'Cancel')?.trigger('click')

    expect(wrapper.find('#key-label').exists()).toBe(false)
    expect(findButtonByText(wrapper, '+ Generate key')).toBeDefined()
    expect(createMock).not.toHaveBeenCalled()
  })
})
