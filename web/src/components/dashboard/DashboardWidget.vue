<template>
  <div
    class="bg-white overflow-hidden shadow rounded-lg"
    :aria-busy="loading"
    role="region"
    :aria-label="title"
  >
    <!-- Widget Header -->
    <div class="px-4 py-5 sm:px-6 border-b border-gray-200">
      <div class="flex items-center justify-between">
        <h3 class="text-lg leading-6 font-medium text-gray-900">
          {{ title }}
        </h3>
        <div v-if="$slots.actions" class="flex items-center space-x-2">
          <slot name="actions" />
        </div>
      </div>
    </div>

    <!-- Widget Content -->
    <div class="px-4 py-5 sm:p-6">
      <!-- Loading State -->
      <div v-if="loading" class="animate-pulse space-y-4">
        <div class="h-4 bg-gray-200 rounded w-3/4"></div>
        <div class="h-4 bg-gray-200 rounded w-1/2"></div>
        <div class="h-4 bg-gray-200 rounded w-5/6"></div>
        <div class="h-32 bg-gray-200 rounded"></div>
      </div>

      <!-- Error State -->
      <div
        v-else-if="error"
        class="text-center py-8"
        role="alert"
        aria-live="polite"
      >
        <div class="flex flex-col items-center space-y-4">
          <svg
            class="w-12 h-12 text-red-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <div class="space-y-2">
            <p class="text-sm font-medium text-gray-900">
              Failed to load data
            </p>
            <p class="text-sm text-gray-500">
              {{ error }}
            </p>
          </div>
          <button
            @click="$emit('retry')"
            class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
          >
            <svg
              class="w-4 h-4 mr-2"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            Retry
          </button>
        </div>
      </div>

      <!-- Empty State -->
      <div
        v-else-if="empty"
        class="text-center py-8"
        role="status"
        aria-live="polite"
      >
        <div class="flex flex-col items-center space-y-3">
          <svg
            class="w-12 h-12 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
            />
          </svg>
          <p class="text-sm text-gray-500">
            {{ emptyMessage }}
          </p>
        </div>
      </div>

      <!-- Content -->
      <div v-else>
        <slot />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title: string
  loading?: boolean
  error?: string | null
  empty?: boolean
  emptyMessage?: string
}

withDefaults(defineProps<Props>(), {
  loading: false,
  error: null,
  empty: false,
  emptyMessage: 'No data available'
})

defineEmits<{
  retry: []
}>()
</script>
