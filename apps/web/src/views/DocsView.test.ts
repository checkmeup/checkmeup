import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import DocsView from './DocsView.vue'

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

vi.mock('vue-router', () => ({
  RouterLink: { template: '<a><slot /></a>' },
}))

describe('DocsView', () => {
  it('mounts without throwing', () => {
    expect(() => mount(DocsView)).not.toThrow()
  })

  it('renders the hero heading', () => {
    const wrapper = mount(DocsView)

    expect(wrapper.text()).toContain('Documentation')
  })

  it('renders a sidebar nav link for every documented section', () => {
    const wrapper = mount(DocsView)

    const navLinks = wrapper.findAll('nav a')
    const hrefs = navLinks.map((a) => a.attributes('href'))

    expect(hrefs).toEqual([
      '#getting-started',
      '#cron',
      '#uptime',
      '#ssl',
      '#domain',
      '#port',
      '#telegram',
      '#email',
      '#webhook',
      '#slack',
      '#sms',
      '#status-pages',
      '#maintenance',
      '#appearance',
      '#plans',
      '#api',
      '#help',
    ])
  })

  it('renders all key section headings', () => {
    const wrapper = mount(DocsView)

    expect(wrapper.text()).toContain('Getting started')
    expect(wrapper.text()).toContain('Cron job monitoring')
    expect(wrapper.text()).toContain('Uptime monitoring')
    expect(wrapper.text()).toContain('SSL expiry monitoring')
    expect(wrapper.text()).toContain('Domain expiry monitoring')
    expect(wrapper.text()).toContain('Port (TCP) monitoring')
    expect(wrapper.text()).toContain('Telegram alerts')
    expect(wrapper.text()).toContain('Email alerts')
    expect(wrapper.text()).toContain('Slack alerts')
    expect(wrapper.text()).toContain('Status pages')
    expect(wrapper.text()).toContain('Maintenance windows')
    expect(wrapper.text()).toContain('Appearance')
    expect(wrapper.text()).toContain('Plans & limits')
    expect(wrapper.text()).toContain('Public API')
    expect(wrapper.text()).toContain('Need help?')
  })

  it('documents the cron ping URL usage', () => {
    const wrapper = mount(DocsView)

    expect(wrapper.text()).toContain('curl -s https://checkmeup.net/ping/<your-monitor-token>')
  })

  it('documents the public API status endpoint with a curl example', () => {
    const wrapper = mount(DocsView)

    expect(wrapper.text()).toContain('X-API-Key: cmu_live_...')
    expect(wrapper.text()).toContain(
      'https://checkmeup.net/api/v1/public/monitors/cron/<monitor-id>/status',
    )
    expect(wrapper.text()).toContain('lastPingMetadata')
  })

  it('renders the plans and limits table with all four tiers', () => {
    const wrapper = mount(DocsView)

    expect(wrapper.text()).toContain('Hobby — Free')
    expect(wrapper.text()).toContain('Solo — $9/mo')
    expect(wrapper.text()).toContain('Startup — $29/mo')
    expect(wrapper.text()).toContain('Enterprise — $99/mo')
  })

  it('routes help to in-app feedback, email, and GitHub Issues rather than a feature board', () => {
    const wrapper = mount(DocsView)

    expect(wrapper.text()).toContain('Suggest a feature')
    expect(wrapper.text()).toContain('andrew@checkmeup.net')
    expect(wrapper.text()).toContain('Open an issue')

    const githubLink = wrapper
      .findAll('a')
      .find((a) => a.attributes('href') === 'https://github.com/checkmeup/checkmeup/issues')
    expect(githubLink).toBeTruthy()

    const mailLink = wrapper
      .findAll('a')
      .find((a) => a.attributes('href') === 'mailto:andrew@checkmeup.net')
    expect(mailLink).toBeTruthy()
  })
})
