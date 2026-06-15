import { api } from './client'

export interface BillingInfo {
  plan: 'hobbyist' | 'indie' | 'studio' | 'agency'
  subscriptionStatus: string
  planRenewsAt: string | null
  monitorCount: number
  monitorLimit: number
  statusPageCount: number
  statusPageLimit: number
  minIntervalMins: number
  customerPortalUrl: string
}

export const billingApi = {
  async getInfo(): Promise<BillingInfo> {
    return api.get('/api/v1/billing')
  },

  async createCheckout(plan: string): Promise<{ url: string }> {
    return api.post('/api/v1/billing/checkout', { plan })
  },
}
