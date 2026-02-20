<template>
  <div class="bg-white rounded-lg shadow">
    <!-- Header -->
    <div class="px-4 py-3 border-b border-gray-200">
      <h3 class="text-lg font-medium text-gray-900">CSV Data Preview</h3>
      <p class="mt-1 text-sm text-gray-500">
        Showing first {{ previewData.length }} rows of {{ totalRows }} total rows
      </p>
    </div>

    <!-- Table -->
    <div class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th
              v-for="column in columns"
              :key="column"
              class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky top-0 bg-gray-50"
            >
              {{ column }}
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr
            v-for="(row, idx) in previewData"
            :key="idx"
            :class="idx % 2 === 0 ? 'bg-white' : 'bg-gray-50'"
            class="hover:bg-blue-50 transition-colors"
          >
            <td
              v-for="column in columns"
              :key="column"
              class="px-4 py-3 whitespace-nowrap text-sm text-gray-900"
            >
              {{ getCellValue(row, column) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ImportRow } from '@/stores/eaImport'

interface Props {
  data: ImportRow[]
  columns: string[]
}

const props = defineProps<Props>()

const previewData = computed(() => {
  return props.data.slice(0, 10) // Show first 10 rows
})

const totalRows = computed(() => {
  return props.data.length
})

function getCellValue(row: ImportRow, column: string): string {
  const value = (row as any)[column]
  return value !== null && value !== undefined ? String(value) : ''
}
</script>
