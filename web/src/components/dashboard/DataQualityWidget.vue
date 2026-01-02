<template>
  <DashboardWidget
    title="Data Quality Metrics"
    :loading="loading"
    :error="error"
    @retry="fetchDataQualityMetrics"
  >
    <div v-if="metrics" class="space-y-6">
      <!-- Quality Issues with Percentages -->
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <!-- Missing Attributes -->
        <QualityIssueCard
          title="Missing Attributes"
          :count="metrics.missing_attributes.count"
          :percentage="metrics.missing_attributes.percentage"
          severity="high"
          icon-class="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
        />

        <!-- Orphaned CIs -->
        <QualityIssueCard
          title="Orphaned CIs"
          :count="metrics.orphaned_cis.count"
          :percentage="metrics.orphaned_cis.percentage"
          severity="medium"
          icon-class="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
        />

        <!-- No Lifecycle Status -->
        <QualityIssueCard
          title="No Lifecycle Status"
          :count="metrics.no_lifecycle_status.count"
          :percentage="metrics.no_lifecycle_status.percentage"
          severity="high"
          icon-class="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
        />

        <!-- No Tags -->
        <QualityIssueCard
          title="No Tags"
          :count="metrics.no_tags.count"
          :percentage="metrics.no_tags.percentage"
          severity="low"
          icon-class="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
        />
      </div>

      <!-- Stale Data Section -->
      <div>
        <h4 class="text-sm font-medium text-gray-700 mb-3">Stale Data</h4>
        <div class="grid grid-cols-3 gap-4">
          <StaleDataCard
            title="30 Days"
            :count="metrics.stale_30_days"
            severity="medium"
          />
          <StaleDataCard
            title="60 Days"
            :count="metrics.stale_60_days"
            severity="high"
          />
          <StaleDataCard
            title="90 Days"
            :count="metrics.stale_90_days"
            severity="critical"
          />
        </div>
      </div>

      <!-- Duplicates Section -->
      <div
        v-if="metrics.duplicates > 0"
        class="bg-red-50 border border-red-200 rounded-lg p-4"
      >
        <div class="flex items-center">
          <svg
            class="w-6 h-6 text-red-600 mr-3"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
          <div class="flex-1">
            <p class="text-sm font-medium text-red-900">
              {{ metrics.duplicates }} Duplicate{{ metrics.duplicates > 1 ? 's' : '' }} Found
            </p>
            <p class="text-xs text-red-700 mt-1">
              Configuration items with duplicate names detected
            </p>
          </div>
        </div>
      </div>

      <!-- Overall Health Summary -->
      <div class="border-t border-gray-200 pt-4">
        <div class="flex items-center justify-between">
          <span class="text-sm text-gray-600">Overall Data Quality</span>
          <span
            class="text-sm font-semibold px-3 py-1 rounded-full"
            :class="overallQualityClass"
          >
            {{ overallQualityText }}
          </span>
        </div>
      </div>
    </div>
  </DashboardWidget>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/services/api'
import DashboardWidget from './DashboardWidget.vue'
import QualityIssueCard from './QualityIssueCard.vue'
import StaleDataCard from './StaleDataCard.vue'

interface QualityIssue {
  count: number
  percentage: number
}

interface DataQualityMetrics {
  missing_attributes: QualityIssue
  orphaned_cis: QualityIssue
  no_lifecycle_status: QualityIssue
  no_tags: QualityIssue
  stale_30_days: number
  stale_60_days: number
  stale_90_days: number
  duplicates: number
}

const loading = ref(true)
const error = ref<string | null>(null)
const metrics = ref<DataQualityMetrics | null>(null)

// Calculate overall quality score based on all metrics
const overallQualityScore = computed(() => {
  if (!metrics.value) return 0

  // Count issues (each quality issue counts once, stale data counts highest severity)
  let totalIssues = 0
  totalIssues += metrics.value.missing_attributes.count > 0 ? 1 : 0
  totalIssues += metrics.value.orphaned_cis.count > 0 ? 1 : 0
  totalIssues += metrics.value.no_lifecycle_status.count > 0 ? 1 : 0
  totalIssues += metrics.value.no_tags.count > 0 ? 1 : 0
  totalIssues += metrics.value.stale_90_days > 0 ? 1 : 0
  totalIssues += metrics.value.duplicates > 0 ? 1 : 0

  // Maximum possible issues is 6
  const maxIssues = 6
  const qualityPercentage = ((maxIssues - totalIssues) / maxIssues) * 100

  return qualityPercentage
})

const overallQualityText = computed(() => {
  const score = overallQualityScore.value
  if (score >= 90) return 'Excellent'
  if (score >= 70) return 'Good'
  if (score >= 50) return 'Fair'
  return 'Poor'
})

const overallQualityClass = computed(() => {
  const score = overallQualityScore.value
  if (score >= 90) return 'bg-green-100 text-green-800'
  if (score >= 70) return 'bg-blue-100 text-blue-800'
  if (score >= 50) return 'bg-yellow-100 text-yellow-800'
  return 'bg-red-100 text-red-800'
})

const fetchDataQualityMetrics = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await api.get<DataQualityMetrics>('/dashboard/data-quality')
    metrics.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'An error occurred'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchDataQualityMetrics()
})
</script>

<style scoped>
/* Add any component-specific styles here */
</style>
