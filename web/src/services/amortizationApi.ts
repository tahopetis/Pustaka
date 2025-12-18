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
  PaginatedResponse
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
  async getSettings(): Promise<ApiResponse<AmortizationSettings>> {
    const response = await this.api.get('/settings')
    return response.data
  }

  async updateSettings(settings: Partial<AmortizationSettings>): Promise<ApiResponse<AmortizationSettings>> {
    const response = await this.api.put('/settings', settings)
    return response.data
  }

  // Asset Financials Management
  async getAssetFinancials(ciId: string): Promise<ApiResponse<AssetFinancials>> {
    const response = await this.api.get(`/configuration-items/${ciId}`)
    return response.data
  }

  async updateAssetFinancials(ciId: string, financials: Partial<AssetFinancials>): Promise<ApiResponse<AssetFinancials>> {
    const response = await this.api.put(`/configuration-items/${ciId}`, financials)
    return response.data
  }

  async createAssetFinancials(ciId: string, financials: AssetFinancials): Promise<ApiResponse<AssetFinancials>> {
    const response = await this.api.put(`/configuration-items/${ciId}`, financials)
    return response.data
  }

  async getAssetSummaries(filters?: AssetFinancialsFilters): Promise<ApiResponse<PaginatedResponse<AssetSummary>>> {
    const response = await this.api.get('/configuration-items', { params: filters })
    return response.data
  }

  async getAssetSummary(): Promise<ApiResponse<AmortizationMetrics>> {
    const response = await this.api.get('/summaries')
    return response.data
  }

  // Ledger Management
  async getLedgerEntries(filters?: AmortizationLedgerFilters): Promise<ApiResponse<PaginatedResponse<AmortizationLedgerEntry>>> {
    const response = await this.api.get('/ledger', { params: filters })
    return response.data
  }

  async getLedgerEntry(entryId: string): Promise<ApiResponse<AmortizationLedgerEntry>> {
    const response = await this.api.get(`/ledger/${entryId}`)
    return response.data
  }

  // Reports
  async getJournalReport(filters?: AmortizationLedgerFilters): Promise<ApiResponse<PaginatedResponse<AmortizationLedgerEntry>>> {
    const response = await this.api.get('/reports/journal', { params: filters })
    return response.data
  }

  // Scheduler Management
  async runScheduler(): Promise<ApiResponse<AmortizationRun>> {
    const response = await this.api.post('/scheduler/run')
    return response.data
  }

  async getSchedulerStatus(): Promise<ApiResponse<SchedulerStatus>> {
    const response = await this.api.get('/scheduler/status')
    return response.data
  }

  async getSchedulerRuns(): Promise<ApiResponse<PaginatedResponse<AmortizationRun>>> {
    const response = await this.api.get('/scheduler/runs')
    return response.data
  }

  // Bulk Operations
  async bulkUpdateFinancials(update: BulkFinancialsUpdate): Promise<ApiResponse<{ updated: number; failed: number }>> {
    const response = await this.api.post('/bulk/financials', update)
    return response.data
  }

  // Adjustment Entries
  async createAdjustment(ciId: string, adjustment: AdjustmentEntry): Promise<ApiResponse<AmortizationLedgerEntry>> {
    const response = await this.api.post('/adjustments', { ...adjustment, ci_id: ciId })
    return response.data
  }
}

export const amortizationApi = new AmortizationApiService()
export default amortizationApi