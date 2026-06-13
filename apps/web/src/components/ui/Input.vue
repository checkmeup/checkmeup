<script setup lang="ts">
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: string
  modelValue?: string
  type?: string
  placeholder?: string
  disabled?: boolean
  autocomplete?: string
  required?: boolean
  id?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <input
    :id="props.id"
    :type="props.type ?? 'text'"
    :value="props.modelValue"
    :placeholder="props.placeholder"
    :disabled="props.disabled"
    :autocomplete="props.autocomplete"
    :required="props.required"
    :class="
      cn(
        'flex h-10 w-full rounded-md border border-[--border] bg-[--surface] px-3 py-2 text-sm text-[--text] placeholder:text-[--text-muted] focus:outline-none focus:ring-2 focus:ring-[--color-green-500] focus:border-transparent disabled:cursor-not-allowed disabled:opacity-50 transition-colors',
        props.class,
      )
    "
    @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  />
</template>
