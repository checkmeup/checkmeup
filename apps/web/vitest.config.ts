import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'node:path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: { '@': resolve(__dirname, './src') },
  },
  test: {
    environment: 'happy-dom',
    // Stryker copies the whole app into .stryker-tmp/sandbox-*/ per worker. Without
    // this exclude, a leftover sandbox makes Vitest collect every test 5x over
    // (683 -> 3422 seen in practice) and breaks Stryker's own dry run.
    exclude: ['**/node_modules/**', '**/dist/**', '.stryker-tmp/**'],
    passWithNoTests: true,
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['lcov'],
      reportsDirectory: 'coverage',
      exclude: ['vite**.ts', 'dist/**', 'src/blog/**'],
    },
  },
})
