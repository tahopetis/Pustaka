package ci

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// MockRelationshipTypeRepository is a mock for the RelationshipTypeRepository interface
type MockRelationshipTypeRepository struct {
	mock.Mock
}

func (m *MockRelationshipTypeRepository) Create(ctx context.Context, rt *RelationshipType) (*RelationshipType, error) {
	args := m.Called(ctx, rt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RelationshipType), args.Error(1)
}

func (m *MockRelationshipTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*RelationshipType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RelationshipType), args.Error(1)
}

func (m *MockRelationshipTypeRepository) GetByName(ctx context.Context, name string) (*RelationshipType, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RelationshipType), args.Error(1)
}

func (m *MockRelationshipTypeRepository) List(ctx context.Context, filters ListRelationshipTypeFilters, page, limit int) (*RelationshipTypeListResponse, error) {
	args := m.Called(ctx, filters, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RelationshipTypeListResponse), args.Error(1)
}

func (m *MockRelationshipTypeRepository) GetActive(ctx context.Context) ([]RelationshipType, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]RelationshipType), args.Error(1)
}

func (m *MockRelationshipTypeRepository) Update(ctx context.Context, rt *RelationshipType) (*RelationshipType, error) {
	args := m.Called(ctx, rt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RelationshipType), args.Error(1)
}

func (m *MockRelationshipTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRelationshipTypeRepository) GetUsageCount(ctx context.Context, id uuid.UUID) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRelationshipTypeRepository) GetUsage(ctx context.Context) ([]RelationshipTypeUsage, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]RelationshipTypeUsage), args.Error(1)
}

func (m *MockRelationshipTypeRepository) GetStatistics(ctx context.Context) (*RelationshipTypeStatistics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RelationshipTypeStatistics), args.Error(1)
}

// RelationshipTypeServiceTestSuite tests the relationship type service
type RelationshipTypeServiceTestSuite struct {
	suite.Suite
	service  *RelationshipTypeService
	repo     *MockRelationshipTypeRepository
	logger   *pustakaLogger.Logger
	ctx      context.Context
	userID   uuid.UUID
}

func (suite *RelationshipTypeServiceTestSuite) SetupTest() {
	suite.repo = new(MockRelationshipTypeRepository)
	suite.logger, _ = pustakaLogger.NewLogger("debug")
	suite.service = NewRelationshipTypeService(suite.repo, suite.logger)
	suite.ctx = context.Background()
	suite.userID = uuid.New()
}

func (suite *RelationshipTypeServiceTestSuite) TestCreateRelationshipType_Success() {
	// Arrange
	req := &CreateRelationshipTypeRequest{
		Name:               "test_depends",
		DisplayName:        stringPtr("Test Depends"),
		Description:        stringPtr("Test dependency relationship"),
		ForwardLabel:       "depends on",
		ReverseLabel:       "required by",
		Color:              stringPtr("#ff0000"),
		Icon:               stringPtr("test-icon"),
		Category:           stringPtr("Test"),
		AllowedSourceTypes: []string{"Application", "Service"},
		AllowedTargetTypes: []string{"Server", "Database"},
		CardinalitySource:  "many",
		CardinalityTarget:  "many",
		Bidirectional:      false,
		Attributes:         map[string]interface{}{"test": "value"},
	}

	expectedRT := &RelationshipType{
		ID:                 uuid.New(),
		Name:               req.Name,
		DisplayName:        req.DisplayName,
		Description:        req.Description,
		ForwardLabel:       req.ForwardLabel,
		ReverseLabel:       req.ReverseLabel,
		Color:              req.Color,
		Icon:               req.Icon,
		Category:           req.Category,
		AllowedSourceTypes: req.AllowedSourceTypes,
		AllowedTargetTypes: req.AllowedTargetTypes,
		CardinalitySource:  req.CardinalitySource,
		CardinalityTarget:  req.CardinalityTarget,
		Bidirectional:      req.Bidirectional,
		Attributes:         req.Attributes,
		IsActive:           true,
		IsSystem:           false,
		CreatedBy:          suite.userID,
		CreatedAt:          time.Now(),
	}

	// Mock calls
	suite.repo.On("GetByName", suite.ctx, req.Name).Return(nil, assert.AnError)
	suite.repo.On("Create", suite.ctx, mock.AnythingOfType("*ci.RelationshipType")).Return(expectedRT, nil)

	// Act
	result, err := suite.service.CreateRelationshipType(suite.ctx, req, suite.userID)

	// Assert
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedRT.ID, result.ID)
	assert.Equal(suite.T(), req.Name, result.Name)
	assert.Equal(suite.T(), req.ForwardLabel, result.ForwardLabel)
	assert.Equal(suite.T(), req.ReverseLabel, result.ReverseLabel)
	assert.Equal(suite.T(), suite.userID, result.CreatedBy)
	assert.True(suite.T(), result.IsActive)
	assert.False(suite.T(), result.IsSystem)

	suite.repo.AssertExpectations(suite.T())
}

func (suite *RelationshipTypeServiceTestSuite) TestCreateRelationshipType_DuplicateName() {
	// Arrange
	req := &CreateRelationshipTypeRequest{
		Name:          "existing_type",
		ForwardLabel:  "tests",
		ReverseLabel:  "tested by",
		CardinalitySource: "one",
		CardinalityTarget: "many",
	}

	existingRT := &RelationshipType{
		ID:           uuid.New(),
		Name:         req.Name,
		ForwardLabel: "old label",
		ReverseLabel: "old reverse",
	}

	// Mock call
	suite.repo.On("GetByName", suite.ctx, req.Name).Return(existingRT, nil)

	// Act
	result, err := suite.service.CreateRelationshipType(suite.ctx, req, suite.userID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "already exists")

	suite.repo.AssertExpectations(suite.T())
}

func (suite *RelationshipTypeServiceTestSuite) TestGetRelationshipType_Success() {
	// Arrange
	expectedID := uuid.New()
	expectedRT := &RelationshipType{
		ID:           expectedID,
		Name:         "test_type",
		ForwardLabel: "tests",
		ReverseLabel: "tested by",
		IsActive:     true,
	}

	suite.repo.On("GetByID", suite.ctx, expectedID).Return(expectedRT, nil)

	// Act
	result, err := suite.service.GetRelationshipType(suite.ctx, expectedID)

	// Assert
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedRT.ID, result.ID)
	assert.Equal(suite.T(), expectedRT.Name, result.Name)

	suite.repo.AssertExpectations(suite.T())
}

func (suite *RelationshipTypeServiceTestSuite) TestGetRelationshipType_NotFound() {
	// Arrange
	id := uuid.New()

	suite.repo.On("GetByID", suite.ctx, id).Return(nil, assert.AnError)

	// Act
	result, err := suite.service.GetRelationshipType(suite.ctx, id)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "not found")

	suite.repo.AssertExpectations(suite.T())
}

func (suite *RelationshipTypeServiceTestSuite) TestListRelationshipTypes_Success() {
	// Arrange
	filters := ListRelationshipTypeFilters{
		Search:   "test",
		Category: "Test",
		Sort:     "name",
		Order:    "asc",
		IsActive: boolPtr(true),
	}

	expectedResponse := &RelationshipTypeListResponse{
		RelationshipTypes: []RelationshipType{
			{
				ID:           uuid.New(),
				Name:         "test_type_1",
				ForwardLabel: "tests",
				ReverseLabel: "tested by",
				IsActive:     true,
			},
			{
				ID:           uuid.New(),
				Name:         "test_type_2",
				ForwardLabel: "validates",
				ReverseLabel: "validated by",
				IsActive:     true,
			},
		},
		Page:       1,
		Limit:      10,
		Total:      2,
		TotalPages: 1,
	}

	suite.repo.On("List", suite.ctx, filters, 1, 10).Return(expectedResponse, nil)

	// Act
	result, err := suite.service.ListRelationshipTypes(suite.ctx, filters, 1, 10)

	// Assert
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedResponse.Page, result.Page)
	assert.Equal(suite.T(), expectedResponse.Total, result.Total)
	assert.Len(suite.T(), result.RelationshipTypes, 2)

	suite.repo.AssertExpectations(suite.T())
}

func (suite *RelationshipTypeServiceTestSuite) TestGetActiveRelationshipTypes_Success() {
	// Arrange
	expectedTypes := []RelationshipType{
		{
			ID:           uuid.New(),
			Name:         "depends_on",
			ForwardLabel: "depends on",
			ReverseLabel: "required by",
			IsActive:     true,
			IsSystem:     true,
		},
		{
			ID:           uuid.New(),
			Name:         "hosts",
			ForwardLabel: "hosts",
			ReverseLabel: "hosted on",
			IsActive:     true,
			IsSystem:     true,
		},
	}

	suite.repo.On("GetActive", suite.ctx).Return(expectedTypes, nil)

	// Act
	result, err := suite.service.GetActiveRelationshipTypes(suite.ctx)

	// Assert
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.True(suite.T(), result[0].IsActive)
	assert.True(suite.T(), result[1].IsActive)

	suite.repo.AssertExpectations(suite.T())
}

func (suite *RelationshipTypeServiceTestSuite) TestUpdateRelationshipType_Success() {
	// Arrange
	id := uuid.New()
	req := &UpdateRelationshipTypeRequest{
		DisplayName: stringPtr("Updated Display Name"),
		Description: stringPtr("Updated description"),
		Category:    stringPtr("Updated Category"),
		IsActive:    boolPtr(false),
	}

	existingRT := &RelationshipType{
		ID:           id,
		Name:         "test_type",
		DisplayName:  stringPtr("Old Display Name"),
		ForwardLabel: "tests",
		ReverseLabel: "tested by",
		IsActive:     true,
		IsSystem:     false,
		CreatedBy:    suite.userID,
	}

	updatedRT := &RelationshipType{
		ID:           id,
		Name:         "test_type",
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		ForwardLabel: "tests",
		ReverseLabel: "tested by",
		Category:     req.Category,
		IsActive:     *req.IsActive,
		IsSystem:     false,
		CreatedBy:    suite.userID,
		UpdatedBy:    &suite.userID,
		UpdatedAt:    timePtr(time.Now()),
	}

	suite.repo.On("GetByID", suite.ctx, id).Return(existingRT, nil)
	suite.repo.On("Update", suite.ctx, mock.AnythingOfType("*ci.RelationshipType")).Return(updatedRT, nil)

	// Act
	result, err := suite.service.UpdateRelationshipType(suite.ctx, id, req, suite.userID)

	// Assert
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), updatedRT.ID, result.ID)
	assert.Equal(suite.T(), *req.DisplayName, *result.DisplayName)
	assert.Equal(suite.T(), *req.Description, *result.Description)
	assert.Equal(suite.T(), *req.Category, *result.Category)
	assert.Equal(suite.T(), *req.IsActive, result.IsActive)

	suite.repo.AssertExpectations(suite.T())
}

func (suite *RelationshipTypeServiceTestSuite) TestUpdateRelationshipType_SystemType() {
	// Arrange
	id := uuid.New()
	req := &UpdateRelationshipTypeRequest{
		DisplayName: stringPtr("Should not update"),
	}

	existingRT := &RelationshipType{
		ID:       id,
		Name:     "system_type",
		IsSystem: true,
	}

	suite.repo.On("GetByID", suite.ctx, id).Return(existingRT, nil)

	// Act
	result, err := suite.service.UpdateRelationshipType(suite.ctx, id, req, suite.userID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "system relationship type")

	suite.repo.AssertExpectations(suite.T())
}

func (suite *RelationshipTypeServiceTestSuite) TestDeleteRelationshipType_Success() {
	// Arrange
	id := uuid.New()
	existingRT := &RelationshipType{
		ID:       id,
		Name:     "custom_type",
		IsSystem: false,
	}

	suite.repo.On("GetByID", suite.ctx, id).Return(existingRT, nil)
	suite.repo.On("GetUsageCount", suite.ctx, id).Return(int64(0), nil)
	suite.repo.On("Delete", suite.ctx, id).Return(nil)

	// Act
	err := suite.service.DeleteRelationshipType(suite.ctx, id, suite.userID)

	// Assert
	require.NoError(suite.T(), err)

	suite.repo.AssertExpectations(suite.T())
}

func (suite *RelationshipTypeServiceTestSuite) TestDeleteRelationshipType_SystemType() {
	// Arrange
	id := uuid.New()
	existingRT := &RelationshipType{
		ID:       id,
		Name:     "system_type",
		IsSystem: true,
	}

	suite.repo.On("GetByID", suite.ctx, id).Return(existingRT, nil)

	// Act
	err := suite.service.DeleteRelationshipType(suite.ctx, id, suite.userID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "system relationship type")

	suite.repo.AssertExpectations(suite.T())
}

func (suite *RelationshipTypeServiceTestSuite) TestDeleteRelationshipType_InUse() {
	// Arrange
	id := uuid.New()
	existingRT := &RelationshipType{
		ID:       id,
		Name:     "used_type",
		IsSystem: false,
	}

	suite.repo.On("GetByID", suite.ctx, id).Return(existingRT, nil)
	suite.repo.On("GetUsageCount", suite.ctx, id).Return(int64(5), nil)

	// Act
	err := suite.service.DeleteRelationshipType(suite.ctx, id, suite.userID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "is being used by")

	suite.repo.AssertExpectations(suite.T())
}

func (suite *RelationshipTypeServiceTestSuite) TestValidateRelationship_Valid() {
	// Arrange
	req := &RelationshipCompatibilityRequest{
		RelationshipType: "depends_on",
		SourceType:       "Application",
		TargetType:       "Server",
	}

	relationshipType := &RelationshipType{
		ID:                 uuid.New(),
		Name:               "depends_on",
		ForwardLabel:       "depends on",
		ReverseLabel:       "required by",
		AllowedSourceTypes: []string{"Application", "Service"},
		AllowedTargetTypes: []string{"Server", "Database"},
		CardinalitySource:  "many",
		CardinalityTarget:  "many",
		IsActive:           true,
	}

	sourceCI := &ConfigurationItem{
		ID:     uuid.New(),
		Name:   "Test App",
		CIType: "Application",
	}

	targetCI := &ConfigurationItem{
		ID:     uuid.New(),
		Name:   "Test Server",
		CIType: "Server",
	}

	suite.repo.On("GetByName", suite.ctx, req.RelationshipType).Return(relationshipType, nil)

	// Act
	result, err := suite.service.ValidateRelationship(suite.ctx, req)

	// Assert
	require.NoError(suite.T(), err)
	assert.True(suite.T(), result.IsValid)
	assert.Empty(suite.T(), result.Errors)

	suite.repo.AssertExpectations(suite.T())
}

func (suite *RelationshipTypeServiceTestSuite) TestValidateRelationship_InactiveType() {
	// Arrange
	req := &RelationshipCompatibilityRequest{
		RelationshipType: "inactive_type",
		SourceType:       "Application",
		TargetType:       "Server",
	}

	relationshipType := &RelationshipType{
		ID:       uuid.New(),
		Name:     "inactive_type",
		IsActive: false,
	}

	suite.repo.On("GetByName", suite.ctx, req.RelationshipType).Return(relationshipType, nil)

	// Act
	result, err := suite.service.ValidateRelationship(suite.ctx, req)

	// Assert
	require.NoError(suite.T(), err)
	assert.False(suite.T(), result.IsValid)
	assert.Contains(suite.T(), result.Errors, "relationship type is not active")

	suite.repo.AssertExpectations(suite.T())
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func TestRelationshipTypeServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RelationshipTypeServiceTestSuite))
}