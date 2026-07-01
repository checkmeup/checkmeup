<script setup lang="ts">
import LandingLayout from '@/layouts/LandingLayout.vue'

const effectiveDate = '2026-06-17'

const subprocessors = [
  { name: 'Resend', use: 'Transactional email — password reset, account notices' },
  { name: 'Telegram', use: 'Delivering down/recovery alerts to the chat you connect' },
  { name: 'LemonSqueezy', use: 'Payment processing and billing for paid plans (merchant of record)' },
  { name: 'Hetzner', use: 'Application hosting and database storage (Germany)' },
]
</script>

<template>
  <LandingLayout>
    <section class="max-w-3xl mx-auto px-4 sm:px-6 pt-16 pb-24 sm:pt-24">
      <h1 class="text-3xl sm:text-4xl font-bold tracking-tight mb-2" style="color: var(--text-strong)">
        Privacy Policy
      </h1>
      <p class="text-sm mb-12" style="color: var(--text-muted)">Effective {{ effectiveDate }}</p>

      <div class="space-y-10 text-sm leading-relaxed" style="color: var(--text-dim)">
        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">1. Scope and controller</h2>
          <p>
            This policy covers checkmeup.net and the public status pages it serves at
            <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)">checkmeup.net/status/*</code>
            (together, "the Service"). It explains what we collect, why, and who we share it with. The data
            controller is Andrew Molyuk, a sole proprietor (עוסק פטור) based in Israel, operating checkmeup
            ("we", "us").
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">2. Information you provide</h2>
          <p>
            Account data: email address and password (stored as a salted hash, never in plain text). Billing
            data: handled directly by LemonSqueezy — we never see or store your card details. Monitor data:
            the URLs, cron schedules, and check results you configure. Telegram chat ID, if you connect
            alerts. Anything you write in a status page's title, description, or a support email.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">3. Information we collect automatically</h2>
          <p>
            Request metadata (IP address, user agent, timestamp, path) for security and abuse prevention.
            Two strictly-necessary, httpOnly session cookies (<code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)">access_token</code>,
            <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)">refresh_token</code>) that keep you signed in — no advertising or
            analytics cookies, and nothing that tracks you across other sites.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">4. How and why we use it</h2>
          <p>
            To run the Service: performing checks, sending alerts, serving status pages, enforcing plan
            limits, and securing accounts against abuse. Legal bases (for visitors covered by GDPR or
            similar law): performing our contract with you, our legitimate interest in operating and
            securing the Service, and compliance with tax/billing law for payment records. We don't sell
            your data or use it for advertising.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">5. Subprocessors</h2>
          <p class="mb-3">
            We share data with the following subprocessors, each limited to what it needs to do its job:
          </p>
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b" style="border-color: var(--border)">
                <th class="text-left py-2 font-medium" style="color: var(--text-strong)">Service</th>
                <th class="text-left py-2 font-medium" style="color: var(--text-strong)">Purpose</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="s in subprocessors" :key="s.name" class="border-b" style="border-color: var(--border)">
                <td class="py-2 pr-4 whitespace-nowrap" style="color: var(--text-dim)">{{ s.name }}</td>
                <td class="py-2" style="color: var(--text-dim)">{{ s.use }}</td>
              </tr>
            </tbody>
          </table>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">6. International transfers</h2>
          <p>
            Application hosting and the database are in Germany (Hetzner, within the EU). Some subprocessors
            above may process or store data outside the EU/EEA as part of their own infrastructure; where
            that happens, we rely on their standard contractual safeguards.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">7. Data retention</h2>
          <p>
            Account and monitor data is kept for as long as your account is active. We don't yet have a
            self-service delete-account button — email us and we'll delete your account, monitors, check
            history, and status pages. Data may persist briefly afterward in routine backups before they
            roll over. We retain minimal billing records as required by tax law.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">8. Security</h2>
          <p>
            Passwords are hashed with bcrypt, never stored in plain text. Sessions use httpOnly, SameSite
            cookies (see <RouterLink to="/terms" class="underline" style="color: var(--color-green-500)">Terms</RouterLink> for how this works). No method of transmission or storage is
            perfectly secure, but we take reasonable measures to protect your data.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">9. Your rights</h2>
          <p>
            Depending on where you live, you may have rights to access, correct, delete, or export your
            data, and to object to certain processing, under GDPR or similar law. Email us and we'll handle
            it directly — no formal request process needed. EU residents may also lodge a complaint with
            their local supervisory authority.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">10. Children's privacy</h2>
          <p>
            The Service is not directed at children and isn't intended for anyone under the age stated in
            our <RouterLink to="/terms" class="underline" style="color: var(--color-green-500)">Terms</RouterLink>' eligibility requirement. We don't knowingly collect data from
            children. If you believe a child has provided us data, email us and we'll delete it.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">11. Changes to this policy</h2>
          <p>
            We may update this policy as the Service evolves. Material changes require re-acceptance on
            your next sign-in — you'll see a blocking notice before continuing to use the Service.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">12. Contact</h2>
          <p>
            Questions about this policy or your data:
            <a href="mailto:andrew@checkmeup.net" class="underline" style="color: var(--color-green-500)"
              >andrew@checkmeup.net</a
            >.
          </p>
        </section>
      </div>
    </section>
  </LandingLayout>
</template>
