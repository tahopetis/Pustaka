package amortization

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// schedulerSimple implements the SchedulerInterface
type schedulerSimple struct {
	repo      Repository
	calculator CalculatorInterface
	logger     *pustakaLogger.Logger
}

// NewScheduler creates a new amortization scheduler
func NewScheduler(repo Repository, cache CacheRepositoryInterface, logger *pustakaLogger.Logger) SchedulerInterface {
	calculator := NewCalculator(logger)
	return &schedulerSimple{
		repo:      repo,
		calculator: calculator,
		logger:     logger,
	}
}

// ScheduleDailyRun schedules the daily amortization run
func (s *schedulerSimple) ScheduleDailyRun(ctx context.Context) error {
	s.logger.Info().Msg("Starting daily amortization scheduler")
	return nil
}

// UnscheduleDailyRun stops the daily scheduling
func (s *schedulerSimple) UnscheduleDailyRun() error {
	s.logger.Info().Msg("Stopping daily amortization scheduler")
	return nil
}

// ExecuteScheduledRun executes a scheduled amortization run
func (s *schedulerSimple) ExecuteScheduledRun(ctx context.Context, processingDate time.Time) (*AmortizationRun, error) {
	s.logger.Info().Time("processing_date", processingDate).Msg("Executing scheduled amortization run")

	// Create amortization run
	runID := uuid.New()
	now := time.Now()
	run := &AmortizationRun{
		ID:                 runID,
		Status:             "started",
		ProcessingDate:     processingDate,
		StartedAt:          &now,
		TotalAmortizableCIs: 0,
		IsManual:           false,
		CreatedAt:          now,
	}

	// Save run to database
	if err := s.repo.CreateAmortizationRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to create amortization run: %w", err)
	}

	// Get CIs to process
	ciIDs, err := s.repo.GetCIsForProcessing(ctx, processingDate, 1000) // Process in batches of 1000
	if err != nil {
		return s.markRunFailed(ctx, run, fmt.Sprintf("Failed to get CIs for processing: %v", err))
	}

	run.TotalAmortizableCIs = len(ciIDs)
	totalDepreciation := 0.0
	processed := 0
	failed := 0
	skipped := 0

	// Process each CI
	for _, ciID := range ciIDs {
		result, err := s.processCIForAmortization(ctx, ciID, processingDate, run.ID, false)
		if err != nil {
			failed++
			s.logger.Warn().Err(err).Str("ci_id", ciID.String()).Msg("Failed to process CI for amortization")
			continue
		}

		switch result.Status {
		case "processed":
			processed++
			if result.DepreciationAmount > 0 {
				totalDepreciation += result.DepreciationAmount
			}
		case "skipped":
			skipped++
		case "failed":
			failed++
		}
	}

	// Complete the run
	completedAt := time.Now()
	status := "completed"
	if failed > 0 {
		status = "partial"
	}

	updates := &AmortizationRunUpdates{
		Status:            &status,
		CompletedAt:       &completedAt,
		ProcessedCIs:      &processed,
		FailedCIs:         &failed,
		SkippedCIs:        &skipped,
		TotalDepreciation: &totalDepreciation,
	}

	if err := s.repo.UpdateAmortizationRun(ctx, runID, updates); err != nil {
		s.logger.Warn().Err(err).Str("run_id", runID.String()).Msg("Failed to update run completion status")
	}

	// Get updated run
	completedRun, err := s.repo.GetAmortizationRun(ctx, runID)
	if err != nil {
		s.logger.Warn().Err(err).Str("run_id", runID.String()).Msg("Failed to get completed run")
		return run, nil // Return the original run if we can't get the updated one
	}

	s.logger.Info().
		Str("run_id", runID.String()).
		Time("processing_date", processingDate).
		Int("total_cis", run.TotalAmortizableCIs).
		Int("processed", processed).
		Int("failed", failed).
		Int("skipped", skipped).
		Float64("total_depreciation", totalDepreciation).
		Msg("Completed scheduled amortization run")

	return completedRun, nil
}

// ExecuteManualRun executes a manual amortization run
func (s *schedulerSimple) ExecuteManualRun(ctx context.Context, req *ManualRunRequest, userID uuid.UUID) (*AmortizationRun, error) {
	s.logger.Info().
		Str("user_id", userID.String()).
		Bool("dry_run", req.DryRun).
		Msg("Executing manual amortization run")

	// Determine processing date
	processingDate := time.Now()
	if req.DateOverride != nil {
		processingDate = *req.DateOverride
	}

	// Create manual run
	runID := uuid.New()
	now := time.Now()
	run := &AmortizationRun{
		ID:                 runID,
		Status:             "started",
		ProcessingDate:     processingDate,
		StartedAt:          &now,
		TotalAmortizableCIs: 0,
		IsManual:           true,
		DryRun:             req.DryRun,
		TriggeredBy:        &userID,
		CreatedAt:          now,
	}

	// Save run to database
	if err := s.repo.CreateAmortizationRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to create manual amortization run: %w", err)
	}

	// If specific CIs are provided, process only those
	ciIDs := req.CIIDs
	if len(ciIDs) == 0 {
		// Get all CIs to process
		var err error
		ciIDs, err = s.repo.GetCIsForProcessing(ctx, processingDate, 1000)
		if err != nil {
			return s.markRunFailed(ctx, run, fmt.Sprintf("Failed to get CIs for processing: %v", err))
		}
	}

	run.TotalAmortizableCIs = len(ciIDs)
	totalDepreciation := 0.0
	processed := 0
	failed := 0
	skipped := 0

	// Process each CI
	for _, ciID := range ciIDs {
		result, err := s.processCIForAmortization(ctx, ciID, processingDate, runID, req.DryRun)
		if err != nil {
			failed++
			s.logger.Warn().Err(err).Str("ci_id", ciID.String()).Msg("Failed to process CI for amortization")
			continue
		}

		switch result.Status {
		case "processed":
			processed++
			if result.DepreciationAmount > 0 {
				totalDepreciation += result.DepreciationAmount
			}
		case "skipped":
			skipped++
		case "failed":
			failed++
		}
	}

	// Complete the run
	completedAt := time.Now()
	status := "completed"
	if failed > 0 {
		status = "partial"
	}

	updates := &AmortizationRunUpdates{
		Status:            &status,
		CompletedAt:       &completedAt,
		ProcessedCIs:      &processed,
		FailedCIs:         &failed,
		SkippedCIs:        &skipped,
		TotalDepreciation: &totalDepreciation,
	}

	if err := s.repo.UpdateAmortizationRun(ctx, runID, updates); err != nil {
		s.logger.Warn().Err(err).Str("run_id", runID.String()).Msg("Failed to update run completion status")
	}

	// Get updated run
	completedRun, err := s.repo.GetAmortizationRun(ctx, runID)
	if err != nil {
		s.logger.Warn().Err(err).Str("run_id", runID.String()).Msg("Failed to get completed run")
		return run, nil
	}

	s.logger.Info().
		Str("run_id", runID.String()).
		Str("user_id", userID.String()).
		Bool("dry_run", req.DryRun).
		Time("processing_date", processingDate).
		Int("total_cis", run.TotalAmortizableCIs).
		Int("processed", processed).
		Int("failed", failed).
		Int("skipped", skipped).
		Float64("total_depreciation", totalDepreciation).
		Msg("Completed manual amortization run")

	return completedRun, nil
}

// Helper methods

func (s *schedulerSimple) processCIForAmortization(ctx context.Context, ciID uuid.UUID, processingDate time.Time, runID uuid.UUID, dryRun bool) (*ProcessingResult, error) {
	// Get CI
	ci, err := s.repo.GetAmortizableCI(ctx, ciID)
	if err != nil {
		return &ProcessingResult{
			CIID:        ciID,
			Status:      "failed",
			ErrorMessage: stringPtr(fmt.Sprintf("Failed to get CI: %v", err)),
			ProcessedAt:  time.Now(),
		}, nil
	}

	// Check if CI should be processed
	if !s.shouldProcessCI(ci, processingDate) {
		return &ProcessingResult{
			CIID:        ciID,
			Status:      "skipped",
			ProcessedAt: time.Now(),
		}, nil
	}

	// Check if this CI has existing ledger entries
	hasExistingEntries, err := s.repo.HasLedgerEntriesForCI(ctx, ciID)
	if err != nil {
		return &ProcessingResult{
			CIID:        ciID,
			Status:      "failed",
			ErrorMessage: stringPtr(fmt.Sprintf("Failed to check ledger entries: %v", err)),
			ProcessedAt:  time.Now(),
		}, nil
	}

	// Calculate depreciation
	var depreciationAmount float64
	var bookValueBefore, bookValueAfter, accumulatedDepreciationAfter float64
	var entryType string

	// Check if this is a terminal status asset that needs write-off
	if ci.LifecycleStatus != nil && ci.LifecycleStatus.AmortizationBehavior == "terminal" {
		// Handle terminal status write-off
		writeOffCalc, err := s.calculateWriteOff(ctx, ci, processingDate)
		if err != nil {
			return &ProcessingResult{
				CIID:        ciID,
				Status:      "failed",
				ErrorMessage: stringPtr(fmt.Sprintf("Write-off calculation failed: %v", err)),
				ProcessedAt:  time.Now(),
			}, nil
		}

		depreciationAmount = writeOffCalc.WriteOffAmount
		bookValueBefore = writeOffCalc.BookValueBefore
		bookValueAfter = writeOffCalc.BookValueAfter
		accumulatedDepreciationAfter = writeOffCalc.AccumulatedDepreciationAfter
		entryType = "write_off"
	} else if !hasExistingEntries {
		// Use catch-up depreciation for first-time processing
		catchUpCalc, err := s.calculator.CalculateCatchUpDepreciation(ctx, ci, processingDate)
		if err != nil {
			return &ProcessingResult{
				CIID:        ciID,
				Status:      "failed",
				ErrorMessage: stringPtr(fmt.Sprintf("Catch-up calculation failed: %v", err)),
				ProcessedAt:  time.Now(),
			}, nil
		}

		depreciationAmount = catchUpCalc.TotalDepreciationAmount
		bookValueBefore = catchUpCalc.BookValueBefore
		bookValueAfter = catchUpCalc.BookValueAfter
		accumulatedDepreciationAfter = catchUpCalc.AccumulatedDepreciationAfter
		entryType = "catch_up_depreciation"
	} else {
		// Use regular monthly depreciation
		calculation, err := s.calculateDepreciation(ci, processingDate)
		if err != nil {
			return &ProcessingResult{
				CIID:        ciID,
				Status:      "failed",
				ErrorMessage: stringPtr(fmt.Sprintf("Calculation failed: %v", err)),
				ProcessedAt:  time.Now(),
			}, nil
		}

		depreciationAmount = calculation.Amount
		bookValueBefore = calculation.BookValueBefore
		bookValueAfter = calculation.BookValueAfter
		accumulatedDepreciationAfter = calculation.AccumulatedDepreciationAfter
		entryType = "monthly_depreciation"
	}

	// If this is a dry run, just return the calculation
	if dryRun {
		return &ProcessingResult{
			CIID:              ciID,
			Status:            "processed",
			DepreciationAmount: depreciationAmount,
			ProcessedAt:       time.Now(),
		}, nil
	}

	// Create description based on entry type
	var description *string
	switch entryType {
	case "write_off":
		description = stringPtr(fmt.Sprintf("Automatic write-off for terminal status (%s)", ci.LifecycleStatus.DisplayName))
	case "catch_up_depreciation":
		description = stringPtr("Catch-up depreciation for amortization start date")
	case "monthly_depreciation":
		description = stringPtr("Monthly depreciation")
	}

	// Create ledger entry (skip if dry run)
	entry := &LedgerEntry{
		ID:                         uuid.New(),
		CIID:                       ciID,
		EntryType:                  entryType,
		EntryDate:                  processingDate,
		Description:                description,
		Amount:                     depreciationAmount,
		BookValueBefore:            bookValueBefore,
		BookValueAfter:             bookValueAfter,
		AccumulatedDepreciation: accumulatedDepreciationAfter,
		AmortizationRunID:          &runID,
		CreatedAt:                  time.Now(),
		CreatedBy:                  func() *uuid.UUID { id := uuid.MustParse("fd13d040-48a4-45c1-b7fa-1e71b20a29de"); return &id }(), // Admin user for now
	}

	if err := s.repo.CreateLedgerEntry(ctx, entry); err != nil {
		return &ProcessingResult{
			CIID:        ciID,
			Status:      "failed",
			ErrorMessage: stringPtr(fmt.Sprintf("Failed to create ledger entry: %v", err)),
			ProcessedAt:  time.Now(),
		}, nil
	}

	// Update CI book value
	updates := &AmortizationConfigUpdates{
		CurrentBookValue:           &bookValueAfter,
		AccumulatedDepreciation:    &accumulatedDepreciationAfter,
		UpdatedBy:                 func() *uuid.UUID { id := uuid.MustParse("fd13d040-48a4-45c1-b7fa-1e71b20a29de"); return &id }(), // Admin user for now
		UpdatedAt:                 func(t time.Time) *time.Time { return &t }(time.Now()),
	}

	if err := s.repo.UpdateAmortizationConfig(ctx, ciID, updates); err != nil {
		s.logger.Warn().Err(err).Str("ci_id", ciID.String()).Msg("Failed to update CI book value after depreciation")
	}

	return &ProcessingResult{
		CIID:              ciID,
		Status:            "processed",
		DepreciationAmount: depreciationAmount,
		ProcessedAt:       time.Now(),
	}, nil
}

func (s *schedulerSimple) shouldProcessCI(ci *AmortizableCI, processingDate time.Time) bool {
	// Check if amortization has started
	if ci.AmortStartDate == nil || processingDate.Before(*ci.AmortStartDate) {
		return false
	}

	// Check if CI has remaining book value
	if ci.CurrentBookValue <= ci.SalvageValue {
		return false
	}

	// Check amortization behavior
	if ci.LifecycleStatus != nil {
		switch ci.LifecycleStatus.AmortizationBehavior {
		case "pending":
			return false
		case "terminal":
			// For terminal status, only process if there's remaining book value to write off
			return ci.CurrentBookValue > ci.SalvageValue
		case "active":
			return true
		}
	}

	// Default to processing if behavior is not explicitly set
	return true
}

func (s *schedulerSimple) calculateDepreciation(ci *AmortizableCI, processingDate time.Time) (*DepreciationCalculation, error) {
	// Simplified calculation - monthly depreciation
	if ci.UsefulLifeMonths <= 0 {
		return nil, fmt.Errorf("invalid useful life months")
	}

	monthlyDepreciation := (ci.PurchaseCost - ci.SalvageValue) / float64(ci.UsefulLifeMonths)

	bookValueAfter := ci.CurrentBookValue - monthlyDepreciation
	if bookValueAfter < ci.SalvageValue {
		monthlyDepreciation = ci.CurrentBookValue - ci.SalvageValue
		bookValueAfter = ci.SalvageValue
	}

	return &DepreciationCalculation{
		Amount:                     monthlyDepreciation,
		BookValueBefore:            ci.CurrentBookValue,
		BookValueAfter:             bookValueAfter,
		AccumulatedDepreciationBefore: ci.AccumulatedDepreciation,
		AccumulatedDepreciationAfter:  ci.AccumulatedDepreciation + monthlyDepreciation,
		CalculationDate:            processingDate,
	}, nil
}

func (s *schedulerSimple) calculateWriteOff(ctx context.Context, ci *AmortizableCI, processingDate time.Time) (*WriteOffCalculation, error) {
	// Write-off calculation: write off all remaining book value above salvage value
	if ci.CurrentBookValue <= ci.SalvageValue {
		return nil, fmt.Errorf("no remaining book value to write off")
	}

	writeOffAmount := ci.CurrentBookValue - ci.SalvageValue

	return &WriteOffCalculation{
		WriteOffAmount:              writeOffAmount,
		BookValueBefore:             ci.CurrentBookValue,
		BookValueAfter:              ci.SalvageValue,
		AccumulatedDepreciationBefore: ci.AccumulatedDepreciation,
		AccumulatedDepreciationAfter:  ci.AccumulatedDepreciation + writeOffAmount,
		WriteOffDate:                processingDate,
		Reason:                      fmt.Sprintf("Terminal status: %s", ci.LifecycleStatus.DisplayName),
	}, nil
}

func (s *schedulerSimple) markRunFailed(ctx context.Context, run *AmortizationRun, errorMessage string) (*AmortizationRun, error) {
	status := "failed"
	completedAt := time.Now()

	updates := &AmortizationRunUpdates{
		Status:       &status,
		CompletedAt:  &completedAt,
		ErrorSummary: &errorMessage,
	}

	if err := s.repo.UpdateAmortizationRun(ctx, run.ID, updates); err != nil {
		return nil, fmt.Errorf("failed to mark run as failed: %w", err)
	}

	// Get updated run
	failedRun, err := s.repo.GetAmortizationRun(ctx, run.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed run: %w", err)
	}

	return failedRun, nil
}

