<script setup lang="ts">
import LandingLayout from '@/layouts/LandingLayout.vue'
import { useSeo } from '@/composables/useSeo'

useSeo({
  title: 'Refund Policy — Checkmeup',
  description: "Checkmeup's refund policy for paid plans.",
  path: '/refund',
})

const effectiveDate = '2026-07-01'
</script>

<template>
  <LandingLayout>
    <section class="max-w-3xl mx-auto px-4 sm:px-6 pt-16 pb-24 sm:pt-24">
      <h1 class="text-3xl sm:text-4xl font-bold tracking-tight mb-2" style="color: var(--text-strong)">
        Refund Policy
      </h1>
      <p class="text-sm mb-12" style="color: var(--text-muted)">Effective {{ effectiveDate }}</p>

      <div class="space-y-10 text-sm leading-relaxed" style="color: var(--text-dim)">
        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">1. Scope</h2>
          <p>
            This policy covers charges for Checkmeup's paid plans (Solo, Startup, Enterprise), billed and
            processed by Paddle, our merchant of record. The Hobby plan is free and has no charges to
            refund.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">2. The policy</h2>
          <p>
            Refund requests made within 30 days of any charge — your first subscription payment, a renewal,
            or an upgrade — are honored in full, no questions asked. This applies to both monthly and annual
            billing.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">3. How to request one</h2>
          <p>
            Email
            <a href="mailto:andrew@checkmeup.net" class="underline" style="color: var(--color-green-500)"
              >andrew@checkmeup.net</a
            >
            with the email address on your account. Refunds are issued back to your original payment method
            via Paddle, typically within a few business days.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">4. Related terms</h2>
          <p>
            This policy is part of, and should be read alongside, our
            <RouterLink to="/terms" class="underline" style="color: var(--color-green-500)">Terms of Service</RouterLink>,
            which covers plan changes, cancellations, and billing more broadly.
          </p>
        </section>
      </div>
    </section>
  </LandingLayout>
</template>
