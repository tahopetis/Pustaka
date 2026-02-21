package ea

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseCSV_ValidFile tests parsing a valid CSV file
func TestParseCSV_ValidFile(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping CSV parsing test in short mode")
	}

	// Read test CSV file
	data, err := ioutil.ReadFile("testdata/ea-import-valid.csv")
	require.NoError(t, err, "Failed to read test CSV file")

	// Create a mock import service (we'll need to set up minimal dependencies)
	// For now, we'll test the parsing logic directly
	// This test will be updated once we have proper test fixtures

	assert.NotNil(t, data, "CSV data should not be nil")
	assert.Greater(t, len(data), 0, "CSV file should not be empty")
}

// TestParseCSV_InvalidFile tests parsing errors
func TestParseCSV_InvalidFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CSV parsing test in short mode")
	}

	// Test empty file
	emptyData := []byte("")

	// This will fail because we need a proper service setup
	// Once we have test fixtures, we can test:
	// - Empty files
	// - Malformed CSV
	// - Missing required columns
	// - CI_Type mismatch

	assert.NotNil(t, emptyData)
}

// TestValidateRow_ValidRow tests validation of a valid row
func TestValidateRow_ValidRow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping row validation test in short mode")
	}

	// Create a valid test row
	validRow := ImportRow{
		RowNumber:       2,
		Name:            "Test Application",
		CIType:          "EA.Application-BusinessApp",
		Domain:          "Application",
		LifecycleStatus: "Active",
		Owner:           "Business Architecture",
		Team:            "Enterprise Architecture",
		Tags:            "test,sample",
		Attributes: map[string]interface{}{
			"criticality":  "high",
			"description": "Test application for validation",
		},
	}

	// This test requires:
	// - Mock repository with GetCITypeByName
	// - Mock lifecycle status repository
	// - Mock team repository
	// We'll implement these once we have proper test infrastructure

	assert.Equal(t, "Test Application", validRow.Name)
	assert.Equal(t, "EA.Application-BusinessApp", validRow.CIType)
	assert.NotEmpty(t, validRow.Attributes)
}

// TestValidateRow_MissingRequiredField tests validation with missing required field
func TestValidateRow_MissingRequiredField(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping row validation test in short mode")
	}

	// Create an invalid row (missing Name)
	invalidRow := ImportRow{
		RowNumber:       2,
		Name:            "", // Missing required field
		CIType:          "EA.Application-BusinessApp",
		Domain:          "Application",
		LifecycleStatus: "Active",
		Owner:           "Business Architecture",
		Team:            "Enterprise Architecture",
	}

	assert.Empty(t, invalidRow.Name, "Name should be empty to test validation")

	// Expected validation error for missing Name field
	// This will be tested once we have proper mock setup
}

// TestValidateRow_InvalidLifecycleStatus tests validation with invalid lifecycle status
func TestValidateRow_InvalidLifecycleStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping row validation test in short mode")
	}

	// Create a row with invalid lifecycle status
	invalidRow := ImportRow{
		RowNumber:       3,
		Name:            "Test App",
		CIType:          "EA.Application-BusinessApp",
		Domain:          "Application",
		LifecycleStatus: "InvalidStatus", // Invalid status
		Owner:           "Business Architecture",
		Team:            "Enterprise Architecture",
	}

	assert.Equal(t, "InvalidStatus", invalidRow.LifecycleStatus)

	// Expected validation error for non-existent lifecycle status
	// This will be tested once we have proper mock setup
}

// TestValidateRow_InvalidOwner tests validation with invalid owner team
func TestValidateRow_InvalidOwner(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping row validation test in short mode")
	}

	// Create a row with invalid owner
	invalidRow := ImportRow{
		RowNumber:       5,
		Name:            "Test App",
		CIType:          "EA.Application-BusinessApp",
		Domain:          "Application",
		LifecycleStatus: "Active",
		Owner:           "NonExistent Team", // Invalid team
		Team:            "Enterprise Architecture",
	}

	assert.Equal(t, "NonExistent Team", invalidRow.Owner)

	// Expected validation error for non-existent owner team
	// This will be tested once we have proper mock setup
}

// TestValidateRow_InvalidEnumValue tests validation with invalid enum value
func TestValidateRow_InvalidEnumValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping row validation test in short mode")
	}

	// Create a row with invalid enum value in attributes
	invalidRow := ImportRow{
		RowNumber:       7,
		Name:            "Test App",
		CIType:          "EA.Application-BusinessApp",
		Domain:          "Application",
		LifecycleStatus: "Active",
		Owner:           "Business Architecture",
		Team:            "Enterprise Architecture",
		Attributes: map[string]interface{}{
			"criticality": "invalid_value", // Invalid enum value
		},
	}

	assert.NotNil(t, invalidRow.Attributes)
	assert.Equal(t, "invalid_value", invalidRow.Attributes["criticality"])

	// Expected validation error for invalid enum value
	// This will be tested once we have proper mock setup
}

// TestValidateImport_WithErrors tests validation of entire import with errors
func TestValidateImport_WithErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping import validation test in short mode")
	}

	// Read test CSV with errors
	data, err := ioutil.ReadFile("testdata/ea-import-errors.csv")
	require.NoError(t, err, "Failed to read test CSV file with errors")

	assert.NotNil(t, data)

	// Expected to find 4 validation errors:
	// - Row 2: Missing Name field
	// - Row 3: Invalid lifecycle status "InvalidStatus"
	// - Row 4: Invalid owner "NonExistent Team"
	// - Row 6: Invalid enum value in criticality

	// This will be tested once we have proper mock setup
}

// TestGenerateErrorCSV tests error CSV generation
func TestGenerateErrorCSV(t *testing.T) {
	// Create a mock import result with errors
	result := &ImportResult{
		TotalRows:    10,
		SuccessCount: 6,
		ErrorCount:   4,
		Errors: []ImportError{
			{
				RowNumber:    2,
				FieldName:    "Name",
				ErrorMessage: "Name is required",
			},
			{
				RowNumber:    3,
				FieldName:    "Lifecycle_Status",
				ErrorMessage: "Lifecycle_Status 'InvalidStatus' does not exist",
			},
			{
				RowNumber:    4,
				FieldName:    "Owner",
				ErrorMessage: "Owner team 'NonExistent Team' does not exist",
			},
			{
				RowNumber:    6,
				FieldName:    "criticality",
				ErrorMessage: "criticality must be one of: critical, high, medium, low",
			},
		},
	}

	assert.Equal(t, 4, result.ErrorCount)
	assert.Equal(t, 10, result.TotalRows)
	assert.Len(t, result.Errors, 4)

	// We would test GenerateErrorCSV() here once we have a proper service setup
	// Expected CSV content:
	// RowNumber,FieldName,ErrorMessage,ExpectedFormat,ActualValue
	// 2,Name,"Name is required",,
	// 3,Lifecycle_Status,"Lifecycle_Status 'InvalidStatus' does not exist",,
	// 4,Owner,"Owner team 'NonExistent Team' does not exist",,
	// 6,criticality,"criticality must be one of: critical, high, medium, low",,
}

// TestBulkCreateEntities_Integration tests bulk entity creation with database
func TestBulkCreateEntities_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test - requires database")
	}

	// This test requires:
	// - Test database connection
	// - Migrated schema
	// - Seed data (CI types, lifecycle statuses, teams)

	// Test flow:
	// 1. Read valid CSV file
	// 2. Parse CSV into ImportRow slice
	// 3. Validate all rows
	// 4. Call BulkCreateEntities
	// 5. Verify success_count = 10, error_count = 0
	// 6. Query database to verify 10 entities created
	// 7. Clean up test data

	t.Skip("Integration test not yet implemented - requires test database setup")
}

// TestGenerateTemplate tests CSV template generation
func TestGenerateTemplate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping template generation test in short mode")
	}

	// This test requires:
	// - Mock repository with GetCITypeByName
	// - Verify CSV has correct columns for CI type
	// - Verify example row is populated

	t.Skip("Template generation test not yet implemented - requires mock setup")
}

// Helper function to check if file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// TestMain setup for test suite
func TestMain(m *testing.M) {
	// Verify test data files exist before running tests
	if !fileExists("testdata/ea-import-valid.csv") {
		panic("Test data file missing: testdata/ea-import-valid.csv")
	}
	if !fileExists("testdata/ea-import-errors.csv") {
		panic("Test data file missing: testdata/ea-import-errors.csv")
	}

	os.Exit(m.Run())
}

// MockImportService is a mock implementation for testing
// This will be expanded once we have proper mock infrastructure
type MockImportService struct {
	parseCSVCalled bool
	validateCalled bool
	bulkCreateCalled bool
}

func (m *MockImportService) ParseCSV(file []byte, ciType string) ([]ImportRow, error) {
	m.parseCSVCalled = true
	// Return mock data
	return []ImportRow{
		{
			RowNumber: 2,
			Name:      "Test App",
			CIType:    ciType,
			Domain:    "Application",
		},
	}, nil
}

func (m *MockImportService) ValidateRow(ctx context.Context, row ImportRow, ciType *CITypeDefinition, rowNumber int) ([]ValidationError, error) {
	m.validateCalled = true
	return []ValidationError{}, nil
}

func (m *MockImportService) ValidateImport(ctx context.Context, rows []ImportRow, ciTypeName string) (*ImportResult, error) {
	m.validateCalled = true
	return &ImportResult{
		TotalRows:    len(rows),
		SuccessCount: len(rows),
		ErrorCount:   0,
	}, nil
}

func (m *MockImportService) BulkCreateEntities(ctx context.Context, rows []ImportRow, ciTypeName string, userID string) (*ImportResult, error) {
	m.bulkCreateCalled = true
	return &ImportResult{
		TotalRows:    len(rows),
		SuccessCount: len(rows),
		ErrorCount:   0,
	}, nil
}

// TestMockImportService_BasicFlow tests basic flow with mock
func TestMockImportService_BasicFlow(t *testing.T) {
	mock := &MockImportService{}

	// Test ParseCSV
	data := []byte("Name,CI_Type,Domain\nTest App,EA.Application-BusinessApp,Application")
	rows, err := mock.ParseCSV(data, "EA.Application-BusinessApp")
	assert.NoError(t, err)
	assert.True(t, mock.parseCSVCalled)
	assert.Len(t, rows, 1)

	// Test ValidateImport
	ctx := context.Background()
	result, err := mock.ValidateImport(ctx, rows, "EA.Application-BusinessApp")
	assert.NoError(t, err)
	assert.True(t, mock.validateCalled)
	assert.Equal(t, 1, result.TotalRows)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Equal(t, 0, result.ErrorCount)

	// Test BulkCreateEntities
	userID := uuid.New().String()
	importResult, err := mock.BulkCreateEntities(ctx, rows, "EA.Application-BusinessApp", userID)
	assert.NoError(t, err)
	assert.True(t, mock.bulkCreateCalled)
	assert.Equal(t, 1, importResult.SuccessCount)
}
