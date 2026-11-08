import { defineConfig, configDefaults } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import { chunkNameFor } from './config/chunks'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
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
        // Vite 8 builds with Rolldown, which only accepts the function form.
        // The classifier lives in config/chunks.ts and is unit-tested.
        manualChunks(id) {
          return chunkNameFor(id)
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