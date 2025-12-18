import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  AmortizationState,
  AmortizationSettings,
  AssetFinancials,
  AmortizationLedgerEntry,
  AmortizationRun,
  AssetSummary,
  AmortizationMetrics,
  SchedulerStatus,
  AdjustmentEntry,
  BulkFinancialsUpdate,
  AmortizationLedgerFilters,
  AssetFinancialsFilters,
  ApiResponse
} from '@/types/amortization'
import amortizationApi from '@/services/amortizationApi'

export const useAmortizationStore = defineStore('amortization', () => {
  // State
  const settings = ref<AmortizationSettings | null>(null)
  const assets = ref<AssetSummary[]>([])
  const ledgerEntries = ref<AmortizationLedgerEntry[]>([])
  const metrics = ref<AmortizationMetrics | null>(null)
  const schedulerStatus = ref<SchedulerStatus | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const isAmortizationEnabled = computed(() => settings.value !== null)
  const totalAssets = computed(() => assets.value.length)
  const activeAssets = computed(() => assets.value.filter(asset => asset.status === 'active'))
  const totalBookValue = computed(() =>
    assets.value.reduce((sum, asset) => sum + (asset.current_book_value || 0), 0)
  )

  // Actions
  const setLoading = (isLoading: boolean) => {
    loading.value = isLoading
  }

  const setError = (errorMessage: string | null) => {
    error.value = errorMessage
  }

  const clearError = () => {
    error.value = null
  }

  // Settings Actions
  async function loadSettings() {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.getSettings()
      settings.value = response.data
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load amortization settings')
      throw err
    } finally {
      setLoading(false)
    }
  }

  async function updateSettings(settingsData: Partial<AmortizationSettings>) {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.updateSettings(settingsData)
      settings.value = response.data
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to update amortization settings')
      throw err
    } finally {
      setLoading(false)
    }
  }

  // Asset Financials Actions
  async function loadAssetFinancials(ciId: string) {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.getAssetFinancials(ciId)
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load asset financials')
      throw err
    } finally {
      setLoading(false)
    }
  }

  async function updateAssetFinancials(ciId: string, financials: Partial<AssetFinancials>) {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.updateAssetFinancials(ciId, financials)
      // Refresh assets list to reflect changes
      await loadAssetSummaries()
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to update asset financials')
      throw err
    } finally {
      setLoading(false)
    }
  }

  async function createAssetFinancials(ciId: string, financials: AssetFinancials) {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.createAssetFinancials(ciId, financials)
      // Refresh assets list to reflect changes
      await loadAssetSummaries()
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to create asset financials')
      throw err
    } finally {
      setLoading(false)
    }
  }

  async function loadAssetSummaries(filters?: AssetFinancialsFilters) {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.getAssetSummaries(filters)
      assets.value = response.data.data
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load asset summaries')
      throw err
    } finally {
      setLoading(false)
    }
  }

  async function loadMetrics() {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.getAssetSummary()
      metrics.value = response.data
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load amortization metrics')
      throw err
    } finally {
      setLoading(false)
    }
  }

  // Ledger Actions
  async function loadLedgerEntries(filters?: AmortizationLedgerFilters) {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.getLedgerEntries(filters)
      ledgerEntries.value = response.data.data
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load ledger entries')
      throw err
    } finally {
      setLoading(false)
    }
  }

  async function loadJournalReport(filters?: AmortizationLedgerFilters) {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.getJournalReport(filters)
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load journal report')
      throw err
    } finally {
      setLoading(false)
    }
  }

  // Scheduler Actions
  async function runScheduler() {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.runScheduler()
      // Refresh scheduler status
      await loadSchedulerStatus()
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to run scheduler')
      throw err
    } finally {
      setLoading(false)
    }
  }

  async function loadSchedulerStatus() {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.getSchedulerStatus()
      schedulerStatus.value = response.data
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load scheduler status')
      throw err
    } finally {
      setLoading(false)
    }
  }

  async function loadSchedulerRuns() {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.getSchedulerRuns()
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load scheduler runs')
      throw err
    } finally {
      setLoading(false)
    }
  }

  // Bulk Operations
  async function bulkUpdateFinancials(update: BulkFinancialsUpdate) {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.bulkUpdateFinancials(update)
      // Refresh assets list to reflect changes
      await loadAssetSummaries()
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to bulk update financials')
      throw err
    } finally {
      setLoading(false)
    }
  }

  // Adjustment Entries
  async function createAdjustment(ciId: string, adjustment: AdjustmentEntry) {
    setLoading(true)
    clearError()
    try {
      const response = await amortizationApi.createAdjustment(ciId, adjustment)
      // Refresh ledger entries to show the new adjustment
      await loadLedgerEntries({ ci_id: ciId })
      return response
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to create adjustment entry')
      throw err
    } finally {
      setLoading(false)
    }
  }

  // Helper Functions
  function getAssetById(ciId: string): AssetSummary | undefined {
    return assets.value.find(asset => asset.ci_id === ciId)
  }

  function getLedgerEntriesForAsset(ciId: string): AmortizationLedgerEntry[] {
    return ledgerEntries.value.filter(entry => entry.ci_id === ciId)
  }

  function resetStore() {
    settings.value = null
    assets.value = []
    ledgerEntries.value = []
    metrics.value = null
    schedulerStatus.value = null
    loading.value = false
    error.value = null
  }

  return {
    // State
    settings,
    assets,
    ledgerEntries,
    metrics,
    schedulerStatus,
    loading,
    error,

    // Getters
    isAmortizationEnabled,
    totalAssets,
    activeAssets,
    totalBookValue,

    // Actions
    setLoading,
    setError,
    clearError,
    loadSettings,
    updateSettings,
    loadAssetFinancials,
    updateAssetFinancials,
    createAssetFinancials,
    loadAssetSummaries,
    loadMetrics,
    loadLedgerEntries,
    loadJournalReport,
    runScheduler,
    loadSchedulerStatus,
    loadSchedulerRuns,
    bulkUpdateFinancials,
    createAdjustment,

    // Helpers
    getAssetById,
    getLedgerEntriesForAsset,
    resetStore
  }
})