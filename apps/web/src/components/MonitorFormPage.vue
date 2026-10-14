<script setup lang="ts">
// The page chrome every monitor create/edit screen wraps its fields in:
// layout, back link, title, the form element itself, the error/plan-limit
// row, and the submit/cancel buttons. Fourteen views carried a copy of this;
// only the title, the route to go back to, and the button labels ever
// differed. Fields go in the default slot — this component owns no form
// state and never branches on create vs edit.
import type { RouteLocationRaw } from 'vue-router'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import UpgradePrompt from '@/components/UpgradePrompt.vue'

const props = withDefaults(
  defineProps<{
    title: string
    // Where both "← Back" and "Cancel" go. They have always pointed at the
    // same place, so this is one prop rather than two.
    backTo: RouteLocationRaw
    submitLabel: string
    submittingLabel: string
    error?: string
    submitting?: boolean
    loading?: boolean
    // Renders the error as an UpgradePrompt instead of plain text — set when
    // the API rejected with plan_limit_reached.
    limitReached?: boolean
  }>(),
  { error: '', submitting: false, loading: false, limitReached: false },
)

const emit = defineEmits<{ submit: [] }>()

const router = useRouter()

function goBack() {
  router.push(props.backTo)
}
</script>

<template>
  <AppLayout>
    <div class="mx-auto flex max-w-[1200px] flex-col gap-8 px-4 *:max-w-[720px] pt-8 pb-20 md:px-6">
      <div class="flex flex-col gap-4">
        <button
          type="button"
          class="self-start py-1 text-sm text-[var(--text-dim)] hover:text-[var(--text)]"
          @click="goBack"
        >
          ← Back
        </button>
        <h1
          class="text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] md:text-[34px]"
        >
          {{ title }}
        </h1>
      </div>

      <p v-if="loading" class="text-sm text-[var(--text-dim)]">Loading…</p>

      <form v-else class="flex flex-col gap-6" @submit.prevent="emit('submit')">
        <slot />

        <UpgradePrompt v-if="limitReached" :message="error" />
        <p v-else-if="error" class="text-sm text-[var(--status-down)]">{{ error }}</p>

        <div
          class="flex flex-wrap items-center justify-between gap-3 border-t border-[var(--border)] pt-6"
        >
          <div class="flex gap-3">
            <Button type="submit" :disabled="submitting">
              {{ submitting ? submittingLabel : submitLabel }}
            </Button>
            <Button variant="secondary" type="button" @click="goBack">Cancel</Button>
          </div>
          <!-- Destructive or screen-specific actions (e.g. the maintenance
               window's delete-with-confirm) sit opposite save/cancel. -->
          <slot name="actions" />
        </div>
      </form>
    </div>
  </AppLayout>
</template>
