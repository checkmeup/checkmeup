import { beforeEach, describe, expect, it, vi } from 'vitest'

const { getMock, postMock, patchMock, putMock, deleteMock } = vi.hoisted(() => ({
  getMock: vi.fn(),
  postMock: vi.fn(),
  patchMock: vi.fn(),
  putMock: vi.fn(),
  deleteMock: vi.fn(),
}))

vi.mock('./client', () => ({
  api: { get: getMock, post: postMock, patch: patchMock, put: putMock, delete: deleteMock },
}))

import { statusPagesApi } from './statusPages'

beforeEach(() => {
  getMock.mockReset()
  postMock.mockReset()
  patchMock.mockReset()
  putMock.mockReset()
  deleteMock.mockReset()
})

const statusPage = {
  id: 'sp1',
  slug: 'status',
  title: 'checkmeup status',
  description: 'Live status of our services',
  logoUrl: 'https://checkmeup.net/logo.png',
  hideBranding: false,
  publicUrl: 'https://checkmeup.net/status/status',
  createdAt: '2026-06-01T00:00:00Z',
}

const monitorItem = {
  id: 'mi1',
  monitorType: 'cron' as const,
  monitorId: 'c1',
  displayName: 'Nightly backup',
  displayOrder: 0,
}

const createInput = {
  slug: 'status',
  title: 'checkmeup status',
  description: 'Live status of our services',
  logoUrl: 'https://checkmeup.net/logo.png',
}

const updateInput = {
  title: 'checkmeup status',
  description: 'Live status of our services',
  logoUrl: 'https://checkmeup.net/logo.png',
  hideBranding: false,
}

const setMonitorsInput = {
  monitors: [
    {
      monitorType: 'cron' as const,
      monitorId: 'c1',
      displayName: 'Nightly backup',
      displayOrder: 0,
    },
  ],
}

describe('statusPagesApi.checkSlug', () => {
  it('fetches slug availability with the slug URL-encoded', async () => {
    const result_ = { available: false, reason: 'already taken' }
    getMock.mockResolvedValueOnce(result_)

    const result = await statusPagesApi.checkSlug('my status')

    expect(getMock).toHaveBeenCalledExactlyOnceWith(
      '/api/v1/status-pages/check-slug?slug=my%20status',
    )
    expect(result).toEqual(result_)
  })
})

describe('statusPagesApi.list', () => {
  it('fetches the status page list', async () => {
    getMock.mockResolvedValueOnce([statusPage])

    const result = await statusPagesApi.list()

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/status-pages/')
    expect(result).toEqual([statusPage])
  })
})

describe('statusPagesApi.create', () => {
  it('posts the input to create a status page', async () => {
    postMock.mockResolvedValueOnce(statusPage)

    const result = await statusPagesApi.create(createInput)

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/status-pages/', createInput)
    expect(result).toEqual(statusPage)
  })
})

describe('statusPagesApi.get', () => {
  it('fetches a single status page detail by id', async () => {
    const detail = { ...statusPage, monitors: [monitorItem] }
    getMock.mockResolvedValueOnce(detail)

    const result = await statusPagesApi.get('sp1')

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/status-pages/sp1/')
    expect(result).toEqual(detail)
  })
})

describe('statusPagesApi.update', () => {
  it('patches the input to update a status page by id', async () => {
    patchMock.mockResolvedValueOnce(statusPage)

    const result = await statusPagesApi.update('sp1', updateInput)

    expect(patchMock).toHaveBeenCalledExactlyOnceWith('/api/v1/status-pages/sp1/', updateInput)
    expect(result).toEqual(statusPage)
  })
})

describe('statusPagesApi.delete', () => {
  it('deletes a status page by id', async () => {
    deleteMock.mockResolvedValueOnce(undefined)

    await statusPagesApi.delete('sp1')

    expect(deleteMock).toHaveBeenCalledExactlyOnceWith('/api/v1/status-pages/sp1/')
  })
})

describe('statusPagesApi.setMonitors', () => {
  it('puts the input to replace a status page monitor list by id', async () => {
    putMock.mockResolvedValueOnce([monitorItem])

    const result = await statusPagesApi.setMonitors('sp1', setMonitorsInput)

    expect(putMock).toHaveBeenCalledExactlyOnceWith(
      '/api/v1/status-pages/sp1/monitors',
      setMonitorsInput,
    )
    expect(result).toEqual([monitorItem])
  })
})
