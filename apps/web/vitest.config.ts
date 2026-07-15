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
    passWithNoTests: true,
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['lcov'],
      reportsDirectory: 'coverage',
      // Vitest 4 defaults to reporting only files a test actually imported
      // (previously all matched source files were included, untouched ones
      // at 0%) — without this, files nothing imports vanish from the report
      // instead of counting against coverage, and files that happen to get
      // pulled in transitively (e.g. a layout component loaded by a router
      // test but never mounted) show up part-covered, both shifting the
      // overall percentage in ways unrelated to actual test coverage.
      include: ['src/**/*.ts', 'src/**/*.vue'],
      exclude: ['vite**.ts', 'dist/**', 'src/blog/**'],
    },
  },
})
