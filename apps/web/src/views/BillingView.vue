<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { initializePaddle, type Paddle } from '@paddle/paddle-js'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import SettingsSection from '@/components/SettingsSection.vue'
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

// Annual prices are each exactly 10x the monthly price (~2 months free, PRD annual-billing).
const monthlyPrice: Record<string, number> = { solo: 9, startup: 29, enterprise: 99 }
const annualPrice: Record<string, number> = { solo: 90, startup: 290, enterprise: 990 }
const planRank: Record<string, number> = { hobby: 0, solo: 1, startup: 2, enterprise: 3 }

const planPrice = computed(() => {
  if (!info.value || info.value.plan === 'hobby') return 'Free'
  const plan = info.value.plan
  return info.value.billingCycle === 'annual' ? `$${annualPrice[plan]}/yr` : `$${monthlyPrice[plan]}/mo`
})

const upgradeOptions = computed(() => {
  if (!info.value || cancelScheduled.value) return []
  const currentRank = planRank[info.value.plan]
  return (['solo', 'startup', 'enterprise'] as const).filter((p) => planRank[p] > currentRank)
})

// A cancellation that's already scheduled (subscriptionStatus
// 'cancel_scheduled', set by the webhook the moment ChangePlan's
// cancel-to-Hobby call succeeds) locks the subscription on Paddle's side —
// any further upgrade/downgrade/cancel attempt gets rejected as a conflict
// until it actually takes effect at period end. Hiding the actions here
// (rather than letting them fail) is what stops a second "Cancel
// subscription" click from looking like a broken feature.
//
// Once the org is actually back on Hobby, though, there's nothing left to
// cancel — a stale 'cancel_scheduled' status lingering after the plan has
// already reverted (e.g. an out-of-order or retried webhook) must never
// hide the Upgrade section for a Hobby org; that's a dead end with no way
// back in.
const cancelScheduled = computed(
  () => info.value?.plan !== 'hobby' && info.value?.subscriptionStatus === 'cancel_scheduled',
)

const downgradeOptions = computed(() => {
  if (!info.value || cancelScheduled.value) return []
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
      // The plan itself doesn't change yet (cancellation takes effect at
      // period end) — wait for subscriptionStatus to flip to
      // 'cancel_scheduled' instead, so the cancel button actually
      // disappears once the webhook lands, rather than staying clickable.
      await waitForPlanUpdate(() => info.value?.subscriptionStatus === 'cancel_scheduled')
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

const smsCreditsPct = computed(() => {
  if (!info.value) return 0
  if (info.value.smsCreditsLimit <= 0) return 0
  return Math.min(100, Math.round((info.value.smsCreditsUsed / info.value.smsCreditsLimit) * 100))
})

function limitLabel(used: number, limit: number) {
  return limit === -1 ? `${used} / unlimited` : `${used} / ${limit}`
}
</script>

<template>
  <AppLayout>
    <div class="mx-auto max-w-[1000px] px-4 pt-8 pb-20 md:px-6">
      <h1 class="mb-8 text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] md:text-[34px]">Billing</h1>

      <p v-if="loading" class="text-sm text-[var(--text-dim)]">Loading…</p>
      <p v-else-if="error" class="text-sm text-[var(--status-down)]">{{ error }}</p>

      <template v-else-if="info">
        <SettingsSection title="Current plan">
          <template #description>
            <template v-if="cancelScheduled">
              Cancellation scheduled — you'll keep {{ planLabel[info.plan] }} access until
              {{ info.planRenewsAt }}, then drop to Hobby.
            </template>
            <template v-else-if="info.planRenewsAt">
              {{ info.subscriptionStatus === 'canceled' ? 'Access until' : 'Renews' }}
              {{ info.planRenewsAt }}
            </template>
          </template>
          <div class="flex flex-wrap items-center justify-between gap-4">
            <p class="text-[30px] font-semibold tracking-[-0.035em] text-[var(--text-strong)]">
              {{ planLabel[info.plan] }}
              <span class="ml-1 font-mono text-base font-normal tracking-normal text-[var(--text-dim)]">
                {{ planPrice }}
              </span>
              <span v-if="syncingPlan" class="ml-1 text-xs font-normal tracking-normal text-[var(--text-dim)]">
                updating…
              </span>
            </p>
            <a
              v-if="info.customerPortalUrl && info.plan !== 'hobby'"
              :href="info.customerPortalUrl"
              target="_blank"
              rel="noopener"
              class="rounded-[10px] border border-[var(--border-strong)] px-4 py-2 text-sm font-medium text-[var(--text)] hover:bg-[var(--surface-raised)]"
            >
              Manage subscription →
            </a>
          </div>
        </SettingsSection>

        <SettingsSection title="Usage">
          <template #description>What your plan allows, and how much of it is in use.</template>
          <div class="flex flex-col gap-5">
            <div>
              <div class="mb-1.5 flex justify-between text-sm">
                <span class="text-[var(--text)]">Monitors</span>
                <span class="font-mono text-[13px] text-[var(--text-dim)]">{{ limitLabel(info.monitorCount, info.monitorLimit) }}</span>
              </div>
              <div class="h-1.5 overflow-hidden rounded-full bg-[var(--surface-raised)]">
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
              <div class="mb-1.5 flex justify-between text-sm">
                <span class="text-[var(--text)]">Status pages</span>
                <span class="font-mono text-[13px] text-[var(--text-dim)]">{{ limitLabel(info.statusPageCount, info.statusPageLimit) }}</span>
              </div>
              <div class="h-1.5 overflow-hidden rounded-full bg-[var(--surface-raised)]">
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
              <div class="mb-1.5 flex justify-between text-sm">
                <span class="text-[var(--text)]">Notification channels</span>
                <span class="font-mono text-[13px] text-[var(--text-dim)]">{{ limitLabel(info.notificationChannelCount, info.notificationChannelLimit) }}</span>
              </div>
              <div class="h-1.5 overflow-hidden rounded-full bg-[var(--surface-raised)]">
                <div
                  class="h-full rounded-full transition-all"
                  :style="{
                    width: info.notificationChannelLimit === -1 ? '0%' : `${notificationChannelPct}%`,
                    backgroundColor: notificationChannelPct >= 90 ? 'var(--status-down)' : 'var(--accent)',
                  }"
                />
              </div>
            </div>
            <div>
              <div class="mb-1.5 flex justify-between text-sm">
                <span class="text-[var(--text)]">SMS credits (this month)</span>
                <span v-if="info.smsCreditsLimit <= 0" class="text-[13px] text-[var(--text-dim)]">Not available on your plan</span>
                <span v-else class="font-mono text-[13px] text-[var(--text-dim)]">{{ limitLabel(info.smsCreditsUsed, info.smsCreditsLimit) }}</span>
              </div>
              <div v-if="info.smsCreditsLimit > 0" class="h-1.5 overflow-hidden rounded-full bg-[var(--surface-raised)]">
                <div
                  class="h-full rounded-full transition-all"
                  :style="{
                    width: `${smsCreditsPct}%`,
                    backgroundColor: smsCreditsPct >= 90 ? 'var(--status-down)' : 'var(--accent)',
                  }"
                />
              </div>
            </div>
            <p class="text-[13px] text-[var(--text-dim)]">
              Min uptime check interval: {{ info.minIntervalMins }} min
            </p>
          </div>
        </SettingsSection>

        <SettingsSection v-if="upgradeOptions.length > 0 || downgradeOptions.length > 0" title="Change plan">
          <template #description>Switch plans any time; moving to a cheaper plan is prorated.</template>
          <div class="flex flex-col gap-6">
            <!-- Cycle toggle (shared by Upgrade and paid-tier Downgrade lists) -->
            <div v-if="showCycleToggle" class="inline-flex self-start gap-0.5 rounded-[10px] bg-[var(--surface-raised)] p-[3px]">
              <button
                type="button"
                class="min-h-[34px] rounded-lg px-4 text-sm font-medium transition-colors hover:cursor-pointer"
                :class="cycle === 'monthly' ? 'bg-[var(--bg)] text-[var(--text-strong)] shadow-[0_1px_2px_rgb(0_0_0/12%)]' : 'text-[var(--text-dim)] hover:text-[var(--text)]'"
                @click="cycle = 'monthly'"
              >
                Monthly
              </button>
              <button
                type="button"
                class="min-h-[34px] rounded-lg px-4 text-sm font-medium transition-colors hover:cursor-pointer"
                :class="cycle === 'annual' ? 'bg-[var(--bg)] text-[var(--text-strong)] shadow-[0_1px_2px_rgb(0_0_0/12%)]' : 'text-[var(--text-dim)] hover:text-[var(--text)]'"
                @click="cycle = 'annual'"
              >
                Annual <span class="text-[var(--accent)]">— 2 months free</span>
              </button>
            </div>

            <div v-if="upgradeOptions.length > 0">
              <p class="mb-2 text-[13px] font-semibold text-[var(--text-dim)]">Upgrade</p>
              <ul class="divide-y divide-[var(--border)] rounded-[14px] border border-[var(--border)]">
              <li
                v-for="plan in upgradeOptions"
                :key="plan"
                class="flex flex-wrap items-center justify-between gap-3 px-4 py-3.5"
              >
                <div>
                  <p class="font-medium text-[var(--text-strong)]">{{ planLabel[plan] }}</p>
                  <p class="font-mono text-[13px] text-[var(--text-dim)]">
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
              </li>
              </ul>
              <p v-if="checkoutError" class="mt-3 text-sm text-[var(--status-down)]">{{ checkoutError }}</p>
            </div>

            <div v-if="downgradeOptions.length > 0">
              <p class="mb-2 text-[13px] font-semibold text-[var(--text-dim)]">Downgrade</p>
              <ul class="divide-y divide-[var(--border)] rounded-[14px] border border-[var(--border)]">
              <li
                v-for="plan in downgradeOptions"
                :key="plan"
                class="flex flex-wrap items-center justify-between gap-3 px-4 py-3.5"
              >
                <div>
                  <p class="font-medium text-[var(--text-strong)]">{{ planLabel[plan] }}</p>
                  <p class="font-mono text-[13px] text-[var(--text-dim)]">
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
              </li>
              </ul>
              <p v-if="changePlanError" class="mt-3 text-sm text-[var(--status-down)]">{{ changePlanError }}</p>
            </div>
          </div>
        </SettingsSection>

        <!-- Downgrade confirm dialog -->
        <div
          v-if="downgradeTarget"
          class="fixed inset-0 flex items-center justify-center z-50"
          style="background-color: rgb(0 0 0 / 50%)"
          @click.self="downgradeTarget = null"
        >
          <div class="mx-4 w-full max-w-sm rounded-[18px] border border-[var(--border)] bg-[var(--bg)] p-6 shadow-[0_24px_60px_rgb(0_0_0/18%)]">
            <h3 class="mb-2 text-lg font-semibold tracking-[-0.015em] text-[var(--text-strong)]">
              {{ downgradeTarget === 'hobby' ? 'Cancel subscription?' : `Downgrade to ${downgradeTargetLabel}?` }}
            </h3>
            <p class="mb-5 text-sm text-[var(--text-dim)]">
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
