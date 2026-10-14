<script setup lang="ts">
import LandingLayout from '@/layouts/LandingLayout.vue'
import LabeledSection from '@/components/LabeledSection.vue'
import PageHero from '@/components/PageHero.vue'
import { useSeo } from '@/composables/useSeo'

useSeo({
  title: 'Terms of Service — Checkmeup',
  description: "Checkmeup's terms of service.",
  path: '/terms',
})

const effectiveDate = '2026-06-17'
</script>

<template>
  <LandingLayout>
    <PageHero title="Terms of Service">Effective {{ effectiveDate }}</PageHero>

    <div
      class="mx-auto max-w-[1200px] px-4 pb-24 text-[15px] leading-relaxed text-[var(--text-dim)] sm:px-6"
    >
      <LabeledSection title="1. The service">
        <div class="max-w-[680px] space-y-3">
          <p>
            Checkmeup is operated by Andrew Molyuk, a sole proprietor (עוסק פטור) based in Israel.
            Checkmeup ("we", "us") provides cron, uptime, SSL certificate, domain, port, and DNS
            record monitoring with alerting and public status pages ("the Service"). By creating an
            account you agree to these Terms and to our
            <RouterLink to="/privacy" class="underline text-[var(--accent)]"
              >Privacy Policy</RouterLink
            >.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="2. Eligibility">
        <div class="max-w-[680px] space-y-3">
          <p>
            You must be at least 18 years old, or the age of legal majority in your jurisdiction,
            and have the legal capacity to enter into these Terms. If you're using the Service on
            behalf of an organization, you represent that you have the authority to bind that
            organization to these Terms.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="3. Accounts">
        <div class="max-w-[680px] space-y-3">
          <p>
            You're responsible for the accuracy of the information you provide and for keeping your
            credentials secure. One account is created per organization at sign-up; you're
            responsible for activity under your account.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="4. Acceptable use">
        <div class="max-w-[680px] space-y-3">
          <p>
            Don't use the Service to monitor URLs or systems you don't have the right to monitor, to
            send excessive request volume intended to disrupt a third party, or to violate any
            applicable law. We may suspend or terminate accounts that abuse the Service.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="5. Fees and billing">
        <div class="max-w-[680px] space-y-3">
          <p>
            The Hobby plan is free. Paid plans renew automatically and are billed in advance by
            Paddle, our merchant of record, who handles payment processing and applicable tax. You
            can change or cancel your plan at any time from Billing settings; a cancelled
            subscription stays active until the end of the current billing period, then reverts to
            Hobby. If we change a paid plan's price, we'll give you at least 30 days' notice before
            it takes effect on your next billing cycle. Refund requests made within 30 days of a
            charge are honored in full, no questions asked — see our
            <RouterLink to="/refund" class="underline text-[var(--accent)]"
              >Refund Policy</RouterLink
            >.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="6. Intellectual property">
        <div class="max-w-[680px] space-y-3">
          <p>
            The Service, its branding, and its underlying software are our property or licensed to
            us. These Terms don't grant you any rights to our trademarks or branding. You retain
            ownership of the content you configure (monitor names, status page descriptions, and
            similar) and grant us the license needed to display it back to you and to anyone who
            visits your public status pages.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="7. Third-party links and services">
        <div class="max-w-[680px] space-y-3">
          <p>
            The Service integrates with third parties you choose to connect (e.g. Telegram) and
            relies on subprocessors to operate (see our
            <RouterLink to="/privacy" class="underline text-[var(--accent)]"
              >Privacy Policy</RouterLink
            >). We aren't responsible for the content, availability, or practices of third-party
            sites or services, including the URLs and systems you choose to monitor.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="8. Availability and disclaimer">
        <div class="max-w-[680px] space-y-3">
          <p>
            We aim for high availability but the Service is provided "as is" and "as available",
            without warranty that it will be uninterrupted, error-free, or fully secure, and without
            an uptime guarantee or SLA on the free or paid plans as of this version. We're not
            liable for missed alerts, delayed checks, or downtime of services you monitor.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="9. Limitation of liability">
        <div class="max-w-[680px] space-y-3">
          <p>
            To the maximum extent permitted by law, we won't be liable for any indirect, incidental,
            special, consequential, or punitive damages arising from your use of the Service,
            including lost profits or lost data, even if we've been advised of the possibility of
            such damages.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="10. Indemnification">
        <div class="max-w-[680px] space-y-3">
          <p>
            You agree to indemnify and hold us harmless from any claim, demand, loss, or damage
            (including reasonable legal fees) arising out of your use of the Service, your violation
            of these Terms, or your violation of any right of a third party — including monitoring a
            URL or system you don't have the right to monitor.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="11. Termination">
        <div class="max-w-[680px] space-y-3">
          <p>
            You can delete your account at any time. We may suspend or terminate an account
            immediately, without prior notice, for violation of these Terms or to protect the
            Service or other users. On termination, your data is deleted per the retention terms in
            our
            <RouterLink to="/privacy" class="underline text-[var(--accent)]"
              >Privacy Policy</RouterLink
            >.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="12. Governing law and disputes">
        <div class="max-w-[680px] space-y-3">
          <p>
            These Terms are governed by the laws of the State of Israel, without regard to
            conflict-of-law principles. Any dispute arising out of or relating to these Terms or the
            Service will be brought exclusively in the competent courts of Israel, and you consent
            to their exclusive jurisdiction and venue. Disputes are resolved through those courts,
            not arbitration.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="13. Miscellaneous">
        <div class="max-w-[680px] space-y-3">
          <p>
            These Terms, together with our Privacy Policy and Refund Policy, are the entire
            agreement between you and us regarding the Service, and supersede any prior agreements
            on the subject. If any provision of these Terms is found unenforceable, the remaining
            provisions stay in full effect. Our failure to enforce a provision isn't a waiver of our
            right to do so later. You may not assign these Terms without our consent; we may assign
            them without your consent in connection with a merger, acquisition, or sale of assets.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="14. Changes to these Terms">
        <div class="max-w-[680px] space-y-3">
          <p>
            We may update these Terms as the Service evolves. Material changes require re-acceptance
            on your next sign-in — you'll see a blocking notice before continuing to use the
            Service, so you're never silently bound to a new version.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="15. Contact">
        <div class="max-w-[680px] space-y-3">
          <p>
            Questions about these Terms:
            <a href="mailto:andrew@checkmeup.net" class="underline text-[var(--accent)]"
              >andrew@checkmeup.net</a
            >.
          </p>
        </div>
      </LabeledSection>
    </div>
  </LandingLayout>
</template>
