<script setup lang="ts">
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import MaintenanceMonitorPicker from '@/components/MaintenanceMonitorPicker.vue'
import type { MaintenanceWindowFormState } from '@/lib/maintenanceWindowForm'

// The create and edit screens render exactly these fields with no
// differences — no mode props and no slots needed here.
const form = defineModel<MaintenanceWindowFormState>({ required: true })
</script>

<template>
  <div class="space-y-5">
    <div>
      <Label for="title">Title</Label>
      <Input id="title" v-model="form.title" placeholder="Database migration" class="mt-1" required />
    </div>

    <div>
      <Label for="message">
        Message <span style="color: var(--text-muted)">(optional, shown on status page)</span>
      </Label>
      <Input
        id="message"
        v-model="form.message"
        placeholder="We're upgrading our database, expect brief downtime."
        class="mt-1"
      />
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div>
        <Label for="startsAt">Starts</Label>
        <input
          id="startsAt"
          v-model="form.startsAt"
          type="datetime-local"
          class="mt-1 flex h-10 w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2"
          style="background-color: var(--surface); border-color: var(--border); color: var(--text)"
          required
        />
      </div>
      <div>
        <Label for="endsAt">Ends</Label>
        <input
          id="endsAt"
          v-model="form.endsAt"
          type="datetime-local"
          :disabled="form.noEnd"
          class="mt-1 flex h-10 w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2 disabled:opacity-50"
          style="background-color: var(--surface); border-color: var(--border); color: var(--text)"
        />
        <label class="flex items-center gap-2 mt-1.5 text-xs" style="color: var(--text-muted)">
          <input v-model="form.noEnd" type="checkbox" />
          No end date — I'll end it manually
        </label>
      </div>
    </div>

    <div>
      <Label>Monitors</Label>
      <MaintenanceMonitorPicker v-model="form.monitors" class="mt-1" />
    </div>
  </div>
</template>
