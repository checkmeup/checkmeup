import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import HeroSection from './HeroSection.vue'
import type { StatusSummary } from '@/lib/statusSummary'

const RouterLinkStub = { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' }

function mountHero() {
  return mount(HeroSection, { global: { stubs: { RouterLink: RouterLinkStub } } })
}

const summary: StatusSummary = {
  title: 'checkmeup.net',
  overall: 'All systems operational',
  tone: 'up',
  monitors: 3,
  days: Array.from({ length: 90 }, (_, i) => ({ date: `day ${i}`, tone: 'up' as const })),
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('HeroSection status strip', () => {
  it('shows a skeleton, not the example, while the live status loads', () => {
    vi.stubGlobal('fetch', vi.fn(() => new Promise(() => {})))
    const wrapper = mountHero()

    expect(wrapper.text()).toContain('checkmeup.net')
    expect(wrapper.text()).toContain('Loading status')
    expect(wrapper.text()).not.toContain('nightly-backup · Cron')
    expect(wrapper.text()).not.toContain('Down')
    const ticks = wrapper.findAll('[aria-hidden="true"] span')
    expect(ticks).toHaveLength(90)
    expect(ticks.every((t) => t.attributes('style')?.includes('var(--tick-empty)'))).toBe(true)
  })

  it('shows the live status once it arrives', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => summary }))
    const wrapper = mountHero()
    await flushPromises()

    expect(wrapper.text()).toContain('All systems operational')
    expect(wrapper.text()).not.toContain('Loading status')
    expect(wrapper.find('a[aria-label="checkmeup.net status page"]').exists()).toBe(true)
  })

  it('falls back to the labelled example when the status can not load', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('offline')))
    const wrapper = mountHero()
    await flushPromises()

    expect(wrapper.text()).toContain('example, last')
    expect(wrapper.text()).toContain('Down')
    expect(wrapper.text()).not.toContain('Loading status')
  })
})
