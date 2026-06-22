import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const { clearMock, pushMock } = vi.hoisted(() => ({
  clearMock: vi.fn(),
  pushMock: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ clear: clearMock }),
}))

vi.mock('@/router', () => ({
  router: { push: pushMock },
}))

function jsonResponse(status: number, body: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: 'Status Text',
    json: () => Promise.resolve(body),
  } as Response
}

describe('api client', () => {
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.resetModules()
    clearMock.mockClear()
    pushMock.mockClear()
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  async function loadClient() {
    return import('./client')
  }

  it('GET sends credentials and a JSON content-type header, returns the parsed body', async () => {
    const { api } = await loadClient()
    fetchMock.mockResolvedValueOnce(jsonResponse(200, { hello: 'world' }))

    const result = await api.get('/api/v1/thing')

    expect(fetchMock).toHaveBeenCalledExactlyOnceWith('/api/v1/thing', {
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
    })
    expect(result).toEqual({ hello: 'world' })
  })

  it('POST sends the method and JSON-serialized body', async () => {
    const { api } = await loadClient()
    fetchMock.mockResolvedValueOnce(jsonResponse(200, { ok: true }))

    await api.post('/api/v1/thing', { foo: 'bar' })

    expect(fetchMock).toHaveBeenCalledExactlyOnceWith('/api/v1/thing', {
      method: 'POST',
      body: JSON.stringify({ foo: 'bar' }),
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
    })
  })

  it('PUT sends the method and JSON-serialized body', async () => {
    const { api } = await loadClient()
    fetchMock.mockResolvedValueOnce(jsonResponse(200, { ok: true }))

    await api.put('/api/v1/thing/1', { foo: 'bar' })

    expect(fetchMock).toHaveBeenCalledExactlyOnceWith(
      '/api/v1/thing/1',
      expect.objectContaining({ method: 'PUT', body: JSON.stringify({ foo: 'bar' }) }),
    )
  })

  it('PATCH sends the method and JSON-serialized body', async () => {
    const { api } = await loadClient()
    fetchMock.mockResolvedValueOnce(jsonResponse(200, { ok: true }))

    await api.patch('/api/v1/thing/1', { foo: 'bar' })

    expect(fetchMock).toHaveBeenCalledExactlyOnceWith(
      '/api/v1/thing/1',
      expect.objectContaining({ method: 'PATCH', body: JSON.stringify({ foo: 'bar' }) }),
    )
  })

  it('DELETE sends the method with no body', async () => {
    const { api } = await loadClient()
    fetchMock.mockResolvedValueOnce(jsonResponse(200, {}))

    await api.delete('/api/v1/thing/1')

    expect(fetchMock).toHaveBeenCalledExactlyOnceWith(
      '/api/v1/thing/1',
      expect.objectContaining({ method: 'DELETE' }),
    )
  })

  it('returns undefined for a 204 No Content response', async () => {
    const { api } = await loadClient()
    fetchMock.mockResolvedValueOnce(jsonResponse(204, undefined))

    const result = await api.delete('/api/v1/thing/1')

    expect(result).toBeUndefined()
  })

  it('throws an ApiError using the server-provided message and code', async () => {
    const { api } = await loadClient()
    fetchMock.mockResolvedValueOnce(jsonResponse(400, { error: 'bad request', code: 'BAD' }))

    await expect(api.get('/api/v1/thing')).rejects.toMatchObject({
      name: 'ApiError',
      status: 400,
      message: 'bad request',
      code: 'BAD',
    })
  })

  it('falls back to statusText and an empty code when the error body is not JSON', async () => {
    const { api } = await loadClient()
    fetchMock.mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
      json: () => Promise.reject(new Error('not json')),
    } as Response)

    await expect(api.get('/api/v1/thing')).rejects.toMatchObject({
      status: 500,
      message: 'Internal Server Error',
      code: '',
    })
  })

  it('on a 401, refreshes the session and retries the original request once', async () => {
    const { api } = await loadClient()
    fetchMock
      .mockResolvedValueOnce(jsonResponse(401, {}))
      .mockResolvedValueOnce(jsonResponse(200, {}))
      .mockResolvedValueOnce(jsonResponse(200, { data: 'after-refresh' }))

    const result = await api.get('/api/v1/thing')

    expect(fetchMock).toHaveBeenCalledTimes(3)
    expect(fetchMock.mock.calls[1][0]).toBe('/api/v1/auth/refresh')
    expect(fetchMock.mock.calls[1][1]).toMatchObject({ method: 'POST', credentials: 'include' })
    expect(fetchMock.mock.calls[2][0]).toBe('/api/v1/thing')
    expect(result).toEqual({ data: 'after-refresh' })
  })

  it('does not refresh a second time if the retried request also 401s', async () => {
    const { api } = await loadClient()
    fetchMock
      .mockResolvedValueOnce(jsonResponse(401, {}))
      .mockResolvedValueOnce(jsonResponse(200, {}))
      .mockResolvedValueOnce(jsonResponse(401, { error: 'still unauthorized' }))

    await expect(api.get('/api/v1/thing')).rejects.toMatchObject({
      status: 401,
      message: 'still unauthorized',
    })
    expect(fetchMock).toHaveBeenCalledTimes(3)
  })

  it('clears the auth store and redirects to sign-in when the refresh itself fails', async () => {
    const { api } = await loadClient()
    fetchMock
      .mockResolvedValueOnce(jsonResponse(401, {}))
      .mockResolvedValueOnce(jsonResponse(401, {}))

    await expect(api.get('/api/v1/thing')).rejects.toMatchObject({
      message: 'session expired',
    })
    expect(clearMock).toHaveBeenCalledOnce()
    expect(pushMock).toHaveBeenCalledExactlyOnceWith({ name: 'sign-in' })
  })

  it('dedupes concurrent refresh attempts into a single in-flight request', async () => {
    const { api } = await loadClient()
    fetchMock.mockImplementation((url: string) => {
      if (url === '/api/v1/auth/refresh') return Promise.resolve(jsonResponse(200, {}))
      return Promise.resolve(jsonResponse(401, { error: 'unauthorized' }))
    })

    await Promise.allSettled([api.get('/api/v1/a'), api.get('/api/v1/b')])

    const refreshCalls = fetchMock.mock.calls.filter(([url]) => url === '/api/v1/auth/refresh')
    expect(refreshCalls).toHaveLength(1)
  })
})
