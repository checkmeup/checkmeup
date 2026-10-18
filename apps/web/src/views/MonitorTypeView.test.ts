import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import MonitorTypeView from './MonitorTypeView.vue'
import { routes } from '@/router/routes'

const flushHead = () => new Promise((resolve) => setTimeout(resolve, 0))

vi.mock('vue-router', () => ({
  RouterLink: { name: 'RouterLink', props: ['to'], template: '<a :href="to"><slot /></a>' },
}))

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

// Head entries belong to the mounted component; unmount each so one page's
// title can't leak into the next test.
const mounted: ReturnType<typeof mount>[] = []
function mountType(type: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port' | 'dns') {
  const wrapper = mount(MonitorTypeView, { props: { type } })
  mounted.push(wrapper)
  return wrapper
}

afterEach(() => {
  mounted.splice(0).forEach((wrapper) => wrapper.unmount())
})

const pages = [
  {
    type: 'cron',
    path: '/cron-monitoring',
    h1: 'Cron job monitoring',
    link: '/blog/cron-job-monitoring-guide',
  },
  {
    type: 'uptime',
    path: '/uptime-monitoring',
    h1: 'Uptime monitoring',
    link: '/blog/uptime-monitoring-guide',
  },
  {
    type: 'ssl',
    path: '/ssl-monitoring',
    h1: 'SSL certificate monitoring',
    link: '/tools/ssl-checker',
  },
  {
    type: 'domain',
    path: '/domain-expiry-monitoring',
    h1: 'Domain expiry monitoring',
    link: '/tools/domain-expiry-checker',
  },
  {
    type: 'port',
    path: '/port-monitoring',
    h1: 'Port (TCP) monitoring',
    link: '/blog/port-tcp-monitoring-guide',
  },
  {
    type: 'dns',
    path: '/dns-monitoring',
    h1: 'DNS record monitoring',
    link: '/blog/dns-record-monitoring-guide',
  },
] as const

describe('MonitorTypeView', () => {
  it.each(pages)('serves the $type page at $path', ({ type, path }) => {
    const route = routes.find((r) => r.path === path)
    expect(route?.props).toEqual({ type })
  })

  it.each(pages)(
    'gives the $type page its own heading, title and canonical URL',
    async ({ type, path, h1 }) => {
      const wrapper = mountType(type)
      await flushHead()

      expect(wrapper.find('h1').text()).toBe(h1)
      expect(document.title).toContain(h1.split(' ')[0])
      expect(document.head.querySelector('link[rel="canonical"]')?.getAttribute('href')).toBe(
        `https://checkmeup.net${path}`,
      )
    },
  )

  it.each(pages)('links the $type page to sign-up and to $link', ({ type, link }) => {
    const wrapper = mountType(type)

    expect(wrapper.find('a[href="/sign-up"]').exists()).toBe(true)
    expect(wrapper.find(`a[href="${link}"]`).exists()).toBe(true)
  })

  it('shows every plan with its price, and says what its check does not prove', () => {
    const wrapper = mountType('port')

    for (const plan of ['Hobby', 'Solo', 'Startup', 'Enterprise'])
      expect(wrapper.text()).toContain(plan)
    expect(wrapper.text()).toContain('$29')
    expect(wrapper.text()).toContain("What it doesn't tell you")
    expect(wrapper.text()).toContain('no handshake or data is sent')
  })
})
