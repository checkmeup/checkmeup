import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import CheckStrip from './CheckStrip.vue'

const checks = Array.from({ length: 90 }, (_, i) => ({ ok: true, label: `day ${i}` }))

describe('CheckStrip', () => {
  it("shows each tick's label as its tooltip", () => {
    const ticks = mount(CheckStrip, { props: { checks } }).findAll('span')

    expect(ticks[0].attributes('title')).toBe('day 0')
    expect(ticks[89].attributes('title')).toBe('day 89')
  })

  it('hides all but the newest narrowLast ticks on narrow screens', () => {
    const ticks = mount(CheckStrip, { props: { checks, narrowLast: 30 } }).findAll('span')

    expect(ticks.filter((t) => t.classes('max-sm:hidden'))).toHaveLength(60)
    expect(ticks[59].classes()).toContain('max-sm:hidden')
    expect(ticks[60].classes()).not.toContain('max-sm:hidden')
  })

  it('hides nothing without narrowLast', () => {
    const ticks = mount(CheckStrip, { props: { checks } }).findAll('span')

    expect(ticks.some((t) => t.classes('max-sm:hidden'))).toBe(false)
  })
})
