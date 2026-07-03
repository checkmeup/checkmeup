<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterView } from 'vue-router'
import ConsentBanner from '@/components/ConsentBanner.vue'
import { useConsent } from '@/lib/consent'
import { loadGtm } from '@/lib/analytics'

const { status } = useConsent()

onMounted(() => {
  if (status.value === 'granted') loadGtm()
})
</script>

<template>
  <RouterView />
  <ConsentBanner />
</template>
