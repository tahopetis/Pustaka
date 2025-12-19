package amortization

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)


// service implements the Service interface
type service struct {
	repo             Repository
	cache            CacheRepositoryInterface
	calculator       CalculatorInterface
	scheduler        SchedulerInterface
	eventPublisher   EventPublisherInterface
	auditLogger      AuditLoggerInterface
	ciService        CIServiceInterface
	lifecycleService LifecycleServiceInterface
	logger           *pustakaLogger.Logger
}

// NewAmortizationService creates a new amortization service
func NewAmortizationService(
	repo Repository,
	cache CacheRepositoryInterface,
	calculator CalculatorInterface,
	scheduler SchedulerInterface,
	eventPublisher EventPublisherInterface,
	auditLogger AuditLoggerInterface,
	ciService CIServiceInterface,
	lifecycleService LifecycleServiceInterface,
	logger *pustakaLogger.Logger,
) Service {
	return &service{
		repo:             repo,
		cache:            cache,
		calculator:       calculator,
		scheduler:        scheduler,
		eventPublisher:   eventPublisher,
		auditLogger:      auditLogger,
		ciService:        ciService,
		lifecycleService: lifecycleService,
		logger:           logger,
	}
}

// GetAmortizationDetails retrieves amortization details for a CI
func (s *service) GetAmortizationDetails(ctx context.Context, ciID uuid.UUID) (*AmortizationDetails, error) {
	// Try cache first
	if ci, err := s.cache.GetAmortizableCI(ctx, ciID); err == nil && ci != nil {
		return s.enrichWithCalculatedFields(ctx, ci), nil
	}

	// Fetch from repository
	ci, err := s.repo.GetAmortizableCI(ctx, ciID)
	if err != nil {
		return nil, fmt.Errorf("failed to get amortization details: %w", err)
	}

	// Cache the result
	s.cache.SetAmortizableCI(ctx, ciID, ci, 5*time.Minute)

	return s.enrichWithCalculatedFields(ctx, ci), nil
}

// UpdateAmortizationConfig updates amortization configuration for a CI
func (s *service) UpdateAmortizationConfig(ctx context.Context, ciID uuid.UUID, req *UpdateAmortizationConfig, userID uuid.UUID) (*AmortizationDetails, error) {
	// Get current CI configuration
	currentCI, err := s.repo.GetAmortizableCI(ctx, ciID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current CI configuration: %w", err)
	}

	// Prepare updates
	updates := &AmortizationConfigUpdates{
		PurchaseCost:        req.PurchaseCost,
		SalvageValue:        req.SalvageValue,
		AmortStartDate:      req.AmortStartDate,
		UsefulLifeMonths:    req.UsefulLifeMonths,
		UpdatedBy:           &userID,
		UpdatedAt:           &[]time.Time{time.Now()}[0],
	}

	// Calculate new book value if purchase cost changed
	if req.PurchaseCost != nil && currentCI.PurchaseCost != *req.PurchaseCost {
		updates.CurrentBookValue = req.PurchaseCost
		zero := 0.0
		updates.AccumulatedDepreciation = &zero
	}

	// Use transaction for consistency
	err = s.repo.WithTransaction(ctx, func(ctx context.Context, tx interface{}) error {
		// Update CI configuration
		return s.repo.UpdateAmortizationConfig(ctx, ciID, updates)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update amortization configuration: %w", err)
	}

	// Invalidate cache
	s.cache.InvalidateAmortizableCI(ctx, ciID)

	// Publish event
	event := &AmortizationConfigUpdatedEvent{
		CIID:      ciID,
		UserID:    userID,
		Changes:   s.buildChangesMap(currentCI, req),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	s.eventPublisher.PublishAmortizationConfigUpdated(ctx, event)

	// Log audit
	s.auditLogger.LogAmortizationConfigUpdated(ctx, ciID, userID, s.buildChangesMap(currentCI, req))

	// Get updated details
	return s.GetAmortizationDetails(ctx, ciID)
}

// ListAmortizableCIs retrieves a paginated list of amortizable CIs
func (s *service) ListAmortizableCIs(ctx context.Context, filters *AmortizableCIFilters) (*AmortizationCIList, error) {
	result, err := s.repo.ListAmortizableCIs(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list amortizable CIs: %w", err)
	}

	return result, nil
}

// GetLedgerEntries retrieves ledger entries with filtering
func (s *service) GetLedgerEntries(ctx context.Context, filters *LedgerFilters) (*LedgerEntryList, error) {
	result, err := s.repo.GetLedgerEntries(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger entries: %w", err)
	}

	return result, nil
}

// GetLedgerEntry retrieves a specific ledger entry
func (s *service) GetLedgerEntry(ctx context.Context, entryID uuid.UUID) (*LedgerEntry, error) {
	entry, err := s.repo.GetLedgerEntry(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger entry: %w", err)
	}

	return entry, nil
}

// CreateAdjustment creates a manual adjustment entry
func (s *service) CreateAdjustment(ctx context.Context, req *CreateAdjustmentRequest, userID uuid.UUID) (*LedgerEntry, error) {
	// Get current CI
	ci, err := s.repo.GetAmortizableCI(ctx, req.CIID)
	if err != nil {
		return nil, fmt.Errorf("failed to get CI for adjustment: %w", err)
	}

	// Set effective date if not provided
	effectiveDate := time.Now()
	if req.EffectiveDate != nil {
		effectiveDate = *req.EffectiveDate
	}

	// Create adjustment entry
	entry := &LedgerEntry{
		ID:                         uuid.New(),
		CIID:                       req.CIID,
		EntryType:                  "adjustment",
		EntryDate:                  effectiveDate,
		Amount:                     req.Amount,
		BookValueBefore:            ci.CurrentBookValue,
		BookValueAfter:             ci.CurrentBookValue + req.Amount,
		AccumulatedDepreciation: ci.AccumulatedDepreciation,
		Description:                &req.Description,
		CreatedAt:                  time.Now(),
		CreatedBy:                  &userID,
	}

	// Save entry
	if err := s.repo.CreateLedgerEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create adjustment entry: %w", err)
	}

	// Update CI book value
	updates := &AmortizationConfigUpdates{
		CurrentBookValue: &entry.BookValueAfter,
		UpdatedBy:        &userID,
		UpdatedAt:        &[]time.Time{time.Now()}[0],
	}

	if err := s.repo.UpdateAmortizationConfig(ctx, req.CIID, updates); err != nil {
		s.logger.Warn().Err(err).Str("ci_id", req.CIID.String()).Msg("Failed to update CI book value after adjustment")
	}

	// Invalidate cache
	s.cache.InvalidateAmortizableCI(ctx, req.CIID)

	// Publish event
	event := &AmortizationAdjustmentCreatedEvent{
		CIID:         req.CIID,
		UserID:       userID,
		AdjustmentID: entry.ID,
		Amount:       req.Amount,
		Description:  req.Description,
		Timestamp:    time.Now().Format(time.RFC3339),
	}
	s.eventPublisher.PublishAmortizationAdjustmentCreated(ctx, event)

	// Log audit
	s.auditLogger.LogAmortizationAdjustmentCreated(ctx, req.CIID, userID, req.Amount, req.Description)

	return entry, nil
}

// ProcessDailyAmortization processes daily amortization for all eligible CIs
func (s *service) ProcessDailyAmortization(ctx context.Context, processingDate time.Time) (*AmortizationRun, error) {
	return s.scheduler.ExecuteScheduledRun(ctx, processingDate)
}

// TriggerManualRun triggers a manual amortization run
func (s *service) TriggerManualRun(ctx context.Context, req *ManualRunRequest, userID uuid.UUID) (*AmortizationRun, error) {
	return s.scheduler.ExecuteManualRun(ctx, req, userID)
}

// GetAmortizationRun retrieves a specific amortization run
func (s *service) GetAmortizationRun(ctx context.Context, runID uuid.UUID) (*AmortizationRun, error) {
	run, err := s.repo.GetAmortizationRun(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to get amortization run: %w", err)
	}

	return run, nil
}

// ListAmortizationRuns retrieves amortization runs with filtering
func (s *service) ListAmortizationRuns(ctx context.Context, filters *AmortizationRunFilters) (*AmortizationRunList, error) {
	result, err := s.repo.ListAmortizationRuns(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list amortization runs: %w", err)
	}

	return result, nil
}

// GetAmortizationSummaries retrieves amortization summaries for reporting
func (s *service) GetAmortizationSummaries(ctx context.Context, req *SummaryRequest) (*AmortizationSummary, error) {
	result, err := s.repo.GetAmortizationSummaries(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get amortization summaries: %w", err)
	}

	return result, nil
}

// GenerateDepreciationSchedule generates a depreciation schedule
func (s *service) GenerateDepreciationSchedule(ctx context.Context, req *DepreciationScheduleRequest) (*DepreciationSchedule, error) {
	result := &DepreciationSchedule{
		ReportID: uuid.New(),
		DateRange: DepreciationScheduleRange{
			StartDate: req.DateFrom,
			EndDate:   req.DateTo,
		},
		Schedule: []DepreciationScheduleEntry{},
	}

	// Get schedule data from repository
	entries, err := s.repo.GetDepreciationScheduleData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate depreciation schedule: %w", err)
	}

	result.Schedule = entries

	return result, nil
}

// HandleTerminalStatusChange handles terminal status changes from CI lifecycle
func (s *service) HandleTerminalStatusChange(ctx context.Context, ciID uuid.UUID, oldStatusID, newStatusID uuid.UUID, userID uuid.UUID) error {
	// Get new lifecycle status to check amortization behavior
	newStatus, err := s.lifecycleService.GetLifecycleStatus(ctx, newStatusID)
	if err != nil {
		return fmt.Errorf("failed to get new lifecycle status: %w", err)
	}

	// Check if new status has terminal behavior
	if newStatus.AmortizationBehavior != "terminal" {
		return nil // No action needed for non-terminal status
	}

	// Get CI details
	ci, err := s.repo.GetAmortizableCI(ctx, ciID)
	if err != nil {
		return fmt.Errorf("failed to get CI for write-off: %w", err)
	}

	// Check if CI is amortizable
	if !ci.IsAmortizable {
		return nil // No amortization to handle
	}

	// Check if CI already has book value at salvage value
	if ci.CurrentBookValue <= ci.SalvageValue {
		return nil // Already at salvage value, no write-off needed
	}

	// Calculate and process write-off
	writeOffCalc, err := s.calculator.CalculateWriteOff(ctx, ci, time.Now())
	if err != nil {
		return fmt.Errorf("failed to calculate write-off: %w", err)
	}

	// Create write-off ledger entry
	writeOffEntry := &LedgerEntry{
		ID:                         uuid.New(),
		CIID:                       ciID,
		EntryType:                  "write_off",
		EntryDate:                  writeOffCalc.WriteOffDate,
		Amount:                     writeOffCalc.WriteOffAmount,
		BookValueBefore:            writeOffCalc.BookValueBefore,
		BookValueAfter:             writeOffCalc.BookValueAfter,
		AccumulatedDepreciation: writeOffCalc.AccumulatedDepreciationAfter,
		Description:                stringPtr(fmt.Sprintf("Automatic write-off due to terminal status: %s", newStatus.Name)),
		CreatedAt:                  time.Now(),
		CreatedBy:                  &userID,
	}

	// Save write-off entry
	if err := s.repo.CreateLedgerEntry(ctx, writeOffEntry); err != nil {
		return fmt.Errorf("failed to create write-off entry: %w", err)
	}

	// Update CI book value
	updates := &AmortizationConfigUpdates{
		CurrentBookValue: &writeOffCalc.BookValueAfter,
		UpdatedBy:        &userID,
		UpdatedAt:        &[]time.Time{time.Now()}[0],
	}

	if err := s.repo.UpdateAmortizationConfig(ctx, ciID, updates); err != nil {
		s.logger.Warn().Err(err).Str("ci_id", ciID.String()).Msg("Failed to update CI book value after write-off")
	}

	// Invalidate cache
	s.cache.InvalidateAmortizableCI(ctx, ciID)

	// Publish event
	event := &AmortizationWrittenOffEvent{
		CIID:          ciID,
		UserID:        userID,
		WriteOffAmount: writeOffCalc.WriteOffAmount,
		Reason:        writeOffCalc.Reason,
		Timestamp:     time.Now().Format(time.RFC3339),
	}
	s.eventPublisher.PublishAmortizationWrittenOff(ctx, event)

	// Log audit
	s.auditLogger.LogAmortizationWriteOff(ctx, ciID, userID, writeOffCalc.WriteOffAmount, writeOffCalc.Reason)

	s.logger.Info().
		Str("ci_id", ciID.String()).
		Str("old_status", oldStatusID.String()).
		Str("new_status", newStatusID.String()).
		Float64("write_off_amount", writeOffCalc.WriteOffAmount).
		Msg("Handled terminal status change")

	return nil
}

// RecalculateAmortization recalculates amortization for a CI
func (s *service) RecalculateAmortization(ctx context.Context, ciID uuid.UUID, userID uuid.UUID) error {
	// Get current CI
	ci, err := s.repo.GetAmortizableCI(ctx, ciID)
	if err != nil {
		return fmt.Errorf("failed to get CI for recalculation: %w", err)
	}

	// Calculate depreciation from start date to today
	calculations, err := s.calculator.CalculateDepreciationToDate(ctx, ci, time.Now())
	if err != nil {
		return fmt.Errorf("failed to calculate depreciation to date: %w", err)
	}

	s.logger.Info().
		Str("ci_id", ciID.String()).
		Int("calculations", len(calculations)).
		Str("user_id", userID.String()).
		Msg("Recalculated amortization")

	return nil
}

// Helper methods

func (s *service) enrichWithCalculatedFields(ctx context.Context, ci *AmortizableCI) *AmortizationDetails {
	details := &AmortizationDetails{
		AmortizableCI: *ci,
	}

	// Calculate monthly depreciation
	if ci.UsefulLifeMonths > 0 {
		monthlyDep := (ci.PurchaseCost - ci.SalvageValue) / float64(ci.UsefulLifeMonths)
		details.MonthlyDepreciation = &monthlyDep
	}

	// Calculate remaining life
	if remainingLife, err := s.calculator.CalculateRemainingLife(ctx, ci, time.Now()); err == nil {
		details.RemainingLifeMonths = &remainingLife
	}

	return details
}

func (s *service) buildChangesMap(currentCI *AmortizableCI, req *UpdateAmortizationConfig) map[string]interface{} {
	changes := make(map[string]interface{})

	if req.PurchaseCost != nil && currentCI.PurchaseCost != *req.PurchaseCost {
		changes["purchase_cost"] = map[string]interface{}{
			"old": currentCI.PurchaseCost,
			"new": *req.PurchaseCost,
		}
	}

	if req.SalvageValue != nil && currentCI.SalvageValue != *req.SalvageValue {
		changes["salvage_value"] = map[string]interface{}{
			"old": currentCI.SalvageValue,
			"new": *req.SalvageValue,
		}
	}

	if req.UsefulLifeMonths != nil && currentCI.UsefulLifeMonths != *req.UsefulLifeMonths {
		changes["useful_life_months"] = map[string]interface{}{
			"old": currentCI.UsefulLifeMonths,
			"new": *req.UsefulLifeMonths,
		}
	}

	return changes
}

