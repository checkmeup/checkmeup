import { api } from './client'

export type BillingCycle = 'monthly' | 'annual'

export interface BillingInfo {
  plan: 'hobby' | 'solo' | 'startup' | 'enterprise'
  billingCycle: BillingCycle
  subscriptionStatus: string
  planRenewsAt: string | null
  monitorCount: number
  monitorLimit: number
  statusPageCount: number
  statusPageLimit: number
  minIntervalMins: number
  keywordMonitoringEnabled: boolean
  customerPortalUrl: string
}

export const billingApi = {
  async getInfo(): Promise<BillingInfo> {
    return api.get('/api/v1/billing')
  },

  async createCheckout(plan: string, cycle: BillingCycle = 'monthly'): Promise<{ url: string }> {
    return api.post('/api/v1/billing/checkout', { plan, cycle })
  },
}
