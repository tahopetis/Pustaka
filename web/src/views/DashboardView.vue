<template>
  <div class="page-container page-content">
    <!-- Page header -->
    <div class="page-header flex justify-between items-center mb-5">
      <div>
        <h1 class="page-title">Dashboard</h1>
        <p class="page-subtitle">Welcome back, {{ user?.username }}!</p>
      </div>

      <!-- Global Search -->
      <GlobalSearch class="mx-4" />

      <!-- Refresh button -->
      <button
        @click="refreshAllData"
        :disabled="isLoading"
        class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
      >
        <svg
          class="w-4 h-4 mr-2"
          :class="{ 'animate-spin': isLoading }"
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

    <!-- Time Range Filter -->
    <div class="mb-4">
      <TimeRangeFilter @change="handleTimeRangeChange" />
    </div>

    <!-- Stats cards -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 mb-5">
      <DashboardWidget
        title="Total CIs"
        :loading="loadingState.stats"
        :error="errorState.stats"
        @retry="fetchAllData"
      >
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-7 h-7 bg-blue-500 rounded-md flex items-center justify-center">
              <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
              </svg>
            </div>
          </div>
          <div class="ml-5 w-0 flex-1">
            <dl>
              <dt class="text-sm font-medium text-gray-500 truncate">Configuration Items</dt>
              <dd class="text-2xl font-bold text-gray-900">{{ stats?.total_cis || 0 }}</dd>
            </dl>
          </div>
        </div>
      </DashboardWidget>

      <DashboardWidget
        title="CI Types"
        :loading="loadingState.stats"
        :error="errorState.stats"
        @retry="fetchAllData"
      >
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-7 h-7 bg-green-500 rounded-md flex items-center justify-center">
              <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"></path>
              </svg>
            </div>
          </div>
          <div class="ml-5 w-0 flex-1">
            <dl>
              <dt class="text-sm font-medium text-gray-500 truncate">Type Definitions</dt>
              <dd class="text-2xl font-bold text-gray-900">{{ stats?.total_ci_types || 0 }}</dd>
            </dl>
          </div>
        </div>
      </DashboardWidget>

      <DashboardWidget
        title="Relationships"
        :loading="loadingState.stats"
        :error="errorState.stats"
        @retry="fetchAllData"
      >
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-7 h-7 bg-purple-500 rounded-md flex items-center justify-center">
              <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"></path>
              </svg>
            </div>
          </div>
          <div class="ml-5 w-0 flex-1">
            <dl>
              <dt class="text-sm font-medium text-gray-500 truncate">Total Connections</dt>
              <dd class="text-2xl font-bold text-gray-900">{{ stats?.total_relationships || 0 }}</dd>
            </dl>
          </div>
        </div>
      </DashboardWidget>

      <DashboardWidget
        title="Users"
        :loading="loadingState.stats"
        :error="errorState.stats"
        @retry="fetchAllData"
      >
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-7 h-7 bg-yellow-500 rounded-md flex items-center justify-center">
              <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path>
              </svg>
            </div>
          </div>
          <div class="ml-5 w-0 flex-1">
            <dl>
              <dt class="text-sm font-medium text-gray-500 truncate">Active Users</dt>
              <dd class="text-2xl font-bold text-gray-900">{{ stats?.total_users || 0 }}</dd>
            </dl>
          </div>
        </div>
      </DashboardWidget>

      <!-- Amortization Widget -->
      <DashboardWidget
        v-if="hasAnyPermission(['amortization:read'])"
        title="Amortizable Assets"
        :loading="loadingState.amortizationMetrics"
        :error="errorState.amortizationMetrics"
        @retry="loadAmortizationMetrics"
      >
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="w-7 h-7 bg-indigo-500 rounded-md flex items-center justify-center">
              <svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
            </div>
          </div>
          <div class="ml-5 w-0 flex-1">
            <dl>
              <dt class="text-sm font-medium text-gray-500 truncate">Amortizable Assets</dt>
              <dd class="text-2xl font-bold text-gray-900">{{ amortizationMetrics?.total_amortizable_assets || 0 }}</dd>
            </dl>
          </div>
        </div>
        <div class="mt-3 pt-3 border-t border-gray-100">
          <div class="flex items-center text-sm">
            <span class="text-gray-500">Monthly Depreciation:</span>
            <span class="ml-2 font-medium text-gray-900">
              {{ formatCurrency(amortizationMetrics?.monthly_depreciation || 0) }}
            </span>
          </div>
        </div>
      </DashboardWidget>

      <!-- CI Status Distribution Widget -->
      <DashboardWidget
        title="CI Status Distribution"
        :loading="loadingState.ciStatusDistribution"
        :error="errorState.ciStatusDistribution"
        @retry="loadCIStatusDistribution"
      >
        <div class="px-2 py-3">
          <div v-if="ciStatusDistribution && ciStatusDistribution.length > 0" class="space-y-2">
            <div
              v-for="status in ciStatusDistribution"
              :key="status.status_name"
              class="flex items-center justify-between"
            >
              <div class="flex items-center flex-1">
                <div
                  v-if="status.color"
                  class="w-3 h-3 rounded-full mr-2"
                  :style="{ backgroundColor: status.color }"
                ></div>
                <span class="text-sm font-medium text-gray-700">{{ status.display_name }}</span>
              </div>
              <div class="flex items-center space-x-4">
                <div class="text-right">
                  <span class="text-sm font-bold text-gray-900">{{ status.count }}</span>
                  <span class="text-xs text-gray-500">CIs</span>
                </div>
                <div class="w-24 bg-gray-200 rounded-full h-2">
                  <div
                    class="h-2 rounded-full"
                    :style="{
                      backgroundColor: status.color,
                      width: `${status.percentage}%`
                    }"
                  ></div>
                </div>
                <span class="text-xs text-gray-500 w-12 text-right">{{ status.percentage.toFixed(1) }}%</span>
              </div>
            </div>
            <div v-if="ciStatusDistribution.length === 0" class="text-center py-4 text-gray-500">
              <p class="text-sm">No CIs with lifecycle status assigned</p>
            </div>
          </div>
        </div>
      </DashboardWidget>
    </div>

    <!-- Health Score Widget -->
    <div class="mb-5">
      <HealthScoreWidget />
    </div>

    <!-- Data Quality Widget -->
    <div class="mb-5">
      <DataQualityWidget />
    </div>

    <!-- Asset Aging Widget -->
    <div class="mb-5">
      <AssetAgingWidget />
    </div>

    <!-- Financial Summary Widget -->
    <div class="mb-5">
      <FinancialSummaryWidget />
    </div>

    <!-- Risk Metrics Widget -->
    <div class="mb-5">
      <RiskMetricsWidget />
    </div>

    <!-- Analytics Charts - Custom Layout -->
    <div class="space-y-4 mb-5">
      <!-- Activity Trend and Heatmap - Side by Side -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <!-- Activity Trend Chart -->
        <DashboardWidget
          title="Activity Trend"
          :loading="loadingState.auditStats"
          :error="errorState.auditStats"
          :empty="!hasActivityData"
          @retry="retryDataSource('auditStats')"
        >
          <div class="px-2 py-3">
            <TrendChart
              v-if="hasActivityData"
              :data="activityTrendData"
              title=""
              :height="280"
              x-axis-label="Date"
              y-axis-label="Events"
            />
          </div>
        </DashboardWidget>

        <!-- Activity Heatmap -->
        <DashboardWidget
          title="Activity Heatmap"
          :loading="loadingState.auditStats"
          :error="errorState.auditStats"
          :empty="!hasActivityData"
          @retry="retryDataSource('auditStats')"
        >
          <div class="px-3 py-4">
            <ActivityHeatmap
              v-if="hasActivityData"
              :data="auditStats?.daily_activity || {}"
              title=""
              :weeks="12"
              @day-click="handleDayClick"
            />
          </div>
        </DashboardWidget>
      </div>

      <!-- CI Type Distribution - Full Width -->
      <div class="w-full">
        <DashboardWidget
          title="CI Type Distribution"
          :loading="loadingState.ciTypeUsage"
          :error="errorState.ciTypeUsage"
          :empty="!ciTypeUsage || ciTypeUsage.length === 0"
          @retry="retryDataSource('ciTypeUsage')"
        >
          <div class="flex items-center justify-center px-3 py-4">
            <DonutChart
              v-if="ciTypeUsage && ciTypeUsage.length > 0"
              :data="ciTypeDistributionData"
              title=""
              :size="280"
            />
          </div>
        </DashboardWidget>
      </div>

      <!-- Most Connected CIs and Relationship Type Distribution - Side by Side -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <!-- Most Connected CIs -->
        <DashboardWidget
          title="Most Connected CIs"
          :loading="loadingState.mostConnected"
          :error="errorState.mostConnected"
          :empty="!mostConnected || mostConnected.length === 0"
          @retry="retryDataSource('mostConnected')"
        >
          <div class="px-2 py-2">
            <NetworkAnalyticsCard
              v-if="mostConnected && mostConnected.length > 0"
              :data="mostConnected"
              title=""
              :limit="10"
              @item-click="handleCIClick"
            />
          </div>
        </DashboardWidget>

        <!-- Relationship Type Distribution -->
        <DashboardWidget
          title="Relationship Type Distribution"
          :loading="loadingState.relationshipTypeUsage"
          :error="errorState.relationshipTypeUsage"
          :empty="!relationshipTypeUsage || relationshipTypeUsage.length === 0"
          @retry="retryDataSource('relationshipTypeUsage')"
        >
          <div class="flex items-center justify-center px-3 py-4">
            <DonutChart
              v-if="relationshipTypeUsage && relationshipTypeUsage.length > 0"
              :data="relationshipTypeDistributionData"
              title=""
              :size="280"
            />
          </div>
        </DashboardWidget>
      </div>
    </div>

    <!-- Quick actions -->
    <div v-if="hasCreatePermissions" class="bg-white shadow-sm rounded-lg mb-5">
      <div class="px-6 py-5 sm:px-8">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Quick Actions</h3>
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <router-link
            v-if="hasPermission('ci:create')"
            to="/ci/new"
            class="inline-flex items-center justify-center px-5 py-3 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors duration-150"
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
            </svg>
            New Configuration Item
          </router-link>

          <router-link
            v-if="hasPermission('ci_type:create')"
            to="/ci-types/new"
            class="inline-flex items-center justify-center px-5 py-3 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors duration-150"
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
            </svg>
            New CI Type
          </router-link>

          <router-link
            v-if="hasPermission('user:create')"
            to="/users/new"
            class="inline-flex items-center justify-center px-5 py-3 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors duration-150"
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"></path>
            </svg>
            New User
          </router-link>
        </div>
      </div>
    </div>

    <!-- Recent activity -->
    <div class="bg-white shadow-sm rounded-lg">
      <div class="px-6 py-5 sm:px-8">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Recent Activity</h3>
        <div class="space-y-3">
          <div v-if="loadingState.auditStats" class="text-gray-500 text-center py-10">
            <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mb-3"></div>
            <p>Loading recent activity...</p>
          </div>
          <div v-else-if="!auditStats || !auditStats.recent_activity || auditStats.recent_activity.length === 0" class="text-gray-500 text-center py-10">
            <svg class="mx-auto h-12 w-12 text-gray-400 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"></path>
            </svg>
            <p>No recent activity to display.</p>
          </div>
          <div v-else v-for="activity in auditStats.recent_activity.slice(0, 10)" :key="activity.id" class="flex items-center space-x-4 py-2 border-b border-gray-100 last:border-0">
            <div class="flex-shrink-0">
              <div class="w-9 h-9 bg-gradient-to-br from-blue-50 to-blue-100 rounded-full flex items-center justify-center">
                <svg class="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
              </div>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm text-gray-900">
                <span class="font-semibold text-gray-800">{{ activity.username || 'Unknown' }}</span>
                <span class="text-gray-600 mx-1">{{ activity.action }}</span>
                <span class="text-gray-700">{{ activity.entity_type }}</span>
                <span class="text-gray-500 ml-1 font-mono text-xs">{{ activity.entity_id?.substring(0, 8) }}</span>
              </p>
              <p class="text-xs text-gray-500 mt-0.5">{{ formatDate(activity.timestamp) }}</p>
            </div>
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
import { useLifecycleStatusStore } from '@/stores/lifecycleStatus'
import { useAmortizationStore } from '@/stores/amortization'
import { useDashboardData } from '@/composables/useDashboardData'
import TimeRangeFilter from '@/components/dashboard/TimeRangeFilter.vue'
import DashboardWidget from '@/components/dashboard/DashboardWidget.vue'
import HealthScoreWidget from '@/components/dashboard/HealthScoreWidget.vue'
import DataQualityWidget from '@/components/dashboard/DataQualityWidget.vue'
import AssetAgingWidget from '@/components/dashboard/AssetAgingWidget.vue'
import FinancialSummaryWidget from '@/components/dashboard/FinancialSummaryWidget.vue'
import RiskMetricsWidget from '@/components/dashboard/RiskMetricsWidget.vue'
import GlobalSearch from '@/components/dashboard/GlobalSearch.vue'
import TrendChart from '@/components/dashboard/TrendChart.vue'
import DonutChart from '@/components/dashboard/DonutChart.vue'
import NetworkAnalyticsCard from '@/components/dashboard/NetworkAnalyticsCard.vue'
import ActivityHeatmap from '@/components/dashboard/ActivityHeatmap.vue'
import type { TimeRange } from '@/types/dashboard'
import type { MostConnectedCI } from '@/types/dashboard'

const router = useRouter()
const authStore = useAuthStore()
const lifecycleStatusStore = useLifecycleStatusStore()
const amortizationStore = useAmortizationStore()

const user = computed(() => authStore.user)
const ciStatusDistribution = ref<Array<any>>([])
const amortizationMetrics = ref<any>(null)

// Use dashboard data composable
const {
  stats,
  auditStats,
  ciTypeUsage,
  mostConnected,
  relationshipTypeUsage,
  loading: loadingState,
  errors: errorState,
  isLoading,
  fetchAllData,
  refreshData,
  updateTimeRange,
  retryDataSource,
} = useDashboardData()

// Add CI status distribution to loading states
loadingState.value.ciStatusDistribution = ref(false)
errorState.value.ciStatusDistribution = ref(null)

// Add amortization metrics to loading states
loadingState.value.amortizationMetrics = ref(false)
errorState.value.amortizationMetrics = ref(null)

const hasPermission = (permission: string) => {
  return authStore.hasPermission(permission)
}

const hasAnyPermission = (permissions: string[]) => {
  return authStore.hasAnyPermission(permissions)
}

const formatCurrency = (amount: number): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD'
  }).format(amount)
}

const loadCIStatusDistribution = async () => {
  loadingState.value.ciStatusDistribution = true
  try {
    const distribution = await lifecycleStatusStore.getCIStatusDistribution()
    ciStatusDistribution.value = distribution
    errorState.value.ciStatusDistribution = null
  } catch (error) {
    console.error('Failed to load CI status distribution:', error)
    ciStatusDistribution.value = []
    errorState.value.ciStatusDistribution = error
  } finally {
    loadingState.value.ciStatusDistribution = false
  }
}

const loadAmortizationMetrics = async () => {
  loadingState.value.amortizationMetrics = true
  try {
    const response = await amortizationStore.loadMetrics()
    amortizationMetrics.value = response.data
    errorState.value.amortizationMetrics = null
  } catch (error) {
    console.error('Failed to load amortization metrics:', error)
    amortizationMetrics.value = null
    errorState.value.amortizationMetrics = error
  } finally {
    loadingState.value.amortizationMetrics = false
  }
}

const hasCreatePermissions = computed(() => {
  return authStore.hasAnyPermission(['ci:create', 'ci_type:create', 'user:create'])
})

const formatDate = (timestamp: string) => {
  return new Date(timestamp).toLocaleString()
}

// Chart data transformations
const hasActivityData = computed(() => {
  return auditStats.value?.daily_activity && Object.keys(auditStats.value.daily_activity).length > 0
})

const activityTrendData = computed(() => {
  if (!hasActivityData.value) return []

  const dailyActivity = auditStats.value!.daily_activity
  const data = Object.entries(dailyActivity)
    .map(([date, count]) => ({
      x: date,
      y: count as number,
      label: date,
    }))
    .sort((a, b) => a.x.localeCompare(b.x))

  return [
    {
      id: 'activity',
      name: 'Events',
      color: '#3B82F6',
      data,
    },
  ]
})

const ciTypeDistributionData = computed(() => {
  if (!ciTypeUsage.value || ciTypeUsage.value.length === 0) return []

  return ciTypeUsage.value.map((item, index) => ({
    label: item.type,
    value: item.count,
    color: getColor(index),
  }))
})

const relationshipTypeDistributionData = computed(() => {
  if (!relationshipTypeUsage.value || relationshipTypeUsage.value.length === 0) return []

  return relationshipTypeUsage.value.map((item, index) => ({
    label: item.relationship_type.name,
    value: item.usage_count,
    color: getColor(index),
  }))
})

function getColor(index: number): string {
  const colors = [
    '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6',
    '#EC4899', '#14B8A6', '#F97316', '#6366F1', '#84CC16',
  ]
  return colors[index % colors.length]
}

// Event handlers
const handleTimeRangeChange = async (range: TimeRange) => {
  await updateTimeRange(range)
}

const refreshAllData = async () => {
  await Promise.all([
    refreshData(),
    loadCIStatusDistribution()
  ])
}

const handleCIClick = (ci: MostConnectedCI) => {
  router.push(`/ci/${ci.id}`)
}

const handleDayClick = (date: string, count: number) => {
  console.log(`Activity on ${date}: ${count} events`)
  // Could navigate to audit log filtered by this date
}

// Load data on mount
onMounted(async () => {
  await Promise.all([
    fetchAllData(),
    loadCIStatusDistribution(),
    loadAmortizationMetrics()
  ])
})
</script>
