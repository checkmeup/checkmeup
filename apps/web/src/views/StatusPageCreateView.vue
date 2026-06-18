<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { statusPagesApi } from '@/api/statusPages'
import { ApiError } from '@/api/client'
import UpgradePrompt from '@/components/UpgradePrompt.vue'

const router = useRouter()

const title = ref('')
const slug = ref('')
const description = ref('')
const logoUrl = ref('')
const submitting = ref(false)
const error = ref('')
const limitReached = ref(false)
const slugStatus = ref<'idle' | 'checking' | 'available' | 'taken' | 'invalid'>('idle')
const slugMessage = ref('')

let slugTimer: ReturnType<typeof setTimeout> | null = null

function toSlug(s: string) {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
}

watch(title, (val) => {
  if (!slug.value || slug.value === toSlug(title.value.slice(0, -1))) {
    slug.value = toSlug(val)
  }
})

watch(slug, (val) => {
  if (!val) {
    slugStatus.value = 'idle'
    slugMessage.value = ''
    return
  }
  slugStatus.value = 'checking'
  if (slugTimer) clearTimeout(slugTimer)
  slugTimer = setTimeout(async () => {
    try {
      const result = await statusPagesApi.checkSlug(val)
      if (result.available) {
        slugStatus.value = 'available'
        slugMessage.value = 'Available'
      } else {
        slugStatus.value = result.reason.includes('format') || result.reason.includes('letter') ? 'invalid' : 'taken'
        slugMessage.value = result.reason
      }
    } catch {
      slugStatus.value = 'idle'
    }
  }, 400)
})

const slugColors: Record<string, string> = {
  available: 'var(--status-up)',
  taken: 'var(--status-down)',
  invalid: 'var(--status-down)',
  checking: 'var(--text-muted)',
}

async function submit() {
  error.value = ''
  limitReached.value = false
  if (!title.value.trim()) {
    error.value = 'Title is required'
    return
  }
  if (slugStatus.value !== 'available') {
    error.value = slugMessage.value || 'Please enter a valid, available slug'
    return
  }

  submitting.value = true
  try {
    const page = await statusPagesApi.create({
      slug: slug.value,
      title: title.value.trim(),
      description: description.value.trim(),
      logoUrl: logoUrl.value.trim(),
    })
    router.push({ name: 'status-page-detail', params: { id: page.id } })
  } catch (e: unknown) {
    if (e instanceof ApiError && e.code === 'plan_limit_reached') {
      limitReached.value = true
      error.value = e.message
    } else {
      error.value = e instanceof Error ? e.message : 'Failed to create page'
    }
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
          class="text-sm transition-colors"
          style="color: var(--text-muted)"
          @click="router.push({ name: 'status-pages' })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">New status page</h1>
      </div>

      <form
        class="rounded-xl border p-6 space-y-5"
        style="background-color: var(--surface); border-color: var(--border)"
        @submit.prevent="submit"
      >
        <div>
          <Label for="title">Page title</Label>
          <Input
            id="title"
            v-model="title"
            placeholder="Acme Status"
            class="mt-1"
            required
          />
        </div>

        <div>
          <Label for="slug">Slug</Label>
          <Input
            id="slug"
            v-model="slug"
            placeholder="acme"
            class="mt-1"
          />
          <p v-if="slugStatus !== 'idle'" class="text-xs mt-1" :style="{ color: slugColors[slugStatus] ?? 'var(--text-muted)' }">
            {{ slugStatus === 'checking' ? 'Checking…' : slugMessage }}
          </p>
          <p v-else class="text-xs mt-1" style="color: var(--text-muted)">
            Public URL will be <span class="font-mono">/status/{{ slug || 'your-slug' }}</span>
          </p>
        </div>

        <div>
          <Label for="description">Description <span style="color: var(--text-muted)">(optional)</span></Label>
          <Input
            id="description"
            v-model="description"
            placeholder="Live status of Acme services"
            class="mt-1"
          />
        </div>

        <div>
          <Label for="logoUrl">Logo URL <span style="color: var(--text-muted)">(optional)</span></Label>
          <Input
            id="logoUrl"
            v-model="logoUrl"
            placeholder="https://example.com/logo.png"
            class="mt-1"
          />
        </div>

        <UpgradePrompt v-if="limitReached" :message="error" />
        <p v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div class="flex gap-3 pt-1">
          <Button type="submit" :disabled="submitting || slugStatus === 'taken' || slugStatus === 'invalid'">
            {{ submitting ? 'Creating…' : 'Create page' }}
          </Button>
          <Button variant="secondary" type="button" @click="router.push({ name: 'status-pages' })">
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
