<script setup lang="ts">
import LandingLayout from '@/layouts/LandingLayout.vue'
import LabeledSection from '@/components/LabeledSection.vue'
import PageHero from '@/components/PageHero.vue'
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
    <PageHero title="Refund Policy">Effective {{ effectiveDate }}</PageHero>

    <div
      class="mx-auto max-w-[1200px] px-4 pb-24 text-[15px] leading-relaxed text-[var(--text-dim)] sm:px-6"
    >
      <LabeledSection title="1. Scope">
        <div class="max-w-[680px] space-y-3">
          <p>
            This policy covers charges for Checkmeup's paid plans (Solo, Startup, Enterprise),
            billed and processed by Paddle, our merchant of record. The Hobby plan is free and has
            no charges to refund.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="2. The policy">
        <div class="max-w-[680px] space-y-3">
          <p>
            Refund requests made within 30 days of any charge — your first subscription payment, a
            renewal, or an upgrade — are honored in full, no questions asked. This applies to both
            monthly and annual billing.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="3. How to request one">
        <div class="max-w-[680px] space-y-3">
          <p>
            Email
            <a href="mailto:andrew@checkmeup.net" class="underline text-[var(--accent)]"
              >andrew@checkmeup.net</a
            >
            with the email address on your account. Refunds are issued back to your original payment
            method via Paddle, typically within a few business days.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="4. Related terms">
        <div class="max-w-[680px] space-y-3">
          <p>
            This policy is part of, and should be read alongside, our
            <RouterLink to="/terms" class="underline text-[var(--accent)]"
              >Terms of Service</RouterLink
            >, which covers plan changes, cancellations, and billing more broadly.
          </p>
        </div>
      </LabeledSection>
    </div>
  </LandingLayout>
</template>
