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

    expect(wrapper.find('h1').text()).toBe('Know before your client does.')
    expect(wrapper.text()).toContain('Add one line to the end of a cron job')
  })

  it('renders the hero CTA links', () => {
    const wrapper = mountHome()

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const signUp = links.find((c) => c.text().includes('Start free'))
    const docs = links.find((c) => c.text().includes('Read the docs'))

    expect(signUp).toBeTruthy()
    expect(signUp!.props('to')).toBe('/sign-up')
    expect(docs).toBeTruthy()
    expect(docs!.props('to')).toBe('/docs')
  })

  it('lists all six monitor types in the spec table', () => {
    const wrapper = mountHome()

    expect(wrapper.text()).toContain('Cron job monitoring')
    expect(wrapper.text()).toContain('Uptime monitoring')
    expect(wrapper.text()).toContain('SSL expiry monitoring')
    expect(wrapper.text()).toContain('Domain expiry monitoring')
    expect(wrapper.text()).toContain('Port (TCP) monitoring')
    expect(wrapper.text()).toContain('DNS record monitoring')
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

  it('shows the cron line and the alert it produces in the hero', () => {
    const wrapper = mountHome()

    const example = wrapper.find('[aria-label="Example: a cron job and its missed-ping alert"]')
    expect(example.text()).toContain('curl -fsS checkmeup.net/ping/')
    expect(example.text()).toContain('nightly-backup missed its 03:00 ping')
  })

  it('renders the status page live mockup', () => {
    const wrapper = mountHome()

    expect(wrapper.text()).toContain('All systems operational')
    expect(wrapper.text()).toContain('Hourly Cron Monitor')
    expect(wrapper.text()).toContain('Operational')
  })
})
