<script setup lang="ts">
import CheckStrip from '@/components/CheckStrip.vue'
import StatusLabel from '@/components/StatusLabel.vue'
import type { Tone } from '@/lib/monitorRows'

// The product preview beside the headline is built from the app's own
// components with example data, so it always looks like the real dashboard.
function ticks(pattern: string) {
  return pattern.split('').map((ch) => ({ ok: ch === '-' ? null : ch === 'u', slow: ch === 's' }))
}

const previewRows: {
  name: string
  type: string
  tone: Tone
  label: string
  metric: string
  pattern: string
}[] = [
  {
    name: 'shop.example.com',
    type: 'Uptime',
    tone: 'down',
    label: 'Down',
    metric: 'HTTP 502',
    pattern: 'uuuuuuuuuuuuuuuuuuuuduuuuuuuuuuuuuudddddd',
  },
  {
    name: 'nightly-db-backup',
    type: 'Cron',
    tone: 'up',
    label: 'Up',
    metric: 'ran 03:00',
    pattern: 'uuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuu',
  },
  {
    name: 'api.example.com',
    type: 'Uptime',
    tone: 'up',
    label: 'Up',
    metric: '142 ms',
    pattern: 'uuuuuuuuuuuuuusuuuuuuuuuuuuuuuuuuuuuuuuu',
  },
  {
    name: 'example.com',
    type: 'SSL',
    tone: 'warn',
    label: 'Expiring soon',
    metric: '12 days',
    pattern: 'uuuuuuuuuuuuuuuuuuuussssssssssssssssssss',
  },
]
</script>

<template>
  <section class="mx-auto max-w-[1200px] px-4 pt-16 pb-20 sm:px-6 md:pt-24">
    <div class="grid items-center gap-14 lg:grid-cols-[1fr_1.05fr]">
      <div class="flex flex-col gap-6">
        <h1
          class="text-[44px] leading-[1.02] font-semibold tracking-[-0.045em] text-[var(--text-strong)] sm:text-[60px]"
        >
          Know before<br />
          <span class="text-[var(--text-muted)]">your client does.</span>
        </h1>

        <p class="max-w-[500px] text-[17px] leading-relaxed text-[var(--text-dim)]">
          Cron job monitoring, uptime checks, SSL and domain expiry alerts, and raw TCP port
          checks — for every client site you maintain. Get alerted on Telegram, Slack, or email
          before your client ever notices something's wrong.
        </p>

        <div class="flex flex-col gap-3 sm:flex-row">
          <RouterLink
            to="/sign-up"
            class="rounded-[10px] bg-[var(--accent)] px-6 py-3 text-center text-[15px] font-semibold text-[var(--on-accent)] transition-colors hover:bg-[var(--accent-hover)]"
          >
            Start free — no credit card
          </RouterLink>
          <RouterLink
            to="/sign-in"
            class="rounded-[10px] border border-[var(--border-strong)] px-6 py-3 text-center text-[15px] font-medium text-[var(--text)] transition-colors hover:bg-[var(--surface-raised)]"
          >
            Sign in →
          </RouterLink>
        </div>
        <p class="text-[13px] text-[var(--text-dim)]">
          Built for freelancers and solo devs with client sites.
        </p>
      </div>

      <!-- Product preview -->
      <div
        aria-label="Example dashboard"
        role="img"
        class="flex flex-col gap-5 rounded-[18px] border border-[var(--border)] bg-[var(--bg)] p-5 shadow-[0_24px_60px_rgb(0_0_0/8%)] sm:p-7"
      >
        <div class="flex flex-col gap-1">
          <span class="text-[22px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]">
            1 monitor needs you.
          </span>
          <span class="text-sm text-[var(--text-dim)]">
            99.62% average uptime today · 3 of 4 up
          </span>
        </div>

        <div class="flex items-center gap-4 border-y border-[var(--border)] py-4">
          <span
            class="h-2.5 w-2.5 flex-shrink-0 rounded-full bg-[var(--status-down)] shadow-[0_0_0_4px_var(--status-down-wash)]"
          ></span>
          <div class="flex min-w-0 flex-col">
            <span class="font-semibold text-[var(--text-strong)]">shop.example.com</span>
            <span class="text-[13.5px] text-[var(--status-down)]">
              Down — the last check failed: 502 Bad Gateway.
            </span>
          </div>
        </div>

        <ul class="flex flex-col">
          <li
            v-for="row in previewRows"
            :key="row.name"
            class="grid grid-cols-[110px_1fr_auto] items-center gap-4 border-t border-[var(--border)] py-2.5 first:border-t-0 sm:grid-cols-[120px_1fr_110px_80px]"
          >
            <StatusLabel :tone="row.tone" :label="row.label" />
            <span class="min-w-0 truncate text-sm">
              <span class="font-medium text-[var(--text-strong)]">{{ row.name }}</span>
              <span class="text-[var(--text-dim)]"> · {{ row.type }}</span>
            </span>
            <CheckStrip size="sm" :checks="ticks(row.pattern)" class="hidden sm:flex" />
            <span class="text-right font-mono text-[13px] text-[var(--text)]">{{ row.metric }}</span>
          </li>
        </ul>
      </div>
    </div>
  </section>
</template>
