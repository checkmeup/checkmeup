<script setup lang="ts">
import LandingLayout from '@/layouts/LandingLayout.vue'
import DocsGettingStarted from '@/components/docs/DocsGettingStarted.vue'
import DocsCron from '@/components/docs/DocsCron.vue'
import DocsUptime from '@/components/docs/DocsUptime.vue'
import DocsSsl from '@/components/docs/DocsSsl.vue'
import DocsDomain from '@/components/docs/DocsDomain.vue'
import DocsPort from '@/components/docs/DocsPort.vue'
import DocsDns from '@/components/docs/DocsDns.vue'
import DocsTelegram from '@/components/docs/DocsTelegram.vue'
import DocsEmail from '@/components/docs/DocsEmail.vue'
import DocsWebhook from '@/components/docs/DocsWebhook.vue'
import DocsSlack from '@/components/docs/DocsSlack.vue'
import DocsSms from '@/components/docs/DocsSms.vue'
import DocsStatusPages from '@/components/docs/DocsStatusPages.vue'
import DocsIncidents from '@/components/docs/DocsIncidents.vue'
import DocsMaintenance from '@/components/docs/DocsMaintenance.vue'
import DocsAppearance from '@/components/docs/DocsAppearance.vue'
import DocsPlans from '@/components/docs/DocsPlans.vue'
import DocsApi from '@/components/docs/DocsApi.vue'
import DocsHelp from '@/components/docs/DocsHelp.vue'
import { useSeo } from '@/composables/useSeo'

useSeo({
  title: 'Docs — Checkmeup',
  description:
    'How to set up cron, uptime, SSL, domain, port, and DNS monitors, alerts, status pages, and the public API on Checkmeup.',
  path: '/docs',
})

const sections = [
  { id: 'getting-started', label: 'Getting started' },
  { id: 'cron', label: 'Cron job monitoring' },
  { id: 'uptime', label: 'Uptime monitoring' },
  { id: 'ssl', label: 'SSL expiry monitoring' },
  { id: 'domain', label: 'Domain expiry monitoring' },
  { id: 'port', label: 'Port (TCP) monitoring' },
  { id: 'dns', label: 'DNS record monitoring' },
  { id: 'telegram', label: 'Telegram alerts' },
  { id: 'email', label: 'Email alerts' },
  { id: 'webhook', label: 'Webhook alerts' },
  { id: 'slack', label: 'Slack alerts' },
  { id: 'sms', label: 'SMS alerts' },
  { id: 'status-pages', label: 'Status pages' },
  { id: 'incidents', label: 'Incident management' },
  { id: 'maintenance', label: 'Maintenance windows' },
  { id: 'appearance', label: 'Appearance' },
  { id: 'plans', label: 'Plans & limits' },
  { id: 'api', label: 'Public API' },
  { id: 'help', label: 'Need help?' },
]
</script>

<template>
  <LandingLayout>
    <!-- Hero -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 pt-16 pb-10 sm:pt-24 sm:pb-12">
      <h1
        class="text-4xl sm:text-5xl font-bold tracking-tight mb-4"
        style="color: var(--text-strong)"
      >
        Documentation
      </h1>
      <p class="text-lg max-w-2xl" style="color: var(--text-dim)">
        Everything Checkmeup does today, in plain language — how to set up a monitor, what triggers
        an alert, and what each plan actually includes.
      </p>
    </section>

    <div class="max-w-6xl mx-auto px-4 sm:px-6 pb-24 flex flex-col lg:flex-row gap-12">
      <!-- Sidebar nav -->
      <nav class="lg:w-56 flex-shrink-0">
        <div
          class="lg:sticky lg:top-24 flex lg:flex-col gap-1 overflow-x-auto pb-2 lg:pb-0 -mx-1 px-1"
        >
          <a
            v-for="s in sections"
            :key="s.id"
            :href="`#${s.id}`"
            class="text-sm px-3 py-1.5 rounded-md whitespace-nowrap transition-colors hover:underline"
            style="color: var(--text-dim)"
          >
            {{ s.label }}
          </a>
        </div>
      </nav>

      <!-- Content -->
      <div class="flex-1 min-w-0 space-y-16">
        <DocsGettingStarted />

        <DocsCron />

        <DocsUptime />

        <DocsSsl />

        <DocsDomain />

        <DocsPort />

        <DocsDns />

        <DocsTelegram />

        <DocsEmail />

        <DocsWebhook />

        <DocsSlack />

        <DocsSms />

        <DocsStatusPages />

        <DocsIncidents />

        <DocsMaintenance />

        <DocsAppearance />

        <DocsPlans />

        <DocsApi />

        <DocsHelp />
      </div>
    </div>
  </LandingLayout>
</template>
