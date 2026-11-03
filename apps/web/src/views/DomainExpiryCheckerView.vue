<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { useHead } from '@unhead/vue'
import LandingLayout from '@/layouts/LandingLayout.vue'
import PageHero from '@/components/PageHero.vue'
import CtaBanner from '@/components/CtaBanner.vue'
import ToolChecker from '@/components/ToolChecker.vue'
import { useSeo } from '@/composables/useSeo'
import { domainExpiryFaqs } from '@/content/domainExpiryFaqs'

useSeo({
  title: 'Domain Expiry Checker: Free Domain Expiration Lookup — Checkmeup',
  description:
    'Look up when any domain expires: its expiration date, days left and registrar, straight from the registry over RDAP or WHOIS. Free, no sign-up.',
  path: '/tools/domain-expiry-checker',
})

// Makes the questions below eligible for Google's expandable FAQ rich results.
const faqSchema = {
  '@context': 'https://schema.org',
  '@type': 'FAQPage',
  mainEntity: domainExpiryFaqs.map((faq) => ({
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
    <PageHero title="Domain expiry checker" eyebrow="Free tool · No sign-up">
      <p>
        Look up when a domain expires: enter it to see its expiration date, how many days are
        left, and which registrar holds it.
      </p>
    </PageHero>

    <div class="mx-auto max-w-[1200px] px-4 pb-20 sm:px-6">
      <div class="max-w-2xl">
        <ToolChecker
          tool="domain"
          endpoint="/api/v1/tools/domain"
          param="domain"
          input-label="Domain"
          placeholder="example.com"
          detail-key="registrar"
          detail-label="Registrar"
          monitor-cta="Get an alert 30, 14 and 7 days before this domain expires."
        />
      </div>

      <div class="mt-16 max-w-2xl space-y-10 text-[15px] leading-relaxed text-[var(--text-dim)]">
        <section>
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            Where the date comes from
          </h2>
          <p>
            The checker asks the domain's registry directly. Most registries answer over RDAP, the
            structured successor to WHOIS. For those that don't, such as .io, .co and .me, it reads
            the registry's WHOIS answer instead. A few registries, among them .de, .eu and .at,
            publish no expiry date at all, so no tool can show one. More in
            <RouterLink to="/blog/rdap-vs-whois-domain-expiry" class="text-[var(--accent)] hover:underline"
              >RDAP vs WHOIS: why some domains hide their expiry</RouterLink
            >.
          </p>
        </section>
        <section>
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            Why domains lapse
          </h2>
          <p>
            Renewal emails go to an inbox nobody reads, a card on file expires, or the domain was
            registered by someone who has since left. A lapsed domain takes the site, email and
            every certificate down with it. Read
            <RouterLink to="/blog/domain-expiry-monitoring" class="text-[var(--accent)] hover:underline"
              >why domains expire silently</RouterLink
            >.
          </p>
        </section>
        <section>
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            Check it every day instead
          </h2>
          <p>
            A Checkmeup domain monitor runs this lookup daily and alerts you by email, Telegram,
            Slack or webhook 30, 14 and 7 days before expiry. The free plan includes 10
            monitors. Check the certificate too with the
            <RouterLink to="/tools/ssl-checker" class="text-[var(--accent)] hover:underline"
              >SSL certificate checker</RouterLink
            >.
          </p>
        </section>
        <section id="faq">
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            Domain expiry questions
          </h2>
          <div class="space-y-6">
            <div v-for="faq in domainExpiryFaqs" :key="faq.q">
              <h3 class="mb-2 text-[17px] font-semibold text-[var(--text-strong)]">{{ faq.q }}</h3>
              <p>{{ faq.a }}</p>
            </div>
          </div>
        </section>
      </div>
    </div>

    <CtaBanner
      title="Never miss a renewal."
      subtitle="10 monitors, no credit card, no expiry date on the free plan."
    />
  </LandingLayout>
</template>
