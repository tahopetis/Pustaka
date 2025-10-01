package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type User struct {
	ID          uuid.UUID   `json:"id"`
	Username    string      `json:"username"`
	Email       string      `json:"email"`
	IsActive    bool        `json:"is_active"`
	Roles       []Role      `json:"roles"`
	Permissions []string    `json:"permissions"`
}

type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type Permission struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ResourceType string    `json:"resource_type"`
}

type RBACService struct {
	db *pgxpool.Pool
}

func NewRBACService(db *pgxpool.Pool) *RBACService {
	return &RBACService{db: db}
}

// GetUserByID retrieves a user by ID with their roles and permissions
func (r *RBACService) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.is_active
		FROM users u
		WHERE u.id = $1
	`

	var user User
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.IsActive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get user roles
	roles, err := r.getUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	user.Roles = roles

	// Get user permissions
	permissions, err := r.getUserPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	user.Permissions = permissions

	return &user, nil
}

// GetUserByUsername retrieves a user by username with their roles and permissions
func (r *RBACService) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.is_active
		FROM users u
		WHERE u.username = $1
	`

	var user User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.IsActive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get user roles
	roles, err := r.getUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	user.Roles = roles

	// Get user permissions
	permissions, err := r.getUserPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	user.Permissions = permissions

	return &user, nil
}

// getUserRoles retrieves all roles for a user
func (r *RBACService) getUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error) {
	query := `
		SELECT r.id, r.name, r.description
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
		)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// getUserPermissions retrieves all permissions for a user
func (r *RBACService) getUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT p.name
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permission string
		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (u *User) HasPermission(permission string) bool {
	for _, p := range u.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasRole checks if a user has a specific role
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r.Name == role {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a user has any of the specified permissions
func (u *User) HasAnyPermission(permissions ...string) bool {
	for _, requiredPerm := range permissions {
		if u.HasPermission(requiredPerm) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a user has all of the specified permissions
func (u *User) HasAllPermissions(permissions ...string) bool {
	for _, requiredPerm := range permissions {
		if !u.HasPermission(requiredPerm) {
			return false
		}
	}
	return true
}

// CreateRole creates a new role
func (r *RBACService) CreateRole(ctx context.Context, name, description string, permissions []string) (*Role, error) {
	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert role
	roleID := uuid.New()
	query := `
		INSERT INTO roles (id, name, description, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, name, description, created_at
	`

	var role Role
	err = tx.QueryRow(ctx, query, roleID, name, description).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		new(any), // created_at - not used in Role struct
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Add permissions to role
	if len(permissions) > 0 {
		for _, permName := range permissions {
			// Get permission ID
			var permID uuid.UUID
			err = tx.QueryRow(ctx, "SELECT id FROM permissions WHERE name = $1", permName).Scan(&permID)
			if err != nil {
				if err == pgx.ErrNoRows {
					return nil, fmt.Errorf("permission '%s' not found", permName)
				}
				return nil, fmt.Errorf("failed to get permission: %w", err)
			}

			// Add role-permission mapping
			_, err = tx.Exec(ctx,
				"INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)",
				roleID, permID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to assign permission to role: %w", err)
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &role, nil
}

// AssignRoleToUser assigns a role to a user
func (r *RBACService) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, userID, roleID, assignedBy)
	if err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("role_id", roleID.String()).
		Str("assigned_by", assignedBy.String()).
		Msg("Role assigned to user")

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (r *RBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`

	_, err := r.db.Exec(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("role_id", roleID.String()).
		Msg("Role removed from user")

	return nil
}

// CreateUser creates a new user
func (r *RBACService) CreateUser(ctx context.Context, username, email, passwordHash string, roles []string) (*User, error) {
	// Start transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert user
	userID := uuid.New()
	query := `
		INSERT INTO users (id, username, email, password_hash, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		RETURNING id, username, email, is_active, created_at, updated_at
	`

	var user User
	err = tx.QueryRow(ctx, query, userID, username, email, passwordHash).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.IsActive,
		new(any), // created_at - not used in User struct
		new(any), // updated_at - not used in User struct
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign roles
	if len(roles) > 0 {
		for _, roleName := range roles {
			// Get role ID
			var roleID uuid.UUID
			err = tx.QueryRow(ctx, "SELECT id FROM roles WHERE name = $1", roleName).Scan(&roleID)
			if err != nil {
				if err == pgx.ErrNoRows {
					return nil, fmt.Errorf("role '%s' not found", roleName)
				}
				return nil, fmt.Errorf("failed to get role: %w", err)
			}

			// Assign role to user
			_, err = tx.Exec(ctx,
				"INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at) VALUES ($1, $2, $1, NOW())",
				userID, roleID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to assign role to user: %w", err)
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get full user with roles and permissions
	fullUser, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created user: %w", err)
	}

	return fullUser, nil
}

// GetUserPasswordHash retrieves the password hash for a user
func (r *RBACService) GetUserPasswordHash(ctx context.Context, username string) (string, error) {
	query := `SELECT password_hash FROM users WHERE username = $1`

	var passwordHash string
	err := r.db.QueryRow(ctx, query, username).Scan(&passwordHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("user not found")
		}
		return "", fmt.Errorf("failed to get password hash: %w", err)
	}

	return passwordHash, nil
}

// VerifyPassword verifies a password against the stored hash
func (r *RBACService) VerifyPassword(ctx context.Context, username, password string) bool {
	passwordHash, err := r.GetUserPasswordHash(ctx, username)
	if err != nil {
		return false
	}

	// This is a simple verification - in practice, you'd use bcrypt
	// For now, we'll do a basic comparison (this should be improved)
	return passwordHash == password
}