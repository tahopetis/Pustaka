package amortization

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// calculatorSimple implements the CalculatorInterface
type calculatorSimple struct {
	logger *pustakaLogger.Logger
}

// NewCalculator creates a new depreciation calculator
func NewCalculator(logger *pustakaLogger.Logger) CalculatorInterface {
	return &calculatorSimple{
		logger: logger,
	}
}

// CalculateMonthlyDepreciation calculates monthly depreciation using straight-line method
func (c *calculatorSimple) CalculateMonthlyDepreciation(ctx context.Context, ci *AmortizableCI, entryDate time.Time) (*DepreciationCalculation, error) {
	// Validation
	if ci.PurchaseCost <= 0 {
		return nil, fmt.Errorf("invalid purchase cost")
	}
	if ci.UsefulLifeMonths <= 0 {
		return nil, fmt.Errorf("invalid useful life months")
	}
	if ci.AmortStartDate == nil {
		return nil, fmt.Errorf("amortization start date is not set")
	}

	// Check if depreciation should be calculated
	if !c.shouldDepreciateForDate(ci, entryDate) {
		return nil, fmt.Errorf("CI does not require depreciation for date %s", entryDate.Format("2006-01-02"))
	}

	// Calculate remaining life
	remainingMonths := c.calculateRemainingLife(ci, entryDate)
	if remainingMonths <= 0 {
		return nil, fmt.Errorf("no remaining useful life for depreciation")
	}

	// Calculate monthly depreciation amount
	monthlyDepreciation := (ci.PurchaseCost - ci.SalvageValue) / float64(ci.UsefulLifeMonths)

	// Ensure we don't depreciate below salvage value
	bookValueBefore := ci.CurrentBookValue
	if bookValueBefore <= ci.SalvageValue {
		return nil, fmt.Errorf("book value (%.2f) is already at or below salvage value (%.2f)",
			bookValueBefore, ci.SalvageValue)
	}

	// Calculate depreciation amount for this period
	depreciationAmount := monthlyDepreciation

	// Check if this is the final depreciation
	bookValueAfter := bookValueBefore - depreciationAmount
	if bookValueAfter < ci.SalvageValue {
		// Adjust final depreciation to hit salvage value exactly
		depreciationAmount = bookValueBefore - ci.SalvageValue
		bookValueAfter = ci.SalvageValue
	}

	// New accumulated depreciation
	newAccumulatedDepreciation := ci.AccumulatedDepreciation + depreciationAmount

	calculation := &DepreciationCalculation{
		Amount:                     depreciationAmount,
		BookValueBefore:            bookValueBefore,
		BookValueAfter:             bookValueAfter,
		AccumulatedDepreciationBefore: ci.AccumulatedDepreciation,
		AccumulatedDepreciationAfter:  newAccumulatedDepreciation,
		CalculationDate:            entryDate,
		Method:                     "straight_line",
	}

	// Fixed logger call
	c.logger.Info().
		Str("ci_id", ci.ID.String()).
		Time("entry_date", entryDate).
		Float64("depreciation_amount", depreciationAmount).
		Float64("book_value_before", bookValueBefore).
		Float64("book_value_after", bookValueAfter).
		Int("remaining_months", remainingMonths).
		Msg("Calculated monthly depreciation")

	return calculation, nil
}

// CalculateCatchUpDepreciation calculates all missed depreciation from start date to current date
func (c *calculatorSimple) CalculateCatchUpDepreciation(ctx context.Context, ci *AmortizableCI, asOfDate time.Time) (*CatchUpDepreciationCalculation, error) {
	// Validation
	if ci.PurchaseCost <= 0 {
		return nil, fmt.Errorf("invalid purchase cost")
	}
	if ci.UsefulLifeMonths <= 0 {
		return nil, fmt.Errorf("invalid useful life months")
	}
	if ci.AmortStartDate == nil {
		return nil, fmt.Errorf("amortization start date is not set")
	}

	// Check if depreciation should be calculated
	if !c.shouldDepreciateForDate(ci, asOfDate) {
		return nil, fmt.Errorf("CI does not require depreciation as of %s", asOfDate.Format("2006-01-02"))
	}

	// Calculate monthly depreciation amount
	monthlyDepreciation := (ci.PurchaseCost - ci.SalvageValue) / float64(ci.UsefulLifeMonths)

	// Calculate how many months of depreciation should have been calculated
	monthsSinceStart := c.calculateElapsedMonths(*ci.AmortStartDate, asOfDate)

	// Don't exceed useful life
	maxDepreciationMonths := ci.UsefulLifeMonths
	if monthsSinceStart > maxDepreciationMonths {
		monthsSinceStart = maxDepreciationMonths
	}

	// Calculate total catch-up depreciation
	totalCatchUpDepreciation := float64(monthsSinceStart) * monthlyDepreciation

	// Ensure we don't depreciate below salvage value
	maxAllowableDepreciation := ci.PurchaseCost - ci.SalvageValue
	if totalCatchUpDepreciation > maxAllowableDepreciation {
		totalCatchUpDepreciation = maxAllowableDepreciation
	}

	// Calculate current book value after catch-up
	bookValueAfter := ci.PurchaseCost - totalCatchUpDepreciation
	if bookValueAfter < ci.SalvageValue {
		bookValueAfter = ci.SalvageValue
	}

	// New accumulated depreciation
	newAccumulatedDepreciation := totalCatchUpDepreciation

	calculation := &CatchUpDepreciationCalculation{
		MonthsDepreciated:          monthsSinceStart,
		TotalDepreciationAmount:    totalCatchUpDepreciation,
		MonthlyDepreciationAmount:  monthlyDepreciation,
		BookValueBefore:            ci.PurchaseCost,
		BookValueAfter:             bookValueAfter,
		AccumulatedDepreciationAfter: newAccumulatedDepreciation,
		CalculationDate:            asOfDate,
		Method:                     "straight_line_catch_up",
	}

	c.logger.Info().
		Str("ci_id", ci.ID.String()).
		Time("start_date", *ci.AmortStartDate).
		Time("as_of_date", asOfDate).
		Int("months_depreciated", monthsSinceStart).
		Float64("total_depreciation", totalCatchUpDepreciation).
		Float64("book_value_after", bookValueAfter).
		Msg("Calculated catch-up depreciation")

	return calculation, nil
}

// CalculateWriteOff calculates the write-off amount for an asset
func (c *calculatorSimple) CalculateWriteOff(ctx context.Context, ci *AmortizableCI, writeOffDate time.Time) (*WriteOffCalculation, error) {
	if ci.PurchaseCost <= 0 {
		return nil, fmt.Errorf("invalid purchase cost")
	}

	// Calculate write-off amount
	bookValueBefore := ci.CurrentBookValue
	writeOffAmount := bookValueBefore - ci.SalvageValue

	// Ensure write-off amount is not negative
	if writeOffAmount < 0 {
		writeOffAmount = 0
	}

	bookValueAfter := ci.SalvageValue

	calculation := &WriteOffCalculation{
		WriteOffAmount:              writeOffAmount,
		BookValueBefore:            bookValueBefore,
		BookValueAfter:             bookValueAfter,
		AccumulatedDepreciationBefore: ci.AccumulatedDepreciation,
		AccumulatedDepreciationAfter:  ci.AccumulatedDepreciation + writeOffAmount,
		WriteOffDate:               writeOffDate,
		Reason:                     "asset_write_off",
	}

	// Fixed logger call
	c.logger.Info().
		Str("ci_id", ci.ID.String()).
		Time("write_off_date", writeOffDate).
		Float64("write_off_amount", writeOffAmount).
		Float64("book_value_before", bookValueBefore).
		Float64("book_value_after", bookValueAfter).
		Msg("Calculated write-off")

	return calculation, nil
}

// CalculateDepreciationToDate calculates all depreciation up to a specific date
func (c *calculatorSimple) CalculateDepreciationToDate(ctx context.Context, ci *AmortizableCI, asOfDate time.Time) ([]DepreciationCalculation, error) {
	var calculations []DepreciationCalculation

	// Validation
	if ci.AmortStartDate == nil {
		return nil, fmt.Errorf("amortization start date is not set")
	}

	if asOfDate.Before(*ci.AmortStartDate) {
		return nil, fmt.Errorf("as-of date is before amortization start date")
	}

	// Generate depreciation calculations for each month from start date to as-of date
	currentDate := *ci.AmortStartDate
	currentCI := *ci // Work with a copy

	for currentDate.Before(asOfDate) || currentDate.Equal(asOfDate) {
		// Check if this month already has depreciation (shouldn't happen in new calculations)
		calc, err := c.CalculateMonthlyDepreciation(ctx, &currentCI, currentDate)
		if err != nil {
			if err.Error() == "no remaining useful life for depreciation" {
				break // Stop if we've reached the end of useful life
			}
			// Fixed logger call
			c.logger.Warn().
				Err(err).
				Str("ci_id", ci.ID.String()).
				Time("date", currentDate).
				Msg("Failed to calculate depreciation for month, skipping")
			continue
		}

		calculations = append(calculations, *calc)

		// Update the CI state for next calculation
		currentCI.CurrentBookValue = calc.BookValueAfter
		currentCI.AccumulatedDepreciation = calc.AccumulatedDepreciationAfter

		// Move to next month
		currentDate = currentDate.AddDate(0, 1, 0)
	}

	return calculations, nil
}

// CalculateRemainingLife calculates remaining useful life in months
func (c *calculatorSimple) CalculateRemainingLife(ctx context.Context, ci *AmortizableCI, asOfDate time.Time) (int, error) {
	if ci.AmortStartDate == nil {
		return 0, fmt.Errorf("amortization start date is not set")
	}

	// Calculate elapsed months from start date
	elapsedMonths := int(asOfDate.Sub(*ci.AmortStartDate).Hours() / 24 / 30)
	remainingMonths := ci.UsefulLifeMonths - elapsedMonths

	if remainingMonths < 0 {
		remainingMonths = 0
	}

	// Additional check based on current book value
	if ci.CurrentBookValue <= ci.SalvageValue {
		remainingMonths = 0
	}

	return remainingMonths, nil
}

// ValidateAdjustmentAmount validates an adjustment amount
func (c *calculatorSimple) ValidateAdjustmentAmount(ci *AmortizableCI, amount float64) error {
	if amount == 0 {
		return fmt.Errorf("adjustment amount cannot be zero")
	}

	// Check if adjustment would result in negative book value
	newBookValue := ci.CurrentBookValue + amount
	if newBookValue < 0 {
		return fmt.Errorf("adjustment amount (%.2f) would result in negative book value (%.2f)",
			amount, newBookValue)
	}

	// Check if adjustment exceeds reasonable limits (e.g., 50% of purchase cost)
	maxAdjustment := ci.PurchaseCost * 0.5
	if math.Abs(amount) > maxAdjustment {
		return fmt.Errorf("adjustment amount (%.2f) exceeds maximum allowed (%.2f)",
			math.Abs(amount), maxAdjustment)
	}

	return nil
}

// ProjectFutureValue projects the future value of an asset (NEW METHOD)
func (c *calculatorSimple) ProjectFutureValue(ctx context.Context, ci *AmortizableCI, projectionDate time.Time) (*ValueProjection, error) {
	// Validation
	if ci.AmortStartDate == nil {
		return nil, fmt.Errorf("amortization start date is not set")
	}

	if projectionDate.Before(*ci.AmortStartDate) {
		return nil, fmt.Errorf("projection date is before amortization start date")
	}

	// Calculate depreciation to projection date
	calculations, err := c.CalculateDepreciationToDate(ctx, ci, projectionDate)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate depreciation to projection date: %w", err)
	}

	var projectedBookValue float64
	var projectedDepreciation float64

	if len(calculations) == 0 {
		// No depreciation calculated, use current values
		projectedBookValue = ci.CurrentBookValue
		projectedDepreciation = ci.AccumulatedDepreciation
	} else {
		lastCalc := calculations[len(calculations)-1]
		projectedBookValue = lastCalc.BookValueAfter
		projectedDepreciation = lastCalc.AccumulatedDepreciationAfter
	}

	// Determine confidence level
	confidence := c.calculateConfidenceLevel(ci, projectionDate)

	// Generate assumptions
	assumptions := []string{
		"Straight-line depreciation method",
		"No changes to useful life or salvage value",
		"Asset remains in active amortization status",
	}

	if ci.AmortStartDate != nil {
		monthsElapsed := int(projectionDate.Sub(*ci.AmortStartDate).Hours() / 24 / 30)
		assumptions = append(assumptions, fmt.Sprintf("Projected over %d months", monthsElapsed))
	}

	projection := &ValueProjection{
		ProjectedDate:        projectionDate,
		ProjectedBookValue:   projectedBookValue,
		ProjectedDepreciation: projectedDepreciation,
		Confidence:          confidence,
		Assumptions:         assumptions,
	}

	return projection, nil
}

// ValidateAmortizationConfig validates amortization configuration
func (c *calculatorSimple) ValidateAmortizationConfig(ctx context.Context, config *AmortizationConfig) error {
	if config.PurchaseCost != nil && *config.PurchaseCost < 0 {
		return fmt.Errorf("purchase cost cannot be negative")
	}

	if config.SalvageValue != nil && *config.SalvageValue < 0 {
		return fmt.Errorf("salvage value cannot be negative")
	}

	if config.PurchaseCost != nil && config.SalvageValue != nil {
		if *config.SalvageValue > *config.PurchaseCost {
			return fmt.Errorf("salvage value cannot exceed purchase cost")
		}
	}

	if config.UsefulLifeMonths != nil && *config.UsefulLifeMonths <= 0 {
		return fmt.Errorf("useful life months must be positive")
	}

	return nil
}

// ReconstructHistoricalValues reconstructs historical book values for an asset
func (c *calculatorSimple) ReconstructHistoricalValues(ciID uuid.UUID, dateFrom, dateTo time.Time) ([]HistoricalValue, error) {
	// This would typically query the ledger entries and reconstruct historical values
	// For now, return an empty slice as this would require database access
	return []HistoricalValue{}, nil
}

// Helper methods

func (c *calculatorSimple) shouldDepreciateForDate(ci *AmortizableCI, entryDate time.Time) bool {
	if ci.AmortStartDate == nil {
		return false
	}

	// Check if entry date is on or after amortization start date
	if entryDate.Before(*ci.AmortStartDate) {
		return false
	}

	// Check if current book value is above salvage value
	if ci.CurrentBookValue <= ci.SalvageValue {
		return false
	}

	// Check amortization behavior if lifecycle status is available
	if ci.LifecycleStatus != nil {
		switch ci.LifecycleStatus.AmortizationBehavior {
		case "pending", "terminal":
			return false
		case "active":
			return true
		default:
			// Default to allowing depreciation if behavior is not explicitly set
			return true
		}
	}

	return true
}

func (c *calculatorSimple) calculateRemainingLife(ci *AmortizableCI, asOfDate time.Time) int {
	if ci.AmortStartDate == nil {
		return 0
	}

	// Calculate elapsed months
	elapsedMonths := int(asOfDate.Sub(*ci.AmortStartDate).Hours() / 24 / 30)
	remainingMonths := ci.UsefulLifeMonths - elapsedMonths

	if remainingMonths < 0 {
		remainingMonths = 0
	}

	// Additional check based on book value
	if ci.CurrentBookValue <= ci.SalvageValue {
		remainingMonths = 0
	}

	return remainingMonths
}

func (c *calculatorSimple) calculateElapsedMonths(startDate, endDate time.Time) int {
	// Calculate full months between two dates
	years := endDate.Year() - startDate.Year()
	months := endDate.Month() - startDate.Month()

	// Adjust for day of month
	if endDate.Day() < startDate.Day() {
		months--
	}

	totalMonths := years*12 + int(months)

	// Ensure we don't return negative months
	if totalMonths < 0 {
		return 0
	}

	return totalMonths
}

func (c *calculatorSimple) calculateConfidenceLevel(ci *AmortizableCI, projectionDate time.Time) string {
	if ci.AmortStartDate == nil {
		return "low"
	}

	// Calculate months to projection
	monthsToProjection := int(projectionDate.Sub(*ci.AmortStartDate).Hours() / 24 / 30)

	// Confidence decreases with longer projection periods
	switch {
	case monthsToProjection <= 12:
		return "high"
	case monthsToProjection <= 36:
		return "medium"
	default:
		return "low"
	}
}

// CalculateRestructuring calculates the prospective recalculation when useful life changes
func (c *calculatorSimple) CalculateRestructuring(ctx context.Context, ci *AmortizableCI, newUsefulLifeMonths int, asOfDate time.Time) (*RestructuringCalculation, error) {
	// Validation
	if newUsefulLifeMonths <= 0 {
		return nil, fmt.Errorf("new useful life months must be positive")
	}

	if ci.AmortStartDate == nil {
		return nil, fmt.Errorf("amortization start date is not set")
	}

	if newUsefulLifeMonths == ci.UsefulLifeMonths {
		return nil, fmt.Errorf("new useful life is the same as current useful life")
	}

	// Calculate current state
	currentMonthlyDepreciation := (ci.PurchaseCost - ci.SalvageValue) / float64(ci.UsefulLifeMonths)

	// Calculate elapsed months and remaining months under OLD useful life
	elapsedMonths := c.calculateElapsedMonths(*ci.AmortStartDate, asOfDate)
	remainingMonthsOld := ci.UsefulLifeMonths - elapsedMonths
	if remainingMonthsOld < 0 {
		remainingMonthsOld = 0
	}

	// Calculate remaining months under NEW useful life
	remainingMonthsNew := newUsefulLifeMonths - elapsedMonths
	if remainingMonthsNew < 0 {
		return &RestructuringCalculation{
			IsValid:           false,
			ValidationMessage: fmt.Sprintf("New useful life (%d months) is less than already elapsed time (%d months)", newUsefulLifeMonths, elapsedMonths),
		}, nil
	}

	// Calculate new monthly depreciation using PROSPECTIVE method
	// Formula: (current_book_value - salvage_value) / remaining_months
	var newMonthlyDepreciation float64
	if remainingMonthsNew > 0 {
		newMonthlyDepreciation = (ci.CurrentBookValue - ci.SalvageValue) / float64(remainingMonthsNew)
	} else {
		newMonthlyDepreciation = 0
	}

	// Calculate changes
	monthlyDepreciationChange := newMonthlyDepreciation - currentMonthlyDepreciation
	percentChange := 0.0
	if currentMonthlyDepreciation != 0 {
		percentChange = (monthlyDepreciationChange / currentMonthlyDepreciation) * 100
	}

	// Calculate remaining life extension
	remainingLifeExtension := remainingMonthsNew - remainingMonthsOld

	// Calculate new end date
	var newEndDate *time.Time
	if remainingMonthsNew > 0 && ci.AmortStartDate != nil {
		endDate := ci.AmortStartDate.AddDate(0, newUsefulLifeMonths, 0)
		newEndDate = &endDate
	}

	calculation := &RestructuringCalculation{
		CurrentUsefulLifeMonths:  ci.UsefulLifeMonths,
		CurrentMonthlyDepreciation: currentMonthlyDepreciation,
		CurrentBookValue:         ci.CurrentBookValue,
		AccumulatedDepreciation:  ci.AccumulatedDepreciation,
		RemainingMonthsOld:       remainingMonthsOld,
		NewUsefulLifeMonths:      newUsefulLifeMonths,
		RemainingMonthsNew:       remainingMonthsNew,
		NewMonthlyDepreciation:   newMonthlyDepreciation,
		MonthlyDepreciationChange: monthlyDepreciationChange,
		PercentChange:            percentChange,
		RemainingLifeExtension:   remainingLifeExtension,
		NewEndDate:                newEndDate,
		IsValid:                  true,
	}

	c.logger.Info().
		Str("ci_id", ci.ID.String()).
		Int("old_useful_life", ci.UsefulLifeMonths).
		Int("new_useful_life", newUsefulLifeMonths).
		Float64("old_monthly_depreciation", currentMonthlyDepreciation).
		Float64("new_monthly_depreciation", newMonthlyDepreciation).
		Float64("percent_change", percentChange).
		Int("remaining_months_new", remainingMonthsNew).
		Msg("Calculated restructuring")

	return calculation, nil
}