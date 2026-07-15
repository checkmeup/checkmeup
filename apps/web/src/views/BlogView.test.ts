import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import BlogView from './BlogView.vue'
import { postsMeta } from '@/blog/posts'

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

  it('renders a link for every post', () => {
    const wrapper = mount(BlogView)

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    expect(links).toHaveLength(postsMeta.length)
  })

  it('renders posts in reverse (newest-first) order with title, date, and read time', () => {
    const wrapper = mount(BlogView)

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const sorted = [...postsMeta].reverse()

    expect(links[0].props('to')).toBe(`/blog/${sorted[0].slug}`)
    expect(links[0].text()).toContain(sorted[0].title)
    expect(links[0].text()).toContain(sorted[0].date)
    expect(links[0].text()).toContain(sorted[0].readTime)

    expect(links[links.length - 1].props('to')).toBe(`/blog/${sorted[sorted.length - 1].slug}`)
    expect(links[links.length - 1].text()).toContain(sorted[sorted.length - 1].title)
  })

  it('renders each post excerpt', () => {
    const wrapper = mount(BlogView)

    for (const meta of postsMeta) {
      expect(wrapper.text()).toContain(meta.excerpt)
    }
  })
})
