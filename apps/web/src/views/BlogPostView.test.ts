import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import BlogPostView from './BlogPostView.vue'
import { posts } from '@/blog/posts'

const routeParams = { slug: '' }

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: routeParams }),
  RouterLink: { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

afterEach(() => {
  vi.clearAllMocks()
})

describe('BlogPostView', () => {
  it('renders the matched post title, date, read time, and excerpt', () => {
    const post = posts[0]
    routeParams.slug = post.slug
    const wrapper = mount(BlogPostView)

    expect(wrapper.text()).toContain(post.title)
    expect(wrapper.text()).toContain(post.date)
    expect(wrapper.text()).toContain(post.readTime)
    expect(wrapper.text()).toContain(post.excerpt)
  })

  it('renders paragraph and heading content blocks', () => {
    const post = posts.find(
      (p) => p.content.some((b) => b.type === 'p') && p.content.some((b) => b.type === 'h2'),
    )!
    routeParams.slug = post.slug
    const wrapper = mount(BlogPostView)

    const firstParagraph = post.content.find((b) => b.type === 'p')
    const firstH2 = post.content.find((b) => b.type === 'h2')

    expect(wrapper.text()).toContain(firstParagraph!.text)
    expect(wrapper.text()).toContain(firstH2!.text)
  })

  it('shows a not-found message for an unknown slug', () => {
    routeParams.slug = 'this-slug-does-not-exist'
    const wrapper = mount(BlogPostView)

    expect(wrapper.text()).toContain('Post not found.')
  })

  it('links back to the blog list from the not-found state', () => {
    routeParams.slug = 'this-slug-does-not-exist'
    const wrapper = mount(BlogPostView)

    const backLink = wrapper.findComponent({ name: 'RouterLink' })
    expect(backLink.exists()).toBe(true)
    expect(backLink.props('to')).toBe('/blog')
  })

  it('does not render post content when the slug is not found', () => {
    routeParams.slug = 'this-slug-does-not-exist'
    const wrapper = mount(BlogPostView)

    expect(wrapper.find('article').exists()).toBe(false)
  })

  it('renders a sign-up CTA link for a matched post', () => {
    const post = posts[0]
    routeParams.slug = post.slug
    const wrapper = mount(BlogPostView)

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const cta = links.find((c) => c.props('to') === '/sign-up')
    expect(cta).toBeTruthy()
  })
})
