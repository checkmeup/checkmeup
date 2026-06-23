import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import PrivacyView from './PrivacyView.vue'

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

const mountOptions = { global: { stubs: { RouterLink: true } } }

describe('PrivacyView', () => {
  it('mounts without throwing', () => {
    expect(() => mount(PrivacyView, mountOptions)).not.toThrow()
  })

  it('renders the heading and effective date', () => {
    const wrapper = mount(PrivacyView, mountOptions)

    expect(wrapper.text()).toContain('Privacy Policy')
    expect(wrapper.text()).toContain('Effective 2026-06-17')
  })

  it('renders the expected section titles', () => {
    const wrapper = mount(PrivacyView, mountOptions)

    expect(wrapper.text()).toContain('1. Scope')
    expect(wrapper.text()).toContain('5. Subprocessors')
    expect(wrapper.text()).toContain('9. Your rights')
    expect(wrapper.text()).toContain('11. Contact')
  })

  it('lists all subprocessors in the table', () => {
    const wrapper = mount(PrivacyView, mountOptions)

    expect(wrapper.text()).toContain('Resend')
    expect(wrapper.text()).toContain('Telegram')
    expect(wrapper.text()).toContain('LemonSqueezy')
    expect(wrapper.text()).toContain('Hetzner')
  })
})
