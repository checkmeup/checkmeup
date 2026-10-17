<script setup lang="ts">
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'
import PageHero from '@/components/PageHero.vue'
import CtaBanner from '@/components/CtaBanner.vue'
import ToolChecker from '@/components/ToolChecker.vue'
import { useSeo } from '@/composables/useSeo'

useSeo({
  title: 'SSL Certificate Checker — Checkmeup',
  description:
    "Check any site's SSL certificate for free: expiry date, days left and issuer. No sign-up. Then let Checkmeup alert you before it expires.",
  path: '/tools/ssl-checker',
})
</script>

<template>
  <LandingLayout>
    <PageHero title="SSL certificate checker" eyebrow="Free tool · No sign-up">
      <p>
        Enter a hostname to see when its SSL certificate expires, how many days are left, and who
        issued it.
      </p>
    </PageHero>

    <div class="mx-auto max-w-[1200px] px-4 pb-20 sm:px-6">
      <div class="max-w-2xl">
        <ToolChecker
          tool="ssl"
          endpoint="/api/v1/tools/ssl"
          param="host"
          input-label="Hostname"
          placeholder="example.com"
          detail-key="issuer"
          detail-label="Issuer"
          monitor-cta="Get an alert 30, 14 and 7 days before this certificate expires."
        />
      </div>

      <div class="mt-16 max-w-2xl space-y-10 text-[15px] leading-relaxed text-[var(--text-dim)]">
        <section>
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            What it checks
          </h2>
          <p>
            The checker connects to the host on port 443 over TLS, the way a browser does, and reads
            the certificate the server sends. You see its expiry date, the whole days left until
            then, and the issuer's name. If the connection or the handshake fails, you see why.
          </p>
        </section>
        <section>
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            Why certificates still expire
          </h2>
          <p>
            Auto-renewal fails quietly: a DNS change breaks the ACME challenge, a renewal job stops
            running, or a certificate was bought by hand a year ago. Nobody notices until visitors
            see a browser warning. Read
            <RouterLink to="/blog/ssl-certificate-expired-story" class="text-[var(--accent)] hover:underline"
              >the 3 AM SSL cert that cost us a client</RouterLink
            >.
          </p>
        </section>
        <section>
          <h2 class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            Check it every day instead
          </h2>
          <p>
            A Checkmeup SSL monitor runs this check daily and alerts you by email, Telegram, Slack or
            webhook 30, 14 and 7 days before expiry. The free plan includes 10 monitors. Need
            the domain itself too? Try the
            <RouterLink to="/tools/domain-expiry-checker" class="text-[var(--accent)] hover:underline"
              >domain expiry checker</RouterLink
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
