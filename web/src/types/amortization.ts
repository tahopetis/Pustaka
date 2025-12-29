export interface AmortizationSettings {
  id: string // always "global"
  currency: string
  default_useful_life_months: number
  created_at: string
  updated_at: string
  created_by?: string
  updated_by?: string
}

export interface AssetFinancials {
  ci_id: string
  purchase_cost?: number
  salvage_value?: number
  amort_start_date?: string
  useful_life_months?: number
  current_book_value?: number
  accumulated_depreciation?: number
  monthly_depreciation?: number
  remaining_months?: number
}

export interface AmortizationLedgerEntry {
  id: string
  ci_id: string
  amortization_run_id?: string
  entry_type: 'depreciation' | 'adjustment' | 'write_off' | 'reversal' | 'restructuring'
  entry_date: string
  description?: string
  amount: number
  book_value_before: number
  book_value_after: number
  accumulated_depreciation: number
  created_at: string
  created_by?: string
  metadata?: Record<string, any>
  ci_name?: string
  ci_type_name?: string
}

export interface AmortizationRun {
  id: string
  run_date: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  started_at?: string
  completed_at?: string
  total_assets_processed: number
  successful_entries: number
  failed_entries: number
  error_message?: string
  metadata?: Record<string, any>
}

export interface AssetSummary {
  ci_id: string
  ci_name: string
  ci_type_name: string
  purchase_cost: number
  current_book_value: number
  accumulated_depreciation: number
  monthly_depreciation: number
  remaining_months: number
  amort_start_date: string
  useful_life_months: number
  status: 'pending' | 'active' | 'terminal'
  last_depreciation_date?: string
}

export interface AmortizationMetrics {
  total_amortizable_assets: number
  total_book_value: number
  monthly_depreciation: number
  total_monthly_depreciation: number
  active_amortizations: number
}

export interface AdjustmentEntry {
  ci_id: string
  entry_type: 'adjustment' | 'write_off' | 'reversal'
  amount: number
  description?: string
  entry_date: string
  corrects_entry_id?: string
  metadata?: Record<string, any>
}

export interface AmortizationLedgerFilters {
  ci_id?: string
  entry_type?: string[]
  date_from?: string
  date_to?: string
  limit?: number
  offset?: number
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export interface AssetFinancialsFilters {
  ci_type_id?: string
  status?: string[]
  has_financial_data?: boolean
  limit?: number
  offset?: number
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export interface BulkFinancialsUpdate {
  updates: Array<{
    ci_id: string
    purchase_cost?: number
    salvage_value?: number
    amort_start_date?: string
    useful_life_months?: number
  }>
}

export interface SchedulerStatus {
  last_run_date?: string
  next_run_date?: string
  is_running: boolean
  current_run_id?: string
  run_frequency: string
}

// API Response Types
export interface ApiResponse<T> {
  data: T
  message?: string
  success: boolean
}

export interface PaginatedResponse<T> {
  data: T[]
  pagination: {
    total: number
    limit: number
    offset: number
    has_next: boolean
    has_prev: boolean
  }
}

// Store Types
export interface AmortizationState {
  settings: AmortizationSettings | null
  assets: AssetSummary[]
  ledgerEntries: AmortizationLedgerEntry[]
  metrics: AmortizationMetrics | null
  schedulerStatus: SchedulerStatus | null
  loading: boolean
  error: string | null
}

// Restructuring Types
export interface RestructuringCalculation {
  current_useful_life_months: number
  current_monthly_depreciation: number
  current_book_value: number
  accumulated_depreciation: number
  remaining_months_old: number
  new_useful_life_months: number
  remaining_months_new: number
  new_monthly_depreciation: number
  monthly_depreciation_change: number
  percent_change: number
  remaining_life_extension: number
  new_end_date?: string
  is_valid: boolean
  validation_message?: string
}

export interface RestructureResult {
  success: boolean
  calculation: RestructuringCalculation
  updated_details?: AssetFinancials
  message?: string
}