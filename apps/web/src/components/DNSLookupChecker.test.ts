import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import DNSLookupChecker from './DNSLookupChecker.vue'

const RouterLinkStub = { name: 'RouterLink', props: ['to'], template: '<a :href="to"><slot /></a>' }

async function lookUp(body: unknown, status = 200) {
  const fetch = vi.fn((_url: string) =>
    Promise.resolve({ ok: status < 400, status, json: () => Promise.resolve(body) }),
  )
  vi.stubGlobal('fetch', fetch)
  const wrapper = mount(DNSLookupChecker, { global: { stubs: { RouterLink: RouterLinkStub } } })
  await wrapper.find('input').setValue('example.com')
  await wrapper.find('form').trigger('submit')
  await flushPromises()
  return { wrapper, fetch }
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('DNSLookupChecker', () => {
  it('lists each record type with its values, with an offer to monitor them', async () => {
    const { wrapper, fetch } = await lookUp({
      host: 'example.com',
      records: [
        { type: 'A', values: ['93.184.215.14'] },
        { type: 'AAAA', values: [] },
        { type: 'CNAME', values: [] },
        { type: 'MX', values: ['10 mail.example.com', '20 backup.example.com'] },
        { type: 'TXT', values: ['v=spf1 -all'] },
        { type: 'NS', values: [], error: 'DNS lookup timeout' },
      ],
    })

    expect(fetch.mock.calls[0][0]).toBe('/api/v1/tools/dns?host=example.com')
    const rows = wrapper.findAll('tbody tr').map((tr) => tr.findAll('td').map((td) => td.text()))
    expect(rows).toEqual([
      ['A', '93.184.215.14'],
      ['AAAA', 'No records'],
      ['CNAME', 'No records'],
      ['MX', '10 mail.example.com20 backup.example.com'],
      ['TXT', 'v=spf1 -all'],
      ['NS', 'DNS lookup timeout'],
    ])
    expect(wrapper.find('a[href="/sign-up"]').text()).toContain('Monitor it free')
  })

  it('shows why a lookup failed instead of a result', async () => {
    const { wrapper } = await lookUp(
      { error: 'missing.example.com has no DNS records', code: 'check_failed' },
      422,
    )

    expect(wrapper.find('[role="alert"]').text()).toBe('missing.example.com has no DNS records')
    expect(wrapper.find('table').exists()).toBe(false)
  })
})
