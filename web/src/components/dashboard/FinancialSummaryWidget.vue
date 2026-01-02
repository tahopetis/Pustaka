<template>
  <DashboardWidget
    title="Financial Summary"
    :loading="loading"
    :error="error"
    @retry="fetchFinancialSummary"
  >
    <div v-if="summary" class="space-y-6">
      <!-- Summary Cards -->
      <div class="grid grid-cols-2 gap-4">
        <div class="bg-green-50 border border-green-200 rounded-lg p-4">
          <div class="text-sm text-green-700 font-medium mb-1">Total Asset Value</div>
          <div class="text-2xl font-bold text-green-900">
            {{ formatCurrency(summary.total_gross_book_value) }}
          </div>
          <div class="text-xs text-green-600 mt-1">
            {{ summary.total_cis }} assets
          </div>
        </div>

        <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div class="text-sm text-blue-700 font-medium mb-1">Net Book Value</div>
          <div class="text-2xl font-bold text-blue-900">
            {{ formatCurrency(summary.total_net_book_value) }}
          </div>
          <div class="text-xs text-blue-600 mt-1">
            After depreciation
          </div>
        </div>

        <div class="bg-orange-50 border border-orange-200 rounded-lg p-4">
          <div class="text-sm text-orange-700 font-medium mb-1">Accumulated Depreciation</div>
          <div class="text-2xl font-bold text-orange-900">
            {{ formatCurrency(summary.total_accumulated_depreciation) }}
          </div>
          <div class="text-xs text-orange-600 mt-1">
            Total to date
          </div>
        </div>

        <div class="bg-purple-50 border border-purple-200 rounded-lg p-4">
          <div class="text-sm text-purple-700 font-medium mb-1">Monthly Depreciation</div>
          <div class="text-2xl font-bold text-purple-900">
            {{ formatCurrency(summary.total_monthly_depreciation) }}
          </div>
          <div class="text-xs text-purple-600 mt-1">
            Per month
          </div>
        </div>
      </div>

      <!-- Value by CI Type -->
      <div v-if="summary.groups && summary.groups.length > 0">
        <h4 class="text-sm font-medium text-gray-700 mb-3">Asset Value by Type</h4>
        <div class="space-y-3">
          <div
            v-for="group in sortedGroups"
            :key="group.group_name"
            class="bg-white border border-gray-200 rounded-lg p-4"
          >
            <div class="flex items-center justify-between mb-2">
              <div class="text-sm font-medium text-gray-900">
                {{ group.group_name }}
              </div>
              <div class="text-sm text-gray-600">
                {{ formatCurrency(group.total_book_value) }}
              </div>
            </div>
            <div class="w-full bg-gray-200 rounded-full h-2">
              <div
                class="h-2 rounded-full transition-all duration-500 bg-blue-500"
                :style="{ width: `${calculatePercentage(group.total_book_value)}%` }"
              />
            </div>
            <div class="flex items-center justify-between mt-1 text-xs text-gray-500">
              <span>{{ group.ci_count }} assets</span>
              <span>{{ calculatePercentage(group.total_book_value).toFixed(1) }}% of total</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Export Button -->
      <div class="flex justify-end pt-4 border-t border-gray-200">
        <button
          @click="exportToExcel"
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          Export to Excel
        </button>
      </div>
    </div>

    <!-- Empty State -->
    <div
      v-else-if="!loading && !error"
      class="text-center py-8 bg-gray-50 border border-gray-200 rounded-lg"
    >
      <svg class="w-12 h-12 mx-auto text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
      </svg>
      <p class="text-sm font-medium text-gray-900">No financial data available</p>
      <p class="text-xs text-gray-600 mt-1">
        Add amortizable assets to see financial summaries
      </p>
    </div>
  </DashboardWidget>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/services/api'
import DashboardWidget from './DashboardWidget.vue'

interface AmortizationGroup {
  group_name: string
  group_id?: string
  ci_count: number
  total_book_value: number
}

interface AmortizationSummary {
  group_by: string
  groups: AmortizationGroup[]
  total_cis: number
  total_gross_book_value: number
  total_net_book_value: number
  total_accumulated_depreciation: number
  total_salvage_value: number
  total_monthly_depreciation: number
  generated_at: string
}

const loading = ref(true)
const error = ref<string | null>(null)
const summary = ref<AmortizationSummary | null>(null)

const sortedGroups = computed(() => {
  if (!summary.value?.groups) return []
  return [...summary.value.groups].sort((a, b) => b.total_book_value - a.total_book_value)
})

const formatCurrency = (value: number) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(value)
}

const calculatePercentage = (value: number) => {
  if (!summary.value || summary.value.total_net_book_value === 0) return 0
  return (value / summary.value.total_net_book_value) * 100
}

const exportToExcel = () => {
  // TODO: Implement Excel export
  alert('Excel export functionality coming soon!')
}

const fetchFinancialSummary = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await api.get<AmortizationSummary>('/amortization/summaries', {
      params: {
        group_by: 'ci_type'
      }
    })
    summary.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'An error occurred'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchFinancialSummary()
})
</script>
