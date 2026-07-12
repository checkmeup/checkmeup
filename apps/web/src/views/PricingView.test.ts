import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import PricingView from './PricingView.vue'

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

vi.mock('vue-router', () => ({
  RouterLink: { template: '<a><slot /></a>' },
}))

function findButtonByText(wrapper: ReturnType<typeof mount>, text: string) {
  return wrapper.findAll('button').find((b) => b.text().includes(text))
}

describe('PricingView', () => {
  it('renders all four plan names and monthly prices by default', () => {
    const wrapper = mount(PricingView)

    expect(wrapper.text()).toContain('Hobby')
    expect(wrapper.text()).toContain('Solo')
    expect(wrapper.text()).toContain('Startup')
    expect(wrapper.text()).toContain('Enterprise')
    expect(wrapper.text()).toContain('Free')
    expect(wrapper.text()).toContain('$9')
    expect(wrapper.text()).toContain('$29')
    expect(wrapper.text()).toContain('$99')
    expect(wrapper.text()).toContain('/month')
  })

  it('does not show annual prices or effective monthly cost by default', () => {
    const wrapper = mount(PricingView)

    expect(wrapper.text()).not.toContain('/year')
    expect(wrapper.text()).not.toContain('billed annually')
  })

  it('switches to annual prices when the Annual toggle is clicked', async () => {
    const wrapper = mount(PricingView)

    await findButtonByText(wrapper, 'Annual')!.trigger('click')

    expect(wrapper.text()).toContain('$90')
    expect(wrapper.text()).toContain('$290')
    expect(wrapper.text()).toContain('$990')
    expect(wrapper.text()).toContain('/year')
    expect(wrapper.text()).toContain('billed annually')
    expect(wrapper.text()).toContain('Free')
  })

  it('shows the effective monthly cost for annual billing', async () => {
    const wrapper = mount(PricingView)

    await findButtonByText(wrapper, 'Annual')!.trigger('click')

    expect(wrapper.text()).toContain('$7.50/mo, billed annually')
    expect(wrapper.text()).toContain('$24.17/mo, billed annually')
    expect(wrapper.text()).toContain('$82.50/mo, billed annually')
  })

  it('switches back to monthly prices when the Monthly toggle is clicked', async () => {
    const wrapper = mount(PricingView)

    await findButtonByText(wrapper, 'Annual')!.trigger('click')
    expect(wrapper.text()).toContain('/year')

    await findButtonByText(wrapper, 'Monthly')!.trigger('click')

    expect(wrapper.text()).not.toContain('/year')
    expect(wrapper.text()).toContain('/month')
  })

  it('renders a CTA link for each plan', () => {
    const wrapper = mount(PricingView)

    expect(wrapper.text()).toContain('Get started free')
    expect(wrapper.text()).toContain('Start Solo')
    expect(wrapper.text()).toContain('Start Startup')
    expect(wrapper.text()).toContain('Start Enterprise')
  })

  it('renders the full feature comparison table', () => {
    const wrapper = mount(PricingView)

    expect(wrapper.text()).toContain('Full feature comparison')
    expect(wrapper.text()).toContain('Monitors (all types combined)')
    expect(wrapper.text()).toContain('Keyword monitoring')
    expect(wrapper.text()).toContain('Incident management')
    expect(wrapper.text()).toContain('White-label status pages')
  })

  it('renders the billing FAQ entries from findFaqCategory', () => {
    const wrapper = mount(PricingView)

    expect(wrapper.text()).toContain('Frequently asked questions')
    expect(wrapper.text()).toContain('Do I need a credit card to start?')
    expect(wrapper.text()).toContain('Is there a refund policy?')
  })
})
