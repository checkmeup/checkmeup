import { beforeEach, describe, expect, it, vi } from 'vitest'

const { getMock, postMock, patchMock, deleteMock } = vi.hoisted(() => ({
  getMock: vi.fn(),
  postMock: vi.fn(),
  patchMock: vi.fn(),
  deleteMock: vi.fn(),
}))

vi.mock('./client', () => ({
  api: { get: getMock, post: postMock, patch: patchMock, delete: deleteMock },
}))

import { notificationChannelsApi } from './notificationChannels'

beforeEach(() => {
  getMock.mockReset()
  postMock.mockReset()
  patchMock.mockReset()
  deleteMock.mockReset()
})

const channel = {
  id: 'ch1',
  type: 'telegram' as const,
  name: 'On-call Telegram',
  config: { chatId: '12345' },
  enabled: true,
  createdAt: '2026-06-01T00:00:00Z',
}

const input = {
  type: 'telegram' as const,
  name: 'On-call Telegram',
  config: { chatId: '12345' },
  enabled: true,
}

describe('notificationChannelsApi.list', () => {
  it('fetches the notification channel list', async () => {
    getMock.mockResolvedValueOnce([channel])

    const result = await notificationChannelsApi.list()

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/notification-channels/')
    expect(result).toEqual([channel])
  })
})

describe('notificationChannelsApi.create', () => {
  it('posts the input to create a notification channel', async () => {
    postMock.mockResolvedValueOnce(channel)

    const result = await notificationChannelsApi.create(input)

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/notification-channels/', input)
    expect(result).toEqual(channel)
  })
})

describe('notificationChannelsApi.update', () => {
  it('patches the input to update a notification channel by id', async () => {
    patchMock.mockResolvedValueOnce(channel)

    const result = await notificationChannelsApi.update('ch1', input)

    expect(patchMock).toHaveBeenCalledExactlyOnceWith('/api/v1/notification-channels/ch1/', input)
    expect(result).toEqual(channel)
  })
})

describe('notificationChannelsApi.delete', () => {
  it('deletes a notification channel by id', async () => {
    deleteMock.mockResolvedValueOnce(undefined)

    await notificationChannelsApi.delete('ch1')

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('/api/v1/notification-channels/ch1/')
  })
})

describe('notificationChannelsApi.test', () => {
  it('posts the type and config to the test endpoint', async () => {
    const testInput = { type: 'webhook' as const, config: { url: 'https://example.com/hook' } }
    postMock.mockResolvedValueOnce(undefined)

    await notificationChannelsApi.test(testInput)

    expect(postMock).toHaveBeenCalledExactlyOnceWith(
      '/api/v1/notification-channels/test',
      testInput,
    )
  })
})

describe('notificationChannelsApi.regenerateWebhookSecret', () => {
  it('posts to the regenerate-secret endpoint with a null body', async () => {
    const regenerated = { ...channel, type: 'webhook' as const }
    postMock.mockResolvedValueOnce(regenerated)

    const result = await notificationChannelsApi.regenerateWebhookSecret('ch1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith(
      '/api/v1/notification-channels/ch1/regenerate-secret',
      null,
    )
    expect(result).toEqual(regenerated)
  })
})
