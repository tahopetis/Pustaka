package ci

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/pustaka/pustaka/internal/testutils"
)

type CIServiceSuite struct {
	suite.Suite
	db        *pgxpool.Pool
	service   *Service
	cleanup   func()
}

func (suite *CIServiceSuite) SetupSuite() {
	suite.db, suite.cleanup = testutils.SetupTestDB(suite.T())

	// Initialize service with required dependencies
	repo := NewRepository(suite.db)
	neo4jService := &MockNeo4jService{} // We'll create a mock for testing
	redisService := &MockRedisService{}  // We'll create a mock for testing
	logger := &MockLogger{}              // We'll create a mock for testing

	suite.service = NewService(repo, neo4jService, redisService, logger)
}

func (suite *CIServiceSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *CIServiceSuite) SetupTest() {
	testutils.CleanupDB(suite.T(), suite.db)
	testutils.InsertTestData(suite.T(), suite.db)
}

func (suite *CIServiceSuite) TestCreateCI() {
	ctx := context.Background()
	userID := uuid.New()

	req := &CreateCIRequest{
		Name:   "test-server",
		CIType: "Server",
		Attributes: map[string]interface{}{
			"hostname":   "test-server-01",
			"ip_address": "192.168.1.100",
			"os":         "Ubuntu 20.04",
		},
		Tags: []string{"production", "web"},
	}

	ci, err := suite.service.CreateCI(ctx, req, userID)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), ci.ID)
	assert.Equal(suite.T(), req.Name, ci.Name)
	assert.Equal(suite.T(), req.CIType, ci.CIType)
	assert.Equal(suite.T(), req.Attributes, ci.Attributes)
	assert.Equal(suite.T(), req.Tags, ci.Tags)
	assert.Equal(suite.T(), userID, ci.CreatedBy)
	assert.NotZero(suite.T(), ci.CreatedAt)
}

func (suite *CIServiceSuite) TestCreateCIValidation() {
	ctx := context.Background()
	userID := uuid.New()

	// Test empty name
	req := &CreateCIRequest{
		Name:   "",
		CIType: "Server",
	}

	_, err := suite.service.CreateCI(ctx, req, userID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "Name is required")

	// Test empty CI type
	req.Name = "test-server"
	req.CIType = ""

	_, err = suite.service.CreateCI(ctx, req, userID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "CI type is required")
}

func (suite *CIServiceSuite) TestGetCI() {
	ctx := context.Background()
	userID := uuid.New()

	// Create a CI first
	createReq := &CreateCIRequest{
		Name:   "test-server",
		CIType: "Server",
		Attributes: map[string]interface{}{
			"hostname": "test-server-01",
		},
	}

	createdCI, err := suite.service.CreateCI(ctx, createReq, userID)
	require.NoError(suite.T(), err)

	// Get the CI
	retrievedCI, err := suite.service.GetCI(ctx, createdCI.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), createdCI.ID, retrievedCI.ID)
	assert.Equal(suite.T(), createdCI.Name, retrievedCI.Name)
	assert.Equal(suite.T(), createdCI.CIType, retrievedCI.CIType)
	assert.Equal(suite.T(), createdCI.Attributes, retrievedCI.Attributes)
}

func (suite *CIServiceSuite) TestGetCINotFound() {
	ctx := context.Background()

	_, err := suite.service.GetCI(ctx, uuid.New())
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not found")
}

func (suite *CIServiceSuite) TestListCIs() {
	ctx := context.Background()
	userID := uuid.New()

	// Create test CIs
	ci1, err := suite.service.CreateCI(ctx, &CreateCIRequest{
		Name:   "web-server",
		CIType: "Server",
		Attributes: map[string]interface{}{
			"hostname": "web-01",
		},
		Tags: []string{"web", "production"},
	}, userID)
	require.NoError(suite.T(), err)

	ci2, err := suite.service.CreateCI(ctx, &CreateCIRequest{
		Name:   "database",
		CIType: "Database",
		Attributes: map[string]interface{}{
			"engine": "postgresql",
		},
		Tags: []string{"database", "production"},
	}, userID)
	require.NoError(suite.T(), err)

	// Test listing all CIs
	response, err := suite.service.ListCIs(ctx, ListCIFilters{}, 1, 10)
	require.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(response.CIs), 2)
	assert.GreaterOrEqual(suite.T(), response.Total, 2)

	// Test filtering by CI type
	filters := ListCIFilters{CIType: "Server"}
	response, err = suite.service.ListCIs(ctx, filters, 1, 10)
	require.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(response.CIs), 1)
	for _, ci := range response.CIs {
		assert.Equal(suite.T(), "Server", ci.CIType)
	}

	// Test filtering by search
	filters = ListCIFilters{Search: "web"}
	response, err = suite.service.ListCIs(ctx, filters, 1, 10)
	require.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(response.CIs), 1)
	for _, ci := range response.CIs {
		assert.Contains(suite.T(), ci.Name, "web")
	}

	// Test filtering by tags
	filters = ListCIFilters{Tags: []string{"web"}}
	response, err = suite.service.ListCIs(ctx, filters, 1, 10)
	require.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(response.CIs), 1)
	for _, ci := range response.CIs {
		assert.Contains(suite.T(), ci.Tags, "web")
	}
}

func (suite *CIServiceSuite) TestUpdateCI() {
	ctx := context.Background()
	userID := uuid.New()

	// Create a CI first
	createReq := &CreateCIRequest{
		Name:   "test-server",
		CIType: "Server",
		Attributes: map[string]interface{}{
			"hostname": "test-server-01",
			"os":       "Ubuntu 18.04",
		},
		Tags: []string{"test"},
	}

	createdCI, err := suite.service.CreateCI(ctx, createReq, userID)
	require.NoError(suite.T(), err)

	// Update the CI
	updateReq := &UpdateCIRequest{
		Name: strPtr("updated-server"),
		Attributes: map[string]interface{}{
			"os":       "Ubuntu 20.04",
			"cpu_cores": 4,
		},
		Tags: []string{"production", "web"},
	}

	updatedCI, err := suite.service.UpdateCI(ctx, createdCI.ID, updateReq, userID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), *updateReq.Name, updatedCI.Name)
	assert.Equal(suite.T(), updateReq.Attributes, updatedCI.Attributes)
	assert.Equal(suite.T(), updateReq.Tags, updatedCI.Tags)
	assert.Equal(suite.T(), userID, updatedCI.UpdatedBy)
	assert.NotZero(suite.T(), updatedCI.UpdatedAt)
}

func (suite *CIServiceSuite) TestDeleteCI() {
	ctx := context.Background()
	userID := uuid.New()

	// Create a CI first
	createReq := &CreateCIRequest{
		Name:   "test-server",
		CIType: "Server",
		Attributes: map[string]interface{}{
			"hostname": "test-server-01",
		},
	}

	createdCI, err := suite.service.CreateCI(ctx, createReq, userID)
	require.NoError(suite.T(), err)

	// Delete the CI
	err = suite.service.DeleteCI(ctx, createdCI.ID, userID)
	require.NoError(suite.T(), err)

	// Verify CI is deleted
	_, err = suite.service.GetCI(ctx, createdCI.ID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not found")
}

func (suite *CIServiceSuite) TestDeleteCIWithRelationships() {
	ctx := context.Background()
	userID := uuid.New()

	// Create two CIs
	ci1, err := suite.service.CreateCI(ctx, &CreateCIRequest{
		Name:   "web-server",
		CIType: "Server",
		Attributes: map[string]interface{}{
			"hostname": "web-01",
		},
	}, userID)
	require.NoError(suite.T(), err)

	ci2, err := suite.service.CreateCI(ctx, &CreateCIRequest{
		Name:   "database",
		CIType: "Database",
		Attributes: map[string]interface{}{
			"engine": "postgresql",
		},
	}, userID)
	require.NoError(suite.T(), err)

	// Create a relationship between them
	relReq := &CreateRelationshipRequest{
		SourceID:         ci1.ID,
		TargetID:         ci2.ID,
		RelationshipType: "connects_to",
		Attributes: map[string]interface{}{
			"port": 5432,
		},
	}

	_, err = suite.service.CreateRelationship(ctx, relReq, userID)
	require.NoError(suite.T(), err)

	// Try to delete CI that has relationships
	err = suite.service.DeleteCI(ctx, ci1.ID, userID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "has relationships")
}

// Mock implementations for testing
type MockNeo4jService struct {
	// Implement required methods for testing
}

type MockRedisService struct {
	// Implement required methods for testing
}

type MockLogger struct {
	// Implement required methods for testing
}

// Helper function
func strPtr(s string) *string {
	return &s
}

func TestCIServiceSuite(t *testing.T) {
	suite.Run(t, new(CIServiceSuite))
}