<script setup lang="ts">
import logoDark from '@/assets/logo-dark.svg'
</script>

<template>
  <div class="min-h-screen flex items-center justify-center px-4" style="background-color: var(--bg)">
    <div class="w-full max-w-sm space-y-6">
      <!-- Logo -->
      <div class="flex justify-center">
        <RouterLink to="/">
          <img :src="logoDark" alt="checkmeup" class="h-8" />
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
