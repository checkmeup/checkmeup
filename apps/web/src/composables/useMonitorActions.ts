import { ref } from 'vue'
import { ApiError } from '@/api/client'

// Pause/resume and delete for a monitor page, with the error and the
// plan-limit upgrade prompt they can raise. Every monitor type wires the same
// flow to its own API calls.
export function useMonitorActions(opts: {
  isPaused: () => boolean
  pause: () => Promise<unknown>
  resume: () => Promise<unknown>
  remove: () => Promise<unknown>
  refetch: () => Promise<unknown>
  afterDelete: () => void
}) {
  const actionError = ref('')
  const limitReached = ref(false)
  const confirmDelete = ref(false)

  async function togglePause() {
    actionError.value = ''
    limitReached.value = false
    try {
      await (opts.isPaused() ? opts.resume() : opts.pause())
      await opts.refetch()
    } catch (e: unknown) {
      if (e instanceof ApiError && e.code === 'plan_limit_reached') {
        limitReached.value = true
        actionError.value = e.message
      } else {
        actionError.value = e instanceof Error ? e.message : 'Action failed'
      }
    }
  }

  async function deleteMonitor() {
    actionError.value = ''
    try {
      await opts.remove()
      opts.afterDelete()
    } catch (e: unknown) {
      actionError.value = e instanceof Error ? e.message : 'Delete failed'
      confirmDelete.value = false
    }
  }

  return { actionError, limitReached, confirmDelete, togglePause, deleteMonitor }
}
