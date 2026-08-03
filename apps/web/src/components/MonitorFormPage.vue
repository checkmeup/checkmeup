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
    <div class="p-8 max-w-xl mx-auto">
      <div class="flex items-center gap-3 mb-6">
        <button class="text-sm transition-colors" style="color: var(--text-muted)" @click="goBack">
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">{{ title }}</h1>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>

      <form
        v-else
        class="rounded-xl border p-6 space-y-5"
        style="background-color: var(--surface); border-color: var(--border)"
        @submit.prevent="emit('submit')"
      >
        <slot />

        <UpgradePrompt v-if="limitReached" :message="error" />
        <p v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div class="flex items-center justify-between pt-1">
          <div class="flex gap-3">
            <Button type="submit" :disabled="submitting">
              {{ submitting ? submittingLabel : submitLabel }}
            </Button>
            <Button variant="secondary" type="button" @click="goBack">Cancel</Button>
          </div>
          <!-- Destructive or screen-specific actions (e.g. the maintenance
               window's delete-with-confirm) sit opposite save/cancel. With no
               actions slot filled, justify-between leaves the single child
               left-aligned, matching the original single-row layout. -->
          <slot name="actions" />
        </div>
      </form>
    </div>
  </AppLayout>
</template>
