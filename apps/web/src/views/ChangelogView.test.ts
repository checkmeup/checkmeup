import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ChangelogView from './ChangelogView.vue'
import { articlesMeta, releasesMeta } from '@/blog/posts'

vi.mock('vue-router', () => ({
  RouterLink: { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

describe('ChangelogView', () => {
  it('renders the page heading', () => {
    const wrapper = mount(ChangelogView)

    expect(wrapper.find('h1').text()).toBe('Changelog')
  })

  it('renders one entry per release, newest first, each anchored by its slug', () => {
    const wrapper = mount(ChangelogView)

    const entries = wrapper.findAll('article')
    const newestFirst = [...releasesMeta].reverse()
    expect(entries).toHaveLength(releasesMeta.length)
    expect(entries.map((e) => e.attributes('id'))).toEqual(newestFirst.map((m) => m.slug))
    expect(entries[0].text()).toContain(newestFirst[0].title)
    expect(entries[0].text()).toContain(newestFirst[0].date)
    expect(entries[0].text()).toContain(newestFirst[0].excerpt)
  })

  it('lists releases by version, newest first, including several from the same day', () => {
    const wrapper = mount(ChangelogView)

    // "v1.49–v1.51: …" → 1.49, compared as [major, minor].
    const versions = wrapper.findAll('article h2').map((h) => {
      const m = /^v(\d+)\.(\d+)/.exec(h.text())!
      return Number(m[1]) * 1000 + Number(m[2])
    })
    expect(versions.length).toBe(releasesMeta.length)
    expect(versions).toEqual([...versions].sort((a, b) => b - a))
  })

  it('links every release to its full notes', () => {
    const wrapper = mount(ChangelogView)

    const targets = wrapper
      .findAllComponents({ name: 'RouterLink' })
      .map((link) => link.props('to'))
    for (const meta of releasesMeta) {
      expect(targets).toContain(`/blog/${meta.slug}`)
    }
  })

  it('does not list articles', () => {
    const wrapper = mount(ChangelogView)

    const targets = wrapper
      .findAllComponents({ name: 'RouterLink' })
      .map((link) => link.props('to'))
    for (const meta of articlesMeta) {
      expect(targets).not.toContain(`/blog/${meta.slug}`)
    }
  })
})
