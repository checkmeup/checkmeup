import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import BlogPostView from './BlogPostView.vue'
import { postsMeta, getPost } from '@/blog/posts'

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
async function mountLoaded(slug: string) {
  routeParams.slug = slug
  const wrapper = mount(BlogPostView)
  mountedWrappers.push(wrapper)
  // onMounted's async loadPost() hasn't resolved yet right after mount() —
  // getPost() goes through a real (unmocked) dynamic import() of the post's
  // own chunk, which can take more than one microtask tick to settle, unlike
  // a plain awaited promise — poll until the loading state actually clears
  // instead of assuming a fixed number of flushes/ticks is enough.
  await vi.waitFor(() => {
    if ((wrapper.vm as unknown as { loading: boolean }).loading) throw new Error('still loading')
  })
  await wrapper.vm.$nextTick()
  return wrapper
}

afterEach(() => {
  mountedWrappers.splice(0).forEach((wrapper) => wrapper.unmount())
  vi.clearAllMocks()
})

describe('BlogPostView', () => {
  it('renders the matched post title, date, read time, and excerpt', async () => {
    const meta = postsMeta[0]
    const wrapper = await mountLoaded(meta.slug)

    expect(wrapper.text()).toContain(meta.title)
    expect(wrapper.text()).toContain(meta.date)
    expect(wrapper.text()).toContain(meta.readTime)
    expect(wrapper.text()).toContain(meta.excerpt)
  })

  it('renders paragraph and heading content blocks', async () => {
    // Metadata alone doesn't say what content blocks a post has — load each
    // full post (cheap in a test env, no real network) to find one with
    // both a paragraph and an h2 to assert against.
    let post
    for (const m of postsMeta) {
      const candidate = await getPost(m.slug)
      if (
        candidate?.content.some((b) => b.type === 'p') &&
        candidate?.content.some((b) => b.type === 'h2')
      ) {
        post = candidate
        break
      }
    }
    const wrapper = await mountLoaded(post!.slug)

    const firstParagraph = post!.content.find((b) => b.type === 'p')
    const firstH2 = post!.content.find((b) => b.type === 'h2')

    expect(wrapper.text()).toContain(firstParagraph!.text)
    expect(wrapper.text()).toContain(firstH2!.text)
  })

  it('shows a not-found message for an unknown slug', async () => {
    const wrapper = await mountLoaded('this-slug-does-not-exist')

    expect(wrapper.text()).toContain("doesn't exist, or it moved without telling anyone")
  })

  it('links back to the blog list and homepage from the not-found state', async () => {
    const wrapper = await mountLoaded('this-slug-does-not-exist')

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const targets = links.map((link) => link.props('to'))
    expect(targets).toContain('/blog')
    expect(targets).toContain('/')
  })

  it('does not render post content when the slug is not found', async () => {
    const wrapper = await mountLoaded('this-slug-does-not-exist')

    expect(wrapper.find('article').exists()).toBe(false)
  })

  it('renders a sign-up CTA link for a matched post', async () => {
    const meta = postsMeta[0]
    const wrapper = await mountLoaded(meta.slug)

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const cta = links.find((c) => c.props('to') === '/sign-up')
    expect(cta).toBeTruthy()
  })

  it('emits Article JSON-LD for a matched post', async () => {
    const meta = postsMeta[0]
    await mountLoaded(meta.slug)
    await flushHead()

    const script = document.head.querySelector('script[type="application/ld+json"]')
    expect(script).toBeTruthy()
    const schema = JSON.parse(script!.innerHTML)
    expect(schema['@type']).toBe('Article')
    expect(schema.headline).toBe(meta.title)
    expect(schema.datePublished).toMatch(/^\d{4}-\d{2}-\d{2}T/)
    expect(schema.author).toEqual({
      '@type': 'Person',
      name: 'Andrew Molyuk',
      url: 'https://checkmeup.net/about',
    })
  })

  it('emits no JSON-LD script for an unknown slug', async () => {
    await mountLoaded('this-slug-does-not-exist')
    await flushHead()

    expect(document.head.querySelector('script[type="application/ld+json"]')).toBeFalsy()
  })
})
