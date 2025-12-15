/**
 * Dashboard Data Composable
 *
 * Centralized data fetching and state management for the dashboard view.
 * Handles parallel API calls, time-range filtering, loading states, and error handling.
 *
 * Features:
 * - Parallel data fetching from multiple endpoints using Promise.all()
 * - Individual error handling per endpoint (failures are isolated)
 * - Time-range filter application for audit and analytics data
 * - Reactive loading and error states per data source
 * - Manual refresh functionality
 *
 * @module useDashboardData
 */

import { ref, computed, type Ref } from 'vue'
import { api } from '@/services/api'
import type {
  TimeRange,
  DashboardStats,
  AuditStats,
  CITypeUsage,
  MostConnectedCI,
  RelationshipTypeUsage,
  DashboardLoadingState,
  DashboardErrorState,
} from '@/types/dashboard'

/**
 * Dashboard data composable
 *
 * @param initialTimeRange - Initial time range filter (defaults to last 30 days)
 * @returns Dashboard data state and methods
 */
export function useDashboardData(initialTimeRange?: TimeRange) {
  // ============================================================================
  // State - Reactive data stores
  // ============================================================================

  /**
   * Basic dashboard statistics (counts)
   */
  const stats: Ref<DashboardStats | null> = ref(null)

  /**
   * Audit statistics with time-series data
   */
  const auditStats: Ref<AuditStats | null> = ref(null)

  /**
   * CI type distribution data
   */
  const ciTypeUsage: Ref<CITypeUsage[]> = ref([])

  /**
   * Most connected CIs for network analytics
   */
  const mostConnected: Ref<MostConnectedCI[]> = ref([])

  /**
   * Relationship type usage statistics
   */
  const relationshipTypeUsage: Ref<RelationshipTypeUsage[]> = ref([])

  /**
   * Current time range filter
   */
  const timeRange: Ref<TimeRange> = ref(
    initialTimeRange || {
      startDate: getDefaultStartDate(),
      endDate: getDefaultEndDate(),
      preset: '30',
    }
  )

  /**
   * Loading states for each data source
   */
  const loading: Ref<DashboardLoadingState> = ref({
    stats: false,
    audit: false,
    ciTypes: false,
    network: false,
    relationshipTypes: false,
    growth: false,
  })

  /**
   * Error states for each data source
   */
  const errors: Ref<DashboardErrorState> = ref({
    stats: null,
    audit: null,
    ciTypes: null,
    network: null,
    relationshipTypes: null,
    growth: null,
  })

  /**
   * Timestamp of last successful data refresh
   */
  const lastRefresh: Ref<Date | null> = ref(null)

  // ============================================================================
  // Computed Properties
  // ============================================================================

  /**
   * Whether any data source is currently loading
   */
  const isLoading = computed(() => {
    return Object.values(loading.value).some((state) => state === true)
  })

  /**
   * Whether any data source has an error
   */
  const hasErrors = computed(() => {
    return Object.values(errors.value).some((error) => error !== null)
  })

  /**
   * Count of failed data sources
   */
  const errorCount = computed(() => {
    return Object.values(errors.value).filter((error) => error !== null).length
  })

  /**
   * Whether all critical data has loaded successfully
   */
  const isReady = computed(() => {
    return (
      stats.value !== null &&
      !loading.value.stats &&
      !loading.value.audit &&
      !loading.value.ciTypes
    )
  })

  // ============================================================================
  // Helper Functions
  // ============================================================================

  /**
   * Get default start date (30 days ago)
   */
  function getDefaultStartDate(): string {
    const date = new Date()
    date.setDate(date.getDate() - 30)
    return date.toISOString().split('T')[0]
  }

  /**
   * Get default end date (today)
   */
  function getDefaultEndDate(): string {
    return new Date().toISOString().split('T')[0]
  }

  /**
   * Format date for API query parameters
   */
  function formatDateForAPI(date: string | null): string | undefined {
    return date || undefined
  }

  /**
   * Build audit stats query parameters
   */
  function buildAuditStatsParams() {
    const params: Record<string, string> = {}

    if (timeRange.value.startDate) {
      params.from_date = timeRange.value.startDate
    }

    if (timeRange.value.endDate) {
      params.to_date = timeRange.value.endDate
    }

    return params
  }

  // ============================================================================
  // Data Fetching Functions
  // ============================================================================

  /**
   * Fetch basic dashboard statistics
   */
  async function fetchStats(): Promise<void> {
    loading.value.stats = true
    errors.value.stats = null

    try {
      const response = await api.get<DashboardStats>('/dashboard/stats')
      stats.value = response.data
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error?.message ||
        error.message ||
        'Failed to load dashboard statistics'
      errors.value.stats = errorMessage
      console.error('Dashboard stats fetch error:', error)
    } finally {
      loading.value.stats = false
    }
  }

  /**
   * Fetch audit statistics with time filtering
   */
  async function fetchAuditStats(): Promise<void> {
    loading.value.audit = true
    errors.value.audit = null

    try {
      const params = buildAuditStatsParams()
      const response = await api.get<AuditStats>('/audit/stats', { params })
      auditStats.value = response.data
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error?.message ||
        error.message ||
        'Failed to load audit statistics'
      errors.value.audit = errorMessage
      console.error('Audit stats fetch error:', error)
    } finally {
      loading.value.audit = false
    }
  }

  /**
   * Fetch CI type usage distribution
   */
  async function fetchCITypeUsage(): Promise<void> {
    loading.value.ciTypes = true
    errors.value.ciTypes = null

    try {
      const response = await api.get<CITypeUsage[]>('/analytics/ci-types/usage')
      ciTypeUsage.value = response.data || []
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error?.message ||
        error.message ||
        'Failed to load CI type usage'
      errors.value.ciTypes = errorMessage
      console.error('CI type usage fetch error:', error)
      ciTypeUsage.value = []
    } finally {
      loading.value.ciTypes = false
    }
  }

  /**
   * Fetch most connected CIs for network analytics
   */
  async function fetchMostConnected(): Promise<void> {
    loading.value.network = true
    errors.value.network = null

    try {
      const response = await api.get<MostConnectedCI[]>('/analytics/most-connected', {
        params: { limit: 10 },
      })
      mostConnected.value = response.data || []
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error?.message ||
        error.message ||
        'Failed to load network analytics'
      errors.value.network = errorMessage
      console.error('Most connected fetch error:', error)
      mostConnected.value = []
    } finally {
      loading.value.network = false
    }
  }

  /**
   * Fetch relationship type usage statistics
   */
  async function fetchRelationshipTypeUsage(): Promise<void> {
    loading.value.relationshipTypes = true
    errors.value.relationshipTypes = null

    try {
      const response = await api.get<RelationshipTypeUsage[]>(
        '/analytics/relationship-types/usage'
      )
      relationshipTypeUsage.value = response.data || []
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error?.message ||
        error.message ||
        'Failed to load relationship type usage'
      errors.value.relationshipTypes = errorMessage
      console.error('Relationship type usage fetch error:', error)
      relationshipTypeUsage.value = []
    } finally {
      loading.value.relationshipTypes = false
    }
  }

  // ============================================================================
  // Public Methods
  // ============================================================================

  /**
   * Fetch all dashboard data in parallel
   *
   * Uses Promise.allSettled to ensure one endpoint failure doesn't break others.
   * Each endpoint handles its own error state independently.
   */
  async function fetchAllData(): Promise<void> {
    // Use Promise.allSettled to run all fetches in parallel
    // This ensures that even if one fails, others can still succeed
    const results = await Promise.allSettled([
      fetchStats(),
      fetchAuditStats(),
      fetchCITypeUsage(),
      fetchMostConnected(),
      fetchRelationshipTypeUsage(),
    ])

    // Check if all requests succeeded
    const allSucceeded = results.every((result) => result.status === 'fulfilled')

    // Only update lastRefresh if at least one request succeeded
    const anySucceeded = results.some((result) => result.status === 'fulfilled')
    if (anySucceeded) {
      lastRefresh.value = new Date()
    }

    // Log any failures for debugging
    results.forEach((result, index) => {
      if (result.status === 'rejected') {
        const endpoints = [
          'dashboard/stats',
          'audit/stats',
          'analytics/ci-types/usage',
          'analytics/most-connected',
          'analytics/relationship-types/usage',
        ]
        console.error(`Failed to fetch ${endpoints[index]}:`, result.reason)
      }
    })
  }

  /**
   * Refresh all dashboard data
   *
   * This is the primary method for manually reloading all data sources.
   */
  async function refreshData(): Promise<void> {
    await fetchAllData()
  }

  /**
   * Update the time range filter and refresh time-sensitive data
   *
   * @param newTimeRange - New time range configuration
   */
  async function updateTimeRange(newTimeRange: TimeRange): Promise<void> {
    timeRange.value = newTimeRange

    // Only refresh time-sensitive endpoints (audit stats)
    // Stats, CI types, network, and relationship types don't change with time filtering
    await fetchAuditStats()
  }

  /**
   * Clear all errors
   */
  function clearErrors(): void {
    errors.value = {
      stats: null,
      audit: null,
      ciTypes: null,
      network: null,
      relationshipTypes: null,
      growth: null,
    }
  }

  /**
   * Retry a specific failed data source
   *
   * @param source - Data source to retry
   */
  async function retryDataSource(
    source: keyof DashboardLoadingState
  ): Promise<void> {
    switch (source) {
      case 'stats':
        await fetchStats()
        break
      case 'audit':
        await fetchAuditStats()
        break
      case 'ciTypes':
        await fetchCITypeUsage()
        break
      case 'network':
        await fetchMostConnected()
        break
      case 'relationshipTypes':
        await fetchRelationshipTypeUsage()
        break
      default:
        console.warn(`Unknown data source: ${source}`)
    }
  }

  /**
   * Reset all data to initial state
   */
  function resetData(): void {
    stats.value = null
    auditStats.value = null
    ciTypeUsage.value = []
    mostConnected.value = []
    relationshipTypeUsage.value = []
    lastRefresh.value = null
    clearErrors()
  }

  // ============================================================================
  // Return Public API
  // ============================================================================

  return {
    // Data
    stats,
    auditStats,
    ciTypeUsage,
    mostConnected,
    relationshipTypeUsage,
    timeRange,
    lastRefresh,

    // State
    loading,
    errors,
    isLoading,
    hasErrors,
    errorCount,
    isReady,

    // Methods
    fetchAllData,
    refreshData,
    updateTimeRange,
    retryDataSource,
    clearErrors,
    resetData,
  }
}
