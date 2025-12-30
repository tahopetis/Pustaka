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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/pustaka/pustaka/internal/api/middleware"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// AmortizationContractTestSuite provides comprehensive contract testing
// for the amortization module API endpoints against OpenAPI specification
type AmortizationContractTestSuite struct {
	suite.Suite
	handler        *AmortizationHandler
	mockService    *MockAmortizationService
	router         *httptest.Server
	logger         *pustakaLogger.Logger
	adminUserID    uuid.UUID
	userUserID     uuid.UUID
	validAuthToken string
}

func (suite *AmortizationContractTestSuite) SetupSuite() {
	suite.logger = pustakaLogger.NewLogger()
	suite.mockService = NewMockAmortizationService()
	suite.handler = NewAmortizationHandler(suite.mockService, suite.logger)

	// Setup test users
	suite.adminUserID = uuid.New()
	suite.userUserID = uuid.New()
	suite.validAuthToken = "valid.jwt.token"

	// Setup test router
	router := http.NewServeMux()

	// Register amortization routes
	chiRouter := chi.NewRouter()
	suite.handler.RegisterRoutes(chiRouter)

	// Wrap with authentication middleware
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Mock authenticated user
			user := &middleware.AuthenticatedUser{
				UserID:      suite.adminUserID,
				Username:    "admin",
				Email:       "admin@example.com",
				Role:        "admin",
				Permissions: []string{"amortization:read", "amortization:write", "amortization:adjust", "amortization:admin"},
			}
			ctx := context.WithValue(r.Context(), middleware.UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	finalRouter := authMiddleware(chiRouter)
	suite.router = httptest.NewServer(finalRouter)
}

func (suite *AmortizationContractTestSuite) TearDownSuite() {
	suite.router.Close()
}

func (suite *AmortizationContractTestSuite) SetupTest() {
	suite.mockService.Reset()
}

// Test Configuration Items Endpoints

func (suite *AmortizationContractTestSuite) TestListAmortizableCIs_Contract() {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "Default parameters",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"cis", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "With pagination",
			queryParams:    "page=2&limit=10",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"cis", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "With filters",
			queryParams:    "ci_type=Server&lifecycle_status=active&amortization_behavior=active",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"cis", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "With search",
			queryParams:    "search=web",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"cis", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "With sorting",
			queryParams:    "sort=current_book_value&order=desc",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"cis", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "Invalid page parameter",
			queryParams:    "page=0",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid limit parameter",
			queryParams:    "limit=0",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid sort parameter",
			queryParams:    "sort=invalid_field",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid order parameter",
			queryParams:    "order=invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid amortization behavior",
			queryParams:    "amortization_behavior=invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			url := "/amortization/configuration-items"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(suite.T(), err)
			req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)

			rr := httptest.NewRecorder()
			suite.handler.ListAmortizableCIs(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				// Validate required fields
				for _, field := range tt.expectedFields {
					_, exists := response[field]
					assert.True(suite.T(), exists, "Expected field %s not found in response", field)
				}

				// Validate pagination structure
				assert.IsType(suite.T(), 1, response["page"])
				assert.IsType(suite.T(), 1, response["limit"])
				assert.IsType(suite.T(), int64(0), response["total"])
				assert.IsType(suite.T(), 1, response["total_pages"])

				// Validate CI structure if present
				if cis, ok := response["cis"].([]interface{}); ok && len(cis) > 0 {
					ci := cis[0].(map[string]interface{})
					suite.validateAmortizableCIFields(ci)
				}
			}
		})
	}
}

func (suite *AmortizationContractTestSuite) TestGetAmortizationDetails_Contract() {
	tests := []struct {
		name           string
		ciID           string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "Valid CI ID",
			ciID:           uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectedFields: []string{
				"id", "name", "ci_type", "purchase_cost", "salvage_value",
				"amort_start_date", "useful_life_months", "current_book_value",
				"accumulated_depreciation", "depreciation_method", "monthly_depreciation",
				"remaining_life_months", "amortization_behavior", "recent_ledger_entries",
			},
		},
		{
			name:           "Invalid UUID format",
			ciID:           "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent CI",
			ciID:           uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.name == "Valid CI ID" {
				// Setup mock data
				ciID := uuid.MustParse(tt.ciID)
				suite.mockService.AddAmortizableCI(&AmortizableCI{
					ID:                      ciID,
					Name:                    "Test Server",
					CIType:                  "Server",
					PurchaseCost:            10000.0,
					SalvageValue:            500.0,
					UsefulLifeMonths:        60,
					CurrentBookValue:        8000.0,
					AccumulatedDepreciation: 2000.0,
					DepreciationMethod:      "straight_line",
					AmortizationBehavior:    "active",
					IsAmortizable:           true,
				})
			}

			url := fmt.Sprintf("/amortization/configuration-items/%s", tt.ciID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(suite.T(), err)
			req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)

			rr := httptest.NewRecorder()
			suite.handler.GetAmortizationDetails(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				// Validate required fields
				for _, field := range tt.expectedFields {
					_, exists := response[field]
					assert.True(suite.T(), exists, "Expected field %s not found in response", field)
				}

				// Validate field types
				suite.validateAmortizationDetails(response)
			}
		})
	}
}

func (suite *AmortizationContractTestSuite) TestUpdateAmortizationConfig_Contract() {
	tests := []struct {
		name           string
		ciID           string
		requestBody    map[string]interface{}
		expectedStatus int
		validateFields bool
	}{
		{
			name: "Valid full update",
			ciID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"purchase_cost":       12000.0,
				"salvage_value":       600.0,
				"amort_start_date":    "2024-01-01",
				"useful_life_months":  72,
				"depreciation_method": "straight_line",
			},
			expectedStatus: http.StatusOK,
			validateFields: true,
		},
		{
			name: "Partial update",
			ciID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"purchase_cost":      15000.0,
				"useful_life_months": 84,
			},
			expectedStatus: http.StatusOK,
			validateFields: true,
		},
		{
			name: "Invalid purchase cost (negative)",
			ciID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"purchase_cost": -1000.0,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid useful life (zero)",
			ciID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"useful_life_months": 0,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid depreciation method",
			ciID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"depreciation_method": "invalid_method",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid date format",
			ciID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"amort_start_date": "2024-13-45",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty request body",
			ciID:           uuid.New().String(),
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			ciID:           uuid.New().String(),
			requestBody:    map[string]interface{}{"invalid": json.RawMessage("{invalid")},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid CI ID",
			ciID:           "invalid-uuid",
			requestBody:    map[string]interface{}{"purchase_cost": 10000.0},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Non-existent CI",
			ciID: uuid.New().String(),
			requestBody: map[string]interface{}{
				"purchase_cost": 10000.0,
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.name == "Valid full update" || tt.name == "Partial update" || tt.name == "Non-existent CI" {
				ciID := uuid.MustParse(tt.ciID)
				if tt.name != "Non-existent CI" {
					suite.mockService.AddAmortizableCI(&AmortizableCI{
						ID:                      ciID,
						Name:                    "Test Server",
						CIType:                  "Server",
						PurchaseCost:            10000.0,
						SalvageValue:            500.0,
						UsefulLifeMonths:        60,
						CurrentBookValue:        8000.0,
						AccumulatedDepreciation: 2000.0,
						IsAmortizable:           true,
					})
				}
			}

			url := fmt.Sprintf("/amortization/configuration-items/%s", tt.ciID)

			var body bytes.Buffer
			err := json.NewEncoder(&body).Encode(tt.requestBody)
			require.NoError(suite.T(), err)

			req, err := http.NewRequest(http.MethodPut, url, &body)
			require.NoError(suite.T(), err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)

			rr := httptest.NewRecorder()
			suite.handler.UpdateAmortizationConfig(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.validateFields && tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				// Validate response structure
				suite.validateAmortizationDetails(response)
			}
		})
	}
}

// Test Ledger Endpoints

func (suite *AmortizationContractTestSuite) TestGetLedgerEntries_Contract() {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "Default parameters",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"entries", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "With CI filter",
			queryParams:    fmt.Sprintf("ci_id=%s", uuid.New().String()),
			expectedStatus: http.StatusOK,
			expectedFields: []string{"entries", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "With entry type filter",
			queryParams:    "entry_type=monthly_depreciation",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"entries", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "With date range",
			queryParams:    "date_from=2024-01-01&date_to=2024-12-31",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"entries", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "With sorting",
			queryParams:    "sort=entry_date&order=desc",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"entries", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "Invalid UUID format",
			queryParams:    "ci_id=invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid date format",
			queryParams:    "date_from=2024-13-45",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid entry type",
			queryParams:    "entry_type=invalid_type",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "With CI type filter",
			queryParams:    fmt.Sprintf("ci_type_id=%s", uuid.New().String()),
			expectedStatus: http.StatusOK,
			expectedFields: []string{"entries", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "With CI name search",
			queryParams:    "ci_name_search=test",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"entries", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "Invalid CI type ID format",
			queryParams:    "ci_type_id=invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Combined filters",
			queryParams:    fmt.Sprintf("ci_type_id=%s&ci_name_search=test&date_from=2024-01-01", uuid.New().String()),
			expectedStatus: http.StatusOK,
			expectedFields: []string{"entries", "page", "limit", "total", "total_pages"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock data for successful cases
			if tt.expectedStatus == http.StatusOK {
				ciID := uuid.New()
				suite.mockService.AddLedgerEntry(&AmortizationEntry{
					ID:                            uuid.New(),
					CIID:                          ciID,
					EntryType:                     "monthly_depreciation",
					EntryDate:                     time.Now(),
					Amount:                        100.0,
					BookValueBefore:               5000.0,
					BookValueAfter:                4900.0,
					AccumulatedDepreciationBefore: 1000.0,
					AccumulatedDepreciationAfter:  1100.0,
				})
			}

			url := "/amortization/ledger"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(suite.T(), err)
			req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)

			rr := httptest.NewRecorder()
			suite.handler.GetLedgerEntries(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				// Validate required fields
				for _, field := range tt.expectedFields {
					_, exists := response[field]
					assert.True(suite.T(), exists, "Expected field %s not found in response", field)
				}

				// Validate entry structure if present
				if entries, ok := response["entries"].([]interface{}); ok && len(entries) > 0 {
					entry := entries[0].(map[string]interface{})
					suite.validateLedgerEntryFields(entry)
				}
			}
		})
	}
}

func (suite *AmortizationContractTestSuite) TestGetLedgerEntry_Contract() {
	tests := []struct {
		name           string
		entryID        string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "Valid entry ID",
			entryID:        uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectedFields: []string{
				"id", "ci_id", "entry_type", "entry_date", "amount",
				"book_value_before", "book_value_after", "accumulated_depreciation_before",
				"accumulated_depreciation_after", "ci_details",
			},
		},
		{
			name:           "Invalid UUID format",
			entryID:        "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent entry",
			entryID:        uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.name == "Valid entry ID" {
				entryID := uuid.MustParse(tt.entryID)
				ciID := uuid.New()
				suite.mockService.AddLedgerEntry(&AmortizationEntry{
					ID:                            entryID,
					CIID:                          ciID,
					EntryType:                     "monthly_depreciation",
					EntryDate:                     time.Now(),
					Amount:                        100.0,
					BookValueBefore:               5000.0,
					BookValueAfter:                4900.0,
					AccumulatedDepreciationBefore: 1000.0,
					AccumulatedDepreciationAfter:  1100.0,
				})
				suite.mockService.AddAmortizableCI(&AmortizableCI{
					ID:            ciID,
					Name:          "Test Server",
					CIType:        "Server",
					IsAmortizable: true,
				})
			}

			url := fmt.Sprintf("/amortization/ledger/%s", tt.entryID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(suite.T(), err)
			req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)

			rr := httptest.NewRecorder()
			suite.handler.GetLedgerEntry(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				// Validate required fields
				for _, field := range tt.expectedFields {
					_, exists := response[field]
					assert.True(suite.T(), exists, "Expected field %s not found in response", field)
				}

				// Validate field types and structure
				suite.validateLedgerEntryResponse(response)
			}
		})
	}
}

func (suite *AmortizationContractTestSuite) TestCreateAdjustment_Contract() {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		validateFields bool
	}{
		{
			name: "Valid adjustment",
			requestBody: map[string]interface{}{
				"ci_id":          uuid.New().String(),
				"amount":         500.0,
				"description":    "Correction for calculation error",
				"effective_date": "2024-01-15",
			},
			expectedStatus: http.StatusCreated,
			validateFields: true,
		},
		{
			name: "Valid adjustment without effective date",
			requestBody: map[string]interface{}{
				"ci_id":       uuid.New().String(),
				"amount":      250.0,
				"description": "Manual adjustment",
			},
			expectedStatus: http.StatusCreated,
			validateFields: true,
		},
		{
			name: "Missing required ci_id",
			requestBody: map[string]interface{}{
				"amount":      500.0,
				"description": "Missing CI ID",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing required amount",
			requestBody: map[string]interface{}{
				"ci_id":       uuid.New().String(),
				"description": "Missing amount",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing required description",
			requestBody: map[string]interface{}{
				"ci_id":  uuid.New().String(),
				"amount": 500.0,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid CI ID format",
			requestBody: map[string]interface{}{
				"ci_id":       "invalid-uuid",
				"amount":      500.0,
				"description": "Invalid CI ID",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Description too long",
			requestBody: map[string]interface{}{
				"ci_id":       uuid.New().String(),
				"amount":      500.0,
				"description": string(make([]byte, 501)), // 501 characters
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid date format",
			requestBody: map[string]interface{}{
				"ci_id":          uuid.New().String(),
				"amount":         500.0,
				"description":    "Invalid date",
				"effective_date": "2024-13-45",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty request body",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			requestBody:    map[string]interface{}{"invalid": json.RawMessage("{invalid")},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock data for valid CI cases
			if tt.validateFields || tt.name == "Invalid CI ID format" {
				if ciIDStr, ok := tt.requestBody["ci_id"].(string); ok {
					if ciID, err := uuid.Parse(ciIDStr); err == nil {
						suite.mockService.AddAmortizableCI(&AmortizableCI{
							ID:            ciID,
							Name:          "Test Server",
							CIType:        "Server",
							IsAmortizable: true,
						})
					}
				}
			}

			var body bytes.Buffer
			err := json.NewEncoder(&body).Encode(tt.requestBody)
			require.NoError(suite.T(), err)

			req, err := http.NewRequest(http.MethodPost, "/amortization/adjustments", &body)
			require.NoError(suite.T(), err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)

			rr := httptest.NewRecorder()
			suite.handler.CreateAdjustment(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.validateFields && tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				// Validate response structure
				suite.validateLedgerEntryResponse(response)

				// Verify entry type
				assert.Equal(suite.T(), "adjustment", response["entry_type"])
			}
		})
	}
}

// Test Amortization Runs Endpoints

func (suite *AmortizationContractTestSuite) TestListAmortizationRuns_Contract() {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "Default parameters",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"runs", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "With status filter",
			queryParams:    "status=completed",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"runs", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "With multiple status filters",
			queryParams:    "status=completed,failed",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"runs", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "With date range",
			queryParams:    "date_from=2024-01-01&date_to=2024-12-31",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"runs", "page", "limit", "total", "total_pages"},
		},
		{
			name:           "Invalid date format",
			queryParams:    "date_from=2024-13-45",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid status",
			queryParams:    "status=invalid_status",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock data for successful cases
			if tt.expectedStatus == http.StatusOK {
				suite.mockService.AddAmortizationRun(&AmortizationRun{
					ID:                  uuid.New(),
					Status:              "completed",
					ProcessingDate:      time.Now(),
					TotalAmortizableCIs: 100,
					ProcessedCIs:        intPtr(95),
					FailedCIs:           intPtr(5),
					SkippedCIs:          intPtr(0),
					TotalDepreciation:   floatPtr(5000.0),
					IsManual:            false,
					CreatedAt:           time.Now(),
				})
			}

			url := "/amortization/runs"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(suite.T(), err)
			req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)

			rr := httptest.NewRecorder()
			suite.handler.ListAmortizationRuns(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				// Validate required fields
				for _, field := range tt.expectedFields {
					_, exists := response[field]
					assert.True(suite.T(), exists, "Expected field %s not found in response", field)
				}

				// Validate run structure if present
				if runs, ok := response["runs"].([]interface{}); ok && len(runs) > 0 {
					run := runs[0].(map[string]interface{})
					suite.validateAmortizationRunFields(run)
				}
			}
		})
	}
}

func (suite *AmortizationContractTestSuite) TestGetAmortizationRun_Contract() {
	tests := []struct {
		name           string
		runID          string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "Valid run ID",
			runID:          uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectedFields: []string{
				"id", "status", "processing_date", "started_at", "completed_at",
				"total_amortizable_cis", "processed_cis", "failed_cis", "skipped_cis",
				"total_depreciation", "error_summary", "is_manual", "dry_run",
				"processed_items",
			},
		},
		{
			name:           "Invalid UUID format",
			runID:          "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent run",
			runID:          uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.name == "Valid run ID" {
				runID := uuid.MustParse(tt.runID)
				suite.mockService.AddAmortizationRun(&AmortizationRun{
					ID:                  runID,
					Status:              "completed",
					ProcessingDate:      time.Now(),
					StartedAt:           timePtr(time.Now().Add(-1 * time.Hour)),
					CompletedAt:         timePtr(time.Now()),
					TotalAmortizableCIs: 100,
					ProcessedCIs:        intPtr(95),
					FailedCIs:           intPtr(5),
					SkippedCIs:          intPtr(0),
					TotalDepreciation:   floatPtr(5000.0),
					IsManual:            true,
					DryRun:              false,
					ErrorSummary:        stringPtr(""),
					CreatedAt:           time.Now(),
				})
			}

			url := fmt.Sprintf("/amortization/runs/%s", tt.runID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(suite.T(), err)
			req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)

			rr := httptest.NewRecorder()
			suite.handler.GetAmortizationRun(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				// Validate required fields
				for _, field := range tt.expectedFields {
					_, exists := response[field]
					assert.True(suite.T(), exists, "Expected field %s not found in response", field)
				}

				// Validate field types and structure
				suite.validateAmortizationRunResponse(response)
			}
		})
	}
}

func (suite *AmortizationContractTestSuite) TestTriggerManualRun_Contract() {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "Valid dry run",
			requestBody: map[string]interface{}{
				"dry_run": true,
			},
			expectedStatus: http.StatusAccepted,
			expectedFields: []string{"run_id", "status", "message"},
		},
		{
			name: "Valid run with specific CIs",
			requestBody: map[string]interface{}{
				"dry_run": false,
				"ci_ids":  []string{uuid.New().String(), uuid.New().String()},
			},
			expectedStatus: http.StatusAccepted,
			expectedFields: []string{"run_id", "status", "message"},
		},
		{
			name: "Valid run with date override",
			requestBody: map[string]interface{}{
				"dry_run":       false,
				"date_override": "2024-01-01",
			},
			expectedStatus: http.StatusAccepted,
			expectedFields: []string{"run_id", "status", "message"},
		},
		{
			name: "Valid run with all options",
			requestBody: map[string]interface{}{
				"dry_run":       false,
				"ci_ids":        []string{uuid.New().String()},
				"date_override": "2024-01-15",
			},
			expectedStatus: http.StatusAccepted,
			expectedFields: []string{"run_id", "status", "message"},
		},
		{
			name:           "Empty request body",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusAccepted,
			expectedFields: []string{"run_id", "status", "message"},
		},
		{
			name: "Invalid CI ID format",
			requestBody: map[string]interface{}{
				"ci_ids": []string{"invalid-uuid"},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid date format",
			requestBody: map[string]interface{}{
				"date_override": "2024-13-45",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			requestBody:    map[string]interface{}{"invalid": json.RawMessage("{invalid")},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock data for CI-specific runs
			if ciIDs, ok := tt.requestBody["ci_ids"].([]interface{}); ok {
				for _, ciID := range ciIDs {
					if ciIDStr, ok := ciID.(string); ok {
						if ciID, err := uuid.Parse(ciIDStr); err == nil {
							suite.mockService.AddAmortizableCI(&AmortizableCI{
								ID:            ciID,
								Name:          "Test Server",
								CIType:        "Server",
								IsAmortizable: true,
							})
						}
					}
				}
			}

			var body bytes.Buffer
			err := json.NewEncoder(&body).Encode(tt.requestBody)
			require.NoError(suite.T(), err)

			req, err := http.NewRequest(http.MethodPost, "/amortization/runs", &body)
			require.NoError(suite.T(), err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)

			rr := httptest.NewRecorder()
			suite.handler.TriggerManualRun(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusAccepted {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				// Validate required fields
				for _, field := range tt.expectedFields {
					_, exists := response[field]
					assert.True(suite.T(), exists, "Expected field %s not found in response", field)
				}

				// Validate response structure
				assert.Equal(suite.T(), "started", response["status"])
				assert.Equal(suite.T(), "Amortization run initiated", response["message"])

				// Validate run_id is a valid UUID
				if runID, ok := response["run_id"].(string); ok {
					_, err := uuid.Parse(runID)
					assert.NoError(suite.T(), err, "run_id should be a valid UUID")
				}
			}
		})
	}
}

// Test Reports and Summaries Endpoints

func (suite *AmortizationContractTestSuite) TestGetAmortizationSummaries_Contract() {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "Default group by ci_type",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"summaries", "total_book_value", "total_accumulated_depreciation", "date_as_of"},
		},
		{
			name:           "Group by lifecycle_status",
			queryParams:    "group_by=lifecycle_status",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"summaries", "total_book_value", "total_accumulated_depreciation", "date_as_of"},
		},
		{
			name:           "Group by age_bucket",
			queryParams:    "group_by=age_bucket",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"summaries", "total_book_value", "total_accumulated_depreciation", "date_as_of"},
		},
		{
			name:           "Group by depreciation_method",
			queryParams:    "group_by=depreciation_method",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"summaries", "total_book_value", "total_accumulated_depreciation", "date_as_of"},
		},
		{
			name:           "With date as of",
			queryParams:    "date_as_of=2024-06-30",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"summaries", "total_book_value", "total_accumulated_depreciation", "date_as_of"},
		},
		{
			name:           "Include zero value assets",
			queryParams:    "include_zero_value=true",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"summaries", "total_book_value", "total_accumulated_depreciation", "date_as_of"},
		},
		{
			name:           "Invalid group_by",
			queryParams:    "group_by=invalid_field",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid date format",
			queryParams:    "date_as_of=2024-13-45",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock data for successful cases
			if tt.expectedStatus == http.StatusOK {
				suite.mockService.AddAmortizationSummary(&AmortizationSummary{
					GroupBy:     "ci_type",
					TotalCIs:    50,
					GeneratedAt: time.Now(),
					Groups: []AmortizationGroup{
						{
							GroupName:         "Server",
							CICount:           25,
							TotalBookValue:    250000.0,
							TotalDepreciation: 50000.0,
							AverageAge:        24.5,
						},
					},
				})
			}

			url := "/amortization/summaries"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(suite.T(), err)
			req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)

			rr := httptest.NewRecorder()
			suite.handler.GetAmortizationSummaries(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				// Validate required fields
				for _, field := range tt.expectedFields {
					_, exists := response[field]
					assert.True(suite.T(), exists, "Expected field %s not found in response", field)
				}

				// Validate field types
				assert.IsType(suite.T(), []interface{}{}, response["summaries"])
				assert.IsType(suite.T(), 0.0, response["total_book_value"])
				assert.IsType(suite.T(), 0.0, response["total_accumulated_depreciation"])
				assert.IsType(suite.T(), "", response["date_as_of"])

				// Validate summary structure if present
				if summaries, ok := response["summaries"].([]interface{}); ok && len(summaries) > 0 {
					summary := summaries[0].(map[string]interface{})
					suite.validateAmortizationSummaryFields(summary)
				}
			}
		})
	}
}

func (suite *AmortizationContractTestSuite) TestGenerateDepreciationSchedule_Contract() {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedFormat string
	}{
		{
			name:           "JSON format default",
			queryParams:    "date_from=2024-01-01&date_to=2024-12-31",
			expectedStatus: http.StatusOK,
			expectedFormat: "json",
		},
		{
			name:           "JSON format explicit",
			queryParams:    "date_from=2024-01-01&date_to=2024-12-31&format=json",
			expectedStatus: http.StatusOK,
			expectedFormat: "json",
		},
		{
			name:           "CSV format",
			queryParams:    "date_from=2024-01-01&date_to=2024-12-31&format=csv",
			expectedStatus: http.StatusOK,
			expectedFormat: "csv",
		},
		{
			name:           "With specific CIs",
			queryParams:    fmt.Sprintf("date_from=2024-01-01&date_to=2024-12-31&ci_ids=%s", uuid.New().String()),
			expectedStatus: http.StatusOK,
			expectedFormat: "json",
		},
		{
			name:           "Missing date_from",
			queryParams:    "date_to=2024-12-31",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing date_to",
			queryParams:    "date_from=2024-01-01",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid date format",
			queryParams:    "date_from=2024-13-45&date_to=2024-12-31",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid format",
			queryParams:    "date_from=2024-01-01&date_to=2024-12-31&format=invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid CI ID",
			queryParams:    "date_from=2024-01-01&date_to=2024-12-31&ci_ids=invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock data for successful cases
			if tt.expectedStatus == http.StatusOK {
				ciID := uuid.New()
				suite.mockService.AddAmortizableCI(&AmortizableCI{
					ID:            ciID,
					Name:          "Test Server",
					CIType:        "Server",
					IsAmortizable: true,
				})

				reportID := uuid.New()
				suite.mockService.AddDepreciationSchedule(&DepreciationSchedule{
					ReportID: reportID,
					DateRange: DepreciationScheduleRange{
						StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						EndDate:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
					},
					Schedule: []DepreciationScheduleEntry{
						{
							CIID:                    ciID,
							CIName:                  "Test Server",
							CIType:                  "Server",
							PeriodStart:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							PeriodEnd:               time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
							OpeningBookValue:        10000.0,
							DepreciationAmount:      166.67,
							ClosingBookValue:        9833.33,
							AccumulatedDepreciation: 166.67,
						},
					},
				})
			}

			url := "/amortization/reports/depreciation-schedule"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(suite.T(), err)
			req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)

			rr := httptest.NewRecorder()
			suite.handler.GenerateDepreciationSchedule(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				if tt.expectedFormat == "json" {
					assert.Equal(suite.T(), "application/json", rr.Header().Get("Content-Type"))

					var response map[string]interface{}
					err := json.Unmarshal(rr.Body.Bytes(), &response)
					require.NoError(suite.T(), err)

					// Validate response structure
					expectedFields := []string{"report_id", "date_range", "schedule"}
					for _, field := range expectedFields {
						_, exists := response[field]
						assert.True(suite.T(), exists, "Expected field %s not found in response", field)
					}

					// Validate date range structure
					if dateRange, ok := response["date_range"].(map[string]interface{}); ok {
						_, hasStart := dateRange["start_date"]
						_, hasEnd := dateRange["end_date"]
						assert.True(suite.T(), hasStart, "date_range should have start_date")
						assert.True(suite.T(), hasEnd, "date_range should have end_date")
					}

					// Validate schedule structure
					if schedule, ok := response["schedule"].([]interface{}); ok && len(schedule) > 0 {
						entry := schedule[0].(map[string]interface{})
						suite.validateDepreciationScheduleEntry(entry)
					}
				} else if tt.expectedFormat == "csv" {
					assert.Equal(suite.T(), "text/csv", rr.Header().Get("Content-Type"))
					assert.Contains(suite.T(), rr.Header().Get("Content-Disposition"), "attachment")
					assert.Contains(suite.T(), rr.Header().Get("Content-Disposition"), "depreciation_schedule")

					// Validate CSV content
					csvContent := rr.Body.String()
					assert.Contains(suite.T(), csvContent, "CI ID,CI Name,CI Type")
				}
			}
		})
	}
}

// Authentication and Authorization Tests

func (suite *AmortizationContractTestSuite) TestAuthentication_Contract() {
	tests := []struct {
		name           string
		authToken      string
		expectedStatus int
	}{
		{
			name:           "Valid JWT token",
			authToken:      suite.validAuthToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing token",
			authToken:      "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token format",
			authToken:      "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Expired token",
			authToken:      "expired.jwt.token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock data
			ciID := uuid.New()
			suite.mockService.AddAmortizableCI(&AmortizableCI{
				ID:            ciID,
				Name:          "Test Server",
				CIType:        "Server",
				IsAmortizable: true,
			})

			url := fmt.Sprintf("/amortization/configuration-items/%s", ciID.String())
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(suite.T(), err)

			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
			}

			rr := httptest.NewRecorder()
			suite.handler.GetAmortizationDetails(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusUnauthorized {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				// Validate error response structure
				if errorObj, ok := response["error"].(map[string]interface{}); ok {
					assert.Contains(suite.T(), errorObj, "message")
					assert.Contains(suite.T(), errorObj, "code")
				}
			}
		})
	}
}

func (suite *AmortizationContractTestSuite) TestAuthorization_Contract() {
	tests := []struct {
		name           string
		permissions    []string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "Admin with full permissions",
			permissions:    []string{"amortization:read", "amortization:write", "amortization:adjust", "amortization:admin"},
			method:         http.MethodGet,
			url:            "/amortization/configuration-items",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Read-only user accessing read endpoint",
			permissions:    []string{"amortization:read"},
			method:         http.MethodGet,
			url:            "/amortization/configuration-items",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Read-only user accessing write endpoint",
			permissions:    []string{"amortization:read"},
			method:         http.MethodPut,
			url:            "/amortization/configuration-items/" + uuid.New().String(),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "User without adjust permission",
			permissions:    []string{"amortization:read", "amortization:write"},
			method:         http.MethodPost,
			url:            "/amortization/adjustments",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "User without admin permission",
			permissions:    []string{"amortization:read", "amortization:write", "amortization:adjust"},
			method:         http.MethodPost,
			url:            "/amortization/runs",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "User with no amortization permissions",
			permissions:    []string{},
			method:         http.MethodGet,
			url:            "/amortization/configuration-items",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup mock data for needed endpoints
			if tt.method == http.MethodPut {
				ciID := uuid.New()
				suite.mockService.AddAmortizableCI(&AmortizableCI{
					ID:            ciID,
					Name:          "Test Server",
					CIType:        "Server",
					IsAmortizable: true,
				})
				tt.url = "/amortization/configuration-items/" + ciID.String()
			}

			var body bytes.Buffer
			if tt.method == http.MethodPut {
				json.NewEncoder(&body).Encode(map[string]interface{}{
					"purchase_cost": 10000.0,
				})
			} else if tt.method == http.MethodPost && tt.url == "/amortization/adjustments" {
				ciID := uuid.New()
				suite.mockService.AddAmortizableCI(&AmortizableCI{
					ID:            ciID,
					Name:          "Test Server",
					CIType:        "Server",
					IsAmortizable: true,
				})
				json.NewEncoder(&body).Encode(map[string]interface{}{
					"ci_id":       ciID.String(),
					"amount":      100.0,
					"description": "Test adjustment",
				})
			} else if tt.method == http.MethodPost && tt.url == "/amortization/runs" {
				json.NewEncoder(&body).Encode(map[string]interface{}{
					"dry_run": true,
				})
			}

			req, err := http.NewRequest(tt.method, tt.url, &body)
			require.NoError(suite.T(), err)
			req.Header.Set("Content-Type", "application/json")

			// Create custom middleware with specified permissions
			authMiddleware := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					user := &middleware.AuthenticatedUser{
						UserID:      suite.userUserID,
						Username:    "user",
						Email:       "user@example.com",
						Role:        "user",
						Permissions: tt.permissions,
					}
					ctx := context.WithValue(r.Context(), middleware.UserContextKey, user)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			}

			finalRouter := authMiddleware(suite.handler)
			rr := httptest.NewRecorder()
			finalRouter.ServeHTTP(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusForbidden {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				// Validate error response structure
				if errorObj, ok := response["error"].(map[string]interface{}); ok {
					assert.Contains(suite.T(), errorObj, "message")
					assert.Equal(suite.T(), "Forbidden", errorObj["code"])
				}
			}
		})
	}
}

// Validation Helper Methods

func (suite *AmortizationContractTestSuite) validateAmortizableCIFields(ci map[string]interface{}) {
	expectedFields := []string{
		"id", "name", "ci_type", "purchase_cost", "salvage_value",
		"amort_start_date", "useful_life_months", "current_book_value",
		"accumulated_depreciation", "amortization_behavior", "created_at",
	}

	for _, field := range expectedFields {
		_, exists := ci[field]
		assert.True(suite.T(), exists, "Expected field %s not found in AmortizableCI", field)
	}

	// Validate field types
	assert.IsType(suite.T(), "", ci["id"])
	assert.IsType(suite.T(), "", ci["name"])
	assert.IsType(suite.T(), "", ci["ci_type"])
	assert.IsType(suite.T(), 0.0, ci["purchase_cost"])
	assert.IsType(suite.T(), 0.0, ci["salvage_value"])
	assert.IsType(suite.T(), "", ci["amort_start_date"])
	assert.IsType(suite.T(), 0, ci["useful_life_months"])
	assert.IsType(suite.T(), 0.0, ci["current_book_value"])
	assert.IsType(suite.T(), 0.0, ci["accumulated_depreciation"])
	assert.IsType(suite.T(), "", ci["amortization_behavior"])
}

func (suite *AmortizationContractTestSuite) validateAmortizationDetails(details map[string]interface{}) {
	suite.validateAmortizableCIFields(details)

	// Additional fields for details
	detailFields := []string{
		"depreciation_method", "monthly_depreciation",
		"remaining_life_months", "recent_ledger_entries",
	}

	for _, field := range detailFields {
		_, exists := details[field]
		assert.True(suite.T(), exists, "Expected field %s not found in AmortizationDetails", field)
	}

	// Validate calculated fields
	if monthlyDepreciation, ok := details["monthly_depreciation"]; ok {
		assert.IsType(suite.T(), 0.0, monthlyDepreciation)
	}
	if remainingLife, ok := details["remaining_life_months"]; ok {
		assert.IsType(suite.T(), 0, remainingLife)
	}
	if recentEntries, ok := details["recent_ledger_entries"]; ok {
		assert.IsType(suite.T(), []interface{}{}, recentEntries)
	}
}

func (suite *AmortizationContractTestSuite) validateLedgerEntryFields(entry map[string]interface{}) {
	expectedFields := []string{
		"id", "ci_id", "entry_type", "entry_date", "amount",
		"book_value_before", "book_value_after", "accumulated_depreciation_before",
		"accumulated_depreciation_after", "created_at",
	}

	for _, field := range expectedFields {
		_, exists := entry[field]
		assert.True(suite.T(), exists, "Expected field %s not found in LedgerEntry", field)
	}

	// Validate field types
	assert.IsType(suite.T(), "", entry["id"])
	assert.IsType(suite.T(), "", entry["ci_id"])
	assert.IsType(suite.T(), "", entry["entry_type"])
	assert.IsType(suite.T(), "", entry["entry_date"])
	assert.IsType(suite.T(), 0.0, entry["amount"])
	assert.IsType(suite.T(), 0.0, entry["book_value_before"])
	assert.IsType(suite.T(), 0.0, entry["book_value_after"])
	assert.IsType(suite.T(), 0.0, entry["accumulated_depreciation_before"])
	assert.IsType(suite.T(), 0.0, entry["accumulated_depreciation_after"])
	assert.IsType(suite.T(), "", entry["created_at"])

	// Validate entry type is valid
	if entryType, ok := entry["entry_type"].(string); ok {
		validTypes := []string{"monthly_depreciation", "write_off", "adjustment", "correction"}
		assert.Contains(suite.T(), validTypes, entryType, "Invalid entry_type: %s", entryType)
	}
}

func (suite *AmortizationContractTestSuite) validateLedgerEntryResponse(response map[string]interface{}) {
	suite.validateLedgerEntryFields(response)

	// Additional fields for response
	_, exists := response["ci_details"]
	assert.True(suite.T(), exists, "Expected field ci_details not found in LedgerEntryResponse")
}

func (suite *AmortizationContractTestSuite) validateAmortizationRunFields(run map[string]interface{}) {
	expectedFields := []string{
		"id", "status", "processing_date", "total_amortizable_cis",
		"created_at", "is_manual", "dry_run",
	}

	for _, field := range expectedFields {
		_, exists := run[field]
		assert.True(suite.T(), exists, "Expected field %s not found in AmortizationRun", field)
	}

	// Validate field types
	assert.IsType(suite.T(), "", run["id"])
	assert.IsType(suite.T(), "", run["status"])
	assert.IsType(suite.T(), "", run["processing_date"])
	assert.IsType(suite.T(), 0, run["total_amortizable_cis"])
	assert.IsType(suite.T(), "", run["created_at"])
	assert.IsType(suite.T(), false, run["is_manual"])
	assert.IsType(suite.T(), false, run["dry_run"])

	// Validate status is valid
	if status, ok := run["status"].(string); ok {
		validStatuses := []string{"started", "completed", "failed", "partial"}
		assert.Contains(suite.T(), validStatuses, status, "Invalid status: %s", status)
	}
}

func (suite *AmortizationContractTestSuite) validateAmortizationRunResponse(response map[string]interface{}) {
	suite.validateAmortizationRunFields(response)

	// Additional fields for response
	responseFields := []string{
		"started_at", "completed_at", "processed_cis", "failed_cis",
		"skipped_cis", "total_depreciation", "error_summary", "processed_items",
	}

	for _, field := range responseFields {
		_, exists := response[field]
		assert.True(suite.T(), exists, "Expected field %s not found in AmortizationRunResponse", field)
	}

	// Validate processed items structure
	if processedItems, ok := response["processed_items"].([]interface{}); ok {
		for _, item := range processedItems {
			if itemMap, ok := item.(map[string]interface{}); ok {
				_, hasCIID := itemMap["ci_id"]
				_, hasStatus := itemMap["status"]
				assert.True(suite.T(), hasCIID, "processed_items should have ci_id")
				assert.True(suite.T(), hasStatus, "processed_items should have status")
			}
		}
	}
}

func (suite *AmortizationContractTestSuite) validateAmortizationSummaryFields(summary map[string]interface{}) {
	expectedFields := []string{
		"group_key", "group_label", "ci_count", "total_purchase_cost",
		"total_book_value", "total_accumulated_depreciation", "depreciation_percentage",
	}

	for _, field := range expectedFields {
		_, exists := summary[field]
		assert.True(suite.T(), exists, "Expected field %s not found in AmortizationSummary", field)
	}

	// Validate field types
	assert.IsType(suite.T(), "", summary["group_key"])
	assert.IsType(suite.T(), "", summary["group_label"])
	assert.IsType(suite.T(), 0, summary["ci_count"])
	assert.IsType(suite.T(), 0.0, summary["total_purchase_cost"])
	assert.IsType(suite.T(), 0.0, summary["total_book_value"])
	assert.IsType(suite.T(), 0.0, summary["total_accumulated_depreciation"])
	assert.IsType(suite.T(), 0.0, summary["depreciation_percentage"])
}

func (suite *AmortizationContractTestSuite) validateDepreciationScheduleEntry(entry map[string]interface{}) {
	expectedFields := []string{
		"ci_id", "ci_name", "ci_type", "period_start", "period_end",
		"opening_book_value", "depreciation_amount", "closing_book_value",
		"accumulated_depreciation",
	}

	for _, field := range expectedFields {
		_, exists := entry[field]
		assert.True(suite.T(), exists, "Expected field %s not found in DepreciationScheduleEntry", field)
	}

	// Validate field types
	assert.IsType(suite.T(), "", entry["ci_id"])
	assert.IsType(suite.T(), "", entry["ci_name"])
	assert.IsType(suite.T(), "", entry["ci_type"])
	assert.IsType(suite.T(), "", entry["period_start"])
	assert.IsType(suite.T(), "", entry["period_end"])
	assert.IsType(suite.T(), 0.0, entry["opening_book_value"])
	assert.IsType(suite.T(), 0.0, entry["depreciation_amount"])
	assert.IsType(suite.T(), 0.0, entry["closing_book_value"])
	assert.IsType(suite.T(), 0.0, entry["accumulated_depreciation"])
}

// Performance and Load Tests

func (suite *AmortizationContractTestSuite) TestPerformance_Contract() {
	tests := []struct {
		name         string
		setupData    func()
		request      func() *http.Request
		maxDuration  time.Duration
		expectedCode int
	}{
		{
			name: "List amortizable CIs performance",
			setupData: func() {
				for i := 0; i < 100; i++ {
					ci := &AmortizableCI{
						ID:               uuid.New(),
						Name:             fmt.Sprintf("Server %d", i),
						CIType:           "Server",
						PurchaseCost:     10000.0,
						UsefulLifeMonths: 60,
						CurrentBookValue: 8000.0,
						IsAmortizable:    true,
					}
					suite.mockService.AddAmortizableCI(ci)
				}
			},
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "/amortization/configuration-items?limit=50", nil)
				req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)
				return req
			},
			maxDuration:  100 * time.Millisecond,
			expectedCode: http.StatusOK,
		},
		{
			name: "Get amortization details performance",
			setupData: func() {
				ci := &AmortizableCI{
					ID:               uuid.New(),
					Name:             "Performance Test Server",
					CIType:           "Server",
					PurchaseCost:     10000.0,
					UsefulLifeMonths: 60,
					CurrentBookValue: 8000.0,
					IsAmortizable:    true,
				}
				suite.mockService.AddAmortizableCI(ci)
			},
			request: func() *http.Request {
				ciID := suite.mockService.cis[0].ID.String()
				req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/amortization/configuration-items/%s", ciID), nil)
				req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)
				return req
			},
			maxDuration:  50 * time.Millisecond,
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setupData()

			start := time.Now()
			rr := httptest.NewRecorder()

			var handler func(http.ResponseWriter, *http.Request)
			if strings.Contains(tt.request().URL.Path, "configuration-items") && tt.request().Method == http.MethodGet {
				if strings.Contains(tt.request().URL.String(), "limit=") {
					handler = suite.handler.ListAmortizableCIs
				} else {
					handler = suite.handler.GetAmortizationDetails
				}
			}

			handler(rr, tt.request())
			duration := time.Since(start)

			assert.Equal(suite.T(), tt.expectedCode, rr.Code)
			assert.Less(suite.T(), duration, tt.maxDuration,
				"Request took %v, expected less than %v", duration, tt.maxDuration)
		})
	}
}

// Edge Cases and Error Handling

func (suite *AmortizationContractTestSuite) TestEdgeCases_Contract() {
	tests := []struct {
		name           string
		setupMock      func()
		request        func() *http.Request
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Maximum pagination limit",
			setupMock: func() {
				for i := 0; i < 200; i++ {
					ci := &AmortizableCI{
						ID:            uuid.New(),
						Name:          fmt.Sprintf("Server %d", i),
						CIType:        "Server",
						PurchaseCost:  10000.0,
						IsAmortizable: true,
					}
					suite.mockService.AddAmortizableCI(ci)
				}
			},
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "/amortization/configuration-items?limit=200", nil)
				req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Zero value financials",
			setupMock: func() {
				ci := &AmortizableCI{
					ID:                      uuid.New(),
					Name:                    "Zero Value Server",
					CIType:                  "Server",
					PurchaseCost:            0.0,
					SalvageValue:            0.0,
					CurrentBookValue:        0.0,
					AccumulatedDepreciation: 0.0,
					IsAmortizable:           true,
				}
				suite.mockService.AddAmortizableCI(ci)
			},
			request: func() *http.Request {
				ciID := suite.mockService.cis[0].ID.String()
				req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/amortization/configuration-items/%s", ciID), nil)
				req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Very large financial values",
			setupMock: func() {
				ci := &AmortizableCI{
					ID:                      uuid.New(),
					Name:                    "High Value Server",
					CIType:                  "Server",
					PurchaseCost:            999999999.99,
					SalvageValue:            1000000.00,
					CurrentBookValue:        500000000.00,
					AccumulatedDepreciation: 499999999.99,
					IsAmortizable:           true,
				}
				suite.mockService.AddAmortizableCI(ci)
			},
			request: func() *http.Request {
				ciID := suite.mockService.cis[0].ID.String()
				req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/amortization/configuration-items/%s", ciID), nil)
				req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Special characters in CI name",
			setupMock: func() {
				ci := &AmortizableCI{
					ID:            uuid.New(),
					Name:          "Server-测试_🖥️@#$%^&*()",
					CIType:        "Server",
					PurchaseCost:  10000.0,
					IsAmortizable: true,
				}
				suite.mockService.AddAmortizableCI(ci)
			},
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "/amortization/configuration-items?search=Server-测试", nil)
				req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Unicode in description fields",
			setupMock: func() {
				ciID := uuid.New()
				suite.mockService.AddAmortizableCI(&AmortizableCI{
					ID:            ciID,
					Name:          "Unicode Server",
					CIType:        "Server",
					PurchaseCost:  10000.0,
					IsAmortizable: true,
				})
				description := "Adjustment for currency revaluation Ðóõ₤€£¥١٢٣"
				suite.mockService.AddLedgerEntry(&AmortizationEntry{
					ID:          uuid.New(),
					CIID:        ciID,
					EntryType:   "adjustment",
					EntryDate:   time.Now(),
					Amount:      1000.0,
					Description: &description,
				})
			},
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "/amortization/ledger", nil)
				req.Header.Set("Authorization", "Bearer "+suite.validAuthToken)
				return req
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setupMock()

			rr := httptest.NewRecorder()

			// Determine handler based on request
			var handler func(http.ResponseWriter, *http.Request)
			req := tt.request()
			path := req.URL.Path

			if strings.Contains(path, "/configuration-items") && strings.Contains(path, "{") == false {
				if req.Method == http.MethodGet {
					handler = suite.handler.ListAmortizableCIs
				} else {
					handler = suite.handler.GetAmortizationDetails
				}
			} else if strings.Contains(path, "/ledger") {
				handler = suite.handler.GetLedgerEntries
			}

			handler(rr, req)

			assert.Equal(suite.T(), tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(suite.T(), err)

				if errorObj, ok := response["error"].(map[string]interface{}); ok {
					if message, ok := errorObj["message"].(string); ok {
						assert.Contains(suite.T(), message, tt.expectedError)
					}
				}
			}
		})
	}
}

// Test Suite Runner

func TestAmortizationContractTestSuite(t *testing.T) {
	suite.Run(t, new(AmortizationContractTestSuite))
}

// Helper Functions

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func stringPtr(s string) *string {
	return &s
}
