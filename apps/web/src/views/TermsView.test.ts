import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import TermsView from './TermsView.vue'

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

const mountOptions = { global: { stubs: { RouterLink: true } } }

describe('TermsView', () => {
  it('mounts without throwing', () => {
    expect(() => mount(TermsView, mountOptions)).not.toThrow()
  })

  it('renders the heading and effective date', () => {
    const wrapper = mount(TermsView, mountOptions)

    expect(wrapper.text()).toContain('Terms of Service')
    expect(wrapper.text()).toContain('Effective 2026-06-17')
  })

  it('renders the expected section titles', () => {
    const wrapper = mount(TermsView, mountOptions)

    expect(wrapper.text()).toContain('1. The service')
    expect(wrapper.text()).toContain('4. Fees and billing')
    expect(wrapper.text()).toContain('10. Governing law')
    expect(wrapper.text()).toContain('12. Contact')
  })

  it('mentions LemonSqueezy as the merchant of record for billing', () => {
    const wrapper = mount(TermsView, mountOptions)

    expect(wrapper.text()).toContain('LemonSqueezy')
    expect(wrapper.text()).toContain('merchant of record')
  })
})
