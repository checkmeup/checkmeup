import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import AlternativeView from './AlternativeView.vue'
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
function mountAlternative(slug: 'healthchecks-io' | 'uptimerobot' | 'cronitor') {
  const wrapper = mount(AlternativeView, { props: { slug } })
  mounted.push(wrapper)
  return wrapper
}

afterEach(() => {
  mounted.splice(0).forEach((wrapper) => wrapper.unmount())
})

const pages = [
  { slug: 'healthchecks-io', name: 'Healthchecks.io', source: 'https://healthchecks.io/pricing/' },
  { slug: 'uptimerobot', name: 'UptimeRobot', source: 'https://uptimerobot.com/pricing/' },
  { slug: 'cronitor', name: 'Cronitor', source: 'https://cronitor.io/pricing' },
] as const

describe('AlternativeView', () => {
  it.each(pages)('serves the $name page at /alternatives/$slug', ({ slug }) => {
    const route = routes.find((r) => r.path === `/alternatives/${slug}`)
    expect(route?.props).toEqual({ slug })
  })

  it.each(pages)(
    'gives the $name page its own heading, title and canonical URL',
    async ({ slug, name }) => {
      const wrapper = mountAlternative(slug)
      await flushHead()

      expect(wrapper.find('h1').text()).toContain(`${name} alternative`)
      expect(document.title).toContain(`${name} Alternative`)
      expect(document.head.querySelector('link[rel="canonical"]')?.getAttribute('href')).toBe(
        `https://checkmeup.net/alternatives/${slug}`,
      )
    },
  )

  it.each(pages)(
    'compares Checkmeup with $name side by side, dated, with its pricing page as a source',
    ({ slug, name, source }) => {
      const wrapper = mountAlternative(slug)

      const headers = wrapper.findAll('thead th').map((th) => th.text())
      expect(headers).toContain('Checkmeup')
      expect(headers).toContain(name)
      expect(wrapper.findAll('tbody tr').length).toBeGreaterThan(3)
      expect(wrapper.text()).toContain('As of October 18, 2026')
      expect(wrapper.find(`a[href="${source}"]`).exists()).toBe(true)
    },
  )

  it.each(pages)('says when $name is the better choice, and links to sign-up', ({ slug, name }) => {
    const wrapper = mountAlternative(slug)

    expect(wrapper.text()).toContain(`When ${name} is the better choice`)
    expect(wrapper.find('a[href="/sign-up"]').exists()).toBe(true)
  })
})
