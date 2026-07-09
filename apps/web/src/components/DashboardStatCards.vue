<script setup lang="ts">
defineProps<{
  healthyMonitors: number
  totalMonitors: number
  healthyPct: number
  avgUptimeStr: string
  uptimeSamplesCount: number
  attentionColor: string
  attentionCount: number
  attentionNote: string
  smsUsed: number
  smsTotal: number
  smsPct: number
}>()
</script>

<template>
  <div class="grid gap-3.5 sm:grid-cols-2 lg:grid-cols-4 mb-7">
    <div
      class="rounded-2xl border p-[22px]"
      style="background-color: var(--surface); border-color: var(--border)"
    >
      <div class="text-xs font-medium mb-2.5" style="color: var(--text-dim)">
        Monitors healthy
      </div>
      <div class="flex items-baseline gap-1.5 mb-3">
        <span
          class="font-mono text-[28px] font-semibold"
          style="color: var(--text-strong); letter-spacing: -0.02em"
          >{{ healthyMonitors }}</span
        >
        <span class="text-[15px]" style="color: var(--text-muted)">/ {{ totalMonitors }}</span>
      </div>
      <div class="h-1.5 rounded overflow-hidden" style="background-color: var(--accent-wash)">
        <div
          class="h-full rounded"
          :style="{ backgroundColor: 'var(--accent)', width: `${healthyPct}%` }"
        ></div>
      </div>
      <div class="text-xs mt-2" style="color: var(--text-muted)">
        {{ healthyPct.toFixed(1) }}% healthy
      </div>
    </div>

    <div
      class="rounded-2xl border p-[22px]"
      style="background-color: var(--surface); border-color: var(--border)"
    >
      <div class="text-xs font-medium mb-2.5" style="color: var(--text-dim)">
        Avg uptime (24h)
      </div>
      <div
        class="font-mono text-[28px] font-semibold"
        style="color: var(--text-strong); letter-spacing: -0.02em"
      >
        {{ avgUptimeStr }}
      </div>
      <div class="text-xs mt-3" style="color: var(--text-muted)">
        across {{ uptimeSamplesCount }} uptime/port monitor{{ uptimeSamplesCount === 1 ? '' : 's' }}
      </div>
    </div>

    <div
      class="rounded-2xl border p-[22px]"
      style="background-color: var(--surface); border-color: var(--border)"
    >
      <div class="text-xs font-medium mb-2.5" style="color: var(--text-dim)">Needs attention</div>
      <div
        class="font-mono text-[28px] font-semibold"
        :style="{ color: attentionColor, letterSpacing: '-0.02em' }"
      >
        {{ attentionCount }}
      </div>
      <div class="text-xs mt-3 truncate" style="color: var(--text-muted)">
        {{ attentionNote }}
      </div>
    </div>

    <div
      class="rounded-2xl border p-[22px]"
      style="background-color: var(--surface); border-color: var(--border)"
    >
      <div class="text-xs font-medium mb-2.5" style="color: var(--text-dim)">SMS credits used</div>
      <template v-if="smsTotal > 0">
        <div class="flex items-baseline gap-1.5 mb-3">
          <span
            class="font-mono text-[28px] font-semibold"
            style="color: var(--text-strong); letter-spacing: -0.02em"
            >{{ smsUsed }}</span
          >
          <span class="text-[15px]" style="color: var(--text-muted)">/ {{ smsTotal }}</span>
        </div>
        <div class="h-1.5 rounded overflow-hidden" style="background-color: var(--surface-raised)">
          <div
            class="h-full rounded"
            :style="{ backgroundColor: 'var(--accent)', width: `${smsPct}%` }"
          ></div>
        </div>
        <div class="text-xs mt-2" style="color: var(--text-muted)">this billing cycle</div>
      </template>
      <div v-else class="font-mono text-[15px] font-normal" style="color: var(--text-muted)">
        Not available on your plan
      </div>
    </div>
  </div>
</template>
