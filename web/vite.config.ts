import { defineConfig, configDefaults } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      output: {
        // Vite 8 builds with Rolldown, which only accepts the function form
        // (the object/array shorthand was removed).
        manualChunks(id) {
          if (!id.includes('node_modules')) return
          if (/node_modules\/(vue|@vue\/|vue-router|pinia)\//.test(id)) return 'vendor'
          if (/node_modules\/(@headlessui\/vue|@heroicons\/vue)\//.test(id)) return 'ui'
          if (/node_modules\/(d3|d3-[^/]+|vis-network|vis-data|vis-util)\//.test(id)) return 'charts'
        },
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    // Playwright specs live under tests/ too; keep them out of Vitest discovery.
    exclude: [...configDefaults.exclude, 'tests/e2e/**', 'tests/**/*.spec.ts'],
  },
})