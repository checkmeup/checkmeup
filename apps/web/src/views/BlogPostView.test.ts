import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import BlogPostView from './BlogPostView.vue'
import { posts } from '@/blog/posts'

const flushHead = () => new Promise((resolve) => setTimeout(resolve, 0))

const routeParams = { slug: '' }

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: routeParams }),
  RouterLink: { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

// @unhead/vue ties head entries (script/meta tags) to the mounting
// component's lifecycle — leaving a wrapper mounted across tests lets one
// test's JSON-LD/robots tags leak into the next test's document.head
// assertions, so every wrapper gets unmounted here rather than per-test.
const mountedWrappers: ReturnType<typeof mount>[] = []
function mountTracked(component: Parameters<typeof mount>[0]) {
  const wrapper = mount(component)
  mountedWrappers.push(wrapper)
  return wrapper
}

afterEach(() => {
  mountedWrappers.splice(0).forEach((wrapper) => wrapper.unmount())
  vi.clearAllMocks()
})

describe('BlogPostView', () => {
  it('renders the matched post title, date, read time, and excerpt', () => {
    const post = posts[0]
    routeParams.slug = post.slug
    const wrapper = mountTracked(BlogPostView)

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
    const wrapper = mountTracked(BlogPostView)

    const firstParagraph = post.content.find((b) => b.type === 'p')
    const firstH2 = post.content.find((b) => b.type === 'h2')

    expect(wrapper.text()).toContain(firstParagraph!.text)
    expect(wrapper.text()).toContain(firstH2!.text)
  })

  it('shows a not-found message for an unknown slug', () => {
    routeParams.slug = 'this-slug-does-not-exist'
    const wrapper = mountTracked(BlogPostView)

    expect(wrapper.text()).toContain("doesn't exist, or it moved without telling anyone")
  })

  it('links back to the blog list and homepage from the not-found state', () => {
    routeParams.slug = 'this-slug-does-not-exist'
    const wrapper = mountTracked(BlogPostView)

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const targets = links.map((link) => link.props('to'))
    expect(targets).toContain('/blog')
    expect(targets).toContain('/')
  })

  it('does not render post content when the slug is not found', () => {
    routeParams.slug = 'this-slug-does-not-exist'
    const wrapper = mountTracked(BlogPostView)

    expect(wrapper.find('article').exists()).toBe(false)
  })

  it('renders a sign-up CTA link for a matched post', () => {
    const post = posts[0]
    routeParams.slug = post.slug
    const wrapper = mountTracked(BlogPostView)

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const cta = links.find((c) => c.props('to') === '/sign-up')
    expect(cta).toBeTruthy()
  })

  it('emits Article JSON-LD for a matched post', async () => {
    const post = posts[0]
    routeParams.slug = post.slug
    mountTracked(BlogPostView)
    await nextTick()
    await flushHead()

    const script = document.head.querySelector('script[type="application/ld+json"]')
    expect(script).toBeTruthy()
    const schema = JSON.parse(script!.innerHTML)
    expect(schema['@type']).toBe('Article')
    expect(schema.headline).toBe(post.title)
    expect(schema.datePublished).toMatch(/^\d{4}-\d{2}-\d{2}T/)
    expect(schema.author).toEqual({
      '@type': 'Person',
      name: 'Andrew Molyuk',
      url: 'https://checkmeup.net/about',
    })
  })

  it('emits no JSON-LD script for an unknown slug', async () => {
    routeParams.slug = 'this-slug-does-not-exist'
    mountTracked(BlogPostView)
    await nextTick()
    await flushHead()

    expect(document.head.querySelector('script[type="application/ld+json"]')).toBeFalsy()
  })
})
