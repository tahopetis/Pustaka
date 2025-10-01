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

// MockRelationshipService is a mock implementation of the relationship service interface
type MockRelationshipService struct {
	mock.Mock
}

func (m *MockRelationshipService) CreateRelationship(ctx context.Context, req *ci.CreateRelationshipRequest, userID uuid.UUID) (*ci.Relationship, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ci.Relationship), args.Error(1)
}

func (m *MockRelationshipService) GetRelationship(ctx context.Context, relationshipID uuid.UUID) (*ci.Relationship, error) {
	args := m.Called(ctx, relationshipID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ci.Relationship), args.Error(1)
}

func (m *MockRelationshipService) GetCIRelationships(ctx context.Context, ciID uuid.UUID) ([]ci.Relationship, error) {
	args := m.Called(ctx, ciID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ci.Relationship), args.Error(1)
}

func (m *MockRelationshipService) DeleteRelationship(ctx context.Context, relationshipID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, relationshipID, userID)
	return args.Error(0)
}

func (m *MockRelationshipService) GetGraphData(ctx context.Context, filters ci.GraphFilters) (*ci.GraphData, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ci.GraphData), args.Error(1)
}

// RelationshipHandlerSuite contains tests for relationship handlers
type RelationshipHandlerSuite struct {
	suite.Suite
	mockRelationship *MockRelationshipService
	logger           *pustakaLogger.Logger
}

func (suite *RelationshipHandlerSuite) SetupTest() {
	suite.mockRelationship = new(MockRelationshipService)
	suite.logger = pustakaLogger.New(pustakaLogger.Config{Level: "info"})
}

func (suite *RelationshipHandlerSuite) TestCreateRelationship() {
	handler := &Handler{logger: suite.logger}
	relationshipHandlers := NewRelationshipHandlers(handler, suite.mockRelationship)

	userID := uuid.New()
	sourceID := uuid.New()
	targetID := uuid.New()

	testRelationship := &ci.Relationship{
		ID:                uuid.New(),
		SourceID:          sourceID,
		TargetID:          targetID,
		RelationshipType:  "depends_on",
		Attributes:        map[string]interface{}{"description": "Application depends on database"},
	}

	reqBody := ci.CreateRelationshipRequest{
		SourceID:         sourceID,
		TargetID:         targetID,
		RelationshipType: "depends_on",
		Attributes:       map[string]interface{}{"description": "Application depends on database"},
	}

	suite.mockRelationship.On("CreateRelationship", mock.Anything, &reqBody, userID).Return(testRelationship, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/relationships", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	relationshipHandlers.CreateRelationship(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response ci.Relationship
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), testRelationship.SourceID, response.SourceID)
	assert.Equal(suite.T(), testRelationship.TargetID, response.TargetID)
	assert.Equal(suite.T(), testRelationship.RelationshipType, response.RelationshipType)

	suite.mockRelationship.AssertExpectations(suite.T())
}

func (suite *RelationshipHandlerSuite) TestCreateRelationshipValidationError() {
	handler := &Handler{logger: suite.logger}
	relationshipHandlers := NewRelationshipHandlers(handler, suite.mockRelationship)

	userID := uuid.New()
	reqBody := ci.CreateRelationshipRequest{
		SourceID:         uuid.New(),
		TargetID:         uuid.New(), // Same as source - should cause validation error
		RelationshipType: "", // Empty type should cause validation error
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/relationships", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	relationshipHandlers.CreateRelationship(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Source and target cannot be the same", response["error"])
}

func (suite *RelationshipHandlerSuite) TestGetRelationship() {
	handler := &Handler{logger: suite.logger}
	relationshipHandlers := NewRelationshipHandlers(handler, suite.mockRelationship)

	relationshipID := uuid.New()
	testRelationship := &ci.Relationship{
		ID:               relationshipID,
		SourceID:         uuid.New(),
		TargetID:         uuid.New(),
		RelationshipType: "depends_on",
		Attributes:       map[string]interface{}{"description": "Test relationship"},
	}

	suite.mockRelationship.On("GetRelationship", mock.Anything, relationshipID).Return(testRelationship, nil)

	req := httptest.NewRequest(http.MethodGet, "/relationships/"+relationshipID.String(), nil)
	w := httptest.NewRecorder()

	relationshipHandlers.GetRelationship(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response ci.Relationship
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), testRelationship.ID, response.ID)
	assert.Equal(suite.T(), testRelationship.RelationshipType, response.RelationshipType)

	suite.mockRelationship.AssertExpectations(suite.T())
}

func (suite *RelationshipHandlerSuite) TestGetRelationshipNotFound() {
	handler := &Handler{logger: suite.logger}
	relationshipHandlers := NewRelationshipHandlers(handler, suite.mockRelationship)

	relationshipID := uuid.New()
	suite.mockRelationship.On("GetRelationship", mock.Anything, relationshipID).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/relationships/"+relationshipID.String(), nil)
	w := httptest.NewRecorder()

	relationshipHandlers.GetRelationship(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Relationship not found", response["error"])

	suite.mockRelationship.AssertExpectations(suite.T())
}

func (suite *RelationshipHandlerSuite) TestGetCIRelationships() {
	handler := &Handler{logger: suite.logger}
	relationshipHandlers := NewRelationshipHandlers(handler, suite.mockRelationship)

	ciID := uuid.New()
	testRelationships := []ci.Relationship{
		{
			ID:               uuid.New(),
			SourceID:         ciID,
			TargetID:         uuid.New(),
			RelationshipType: "depends_on",
		},
		{
			ID:               uuid.New(),
			SourceID:         uuid.New(),
			TargetID:         ciID,
			RelationshipType: "hosts",
		},
	}

	suite.mockRelationship.On("GetCIRelationships", mock.Anything, ciID).Return(testRelationships, nil)

	req := httptest.NewRequest(http.MethodGet, "/ci/"+ciID.String()+"/relationships", nil)
	w := httptest.NewRecorder()

	relationshipHandlers.GetCIRelationships(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response []ci.Relationship
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), response, 2)

	suite.mockRelationship.AssertExpectations(suite.T())
}

func (suite *RelationshipHandlerSuite) TestDeleteRelationship() {
	handler := &Handler{logger: suite.logger}
	relationshipHandlers := NewRelationshipHandlers(handler, suite.mockRelationship)

	userID := uuid.New()
	relationshipID := uuid.New()

	suite.mockRelationship.On("DeleteRelationship", mock.Anything, relationshipID, userID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/relationships/"+relationshipID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	relationshipHandlers.DeleteRelationship(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
	assert.Equal(suite.T(), 0, w.Body.Len())

	suite.mockRelationship.AssertExpectations(suite.T())
}

func (suite *RelationshipHandlerSuite) TestGetGraphData() {
	handler := &Handler{logger: suite.logger}
	relationshipHandlers := NewRelationshipHandlers(handler, suite.mockRelationship)

	testGraphData := &ci.GraphData{
		Nodes: []ci.GraphNode{
			{
				ID:       uuid.New().String(),
				Label:    "Web Server",
				Type:     "Server",
				Group:    "infrastructure",
				Properties: map[string]interface{}{
					"hostname": "web-01",
					"status":   "active",
				},
			},
			{
				ID:       uuid.New().String(),
				Label:    "Database",
				Type:     "Database",
				Group:    "infrastructure",
				Properties: map[string]interface{}{
					"engine": "PostgreSQL",
					"version": "14",
				},
			},
		},
		Edges: []ci.GraphEdge{
			{
				Source: "node-1",
				Target: "node-2",
				Type:   "dependency",
				Label:  "connects to",
				Properties: map[string]interface{}{
					"protocol": "tcp",
					"port":     5432,
				},
			},
		},
	}

	expectedFilters := ci.GraphFilters{
		CITypes: []string{"Server"},
		Search:  "depends_on",
		Limit:   2,
	}

	suite.mockRelationship.On("GetGraphData", mock.Anything, expectedFilters).Return(testGraphData, nil)

	req := httptest.NewRequest(http.MethodGet, "/graph?ci_type=Server&relationship=depends_on&depth=2", nil)
	w := httptest.NewRecorder()

	relationshipHandlers.GetGraphData(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response ci.GraphData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), response.Nodes, 2)
	assert.Len(suite.T(), response.Links, 1)

	suite.mockRelationship.AssertExpectations(suite.T())
}

func (suite *RelationshipHandlerSuite) TestGetGraphDataEmpty() {
	handler := &Handler{logger: suite.logger}
	relationshipHandlers := NewRelationshipHandlers(handler, suite.mockRelationship)

	testGraphData := &ci.GraphData{
		Nodes: []ci.GraphNode{},
		Edges: []ci.GraphEdge{},
	}

	suite.mockRelationship.On("GetGraphData", mock.Anything, ci.GraphFilters{}).Return(testGraphData, nil)

	req := httptest.NewRequest(http.MethodGet, "/graph", nil)
	w := httptest.NewRecorder()

	relationshipHandlers.GetGraphData(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response ci.GraphData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), response.Nodes, 0)
	assert.Len(suite.T(), response.Links, 0)

	suite.mockRelationship.AssertExpectations(suite.T())
}

func TestRelationshipHandlerSuite(t *testing.T) {
	suite.Run(t, new(RelationshipHandlerSuite))
}