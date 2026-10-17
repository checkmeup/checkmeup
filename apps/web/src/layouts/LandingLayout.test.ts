import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import LandingLayout from './LandingLayout.vue'

const Page = { template: '<p>page</p>' }

// A real router, with /blog/:slug a sibling of /blog as in src/router/routes.ts:
// RouterLink's own active-class only covers nested routes, which is the case
// these tests pin.
async function mountAt(path: string) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      '/',
      '/docs',
      '/faq',
      '/pricing',
      '/blog',
      '/blog/:slug',
      '/about',
      '/sign-in',
      '/sign-up',
    ].map((p) => ({ path: p, component: Page })),
  })
  await router.push(path)
  await router.isReady()
  return mount(LandingLayout, { global: { plugins: [router] } })
}

const navLink = (wrapper: Awaited<ReturnType<typeof mountAt>>, label: string) =>
  wrapper.findAll('header nav a').find((a) => a.text() === label)!

describe('LandingLayout', () => {
  it('highlights Blog in the nav on a blog post', async () => {
    const wrapper = await mountAt('/blog/some-post')
    expect(navLink(wrapper, 'Blog').classes()).toContain('font-medium')
  })

  it('highlights Blog in the nav on the blog list', async () => {
    const wrapper = await mountAt('/blog')
    expect(navLink(wrapper, 'Blog').classes()).toContain('font-medium')
  })

  it('keeps Sign in and Sign up alone in the footer middle column, with About in the first', async () => {
    const columns = (await mountAt('/'))
      .findAll('footer nav')
      .map((nav) => nav.findAll('a').map((a) => a.text()))
    expect(columns[0]).toContain('About')
    expect(columns[1]).toEqual(['Sign in', 'Sign up'])
  })

  it("links the footer's bottom row to Checkmeup's own status page and GitHub", async () => {
    const hrefs = (await mountAt('/'))
      .findAll('footer > div > div:last-child a')
      .map((a) => a.attributes('href'))
    expect(hrefs).toContain('/status/checkmeup-net')
    expect(hrefs).toContain('https://github.com/checkmeup/checkmeup')
  })

  it('leaves Blog unhighlighted on another page', async () => {
    const wrapper = await mountAt('/docs')
    expect(navLink(wrapper, 'Blog').classes()).not.toContain('font-medium')
    expect(navLink(wrapper, 'Docs').classes()).toContain('font-medium')
  })
})
