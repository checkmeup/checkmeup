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

import { maintenanceApi } from './maintenance'

beforeEach(() => {
  getMock.mockReset()
  postMock.mockReset()
  patchMock.mockReset()
  deleteMock.mockReset()
})

const window = {
  id: 'w1',
  title: 'Scheduled maintenance',
  message: 'We will be down briefly',
  startsAt: '2026-07-01T00:00:00Z',
  endsAt: '2026-07-01T01:00:00Z',
  status: 'upcoming' as const,
  monitorCount: 1,
  createdAt: '2026-06-20T00:00:00Z',
}

const input = {
  title: 'Scheduled maintenance',
  message: 'We will be down briefly',
  startsAt: '2026-07-01T00:00:00Z',
  endsAt: '2026-07-01T01:00:00Z' as string | null,
  monitors: [{ monitorType: 'cron' as const, monitorId: 'm1' }],
}

describe('maintenanceApi.list', () => {
  it('fetches the maintenance window list', async () => {
    getMock.mockResolvedValueOnce([window])

    const result = await maintenanceApi.list()

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/maintenance-windows/')
    expect(result).toEqual([window])
  })
})

describe('maintenanceApi.get', () => {
  it('fetches a single maintenance window by id', async () => {
    getMock.mockResolvedValueOnce(window)

    const result = await maintenanceApi.get('w1')

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/maintenance-windows/w1/')
    expect(result).toEqual(window)
  })
})

describe('maintenanceApi.create', () => {
  it('posts the input to create a maintenance window', async () => {
    postMock.mockResolvedValueOnce(window)

    const result = await maintenanceApi.create(input)

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/maintenance-windows/', input)
    expect(result).toEqual(window)
  })
})

describe('maintenanceApi.update', () => {
  it('patches the input to update a maintenance window by id', async () => {
    patchMock.mockResolvedValueOnce(window)

    const result = await maintenanceApi.update('w1', input)

    expect(patchMock).toHaveBeenCalledExactlyOnceWith('/api/v1/maintenance-windows/w1/', input)
    expect(result).toEqual(window)
  })
})

describe('maintenanceApi.delete', () => {
  it('deletes a maintenance window by id', async () => {
    deleteMock.mockResolvedValueOnce(undefined)

    await maintenanceApi.delete('w1')

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('/api/v1/maintenance-windows/w1/')
  })
})

describe('maintenanceApi.endNow', () => {
  it('posts to the end endpoint with no body', async () => {
    postMock.mockResolvedValueOnce({ ...window, status: 'ended' })

    const result = await maintenanceApi.endNow('w1')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/maintenance-windows/w1/end', {})
    expect(result).toEqual({ ...window, status: 'ended' })
  })
})
