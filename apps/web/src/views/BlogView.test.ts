import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import BlogView from './BlogView.vue'
import { posts } from '@/blog/posts'

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
    expect(links).toHaveLength(posts.length)
  })

  it('renders posts in reverse order with title, date, and read time', () => {
    const wrapper = mount(BlogView)

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const lastPost = posts[posts.length - 1]
    const firstPost = posts[0]

    expect(links[0].props('to')).toBe(`/blog/${lastPost.slug}`)
    expect(links[0].text()).toContain(lastPost.title)
    expect(links[0].text()).toContain(lastPost.date)
    expect(links[0].text()).toContain(lastPost.readTime)

    expect(links[links.length - 1].props('to')).toBe(`/blog/${firstPost.slug}`)
    expect(links[links.length - 1].text()).toContain(firstPost.title)
  })

  it('renders each post excerpt', () => {
    const wrapper = mount(BlogView)

    for (const post of posts) {
      expect(wrapper.text()).toContain(post.excerpt)
    }
  })
})
