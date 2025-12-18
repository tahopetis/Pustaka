package amortization

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"
)

// MockAmortizationService implements the AmortizationService interface for testing
type MockAmortizationService struct {
	cis                   map[uuid.UUID]*AmortizableCI
	ledgerEntries         map[uuid.UUID]*AmortizationEntry
	amortizationRuns      map[uuid.UUID]*AmortizationRun
	summaries             []*AmortizationSummary
	depreciationSchedules map[uuid.UUID]*DepreciationSchedule
	errors                map[string]error
}

// NewMockAmortizationService creates a new mock amortization service
func NewMockAmortizationService() *MockAmortizationService {
	return &MockAmortizationService{
		cis:                   make(map[uuid.UUID]*AmortizableCI),
		ledgerEntries:         make(map[uuid.UUID]*AmortizationEntry),
		amortizationRuns:      make(map[uuid.UUID]*AmortizationRun),
		summaries:             make([]*AmortizationSummary, 0),
		depreciationSchedules: make(map[uuid.UUID]*DepreciationSchedule),
		errors:                make(map[string]error),
	}
}

// Reset clears all mock data
func (m *MockAmortizationService) Reset() {
	m.cis = make(map[uuid.UUID]*AmortizableCI)
	m.ledgerEntries = make(map[uuid.UUID]*AmortizationEntry)
	m.amortizationRuns = make(map[uuid.UUID]*AmortizationRun)
	m.summaries = make([]*AmortizationSummary, 0)
	m.depreciationSchedules = make(map[uuid.UUID]*DepreciationSchedule)
	m.errors = make(map[string]error)
}

// Helper methods to add test data
func (m *MockAmortizationService) AddAmortizableCI(ci *AmortizableCI) {
	m.cis[ci.ID] = ci
}

func (m *MockAmortizationService) AddLedgerEntry(entry *AmortizationEntry) {
	m.ledgerEntries[entry.ID] = entry
}

func (m *MockAmortizationService) AddAmortizationRun(run *AmortizationRun) {
	m.amortizationRuns[run.ID] = run
}

func (m *MockAmortizationService) AddAmortizationSummary(summary *AmortizationSummary) {
	m.summaries = append(m.summaries, summary)
}

func (m *MockAmortizationService) AddDepreciationSchedule(schedule *DepreciationSchedule) {
	m.depreciationSchedules[schedule.ReportID] = schedule
}

// SetError sets an error to be returned for a specific operation
func (m *MockAmortizationService) SetError(operation string, err error) {
	m.errors[operation] = err
}

// Service interface implementation

func (m *MockAmortizationService) ListAmortizableCIs(ctx context.Context, filters *AmortizableCIFilters) (*AmortizationCIList, error) {
	if err, ok := m.errors["ListAmortizableCIs"]; ok {
		return nil, err
	}

	var results []AmortizableCI
	for _, ci := range m.cis {
		if m.matchesCIFilters(ci, filters) {
			results = append(results, *ci)
		}
	}

	page := 1
	pageSize := 20
	if filters.Page != nil {
		page = *filters.Page
	}
	if filters.PageSize != nil {
		pageSize = *filters.PageSize
	}

	total := len(results)
	totalPages := (total + pageSize - 1) / pageSize

	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	if start >= end {
		results = []AmortizableCI{}
	} else {
		results = results[start:end]
	}

	return &AmortizationCIList{
		CIs:         results,
		TotalCount:  total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	}, nil
}

func (m *MockAmortizationService) GetAmortizationDetails(ctx context.Context, ciID uuid.UUID) (*AmortizationDetails, error) {
	if err, ok := m.errors["GetAmortizationDetails"]; ok {
		return nil, err
	}

	ci, exists := m.cis[ciID]
	if !exists {
		return nil, fmt.Errorf("CI not found")
	}

	// Create amortization details with calculated fields
	details := &AmortizationDetails{
		AmortizableCI: *ci,
	}

	// Calculate monthly depreciation
	if ci.UsefulLifeMonths > 0 && ci.PurchaseCost > ci.SalvageValue {
		monthlyDepreciation := (ci.PurchaseCost - ci.SalvageValue) / float64(ci.UsefulLifeMonths)
		details.MonthlyDepreciation = &monthlyDepreciation

		// Calculate remaining life
		remainingMonths := ci.UsefulLifeMonths
		if ci.AmortStartDate != nil {
			monthsElapsed := int(time.Since(*ci.AmortStartDate).Hours() / 24 / 30)
			remainingMonths = ci.UsefulLifeMonths - monthsElapsed
			if remainingMonths < 0 {
				remainingMonths = 0
			}
		}
		details.RemainingLifeMonths = &remainingMonths
	}

	// Get recent ledger entries
	var recentEntries []LedgerEntry
	for _, entry := range m.ledgerEntries {
		if entry.CIID == ciID {
			recentEntries = append(recentEntries, *entry)
		}
	}

	// Sort by date descending and limit to 10
	if len(recentEntries) > 10 {
		recentEntries = recentEntries[:10]
	}
	details.RecentLedgerEntries = recentEntries

	return details, nil
}

func (m *MockAmortizationService) UpdateAmortizationConfig(ctx context.Context, ciID uuid.UUID, req *UpdateAmortizationConfig, userID uuid.UUID) (*AmortizationDetails, error) {
	if err, ok := m.errors["UpdateAmortizationConfig"]; ok {
		return nil, err
	}

	ci, exists := m.cis[ciID]
	if !exists {
		return nil, fmt.Errorf("CI not found")
	}

	// Apply updates
	if req.PurchaseCost != nil {
		ci.PurchaseCost = *req.PurchaseCost
	}
	if req.SalvageValue != nil {
		ci.SalvageValue = *req.SalvageValue
	}
	if req.AmortStartDate != nil {
		ci.AmortStartDate = req.AmortStartDate
	}
	if req.UsefulLifeMonths != nil {
		ci.UsefulLifeMonths = *req.UsefulLifeMonths
	}
	if req.DepreciationMethod != nil {
		ci.DepreciationMethod = *req.DepreciationMethod
	}

	ci.UpdatedBy = &userID
	now := time.Now()
	ci.UpdatedAt = &now

	// Reset book value if purchase cost changed
	if req.PurchaseCost != nil {
		ci.CurrentBookValue = *req.PurchaseCost
		ci.AccumulatedDepreciation = 0.0
	}

	return m.GetAmortizationDetails(ctx, ciID)
}

func (m *MockAmortizationService) GetLedgerEntries(ctx context.Context, filters *LedgerFilters) (*LedgerEntryList, error) {
	if err, ok := m.errors["GetLedgerEntries"]; ok {
		return nil, err
	}

	var results []AmortizationEntry
	for _, entry := range m.ledgerEntries {
		if m.matchesLedgerFilters(entry, filters) {
			results = append(results, *entry)
		}
	}

	page := 1
	pageSize := 50
	if filters.Page != nil {
		page = *filters.Page
	}
	if filters.PageSize != nil {
		pageSize = *filters.PageSize
	}

	total := len(results)
	totalPages := (total + pageSize - 1) / pageSize

	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	if start >= end {
		results = []AmortizationEntry{}
	} else {
		results = results[start:end]
	}

	return &LedgerEntryList{
		Entries:    results,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (m *MockAmortizationService) GetLedgerEntry(ctx context.Context, entryID uuid.UUID) (*LedgerEntryResponse, error) {
	if err, ok := m.errors["GetLedgerEntry"]; ok {
		return nil, err
	}

	entry, exists := m.ledgerEntries[entryID]
	if !exists {
		return nil, fmt.Errorf("Ledger entry not found")
	}

	response := &LedgerEntryResponse{
		LedgerEntry: *entry,
	}

	// Add CI details
	if ci, exists := m.cis[entry.CIID]; exists {
		response.CIDetails = &BaseCI{
			ID:   ci.ID,
			Name: ci.Name,
			CIType: ci.CIType,
			Attributes: ci.Attributes,
			Tags: ci.Tags,
			CreatedAt: ci.CreatedAt,
			UpdatedAt: ci.UpdatedAt,
		}
	}

	return response, nil
}

func (m *MockAmortizationService) CreateAdjustment(ctx context.Context, req *CreateAdjustmentRequest, userID uuid.UUID) (*LedgerEntryResponse, error) {
	if err, ok := m.errors["CreateAdjustment"]; ok {
		return nil, err
	}

	// Validate CI exists
	ci, exists := m.cis[req.CIID]
	if !exists {
		return nil, fmt.Errorf("CI not found")
	}

	// Create adjustment entry
	entry := &AmortizationEntry{
		ID:                         uuid.New(),
		CIID:                       req.CIID,
		EntryType:                  "adjustment",
		EntryDate:                  time.Now(),
		Amount:                     req.Amount,
		BookValueBefore:            ci.CurrentBookValue,
		AccumulatedDepreciationBefore: ci.AccumulatedDepreciation,
		Description:                &req.Description,
		CreatedAt:                  time.Now(),
		CreatedBy:                  &userID,
	}

	// Calculate new book value
	entry.BookValueAfter = entry.BookValueBefore + req.Amount
	entry.AccumulatedDepreciationAfter = entry.AccumulatedDepreciationBefore + req.Amount

	// Update CI
	ci.CurrentBookValue = entry.BookValueAfter
	ci.AccumulatedDepreciation = entry.AccumulatedDepreciationAfter

	m.ledgerEntries[entry.ID] = entry

	return m.GetLedgerEntry(ctx, entry.ID)
}

func (m *MockAmortizationService) ListAmortizationRuns(ctx context.Context, filters *AmortizationRunFilters) (*AmortizationRunList, error) {
	if err, ok := m.errors["ListAmortizationRuns"]; ok {
		return nil, err
	}

	var results []AmortizationRun
	for _, run := range m.amortizationRuns {
		if m.matchesRunFilters(run, filters) {
			results = append(results, *run)
		}
	}

	page := 1
	pageSize := 20
	if filters.Page != nil {
		page = *filters.Page
	}
	if filters.PageSize != nil {
		pageSize = *filters.PageSize
	}

	total := len(results)
	totalPages := (total + pageSize - 1) / pageSize

	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	if start >= end {
		results = []AmortizationRun{}
	} else {
		results = results[start:end]
	}

	return &AmortizationRunList{
		Runs:       results,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (m *MockAmortizationService) GetAmortizationRun(ctx context.Context, runID uuid.UUID) (*AmortizationRunResponse, error) {
	if err, ok := m.errors["GetAmortizationRun"]; ok {
		return nil, err
	}

	run, exists := m.amortizationRuns[runID]
	if !exists {
		return nil, fmt.Errorf("Amortization run not found")
	}

	response := &AmortizationRunResponse{
		AmortizationRun: *run,
		ProcessedItems: []struct {
			CIID           uuid.UUID `json:"ci_id"`
			CIName         string    `json:"ci_name"`
			Status         string    `json:"status"`
			ErrorMessage   string    `json:"error_message,omitempty"`
			DepreciationAmount float64 `json:"depreciation_amount"`
		}{},
	}

	// Add some mock processed items
	for _, ci := range m.cis {
		response.ProcessedItems = append(response.ProcessedItems, struct {
			CIID           uuid.UUID `json:"ci_id"`
			CIName         string    `json:"ci_name"`
			Status         string    `json:"status"`
			ErrorMessage   string    `json:"error_message,omitempty"`
			DepreciationAmount float64 `json:"depreciation_amount"`
		}{
			CIID: ci.ID,
			CIName: ci.Name,
			Status: "processed",
			DepreciationAmount: 100.0,
		})
	}

	return response, nil
}

func (m *MockAmortizationService) TriggerManualRun(ctx context.Context, req *ManualRunRequest, userID uuid.UUID) (*AmortizationRun, error) {
	if err, ok := m.errors["TriggerManualRun"]; ok {
		return nil, err
	}

	run := &AmortizationRun{
		ID:                 uuid.New(),
		Status:             "started",
		ProcessingDate:     time.Now(),
		StartedAt:          timePtr(time.Now()),
		TotalAmortizableCIs: len(m.cis),
		IsManual:           true,
		DryRun:             req.DryRun,
		TriggeredBy:        &userID,
		CreatedAt:          time.Now(),
	}

	if !req.DryRun {
		processed := len(m.cis)
		run.ProcessedCIs = &processed
		run.Status = "completed"
		run.CompletedAt = timePtr(time.Now())

		totalDepreciation := float64(len(m.cis)) * 100.0
		run.TotalDepreciation = &totalDepreciation
	}

	m.amortizationRuns[run.ID] = run
	return run, nil
}

func (m *MockAmortizationService) GetAmortizationSummaries(ctx context.Context, req *SummaryRequest) (*AmortizationSummary, error) {
	if err, ok := m.errors["GetAmortizationSummaries"]; ok {
		return nil, err
	}

	if len(m.summaries) > 0 {
		// Return the first matching summary
		for _, summary := range m.summaries {
			if summary.GroupBy == req.GroupBy {
				return summary, nil
			}
		}
	}

	// Create a default summary if none exist
	summary := &AmortizationSummary{
		GroupBy:    req.GroupBy,
		TotalCIs:   len(m.cis),
		GeneratedAt: time.Now(),
		Groups: []AmortizationGroup{
			{
				GroupName:          "Server",
				CICount:            len(m.cis),
				TotalBookValue:     100000.0,
				TotalDepreciation:  20000.0,
				AverageAge:         24.5,
			},
		},
	}

	return summary, nil
}

func (m *MockAmortizationService) GenerateDepreciationSchedule(ctx context.Context, req *DepreciationScheduleRequest) (*DepreciationSchedule, error) {
	if err, ok := m.errors["GenerateDepreciationSchedule"]; ok {
		return nil, err
	}

	reportID := uuid.New()
	schedule := &DepreciationSchedule{
		ReportID: reportID,
		DateRange: DepreciationScheduleRange{
			StartDate: req.DateFrom,
			EndDate:   req.DateTo,
		},
		Schedule: []DepreciationScheduleEntry{},
	}

	// Generate schedule entries for each CI
	for _, ci := range m.cis {
		if len(req.CIIDs) > 0 {
			found := false
			for _, id := range req.CIIDs {
				if ci.ID == id {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		entry := DepreciationScheduleEntry{
			CIID:                    ci.ID,
			CIName:                  ci.Name,
			CIType:                  ci.CIType,
			PeriodStart:             req.DateFrom,
			PeriodEnd:               req.DateTo,
			OpeningBookValue:        ci.CurrentBookValue,
			DepreciationAmount:      166.67, // Mock monthly depreciation
			ClosingBookValue:        ci.CurrentBookValue - 166.67,
			AccumulatedDepreciation: ci.AccumulatedDepreciation + 166.67,
		}
		schedule.Schedule = append(schedule.Schedule, entry)
	}

	m.depreciationSchedules[reportID] = schedule
	return schedule, nil
}

// Helper methods for filtering

func (m *MockAmortizationService) matchesCIFilters(ci *AmortizableCI, filters *AmortizableCIFilters) bool {
	if filters == nil {
		return true
	}

	// Search filter
	if filters.Search != nil && *filters.Search != "" {
		if !contains(ci.Name, *filters.Search) {
			return false
		}
	}

	// CI type filter
	if len(filters.CITypeIDs) > 0 {
		found := false
		for _, typeID := range filters.CITypeIDs {
			if ci.CITypeID == typeID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Lifecycle status filter
	if len(filters.LifecycleStatusIDs) > 0 {
		found := false
		for _, statusID := range filters.LifecycleStatusIDs {
			if ci.LifecycleStatusID != nil && *ci.LifecycleStatusID == statusID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Is amortizable filter
	if filters.IsAmortizable != nil && ci.IsAmortizable != *filters.IsAmortizable {
		return false
	}

	// Book value range filter
	if filters.MinBookValue != nil && ci.CurrentBookValue < *filters.MinBookValue {
		return false
	}
	if filters.MaxBookValue != nil && ci.CurrentBookValue > *filters.MaxBookValue {
		return false
	}

	return true
}

func (m *MockAmortizationService) matchesLedgerFilters(entry *AmortizationEntry, filters *LedgerFilters) bool {
	if filters == nil {
		return true
	}

	// CI filter
	if filters.CIID != nil && entry.CIID != *filters.CIID {
		return false
	}

	// Entry type filter
	if len(filters.EntryTypes) > 0 {
		found := false
		for _, entryType := range filters.EntryTypes {
			if entry.EntryType == entryType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Date range filter
	if filters.DateFrom != nil && entry.EntryDate.Before(*filters.DateFrom) {
		return false
	}
	if filters.DateTo != nil && entry.EntryDate.After(*filters.DateTo) {
		return false
	}

	// Amount range filter
	if filters.MinAmount != nil && entry.Amount < *filters.MinAmount {
		return false
	}
	if filters.MaxAmount != nil && entry.Amount > *filters.MaxAmount {
		return false
	}

	return true
}

func (m *MockAmortizationService) matchesRunFilters(run *AmortizationRun, filters *AmortizationRunFilters) bool {
	if filters == nil {
		return true
	}

	// Status filter
	if len(filters.Status) > 0 {
		found := false
		for _, status := range filters.Status {
			if run.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Date range filter
	if filters.DateFrom != nil && run.ProcessingDate.Before(*filters.DateFrom) {
		return false
	}
	if filters.DateTo != nil && run.ProcessingDate.After(*filters.DateTo) {
		return false
	}

	// Is manual filter
	if filters.IsManual != nil && run.IsManual != *filters.IsManual {
		return false
	}

	// Triggered by filter
	if filters.TriggeredBy != nil && (run.TriggeredBy == nil || *run.TriggeredBy != *filters.TriggeredBy) {
		return false
	}

	return true
}

// Utility functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) &&
		(s[:len(substr)] == substr ||
		 s[len(s)-len(substr):] == substr ||
		 indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}