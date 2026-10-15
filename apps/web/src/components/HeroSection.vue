<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import CheckStrip from '@/components/CheckStrip.vue'
import StatusLabel from '@/components/StatusLabel.vue'
import { monitorTypes, ownStatusSlug } from '@/content/home'
import { liveStrip, type StatusSummary } from '@/lib/statusSummary'

// Until checkmeup.net's own status page answers (and wherever it can't, like
// a local dev database without that page), the strip shows the example job
// from the alert above: eighty-nine good nights, then tonight's miss.
const example = Array.from({ length: 90 }, (_, i) => ({ ok: i !== 89 }))

const summary = ref<StatusSummary | null>(null)
const live = computed(() => (summary.value ? liveStrip(summary.value) : null))

onMounted(async () => {
  try {
    const res = await fetch(`/status/${ownStatusSlug}/summary.json`)
    if (res.ok) summary.value = (await res.json()) as StatusSummary
  } catch {
    // Network error: keep the example.
  }
})
</script>

<template>
  <!-- The hero shows the product's one moment of value: the line you add to a
       cron job, and the alert you get when that job stops. -->
  <section class="mx-auto max-w-[1200px] px-4 pt-16 pb-20 sm:px-6 md:pt-24">
    <div class="grid items-center gap-14 lg:grid-cols-2">
      <div class="flex min-w-0 flex-col gap-6">
        <div class="flex flex-col gap-3">
          <!-- Each monitor type links to its row in the spec table below. -->
          <p class="font-mono text-[13px] text-[var(--text-dim)]">
            <template v-for="(m, i) in monitorTypes" :key="m.id"
              ><span v-if="i > 0" aria-hidden="true"> · </span
              ><RouterLink
                :to="{ hash: `#${m.id}` }"
                class="underline decoration-[var(--border-strong)] decoration-dotted underline-offset-4 transition-colors hover:text-[var(--accent)] hover:decoration-[var(--accent)]"
                >{{ m.id }}</RouterLink
              ></template
            >
          </p>
          <h1
            class="text-[44px] leading-[1.0] font-semibold tracking-[-0.045em] text-[var(--text-strong)] sm:text-[64px]"
          >
            Know before your client does.
          </h1>
        </div>
        <p class="max-w-[480px] text-[17px] leading-relaxed text-[var(--text-dim)] sm:text-lg">
          Add one line to the end of a cron job, or paste a URL. When something stops, you hear
          about it on Telegram, Slack or email — before the person who pays you does.
        </p>
        <div class="flex flex-col gap-3 sm:flex-row">
          <RouterLink
            to="/sign-up"
            class="rounded-[10px] bg-[var(--accent)] px-6 py-3 text-center text-[15px] font-semibold text-[var(--on-accent)] transition-colors hover:bg-[var(--accent-hover)]"
          >
            Start free — no credit card
          </RouterLink>
          <RouterLink
            to="/docs"
            class="rounded-[10px] border border-[var(--border-strong)] px-6 py-3 text-center text-[15px] font-medium text-[var(--text)] transition-colors hover:bg-[var(--surface-raised)]"
          >
            Read the docs
          </RouterLink>
        </div>
      </div>

      <div
        aria-label="Example: a cron job and its missed-ping alert"
        role="img"
        class="flex min-w-0 flex-col gap-3.5"
      >
        <!-- From sm up the type scales with the card's own width (cqi), so the
             64-character crontab line always fits on one line: Geist Mono is
             0.6em a character, so 64 of them need 38.4em, and 2.5cqi keeps
             that under the card's content width with room to spare. -->
        <div
          class="@container rounded-[14px] bg-[var(--terminal-bg)] px-5 py-5 font-mono text-[12.5px] leading-[1.75] text-[var(--terminal-text)] sm:text-[length:clamp(11px,2.5cqi,13.5px)]"
        >
          <div class="mb-2.5 flex justify-between text-xs text-[var(--terminal-dim)]">
            <span>crontab -e</span>
            <span>server-01</span>
          </div>
          <div class="text-[var(--terminal-dim)]"># nightly database backup</div>
          <!-- One crontab line. Only on phones, where it can't fit, it
               soft-wraps like a terminal would, with a hanging indent rather
               than being cut off. A backslash would be wrong here — crontab
               has no line continuation. -->
          <div
            class="max-sm:pl-[2ch] max-sm:-indent-[2ch] max-sm:[overflow-wrap:anywhere] sm:whitespace-nowrap"
          >
            0 3 * * * /opt/backup.sh
            <span class="text-[var(--terminal-accent)]"
              >&amp;&amp; curl -fsS checkmeup.net/ping/k7f2c9</span
            >
          </div>
        </div>

        <div class="flex items-center gap-2.5 pl-5 font-mono text-xs text-[var(--text-muted)]">
          <svg
            aria-hidden="true"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <line x1="12" y1="4" x2="12" y2="20" />
            <polyline points="6 14 12 20 18 14" />
          </svg>
          03:00 + 5 min grace, no ping
        </div>

        <div
          class="flex items-start gap-3.5 rounded-[14px] border border-[var(--border)] bg-[var(--bg)] px-5 py-[18px] shadow-[0_18px_40px_rgb(0_0_0/7%)]"
        >
          <img src="/favicon.svg" alt="" class="h-9 w-9 flex-shrink-0" />
          <div class="flex min-w-0 flex-1 flex-col gap-1">
            <div class="flex items-center justify-between gap-3 text-[13px] text-[var(--text-muted)]">
              <span class="font-semibold text-[var(--text-strong)]">Checkmeup</span>
              <span class="flex items-center gap-2">
                <span
                  class="rounded-full border border-[var(--border-strong)] px-2 py-px text-xs text-[var(--text-dim)]"
                  >via Telegram</span
                >
                <span class="font-mono">03:05</span>
              </span>
            </div>
            <div
              class="flex items-center gap-2 text-[15px] font-semibold text-[var(--text-strong)]"
            >
              <span class="h-2 w-2 flex-shrink-0 rounded-full bg-[var(--status-down)]"></span>
              nightly-backup missed its ping
            </div>
            <!-- The same facts as the real Telegram alert further down. -->
            <div class="text-sm text-[var(--text-dim)]">
              Expected at 03:05 UTC · missed by 18s
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- checkmeup.net's own status, live; the example job until it loads.
         Phones show the last 30 days: 90 ticks in ~360px are hairlines. -->
    <div class="mt-16 flex flex-col gap-3 border-t border-[var(--border)] pt-6">
      <div class="flex flex-wrap items-center justify-between gap-3 text-sm">
        <span v-if="live && summary">
          <a
            :href="`/status/${ownStatusSlug}`"
            class="font-semibold text-[var(--text-strong)] hover:underline"
            >{{ summary.title }}</a
          >
          <span class="text-[var(--text-dim)]">
            · monitored by Checkmeup · live, last
            <span class="sm:hidden">30</span><span class="max-sm:hidden">90</span> days</span
          >
        </span>
        <span v-else>
          <span class="font-semibold text-[var(--text-strong)]">nightly-backup</span>
          <span class="text-[var(--text-dim)]">
            · Cron · example, last <span class="sm:hidden">30</span
            ><span class="max-sm:hidden">90</span> days</span
          >
        </span>
        <StatusLabel v-if="live" :tone="live.tone" :label="live.label" />
        <StatusLabel v-else tone="down" label="Down" />
      </div>
      <a
        v-if="live"
        :href="`/status/${ownStatusSlug}`"
        :aria-label="`${summary?.title} status page`"
        class="block rounded-[4px] transition-opacity hover:opacity-85"
      >
        <CheckStrip :checks="live.checks" :narrow-last="30" />
      </a>
      <CheckStrip v-else :checks="example" :narrow-last="30" />
      <div class="flex justify-between text-xs text-[var(--text-muted)]">
        <span><span class="sm:hidden">30</span><span class="max-sm:hidden">90</span> days ago</span>
        <span>Today</span>
      </div>
    </div>
  </section>
</template>
