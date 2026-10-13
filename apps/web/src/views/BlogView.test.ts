import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import BlogView from './BlogView.vue'
import { articlesMeta, releasesMeta } from '@/blog/posts'

vi.mock('vue-router', () => ({
  RouterLink: { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

describe('BlogView', () => {
  it('mounts without throwing', () => {
    expect(() => mount(BlogView)).not.toThrow()
  })

  it('renders the page heading', () => {
    const wrapper = mount(BlogView)

    expect(wrapper.text()).toContain('Blog')
  })

  // Every RouterLink except the intro's /changelog link is a post entry.
  const postLinks = (wrapper: ReturnType<typeof mount>) =>
    wrapper
      .findAllComponents({ name: 'RouterLink' })
      .filter((link) => link.props('to') !== '/changelog')

  it('renders a link for every article', () => {
    const wrapper = mount(BlogView)

    expect(postLinks(wrapper)).toHaveLength(articlesMeta.length)
  })

  it('does not list release-notes posts', () => {
    const wrapper = mount(BlogView)

    const targets = postLinks(wrapper).map((link) => link.props('to'))
    for (const meta of releasesMeta) {
      expect(targets).not.toContain(`/blog/${meta.slug}`)
    }
  })

  it('links to the changelog', () => {
    const wrapper = mount(BlogView)

    const targets = wrapper
      .findAllComponents({ name: 'RouterLink' })
      .map((link) => link.props('to'))
    expect(targets).toContain('/changelog')
  })

  it('renders posts in reverse (newest-first) order with title, date, and read time', () => {
    const wrapper = mount(BlogView)

    const links = postLinks(wrapper)
    const sorted = [...articlesMeta].reverse()

    expect(links[0].props('to')).toBe(`/blog/${sorted[0].slug}`)
    expect(links[0].text()).toContain(sorted[0].title)
    expect(links[0].text()).toContain(sorted[0].date)
    expect(links[0].text()).toContain(sorted[0].readTime)

    expect(links[links.length - 1].props('to')).toBe(`/blog/${sorted[sorted.length - 1].slug}`)
    expect(links[links.length - 1].text()).toContain(sorted[sorted.length - 1].title)
  })

  it('renders each article excerpt', () => {
    const wrapper = mount(BlogView)

    for (const meta of articlesMeta) {
      expect(wrapper.text()).toContain(meta.excerpt)
    }
  })
})
