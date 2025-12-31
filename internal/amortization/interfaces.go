package amortization

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/ci"
)

// Core Interfaces

// Service defines the core business logic interface for amortization management
type Service interface {
	// Amortization Configuration
	GetAmortizationDetails(ctx context.Context, ciID uuid.UUID) (*AmortizationDetails, error)
	UpdateAmortizationConfig(ctx context.Context, ciID uuid.UUID, req *UpdateAmortizationConfig, userID uuid.UUID) (*AmortizationDetails, error)
	ListAmortizableCIs(ctx context.Context, filters *AmortizableCIFilters) (*AmortizationCIList, error)

	// Settings Management
	GetAmortizationSettings(ctx context.Context) (*AmortizationSettings, error)
	UpdateAmortizationSettings(ctx context.Context, req *UpdateAmortizationSettings, userID uuid.UUID) (*AmortizationSettings, error)

	// Ledger Management
	GetLedgerEntries(ctx context.Context, filters *LedgerFilters) (*LedgerEntryList, error)
	GetLedgerEntry(ctx context.Context, entryID uuid.UUID) (*LedgerEntry, error)
	CreateAdjustment(ctx context.Context, req *CreateAdjustmentRequest, userID uuid.UUID) (*LedgerEntry, error)

	// Amortization Processing
	ProcessDailyAmortization(ctx context.Context, processingDate time.Time) (*AmortizationRun, error)
	TriggerManualRun(ctx context.Context, req *ManualRunRequest, userID uuid.UUID) (*AmortizationRun, error)
	GetAmortizationRun(ctx context.Context, runID uuid.UUID) (*AmortizationRun, error)
	ListAmortizationRuns(ctx context.Context, filters *AmortizationRunFilters) (*AmortizationRunList, error)

	// Reporting and Summaries
	GetAmortizationSummaries(ctx context.Context, req *SummaryRequest) (*AmortizationSummary, error)
	GenerateDepreciationSchedule(ctx context.Context, req *DepreciationScheduleRequest) (*DepreciationScheduleResponse, error)

	// Lifecycle Integration
	HandleTerminalStatusChange(ctx context.Context, ciID uuid.UUID, oldStatusID, newStatusID uuid.UUID, userID uuid.UUID) error
	RecalculateAmortization(ctx context.Context, ciID uuid.UUID, userID uuid.UUID) error

	// Restructuring (prospective recalculation)
	PreviewRestructuring(ctx context.Context, ciID uuid.UUID, newUsefulLifeMonths int) (*RestructuringCalculation, error)
	RestructureAmortization(ctx context.Context, req *RestructureRequest, userID uuid.UUID) (*RestructureResult, error)
}

// Repository defines the data access interface for amortization operations
type Repository interface {
	// Configuration Items with Amortization
	GetAmortizableCI(ctx context.Context, ciID uuid.UUID) (*AmortizableCI, error)
	ListAmortizableCIs(ctx context.Context, filters *AmortizableCIFilters) (*AmortizationCIList, error)
	UpdateAmortizationConfig(ctx context.Context, ciID uuid.UUID, updates *AmortizationConfigUpdates) error

	// Ledger Operations
	CreateLedgerEntry(ctx context.Context, entry *LedgerEntry) error
	GetLedgerEntries(ctx context.Context, filters *LedgerFilters) (*LedgerEntryList, error)
	GetLedgerEntry(ctx context.Context, entryID uuid.UUID) (*LedgerEntry, error)
	GetCILatestLedgerEntry(ctx context.Context, ciID uuid.UUID) (*LedgerEntry, error)
	HasLedgerEntriesForCI(ctx context.Context, ciID uuid.UUID) (bool, error)

	// Amortization Run Management
	CreateAmortizationRun(ctx context.Context, run *AmortizationRun) error
	UpdateAmortizationRun(ctx context.Context, runID uuid.UUID, updates *AmortizationRunUpdates) error
	GetAmortizationRun(ctx context.Context, runID uuid.UUID) (*AmortizationRun, error)
	ListAmortizationRuns(ctx context.Context, filters *AmortizationRunFilters) (*AmortizationRunList, error)

	// Summary and Reporting
	GetAmortizationSummaries(ctx context.Context, req *SummaryRequest) (*AmortizationSummary, error)
	GetDepreciationScheduleData(ctx context.Context, req *DepreciationScheduleRequest) (*DepreciationScheduleResponse, error)

	// Batch Operations for Scheduler
	GetCIsForProcessing(ctx context.Context, processingDate time.Time, limit int) ([]uuid.UUID, error)
	MarkCIsAsProcessed(ctx context.Context, runID uuid.UUID, processedCIs []ProcessingResult) error

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx interface{}) error) error

	// Settings management
	GetAmortizationSettings(ctx context.Context) (*AmortizationSettings, error)
	UpdateAmortizationSettings(ctx context.Context, settings *AmortizationSettings, userID uuid.UUID) error
}

// CacheRepositoryInterface defines caching operations for amortization
type CacheRepositoryInterface interface {
	// Caching frequently accessed data
	GetAmortizableCI(ctx context.Context, ciID uuid.UUID) (*AmortizableCI, error)
	SetAmortizableCI(ctx context.Context, ciID uuid.UUID, ci *AmortizableCI, ttl time.Duration) error
	InvalidateAmortizableCI(ctx context.Context, ciID uuid.UUID) error

	GetCILatestLedgerEntry(ctx context.Context, ciID uuid.UUID) (*LedgerEntry, error)
	SetCILatestLedgerEntry(ctx context.Context, ciID uuid.UUID, entry *LedgerEntry, ttl time.Duration) error
	InvalidateCILedgerEntries(ctx context.Context, ciID uuid.UUID) error

	// Amortization Type Definitions (cached for performance)
	GetAmortizationBehaviorForStatus(ctx context.Context, statusID uuid.UUID) (string, error)
	SetAmortizationBehaviorForStatus(ctx context.Context, statusID uuid.UUID, behavior string) error
	GetIsAmortizableForCIType(ctx context.Context, ciTypeID uuid.UUID) (bool, error)
	SetIsAmortizableForCIType(ctx context.Context, ciTypeID uuid.UUID, isAmortizable bool) error

	// Distributed locking for scheduler
	AcquireProcessingLock(ctx context.Context, lockKey string, ttl time.Duration) (bool, error)
	ReleaseProcessingLock(ctx context.Context, lockKey string) error

	// Job Queue Management
	EnqueueAmortizationJob(ctx context.Context, job *AmortizationJob) error
	DequeueAmortizationJob(ctx context.Context) (*AmortizationJob, error)
	GetJobQueueStats(ctx context.Context) (*JobQueueStats, error)

	// Summary caching
	GetAmortizationSummaries(ctx context.Context, groupBy string) (*AmortizationSummary, error)
	SetAmortizationSummaries(ctx context.Context, groupBy string, summary *AmortizationSummary) error

	// User session caching
	CacheUserFilters(ctx context.Context, userID uuid.UUID, filters *AmortizableCIFilters) error
	GetUserFilters(ctx context.Context, userID uuid.UUID) (*AmortizableCIFilters, error)

	// Cache management
	InvalidatePattern(ctx context.Context, pattern string) error
	WarmupCache(ctx context.Context) error
	GetCacheStats(ctx context.Context) (map[string]interface{}, error)
}

// CalculatorInterface defines the depreciation calculation interface
type CalculatorInterface interface {
	// Core calculations
	CalculateMonthlyDepreciation(ctx context.Context, ci *AmortizableCI, entryDate time.Time) (*DepreciationCalculation, error)
	CalculateCatchUpDepreciation(ctx context.Context, ci *AmortizableCI, asOfDate time.Time) (*CatchUpDepreciationCalculation, error)
	CalculateWriteOff(ctx context.Context, ci *AmortizableCI, writeOffDate time.Time) (*WriteOffCalculation, error)
	CalculateDepreciationToDate(ctx context.Context, ci *AmortizableCI, asOfDate time.Time) ([]DepreciationCalculation, error)
	CalculateRemainingLife(ctx context.Context, ci *AmortizableCI, asOfDate time.Time) (int, error)

	// Configuration validation
	ValidateAmortizationConfig(ctx context.Context, config *AmortizationConfig) error
	ValidateAdjustmentAmount(ci *AmortizableCI, amount float64) error

	// Restructuring calculations (prospective recalculation)
	CalculateRestructuring(ctx context.Context, ci *AmortizableCI, newUsefulLifeMonths int, asOfDate time.Time) (*RestructuringCalculation, error)

	// Projection calculations
	ProjectFutureValue(ctx context.Context, ci *AmortizableCI, futureDate time.Time) (*ValueProjection, error)
	ReconstructHistoricalValues(ciID uuid.UUID, dateFrom, dateTo time.Time) ([]HistoricalValue, error)
}

// SchedulerInterface defines the scheduler interface
type SchedulerInterface interface {
	// Scheduling operations
	ScheduleDailyRun(ctx context.Context) error
	UnscheduleDailyRun() error

	// Run execution
	ExecuteScheduledRun(ctx context.Context, processingDate time.Time) (*AmortizationRun, error)
	ExecuteManualRun(ctx context.Context, req *ManualRunRequest, userID uuid.UUID) (*AmortizationRun, error)
}

// External Service Interfaces

// CIServiceInterface defines the CI service interface
type CIServiceInterface interface {
	GetCI(ctx context.Context, id uuid.UUID) (*ConfigurationItem, error)
	GetCIType(ctx context.Context, id uuid.UUID) (*CITypeDefinition, error)
	UpdateCI(ctx context.Context, id uuid.UUID, req interface{}, userID uuid.UUID) error
}

// LifecycleServiceInterface defines the lifecycle service interface
type LifecycleServiceInterface interface {
	GetLifecycleStatus(ctx context.Context, id uuid.UUID) (*LifecycleStatus, error)
}

// Additional Types needed for interfaces

// Adapter types for existing CI services
type CIServiceAdapter struct {
	Service *ci.Service
}

func (a *CIServiceAdapter) GetCI(ctx context.Context, id uuid.UUID) (*ConfigurationItem, error) {
	ci, err := a.Service.GetCI(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create a minimal amortization.ConfigurationItem from the ci.ConfigurationItem
	result := &ConfigurationItem{
		ID:                ci.ID,
		Name:              ci.Name,
		CIType:            ci.CIType,
		Attributes:        ci.Attributes,
		Tags:              ci.Tags,
		LifecycleStatusID: ci.LifecycleStatusID,
		CreatedAt:         ci.CreatedAt,
		UpdatedAt:         &ci.UpdatedAt,
		CreatedBy:         ci.CreatedBy,
		UpdatedBy:         ci.UpdatedBy,
	}

	// Copy lifecycle status if present
	if ci.LifecycleStatus != nil {
		result.LifecycleStatus = &LifecycleStatus{
			ID:                   ci.LifecycleStatus.ID,
			Name:                 ci.LifecycleStatus.Name,
			DisplayName:          ci.LifecycleStatus.DisplayName,
			Description:          ci.LifecycleStatus.Description,
			Color:                ci.LifecycleStatus.Color,
			Icon:                 ci.LifecycleStatus.Icon,
			AmortizationBehavior: "pending", // Default value
			CreatedAt:            ci.LifecycleStatus.CreatedAt,
			UpdatedAt:            ci.LifecycleStatus.UpdatedAt,
		}
	}

	return result, nil
}

func (a *CIServiceAdapter) GetCIType(ctx context.Context, id uuid.UUID) (*CITypeDefinition, error) {
	ciType, err := a.Service.GetCIType(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert ci.CITypeDefinition to amortization.CITypeDefinition
	// Note: ci.CITypeDefinition doesn't have IsAmortizable field in the existing model
	result := &CITypeDefinition{
		ID:                 ciType.ID,
		Name:               ciType.Name,
		Description:        ciType.Description,
		RequiredAttributes: convertAttributes(ciType.RequiredAttributes),
		OptionalAttributes: convertAttributes(ciType.OptionalAttributes),
		IsAmortizable:      false, // Default value, would need to be determined separately
		CreatedAt:          ciType.CreatedAt,
		UpdatedAt:          &ciType.UpdatedAt,
		CreatedBy:          ciType.CreatedBy,
	}

	return result, nil
}

func (a *CIServiceAdapter) UpdateCI(ctx context.Context, id uuid.UUID, req interface{}, userID uuid.UUID) error {
	// Convert amortization request to ci package request type
	// This is a simplified conversion - in production, you'd want proper type handling
	_ = req // Ignore req for now, in full implementation would convert types

	// For now, just pass through with a basic ci.UpdateCIRequest
	// In a full implementation, this would convert amortization.UpdateAmortizationConfig to ci.UpdateCIRequest
	_, err := a.Service.UpdateCI(ctx, id, &ci.UpdateCIRequest{}, userID)
	return err
}

// convertAttributes converts ci.AttributeDefinition to amortization.AttributeDefinition
func convertAttributes(ciAttrs []ci.AttributeDefinition) []AttributeDefinition {
	result := make([]AttributeDefinition, len(ciAttrs))
	for i, attr := range ciAttrs {
		result[i] = AttributeDefinition{
			Name:         attr.Name,
			Type:         attr.Type,
			Required:     false, // Default value, would need determination
			DefaultValue: nil,
			Description:  &attr.Description,
			Validation:   nil,
		}
	}
	return result
}

type LifecycleServiceAdapter struct {
	Service *ci.LifecycleStatusService
}

func (a *LifecycleServiceAdapter) GetLifecycleStatus(ctx context.Context, id uuid.UUID) (*LifecycleStatus, error) {
	status, err := a.Service.GetLifecycleStatus(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert ci.LifecycleStatus to amortization.LifecycleStatus
	// Note: ci.LifecycleStatus doesn't have AmortizationBehavior field
	result := &LifecycleStatus{
		ID:                   status.ID,
		Name:                 status.Name,
		DisplayName:          status.DisplayName,
		Description:          status.Description,
		Color:                status.Color,
		Icon:                 status.Icon,
		AmortizationBehavior: "pending", // Default value
		CreatedAt:            status.CreatedAt,
		UpdatedAt:            status.UpdatedAt,
	}

	return result, nil
}

// AmortizationConfig represents amortization configuration
type AmortizationConfig struct {
	PurchaseCost       *float64   `json:"purchase_cost,omitempty"`
	SalvageValue       *float64   `json:"salvage_value,omitempty"`
	AmortStartDate     *time.Time `json:"amort_start_date,omitempty"`
	UsefulLifeMonths   *int       `json:"useful_life_months,omitempty"`
	DepreciationMethod *string    `json:"depreciation_method,omitempty"`
}

// ProcessingResult represents the result of processing a single CI
type ProcessingResult struct {
	CIID               uuid.UUID `json:"ci_id"`
	Status             string    `json:"status"` // "processed", "skipped", "failed"
	DepreciationAmount float64   `json:"depreciation_amount,omitempty"`
	ErrorMessage       *string   `json:"error_message,omitempty"`
	ProcessedAt        time.Time `json:"processed_at"`
}

// ConfigurationItem represents basic CI information
type ConfigurationItem struct {
	ID                uuid.UUID              `json:"id"`
	Name              string                 `json:"name"`
	CIType            string                 `json:"ci_type"`
	Attributes        map[string]interface{} `json:"attributes"`
	Tags              []string               `json:"tags"`
	LifecycleStatusID *uuid.UUID             `json:"lifecycle_status_id,omitempty"`
	LifecycleStatus   *LifecycleStatus       `json:"lifecycle_status,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         *time.Time             `json:"updated_at,omitempty"`
	CreatedBy         uuid.UUID              `json:"created_by"`
	UpdatedBy         *uuid.UUID             `json:"updated_by,omitempty"`
}

// CITypeDefinition represents CI type information
type CITypeDefinition struct {
	ID                 uuid.UUID             `json:"id"`
	Name               string                `json:"name"`
	Description        *string               `json:"description,omitempty"`
	RequiredAttributes []AttributeDefinition `json:"required_attributes"`
	OptionalAttributes []AttributeDefinition `json:"optional_attributes"`
	IsAmortizable      bool                  `json:"is_amortizable"`
	CreatedAt          time.Time             `json:"created_at"`
	UpdatedAt          *time.Time            `json:"updated_at,omitempty"`
	CreatedBy          uuid.UUID             `json:"created_by"`
	UpdatedBy          *uuid.UUID            `json:"updated_by,omitempty"`
}

// AttributeDefinition represents CI attribute schema
type AttributeDefinition struct {
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Required     bool         `json:"required"`
	DefaultValue *interface{} `json:"default_value,omitempty"`
	Description  *string      `json:"description,omitempty"`
	Validation   *string      `json:"validation,omitempty"`
}
