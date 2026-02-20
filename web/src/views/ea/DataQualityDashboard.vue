<template>
  <div class="page-container page-content">
    <!-- Page Header -->
    <div class="page-header flex justify-between items-center">
      <div>
        <!-- Breadcrumbs -->
        <nav class="flex text-sm text-gray-500 mb-2">
          <router-link to="/entities" class="hover:text-gray-700">
            EA
          </router-link>
          <span class="mx-2">/</span>
          <span class="text-gray-900">Data Quality</span>
        </nav>
        <h1 class="page-title">Data Quality Dashboard</h1>
        <p class="page-subtitle">Monitor EA entity data health and identify quality issues</p>
      </div>
      <button
        @click="handleRefresh"
        :disabled="loading"
        class="btn btn-outline flex items-center gap-2"
      >
        <svg
          class="w-4 h-4"
          :class="{ 'animate-spin': loading }"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
          />
        </svg>
        Refresh
      </button>
    </div>

    <!-- Error State -->
    <div
      v-if="error"
      class="bg-red-50 border border-red-200 rounded-lg p-4 mb-6 flex items-center justify-between"
    >
      <div class="flex items-center">
        <svg class="w-5 h-5 text-red-600 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span class="text-red-800">{{ error }}</span>
      </div>
      <button @click="handleRefresh" class="text-red-600 hover:text-red-800 font-medium text-sm">
        Retry
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading && !metrics" class="flex items-center justify-center h-64">
      <div class="flex flex-col items-center">
        <svg class="animate-spin h-8 w-8 text-blue-600 mb-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p class="text-gray-600">Loading metrics...</p>
      </div>
    </div>

    <!-- Dashboard Content -->
    <div v-else-if="metrics" class="space-y-6">
      <!-- Top Row: Metric Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <!-- Total Entities -->
        <QualityMetricCard
          title="Total Entities"
          :value="metrics.total_entities"
          icon="cube"
          :clickable="false"
        />

        <!-- Completeness % -->
        <QualityMetricCard
          title="Completeness"
          :value="metrics.completeness_pct"
          icon="check-circle"
          :clickable="false"
        />

        <!-- Stale Entities -->
        <QualityMetricCard
          title="Stale Entities"
          :value="metrics.stale_entities_count"
          icon="clock"
          :clickable="true"
          @click="handleMetricClick('stale')"
        />

        <!-- Entities with Errors -->
        <QualityMetricCard
          title="Entities with Errors"
          :value="metrics.entities_with_errors_count"
          icon="exclamation-triangle"
          :clickable="true"
          @click="handleMetricClick('errors')"
        />
      </div>

      <!-- Middle Row: Charts -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Lifecycle Status Breakdown -->
        <QualityChart
          title="Lifecycle Status Breakdown"
          :data="metrics.lifecycle_breakdown"
          type="donut"
        />

        <!-- Errors by Domain -->
        <QualityChart
          title="Errors by Domain"
          :data="metrics.error_breakdown_by_domain"
          type="donut"
        />
      </div>

      <!-- Bottom Row: Detailed Breakdown Tables (Optional) -->
      <div v-if="showDetails" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Recent Stale Entities Table -->
        <div class="bg-white shadow rounded-lg p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">
            Recent Stale Entities
          </h3>
          <div v-if="staleEntities && staleEntities.length > 0" class="overflow-hidden">
            <table class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Domain</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Quality Score</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                <tr
                  v-for="entity in staleEntities.slice(0, 10)"
                  :key="entity.id"
                  class="hover:bg-gray-50 cursor-pointer"
                  @click="navigateToEntity(entity)"
                >
                  <td class="px-4 py-3 text-sm text-gray-900">{{ entity.name }}</td>
                  <td class="px-4 py-3 text-sm text-gray-600">{{ entity.ea_domain }}</td>
                  <td class="px-4 py-3 text-sm">
                    <span
                      class="px-2 py-1 rounded text-xs font-medium"
                      :class="getQualityScoreClass(entity.data_quality_score)"
                    >
                      {{ entity.data_quality_score.toFixed(1) }}%
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
            <div class="mt-4 text-center">
              <router-link
                to="/entities?stale=true"
                class="text-blue-600 hover:text-blue-800 text-sm font-medium"
              >
                View all stale entities →
              </router-link>
            </div>
          </div>
          <div v-else class="text-center text-gray-500 py-8">
            No stale entities found
          </div>
        </div>

        <!-- Entities with Most Errors Table -->
        <div class="bg-white shadow rounded-lg p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">
            Entities with Most Errors
          </h3>
          <div v-if="entitiesWithErrors && entitiesWithErrors.length > 0" class="overflow-hidden">
            <table class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Domain</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Quality Score</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                <tr
                  v-for="entity in entitiesWithErrors.slice(0, 10)"
                  :key="entity.id"
                  class="hover:bg-gray-50 cursor-pointer"
                  @click="navigateToEntity(entity)"
                >
                  <td class="px-4 py-3 text-sm text-gray-900">{{ entity.name }}</td>
                  <td class="px-4 py-3 text-sm text-gray-600">{{ entity.ea_domain }}</td>
                  <td class="px-4 py-3 text-sm">
                    <span
                      class="px-2 py-1 rounded text-xs font-medium"
                      :class="getQualityScoreClass(entity.data_quality_score)"
                    >
                      {{ entity.data_quality_score.toFixed(1) }}%
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
            <div class="mt-4 text-center">
              <router-link
                to="/entities?errors=true"
                class="text-blue-600 hover:text-blue-800 text-sm font-medium"
              >
                View all entities with errors →
              </router-link>
            </div>
          </div>
          <div v-else class="text-center text-gray-500 py-8">
            No entities with errors found
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import QualityMetricCard from '@/components/ea/QualityMetricCard.vue'
import QualityChart from '@/components/ea/QualityChart.vue'
import dataQualityApi, { type DataQualityMetrics, type EAEntitySummary } from '@/services/dataQualityApi'

const router = useRouter()
const authStore = useAuthStore()

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

// State
const metrics = ref<DataQualityMetrics | null>(null)
const staleEntities = ref<EAEntitySummary[]>([])
const entitiesWithErrors = ref<EAEntitySummary[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const showDetails = ref(false)

// ============================================================================
// Methods
// ============================================================================

async function fetchMetrics() {
  loading.value = true
  error.value = null

  try {
    const [metricsData, staleData, errorsData] = await Promise.all([
      dataQualityApi.getMetrics(),
      dataQualityApi.getStaleEntities({ days_threshold: 90, include_incomplete: true }),
      dataQualityApi.getEntitiesWithErrors(),
    ])

    metrics.value = metricsData
    staleEntities.value = staleData.entities || []
    entitiesWithErrors.value = errorsData.entities || []

    // Show details if there's data
    showDetails.value = (staleEntities.value.length > 0 || entitiesWithErrors.value.length > 0)
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : 'Failed to load data quality metrics'
    error.value = message
    console.error('Error fetching metrics:', err)
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  fetchMetrics()
}

function handleMetricClick(type: string) {
  if (type === 'stale') {
    router.push('/entities?stale=true')
  } else if (type === 'errors') {
    router.push('/entities?errors=true')
  }
}

function navigateToEntity(entity: EAEntitySummary) {
  // Navigate to entity details
  router.push(`/entities/${entity.ea_domain}/${entity.id}`)
}

function getQualityScoreClass(score: number): string {
  if (score >= 80) return 'bg-green-100 text-green-800'
  if (score >= 60) return 'bg-yellow-100 text-yellow-800'
  return 'bg-red-100 text-red-800'
}

// ============================================================================
// Lifecycle
// ============================================================================

onMounted(() => {
  fetchMetrics()
})
</script>

<style scoped>
.page-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 1.5rem;
}

.page-content {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.page-header {
  margin-bottom: 2rem;
}

.page-title {
  font-size: 1.875rem;
  font-weight: 700;
  color: #111827;
  line-height: 1.2;
}

.page-subtitle {
  font-size: 1rem;
  color: #6b7280;
  margin-top: 0.25rem;
}

.btn {
  display: inline-flex;
  align-items: center;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
  cursor: pointer;
  border: 1px solid transparent;
}

.btn-primary {
  background-color: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #2563eb;
}

.btn-outline {
  background-color: white;
  border-color: #d1d5db;
  color: #374151;
}

.btn-outline:hover:not(:disabled) {
  background-color: #f9fafb;
  border-color: #9ca3af;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Responsive grid adjustments */
@media (max-width: 1024px) {
  .grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .grid {
    grid-template-columns: 1fr;
  }

  .page-title {
    font-size: 1.5rem;
  }
}
</style>
