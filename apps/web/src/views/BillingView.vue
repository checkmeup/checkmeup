<script setup lang="ts">
import { ref, computed } from 'vue'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { billingApi, type BillingCycle } from '@/api/billing'
import { ApiError } from '@/api/client'
import { useBilling } from '@/composables/useBilling'

const { data: info, isPending: loading, error: queryError } = useBilling()
const error = computed(() => queryError.value?.message ?? '')

const planLabel: Record<string, string> = {
  hobby: 'Hobby',
  solo: 'Solo',
  startup: 'Startup',
  enterprise: 'Enterprise',
}

// Annual prices are each exactly 10x the monthly price (~2 months free, EP-27).
const monthlyPrice: Record<string, number> = { solo: 9, startup: 29, enterprise: 99 }
const annualPrice: Record<string, number> = { solo: 90, startup: 290, enterprise: 990 }
const planRank: Record<string, number> = { hobby: 0, solo: 1, startup: 2, enterprise: 3 }

const planPrice = computed(() => {
  if (!info.value || info.value.plan === 'hobby') return 'Free'
  const plan = info.value.plan
  return info.value.billingCycle === 'annual' ? `$${annualPrice[plan]}/yr` : `$${monthlyPrice[plan]}/mo`
})

const upgradeOptions = computed(() => {
  if (!info.value) return []
  const currentRank = planRank[info.value.plan]
  return (['solo', 'startup', 'enterprise'] as const).filter((p) => planRank[p] > currentRank)
})

const cycle = ref<BillingCycle>('monthly')
const checkingOut = ref<string | null>(null)
const checkoutError = ref('')

function effectiveMonthly(plan: string): string {
  return (annualPrice[plan] / 12).toFixed(2).replace(/\.00$/, '')
}

async function upgrade(plan: string) {
  checkingOut.value = plan
  checkoutError.value = ''
  try {
    const { url } = await billingApi.createCheckout(plan, cycle.value)
    window.location.href = url
  } catch (e: unknown) {
    if (e instanceof ApiError && e.code === 'not_configured') {
      checkoutError.value = "Billing isn't activated yet — check back soon, or email andrew@checkmeup.net."
    } else {
      checkoutError.value = e instanceof Error ? e.message : 'Failed to start checkout'
    }
  } finally {
    checkingOut.value = null
  }
}

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
                  {{ planPrice }}
                </span>
              </p>
            </div>
            <a
              v-if="info.customerPortalUrl && info.plan !== 'hobby'"
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

        <!-- Upgrade -->
        <div
          v-if="upgradeOptions.length > 0"
          class="rounded-xl border p-6"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="flex items-center justify-between mb-5">
            <p class="font-medium" style="color: var(--text-strong)">Upgrade</p>
            <div class="inline-flex rounded-md border p-1" style="border-color: var(--border)">
              <button
                type="button"
                class="px-3 py-1 rounded text-xs transition-colors hover:cursor-pointer"
                :style="
                  cycle === 'monthly'
                    ? 'background-color: var(--surface-raised); color: var(--text-strong)'
                    : 'color: var(--text-muted)'
                "
                @click="cycle = 'monthly'"
              >
                Monthly
              </button>
              <button
                type="button"
                class="px-3 py-1 rounded text-xs transition-colors hover:cursor-pointer"
                :style="
                  cycle === 'annual'
                    ? 'background-color: var(--surface-raised); color: var(--text-strong)'
                    : 'color: var(--text-muted)'
                "
                @click="cycle = 'annual'"
              >
                Annual <span style="color: var(--color-green-500)">— 2 months free</span>
              </button>
            </div>
          </div>

          <div class="space-y-3">
            <div
              v-for="plan in upgradeOptions"
              :key="plan"
              class="flex items-center justify-between p-3 rounded-lg"
              style="background-color: var(--surface-raised)"
            >
              <div>
                <p class="text-sm font-medium" style="color: var(--text-strong)">{{ planLabel[plan] }}</p>
                <p class="text-xs" style="color: var(--text-muted)">
                  {{
                    cycle === 'annual'
                      ? `$${annualPrice[plan]}/yr ($${effectiveMonthly(plan)}/mo)`
                      : `$${monthlyPrice[plan]}/mo`
                  }}
                </p>
              </div>
              <Button size="sm" :disabled="checkingOut === plan" @click="upgrade(plan)">
                {{ checkingOut === plan ? 'Redirecting…' : `Upgrade to ${planLabel[plan]}` }}
              </Button>
            </div>
          </div>

          <p v-if="checkoutError" class="text-sm mt-4" style="color: var(--status-down)">{{ checkoutError }}</p>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
