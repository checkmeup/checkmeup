<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { initializePaddle, type Paddle } from '@paddle/paddle-js'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { billingApi, type BillingCycle } from '@/api/billing'
import { ApiError } from '@/api/client'
import { useBilling } from '@/composables/useBilling'
import { useTheme } from '@/lib/theme'

const { theme } = useTheme()
const route = useRoute()
const router = useRouter()

// Paddle's successUrl redirect fires the instant checkout completes
// client-side, but the plan itself only updates once Paddle's subscription
// webhook lands server-side — which can take a few seconds. Stash the plan
// the user picked so the post-redirect poll below knows what to wait for.
const PENDING_PLAN_KEY = 'checkmeup:billingPendingPlan'

let paddle: Paddle | undefined
async function getPaddle(): Promise<Paddle | undefined> {
  if (paddle) return paddle
  const token = import.meta.env.VITE_PADDLE_CLIENT_TOKEN
  if (!token) return undefined
  paddle = await initializePaddle({
    token,
    environment: import.meta.env.VITE_PADDLE_ENVIRONMENT === 'production' ? 'production' : 'sandbox',
  })
  return paddle
}

const { data: info, isPending: loading, error: queryError, refetch } = useBilling()
const error = computed(() => queryError.value?.message ?? '')

const syncingPlan = ref(false)

// Polls the billing query instead of relying on a single fetch — Paddle's
// webhook that actually persists the plan change can lag behind whatever
// triggered this (the checkout successUrl redirect, or a change-plan call)
// by a few seconds. Gives up after ~12s so a slow/failed webhook doesn't
// spin forever; the org keeps its old plan showing until they refresh again.
async function waitForPlanUpdate(isSynced: () => boolean, attempts = 8, delayMs = 1500) {
  syncingPlan.value = true
  try {
    for (let i = 0; i < attempts; i++) {
      await refetch()
      if (isSynced()) return
      await new Promise((resolve) => setTimeout(resolve, delayMs))
    }
  } finally {
    syncingPlan.value = false
  }
}

onMounted(() => {
  if (route.query.upgraded === 'true') {
    const pendingPlan = sessionStorage.getItem(PENDING_PLAN_KEY)
    sessionStorage.removeItem(PENDING_PLAN_KEY)
    router.replace({ query: {} })
    void waitForPlanUpdate(() => !pendingPlan || info.value?.plan === pendingPlan)
  }
})

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

const downgradeOptions = computed(() => {
  if (!info.value) return []
  const currentRank = planRank[info.value.plan]
  return (['hobby', 'solo', 'startup'] as const).filter((p) => planRank[p] < currentRank)
})

// The cycle toggle matters for downgrades between paid tiers too, not just
// upgrades — shown whenever either list has a plan whose price depends on it.
const showCycleToggle = computed(
  () => upgradeOptions.value.length > 0 || downgradeOptions.value.some((p) => p !== 'hobby'),
)

const cycle = ref<BillingCycle>('monthly')
const checkingOut = ref<string | null>(null)
const checkoutError = ref('')

// Defaults the cycle toggle to the org's current billing cycle once billing
// info loads — otherwise an org on annual billing with no upgrade options
// left (e.g. already on Enterprise) would silently downgrade onto monthly,
// since the toggle only ever renders inside the Upgrade section.
watch(
  info,
  (v) => {
    if (v && v.plan !== 'hobby') cycle.value = v.billingCycle
  },
  { once: true },
)

const downgradeTarget = ref<string | null>(null)
const downgradeTargetLabel = computed(() =>
  downgradeTarget.value ? planLabel[downgradeTarget.value] : '',
)
const changingPlan = ref<string | null>(null)
const changePlanError = ref('')
const cancelNotice = ref('')

function effectiveMonthly(plan: string): string {
  return (annualPrice[plan] / 12).toFixed(2).replace(/\.00$/, '')
}

async function upgrade(plan: string) {
  checkingOut.value = plan
  checkoutError.value = ''
  try {
    const { transactionId } = await billingApi.createCheckout(plan, cycle.value)
    const p = await getPaddle()
    if (!p) {
      checkoutError.value = "Billing isn't activated yet — check back soon, or email andrew@checkmeup.net."
      return
    }
    sessionStorage.setItem(PENDING_PLAN_KEY, plan)
    p.Checkout.open({
      transactionId,
      settings: {
        theme: theme.value,
        successUrl: `${window.location.origin}/billing?upgraded=true`,
      },
    })
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

function confirmDowngrade() {
  if (downgradeTarget.value) downgrade(downgradeTarget.value)
}

async function downgrade(plan: string) {
  changingPlan.value = plan
  changePlanError.value = ''
  try {
    await billingApi.changePlan(plan, cycle.value)
    downgradeTarget.value = null
    if (plan === 'hobby') {
      // Cancellation takes effect at period end — nothing changes in the DB
      // yet for refetch to observe, so just confirm and leave it there.
      cancelNotice.value = 'Your subscription will be canceled at the end of the current billing period.'
    } else {
      await waitForPlanUpdate(() => info.value?.plan === plan)
    }
  } catch (e: unknown) {
    changePlanError.value = e instanceof Error ? e.message : 'Failed to change plan'
  } finally {
    changingPlan.value = null
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

const notificationChannelPct = computed(() => {
  if (!info.value) return 0
  if (info.value.notificationChannelLimit === -1) return 0
  return Math.min(100, Math.round((info.value.notificationChannelCount / info.value.notificationChannelLimit) * 100))
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
                <span v-if="syncingPlan" class="text-xs font-normal ml-1" style="color: var(--text-muted)">
                  updating…
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
            {{ info.subscriptionStatus === 'canceled' ? 'Access until' : 'Renews' }}
            {{ info.planRenewsAt }}
          </p>
          <p v-if="cancelNotice" class="text-xs mb-4" style="color: var(--text-muted)">{{ cancelNotice }}</p>

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

            <div>
              <div class="flex justify-between text-xs mb-1" style="color: var(--text-dim)">
                <span>Notification channels</span>
                <span>{{ limitLabel(info.notificationChannelCount, info.notificationChannelLimit) }}</span>
              </div>
              <div class="h-1.5 rounded-full overflow-hidden" style="background-color: var(--surface-raised)">
                <div
                  class="h-full rounded-full transition-all"
                  :style="{
                    width: info.notificationChannelLimit === -1 ? '0%' : `${notificationChannelPct}%`,
                    backgroundColor: notificationChannelPct >= 90 ? 'var(--status-down)' : 'var(--accent)',
                  }"
                />
              </div>
            </div>

            <p class="text-xs" style="color: var(--text-muted)">
              Min uptime check interval: {{ info.minIntervalMins }} min
            </p>
          </div>
        </div>

        <!-- Cycle toggle (shared by Upgrade and paid-tier Downgrade sections) -->
        <div v-if="showCycleToggle" class="flex items-center justify-end mb-4">
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

        <!-- Upgrade -->
        <div
          v-if="upgradeOptions.length > 0"
          class="rounded-xl border p-6"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <p class="font-medium mb-5" style="color: var(--text-strong)">Upgrade</p>

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
                {{ checkingOut === plan ? 'Opening checkout…' : `Upgrade to ${planLabel[plan]}` }}
              </Button>
            </div>
          </div>

          <p v-if="checkoutError" class="text-sm mt-4" style="color: var(--status-down)">{{ checkoutError }}</p>
        </div>

        <!-- Downgrade -->
        <div
          v-if="downgradeOptions.length > 0"
          class="rounded-xl border p-6 mt-6"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <p class="font-medium mb-5" style="color: var(--text-strong)">Downgrade</p>

          <div class="space-y-3">
            <div
              v-for="plan in downgradeOptions"
              :key="plan"
              class="flex items-center justify-between p-3 rounded-lg"
              style="background-color: var(--surface-raised)"
            >
              <div>
                <p class="text-sm font-medium" style="color: var(--text-strong)">{{ planLabel[plan] }}</p>
                <p class="text-xs" style="color: var(--text-muted)">
                  {{
                    plan === 'hobby'
                      ? 'Free'
                      : cycle === 'annual'
                        ? `$${annualPrice[plan]}/yr ($${effectiveMonthly(plan)}/mo)`
                        : `$${monthlyPrice[plan]}/mo`
                  }}
                </p>
              </div>
              <Button size="sm" variant="secondary" @click="downgradeTarget = plan">
                {{ plan === 'hobby' ? 'Cancel subscription' : `Downgrade to ${planLabel[plan]}` }}
              </Button>
            </div>
          </div>

          <p v-if="changePlanError" class="text-sm mt-4" style="color: var(--status-down)">{{ changePlanError }}</p>
        </div>

        <!-- Downgrade confirm dialog -->
        <div
          v-if="downgradeTarget"
          class="fixed inset-0 flex items-center justify-center z-50"
          style="background-color: rgba(0,0,0,0.6)"
          @click.self="downgradeTarget = null"
        >
          <div class="rounded-xl border p-6 max-w-sm w-full mx-4" style="background-color: var(--surface); border-color: var(--border)">
            <h3 class="font-medium mb-2" style="color: var(--text-strong)">
              {{ downgradeTarget === 'hobby' ? 'Cancel subscription?' : `Downgrade to ${downgradeTargetLabel}?` }}
            </h3>
            <p class="text-sm mb-5" style="color: var(--text-muted)">
              <template v-if="downgradeTarget === 'hobby'">
                You'll keep access to your current plan until the end of the billing period, then drop to Hobby.
              </template>
              <template v-else>
                This takes effect immediately, with a prorated charge or credit for the price difference.
              </template>
            </p>
            <div class="flex gap-3">
              <Button
                :variant="downgradeTarget === 'hobby' ? 'destructive' : 'default'"
                :disabled="changingPlan === downgradeTarget"
                @click="confirmDowngrade"
              >
                {{
                  changingPlan === downgradeTarget
                    ? 'Applying…'
                    : downgradeTarget === 'hobby'
                      ? 'Cancel subscription'
                      : `Downgrade to ${downgradeTargetLabel}`
                }}
              </Button>
              <Button variant="secondary" @click="downgradeTarget = null">Never mind</Button>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
