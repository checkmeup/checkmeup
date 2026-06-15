<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { statusPagesApi } from '@/api/statusPages'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const title = ref('')
const description = ref('')
const logoUrl = ref('')
const slug = ref('')
const loading = ref(true)
const submitting = ref(false)
const error = ref('')

onMounted(async () => {
  try {
    const page = await statusPagesApi.get(id)
    title.value = page.title
    description.value = page.description
    logoUrl.value = page.logoUrl
    slug.value = page.slug
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load page'
  } finally {
    loading.value = false
  }
})

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
