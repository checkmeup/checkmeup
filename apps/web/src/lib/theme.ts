import { ref } from 'vue'

export type Theme = 'light' | 'dark'

const STORAGE_KEY = 'theme'

function systemPreference(): Theme {
  return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark'
}

// Module-level singleton — every component importing this shares one reactive value,
// kept in sync with the `data-theme` attribute that index.html's inline script
// already set before Vue mounted (avoids a flash of the wrong theme).
const theme = ref<Theme>(
  (document.documentElement.dataset.theme as Theme | undefined) ?? systemPreference(),
)

function apply(value: Theme) {
  theme.value = value
  document.documentElement.dataset.theme = value
  localStorage.setItem(STORAGE_KEY, value)
}

export function useTheme() {
  function setTheme(value: Theme) {
    apply(value)
  }

  function toggleTheme() {
    apply(theme.value === 'dark' ? 'light' : 'dark')
  }

  return { theme, setTheme, toggleTheme }
}
