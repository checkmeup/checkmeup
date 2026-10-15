import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import PlatformStats from './PlatformStats.vue'

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('PlatformStats', () => {
  it('shows the live 30-day totals once they arrive', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ monitors: 1234, pings30d: 98765, checks30d: 4321000, incidents30d: 57 }),
      }),
    )
    const wrapper = mount(PlatformStats)
    await flushPromises()

    expect(fetch).toHaveBeenCalledWith('/api/v1/stats')
    expect(wrapper.text()).toContain('1,234')
    expect(wrapper.text()).toContain('98,765')
    expect(wrapper.text()).toContain('4,321,000')
    expect(wrapper.text()).toContain('57')
  })

  it('renders nothing when the request fails', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false }))
    const wrapper = mount(PlatformStats)
    await flushPromises()

    expect(wrapper.find('section').exists()).toBe(false)
  })

  it('renders nothing on a network error', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('offline')))
    const wrapper = mount(PlatformStats)
    await flushPromises()

    expect(wrapper.find('section').exists()).toBe(false)
  })
})
