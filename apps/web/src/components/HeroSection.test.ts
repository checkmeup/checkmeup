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
  // The oldest 60 days ran slower than the last 30.
  days: Array.from({ length: 90 }, (_, i) => ({
    date: `day ${i}`,
    tone: 'up' as const,
    uptime: 100,
    latencyMs: i < 60 ? 300 : 150,
  })),
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

  it('shows the live status, and the platform totals big enough to show, once they arrive', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn((url: string) =>
        Promise.resolve(
          url === '/api/v1/stats'
            ? {
                ok: true,
                json: async () => ({ monitors: 9, pings30d: 0, checks30d: 28396, incidents30d: 7 }),
              }
            : { ok: true, json: async () => summary },
        ),
      ),
    )
    const wrapper = mountHero()
    await flushPromises()

    expect(wrapper.text()).toContain('28,396 checks across Checkmeup in the last 30 days')
    // 90 days for wide screens, the last 30 for phones; the response time
    // stays in the day tooltips, not this line.
    const numbers = wrapper.find('p.min-h-5').text().replace(/\s+/g, ' ')
    expect(numbers).toContain('100% uptime · last 90 days')
    expect(numbers).toContain('100% uptime · last 30 days')
    expect(numbers).not.toContain('ms')
    expect(wrapper.text()).not.toContain('cron pings')
    expect(wrapper.text()).toContain('All systems operational')
    expect(wrapper.text()).not.toContain('Loading status')
    expect(wrapper.find('a[aria-label="checkmeup.net status page"]').exists()).toBe(true)
  })

  it('falls back to the labelled example when the status can not load', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('offline')))
    const wrapper = mountHero()
    await flushPromises()

    expect(wrapper.text()).toContain('nightly-backup · Cron · example')
    expect(wrapper.text()).toContain('Down')
    expect(wrapper.text()).not.toContain('Loading status')
  })
})
