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
    expect(wrapper.text()).toContain('5. Fees and billing')
    expect(wrapper.text()).toContain('12. Governing law')
    expect(wrapper.text()).toContain('15. Contact')
  })

  it('states the eligibility/age requirement', () => {
    const wrapper = mount(TermsView, mountOptions)

    expect(wrapper.text()).toContain('2. Eligibility')
    expect(wrapper.text()).toContain('at least 18 years old')
  })

  it('includes an indemnification clause', () => {
    const wrapper = mount(TermsView, mountOptions)

    expect(wrapper.text()).toContain('10. Indemnification')
    expect(wrapper.text()).toContain('indemnify and hold us harmless')
  })

  it('includes miscellaneous boilerplate (entire agreement, severability, assignment)', () => {
    const wrapper = mount(TermsView, mountOptions)

    expect(wrapper.text()).toContain('13. Miscellaneous')
    expect(wrapper.text()).toContain('entire agreement')
    expect(wrapper.text()).toContain('unenforceable')
    expect(wrapper.text()).toContain('merger, acquisition, or sale of assets')
  })

  it('mentions LemonSqueezy as the merchant of record for billing', () => {
    const wrapper = mount(TermsView, mountOptions)

    expect(wrapper.text()).toContain('LemonSqueezy')
    expect(wrapper.text()).toContain('merchant of record')
  })

  it('names the sole-proprietor operator', () => {
    const wrapper = mount(TermsView, mountOptions)

    expect(wrapper.text()).toContain('Andrew Molyuk')
    expect(wrapper.text()).toContain('sole proprietor')
  })

  it('mentions the refund policy', () => {
    const wrapper = mount(TermsView, mountOptions)

    expect(wrapper.text()).toContain('Refund Policy')
  })

  it('specifies exclusive Israeli court jurisdiction with no arbitration', () => {
    const wrapper = mount(TermsView, mountOptions)

    expect(wrapper.text()).toContain('exclusively in the competent courts of Israel')
    expect(wrapper.text()).toContain('not arbitration')
  })
})
