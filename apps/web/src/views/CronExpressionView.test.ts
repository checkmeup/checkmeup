import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import CronExpressionView from './CronExpressionView.vue'
import { cronExpressionFaqs } from '@/content/cronExpression'

const flushHead = () => new Promise((resolve) => setTimeout(resolve, 0))

vi.mock('vue-router', () => ({
  RouterLink: { name: 'RouterLink', props: ['to'], template: '<a><slot /></a>' },
}))

vi.mock('@/layouts/LandingLayout.vue', () => ({
  default: { name: 'LandingLayout', template: '<div><slot /></div>' },
}))

vi.mock('@/components/CronExpressionChecker.vue', () => ({
  default: { name: 'CronExpressionChecker', template: '<div />' },
}))

// @unhead/vue keeps a mounted wrapper's JSON-LD in document.head, so each
// test unmounts its wrapper to keep the next one's head clean.
let wrapper: ReturnType<typeof mount> | undefined
afterEach(() => wrapper?.unmount())

describe('CronExpressionView', () => {
  it('names each field with its allowed values and gives an example of each special character', () => {
    wrapper = mount(CronExpressionView)

    const fields = wrapper.findAll('[data-test="fields"] tr').map((tr) => tr.text())
    expect(fields).toEqual([
      'Minute0–59',
      'Hour0–23',
      'Day of month1–31',
      'Month1–12 or JAN–DEC',
      'Day of week0–6 or SUN–SAT (0 is Sunday)',
    ])
    const chars = wrapper.findAll('[data-test="characters"] tr').map((tr) => tr.find('td').text())
    expect(chars).toEqual(['*', ',', '-', '/'])
  })

  it('answers every cron expression question on the page', () => {
    wrapper = mount(CronExpressionView)

    expect(cronExpressionFaqs.map((faq) => faq.q)).toEqual([
      'What is a cron expression?',
      'Which time zone does cron use?',
      'How do I run a cron job every 5 minutes?',
      'What happens when I set both the day of month and the day of week?',
      'How do I know a cron job stopped running?',
    ])
    for (const faq of cronExpressionFaqs) {
      expect(wrapper.text()).toContain(faq.q)
      expect(wrapper.text()).toContain(faq.a)
    }
  })

  it('emits FAQPage JSON-LD with the same questions and answers', async () => {
    wrapper = mount(CronExpressionView)
    await flushHead()

    const script = document.head.querySelector('script[type="application/ld+json"]')
    const schema = JSON.parse(script?.textContent ?? '{}')
    expect(schema['@type']).toBe('FAQPage')
    expect(schema.mainEntity).toEqual(
      cronExpressionFaqs.map((faq) => ({
        '@type': 'Question',
        name: faq.q,
        acceptedAnswer: { '@type': 'Answer', text: faq.a },
      })),
    )
  })
})
