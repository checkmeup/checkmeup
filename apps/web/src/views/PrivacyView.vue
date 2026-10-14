<script setup lang="ts">
import LandingLayout from '@/layouts/LandingLayout.vue'
import LabeledSection from '@/components/LabeledSection.vue'
import PageHero from '@/components/PageHero.vue'
import { useSeo } from '@/composables/useSeo'

useSeo({
  title: 'Privacy Policy — Checkmeup',
  description: "Checkmeup's privacy policy and subprocessor list.",
  path: '/privacy',
})

const effectiveDate = '2026-06-17'

const subprocessors = [
  { name: 'Resend', use: 'Transactional email — password reset, account notices' },
  { name: 'Telegram', use: 'Delivering down/recovery alerts to the chat you connect' },
  { name: 'Paddle', use: 'Payment processing and billing for paid plans (merchant of record)' },
  { name: 'Hetzner', use: 'Application hosting and database storage (Germany)' },
]
</script>

<template>
  <LandingLayout>
    <PageHero title="Privacy Policy">Effective {{ effectiveDate }}</PageHero>

    <div
      class="mx-auto max-w-[1200px] px-4 pb-24 text-[15px] leading-relaxed text-[var(--text-dim)] sm:px-6"
    >
      <LabeledSection title="1. Scope and controller">
        <div class="max-w-[680px] space-y-3">
          <p>
            This policy covers checkmeup.net and the public status pages it serves at
            <code
              class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
              >checkmeup.net/status/*</code
            >
            (together, "the Service"). It explains what we collect, why, and who we share it with.
            The data controller is Andrew Molyuk, a sole proprietor (עוסק פטור) based in Israel,
            operating Checkmeup ("we", "us").
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="2. Information you provide">
        <div class="max-w-[680px] space-y-3">
          <p>
            Account data: email address and password (stored as a salted hash, never in plain text).
            Billing data: handled directly by Paddle — we never see or store your card details.
            Monitor data: the URLs, cron schedules, and check results you configure. Telegram chat
            ID, if you connect alerts. Anything you write in a status page's title, description, or
            a support email.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="3. Information we collect automatically">
        <div class="max-w-[680px] space-y-3">
          <p>
            Request metadata (IP address, user agent, timestamp, path) for security and abuse
            prevention. Two strictly-necessary, httpOnly session cookies (<code
              class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
              >access_token</code
            >,
            <code
              class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
              >refresh_token</code
            >) that keep you signed in — no advertising or analytics cookies, and nothing that
            tracks you across other sites.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="4. How and why we use it">
        <div class="max-w-[680px] space-y-3">
          <p>
            To run the Service: performing checks, sending alerts, serving status pages, enforcing
            plan limits, and securing accounts against abuse. Legal bases (for visitors covered by
            GDPR or similar law): performing our contract with you, our legitimate interest in
            operating and securing the Service, and compliance with tax/billing law for payment
            records. We don't sell your data or use it for advertising.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="5. Subprocessors">
        <div class="max-w-[680px] space-y-3">
          <p>
            We share data with the following subprocessors, each limited to what it needs to do its
            job:
          </p>
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-[var(--border)]">
                <th class="text-left py-2 font-medium text-[var(--text-strong)]">Service</th>
                <th class="text-left py-2 font-medium text-[var(--text-strong)]">Purpose</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="s in subprocessors" :key="s.name" class="border-b border-[var(--border)]">
                <td class="py-2 pr-4 whitespace-nowrap text-[var(--text-dim)]">{{ s.name }}</td>
                <td class="py-2 text-[var(--text-dim)]">{{ s.use }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </LabeledSection>

      <LabeledSection title="6. International transfers">
        <div class="max-w-[680px] space-y-3">
          <p>
            Application hosting and the database are in Germany (Hetzner, within the EU). Some
            subprocessors above may process or store data outside the EU/EEA as part of their own
            infrastructure; where that happens, we rely on their standard contractual safeguards.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="7. Data retention">
        <div class="max-w-[680px] space-y-3">
          <p>
            Account and monitor data is kept for as long as your account is active. We don't yet
            have a self-service delete-account button — email us and we'll delete your account,
            monitors, check history, and status pages. Data may persist briefly afterward in routine
            backups before they roll over. We retain minimal billing records as required by tax law.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="8. Security">
        <div class="max-w-[680px] space-y-3">
          <p>
            Passwords are hashed with bcrypt, never stored in plain text. Sessions use httpOnly,
            SameSite cookies (see
            <RouterLink to="/terms" class="underline text-[var(--accent)]">Terms</RouterLink> for
            how this works). No method of transmission or storage is perfectly secure, but we take
            reasonable measures to protect your data.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="9. Your rights">
        <div class="max-w-[680px] space-y-3">
          <p>
            Depending on where you live, you may have rights to access, correct, delete, or export
            your data, and to object to certain processing, under GDPR or similar law. Email us and
            we'll handle it directly — no formal request process needed. EU residents may also lodge
            a complaint with their local supervisory authority.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="10. Children's privacy">
        <div class="max-w-[680px] space-y-3">
          <p>
            The Service is not directed at children and isn't intended for anyone under the age
            stated in our
            <RouterLink to="/terms" class="underline text-[var(--accent)]">Terms</RouterLink>'
            eligibility requirement. We don't knowingly collect data from children. If you believe a
            child has provided us data, email us and we'll delete it.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="11. Changes to this policy">
        <div class="max-w-[680px] space-y-3">
          <p>
            We may update this policy as the Service evolves. Material changes require re-acceptance
            on your next sign-in — you'll see a blocking notice before continuing to use the
            Service.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="12. Contact">
        <div class="max-w-[680px] space-y-3">
          <p>
            Questions about this policy or your data:
            <a href="mailto:andrew@checkmeup.net" class="underline text-[var(--accent)]"
              >andrew@checkmeup.net</a
            >.
          </p>
        </div>
      </LabeledSection>
    </div>
  </LandingLayout>
</template>
