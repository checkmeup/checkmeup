<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { billingApi, type BillingInfo } from '@/api/billing'

const info = ref<BillingInfo | null>(null)
const loading = ref(true)
const error = ref('')
const upgrading = ref('')

onMounted(async () => {
  try {
    info.value = await billingApi.getInfo()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load billing info'
  } finally {
    loading.value = false
  }
})

async function upgrade(plan: string) {
  upgrading.value = plan
  try {
    const { url } = await billingApi.createCheckout(plan)
    window.location.href = url
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to start checkout'
    upgrading.value = ''
  }
}

const planLabel: Record<string, string> = {
  hobbyist: 'Hobbyist',
  indie: 'Indie',
  studio: 'Studio',
  agency: 'Agency',
}

const planPrice: Record<string, string> = {
  hobbyist: 'Free',
  indie: '$12/mo',
  studio: '$39/mo',
  agency: '$99/mo',
}

const upgradePlans = ['indie', 'studio', 'agency']

const monitorPct = computed(() => {
  if (!info.value) return 0
  if (info.value.monitorLimit === -1) return 0
  return Math.min(100, Math.round((info.value.monitorCount / info.value.monitorLimit) * 100))
})

const statusPagePct = computed(() => {
  if (!info.value) return 0
  if (info.value.statusPageLimit === -1) return 0
  return Math.min(100, Math.round((info.value.statusPageCount / info.value.statusPageLimit) * 100))
})

function limitLabel(used: number, limit: number) {
  return limit === -1 ? `${used} / unlimited` : `${used} / ${limit}`
}
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-2xl mx-auto">
      <h1 class="text-2xl font-semibold mb-6" style="color: var(--text-strong)">Billing</h1>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>
      <div v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</div>

      <template v-else-if="info">
        <!-- Current plan -->
        <div
          class="rounded-xl border p-6 mb-6"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="flex items-center justify-between mb-4">
            <div>
              <p class="text-xs font-medium uppercase tracking-wider mb-1" style="color: var(--text-muted)">Current plan</p>
              <p class="text-2xl font-bold" style="color: var(--text-strong)">
                {{ planLabel[info.plan] }}
                <span class="text-base font-normal ml-1" style="color: var(--text-muted)">
                  {{ planPrice[info.plan] }}
                </span>
              </p>
            </div>
            <a
              v-if="info.customerPortalUrl && info.plan !== 'hobbyist'"
              :href="info.customerPortalUrl"
              target="_blank"
              rel="noopener"
              class="text-sm px-3 py-1.5 rounded-md border"
              style="color: var(--text-dim); border-color: var(--border)"
            >
              Manage subscription →
            </a>
          </div>

          <p v-if="info.planRenewsAt" class="text-xs mb-4" style="color: var(--text-muted)">
            {{ info.subscriptionStatus === 'cancelled' ? 'Access until' : 'Renews' }}
            {{ info.planRenewsAt }}
          </p>

          <!-- Usage bars -->
          <div class="space-y-4">
            <div>
              <div class="flex justify-between text-xs mb-1" style="color: var(--text-dim)">
                <span>Monitors</span>
                <span>{{ limitLabel(info.monitorCount, info.monitorLimit) }}</span>
              </div>
              <div class="h-1.5 rounded-full overflow-hidden" style="background-color: var(--surface-raised)">
                <div
                  class="h-full rounded-full transition-all"
                  :style="{
                    width: info.monitorLimit === -1 ? '0%' : `${monitorPct}%`,
                    backgroundColor: monitorPct >= 90 ? 'var(--status-down)' : 'var(--accent)',
                  }"
                />
              </div>
            </div>

            <div>
              <div class="flex justify-between text-xs mb-1" style="color: var(--text-dim)">
                <span>Status pages</span>
                <span>{{ limitLabel(info.statusPageCount, info.statusPageLimit) }}</span>
              </div>
              <div class="h-1.5 rounded-full overflow-hidden" style="background-color: var(--surface-raised)">
                <div
                  class="h-full rounded-full transition-all"
                  :style="{
                    width: info.statusPageLimit === -1 ? '0%' : `${statusPagePct}%`,
                    backgroundColor: statusPagePct >= 90 ? 'var(--status-down)' : 'var(--accent)',
                  }"
                />
              </div>
            </div>

            <p class="text-xs" style="color: var(--text-muted)">
              Min uptime check interval: {{ info.minIntervalMins }} min
            </p>
          </div>
        </div>

        <!-- Upgrade options -->
        <div v-if="info.plan !== 'agency'">
          <h2 class="text-sm font-medium mb-3" style="color: var(--text-dim)">Upgrade your plan</h2>
          <div class="grid gap-3 sm:grid-cols-3">
            <div
              v-for="plan in upgradePlans.filter(p => p !== info!.plan)"
              :key="plan"
              class="rounded-xl border p-4 space-y-3"
              style="background-color: var(--surface); border-color: var(--border)"
            >
              <div>
                <p class="font-semibold" style="color: var(--text-strong)">{{ planLabel[plan] }}</p>
                <p class="text-sm" style="color: var(--text-muted)">{{ planPrice[plan] }}</p>
              </div>
              <Button
                size="sm"
                class="w-full"
                :disabled="upgrading === plan"
                @click="upgrade(plan)"
              >
                {{ upgrading === plan ? 'Redirecting…' : 'Upgrade' }}
              </Button>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
