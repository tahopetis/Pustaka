package amortization

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// TestDataFactory provides utilities for generating test data
type TestDataFactory struct {
	rand *rand.Rand
}

// NewTestDataFactory creates a new test data factory
func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CreateTestCI creates a test amortizable CI with random data
func (f *TestDataFactory) CreateTestCI() *AmortizableCI {
	return &AmortizableCI{
		ID:                     uuid.New(),
		Name:                   fmt.Sprintf("Test-Server-%04d", f.rand.Intn(10000)),
		CIType:                 f.randomChoice([]string{"Server", "Database", "Laptop", "Network", "Storage"}),
		CITypeID:               uuid.New(),
		PurchaseCost:           f.randomFloat(1000.0, 50000.0),
		SalvageValue:           f.randomFloat(50.0, 2500.0),
		UsefulLifeMonths:       f.randomInt(12, 120),
		CurrentBookValue:       0.0, // Will be calculated
		AccumulatedDepreciation: 0.0, // Will be calculated
		DepreciationMethod:     f.randomChoice([]string{"straight_line", "declining_balance"}),
		AmortizationBehavior:   f.randomChoice([]string{"pending", "active", "terminal"}),
		IsAmortizable:          f.rand.Intn(10) > 1, // 90% chance of being amortizable
		CreatedAt:              time.Now().AddDate(0, -f.rand.Intn(24), -f.rand.Intn(30)),
		CreatedBy:              uuid.New(),
		Attributes: map[string]interface{}{
			"hostname":    fmt.Sprintf("host-%04d", f.rand.Intn(10000)),
			"ip_address":  fmt.Sprintf("10.0.%d.%d", f.rand.Intn(255), f.rand.Intn(255)),
			"cpu_cores":   f.randomInt(2, 32),
			"memory_gb":   f.randomChoice([]int{8, 16, 32, 64, 128}),
			"environment": f.randomChoice([]string{"production", "staging", "development"}),
		},
		Tags: f.randomTags(),
	}
}

// CreateTestLedgerEntry creates a test ledger entry
func (f *TestDataFactory) CreateTestLedgerEntry(ciID uuid.UUID) *AmortizationEntry {
	entryTypes := []string{"monthly_depreciation", "adjustment", "write_off", "correction"}
	entryType := entryTypes[f.rand.Intn(len(entryTypes))]

	var amount float64
	switch entryType {
	case "monthly_depreciation":
		amount = f.randomFloat(50.0, 500.0)
	case "adjustment":
		amount = f.randomFloat(-1000.0, 1000.0)
	case "write_off":
		amount = f.randomFloat(0.0, 5000.0)
	case "correction":
		amount = f.randomFloat(-500.0, 500.0)
	}

	description := ""
	if entryType != "monthly_depreciation" {
		description = f.randomChoice([]string{
			"Correction for calculation error",
			"Asset revaluation",
			"Disposal write-off",
			"Currency adjustment",
			"Policy change adjustment",
		})
	}

	entry := &AmortizationEntry{
		ID:                         uuid.New(),
		CIID:                       ciID,
		EntryType:                  entryType,
		EntryDate:                  time.Now().AddDate(0, -f.rand.Intn(12), -f.rand.Intn(30)),
		Amount:                     amount,
		BookValueBefore:            f.randomFloat(1000.0, 20000.0),
		BookValueAfter:             0.0, // Will be calculated
		AccumulatedDepreciationBefore: f.randomFloat(0.0, 5000.0),
		AccumulatedDepreciationAfter:  0.0, // Will be calculated
		CreatedAt:                  time.Now(),
		CreatedBy:                  f.uuidPtr(),
	}

	if description != "" {
		entry.Description = &description
	}

	// Calculate book value after
	entry.BookValueAfter = entry.BookValueBefore + amount
	entry.AccumulatedDepreciationAfter = entry.AccumulatedDepreciationBefore + amount

	return entry
}

// CreateTestAmortizationRun creates a test amortization run
func (f *TestDataFactory) CreateTestAmortizationRun() *AmortizationRun {
	statuses := []string{"started", "completed", "failed", "partial"}
	status := statuses[f.rand.Intn(len(statuses))]

	totalCIs := f.randomInt(10, 100)
	var processed, failed, skipped int
	var totalDepreciation float64

	switch status {
	case "completed":
		processed = totalCIs
		failed = 0
		skipped = 0
		totalDepreciation = f.randomFloat(1000.0, 20000.0)
	case "failed":
		processed = f.randomInt(0, totalCIs/2)
		failed = totalCIs - processed
		skipped = 0
		totalDepreciation = f.randomFloat(100.0, processed*200.0)
	case "partial":
		processed = f.randomInt(totalCIs/3, totalCIs*2/3)
		failed = f.randomInt(0, totalCIs/4)
		skipped = totalCIs - processed - failed
		totalDepreciation = f.randomFloat(500.0, processed*150.0)
	case "started":
		processed = 0
		failed = 0
		skipped = 0
		totalDepreciation = 0.0
	}

	run := &AmortizationRun{
		ID:                 uuid.New(),
		Status:             status,
		ProcessingDate:     time.Now().AddDate(0, -f.rand.Intn(6), -f.rand.Intn(30)),
		TotalAmortizableCIs: totalCIs,
		IsManual:           f.rand.Intn(10) > 7, // 30% chance of being manual
		DryRun:             f.rand.Intn(10) > 9, // 10% chance of being dry run
		CreatedAt:          time.Now(),
	}

	if status != "started" {
		startedAt := run.ProcessingDate.Add(-time.Duration(f.rand.Intn(60)) * time.Minute)
		run.StartedAt = &startedAt
	}

	if status == "completed" || status == "partial" || status == "failed" {
		completedAt := run.ProcessingDate.Add(time.Duration(f.rand.Intn(120)) * time.Minute)
		run.CompletedAt = &completedAt
	}

	if processed > 0 {
		run.ProcessedCIs = &processed
	}
	if failed > 0 {
		run.FailedCIs = &failed
	}
	if skipped > 0 {
		run.SkippedCIs = &skipped
	}
	if totalDepreciation > 0 {
		run.TotalDepreciation = &totalDepreciation
	}

	if run.IsManual {
		run.TriggeredBy = f.uuidPtr()
	}

	if status == "failed" {
		errorMsg := f.randomChoice([]string{
			"Database connection timeout",
			"Insufficient permissions",
			"Calculation overflow",
			"Invalid CI configuration",
			"Network connectivity issues",
		})
		run.ErrorSummary = &errorMsg
	}

	return run
}

// CreateTestAmortizationSummary creates a test amortization summary
func (f *TestDataFactory) CreateTestAmortizationSummary(groupBy string) *AmortizationSummary {
	groupKeys := []string{"Server", "Database", "Laptop", "Network", "Storage"}
	groupCount := f.rand.Intn(3, len(groupKeys))

	var groups []AmortizationGroup
	totalBookValue := 0.0
	totalDepreciation := 0.0
	totalCIs := 0

	for i := 0; i < groupCount; i++ {
		groupKey := groupKeys[i]
		ciCount := f.randomInt(1, 20)
		bookValue := f.randomFloat(10000.0, 100000.0)
		depreciation := f.randomFloat(1000.0, bookValue*0.5)
		avgAge := f.randomFloat(6.0, 60.0)

		groups = append(groups, AmortizationGroup{
			GroupName:          groupKey,
			CICount:            ciCount,
			TotalBookValue:     bookValue,
			TotalDepreciation:  depreciation,
			AverageAge:         avgAge,
		})

		totalBookValue += bookValue
		totalDepreciation += depreciation
		totalCIs += ciCount
	}

	return &AmortizationSummary{
		GroupBy:    groupBy,
		Groups:     groups,
		TotalCIs:   totalCIs,
		TotalBookValue:     totalBookValue,
		TotalDepreciation:  totalDepreciation,
		GeneratedAt: time.Now(),
	}
}

// Helper methods

func (f *TestDataFactory) randomChoice[T any](slice []T) T {
	return slice[f.rand.Intn(len(slice))]
}

func (f *TestDataFactory) randomFloat(min, max float64) float64 {
	return min + f.rand.Float64()*(max-min)
}

func (f *TestDataFactory) randomInt(min, max int) int {
	return min + f.rand.Intn(max-min)
}

func (f *TestDataFactory) randomTags() []string {
	allTags := []string{
		"production", "staging", "development",
		"critical", "important", "standard",
		"web", "database", "application", "infrastructure",
		"linux", "windows", "cloud", "on-premise",
	}

	count := f.rand.Intn(0, 4) // 0 to 3 tags
	if count == 0 {
		return []string{}
	}

	tags := make([]string, count)
	used := make(map[int]bool)

	for i := 0; i < count; i++ {
		for {
			idx := f.rand.Intn(len(allTags))
			if !used[idx] {
				tags[i] = allTags[idx]
				used[idx] = true
				break
			}
		}
	}

	return tags
}

func (f *TestDataFactory) uuidPtr() *uuid.UUID {
	id := uuid.New()
	return &id
}

// TestDataSet represents a complete set of test data
type TestDataSet struct {
	CIs                   map[uuid.UUID]*AmortizableCI
	LedgerEntries         map[uuid.UUID]*AmortizationEntry
	AmortizationRuns      map[uuid.UUID]*AmortizationRun
	Summaries             []*AmortizationSummary
	DepreciationSchedules map[uuid.UUID]*DepreciationSchedule
}

// CreateTestDataSet creates a comprehensive test data set
func (f *TestDataFactory) CreateTestDataSet(numCIs, numEntries, numRuns int) *TestDataSet {
	dataSet := &TestDataSet{
		CIs:                   make(map[uuid.UUID]*AmortizableCI),
		LedgerEntries:         make(map[uuid.UUID]*AmortizationEntry),
		AmortizationRuns:      make(map[uuid.UUID]*AmortizationRun),
		Summaries:             make([]*AmortizationSummary, 0),
		DepreciationSchedules: make(map[uuid.UUID]*DepreciationSchedule),
	}

	// Create CIs
	ciIDs := make([]uuid.UUID, numCIs)
	for i := 0; i < numCIs; i++ {
		ci := f.CreateTestCI()
		dataSet.CIs[ci.ID] = ci
		ciIDs[i] = ci.ID

		// Calculate current book value and accumulated depreciation
		if ci.IsAmortizable {
			monthsElapsed := int(time.Since(*ci.CreatedAt).Hours() / 24 / 30)
			if monthsElapsed > 0 {
				monthlyDepreciation := (ci.PurchaseCost - ci.SalvageValue) / float64(ci.UsefulLifeMonths)
				totalDepreciation := monthlyDepreciation * float64(monthsElapsed)
				if totalDepreciation > ci.PurchaseCost-ci.SalvageValue {
					totalDepreciation = ci.PurchaseCost - ci.SalvageValue
				}
				ci.AccumulatedDepreciation = totalDepreciation
				ci.CurrentBookValue = ci.PurchaseCost - totalDepreciation
			} else {
				ci.CurrentBookValue = ci.PurchaseCost
			}
		}
	}

	// Create ledger entries
	for i := 0; i < numEntries; i++ {
		ciID := ciIDs[f.rand.Intn(len(ciIDs))]
		entry := f.CreateTestLedgerEntry(ciID)
		dataSet.LedgerEntries[entry.ID] = entry
	}

	// Create amortization runs
	for i := 0; i < numRuns; i++ {
		run := f.CreateTestAmortizationRun()
		dataSet.AmortizationRuns[run.ID] = run
	}

	// Create summaries
	groupByOptions := []string{"ci_type", "lifecycle_status", "age_bucket", "depreciation_method"}
	for _, groupBy := range groupByOptions {
		summary := f.CreateTestAmortizationSummary(groupBy)
		dataSet.Summaries = append(dataSet.Summaries, summary)
	}

	return dataSet
}

// Validation utilities

// ValidateAmortizationCI validates that an AmortizableCI follows business rules
func ValidateAmortizationCI(ci *AmortizableCI) []string {
	var errors []string

	if ci.ID == uuid.Nil {
		errors = append(errors, "ID is required")
	}

	if ci.Name == "" {
		errors = append(errors, "Name is required")
	}

	if ci.CIType == "" {
		errors = append(errors, "CI Type is required")
	}

	if ci.IsAmortizable {
		if ci.PurchaseCost <= 0 {
			errors = append(errors, "Purchase cost must be positive for amortizable assets")
		}

		if ci.SalvageValue < 0 {
			errors = append(errors, "Salvage value cannot be negative")
		}

		if ci.SalvageValue > ci.PurchaseCost {
			errors = append(errors, "Salvage value cannot exceed purchase cost")
		}

		if ci.UsefulLifeMonths <= 0 {
			errors = append(errors, "Useful life months must be positive")
		}

		if ci.DepreciationMethod != "" {
			validMethods := []string{"straight_line", "declining_balance"}
			valid := false
			for _, method := range validMethods {
				if ci.DepreciationMethod == method {
					valid = true
					break
				}
			}
			if !valid {
				errors = append(errors, "Invalid depreciation method")
			}
		}

		if ci.AmortizationBehavior != "" {
			validBehaviors := []string{"pending", "active", "terminal"}
			valid := false
			for _, behavior := range validBehaviors {
				if ci.AmortizationBehavior == behavior {
					valid = true
					break
				}
			}
			if !valid {
				errors = append(errors, "Invalid amortization behavior")
			}
		}

		if ci.CurrentBookValue < ci.SalvageValue {
			errors = append(errors, "Current book value cannot be less than salvage value")
		}

		if ci.AccumulatedDepreciation < 0 {
			errors = append(errors, "Accumulated depreciation cannot be negative")
		}

		maxDepreciation := ci.PurchaseCost - ci.SalvageValue
		if ci.AccumulatedDepreciation > maxDepreciation {
			errors = append(errors, "Accumulated depreciation exceeds maximum allowed")
		}
	}

	return errors
}

// ValidateAmortizationEntry validates that an AmortizationEntry follows business rules
func ValidateAmortizationEntry(entry *AmortizationEntry) []string {
	var errors []string

	if entry.ID == uuid.Nil {
		errors = append(errors, "ID is required")
	}

	if entry.CIID == uuid.Nil {
		errors = append(errors, "CI ID is required")
	}

	validEntryTypes := []string{"monthly_depreciation", "write_off", "adjustment", "correction"}
	valid := false
	for _, entryType := range validEntryTypes {
		if entry.EntryType == entryType {
			valid = true
			break
		}
	}
	if !valid {
		errors = append(errors, "Invalid entry type")
	}

	if entry.EntryDate.IsZero() {
		errors = append(errors, "Entry date is required")
	}

	// Validate amount based on entry type
	switch entry.EntryType {
	case "monthly_depreciation":
		if entry.Amount <= 0 {
			errors = append(errors, "Monthly depreciation amount must be positive")
		}
	case "write_off":
		if entry.Amount <= 0 {
			errors = append(errors, "Write-off amount must be positive")
		}
	}

	// Validate book value consistency
	if entry.BookValueBefore < 0 {
		errors = append(errors, "Book value before cannot be negative")
	}

	if entry.BookValueAfter < 0 {
		errors = append(errors, "Book value after cannot be negative")
	}

	if entry.AccumulatedDepreciationBefore < 0 {
		errors = append(errors, "Accumulated depreciation before cannot be negative")
	}

	if entry.AccumulatedDepreciationAfter < 0 {
		errors = append(errors, "Accumulated depreciation after cannot be negative")
	}

	// Validate calculation consistency
	expectedBookValueAfter := entry.BookValueBefore + entry.Amount
	if entry.BookValueAfter != expectedBookValueAfter {
		errors = append(errors, "Book value after calculation is inconsistent")
	}

	expectedAccumulatedDepreciationAfter := entry.AccumulatedDepreciationBefore + entry.Amount
	if entry.AccumulatedDepreciationAfter != expectedAccumulatedDepreciationAfter {
		errors = append(errors, "Accumulated depreciation after calculation is inconsistent")
	}

	return errors
}

// ValidateAmortizationRun validates that an AmortizationRun follows business rules
func ValidateAmortizationRun(run *AmortizationRun) []string {
	var errors []string

	if run.ID == uuid.Nil {
		errors = append(errors, "ID is required")
	}

	validStatuses := []string{"started", "completed", "failed", "partial"}
	valid := false
	for _, status := range validStatuses {
		if run.Status == status {
			valid = true
			break
		}
	}
	if !valid {
		errors = append(errors, "Invalid status")
	}

	if run.ProcessingDate.IsZero() {
		errors = append(errors, "Processing date is required")
	}

	if run.TotalAmortizableCIs < 0 {
		errors = append(errors, "Total amortizable CIs cannot be negative")
	}

	// Validate processed/failed/skipped counts
	processed := 0
	if run.ProcessedCIs != nil {
		processed = *run.ProcessedCIs
	}

	failed := 0
	if run.FailedCIs != nil {
		failed = *run.FailedCIs
	}

	skipped := 0
	if run.SkippedCIs != nil {
		skipped = *run.SkippedCIs
	}

	if processed < 0 || failed < 0 || skipped < 0 {
		errors = append(errors, "Processed, failed, and skipped counts cannot be negative")
	}

	total := processed + failed + skipped
	if total > run.TotalAmortizableCIs {
		errors = append(errors, "Sum of processed, failed, and skipped CIs exceeds total")
	}

	// Validate total depreciation
	if run.TotalDepreciation != nil && *run.TotalDepreciation < 0 {
		errors = append(errors, "Total depreciation cannot be negative")
	}

	// Validate status consistency
	if run.Status == "completed" {
		if run.StartedAt == nil || run.CompletedAt == nil {
			errors = append(errors, "Completed run must have started and completed timestamps")
		}
		if run.StartedAt.After(*run.CompletedAt) {
			errors = append(errors, "Started time cannot be after completed time")
		}
	}

	return errors
}