import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import BillingView from './BillingView.vue'

vi.mock('@/layouts/AppLayout.vue', () => ({
  default: { name: 'AppLayout', template: '<div><slot /></div>' },
}))

const { createCheckoutMock } = vi.hoisted(() => ({
  createCheckoutMock: vi.fn(),
}))

vi.mock('@/api/billing', () => ({
  billingApi: { createCheckout: createCheckoutMock },
}))

const { initializePaddleMock, checkoutOpenMock } = vi.hoisted(() => ({
  initializePaddleMock: vi.fn(),
  checkoutOpenMock: vi.fn(),
}))

vi.mock('@paddle/paddle-js', () => ({
  initializePaddle: initializePaddleMock,
}))

vi.mock('@/lib/theme', () => ({
  useTheme: () => ({ theme: ref('light') }),
}))

const { ApiError } = vi.hoisted(() => ({
  ApiError: class ApiError extends Error {
    status: number
    code: string
    constructor(status: number, message: string, code = '') {
      super(message)
      this.status = status
      this.code = code
    }
  },
}))

vi.mock('@/api/client', () => ({ ApiError }))

const infoData = ref<unknown>(null)
const infoPending = ref(false)
const infoError = ref<{ message: string } | null>(null)

vi.mock('@/composables/useBilling', () => ({
  useBilling: () => ({
    data: infoData,
    isPending: infoPending,
    error: infoError,
  }),
}))

const hobbyInfo = {
  plan: 'hobby' as const,
  billingCycle: 'monthly' as const,
  subscriptionStatus: 'none',
  planRenewsAt: null,
  monitorCount: 2,
  monitorLimit: 5,
  statusPageCount: 0,
  statusPageLimit: 1,
  minIntervalMins: 30,
  customerPortalUrl: '',
}

const soloInfo = {
  plan: 'solo' as const,
  billingCycle: 'monthly' as const,
  subscriptionStatus: 'active',
  planRenewsAt: '2026-07-01',
  monitorCount: 8,
  monitorLimit: 20,
  statusPageCount: 1,
  statusPageLimit: 3,
  minIntervalMins: 5,
  customerPortalUrl: 'https://billing.example.com/portal',
}

const enterpriseInfo = {
  plan: 'enterprise' as const,
  billingCycle: 'annual' as const,
  subscriptionStatus: 'active',
  planRenewsAt: '2027-01-01',
  monitorCount: 50,
  monitorLimit: -1,
  statusPageCount: 2,
  statusPageLimit: -1,
  minIntervalMins: 1,
  customerPortalUrl: 'https://billing.example.com/portal',
}

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

beforeEach(() => {
  infoData.value = null
  infoPending.value = false
  infoError.value = null
  initializePaddleMock.mockResolvedValue({ Checkout: { open: checkoutOpenMock } })
})

afterEach(() => {
  vi.clearAllMocks()
  vi.unstubAllEnvs()
})

describe('BillingView', () => {
  it('shows a loading state while pending', () => {
    infoPending.value = true
    const wrapper = mount(BillingView)

    expect(wrapper.text()).toContain('Loading…')
  })

  it('shows an error message when the query fails', () => {
    infoError.value = { message: 'Failed to load billing info' }
    const wrapper = mount(BillingView)

    expect(wrapper.text()).toContain('Failed to load billing info')
  })

  it('renders the hobby plan as free with upgrade options for all paid plans', () => {
    infoData.value = { ...hobbyInfo }
    const wrapper = mount(BillingView)

    expect(wrapper.text()).toContain('Hobby')
    expect(wrapper.text()).toContain('Free')
    expect(wrapper.text()).toContain('2 / 5')
    expect(wrapper.text()).toContain('0 / 1')
    expect(findButtonByText(wrapper, 'Upgrade to Solo')).toBeTruthy()
    expect(findButtonByText(wrapper, 'Upgrade to Startup')).toBeTruthy()
    expect(findButtonByText(wrapper, 'Upgrade to Enterprise')).toBeTruthy()
  })

  it('does not show the manage subscription link for the hobby plan', () => {
    infoData.value = { ...hobbyInfo }
    const wrapper = mount(BillingView)

    expect(wrapper.text()).not.toContain('Manage subscription')
  })

  it('renders monthly pricing and the manage subscription link for a paid plan', () => {
    infoData.value = { ...soloInfo }
    const wrapper = mount(BillingView)

    expect(wrapper.text()).toContain('Solo')
    expect(wrapper.text()).toContain('$9/mo')
    expect(wrapper.text()).toContain('Manage subscription')
    expect(wrapper.text()).toContain('Renews')
    expect(wrapper.text()).toContain('2026-07-01')
  })

  it('shows "Access until" wording when the subscription is cancelled', () => {
    infoData.value = { ...soloInfo, subscriptionStatus: 'cancelled' }
    const wrapper = mount(BillingView)

    expect(wrapper.text()).toContain('Access until')
  })

  it('shows unlimited usage labels and hides upgrade options for the top enterprise plan', () => {
    infoData.value = { ...enterpriseInfo }
    const wrapper = mount(BillingView)

    expect(wrapper.text()).toContain('50 / unlimited')
    expect(wrapper.text()).toContain('2 / unlimited')
    expect(wrapper.text()).not.toContain('Upgrade')
  })

  it('renders annual pricing with the effective monthly cost', async () => {
    infoData.value = { ...hobbyInfo }
    const wrapper = mount(BillingView)

    const annualToggle = wrapper.findAll('button').find((b) => b.text().includes('Annual'))
    await annualToggle!.trigger('click')

    expect(wrapper.text()).toContain('$90/yr ($7.50/mo)')
  })

  it('starts checkout and opens the Paddle overlay with the returned transaction', async () => {
    vi.stubEnv('VITE_PADDLE_CLIENT_TOKEN', 'test-client-token')
    infoData.value = { ...hobbyInfo }
    createCheckoutMock.mockResolvedValueOnce({ transactionId: 'txn_01example' })
    const wrapper = mount(BillingView)

    await findButtonByText(wrapper, 'Upgrade to Solo')!.trigger('click')
    await flushPromises()

    expect(createCheckoutMock).toHaveBeenCalledExactlyOnceWith('solo', 'monthly')
    expect(initializePaddleMock).toHaveBeenCalledExactlyOnceWith({
      token: 'test-client-token',
      environment: 'sandbox',
    })
    expect(checkoutOpenMock).toHaveBeenCalledExactlyOnceWith({
      transactionId: 'txn_01example',
      settings: {
        theme: 'light',
        successUrl: `${window.location.origin}/billing?upgraded=true`,
      },
    })
  })

  it('shows a not-configured message when the Paddle client token is missing', async () => {
    infoData.value = { ...hobbyInfo }
    createCheckoutMock.mockResolvedValueOnce({ transactionId: 'txn_01example' })
    const wrapper = mount(BillingView)

    await findButtonByText(wrapper, 'Upgrade to Solo')!.trigger('click')
    await flushPromises()

    expect(initializePaddleMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain("Billing isn't activated yet")
  })

  it('shows a not-configured message when billing has not been activated', async () => {
    infoData.value = { ...hobbyInfo }
    createCheckoutMock.mockRejectedValueOnce(new ApiError(400, 'not configured', 'not_configured'))
    const wrapper = mount(BillingView)

    await findButtonByText(wrapper, 'Upgrade to Solo')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain("Billing isn't activated yet")
  })

  it('shows a generic error message when checkout fails for another reason', async () => {
    infoData.value = { ...hobbyInfo }
    createCheckoutMock.mockRejectedValueOnce(new Error('Network error'))
    const wrapper = mount(BillingView)

    await findButtonByText(wrapper, 'Upgrade to Solo')!.trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Network error')
  })
})
