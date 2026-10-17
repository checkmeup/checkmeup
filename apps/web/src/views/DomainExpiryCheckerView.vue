<script setup lang="ts">
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'
import PageHero from '@/components/PageHero.vue'
import CtaBanner from '@/components/CtaBanner.vue'
import ToolChecker from '@/components/ToolChecker.vue'
import { useSeo } from '@/composables/useSeo'

useSeo({
  title: 'Domain Expiry Checker — Checkmeup',
  description:
    "Check when any domain expires, for free: expiry date, days left and registrar, over RDAP or WHOIS. No sign-up. Get alerted before it lapses.",
  path: '/tools/domain-expiry-checker',
})
</script>

<template>
  <LandingLayout>
    <PageHero title="Domain expiry checker" eyebrow="Free tool · No sign-up">
      <p>
        Enter a domain to see when its registration expires, how many days are left, and which
        registrar holds it.
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
      </div>
    </div>

    <CtaBanner
      title="Never miss a renewal."
      subtitle="10 monitors, no credit card, no expiry date on the free plan."
    />
  </LandingLayout>
</template>
