import { beforeEach, describe, expect, it, vi } from 'vitest'

const { getMock, postMock } = vi.hoisted(() => ({
  getMock: vi.fn(),
  postMock: vi.fn(),
}))

vi.mock('./client', () => ({
  api: { get: getMock, post: postMock },
}))

import { billingApi } from './billing'

beforeEach(() => {
  getMock.mockReset()
  postMock.mockReset()
})

describe('billingApi.getInfo', () => {
  it('fetches billing info from the billing endpoint', async () => {
    const info = {
      plan: 'solo',
      billingCycle: 'monthly',
      subscriptionStatus: 'active',
      planRenewsAt: '2026-07-01T00:00:00Z',
      monitorCount: 3,
      monitorLimit: 10,
      statusPageCount: 1,
      statusPageLimit: 1,
      minIntervalMins: 5,
      customerPortalUrl: 'https://portal.example.com',
    }
    getMock.mockResolvedValueOnce(info)

    const result = await billingApi.getInfo()

    expect(getMock).toHaveBeenCalledExactlyOnceWith('/api/v1/billing')
    expect(result).toEqual(info)
  })
})

describe('billingApi.createCheckout', () => {
  it('defaults to a monthly billing cycle', async () => {
    postMock.mockResolvedValueOnce({ url: 'https://checkout.example.com' })

    const result = await billingApi.createCheckout('solo')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/billing/checkout', {
      plan: 'solo',
      cycle: 'monthly',
    })
    expect(result).toEqual({ url: 'https://checkout.example.com' })
  })

  it('passes an explicit annual billing cycle through', async () => {
    postMock.mockResolvedValueOnce({ url: 'https://checkout.example.com' })

    await billingApi.createCheckout('startup', 'annual')

    expect(postMock).toHaveBeenCalledExactlyOnceWith('/api/v1/billing/checkout', {
      plan: 'startup',
      cycle: 'annual',
    })
  })
})
