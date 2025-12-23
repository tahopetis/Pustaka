import axios from 'axios'
import type {
  AmortizationSettings,
  AssetFinancials,
  AmortizationLedgerEntry,
  AmortizationRun,
  AssetSummary,
  AmortizationMetrics,
  AdjustmentEntry,
  BulkFinancialsUpdate,
  SchedulerStatus,
  AmortizationLedgerFilters,
  AssetFinancialsFilters,
  ApiResponse,
  PaginatedResponse,
  RestructuringCalculation,
  RestructureResult
} from '@/types/amortization'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

class AmortizationApiService {
  private api = axios.create({
    baseURL: `${API_BASE_URL}/api/v1/amortization`,
    headers: {
      'Content-Type': 'application/json',
    },
  })

  constructor() {
    // Add request interceptor to include auth token
    this.api.interceptors.request.use((config) => {
      const token = localStorage.getItem('access_token')
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      return config
    })

    // Add response interceptor for error handling
    this.api.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Handle token expiration or invalid auth
          localStorage.removeItem('access_token')
          window.location.href = '/auth/login'
        }
        return Promise.reject(error)
      }
    )
  }

  // Settings Management
  async getSettings(): Promise<AmortizationSettings> {
    const response = await this.api.get('/settings')
    return response.data
  }

  async updateSettings(settings: Partial<AmortizationSettings>): Promise<AmortizationSettings> {
    const response = await this.api.put('/settings', settings)
    return response.data
  }

  // Asset Financials Management
  async getAssetFinancials(ciId: string): Promise<AssetFinancials> {
    const response = await this.api.get(`/configuration-items/${ciId}`)
    return response.data
  }

  async updateAssetFinancials(ciId: string, financials: Partial<AssetFinancials>): Promise<AssetFinancials> {
    // Convert to the format expected by the backend UpdateAmortizationConfig
    const updateData = {
      purchase_cost: financials.purchase_cost,
      salvage_value: financials.salvage_value,
      amort_start_date: financials.amort_start_date ? new Date(financials.amort_start_date) : undefined,
      useful_life_months: financials.useful_life_months,
    }
    const response = await this.api.put(`/configuration-items/${ciId}`, updateData)
    return response.data
  }

  async createAssetFinancials(ciId: string, financials: AssetFinancials): Promise<AssetFinancials> {
    // Convert to the format expected by the backend UpdateAmortizationConfig
    const updateData = {
      purchase_cost: financials.purchase_cost,
      salvage_value: financials.salvage_value,
      amort_start_date: financials.amort_start_date ? new Date(financials.amort_start_date) : undefined,
      useful_life_months: financials.useful_life_months,
    }
    const response = await this.api.put(`/configuration-items/${ciId}`, updateData)
    return response.data
  }

  async getAssetSummaries(filters?: AssetFinancialsFilters): Promise<PaginatedResponse<AssetSummary>> {
    const response = await this.api.get('/configuration-items', { params: filters })
    return response.data
  }

  async getAssetSummary(): Promise<AmortizationMetrics> {
    const response = await this.api.get('/summaries')
    return response.data
  }

  // Ledger Management
  async getLedgerEntries(filters?: AmortizationLedgerFilters): Promise<PaginatedResponse<AmortizationLedgerEntry>> {
    const response = await this.api.get('/ledger', { params: filters })
    return response.data
  }

  async getLedgerEntry(entryId: string): Promise<AmortizationLedgerEntry> {
    const response = await this.api.get(`/ledger/${entryId}`)
    return response.data
  }

  // Reports
  async getJournalReport(filters?: AmortizationLedgerFilters): Promise<PaginatedResponse<AmortizationLedgerEntry>> {
    const response = await this.api.get('/reports/journal', { params: filters })
    return response.data
  }

  // Scheduler Management
  async runScheduler(dryRun = false): Promise<AmortizationRun> {
    const response = await this.api.post('/runs', { dry_run: dryRun })
    return response.data
  }

  async getSchedulerStatus(): Promise<SchedulerStatus> {
    // This endpoint might not be implemented yet, using runs list as fallback
    const response = await this.api.get('/runs?limit=1')
    return {
      is_enabled: true,
      last_run: response.data.data?.[0] || null,
      next_run: null
    }
  }

  async getSchedulerRuns(): Promise<PaginatedResponse<AmortizationRun>> {
    const response = await this.api.get('/runs')
    return response.data
  }

  // Bulk Operations
  async bulkUpdateFinancials(update: BulkFinancialsUpdate): Promise<{ updated: number; failed: number }> {
    const response = await this.api.post('/bulk/financials', update)
    return response.data
  }

  // Adjustment Entries
  async createAdjustment(ciId: string, adjustment: AdjustmentEntry): Promise<AmortizationLedgerEntry> {
    const response = await this.api.post('/adjustments', { ...adjustment, ci_id: ciId })
    return response.data
  }

  // Restructuring (useful life changes with prospective recalculation)
  async previewRestructuring(ciId: string, newUsefulLifeMonths: number): Promise<RestructuringCalculation> {
    const response = await this.api.post('/restructuring/preview', {
      ci_id: ciId,
      new_useful_life_months: newUsefulLifeMonths
    })
    return response.data
  }

  async executeRestructuring(
    ciId: string,
    newUsefulLifeMonths: number,
    reason: string,
    effectiveDate?: Date
  ): Promise<RestructureResult> {
    const response = await this.api.post('/restructuring', {
      ci_id: ciId,
      new_useful_life_months: newUsefulLifeMonths,
      reason,
      effective_date: effectiveDate
    })
    return response.data
  }
}

export const amortizationApi = new AmortizationApiService()
export default amortizationApi