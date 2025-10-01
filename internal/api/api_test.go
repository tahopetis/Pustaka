package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/pustaka/pustaka/internal/ci"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// Mock CI Service for testing
type MockCIService struct {
	cis map[uuid.UUID]ci.ConfigurationItem
}

func NewMockCIService() *MockCIService {
	return &MockCIService{
		cis: make(map[uuid.UUID]ci.ConfigurationItem),
	}
}

func (m *MockCIService) CreateCI(ctx context.Context, req *ci.CreateCIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	ci := &ci.ConfigurationItem{
		ID:        uuid.New(),
		Name:      req.Name,
		CIType:    req.CIType,
		Attributes: req.Attributes,
		Tags:      req.Tags,
		CreatedBy: userID,
	}
	m.cis[ci.ID] = *ci
	return ci, nil
}

func (m *MockCIService) GetCI(ctx context.Context, id uuid.UUID) (*ci.ConfigurationItem, error) {
	if ci, exists := m.cis[id]; exists {
		return &ci, nil
	}
	return nil, fmt.Errorf("CI not found")
}

func (m *MockCIService) ListCIs(ctx context.Context, filters ci.ListCIFilters, page, limit int) (*ci.CIListResponse, error) {
	var cis []ci.ConfigurationItem
	for _, ci := range m.cis {
		cis = append(cis, ci)
	}
	return &ci.CIListResponse{
		CIs: cis,
		Pagination: ci.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      int64(len(cis)),
			TotalPages: 1,
		},
	}, nil
}

// Implement other required methods with minimal functionality
func (m *MockCIService) UpdateCI(ctx context.Context, id uuid.UUID, req *ci.UpdateCIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	return nil, nil
}
func (m *MockCIService) DeleteCI(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return nil
}
func (m *MockCIService) CreateCIType(ctx context.Context, req *ci.CreateCITypeRequest, userID uuid.UUID) (*ci.CITypeDefinition, error) {
	return nil, nil
}
func (m *MockCIService) GetCIType(ctx context.Context, id uuid.UUID) (*ci.CITypeDefinition, error) {
	return nil, nil
}
func (m *MockCIService) ListCITypes(ctx context.Context, page, limit int, search string) (*ci.CITypeListResponse, error) {
	return nil, nil
}
func (m *MockCIService) UpdateCIType(ctx context.Context, id uuid.UUID, req *ci.UpdateCITypeRequest, userID uuid.UUID) (*ci.CITypeDefinition, error) {
	return nil, nil
}
func (m *MockCIService) DeleteCIType(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return nil
}
func (m *MockCIService) CreateRelationship(ctx context.Context, req *ci.CreateRelationshipRequest, userID uuid.UUID) (*ci.Relationship, error) {
	return nil, nil
}
func (m *MockCIService) GetRelationship(ctx context.Context, id uuid.UUID) (*ci.Relationship, error) {
	return nil, nil
}
func (m *MockCIService) ListRelationships(ctx context.Context, filters ci.ListRelationshipFilters, page, limit int) (*ci.RelationshipListResponse, error) {
	return nil, nil
}
func (m *MockCIService) UpdateRelationship(ctx context.Context, id uuid.UUID, req *ci.UpdateRelationshipRequest, userID uuid.UUID) (*ci.Relationship, error) {
	return nil, nil
}
func (m *MockCIService) DeleteRelationship(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return nil
}
func (m *MockCIService) GetCIRelationships(ctx context.Context, id uuid.UUID) ([]ci.RelationshipGraph, error) {
	return nil, nil
}
func (m *MockCIService) GetGraphData(ctx context.Context, filters ci.GraphFilters) (*ci.GraphData, error) {
	return nil, nil
}
func (m *MockCIService) GetCINetwork(ctx context.Context, id uuid.UUID, depth int) (*ci.CINetwork, error) {
	return nil, nil
}
func (m *MockCIService) GetImpactAnalysis(ctx context.Context, id uuid.UUID) (*ci.ImpactAnalysis, error) {
	return nil, nil
}
func (m *MockCIService) GetCITypesByUsage(ctx context.Context) ([]ci.CITypeUsage, error) {
	return nil, nil
}
func (m *MockCIService) FindCycles(ctx context.Context) ([][]uuid.UUID, error) {
	return nil, nil
}
func (m *MockCIService) GetMostConnectedCIs(ctx context.Context, limit int) ([]ci.CIConnectivity, error) {
	return nil, nil
}

func TestHealthCheck(t *testing.T) {
	// Setup
	logger := pustakaLogger.NewLogger()
	mockService := NewMockCIService()
	router := NewRouter(mockService, logger)

	// Test health check
	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status to be 'healthy', got %v", response["status"])
	}
}

func TestCreateCI(t *testing.T) {
	// Setup
	logger := pustakaLogger.NewLogger()
	mockService := NewMockCIService()
	router := NewRouter(mockService, logger)

	// Test data
	ciRequest := ci.CreateCIRequest{
		Name:   "Test Server",
		CIType: "Server",
		Attributes: map[string]interface{}{
			"ip_address": "192.168.1.100",
			"os":         "Ubuntu 20.04",
		},
		Tags: []string{"production", "web"},
	}

	requestBody, _ := json.Marshal(ciRequest)
	req, err := http.NewRequest("POST", "/api/v1/ci", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	var response ci.ConfigurationItem
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Name != "Test Server" {
		t.Errorf("Expected name to be 'Test Server', got %v", response.Name)
	}

	if response.CIType != "Server" {
		t.Errorf("Expected CI type to be 'Server', got %v", response.CIType)
	}
}

func TestListCIs(t *testing.T) {
	// Setup
	logger := pustakaLogger.NewLogger()
	mockService := NewMockCIService()
	router := NewRouter(mockService, logger)

	// Create a CI first
	userID := uuid.New()
	ciRequest := ci.CreateCIRequest{
		Name:   "Test Server",
		CIType: "Server",
		Attributes: map[string]interface{}{
			"ip_address": "192.168.1.100",
		},
	}
	mockService.CreateCI(context.Background(), &ciRequest, userID)

	// Test listing CIs
	req, err := http.NewRequest("GET", "/api/v1/ci", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.GetRouter().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response ci.CIListResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(response.CIs) != 1 {
		t.Errorf("Expected 1 CI, got %v", len(response.CIs))
	}
}

func TestSchemaValidation(t *testing.T) {
	// Test schema validation logic
	serverType := &ci.CITypeDefinition{
		Name: "Server",
		RequiredAttributes: []ci.AttributeDefinition{
			{
				Name: "hostname",
				Type: "string",
				Validation: &ci.AttributeValidation{
					MinLength: intPtr(1),
					MaxLength: intPtr(255),
				},
			},
			{
				Name: "ip_address",
				Type: "string",
				Validation: &ci.AttributeValidation{
					Format: "ipv4",
				},
			},
		},
		OptionalAttributes: []ci.AttributeDefinition{
			{
				Name: "port",
				Type: "integer",
				Validation: &ci.AttributeValidation{
					Min: intPtr(1),
					Max: intPtr(65535),
				},
			},
		},
	}

	// Test valid attributes
	validAttributes := map[string]interface{}{
		"hostname":   "web-server-01",
		"ip_address": "192.168.1.100",
		"port":       80,
	}
	errors := serverType.ValidateAttributes(validAttributes)
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors for valid attributes, got %v", errors)
	}

	// Test missing required attribute
	missingRequired := map[string]interface{}{
		"ip_address": "192.168.1.100",
	}
	errors = serverType.ValidateAttributes(missingRequired)
	if len(errors) == 0 {
		t.Error("Expected validation error for missing required attribute")
	}

	// Test invalid IP format
	invalidIP := map[string]interface{}{
		"hostname":   "web-server-01",
		"ip_address": "invalid-ip",
	}
	errors = serverType.ValidateAttributes(invalidIP)
	if len(errors) == 0 {
		t.Error("Expected validation error for invalid IP format")
	}

	// Test invalid port range
	invalidPort := map[string]interface{}{
		"hostname":   "web-server-01",
		"ip_address": "192.168.1.100",
		"port":       99999,
	}
	errors = serverType.ValidateAttributes(invalidPort)
	if len(errors) == 0 {
		t.Error("Expected validation error for invalid port range")
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}