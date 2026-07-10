<script setup lang="ts">
import { computed } from 'vue'
import { useTheme } from '@/lib/theme'
import logoDark from '@/assets/logo-dark.svg'
import logoLight from '@/assets/logo-light.svg'

const { theme } = useTheme()
const logo = computed(() => (theme.value === 'light' ? logoLight : logoDark))
</script>

<template>
  <div class="min-h-screen flex items-center justify-center px-4" style="background-color: var(--bg)">
    <div class="w-full max-w-sm space-y-6">
      <!-- Logo -->
      <div class="flex justify-center">
        <RouterLink to="/">
          <img :src="logo" alt="Checkmeup" class="h-8" />
        </RouterLink>
      </div>

      <!-- Card -->
      <div class="rounded-xl border p-6 space-y-4"
           style="background-color: var(--surface); border-color: var(--border)">
        <slot />
      </div>
    </div>
  </div>
</template>
