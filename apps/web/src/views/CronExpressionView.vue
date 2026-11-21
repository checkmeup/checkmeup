<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { useHead } from '@unhead/vue'
import LandingLayout from '@/layouts/LandingLayout.vue'
import PageHero from '@/components/PageHero.vue'
import CtaBanner from '@/components/CtaBanner.vue'
import CronExpressionChecker from '@/components/CronExpressionChecker.vue'
import { useSeo } from '@/composables/useSeo'
import { cronCharacters, cronExpressionFaqs, cronFields } from '@/content/cronExpression'

useSeo({
  title: 'Cron Expression Explainer: Read Any Cron Schedule — Checkmeup',
  description:
    'Paste a cron expression to see what it means in plain English and its next five run times, in UTC and your time zone. Free, no sign-up.',
  path: '/tools/cron-expression',
})

// Makes the questions below eligible for Google's expandable FAQ rich results.
const faqSchema = {
  '@context': 'https://schema.org',
  '@type': 'FAQPage',
  mainEntity: cronExpressionFaqs.map((faq) => ({
    '@type': 'Question',
    name: faq.q,
    acceptedAnswer: { '@type': 'Answer', text: faq.a },
  })),
}

useHead({
  script: [{ type: 'application/ld+json', innerHTML: JSON.stringify(faqSchema) }],
})
</script>

<template>
  <LandingLayout>
    <PageHero title="Cron expression explainer" eyebrow="Free tool · No sign-up">
      <p>
        Type a cron expression to see what it means in plain English and when it runs next, in
        UTC and in your own time zone.
      </p>
    </PageHero>

    <div class="mx-auto max-w-[1200px] px-4 pb-20 sm:px-6">
      <div class="max-w-2xl">
        <CronExpressionChecker />
      </div>

      <div class="mt-16 max-w-2xl space-y-10 text-[15px] leading-relaxed text-[var(--text-dim)]">
        <section>
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            The five fields
          </h2>
          <p class="mb-4">A cron expression is five fields separated by spaces, in this order:</p>
          <table class="w-full text-sm" data-test="fields">
            <tbody>
              <tr v-for="field in cronFields" :key="field.name" class="border-b border-[var(--border)] last:border-0">
                <td class="py-2 pr-4 font-medium text-[var(--text-strong)]">{{ field.name }}</td>
                <td class="py-2 font-mono">{{ field.values }}</td>
              </tr>
            </tbody>
          </table>
        </section>
        <section>
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            Special characters
          </h2>
          <table class="w-full text-sm" data-test="characters">
            <tbody>
              <tr v-for="c in cronCharacters" :key="c.char" class="border-b border-[var(--border)] last:border-0">
                <td class="w-10 py-2 pr-4 font-mono font-semibold text-[var(--text-strong)]">{{ c.char }}</td>
                <td class="py-2 pr-4">{{ c.meaning }}</td>
                <td class="py-2">
                  <code class="font-mono text-[var(--text-strong)]">{{ c.example }}</code> runs {{ c.reads }}
                </td>
              </tr>
            </tbody>
          </table>
        </section>
        <section>
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            Know when a scheduled job doesn't run
          </h2>
          <p>
            A Checkmeup
            <RouterLink to="/cron-monitoring" class="text-[var(--accent)] hover:underline">cron monitor</RouterLink>
            reads the same expression and alerts you by email, Telegram, Slack or webhook when a run
            doesn't check in on time. Read the
            <RouterLink to="/blog/cron-job-monitoring-guide" class="text-[var(--accent)] hover:underline"
              >cron job monitoring guide</RouterLink
            >, or how to catch a
            <RouterLink to="/blog/github-actions-cron-monitoring-guide" class="text-[var(--accent)] hover:underline"
              >GitHub Actions schedule</RouterLink
            >
            that stopped. The free plan includes 10 monitors.
          </p>
        </section>
        <section id="faq">
          <h2 class="mb-6 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            Cron expression questions
          </h2>
          <div class="space-y-6">
            <div v-for="faq in cronExpressionFaqs" :key="faq.q">
              <h3 class="mb-2 text-[17px] font-semibold text-[var(--text-strong)]">{{ faq.q }}</h3>
              <p>{{ faq.a }}</p>
            </div>
          </div>
        </section>
      </div>
    </div>

    <CtaBanner
      title="Know when a cron job stops."
      subtitle="10 monitors, no credit card, no expiry date on the free plan."
    />
  </LandingLayout>
</template>
