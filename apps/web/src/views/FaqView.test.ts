import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import FaqView from './FaqView.vue'
import { faqCategories } from '@/faq/faqs'

vi.mock('vue-router', () => ({
  RouterLink: { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

describe('FaqView', () => {
  it('mounts without throwing', () => {
    expect(() => mount(FaqView)).not.toThrow()
  })

  it('renders the page heading', () => {
    const wrapper = mount(FaqView)

    expect(wrapper.text()).toContain('Frequently asked questions')
  })

  it('renders a sidebar nav link for every category', () => {
    const wrapper = mount(FaqView)

    for (const category of faqCategories) {
      const link = wrapper.find(`nav a[href="#${category.id}"]`)
      expect(link.exists()).toBe(true)
      expect(link.text()).toBe(category.label)
    }
  })

  it('renders every category heading and its questions and answers', () => {
    const wrapper = mount(FaqView)

    for (const category of faqCategories) {
      const section = wrapper.find(`#${category.id}`)
      expect(section.exists()).toBe(true)
      expect(section.text()).toContain(category.label)

      for (const entry of category.entries) {
        expect(section.text()).toContain(entry.q)
        expect(section.text()).toContain(entry.a)
      }
    }
  })
})
