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

// MockCIService from testify/mock for more granular testing
type MockCIServiceWithMock struct {
	mock.Mock
}

func (m *MockCIServiceWithMock) CreateCI(ctx context.Context, req *ci.CreateCIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ci.ConfigurationItem), args.Error(1)
}

func (m *MockCIServiceWithMock) GetCI(ctx context.Context, ciID uuid.UUID) (*ci.ConfigurationItem, error) {
	args := m.Called(ctx, ciID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ci.ConfigurationItem), args.Error(1)
}

func (m *MockCIServiceWithMock) ListCIs(ctx context.Context, filters ci.ListCIFilters, page, limit int) (*ci.CIListResponse, error) {
	args := m.Called(ctx, filters, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ci.CIListResponse), args.Error(1)
}

func (m *MockCIServiceWithMock) UpdateCI(ctx context.Context, ciID uuid.UUID, req *ci.UpdateCIRequest, userID uuid.UUID) (*ci.ConfigurationItem, error) {
	args := m.Called(ctx, ciID, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ci.ConfigurationItem), args.Error(1)
}

func (m *MockCIServiceWithMock) DeleteCI(ctx context.Context, ciID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, ciID, userID)
	return args.Error(0)
}

func (m *MockCIServiceWithMock) GetCIRelationships(ctx context.Context, ciID uuid.UUID) ([]ci.Relationship, error) {
	args := m.Called(ctx, ciID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ci.Relationship), args.Error(1)
}

func (m *MockCIServiceWithMock) GetGraphData(ctx context.Context, filters ci.GraphFilters) (*ci.GraphData, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ci.GraphData), args.Error(1)
}

// CIHandlerSuite contains tests for CI handlers
type CIHandlerSuite struct {
	suite.Suite
	mockCI *MockCIServiceWithMock
	logger *pustakaLogger.Logger
}

func (suite *CIHandlerSuite) SetupTest() {
	suite.mockCI = new(MockCIServiceWithMock)
	suite.logger = pustakaLogger.New(pustakaLogger.Config{Level: "info"})
}

func (suite *CIHandlerSuite) TestCreateCI() {
	handler := &Handler{logger: suite.logger}
	ciHandlers := NewCIHandlers(handler, suite.mockCI)

	userID := uuid.New()
	testCI := &ci.ConfigurationItem{
		ID:     uuid.New(),
		Name:   "test-server",
		CIType: "Server",
		Attributes: map[string]interface{}{
			"hostname": "test-server-01",
			"ip_address": "192.168.1.100",
		},
		Tags: []string{"production", "web"},
	}

	reqBody := ci.CreateCIRequest{
		Name:   "test-server",
		CIType: "Server",
		Attributes: map[string]interface{}{
			"hostname": "test-server-01",
			"ip_address": "192.168.1.100",
		},
		Tags: []string{"production", "web"},
	}

	suite.mockCI.On("CreateCI", mock.Anything, &reqBody, userID).Return(testCI, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/ci", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	ciHandlers.CreateCI(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response ci.ConfigurationItem
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), testCI.Name, response.Name)
	assert.Equal(suite.T(), testCI.CIType, response.CIType)

	suite.mockCI.AssertExpectations(suite.T())
}

func (suite *CIHandlerSuite) TestCreateCIValidationError() {
	handler := &Handler{logger: suite.logger}
	ciHandlers := NewCIHandlers(handler, suite.mockCI)

	userID := uuid.New()
	reqBody := ci.CreateCIRequest{
		Name:   "", // Empty name should cause validation error
		CIType: "Server",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/ci", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	ciHandlers.CreateCI(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Name is required", response["error"])
}

func (suite *CIHandlerSuite) TestGetCI() {
	handler := &Handler{logger: suite.logger}
	ciHandlers := NewCIHandlers(handler, suite.mockCI)

	ciID := uuid.New()
	testCI := &ci.ConfigurationItem{
		ID:     ciID,
		Name:   "test-server",
		CIType: "Server",
		Attributes: map[string]interface{}{
			"hostname": "test-server-01",
		},
	}

	suite.mockCI.On("GetCI", mock.Anything, ciID).Return(testCI, nil)

	req := httptest.NewRequest(http.MethodGet, "/ci/"+ciID.String(), nil)
	w := httptest.NewRecorder()

	ciHandlers.GetCI(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response ci.ConfigurationItem
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), testCI.ID, response.ID)
	assert.Equal(suite.T(), testCI.Name, response.Name)

	suite.mockCI.AssertExpectations(suite.T())
}

func (suite *CIHandlerSuite) TestGetCINotFound() {
	handler := &Handler{logger: suite.logger}
	ciHandlers := NewCIHandlers(handler, suite.mockCI)

	ciID := uuid.New()
	suite.mockCI.On("GetCI", mock.Anything, ciID).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/ci/"+ciID.String(), nil)
	w := httptest.NewRecorder()

	ciHandlers.GetCI(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Configuration item not found", response["error"])

	suite.mockCI.AssertExpectations(suite.T())
}

func (suite *CIHandlerSuite) TestListCIs() {
	handler := &Handler{logger: suite.logger}
	ciHandlers := NewCIHandlers(handler, suite.mockCI)

	testResponse := &ci.CIListResponse{
		CIs: []ci.ConfigurationItem{
			{
				ID:     uuid.New(),
				Name:   "server-1",
				CIType: "Server",
			},
			{
				ID:     uuid.New(),
				Name:   "app-1",
				CIType: "Application",
			},
		},
		Total:      2,
		Page:       1,
		Limit:      20,
		TotalPages: 1,
	}

	expectedFilters := ci.ListCIFilters{
		CIType: "Server",
		Search: "test",
		Sort:   "name",
		Order:  "asc",
	}

	suite.mockCI.On("ListCIs", mock.Anything, expectedFilters, 1, 20).Return(testResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/ci?ci_type=Server&search=test&sort=name&order=asc&page=1&limit=20", nil)
	w := httptest.NewRecorder()

	ciHandlers.ListCIs(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response ci.CIListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), testResponse.Total, response.Total)
	assert.Len(suite.T(), response.CIs, 2)

	suite.mockCI.AssertExpectations(suite.T())
}

func (suite *CIHandlerSuite) TestUpdateCI() {
	handler := &Handler{logger: suite.logger}
	ciHandlers := NewCIHandlers(handler, suite.mockCI)

	userID := uuid.New()
	ciID := uuid.New()

	updatedCI := &ci.ConfigurationItem{
		ID:     ciID,
		Name:   "updated-server",
		CIType: "Server",
		Attributes: map[string]interface{}{
			"hostname": "updated-server-01",
		},
	}

	reqBody := ci.UpdateCIRequest{
		Attributes: map[string]interface{}{
			"hostname": "updated-server-01",
		},
	}

	suite.mockCI.On("UpdateCI", mock.Anything, ciID, &reqBody, userID).Return(updatedCI, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/ci/"+ciID.String(), bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	ciHandlers.UpdateCI(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response ci.ConfigurationItem
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), updatedCI.Name, response.Name)

	suite.mockCI.AssertExpectations(suite.T())
}

func (suite *CIHandlerSuite) TestDeleteCI() {
	handler := &Handler{logger: suite.logger}
	ciHandlers := NewCIHandlers(handler, suite.mockCI)

	userID := uuid.New()
	ciID := uuid.New()

	suite.mockCI.On("DeleteCI", mock.Anything, ciID, userID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/ci/"+ciID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	ciHandlers.DeleteCI(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
	assert.Equal(suite.T(), 0, w.Body.Len())

	suite.mockCI.AssertExpectations(suite.T())
}

func (suite *CIHandlerSuite) TestDeleteCIWithRelationships() {
	handler := &Handler{logger: suite.logger}
	ciHandlers := NewCIHandlers(handler, suite.mockCI)

	userID := uuid.New()
	ciID := uuid.New()

	suite.mockCI.On("DeleteCI", mock.Anything, ciID, userID).Return(assert.AnError)

	req := httptest.NewRequest(http.MethodDelete, "/ci/"+ciID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	ciHandlers.DeleteCI(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	suite.mockCI.AssertExpectations(suite.T())
}

func TestCIHandlerSuite(t *testing.T) {
	suite.Run(t, new(CIHandlerSuite))
}