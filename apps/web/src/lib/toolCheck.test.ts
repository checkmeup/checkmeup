import { afterEach, describe, expect, it, vi } from 'vitest'
import { expiryStatus, runToolCheck } from './toolCheck'

function respondWith(status: number, body: unknown) {
  const fetchMock = vi.fn(() =>
    Promise.resolve({ ok: status < 400, status, json: () => Promise.resolve(body) }),
  )
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('runToolCheck', () => {
  it('asks the endpoint about the trimmed, encoded value and returns the result', async () => {
    const result = {
      host: 'example.com',
      expiresAt: '2027-01-02T00:00:00Z',
      daysLeft: 77,
      issuer: 'R11',
    }
    const fetchMock = respondWith(200, result)

    await expect(runToolCheck('/api/v1/tools/ssl', 'host', ' example.com/a b ')).resolves.toEqual({
      ok: true,
      data: result,
    })
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/tools/ssl?host=example.com%2Fa%20b')
  })

  it("gives the server's reason when the check fails", async () => {
    respondWith(422, { error: 'x509: certificate has expired', code: 'check_failed' })

    await expect(runToolCheck('/api/v1/tools/ssl', 'host', 'expired.example.com')).resolves.toEqual(
      {
        ok: false,
        message: 'x509: certificate has expired',
      },
    )
  })

  it('says to try again in a minute when rate limited', async () => {
    respondWith(429, 'Too Many Requests')

    await expect(runToolCheck('/api/v1/tools/domain', 'domain', 'example.com')).resolves.toEqual({
      ok: false,
      message: 'Too many checks right now. Try again in a minute.',
    })
  })

  it('says the check could not run when the network fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.reject(new TypeError('Failed to fetch'))),
    )

    await expect(runToolCheck('/api/v1/tools/domain', 'domain', 'example.com')).resolves.toEqual({
      ok: false,
      message: "The check couldn't run. Try again.",
    })
  })
})

describe('expiryStatus', () => {
  it.each([
    [-1, 'down', 'Expired'],
    [0, 'warn', 'Expires today'],
    [1, 'warn', 'Expires in 1 day'],
    [30, 'warn', 'Expires in 30 days'],
    [31, 'up', 'Valid for 31 days'],
  ])('%i days left reads %s, "%s"', (daysLeft, tone, label) => {
    expect(expiryStatus(daysLeft)).toEqual({ tone, label })
  })
})
