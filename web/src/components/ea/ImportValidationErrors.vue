<template>
  <div class="space-y-4">
    <!-- Error Summary Card -->
    <div v-if="hasData && errorCount > 0" class="bg-red-50 border border-red-200 rounded-lg p-4">
      <div class="flex items-start">
        <div class="flex-shrink-0">
          <svg class="h-6 w-6 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="ml-3 flex-1">
          <h3 class="text-lg font-medium text-red-800">
            {{ errorCount }} Error{{ errorCount > 1 ? 's' : '' }} Found
          </h3>
          <p class="mt-1 text-sm text-red-700">
            Please fix the errors below before importing. You can download the error report for offline review.
          </p>
        </div>
      </div>
    </div>

    <!-- Success Card -->
    <div v-if="hasData && errorCount === 0" class="bg-green-50 border border-green-200 rounded-lg p-4">
      <div class="flex items-start">
        <div class="flex-shrink-0">
          <svg class="h-6 w-6 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="ml-3 flex-1">
          <h3 class="text-lg font-medium text-green-800">
            All Rows Valid!
          </h3>
          <p class="mt-1 text-sm text-green-700">
            Your CSV data is ready to import. Click the Import button to proceed.
          </p>
        </div>
      </div>
    </div>

    <!-- Tabs -->
    <div v-if="hasData && errorCount > 0" class="border border-gray-300 rounded-lg">
      <div class="border-b border-gray-300">
        <nav class="flex -mb-px">
          <button
            @click="activeTab = 'table'"
            :class="[
              activeTab === 'table'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
              'w-1/2 py-3 px-4 text-center border-b-2 font-medium text-sm transition-colors'
            ]"
          >
            <span class="flex items-center justify-center gap-2">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M3 14h18m-9-4v8m-7 0h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
              Error Table
            </span>
          </button>
          <button
            @click="activeTab = 'download'"
            :class="[
              activeTab === 'download'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
              'w-1/2 py-3 px-4 text-center border-b-2 font-medium text-sm transition-colors'
            ]"
          >
            <span class="flex items-center justify-center gap-2">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Download CSV
            </span>
          </button>
        </nav>
      </div>

      <!-- Error Table Tab -->
      <div v-show="activeTab === 'table'" class="p-4">
        <div class="overflow-x-auto max-h-96">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50 sticky top-0">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Row
                </th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Field
                </th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Error
                </th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Expected
                </th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actual Value
                </th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr
                v-for="(error, idx) in errors"
                :key="idx"
                class="hover:bg-red-50 transition-colors"
              >
                <td class="px-4 py-3 whitespace-nowrap text-sm font-medium text-gray-900">
                  {{ error.row_number }}
                </td>
                <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
                  {{ error.field_name }}
                </td>
                <td class="px-4 py-3 text-sm text-red-600">
                  {{ error.error_message }}
                </td>
                <td class="px-4 py-3 text-sm text-gray-500">
                  {{ error.expected_format || '-' }}
                </td>
                <td class="px-4 py-3 text-sm text-gray-900">
                  {{ error.actual_value || '-' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Download CSV Tab -->
      <div v-show="activeTab === 'download'" class="p-4">
        <div class="text-center">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <h3 class="mt-4 text-lg font-medium text-gray-900">Download Error Report</h3>
          <p class="mt-2 text-sm text-gray-500">
            Export all validation errors to a CSV file for offline review and correction.
          </p>
          <button
            @click="$emit('download-errors')"
            class="mt-4 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            <svg class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            Download Error CSV
          </button>
        </div>
      </div>
    </div>

    <!-- Retry Button -->
    <div v-if="errorCount > 0" class="flex justify-end">
      <button
        @click="$emit('retry')"
        class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
      >
        <svg class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        Fix Errors & Retry
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { ImportError } from '@/services/importApi'

interface Props {
  errors: ImportError[]
  errorCount: number
  hasData: boolean
}

defineProps<Props>()

defineEmits<{
  'download-errors': []
  retry: []
}>()

const activeTab = ref<'table' | 'download'>('table')
</script>
