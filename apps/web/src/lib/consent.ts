import { ref } from 'vue'

export type ConsentStatus = 'granted' | 'denied'

const STORAGE_KEY = 'cookie_consent'

function stored(): ConsentStatus | undefined {
  // Guarded for the build-time prerender script (scripts/prerender.mjs),
  // which renders App.vue — and therefore this module — under Node, with no
  // localStorage global. Real browsers always take the localStorage branch.
  if (typeof localStorage === 'undefined') return undefined
  const value = localStorage.getItem(STORAGE_KEY)
  return value === 'granted' || value === 'denied' ? value : undefined
}

// Module-level singleton, same pattern as useTheme — every component sees the same status.
const status = ref<ConsentStatus | undefined>(stored())

function apply(value: ConsentStatus) {
  status.value = value
  localStorage.setItem(STORAGE_KEY, value)
}

export function useConsent() {
  function grant() {
    apply('granted')
  }

  function deny() {
    apply('denied')
  }

  return { status, grant, deny }
}
