<script setup lang="ts">
import { onMounted, ref } from 'vue'

// Checkmeup's own platform-wide totals (GET /api/v1/stats), in place of
// testimonials: what the service actually did in the last 30 days. Nothing
// renders until the numbers arrive, so a failed request (or the prerender,
// which has no backend) shows no section rather than made-up figures.
interface PlatformStats {
  monitors: number
  pings30d: number
  checks30d: number
  incidents30d: number
}

const stats = ref<PlatformStats | null>(null)

onMounted(async () => {
  try {
    const res = await fetch('/api/v1/stats')
    if (res.ok) stats.value = (await res.json()) as PlatformStats
  } catch {
    // Network error: leave the section out.
  }
})

function format(n: number) {
  return n.toLocaleString('en-US')
}
</script>

<template>
  <section v-if="stats" class="border-t border-[var(--border)]">
    <div class="mx-auto max-w-[1200px] px-4 py-16 sm:px-6">
      <p class="mb-8 font-mono text-[13px] text-[var(--text-dim)]">
        checkmeup.net · last 30 days · live
      </p>
      <dl class="grid grid-cols-2 gap-x-6 gap-y-8 lg:grid-cols-4">
        <div class="flex flex-col gap-1.5 border-t border-[var(--border)] pt-4">
          <dt class="text-[13px] text-[var(--text-dim)]">Monitors running now</dt>
          <dd class="text-[30px] font-semibold tracking-[-0.035em] text-[var(--text-strong)]">
            {{ format(stats.monitors) }}
          </dd>
        </div>
        <div class="flex flex-col gap-1.5 border-t border-[var(--border)] pt-4">
          <dt class="text-[13px] text-[var(--text-dim)]">Cron pings received</dt>
          <dd class="text-[30px] font-semibold tracking-[-0.035em] text-[var(--text-strong)]">
            {{ format(stats.pings30d) }}
          </dd>
        </div>
        <div class="flex flex-col gap-1.5 border-t border-[var(--border)] pt-4">
          <dt class="text-[13px] text-[var(--text-dim)]">Uptime, port &amp; DNS checks run</dt>
          <dd class="text-[30px] font-semibold tracking-[-0.035em] text-[var(--text-strong)]">
            {{ format(stats.checks30d) }}
          </dd>
        </div>
        <div class="flex flex-col gap-1.5 border-t border-[var(--border)] pt-4">
          <dt class="text-[13px] text-[var(--text-dim)]">Outages caught</dt>
          <dd class="text-[30px] font-semibold tracking-[-0.035em] text-[var(--text-strong)]">
            {{ format(stats.incidents30d) }}
          </dd>
        </div>
      </dl>
    </div>
  </section>
</template>
