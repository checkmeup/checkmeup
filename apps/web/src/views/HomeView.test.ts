import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import HomeView from './HomeView.vue'

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

const RouterLinkStub = {
  name: 'RouterLink',
  props: ['to'],
  template: '<a><slot /></a>',
}

function mountHome() {
  return mount(HomeView, {
    global: {
      stubs: { RouterLink: RouterLinkStub },
    },
  })
}

describe('HomeView', () => {
  it('mounts without throwing', () => {
    expect(() => mountHome()).not.toThrow()
  })

  it('renders the hero heading and subheading', () => {
    const wrapper = mountHome()

    expect(wrapper.text()).toContain('Know before')
    expect(wrapper.text()).toContain('your client does.')
    expect(wrapper.text()).toContain(
      'Cron job monitoring, uptime checks, SSL and domain expiry alerts',
    )
  })

  it('renders the hero CTA links', () => {
    const wrapper = mountHome()

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const signUp = links.find((c) => c.text().includes('Start free'))
    const signIn = links.find((c) => c.text().includes('Sign in'))

    expect(signUp).toBeTruthy()
    expect(signUp!.props('to')).toBe('/sign-up')
    expect(signIn).toBeTruthy()
    expect(signIn!.props('to')).toBe('/sign-in')
  })

  it('renders all five monitor feature cards plus the status pages card', () => {
    const wrapper = mountHome()

    expect(wrapper.text()).toContain('Cron job monitoring')
    expect(wrapper.text()).toContain('Uptime monitoring')
    expect(wrapper.text()).toContain('SSL expiry monitoring')
    expect(wrapper.text()).toContain('Domain expiry monitoring')
    expect(wrapper.text()).toContain('Port (TCP) monitoring')
    expect(wrapper.text()).toContain('Public status pages')
  })

  it('renders the status pages highlight section', () => {
    const wrapper = mountHome()

    expect(wrapper.text()).toContain('Give your clients a page they can bookmark.')

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const cta = links.find((c) => c.text().includes('Create a status page'))
    expect(cta).toBeTruthy()
    expect(cta!.props('to')).toBe('/sign-up')
  })

  it('renders all four pricing plans with correct prices', () => {
    const wrapper = mountHome()

    expect(wrapper.text()).toContain('Hobby')
    expect(wrapper.text()).toContain('Free')
    expect(wrapper.text()).toContain('Solo')
    expect(wrapper.text()).toContain('$9')
    expect(wrapper.text()).toContain('Startup')
    expect(wrapper.text()).toContain('$29')
    expect(wrapper.text()).toContain('Enterprise')
    expect(wrapper.text()).toContain('$99')
  })

  it('marks the Startup plan as popular', () => {
    const wrapper = mountHome()

    expect(wrapper.text()).toContain('Most popular')
  })

  it('renders the full pricing details link', () => {
    const wrapper = mountHome()

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const pricingLink = links.find((c) => c.text().includes('See full pricing details'))
    expect(pricingLink).toBeTruthy()
    expect(pricingLink!.props('to')).toBe('/pricing')
  })

  it('renders the closing CTA banner', () => {
    const wrapper = mountHome()

    expect(wrapper.text()).toContain('Start monitoring in 60 seconds.')
    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const ctas = links.filter((c) => c.text().includes('Create free account'))
    expect(ctas.length).toBeGreaterThan(0)
    expect(ctas[0].props('to')).toBe('/sign-up')
  })

  it('renders the hero live mockup', () => {
    const wrapper = mountHome()

    expect(wrapper.text()).toContain('checkmeup.net — Monitor')
    expect(wrapper.text()).toContain('Uptime 24h')
    expect(wrapper.find('svg path').exists()).toBe(true)
  })

  it('renders the status page live mockup', () => {
    const wrapper = mountHome()

    expect(wrapper.text()).toContain('All systems operational')
    expect(wrapper.text()).toContain('Hourly Cron Monitor')
    expect(wrapper.text()).toContain('Operational')
  })
})
