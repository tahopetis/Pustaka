package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/pustaka/pustaka/internal/ci"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// MockCITypeService is a mock implementation of the CI type service interface
type MockCITypeService struct {
	mock.Mock
}

func (m *MockCITypeService) CreateCIType(ctx context.Context, req *ci.CreateCITypeRequest, userID uuid.UUID) (*ci.CITypeDefinition, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ci.CITypeDefinition), args.Error(1)
}

func (m *MockCITypeService) GetCIType(ctx context.Context, ciTypeID uuid.UUID) (*ci.CITypeDefinition, error) {
	args := m.Called(ctx, ciTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ci.CITypeDefinition), args.Error(1)
}

func (m *MockCITypeService) ListCITypes(ctx context.Context, filters map[string]interface{}, page, limit int) (*ci.CITypeListResponse, error) {
	args := m.Called(ctx, filters, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ci.CITypeListResponse), args.Error(1)
}

func (m *MockCITypeService) UpdateCIType(ctx context.Context, ciTypeID uuid.UUID, req *ci.UpdateCITypeRequest, userID uuid.UUID) (*ci.CITypeDefinition, error) {
	args := m.Called(ctx, ciTypeID, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ci.CITypeDefinition), args.Error(1)
}

func (m *MockCITypeService) DeleteCIType(ctx context.Context, ciTypeID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, ciTypeID, userID)
	return args.Error(0)
}

// CITypeHandlerSuite contains tests for CI type handlers
type CITypeHandlerSuite struct {
	suite.Suite
	mockCIType *MockCITypeService
	logger     *pustakaLogger.Logger
}

func (suite *CITypeHandlerSuite) SetupTest() {
	suite.mockCIType = new(MockCITypeService)
	suite.logger = pustakaLogger.New(pustakaLogger.Config{Level: "info"})
}

func (suite *CITypeHandlerSuite) TestCreateCIType() {
	handler := &Handler{logger: suite.logger}
	ciTypeHandlers := NewCITypeHandlers(handler, suite.mockCIType)

	userID := uuid.New()
	testCIType := &ci.CITypeDefinition{
		ID:                 uuid.New(),
		Name:               "Server",
		Description:        "Physical or virtual server",
		RequiredAttributes: []ci.AttributeDefinition{{Name: "hostname", Type: "string", Required: true}},
		OptionalAttributes: []ci.AttributeDefinition{{Name: "ip_address", Type: "string", Required: false}},
	}

	reqBody := ci.CreateCITypeRequest{
		Name:        "Server",
		Description: "Physical or virtual server",
		RequiredAttributes: []ci.AttributeDefinition{
			{Name: "hostname", Type: "string", Required: true},
		},
		OptionalAttributes: []ci.AttributeDefinition{
			{Name: "ip_address", Type: "string", Required: false},
		},
	}

	suite.mockCIType.On("CreateCIType", mock.Anything, &reqBody, userID).Return(testCIType, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/ci-types", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	ciTypeHandlers.CreateCIType(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response ci.CITypeDefinition
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), testCIType.Name, response.Name)
	assert.Equal(suite.T(), testCIType.Description, response.Description)

	suite.mockCIType.AssertExpectations(suite.T())
}

func (suite *CITypeHandlerSuite) TestCreateCITypeValidationError() {
	handler := &Handler{logger: suite.logger}
	ciTypeHandlers := NewCITypeHandlers(handler, suite.mockCIType)

	userID := uuid.New()
	reqBody := ci.CreateCITypeRequest{
		Name:        "", // Empty name should cause validation error
		Description: "Test CI type",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/ci-types", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	ciTypeHandlers.CreateCIType(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Name is required", response["error"])
}

func (suite *CITypeHandlerSuite) TestGetCIType() {
	handler := &Handler{logger: suite.logger}
	ciTypeHandlers := NewCITypeHandlers(handler, suite.mockCIType)

	ciTypeID := uuid.New()
	testCIType := &ci.CITypeDefinition{
		ID:          ciTypeID,
		Name:        "Server",
		Description: "Physical or virtual server",
		RequiredAttributes: []ci.AttributeDefinition{
			{Name: "hostname", Type: "string", Required: true},
		},
	}

	suite.mockCIType.On("GetCIType", mock.Anything, ciTypeID).Return(testCIType, nil)

	req := httptest.NewRequest(http.MethodGet, "/ci-types/"+ciTypeID.String(), nil)
	w := httptest.NewRecorder()

	ciTypeHandlers.GetCIType(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response ci.CITypeDefinition
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), testCIType.ID, response.ID)
	assert.Equal(suite.T(), testCIType.Name, response.Name)

	suite.mockCIType.AssertExpectations(suite.T())
}

func (suite *CITypeHandlerSuite) TestGetCITypeNotFound() {
	handler := &Handler{logger: suite.logger}
	ciTypeHandlers := NewCITypeHandlers(handler, suite.mockCIType)

	ciTypeID := uuid.New()
	suite.mockCIType.On("GetCIType", mock.Anything, ciTypeID).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/ci-types/"+ciTypeID.String(), nil)
	w := httptest.NewRecorder()

	ciTypeHandlers.GetCIType(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "CI type definition not found", response["error"])

	suite.mockCIType.AssertExpectations(suite.T())
}

func (suite *CITypeHandlerSuite) TestListCITypes() {
	handler := &Handler{logger: suite.logger}
	ciTypeHandlers := NewCITypeHandlers(handler, suite.mockCIType)

	testResponse := &ci.CITypeListResponse{
		CITypes: []ci.CITypeDefinition{
			{
				ID:          uuid.New(),
				Name:        "Server",
				Description: "Physical or virtual server",
			},
			{
				ID:          uuid.New(),
				Name:        "Application",
				Description: "Software application",
			},
		},
		Total:      2,
		Page:       1,
		Limit:      20,
		TotalPages: 1,
	}

	// For testing, we'll use a simple map as filters since ListCITypeFilters isn't defined yet
	expectedFilters := map[string]interface{}{
		"search": "server",
		"sort":   "name",
		"order":  "asc",
	}

	suite.mockCIType.On("ListCITypes", mock.Anything, expectedFilters, 1, 20).Return(testResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/ci-types?search=server&sort=name&order=asc&page=1&limit=20", nil)
	w := httptest.NewRecorder()

	ciTypeHandlers.ListCITypes(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response ci.CITypeListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), testResponse.Total, response.Total)
	assert.Len(suite.T(), response.CITypes, 2)

	suite.mockCIType.AssertExpectations(suite.T())
}

func (suite *CITypeHandlerSuite) TestUpdateCIType() {
	handler := &Handler{logger: suite.logger}
	ciTypeHandlers := NewCITypeHandlers(handler, suite.mockCIType)

	userID := uuid.New()
	ciTypeID := uuid.New()

	updatedCIType := &ci.CITypeDefinition{
		ID:          ciTypeID,
		Name:        "Updated Server",
		Description: "Updated description",
		RequiredAttributes: []ci.AttributeDefinition{
			{Name: "hostname", Type: "string", Required: true},
		},
	}

	reqBody := ci.UpdateCITypeRequest{
		Description: "Updated description",
		RequiredAttributes: []ci.AttributeDefinition{
			{Name: "hostname", Type: "string", Required: true},
		},
	}

	suite.mockCIType.On("UpdateCIType", mock.Anything, ciTypeID, &reqBody, userID).Return(updatedCIType, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/ci-types/"+ciTypeID.String(), bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	ciTypeHandlers.UpdateCIType(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response ci.CITypeDefinition
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), updatedCIType.Name, response.Name)

	suite.mockCIType.AssertExpectations(suite.T())
}

func (suite *CITypeHandlerSuite) TestDeleteCIType() {
	handler := &Handler{logger: suite.logger}
	ciTypeHandlers := NewCITypeHandlers(handler, suite.mockCIType)

	userID := uuid.New()
	ciTypeID := uuid.New()

	suite.mockCIType.On("DeleteCIType", mock.Anything, ciTypeID, userID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/ci-types/"+ciTypeID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	ciTypeHandlers.DeleteCIType(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
	assert.Equal(suite.T(), 0, w.Body.Len())

	suite.mockCIType.AssertExpectations(suite.T())
}

func (suite *CITypeHandlerSuite) TestDeleteCITypeInUse() {
	handler := &Handler{logger: suite.logger}
	ciTypeHandlers := NewCITypeHandlers(handler, suite.mockCIType)

	userID := uuid.New()
	ciTypeID := uuid.New()

	suite.mockCIType.On("DeleteCIType", mock.Anything, ciTypeID, userID).Return(assert.AnError)

	req := httptest.NewRequest(http.MethodDelete, "/ci-types/"+ciTypeID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	ciTypeHandlers.DeleteCIType(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	suite.mockCIType.AssertExpectations(suite.T())
}

func TestCITypeHandlerSuite(t *testing.T) {
	suite.Run(t, new(CITypeHandlerSuite))
}