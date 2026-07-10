import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import NotFoundHero from './NotFoundHero.vue'

vi.mock('vue-router', () => ({
  RouterLink: { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' },
}))

const props = {
  badge: 'Post status: not found',
  heading: 'This post is gone.',
  description: 'It moved or never existed.',
  primaryCta: { label: 'Back to blog', to: '/blog' },
  secondaryCta: { label: 'Go to homepage', to: '/' },
}

describe('NotFoundHero', () => {
  it('renders the badge, heading, description, and both CTAs', () => {
    const wrapper = mount(NotFoundHero, { props })

    expect(wrapper.text()).toContain(props.badge)
    expect(wrapper.text()).toContain(props.heading)
    expect(wrapper.text()).toContain(props.description)

    const links = wrapper.findAllComponents({ name: 'RouterLink' })
    const targets = links.map((link) => link.props('to'))
    expect(targets).toEqual([props.primaryCta.to, props.secondaryCta.to])
  })
})
