package amortization

import (
	"time"

	"github.com/google/uuid"
)

// Core Data Models

// AmortizableCI represents a configuration item with amortization capabilities
type AmortizableCI struct {
	// Base CI Information
	ID         uuid.UUID              `json:"id"`
	Name       string                 `json:"name"`
	CIType     string                 `json:"ci_type"`
	CITypeID   uuid.UUID              `json:"ci_type_id"`
	Attributes map[string]interface{} `json:"attributes"`
	Tags       []string               `json:"tags"`

	// Lifecycle Information
	LifecycleStatusID *uuid.UUID       `json:"lifecycle_status_id,omitempty"`
	LifecycleStatus   *LifecycleStatus `json:"lifecycle_status,omitempty"`

	// Financial Information
	PurchaseCost            float64    `json:"purchase_cost"`
	SalvageValue            float64    `json:"salvage_value"`
	AmortStartDate          *time.Time `json:"amort_start_date,omitempty"`
	UsefulLifeMonths        int        `json:"useful_life_months"`
	CurrentBookValue        float64    `json:"current_book_value"`
	AccumulatedDepreciation float64    `json:"accumulated_depreciation"`
	MonthlyDepreciation     *float64   `json:"monthly_depreciation,omitempty"` // Calculated field

	// Amortization Configuration
	DepreciationMethod   string `json:"depreciation_method"`   // "straight_line", "declining_balance"
	AmortizationBehavior string `json:"amortization_behavior"` // "pending", "active", "terminal"
	IsAmortizable        bool   `json:"is_amortizable"`

	// Metadata
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// AmortizationSettings represents global amortization settings
type AmortizationSettings struct {
	ID                      string     `json:"id"` // always "global"
	Currency                string     `json:"currency"`
	DefaultUsefulLifeMonths int        `json:"default_useful_life_months"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	CreatedBy               *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy               *uuid.UUID `json:"updated_by,omitempty"`
}

// AmortizationEntry represents a single amortization ledger entry
type AmortizationEntry struct {
	ID                      uuid.UUID              `json:"id"`
	CIID                    uuid.UUID              `json:"ci_id"`
	CIName                  string                 `json:"ci_name"`
	AmortizationRunID       *uuid.UUID             `json:"amortization_run_id,omitempty"`
	EntryType               string                 `json:"entry_type"` // "depreciation", "adjustment", "write_off", "reversal", "restructuring"
	EntryDate               time.Time              `json:"entry_date"`
	Description             *string                `json:"description,omitempty"`
	Amount                  float64                `json:"amount"`
	BookValueBefore         float64                `json:"book_value_before"`
	BookValueAfter          float64                `json:"book_value_after"`
	AccumulatedDepreciation float64                `json:"accumulated_depreciation"`
	CreatedAt               time.Time              `json:"created_at"`
	CreatedBy               *uuid.UUID             `json:"created_by,omitempty"`
	Metadata                map[string]interface{} `json:"metadata,omitempty"`
}

// AmortizationRun represents a batch amortization processing run
type AmortizationRun struct {
	ID                  uuid.UUID  `json:"id"`
	Status              string     `json:"status"` // "started", "completed", "failed", "partial"
	ProcessingDate      time.Time  `json:"processing_date"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	TotalAmortizableCIs int        `json:"total_amortizable_cis"`
	ProcessedCIs        *int       `json:"processed_cis,omitempty"`
	FailedCIs           *int       `json:"failed_cis,omitempty"`
	SkippedCIs          *int       `json:"skipped_cis,omitempty"`
	TotalDepreciation   *float64   `json:"total_depreciation,omitempty"`
	ErrorSummary        *string    `json:"error_summary,omitempty"`
	IsManual            bool       `json:"is_manual"`
	DryRun              bool       `json:"dry_run"`
	TriggeredBy         *uuid.UUID `json:"triggered_by,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

// AmortizationSummary represents summarized amortization data
type AmortizationSummary struct {
	GroupBy                 string              `json:"group_by"` // "ci_type", "lifecycle_status", "department", etc.
	Groups                  []AmortizationGroup `json:"groups"`
	TotalCIs                int                 `json:"total_cis"`
	TotalBookValue          float64             `json:"total_book_value"`
	TotalDepreciation       float64             `json:"total_depreciation"`        // Accumulated depreciation to date
	TotalMonthlyDepreciation float64             `json:"total_monthly_depreciation"` // Sum of monthly depreciation rates
	GeneratedAt             time.Time           `json:"generated_at"`
}

// AmortizationGroup represents a grouped summary
type AmortizationGroup struct {
	GroupName         string  `json:"group_name"`
	GroupID           *string `json:"group_id,omitempty"`
	CICount           int     `json:"ci_count"`
	TotalBookValue    float64 `json:"total_book_value"`
	TotalDepreciation float64 `json:"total_depreciation"`
	AverageAge        float64 `json:"average_age"` // in months
}

// AmortizationCIList represents a paginated list of amortizable CIs
type AmortizationCIList struct {
	CIs        []AmortizableCI `json:"cis"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// LifecycleStatus represents lifecycle status information
type LifecycleStatus struct {
	ID                   uuid.UUID  `json:"id"`
	Name                 string     `json:"name"`
	DisplayName          string     `json:"display_name"`
	Description          *string    `json:"description,omitempty"`
	Color                *string    `json:"color,omitempty"`
	Icon                 *string    `json:"icon,omitempty"`
	AmortizationBehavior string     `json:"amortization_behavior"` // "pending", "active", "terminal"
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at,omitempty"`
}

// Request/Response Models

// UpdateAmortizationConfig represents a request to update amortization configuration
type UpdateAmortizationConfig struct {
	PurchaseCost       *float64   `json:"purchase_cost,omitempty"`
	SalvageValue       *float64   `json:"salvage_value,omitempty"`
	AmortStartDate     *time.Time `json:"amort_start_date,omitempty"`
	UsefulLifeMonths   *int       `json:"useful_life_months,omitempty"`
	DepreciationMethod *string    `json:"depreciation_method,omitempty"`
}

// UpdateAmortizationSettings represents a request to update amortization settings
type UpdateAmortizationSettings struct {
	Currency                *string `json:"currency,omitempty"`
	DefaultUsefulLifeMonths *int    `json:"default_useful_life_months,omitempty"`
}

// CreateAdjustmentRequest represents a request to create a manual adjustment
type CreateAdjustmentRequest struct {
	CIID          uuid.UUID  `json:"ci_id"`
	Amount        float64    `json:"amount"`
	Description   string     `json:"description"`
	EffectiveDate *time.Time `json:"effective_date,omitempty"`
}

// ManualRunRequest represents a request to trigger manual amortization run
type ManualRunRequest struct {
	DateOverride *time.Time  `json:"date_override,omitempty"`
	CIIDs        []uuid.UUID `json:"ci_ids,omitempty"`
	DryRun       bool        `json:"dry_run"`
}

// SummaryRequest represents a request for amortization summaries
type SummaryRequest struct {
	GroupBy   string      `json:"group_by"`
	CITypeIDs []uuid.UUID `json:"ci_type_ids,omitempty"`
	StatusIDs []uuid.UUID `json:"status_ids,omitempty"`
	DateFrom  *time.Time  `json:"date_from,omitempty"`
	DateTo    *time.Time  `json:"date_to,omitempty"`
}

// DepreciationScheduleRequest represents a request for depreciation schedule
type DepreciationScheduleRequest struct {
	DateFrom  time.Time   `json:"date_from"`
	DateTo    time.Time   `json:"date_to"`
	CIIDs     []uuid.UUID `json:"ci_ids,omitempty"`
	CITypeIDs []uuid.UUID `json:"ci_type_ids,omitempty"`
}

// Filtering and Pagination Models

// AmortizableCIFilters represents filters for listing amortizable CIs
type AmortizableCIFilters struct {
	CITypeIDs          []uuid.UUID `json:"ci_type_ids,omitempty"`
	LifecycleStatusIDs []uuid.UUID `json:"lifecycle_status_ids,omitempty"`
	IsAmortizable      *bool       `json:"is_amortizable,omitempty"`
	MinBookValue       *float64    `json:"min_book_value,omitempty"`
	MaxBookValue       *float64    `json:"max_book_value,omitempty"`
	DateFrom           *time.Time  `json:"date_from,omitempty"`
	DateTo             *time.Time  `json:"date_to,omitempty"`
	Search             *string     `json:"search,omitempty"`
	SortBy             *string     `json:"sort_by,omitempty"`
	SortOrder          *string     `json:"sort_order,omitempty"`
	Page               *int        `json:"page,omitempty"`
	PageSize           *int        `json:"page_size,omitempty"`
}

// LedgerFilters represents filters for listing ledger entries
type LedgerFilters struct {
	CIID       *uuid.UUID `json:"ci_id,omitempty"`
	EntryTypes []string   `json:"entry_types,omitempty"`
	DateFrom   *time.Time `json:"date_from,omitempty"`
	DateTo     *time.Time `json:"date_to,omitempty"`
	MinAmount  *float64   `json:"min_amount,omitempty"`
	MaxAmount  *float64   `json:"max_amount,omitempty"`
	Page       *int       `json:"page,omitempty"`
	PageSize   *int       `json:"page_size,omitempty"`
	SortBy     *string    `json:"sort_by,omitempty"`
	SortOrder  *string    `json:"sort_order,omitempty"`
}

// AmortizationRunFilters represents filters for listing amortization runs
type AmortizationRunFilters struct {
	Status      []string   `json:"status,omitempty"`
	DateFrom    *time.Time `json:"date_from,omitempty"`
	DateTo      *time.Time `json:"date_to,omitempty"`
	IsManual    *bool      `json:"is_manual,omitempty"`
	TriggeredBy *uuid.UUID `json:"triggered_by,omitempty"`
	Page        *int       `json:"page,omitempty"`
	PageSize    *int       `json:"page_size,omitempty"`
}

// Pagination Response Models

// LedgerEntryList represents a paginated list of ledger entries
type LedgerEntryList struct {
	Entries    []AmortizationEntry `json:"entries"`
	TotalCount int                 `json:"total_count"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// AmortizationRunList represents a paginated list of amortization runs
type AmortizationRunList struct {
	Runs       []AmortizationRun `json:"runs"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// Internal Processing Models

// ProcessingResult represents the result of processing a single CI

// AmortizationConfigUpdates represents database update structure
type AmortizationConfigUpdates struct {
	PurchaseCost            *float64   `db:"purchase_cost,omitempty"`
	SalvageValue            *float64   `db:"salvage_value,omitempty"`
	AmortStartDate          *time.Time `db:"amort_start_date,omitempty"`
	UsefulLifeMonths        *int       `db:"useful_life_months,omitempty"`
	CurrentBookValue        *float64   `db:"current_book_value,omitempty"`
	AccumulatedDepreciation *float64   `db:"accumulated_depreciation,omitempty"`
	UpdatedBy               *uuid.UUID `db:"updated_by,omitempty"`
	UpdatedAt               *time.Time `db:"updated_at,omitempty"`
}

// AmortizationRunUpdates represents database update structure for runs
type AmortizationRunUpdates struct {
	Status            *string    `db:"status,omitempty"`
	StartedAt         *time.Time `db:"started_at,omitempty"`
	CompletedAt       *time.Time `db:"completed_at,omitempty"`
	ProcessedCIs      *int       `db:"processed_cis,omitempty"`
	FailedCIs         *int       `db:"failed_cis,omitempty"`
	SkippedCIs        *int       `db:"skipped_cis,omitempty"`
	TotalDepreciation *float64   `db:"total_depreciation,omitempty"`
	ErrorSummary      *string    `db:"error_summary,omitempty"`
}

// Scheduler Configuration

type SchedulerConfig struct {
	Enabled       bool   `json:"enabled"`        // Whether scheduler is enabled
	ScheduleTime  string `json:"schedule_time"`  // Daily run time (e.g., "00:00")
	Timezone      string `json:"timezone"`       // Timezone for scheduling
	Concurrency   int    `json:"concurrency"`    // Number of concurrent processing
	BatchSize     int    `json:"batch_size"`     // Batch size for processing
	RetryAttempts int    `json:"retry_attempts"` // Number of retry attempts
	RetryDelay    string `json:"retry_delay"`    // Delay between retries
	LockTimeout   string `json:"lock_timeout"`   // Lock timeout for distributed processing
}

// Supporting Types

// LedgerEntry alias for consistency
type LedgerEntry = AmortizationEntry

// Additional models for interfaces compatibility
type JobQueueStats struct {
	PendingJobs int `json:"pending_jobs"`
}

type AmortizationJob struct {
	ID        uuid.UUID `json:"id"`
	CIID      uuid.UUID `json:"ci_id"`
	Type      string    `json:"type"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

// Additional calculation and reporting types for interface compatibility

// AmortizationDetails extends AmortizableCI with calculated fields
type AmortizationDetails struct {
	AmortizableCI
	// Calculated Fields
	MonthlyDepreciation *float64      `json:"monthly_depreciation,omitempty"`
	RemainingLifeMonths *int          `json:"remaining_life_months,omitempty"`
	RecentLedgerEntries []LedgerEntry `json:"recent_ledger_entries,omitempty"`
}

// DepreciationCalculation represents a depreciation calculation result
type DepreciationCalculation struct {
	Amount                        float64   `json:"amount"`
	BookValueBefore               float64   `json:"book_value_before"`
	BookValueAfter                float64   `json:"book_value_after"`
	AccumulatedDepreciationBefore float64   `json:"accumulated_depreciation_before"`
	AccumulatedDepreciationAfter  float64   `json:"accumulated_depreciation_after"`
	CalculationDate               time.Time `json:"calculation_date"`
	Method                        string    `json:"method"`
}

// CatchUpDepreciationCalculation represents a catch-up depreciation calculation result
type CatchUpDepreciationCalculation struct {
	MonthsDepreciated            int       `json:"months_depreciated"`
	TotalDepreciationAmount      float64   `json:"total_depreciation_amount"`
	MonthlyDepreciationAmount    float64   `json:"monthly_depreciation_amount"`
	BookValueBefore              float64   `json:"book_value_before"`
	BookValueAfter               float64   `json:"book_value_after"`
	AccumulatedDepreciationAfter float64   `json:"accumulated_depreciation_after"`
	CalculationDate              time.Time `json:"calculation_date"`
	Method                       string    `json:"method"`
}

// WriteOffCalculation represents a write-off calculation result
type WriteOffCalculation struct {
	WriteOffAmount                float64   `json:"write_off_amount"`
	BookValueBefore               float64   `json:"book_value_before"`
	BookValueAfter                float64   `json:"book_value_after"`
	AccumulatedDepreciationBefore float64   `json:"accumulated_depreciation_before"`
	AccumulatedDepreciationAfter  float64   `json:"accumulated_depreciation_after"`
	WriteOffDate                  time.Time `json:"write_off_date"`
	Reason                        string    `json:"reason"`
}

// ValueProjection represents a future value projection
type ValueProjection struct {
	ProjectedDate         time.Time `json:"projected_date"`
	ProjectedBookValue    float64   `json:"projected_book_value"`
	ProjectedDepreciation float64   `json:"projected_depreciation"`
	Confidence            string    `json:"confidence"`
	Assumptions           []string  `json:"assumptions"`
}

// HistoricalValue represents historical book value data
type HistoricalValue struct {
	Date         time.Time `json:"date"`
	BookValue    float64   `json:"book_value"`
	Depreciation float64   `json:"depreciation"`
}

// DepreciationScheduleEntry represents a single entry in a depreciation schedule report
type DepreciationScheduleEntry struct {
	CIID                    uuid.UUID `json:"ci_id"`
	CIName                  string    `json:"ci_name"`
	CIType                  string    `json:"ci_type"`
	PeriodStart             time.Time `json:"period_start"`
	PeriodEnd               time.Time `json:"period_end"`
	OpeningBookValue        float64   `json:"opening_book_value"`
	DepreciationAmount      float64   `json:"depreciation_amount"`
	ClosingBookValue        float64   `json:"closing_book_value"`
	AccumulatedDepreciation float64   `json:"accumulated_depreciation"`
}

// DepreciationScheduleRange represents the date range for a depreciation schedule
type DepreciationScheduleRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// DepreciationSchedule represents a complete depreciation schedule report
type DepreciationSchedule struct {
	ReportID  uuid.UUID                   `json:"report_id"`
	DateRange DepreciationScheduleRange   `json:"date_range"`
	Schedule  []DepreciationScheduleEntry `json:"schedule"`
}

// RestructuringCalculation represents a prospective recalculation when useful life changes
type RestructuringCalculation struct {
	// Current state
	CurrentUsefulLifeMonths    int     `json:"current_useful_life_months"`
	CurrentMonthlyDepreciation float64 `json:"current_monthly_depreciation"`
	CurrentBookValue           float64 `json:"current_book_value"`
	AccumulatedDepreciation    float64 `json:"accumulated_depreciation"`
	RemainingMonthsOld         int     `json:"remaining_months_old"`

	// New configuration
	NewUsefulLifeMonths    int     `json:"new_useful_life_months"`
	RemainingMonthsNew     int     `json:"remaining_months_new"`
	NewMonthlyDepreciation float64 `json:"new_monthly_depreciation"`

	// Impact
	MonthlyDepreciationChange float64    `json:"monthly_depreciation_change"`
	PercentChange             float64    `json:"percent_change"`
	RemainingLifeExtension    int        `json:"remaining_life_extension"`
	NewEndDate                *time.Time `json:"new_end_date,omitempty"`

	// Validation
	IsValid           bool   `json:"is_valid"`
	ValidationMessage string `json:"validation_message,omitempty"`
}

// RestructureRequest represents a request to restructure amortization
type RestructureRequest struct {
	CIID                uuid.UUID  `json:"ci_id"`
	NewUsefulLifeMonths int        `json:"new_useful_life_months"`
	Reason              string     `json:"reason"`
	EffectiveDate       *time.Time `json:"effective_date,omitempty"`
}

// RestructureResult represents the result of a restructuring operation
type RestructureResult struct {
	Success        bool                      `json:"success"`
	Calculation    *RestructuringCalculation `json:"calculation,omitempty"`
	LedgerEntryID  uuid.UUID                 `json:"ledger_entry_id,omitempty"`
	UpdatedDetails *AmortizationDetails      `json:"updated_details,omitempty"`
	Message        string                    `json:"message,omitempty"`
}
