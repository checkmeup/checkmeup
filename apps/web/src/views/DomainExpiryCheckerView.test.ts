import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import DomainExpiryCheckerView from './DomainExpiryCheckerView.vue'
import { domainExpiryFaqs } from '@/content/domainExpiryFaqs'

const flushHead = () => new Promise((resolve) => setTimeout(resolve, 0))

vi.mock('vue-router', () => ({
  RouterLink: { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

vi.mock('@/components/ToolChecker.vue', () => ({
  default: { name: 'ToolChecker', template: '<div />' },
}))

// @unhead/vue keeps a mounted wrapper's JSON-LD in document.head, so each
// test unmounts its wrapper to keep the next one's head clean.
let wrapper: ReturnType<typeof mount> | undefined
afterEach(() => wrapper?.unmount())

describe('DomainExpiryCheckerView', () => {
  it('answers every domain expiry question on the page', () => {
    wrapper = mount(DomainExpiryCheckerView)

    for (const faq of domainExpiryFaqs) {
      expect(wrapper.text()).toContain(faq.q)
      expect(wrapper.text()).toContain(faq.a)
    }
  })

  it('emits FAQPage JSON-LD with the same questions and answers', async () => {
    wrapper = mount(DomainExpiryCheckerView)
    await flushHead()

    const script = document.head.querySelector('script[type="application/ld+json"]')
    const schema = JSON.parse(script?.textContent ?? '{}')
    expect(schema['@type']).toBe('FAQPage')
    expect(schema.mainEntity).toEqual(
      domainExpiryFaqs.map((faq) => ({
        '@type': 'Question',
        name: faq.q,
        acceptedAnswer: { '@type': 'Answer', text: faq.a },
      })),
    )
  })
})
