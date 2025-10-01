package auth

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

type RBACSuite struct {
	suite.Suite
	db   *pgxpool.Pool
	rbac *RBACService
	cleanup func()
}

func (suite *RBACSuite) SetupSuite() {
	suite.db, suite.cleanup = testutils.SetupTestDB(suite.T())
	suite.rbac = NewRBACService(suite.db)
}

func (suite *RBACSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *RBACSuite) SetupTest() {
	// Clean up database before each test
	testutils.CleanupDB(suite.T(), suite.db)

	// Insert test data
	testutils.InsertTestData(suite.T(), suite.db)
}

func (suite *RBACSuite) TestCreateUser() {
	ctx := context.Background()

	// Test creating a user
	user, err := suite.rbac.CreateUser(ctx, "testuser", "test@example.com", "hashedpassword", []string{"viewer"})
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), user.ID)
	assert.Equal(suite.T(), "testuser", user.Username)
	assert.Equal(suite.T(), "test@example.com", user.Email)
	assert.True(suite.T(), user.IsActive)
	assert.Len(suite.T(), user.Roles, 1)
	assert.Equal(suite.T(), "viewer", user.Roles[0].Name)
}

func (suite *RBACSuite) TestGetUserByID() {
	ctx := context.Background()

	// Create a user first
	user, err := suite.rbac.CreateUser(ctx, "testuser", "test@example.com", "hashedpassword", []string{"viewer"})
	require.NoError(suite.T(), err)

	// Get user by ID
	retrievedUser, err := suite.rbac.GetUserByID(ctx, user.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID, retrievedUser.ID)
	assert.Equal(suite.T(), user.Username, retrievedUser.Username)
	assert.Equal(suite.T(), user.Email, retrievedUser.Email)
}

func (suite *RBACSuite) TestGetUserByUsername() {
	ctx := context.Background()

	// Create a user first
	user, err := suite.rbac.CreateUser(ctx, "testuser", "test@example.com", "hashedpassword", []string{"viewer"})
	require.NoError(suite.T(), err)

	// Get user by username
	retrievedUser, err := suite.rbac.GetUserByUsername(ctx, user.Username)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID, retrievedUser.ID)
	assert.Equal(suite.T(), user.Username, retrievedUser.Username)
	assert.Equal(suite.T(), user.Email, retrievedUser.Email)
}

func (suite *RBACSuite) TestGetUserNotFound() {
	ctx := context.Background()

	// Test getting non-existent user by ID
	_, err := suite.rbac.GetUserByID(ctx, uuid.New())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "user not found", err.Error())

	// Test getting non-existent user by username
	_, err = suite.rbac.GetUserByUsername(ctx, "nonexistent")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "user not found", err.Error())
}

func (suite *RBACSuite) TestCreateRole() {
	ctx := context.Background()

	// Test creating a role
	role, err := suite.rbac.CreateRole(ctx, "testrole", "Test role description", []string{"ci:read", "ci:create"})
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), role.ID)
	assert.Equal(suite.T(), "testrole", role.Name)
	assert.Equal(suite.T(), "Test role description", role.Description)
}

func (suite *RBACSuite) TestAssignRoleToUser() {
	ctx := context.Background()

	// Create user and role
	user, err := suite.rbac.CreateUser(ctx, "testuser", "test@example.com", "hashedpassword", []string{})
	require.NoError(suite.T(), err)

	role, err := suite.rbac.CreateRole(ctx, "testrole", "Test role", []string{"ci:read"})
	require.NoError(suite.T(), err)

	// Assign role to user
	err = suite.rbac.AssignRoleToUser(ctx, user.ID, role.ID, user.ID)
	require.NoError(suite.T(), err)

	// Verify user has the role
	retrievedUser, err := suite.rbac.GetUserByID(ctx, user.ID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrievedUser.Roles, 1)
	assert.Equal(suite.T(), role.Name, retrievedUser.Roles[0].Name)
}

func (suite *RBACSuite) TestRemoveRoleFromUser() {
	ctx := context.Background()

	// Create user with role
	user, err := suite.rbac.CreateUser(ctx, "testuser", "test@example.com", "hashedpassword", []string{"viewer"})
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), user.Roles, 1)

	// Get the role ID
	roleID := user.Roles[0].ID

	// Remove role from user
	err = suite.rbac.RemoveRoleFromUser(ctx, user.ID, roleID)
	require.NoError(suite.T(), err)

	// Verify user no longer has the role
	retrievedUser, err := suite.rbac.GetUserByID(ctx, user.ID)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), retrievedUser.Roles, 0)
}

func (suite *RBACSuite) TestGetUserPasswordHash() {
	ctx := context.Background()

	// Create a user
	user, err := suite.rbac.CreateUser(ctx, "testuser", "test@example.com", "hashedpassword123", []string{"viewer"})
	require.NoError(suite.T(), err)

	// Get password hash
	hash, err := suite.rbac.GetUserPasswordHash(ctx, user.Username)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "hashedpassword123", hash)
}

func (suite *RBACSuite) TestVerifyPassword() {
	ctx := context.Background()

	// Create a user with known password hash
	user, err := suite.rbac.CreateUser(ctx, "testuser", "test@example.com", "hashedpassword123", []string{"viewer"})
	require.NoError(suite.T(), err)

	// Test password verification (basic implementation)
	// In real implementation, this would use bcrypt
	valid := suite.rbac.VerifyPassword(ctx, user.Username, "hashedpassword123")
	assert.True(suite.T(), valid)

	// Test invalid password
	invalid := suite.rbac.VerifyPassword(ctx, user.Username, "wrongpassword")
	assert.False(suite.T(), invalid)
}

func TestRBACSuite(t *testing.T) {
	suite.Run(t, new(RBACSuite))
}

// Unit tests for User methods
func TestUserHasPermission(t *testing.T) {
	user := &User{
		Permissions: []string{"ci:read", "ci:create", "user:read"},
	}

	assert.True(t, user.HasPermission("ci:read"))
	assert.True(t, user.HasPermission("ci:create"))
	assert.False(t, user.HasPermission("ci:delete"))
}

func TestUserHasRole(t *testing.T) {
	user := &User{
		Roles: []Role{
			{Name: "admin"},
			{Name: "editor"},
		},
	}

	assert.True(t, user.HasRole("admin"))
	assert.True(t, user.HasRole("editor"))
	assert.False(t, user.HasRole("viewer"))
}

func TestUserHasAnyPermission(t *testing.T) {
	user := &User{
		Permissions: []string{"ci:read", "ci:create"},
	}

	assert.True(t, user.HasAnyPermission("ci:read", "ci:delete"))
	assert.True(t, user.HasAnyPermission("ci:create", "ci:update"))
	assert.False(t, user.HasAnyPermission("ci:delete", "ci:update"))
}

func TestUserHasAllPermissions(t *testing.T) {
	user := &User{
		Permissions: []string{"ci:read", "ci:create", "ci:update"},
	}

	assert.True(t, user.HasAllPermissions("ci:read", "ci:create"))
	assert.True(t, user.HasAllPermissions("ci:read", "ci:create", "ci:update"))
	assert.False(t, user.HasAllPermissions("ci:read", "ci:delete"))
}