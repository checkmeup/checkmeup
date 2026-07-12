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

import { incidentsApi } from './incidents'

beforeEach(() => {
  getMock.mockReset()
  postMock.mockReset()
  patchMock.mockReset()
  deleteMock.mockReset()
})

const incident = {
  id: 'i1',
  title: 'Elevated latency',
  severity: 'major' as const,
  status: 'investigating' as const,
  monitorCount: 1,
  createdAt: '2026-06-20T00:00:00Z',
  updatedAt: '2026-06-20T00:00:00Z',
  resolvedAt: null,
}

const input = {
  title: 'Elevated latency',
  message: 'Investigating',
  severity: 'major' as const,
  monitors: [{ monitorType: 'uptime' as const, monitorId: 'm1' }],
}

describe('incidentsApi.list', () => {
  it('fetches the incident list', async () => {
    getMock.mockResolvedValueOnce([incident])

    const result = await incidentsApi.list()

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/incidents/')
    expect(result).toEqual([incident])
  })
})

describe('incidentsApi.get', () => {
  it('fetches a single incident by id', async () => {
    getMock.mockResolvedValueOnce(incident)

    const result = await incidentsApi.get('i1')

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/incidents/i1/')
    expect(result).toEqual(incident)
  })
})

describe('incidentsApi.create', () => {
  it('posts the input to create an incident', async () => {
    postMock.mockResolvedValueOnce(incident)

    const result = await incidentsApi.create(input)

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/incidents/', input)
    expect(result).toEqual(incident)
  })
})

describe('incidentsApi.updateTitle', () => {
  it('patches the title of an incident by id', async () => {
    patchMock.mockResolvedValueOnce({ ...incident, title: 'Renamed' })

    const result = await incidentsApi.updateTitle('i1', 'Renamed')

    expect(patchMock).toHaveBeenCalledExactlyOnceWith('/api/v1/incidents/i1/', { title: 'Renamed' })
    expect(result).toEqual({ ...incident, title: 'Renamed' })
  })
})

describe('incidentsApi.delete', () => {
  it('deletes an incident by id', async () => {
    deleteMock.mockResolvedValueOnce(undefined)

    await incidentsApi.delete('i1')

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('/api/v1/incidents/i1/')
  })
})

describe('incidentsApi.postUpdate', () => {
  it('posts a message and status to the updates endpoint', async () => {
    postMock.mockResolvedValueOnce({ ...incident, status: 'identified' })

    const result = await incidentsApi.postUpdate('i1', 'Found the cause', 'identified')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/incidents/i1/updates', {
      message: 'Found the cause',
      status: 'identified',
    })
    expect(result).toEqual({ ...incident, status: 'identified' })
  })
})

describe('incidentsApi.updateUpdateMessage', () => {
  it('patches the message of a specific update', async () => {
    patchMock.mockResolvedValueOnce(incident)

    const result = await incidentsApi.updateUpdateMessage('i1', 'u1', 'Corrected message')

    expect(patchMock).toHaveBeenCalledExactlyOnceWith('/api/v1/incidents/i1/updates/u1', {
      message: 'Corrected message',
    })
    expect(result).toEqual(incident)
  })
})
