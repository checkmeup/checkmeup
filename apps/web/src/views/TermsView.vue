<script setup lang="ts">
import LandingLayout from '@/layouts/LandingLayout.vue'

const effectiveDate = '2026-06-17'
</script>

<template>
  <LandingLayout>
    <section class="max-w-3xl mx-auto px-4 sm:px-6 pt-16 pb-24 sm:pt-24">
      <h1 class="text-3xl sm:text-4xl font-bold tracking-tight mb-2" style="color: var(--text-strong)">
        Terms of Service
      </h1>
      <p class="text-sm mb-12" style="color: var(--text-muted)">Effective {{ effectiveDate }}</p>

      <div class="space-y-10 text-sm leading-relaxed" style="color: var(--text-dim)">
        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">1. The service</h2>
          <p>
            checkmeup ("we", "us") provides cron, uptime, and SSL certificate monitoring with alerting and
            public status pages ("the Service"). By creating an account you agree to these Terms and to our
            <RouterLink to="/privacy" class="underline" style="color: var(--color-green-500)">Privacy Policy</RouterLink>.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">2. Accounts</h2>
          <p>
            You're responsible for the accuracy of the information you provide and for keeping your
            credentials secure. One account is created per organization at sign-up; you're responsible for
            activity under your account.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">3. Acceptable use</h2>
          <p>
            Don't use the Service to monitor URLs or systems you don't have the right to monitor, to send
            excessive request volume intended to disrupt a third party, or to violate any applicable law. We
            may suspend or terminate accounts that abuse the Service.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">4. Fees and billing</h2>
          <p>
            The Hobby plan is free. Paid plans renew automatically and are billed in advance by
            LemonSqueezy, our merchant of record, who handles payment processing and applicable tax. You can
            change or cancel your plan at any time from Billing settings; a cancelled subscription stays
            active until the end of the current billing period, then reverts to Hobby. If we change a paid
            plan's price, we'll give you at least 30 days' notice before it takes effect on your next billing
            cycle. Refund requests made within 30 days of a charge are honored in full, no questions asked.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">5. Intellectual property</h2>
          <p>
            The Service, its branding, and its underlying software are our property or licensed to us. These
            Terms don't grant you any rights to our trademarks or branding. You retain ownership of the
            content you configure (monitor names, status page descriptions, and similar) and grant us the
            license needed to display it back to you and to anyone who visits your public status pages.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">6. Third-party links and services</h2>
          <p>
            The Service integrates with third parties you choose to connect (e.g. Telegram) and relies on
            subprocessors to operate (see our <RouterLink to="/privacy" class="underline" style="color: var(--color-green-500)">Privacy Policy</RouterLink>). We aren't responsible for the content, availability, or practices of
            third-party sites or services, including the URLs and systems you choose to monitor.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">7. Availability and disclaimer</h2>
          <p>
            We aim for high availability but the Service is provided "as is" and "as available", without
            warranty that it will be uninterrupted, error-free, or fully secure, and without an uptime
            guarantee or SLA on the free or paid plans as of this version. We're not liable for missed
            alerts, delayed checks, or downtime of services you monitor.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">8. Limitation of liability</h2>
          <p>
            To the maximum extent permitted by law, we won't be liable for any indirect, incidental,
            special, consequential, or punitive damages arising from your use of the Service, including lost
            profits or lost data, even if we've been advised of the possibility of such damages.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">9. Termination</h2>
          <p>
            You can delete your account at any time. We may suspend or terminate an account immediately, without
            prior notice, for violation of these Terms or to protect the Service or other users. On
            termination, your data is deleted per the retention terms in our
            <RouterLink to="/privacy" class="underline" style="color: var(--color-green-500)">Privacy Policy</RouterLink>.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">10. Governing law</h2>
          <p>
            These Terms are governed by the laws of the State of Israel, without regard to conflict-of-law
            principles.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">11. Changes to these Terms</h2>
          <p>
            We may update these Terms as the Service evolves. Material changes require re-acceptance on your
            next sign-in — you'll see a blocking notice before continuing to use the Service, so you're never
            silently bound to a new version.
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold mb-2" style="color: var(--text-strong)">12. Contact</h2>
          <p>
            Questions about these Terms:
            <a href="mailto:andrew@checkmeup.net" class="underline" style="color: var(--color-green-500)"
              >andrew@checkmeup.net</a
            >.
          </p>
        </section>
      </div>
    </section>
  </LandingLayout>
</template>
