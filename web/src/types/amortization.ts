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
  created_by_name?: string
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

  // New comprehensive accounting metrics
  total_original_cost: number           // OCC
  total_gross_book_value: number        // GVB
  total_net_book_value: number          // NBV
  total_accumulated_depreciation: number // AD
  total_salvage_value: number           // SV

  // Legacy fields (for backward compatibility)
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
  ci_type_id?: string
  ci_name_search?: string
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

// Depreciation Schedule Types
export interface DepreciationScheduleRequest {
  start_date: string
  end_date: string
  ci_type_ids?: string[]
  ci_ids?: string[]
}

export interface MonthlyScheduleEntry {
  month: string
  is_projected: boolean
  opening_book_value: number          // NBV opening
  gross_book_value: number            // GVB for this month
  depreciation_amount: number
  write_off_amount: number
  adjustment_amount: number           // ± impact to GVB
  closing_book_value: number          // NBV closing
  accumulated_depreciation: number    // Running AD
  active_assets_count: number
}

export interface ScheduleSummary {
  total_original_cost: number          // OCC
  total_gross_book_value: number        // GVB
  total_net_book_value: number          // NBV
  total_depreciation: number            // AD
  total_write_offs: number
  total_adjustments: number             // Net ±
  total_salvage_value: number           // SV
  average_monthly_expense: number
  projected_end_value: number
  depreciation_percentage: number       // AD/GVB × 100
  remaining_percentage: number          // NBV/GVB × 100
}

export interface CITypeScheduleSummary {
  ci_type_id: string
  ci_type_name: string
  asset_count: number
  total_book_value: number
  monthly_depreciation: number
}

export interface AssetScheduleSummary {
  ci_id: string
  ci_name: string
  ci_type_name: string
  current_book_value: number
  monthly_depreciation: number
  remaining_months: number
  projected_end_date?: string
}

export interface PeriodSummary {
  opening_book_value: number        // NBV at start of period
  closing_book_value: number        // NBV at end of period
  period_depreciation: number       // Total depreciation during period
  period_write_offs: number         // Total write-offs during period
  period_adjustments: number        // Net adjustments during period
  average_monthly_expense: number   // Avg monthly depreciation for period
  opening_date: string              // Formatted opening date
  closing_date: string              // Formatted closing date
  months_count: number              // Number of months in period
}

export interface DepreciationScheduleResponse {
  currency: string
  start_date: string
  end_date: string

  // New comprehensive metrics (cumulative/lifetime values)
  total_original_cost: number           // OCC
  total_gross_book_value: number        // GVB
  total_net_book_value: number          // NBV (current)
  total_salvage_value: number           // SV
  total_accumulated_depreciation: number // AD

  // Period-specific metrics (for the selected date range only)
  period_summary: PeriodSummary

  summary: ScheduleSummary
  monthly_data: MonthlyScheduleEntry[]
  by_ci_type?: CITypeScheduleSummary[]
  by_asset?: AssetScheduleSummary[]
}
