import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import RefundPolicyView from './RefundPolicyView.vue'

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

const mountOptions = { global: { stubs: { RouterLink: true } } }

describe('RefundPolicyView', () => {
  it('mounts without throwing', () => {
    expect(() => mount(RefundPolicyView, mountOptions)).not.toThrow()
  })

  it('renders the heading and effective date', () => {
    const wrapper = mount(RefundPolicyView, mountOptions)

    expect(wrapper.text()).toContain('Refund Policy')
    expect(wrapper.text()).toContain('Effective 2026-07-01')
  })

  it('states the 30-day no-questions-asked policy', () => {
    const wrapper = mount(RefundPolicyView, mountOptions)

    expect(wrapper.text()).toContain('30 days')
    expect(wrapper.text()).toContain('no questions asked')
  })

  it('renders the contact email', () => {
    const wrapper = mount(RefundPolicyView, mountOptions)

    const mailLink = wrapper.find('a[href="mailto:andrew@checkmeup.net"]')
    expect(mailLink.exists()).toBe(true)
  })
})
