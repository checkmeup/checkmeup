import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import CronExpressionChecker from './CronExpressionChecker.vue'
import { commonCronExpressions } from '@/content/cronExpression'
import { trackToolCheck } from '@/lib/analytics'

vi.mock('@/lib/analytics', () => ({ trackToolCheck: vi.fn(() => true) }))

const RouterLinkStub = { name: 'RouterLink', props: ['to'], template: '<a :href="to"><slot /></a>' }

const weekdays = {
  expression: '0 9 * * 1-5',
  description: 'At 09:00, Monday through Friday',
  runs: [
    '2026-11-23T09:00:00Z',
    '2026-11-24T09:00:00Z',
    '2026-11-25T09:00:00Z',
    '2026-11-26T09:00:00Z',
    '2026-11-27T09:00:00Z',
  ],
}

function stubFetch(body: unknown, status = 200) {
  const fetch = vi.fn((_url: string) =>
    Promise.resolve({ ok: status < 400, status, json: () => Promise.resolve(body) }),
  )
  vi.stubGlobal('fetch', fetch)
  return fetch
}

async function mountChecker() {
  const wrapper = mount(CronExpressionChecker, {
    global: { stubs: { RouterLink: RouterLinkStub } },
  })
  await flushPromises()
  return wrapper
}

const TZ = process.env.TZ

beforeEach(() => {
  vi.useFakeTimers()
  vi.mocked(trackToolCheck).mockClear()
  vi.mocked(trackToolCheck).mockReturnValue(true)
})

afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
  process.env.TZ = TZ
})

describe('CronExpressionChecker', () => {
  it('explains the expression a moment after typing stops, with its runs in UTC and an offer to monitor it', async () => {
    const fetch = stubFetch(weekdays)
    const wrapper = await mountChecker()
    fetch.mockClear()

    await wrapper.find('input').setValue('0 9 * * 1-')
    await wrapper.find('input').setValue('0 9 * * 1-5')
    expect(fetch).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(400)
    await flushPromises()

    expect(fetch).toHaveBeenCalledTimes(1)
    expect(fetch.mock.calls[0][0]).toBe('/api/v1/tools/cron?expression=0%209%20*%20*%201-5')
    expect(wrapper.find('[data-test="description"]').text()).toBe('At 09:00, Monday through Friday')
    const utc = wrapper.findAll('tbody tr').map((tr) => tr.findAll('td')[0].text())
    expect(utc).toEqual([
      'Mon, Nov 23, 2026, 09:00',
      'Tue, Nov 24, 2026, 09:00',
      'Wed, Nov 25, 2026, 09:00',
      'Thu, Nov 26, 2026, 09:00',
      'Fri, Nov 27, 2026, 09:00',
    ])
    expect(wrapper.find('a[href="/sign-up"]').text()).toContain('Monitor it free')
  })

  it('shows why an expression is not valid instead of a result', async () => {
    stubFetch({ error: 'Minute: 60 is above the maximum, 59' }, 400)
    const wrapper = await mountChecker()

    expect(wrapper.find('[role="alert"]').text()).toBe('Minute: 60 is above the maximum, 59')
    expect(wrapper.find('[data-test="description"]').exists()).toBe(false)
  })

  it('puts a common expression in the input and explains it at once', async () => {
    const fetch = stubFetch(weekdays)
    const wrapper = await mountChecker()
    fetch.mockClear()

    const button = wrapper.findAll('button').find((b) => b.text() === '0 9 * * 1-5')
    await button!.trigger('click')
    await flushPromises()
    await vi.advanceTimersByTimeAsync(400)

    expect((wrapper.find('input').element as HTMLInputElement).value).toBe('0 9 * * 1-5')
    expect(fetch).toHaveBeenCalledTimes(1)
    expect(wrapper.find('[data-test="description"]').text()).toBe('At 09:00, Monday through Friday')
    expect(wrapper.findAll('button').map((b) => b.text())).toEqual(commonCronExpressions)
  })

  it("shows each run in the browser's time zone beside UTC", async () => {
    process.env.TZ = 'Asia/Tokyo'
    stubFetch(weekdays)
    const wrapper = await mountChecker()

    expect(wrapper.findAll('thead th')[1].text()).toBe('Your time (Asia/Tokyo)')
    expect(wrapper.findAll('tbody tr')[0].findAll('td')[1].text()).toBe('Mon, Nov 23, 2026, 18:00')
  })

  it("counts the visitor's first own explanation in GA4, not the example on page load", async () => {
    stubFetch(weekdays)
    const wrapper = await mountChecker()
    expect(trackToolCheck).not.toHaveBeenCalled()

    await wrapper.find('input').setValue('0 9 * * 1-5')
    await vi.advanceTimersByTimeAsync(400)
    await flushPromises()
    await wrapper.find('input').setValue('0 10 * * 1-5')
    await vi.advanceTimersByTimeAsync(400)
    await flushPromises()

    expect(trackToolCheck).toHaveBeenCalledTimes(1)
    expect(trackToolCheck).toHaveBeenCalledWith('cron', true)
  })

  it('counts a later explanation when the first went unsent before consent', async () => {
    vi.mocked(trackToolCheck).mockReturnValueOnce(false)
    stubFetch(weekdays)
    const wrapper = await mountChecker()

    await wrapper.find('input').setValue('0 9 * * 1-5')
    await vi.advanceTimersByTimeAsync(400)
    await flushPromises()
    await wrapper.find('input').setValue('0 10 * * 1-5')
    await vi.advanceTimersByTimeAsync(400)
    await flushPromises()

    expect(trackToolCheck).toHaveBeenCalledTimes(2)
  })
})
