import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ToolChecker from './ToolChecker.vue'

const RouterLinkStub = { name: 'RouterLink', props: ['to'], template: '<a :href="to"><slot /></a>' }

function mountChecker() {
  return mount(ToolChecker, {
    props: {
      tool: 'domain',
      endpoint: '/api/v1/tools/domain',
      param: 'domain',
      inputLabel: 'Domain',
      placeholder: 'example.com',
      detailKey: 'registrar',
      detailLabel: 'Registrar',
      monitorCta: 'Get an alert before this domain expires.',
    },
    global: { stubs: { RouterLink: RouterLinkStub } },
  })
}

async function checkDomain(body: unknown, status = 200) {
  vi.stubGlobal(
    'fetch',
    vi.fn(() => Promise.resolve({ ok: status < 400, status, json: () => Promise.resolve(body) })),
  )
  const wrapper = mountChecker()
  await wrapper.find('input').setValue('example.com')
  await wrapper.find('form').trigger('submit')
  await flushPromises()
  return wrapper
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('ToolChecker', () => {
  it('shows the expiry, days left and registrar, with an offer to monitor it', async () => {
    const wrapper = await checkDomain({
      domain: 'example.com',
      expiresAt: '2027-06-01T00:00:00Z',
      daysLeft: 12,
      registrar: 'Example Registrar',
    })

    expect(wrapper.text()).toContain('Expires in 12 days')
    expect(wrapper.text()).toContain('June 1, 2027')
    expect(wrapper.text()).toContain('Example Registrar')
    expect(wrapper.text()).toContain('Get an alert before this domain expires.')
    expect(wrapper.find('a[href="/sign-up"]').text()).toContain('Monitor it free')
  })

  it('shows why a check failed instead of a result', async () => {
    const wrapper = await checkDomain(
      { error: 'domain not found in registry (whois)', code: 'check_failed' },
      422,
    )

    expect(wrapper.find('[role="alert"]').text()).toBe('domain not found in registry (whois)')
    expect(wrapper.find('a[href="/sign-up"]').exists()).toBe(false)
  })
})
