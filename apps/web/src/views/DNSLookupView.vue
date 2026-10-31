<script setup lang="ts">
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'
import PageHero from '@/components/PageHero.vue'
import CtaBanner from '@/components/CtaBanner.vue'
import DNSLookupChecker from '@/components/DNSLookupChecker.vue'
import { useSeo } from '@/composables/useSeo'

useSeo({
  title: 'DNS Lookup — Checkmeup',
  description:
    "Look up any domain's DNS records for free: A, AAAA, CNAME, MX, TXT and NS in one go. No sign-up. Then get alerted when a record changes.",
  path: '/tools/dns-lookup',
})
</script>

<template>
  <LandingLayout>
    <PageHero title="DNS lookup" eyebrow="Free tool · No sign-up">
      <p>
        Enter a hostname to see its A, AAAA, CNAME, MX, TXT and NS records at once.
      </p>
    </PageHero>

    <div class="mx-auto max-w-[1200px] px-4 pb-20 sm:px-6">
      <div class="max-w-2xl">
        <DNSLookupChecker />
      </div>

      <div class="mt-16 max-w-2xl space-y-10 text-[15px] leading-relaxed text-[var(--text-dim)]">
        <section>
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            What it checks
          </h2>
          <p>
            The lookup asks a DNS resolver for each record type, the way your visitors' browsers
            and mail servers do. You see IPv4 and IPv6 addresses, the alias a name points to, mail
            servers with their priority, TXT records such as SPF and domain verification, and the
            name servers. If a lookup fails, you see why.
          </p>
        </section>
        <section>
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            Why DNS records change without warning
          </h2>
          <p>
            A registrar transfer resets the zone, a teammate edits the wrong record, or a provider
            migration drops the MX records, and mail quietly stops arriving. Nothing breaks loudly
            until someone notices. Read the
            <RouterLink to="/blog/dns-record-monitoring-guide" class="text-[var(--accent)] hover:underline"
              >DNS record monitoring guide</RouterLink
            >.
          </p>
        </section>
        <section>
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            Check it every few minutes instead
          </h2>
          <p>
            A Checkmeup
            <RouterLink to="/dns-monitoring" class="text-[var(--accent)] hover:underline">DNS monitor</RouterLink>
            runs this lookup on a schedule and alerts you by email, Telegram, Slack or webhook when a
            record changes or stops resolving. The free plan includes 10 monitors.
          </p>
        </section>
      </div>
    </div>

    <CtaBanner
      title="Know when your DNS changes."
      subtitle="10 monitors, no credit card, no expiry date on the free plan."
    />
  </LandingLayout>
</template>
