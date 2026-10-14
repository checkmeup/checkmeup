<script setup lang="ts">
import { computed, ref } from 'vue'
import AppLayout from '@/layouts/AppLayout.vue'
import AddMonitorMenu from '@/components/AddMonitorMenu.vue'
import AttentionPanel from '@/components/AttentionPanel.vue'
import MonitorsTable, { type MonitorFilter } from '@/components/MonitorsTable.vue'
import UpcomingExpirationsCard from '@/components/UpcomingExpirationsCard.vue'
import { useMonitorRows } from '@/composables/useMonitorRows'
import { useBilling } from '@/composables/useBilling'

const { rows, isPending } = useMonitorRows()
const { data: billingInfo } = useBilling()

const filter = ref<MonitorFilter>('all')

const total = computed(() => rows.value.length)
const problems = computed(() => rows.value.filter((r) => r.tone === 'down' || r.tone === 'warn'))
const upCount = computed(() => rows.value.filter((r) => r.tone === 'up').length)

// The headline answers "do I need to do anything?" before any number does.
const headline = computed(() => {
  const n = problems.value.length
  if (n === 0) {
    return {
      main: total.value === 1 ? 'Your monitor is fine.' : `All ${total.value} monitors are fine.`,
      soft: '',
    }
  }
  const others = total.value - n
  return {
    main: n === 1 ? '1 monitor needs you.' : `${n} monitors need you.`,
    soft:
      others === 0 ? '' : others === 1 ? 'The other one is fine.' : `The other ${others} are fine.`,
  }
})

const avgUptime = computed(() => {
  const samples = rows.value.map((r) => r.uptime24h).filter((v): v is number => v !== null)
  if (samples.length === 0) return null
  return samples.reduce((a, b) => a + b, 0) / samples.length
})

const smsUsed = computed(() => billingInfo.value?.smsCreditsUsed ?? 0)
const smsTotal = computed(() => billingInfo.value?.smsCreditsLimit ?? 0)

const facts = computed(() => {
  const list = [`${upCount.value} of ${total.value} up`]
  if (avgUptime.value !== null) list.unshift(`${avgUptime.value.toFixed(2)}% average uptime today`)
  if (smsTotal.value > 0) list.push(`${smsUsed.value} of ${smsTotal.value} SMS credits used`)
  return list.join(' · ')
})
</script>

<template>
  <AppLayout>
    <div class="mx-auto flex max-w-[1200px] flex-col gap-12 px-4 pt-10 pb-20 md:px-6 md:pt-14">
      <p v-if="isPending" class="text-sm text-[var(--text-dim)]">Loading…</p>

      <!-- No monitors yet -->
      <section
        v-else-if="total === 0"
        class="flex flex-wrap items-end justify-between gap-6 border-b border-[var(--border)] pb-12"
      >
        <div class="flex max-w-[640px] flex-col gap-3">
          <h1
            class="text-[34px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] md:text-[44px]"
          >
            Nothing to watch yet.
          </h1>
          <p class="text-base text-[var(--text-dim)]">
            No monitors yet. Add a cron, uptime, SSL, domain, port, or DNS monitor to get started.
          </p>
        </div>
        <AddMonitorMenu />
      </section>

      <template v-else>
        <section class="flex flex-wrap items-end justify-between gap-6">
          <div class="flex max-w-[720px] flex-col gap-3">
            <h1
              class="text-[34px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] md:text-[44px]"
            >
              {{ headline.main }}
              <template v-if="headline.soft">
                <br />
                <span class="text-[var(--text-muted)]">{{ headline.soft }}</span>
              </template>
            </h1>
            <p class="text-base text-[var(--text-dim)]">{{ facts }}</p>
          </div>
          <AddMonitorMenu />
        </section>

        <AttentionPanel :rows="problems" />

        <section aria-labelledby="all-monitors-heading" class="flex flex-col gap-3.5">
          <h2
            id="all-monitors-heading"
            class="text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]"
          >
            All monitors
          </h2>
          <MonitorsTable v-model:filter="filter" :rows="rows" />
        </section>

        <UpcomingExpirationsCard :rows="rows" />
      </template>
    </div>
  </AppLayout>
</template>
