package amortization

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/pustaka/pustaka/internal/api/middleware"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// AmortizationIntegrationTestSuite provides end-to-end testing of the amortization API
type AmortizationIntegrationTestSuite struct {
	suite.Suite
	handler        *AmortizationHandler
	mockService    *MockAmortizationService
	router         chi.Router
	logger         *pustakaLogger.Logger
	adminUserID    uuid.UUID
	userUserID     uuid.UUID
	testCIs        map[uuid.UUID]*AmortizableCI
	testEntries    map[uuid.UUID]*AmortizationEntry
	testRuns       map[uuid.UUID]*AmortizationRun
}

func (suite *AmortizationIntegrationTestSuite) SetupSuite() {
	suite.logger = pustakaLogger.NewLogger()
	suite.mockService = NewMockAmortizationService()
	suite.handler = NewAmortizationHandler(suite.mockService, suite.logger)

	// Setup test users
	suite.adminUserID = uuid.New()
	suite.userUserID = uuid.New()

	// Initialize router
	suite.router = chi.NewRouter()
	suite.handler.RegisterRoutes(suite.router)

	// Setup test data
	suite.setupTestData()
}

func (suite *AmortizationIntegrationTestSuite) TearDownSuite() {
	// Cleanup if needed
}

func (suite *AmortizationIntegrationTestSuite) SetupTest() {
	suite.mockService.Reset()
	suite.setupTestData()
}

func (suite *AmortizationIntegrationTestSuite) setupTestData() {
	suite.testCIs = make(map[uuid.UUID]*AmortizableCI)
	suite.testEntries = make(map[uuid.UUID]*AmortizationEntry)
	suite.testRuns = make(map[uuid.UUID]*AmortizationRun)

	// Create test CIs
	testCI1 := &AmortizableCI{
		ID:                     uuid.New(),
		Name:                   "Production Web Server 01",
		CIType:                 "Server",
		CITypeID:               uuid.New(),
		PurchaseCost:           15000.0,
		SalvageValue:           750.0,
		UsefulLifeMonths:       60,
		CurrentBookValue:       12000.0,
		AccumulatedDepreciation: 3000.0,
		DepreciationMethod:     "straight_line",
		AmortizationBehavior:   "active",
		IsAmortizable:          true,
		CreatedAt:              time.Now().Add(-6 * 30 * 24 * time.Hour), // 6 months ago
		CreatedBy:              suite.adminUserID,
		Attributes: map[string]interface{}{
			"hostname":    "web-prod-01",
			"ip_address":  "10.0.1.10",
			"cpu_cores":   8,
			"memory_gb":   32,
			"environment": "production",
		},
		Tags: []string{"production", "web", "critical"},
	}

	testCI2 := &AmortizableCI{
		ID:                     uuid.New(),
		Name:                   "Database Server 01",
		CIType:                 "Database",
		CITypeID:               uuid.New(),
		PurchaseCost:           25000.0,
		SalvageValue:           1250.0,
		UsefulLifeMonths:       84,
		CurrentBookValue:       22000.0,
		AccumulatedDepreciation: 3000.0,
		DepreciationMethod:     "straight_line",
		AmortizationBehavior:   "active",
		IsAmortizable:          true,
		CreatedAt:              time.Now().Add(-3 * 30 * 24 * time.Hour), // 3 months ago
		CreatedBy:              suite.adminUserID,
		Attributes: map[string]interface{}{
			"hostname":    "db-prod-01",
			"ip_address":  "10.0.2.10",
			"cpu_cores":   16,
			"memory_gb":   64,
			"storage_gb":  1000,
			"engine":      "postgresql",
		},
		Tags: []string{"production", "database", "critical"},
	}

	testCI3 := &AmortizableCI{
		ID:                     uuid.New(),
		Name:                   "Development Laptop",
		CIType:                 "Laptop",
		CITypeID:               uuid.New(),
		PurchaseCost:           2000.0,
		SalvageValue:           100.0,
		UsefulLifeMonths:       36,
		CurrentBookValue:       1500.0,
		AccumulatedDepreciation: 500.0,
		DepreciationMethod:     "straight_line",
		AmortizationBehavior:   "active",
		IsAmortizable:          true,
		CreatedAt:              time.Now().Add(-12 * 30 * 24 * time.Hour), // 12 months ago
		CreatedBy:              suite.userUserID,
		Attributes: map[string]interface{}{
			"hostname":    "dev-laptop-01",
			"model":       "MacBook Pro",
			"cpu_cores":   8,
			"memory_gb":   16,
			"storage_gb":  512,
		},
		Tags: []string{"development", "laptop"},
	}

	suite.testCIs[testCI1.ID] = testCI1
	suite.testCIs[testCI2.ID] = testCI2
	suite.testCIs[testCI3.ID] = testCI3

	// Create test ledger entries
	entry1 := &AmortizationEntry{
		ID:                         uuid.New(),
		CIID:                       testCI1.ID,
		EntryType:                  "monthly_depreciation",
		EntryDate:                  time.Now().AddDate(0, -1, 0), // 1 month ago
		Amount:                     237.5,
		BookValueBefore:            12237.5,
		BookValueAfter:             12000.0,
		AccumulatedDepreciationBefore: 2762.5,
		AccumulatedDepreciationAfter:  3000.0,
		CreatedAt:                  time.Now().AddDate(0, -1, 0),
		CreatedBy:                  &suite.adminUserID,
	}

	entry2 := &AmortizationEntry{
		ID:                         uuid.New(),
		CIID:                       testCI2.ID,
		EntryType:                  "monthly_depreciation",
		EntryDate:                  time.Now().AddDate(0, -1, 0), // 1 month ago
		Amount:                     281.25,
		BookValueBefore:            22281.25,
		BookValueAfter:             22000.0,
		AccumulatedDepreciationBefore: 2718.75,
		AccumulatedDepreciationAfter:  3000.0,
		CreatedAt:                  time.Now().AddDate(0, -1, 0),
		CreatedBy:                  &suite.adminUserID,
	}

	entry3 := &AmortizationEntry{
		ID:                         uuid.New(),
		CIID:                       testCI1.ID,
		EntryType:                  "adjustment",
		EntryDate:                  time.Now().AddDate(0, -2, 0), // 2 months ago
		Amount:                     100.0,
		BookValueBefore:            12137.5,
		BookValueAfter:             12237.5,
		AccumulatedDepreciationBefore: 2662.5,
		AccumulatedDepreciationAfter:  2762.5,
		Description:                stringPtr("Correction for calculation error"),
		CreatedAt:                  time.Now().AddDate(0, -2, 0),
		CreatedBy:                  &suite.adminUserID,
	}

	suite.testEntries[entry1.ID] = entry1
	suite.testEntries[entry2.ID] = entry2
	suite.testEntries[entry3.ID] = entry3

	// Create test amortization runs
	run1 := &AmortizationRun{
		ID:                 uuid.New(),
		Status:             "completed",
		ProcessingDate:     time.Now().AddDate(0, -1, 0), // 1 month ago
		StartedAt:          timePtr(time.Now().AddDate(0, -1, 0, -1, 0, 0)),
		CompletedAt:        timePtr(time.Now().AddDate(0, -1, 0)),
		TotalAmortizableCIs: 50,
		ProcessedCIs:       intPtr(48),
		FailedCIs:          intPtr(2),
		SkippedCIs:         intPtr(0),
		TotalDepreciation:  floatPtr(12500.0),
		IsManual:           false,
		DryRun:             false,
		CreatedAt:          time.Now().AddDate(0, -1, 0, -1, 0, 0),
	}

	run2 := &AmortizationRun{
		ID:                 uuid.New(),
		Status:             "completed",
		ProcessingDate:     time.Now(),
		StartedAt:          timePtr(time.Now().Add(-1 * time.Hour)),
		CompletedAt:        timePtr(time.Now()),
		TotalAmortizableCIs: 52,
		ProcessedCIs:       intPtr(52),
		FailedCIs:          intPtr(0),
		SkippedCIs:         intPtr(0),
		TotalDepreciation:  floatPtr(13200.0),
		IsManual:           true,
		DryRun:             false,
		TriggeredBy:        &suite.adminUserID,
		CreatedAt:          time.Now().Add(-1 * time.Hour),
	}

	suite.testRuns[run1.ID] = run1
	suite.testRuns[run2.ID] = run2

	// Add all data to mock service
	for _, ci := range suite.testCIs {
		suite.mockService.AddAmortizableCI(ci)
	}
	for _, entry := range suite.testEntries {
		suite.mockService.AddLedgerEntry(entry)
	}
	for _, run := range suite.testRuns {
		suite.mockService.AddAmortizationRun(run)
	}
}

// Helper methods for HTTP testing

func (suite *AmortizationIntegrationTestSuite) makeRequest(method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		reqBody = new(bytes.Buffer)
		json.NewEncoder(reqBody).Encode(body)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(suite.T(), err)

	// Set default headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add authentication
	if _, hasAuth := headers["Authorization"]; !hasAuth {
		req.Header.Set("Authorization", "Bearer valid.test.token")
	}

	// Add user context
	user := &middleware.AuthenticatedUser{
		UserID:      suite.adminUserID,
		Username:    "admin",
		Email:       "admin@example.com",
		Role:        "admin",
		Permissions: []string{"amortization:read", "amortization:write", "amortization:adjust", "amortization:admin"},
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	suite.router.ServeHTTP(rr, req)

	return rr
}

func (suite *AmortizationIntegrationTestSuite) assertJSONResponse(rr *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	assert.Equal(suite.T(), expectedStatus, rr.Code)
	assert.Equal(suite.T(), "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	return response
}

func (suite *AmortizationIntegrationTestSuite) assertErrorResponse(rr *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) {
	assert.Equal(suite.T(), expectedStatus, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	errorObj, ok := response["error"].(map[string]interface{})
	require.True(suite.T(), ok, "Response should contain error object")

	if expectedMessage != "" {
		if message, ok := errorObj["message"].(string); ok {
			assert.Contains(suite.T(), message, expectedMessage)
		}
	}
}

// Integration Tests

func (suite *AmortizationIntegrationTestSuite) TestListAmortizableCIs_Workflow() {
	// Test listing all CIs
	rr := suite.makeRequest("GET", "/amortization/configuration-items", nil, nil)
	response := suite.assertJSONResponse(rr, http.StatusOK)

	// Validate response structure
	assert.Contains(suite.T(), response, "cis")
	assert.Contains(suite.T(), response, "total")
	assert.Contains(suite.T(), response, "page")
	assert.Contains(suite.T(), response, "total_pages")

	cis := response["cis"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(cis), 3, "Should have at least 3 test CIs")

	// Test filtering by CI type
	rr = suite.makeRequest("GET", "/amortization/configuration-items?ci_type=Server", nil, nil)
	response = suite.assertJSONResponse(rr, http.StatusOK)

	cis = response["cis"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(cis), 1, "Should have at least 1 server")

	// Test search functionality
	rr = suite.makeRequest("GET", "/amortization/configuration-items?search=Database", nil, nil)
	response = suite.assertJSONResponse(rr, http.StatusOK)

	cis = response["cis"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(cis), 1, "Should find database server in search")
}

func (suite *AmortizationIntegrationTestSuite) TestAmortizationDetails_Workflow() {
	// Get a test CI
	var testCI *AmortizableCI
	for _, ci := range suite.testCIs {
		testCI = ci
		break
	}
	require.NotNil(suite.T(), testCI)

	// Test getting amortization details
	path := fmt.Sprintf("/amortization/configuration-items/%s", testCI.ID.String())
	rr := suite.makeRequest("GET", path, nil, nil)
	response := suite.assertJSONResponse(rr, http.StatusOK)

	// Validate response contains all expected fields
	expectedFields := []string{
		"id", "name", "ci_type", "purchase_cost", "salvage_value",
		"amort_start_date", "useful_life_months", "current_book_value",
		"accumulated_depreciation", "depreciation_method", "monthly_depreciation",
		"remaining_life_months", "amortization_behavior", "recent_ledger_entries",
	}

	for _, field := range expectedFields {
		assert.Contains(suite.T(), response, field, fmt.Sprintf("Response should contain field: %s", field))
	}

	assert.Equal(suite.T(), testCI.ID.String(), response["id"])
	assert.Equal(suite.T(), testCI.Name, response["name"])
	assert.Equal(suite.T(), testCI.CIType, response["ci_type"])
	assert.Equal(suite.T(), testCI.PurchaseCost, response["purchase_cost"])
	assert.Equal(suite.T(), testCI.CurrentBookValue, response["current_book_value"])

	// Validate calculated fields
	assert.NotNil(suite.T(), response["monthly_depreciation"], "Should calculate monthly depreciation")
	assert.NotNil(suite.T(), response["remaining_life_months"], "Should calculate remaining life")

	// Validate recent ledger entries
	recentEntries := response["recent_ledger_entries"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(recentEntries), 1, "Should have recent ledger entries")
}

func (suite *AmortizationIntegrationTestSuite) TestUpdateAmortizationConfig_Workflow() {
	// Get a test CI
	testCI := suite.testCIs[uuid.New()] // This will be nil, need to get a real one
	for _, ci := range suite.testCIs {
		testCI = ci
		break
	}
	require.NotNil(suite.T(), testCI)

	path := fmt.Sprintf("/amortization/configuration-items/%s", testCI.ID.String())

	// Test updating financial information
	updateReq := map[string]interface{}{
		"purchase_cost":       18000.0,
		"salvage_value":       900.0,
		"useful_life_months":  72,
		"depreciation_method": "straight_line",
	}

	rr := suite.makeRequest("PUT", path, updateReq, nil)
	response := suite.assertJSONResponse(rr, http.StatusOK)

	// Validate the update
	assert.Equal(suite.T(), 18000.0, response["purchase_cost"])
	assert.Equal(suite.T(), 900.0, response["salvage_value"])
	assert.Equal(suite.T(), 72, response["useful_life_months"])
	assert.Equal(suite.T(), 18000.0, response["current_book_value"]) // Reset to purchase cost
	assert.Equal(suite.T(), 0.0, response["accumulated_depreciation"]) // Reset

	// Test partial update
	partialReq := map[string]interface{}{
		"useful_life_months": 84,
	}

	rr = suite.makeRequest("PUT", path, partialReq, nil)
	response = suite.assertJSONResponse(rr, http.StatusOK)

	assert.Equal(suite.T(), 18000.0, response["purchase_cost"]) // Unchanged
	assert.Equal(suite.T(), 84, response["useful_life_months"]) // Updated
}

func (suite *AmortizationIntegrationTestSuite) TestLedgerManagement_Workflow() {
	// Test listing all ledger entries
	rr := suite.makeRequest("GET", "/amortization/ledger", nil, nil)
	response := suite.assertJSONResponse(rr, http.StatusOK)

	entries := response["entries"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(entries), 3, "Should have at least 3 test entries")

	// Test filtering by CI
	testCI := suite.testCIs[uuid.New()]
	for _, ci := range suite.testCIs {
		testCI = ci
		break
	}
	require.NotNil(suite.T(), testCI)

	path := fmt.Sprintf("/amortization/ledger?ci_id=%s", testCI.ID.String())
	rr = suite.makeRequest("GET", path, nil, nil)
	response = suite.assertJSONResponse(rr, http.StatusOK)

	entries = response["entries"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(entries), 1, "Should have entries for test CI")

	// Test filtering by entry type
	rr = suite.makeRequest("GET", "/amortization/ledger?entry_type=adjustment", nil, nil)
	response = suite.assertJSONResponse(rr, http.StatusOK)

	entries = response["entries"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(entries), 1, "Should have adjustment entries")

	// Test getting specific ledger entry
	var testEntry *AmortizationEntry
	for _, entry := range suite.testEntries {
		testEntry = entry
		break
	}
	require.NotNil(suite.T(), testCI)

	path = fmt.Sprintf("/amortization/ledger/%s", testEntry.ID.String())
	rr = suite.makeRequest("GET", path, nil, nil)
	response = suite.assertJSONResponse(rr, http.StatusOK)

	// Validate entry details
	assert.Equal(suite.T(), testEntry.ID.String(), response["id"])
	assert.Equal(suite.T(), testEntry.CIID.String(), response["ci_id"])
	assert.Equal(suite.T(), testEntry.EntryType, response["entry_type"])
	assert.NotNil(suite.T(), response["ci_details"], "Should include CI details")

	// Test creating adjustment
	adjustmentReq := map[string]interface{}{
		"ci_id":       testCI.ID.String(),
		"amount":      250.0,
		"description": "Test adjustment for integration test",
	}

	rr = suite.makeRequest("POST", "/amortization/adjustments", adjustmentReq, nil)
	response = suite.assertJSONResponse(rr, http.StatusCreated)

	// Validate adjustment was created
	assert.Equal(suite.T(), "adjustment", response["entry_type"])
	assert.Equal(suite.T(), 250.0, response["amount"])
	assert.NotNil(suite.T(), response["id"], "Should have new entry ID")
}

func (suite *AmortizationIntegrationTestSuite) TestAmortizationRuns_Workflow() {
	// Test listing amortization runs
	rr := suite.makeRequest("GET", "/amortization/runs", nil, nil)
	response := suite.assertJSONResponse(rr, http.StatusOK)

	runs := response["runs"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(runs), 2, "Should have at least 2 test runs")

	// Test filtering by status
	rr = suite.makeRequest("GET", "/amortization/runs?status=completed", nil, nil)
	response = suite.assertJSONResponse(rr, http.StatusOK)

	runs = response["runs"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(runs), 2, "Should have completed runs")

	// Test getting specific run details
	var testRun *AmortizationRun
	for _, run := range suite.testRuns {
		testRun = run
		break
	}
	require.NotNil(suite.T(), testRun)

	path := fmt.Sprintf("/amortization/runs/%s", testRun.ID.String())
	rr = suite.makeRequest("GET", path, nil, nil)
	response = suite.assertJSONResponse(rr, http.StatusOK)

	// Validate run details
	assert.Equal(suite.T(), testRun.ID.String(), response["id"])
	assert.Equal(suite.T(), testRun.Status, response["status"])
	assert.NotNil(suite.T(), response["processed_items"], "Should include processed items")

	processedItems := response["processed_items"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(processedItems), 1, "Should have processed items")

	// Test triggering manual run
	manualRunReq := map[string]interface{}{
		"dry_run": true,
	}

	rr = suite.makeRequest("POST", "/amortization/runs", manualRunReq, nil)
	response = suite.assertJSONResponse(rr, http.StatusAccepted)

	// Validate run was triggered
	assert.Equal(suite.T(), "started", response["status"])
	assert.Equal(suite.T(), "Amortization run initiated", response["message"])
	assert.NotNil(suite.T(), response["run_id"], "Should have run ID")

	// Test run with specific CIs
	ciIDs := []string{}
	for _, ci := range suite.testCIs {
		ciIDs = append(ciIDs, ci.ID.String())
		if len(ciIDs) >= 2 {
			break
		}
	}

	manualRunReq = map[string]interface{}{
		"dry_run": false,
		"ci_ids":  ciIDs,
	}

	rr = suite.makeRequest("POST", "/amortization/runs", manualRunReq, nil)
	response = suite.assertJSONResponse(rr, http.StatusAccepted)

	// Validate run with specific CIs
	assert.Equal(suite.T(), "started", response["status"])
	assert.NotNil(suite.T(), response["run_id"])
}

func (suite *AmortizationIntegrationTestSuite) TestReportsAndSummaries_Workflow() {
	// Test amortization summaries
	rr := suite.makeRequest("GET", "/amortization/summaries", nil, nil)
	response := suite.assertJSONResponse(rr, http.StatusOK)

	// Validate summary structure
	expectedFields := []string{"summaries", "total_book_value", "total_accumulated_depreciation", "date_as_of"}
	for _, field := range expectedFields {
		assert.Contains(suite.T(), response, field, fmt.Sprintf("Response should contain field: %s", field))
	}

	summaries := response["summaries"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(summaries), 1, "Should have summary data")

	// Test different groupings
	groupByOptions := []string{"ci_type", "lifecycle_status", "depreciation_method"}
	for _, groupBy := range groupByOptions {
		rr = suite.makeRequest("GET", fmt.Sprintf("/amortization/summaries?group_by=%s", groupBy), nil, nil)
		response = suite.assertJSONResponse(rr, http.StatusOK)
		assert.Contains(suite.T(), response, "summaries")
	}

	// Test depreciation schedule report
	dateFrom := time.Now().AddDate(0, -3, 0).Format("2006-01-02")
	dateTo := time.Now().AddDate(0, 3, 0).Format("2006-01-02")

	path := fmt.Sprintf("/amortization/reports/depreciation-schedule?date_from=%s&date_to=%s", dateFrom, dateTo)
	rr = suite.makeRequest("GET", path, nil, nil)
	response = suite.assertJSONResponse(rr, http.StatusOK)

	// Validate schedule structure
	assert.Contains(suite.T(), response, "report_id")
	assert.Contains(suite.T(), response, "date_range")
	assert.Contains(suite.T(), response, "schedule")

	dateRange := response["date_range"].(map[string]interface{})
	assert.Contains(suite.T(), dateRange, "start_date")
	assert.Contains(suite.T(), dateRange, "end_date")

	schedule := response["schedule"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(schedule), 1, "Should have schedule entries")

	// Test CSV format
	path += "&format=csv"
	rr = suite.makeRequest("GET", path, nil, nil)
	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	assert.Equal(suite.T(), "text/csv", rr.Header().Get("Content-Type"))
	assert.Contains(suite.T(), rr.Header().Get("Content-Disposition"), "attachment")
}

func (suite *AmortizationIntegrationTestSuite) TestErrorHandling_Workflow() {
	// Test not found errors
	nonExistentID := uuid.New().String()

	path := fmt.Sprintf("/amortization/configuration-items/%s", nonExistentID)
	rr := suite.makeRequest("GET", path, nil, nil)
	suite.assertErrorResponse(rr, http.StatusNotFound, "not found")

	path = fmt.Sprintf("/amortization/ledger/%s", nonExistentID)
	rr = suite.makeRequest("GET", path, nil, nil)
	suite.assertErrorResponse(rr, http.StatusNotFound, "not found")

	path = fmt.Sprintf("/amortization/runs/%s", nonExistentID)
	rr = suite.makeRequest("GET", path, nil, nil)
	suite.assertErrorResponse(rr, http.StatusNotFound, "not found")

	// Test validation errors
	path = fmt.Sprintf("/amortization/configuration-items/%s", uuid.New().String())
	invalidReq := map[string]interface{}{
		"purchase_cost": -1000.0, // Negative cost
	}

	rr = suite.makeRequest("PUT", path, invalidReq, nil)
	suite.assertErrorResponse(rr, http.StatusBadRequest, "")

	// Test invalid UUID formats
	rr = suite.makeRequest("GET", "/amortization/configuration-items/invalid-uuid", nil, nil)
	suite.assertErrorResponse(rr, http.StatusBadRequest, "Invalid CI ID")

	rr = suite.makeRequest("POST", "/amortization/adjustments", map[string]interface{}{
		"ci_id": "invalid-uuid",
		"amount": 100.0,
		"description": "Test",
	}, nil)
	suite.assertErrorResponse(rr, http.StatusBadRequest, "")

	// Test missing required fields
	rr = suite.makeRequest("POST", "/amortization/adjustments", map[string]interface{}{
		"ci_id": uuid.New().String(),
		// Missing amount and description
	}, nil)
	suite.assertErrorResponse(rr, http.StatusBadRequest, "")

	// Test invalid JSON
	rr = suite.makeRequest("PUT", "/amortization/configuration-items/"+uuid.New().String(),
		"{invalid json}", nil)
	suite.assertErrorResponse(rr, http.StatusBadRequest, "Invalid request body")
}

func (suite *AmortizationIntegrationTestSuite) TestAuthentication_Workflow() {
	// Test without authentication
	headers := map[string]string{"Authorization": ""}
	rr := suite.makeRequest("GET", "/amortization/configuration-items", nil, headers)
	suite.assertErrorResponse(rr, http.StatusUnauthorized, "")

	// Test with invalid token
	headers = map[string]string{"Authorization": "Bearer invalid.token"}
	rr = suite.makeRequest("GET", "/amortization/configuration-items", nil, headers)
	suite.assertErrorResponse(rr, http.StatusUnauthorized, "")

	// Test with different user roles
	userHeaders := map[string]string{
		"Authorization": "Bearer user.token",
	}

	// This would require modifying the makeRequest method to support different user contexts
	// For now, we assume the middleware properly handles role-based access
}

func (suite *AmortizationIntegrationTestSuite) TestPerformance_Workflow() {
	// Test response times for various endpoints
	start := time.Now()
	rr := suite.makeRequest("GET", "/amortization/configuration-items", nil, nil)
	duration := time.Since(start)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	assert.Less(suite.T(), duration, 100*time.Millisecond, "List CIs should respond within 100ms")

	start = time.Now()
	testCI := suite.testCIs[uuid.New()]
	for _, ci := range suite.testCIs {
		testCI = ci
		break
	}
	require.NotNil(suite.T(), testCI)

	path := fmt.Sprintf("/amortization/configuration-items/%s", testCI.ID.String())
	rr = suite.makeRequest("GET", path, nil, nil)
	duration = time.Since(start)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	assert.Less(suite.T(), duration, 50*time.Millisecond, "Get CI details should respond within 50ms")
}

func (suite *AmortizationIntegrationTestSuite) TestConsistency_Workflow() {
	// Test data consistency across operations
	testCI := suite.testCIs[uuid.New()]
	for _, ci := range suite.testCIs {
		testCI = ci
		break
	}
	require.NotNil(suite.T(), testCI)

	// Get initial state
	path := fmt.Sprintf("/amortization/configuration-items/%s", testCI.ID.String())
	rr := suite.makeRequest("GET", path, nil, nil)
	initialResponse := suite.assertJSONResponse(rr, http.StatusOK)
	initialBookValue := initialResponse["current_book_value"].(float64)

	// Create adjustment
	adjustmentReq := map[string]interface{}{
		"ci_id":       testCI.ID.String(),
		"amount":      -500.0,
		"description": "Consistency test adjustment",
	}

	rr = suite.makeRequest("POST", "/amortization/adjustments", adjustmentReq, nil)
	suite.assertJSONResponse(rr, http.StatusCreated)

	// Verify book value changed
	rr = suite.makeRequest("GET", path, nil, nil)
	updatedResponse := suite.assertJSONResponse(rr, http.StatusOK)
	updatedBookValue := updatedResponse["current_book_value"].(float64)

	assert.Equal(suite.T(), initialBookValue-500.0, updatedBookValue,
		"Book value should reflect adjustment")
}

// Test Suite Runner

func TestAmortizationIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AmortizationIntegrationTestSuite))
}

// Helper functions

func stringPtr(s string) *string {
	return &s
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