import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import NotFoundView from './NotFoundView.vue'

vi.mock('vue-router', () => ({
  RouterLink: { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' },
  useRoute: () => ({ fullPath: '/nonexistent-page' }),
}))

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

const flush = () => new Promise((resolve) => setTimeout(resolve, 0))

describe('NotFoundView', () => {
  it('mounts without throwing', () => {
    expect(() => mount(NotFoundView)).not.toThrow()
  })

  it('renders a 404 message with links back to the dashboard and homepage', () => {
    const wrapper = mount(NotFoundView)

    expect(wrapper.text()).toContain('404')
    expect(wrapper.text()).toContain('went down and stayed down')

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const targets = links.map((link) => link.props('to'))
    expect(targets).toContain('/dashboard')
    expect(targets).toContain('/')
  })

  it('sets a noindex robots meta tag', async () => {
    mount(NotFoundView)
    await nextTick()
    await flush()

    const robots = document.head.querySelector('meta[name="robots"]')
    expect(robots?.getAttribute('content')).toBe('noindex')
  })
})
