<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { statusPagesApi } from '@/api/statusPages'
import { useStatusPage } from '@/composables/useStatusPages'
import { useBilling } from '@/composables/useBilling'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const title = ref('')
const description = ref('')
const logoUrl = ref('')
const slug = ref('')
const hideBranding = ref(false)
const layout = ref<'classic' | 'grid'>('classic')
const submitting = ref(false)
const error = ref('')

const { data: page, isPending: loading, error: loadError } = useStatusPage(id)
watch(
  page,
  (p) => {
    if (!p) return
    title.value = p.title
    description.value = p.description
    logoUrl.value = p.logoUrl
    slug.value = p.slug
    hideBranding.value = p.hideBranding
    layout.value = p.layout
  },
  { immediate: true },
)
watch(loadError, (e) => {
  if (e) error.value = e.message
})

const { data: billingInfo } = useBilling()
const hobbyPlanNoBrandingToggle = computed(() => billingInfo.value?.plan === 'hobby')

async function submit() {
  error.value = ''
  if (!title.value.trim()) {
    error.value = 'Title is required'
    return
  }

  submitting.value = true
  try {
    await statusPagesApi.update(id, {
      title: title.value.trim(),
      description: description.value.trim(),
      logoUrl: logoUrl.value.trim(),
      hideBranding: hideBranding.value,
      layout: layout.value,
    })
    router.push({ name: 'status-page-detail', params: { id } })
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to update page'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-xl mx-auto">
      <div class="flex items-center gap-3 mb-6">
        <button
          class="text-sm"
          style="color: var(--text-muted)"
          @click="router.push({ name: 'status-page-detail', params: { id } })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">Edit status page</h1>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>

      <form
        v-else
        class="rounded-xl border p-6 space-y-5"
        style="background-color: var(--surface); border-color: var(--border)"
        @submit.prevent="submit"
      >
        <div>
          <Label>Slug</Label>
          <p class="mt-1 text-sm font-mono" style="color: var(--text-dim)">/status/{{ slug }}</p>
          <p class="text-xs mt-0.5" style="color: var(--text-muted)">
            Slug cannot be changed after creation.
          </p>
        </div>

        <div>
          <Label for="title">Page title</Label>
          <Input id="title" v-model="title" class="mt-1" required />
        </div>

        <div>
          <Label for="description">Description <span style="color: var(--text-muted)">(optional)</span></Label>
          <Input id="description" v-model="description" placeholder="Live status of our services" class="mt-1" />
        </div>

        <div>
          <Label for="logoUrl">Logo URL <span style="color: var(--text-muted)">(optional)</span></Label>
          <Input id="logoUrl" v-model="logoUrl" placeholder="https://example.com/logo.png" class="mt-1" />
        </div>

        <div>
          <Label>Layout</Label>
          <div class="mt-1 flex flex-col gap-2">
            <label class="flex items-center gap-2 text-sm" style="color: var(--text)">
              <input v-model="layout" type="radio" value="classic" name="layout" />
              Classic — single-column, monitors and incidents stacked
            </label>
            <label class="flex items-center gap-2 text-sm" style="color: var(--text)">
              <input v-model="layout" type="radio" value="grid" name="layout" />
              Grid — monitor cards in a wide grid with an incident sidebar
            </label>
          </div>
        </div>

        <div>
          <label
            class="flex items-center gap-2 text-sm"
            :style="{ color: hobbyPlanNoBrandingToggle ? 'var(--text-muted)' : 'var(--text)' }"
          >
            <input v-model="hideBranding" type="checkbox" :disabled="hobbyPlanNoBrandingToggle" />
            Hide "Powered by Checkmeup" and FAQ/Terms/Privacy links on this page
          </label>
          <p v-if="hobbyPlanNoBrandingToggle" class="mt-1 text-xs" style="color: var(--text-muted)">
            Hiding branding requires a paid plan —
            <RouterLink to="/billing" class="underline" style="color: var(--color-green-500)"
              >view plans</RouterLink
            >.
          </p>
        </div>

        <p v-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div class="flex gap-3 pt-1">
          <Button type="submit" :disabled="submitting">
            {{ submitting ? 'Saving…' : 'Save changes' }}
          </Button>
          <Button
            variant="secondary"
            type="button"
            @click="router.push({ name: 'status-page-detail', params: { id } })"
          >
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
