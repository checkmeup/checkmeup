import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import AboutView from './AboutView.vue'

vi.mock('vue-router', () => ({
  RouterLink: { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

describe('AboutView', () => {
  it('mounts without throwing', () => {
    expect(() => mount(AboutView)).not.toThrow()
  })

  it('renders the hero heading', () => {
    const wrapper = mount(AboutView)

    expect(wrapper.text()).toContain('Built by a developer')
    expect(wrapper.text()).toContain('who got paged at 2')
  })

  it('renders the founder name and bio', () => {
    const wrapper = mount(AboutView)

    expect(wrapper.text()).toContain('Andrew Molyuk')
    expect(wrapper.text()).toContain('25+ years writing software')
  })

  it('renders the stack list', () => {
    const wrapper = mount(AboutView)

    expect(wrapper.text()).toContain('Go · Chi · sqlc · goose')
    expect(wrapper.text()).toContain('PostgreSQL')
  })

  it('renders the license note', () => {
    const wrapper = mount(AboutView)

    expect(wrapper.text()).toContain('Business Source License')
    const licenseLink = wrapper.find(
      'a[href="https://github.com/checkmeup/checkmeup/blob/main/LICENSE.md"]',
    )
    expect(licenseLink.exists()).toBe(true)
  })

  it('renders the values section', () => {
    const wrapper = mount(AboutView)

    expect(wrapper.text()).toContain('Built by developers, for developers')
    expect(wrapper.text()).toContain('Transparency by default')
    expect(wrapper.text()).toContain('Shared success')
  })

  it('renders the final CTA linking to sign-up', () => {
    const wrapper = mount(AboutView)

    const cta = wrapper
      .findAllComponents({ name: 'RouterLink' })
      .find((c) => c.text().includes('Create free account'))
    expect(cta).toBeTruthy()
    expect(cta!.props('to')).toBe('/sign-up')
  })

  it('renders contact links', () => {
    const wrapper = mount(AboutView)

    const mailLink = wrapper.find('a[href="mailto:andrew@checkmeup.net"]')
    expect(mailLink.exists()).toBe(true)

    const issueLink = wrapper.find('a[href="https://github.com/checkmeup/checkmeup/issues"]')
    expect(issueLink.exists()).toBe(true)
  })
})
