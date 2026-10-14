<script setup lang="ts">
import { computed } from 'vue'
import { useTheme } from '@/lib/theme'
import logoDark from '@/assets/logo-dark.svg'
import logoLight from '@/assets/logo-light.svg'

const { theme } = useTheme()
const logo = computed(() => (theme.value === 'light' ? logoLight : logoDark))
</script>

<template>
  <div class="min-h-screen flex items-center justify-center px-4 bg-[var(--bg)]">
    <div class="w-full max-w-[400px] space-y-8 py-12">
      <!-- Logo -->
      <div class="flex justify-center">
        <RouterLink to="/">
          <img :src="logo" alt="Checkmeup" class="h-8" />
        </RouterLink>
      </div>

      <!-- Card -->
      <div class="space-y-4 rounded-[18px] border border-[var(--border)] p-7">
        <slot />
      </div>
    </div>
  </div>
</template>
