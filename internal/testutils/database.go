package testutils

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// testDatabaseURLEnv names the DSN of the PostgreSQL instance used by
// database-backed tests. When it is unset those tests skip instead of starting
// a container: the previous dockertest helper required a Docker daemon and
// pulled the retired github.com/docker/docker module into the module graph.
const testDatabaseURLEnv = "TEST_DATABASE_URL"

// databaseURLFromEnv resolves the test DSN, trimming accidental whitespace.
func databaseURLFromEnv(getenv func(string) string) string {
	return strings.TrimSpace(getenv(testDatabaseURLEnv))
}

// skipMessage is the reason reported when no test database is configured.
func skipMessage() string {
	return fmt.Sprintf("%s not set; skipping database-backed test", testDatabaseURLEnv)
}

// SetupTestDB connects to the PostgreSQL instance configured by
// TEST_DATABASE_URL and applies the minimal test schema. It skips the test when
// the variable is unset, so the default suite needs no Docker daemon.
func SetupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	dsn := databaseURLFromEnv(os.Getenv)
	if dsn == "" {
		t.Skip(skipMessage())
	}

	ctx := context.Background()

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("invalid %s: %v", testDatabaseURLEnv, err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("could not create test database pool: %v", err)
	}

	cleanup := func() { pool.Close() }

	if err := waitForDatabase(ctx, pool); err != nil {
		cleanup()
		t.Fatalf("could not connect to test database: %v", err)
	}

	if err := runMigrations(pool); err != nil {
		cleanup()
		t.Fatalf("could not apply test schema: %v", err)
	}

	return pool, cleanup
}

// waitForDatabase retries the ping until PostgreSQL accepts connections.
func waitForDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		if lastErr = pool.Ping(ctx); lastErr == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return lastErr
}

// CleanupDB cleans up all data from test database
func CleanupDB(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()

	// Delete all data in correct order respecting foreign keys
	tables := []string{
		"audit_logs",
		"relationships",
		"configuration_items",
		"user_roles",
		"role_permissions",
		"users",
		"roles",
		"permissions",
		"ci_type_definitions",
	}

	for _, table := range tables {
		_, err := db.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: Could not clean table %s: %v", table, err)
		}
	}
}

// InsertTestData inserts basic test data
func InsertTestData(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()

	// Insert roles
	_, err := db.Exec(ctx, `
		INSERT INTO roles (id, name, description) VALUES
		('550e8400-e29b-41d4-a716-446655440001', 'admin', 'Full system access'),
		('550e8400-e29b-41d4-a716-446655440002', 'editor', 'CI and relationship management'),
		('550e8400-e29b-41d4-a716-446655440003', 'viewer', 'Read-only access')
	`)
	if err != nil {
		t.Fatalf("Could not insert test roles: %v", err)
	}

	// Insert permissions
	_, err = db.Exec(ctx, `
		INSERT INTO permissions (id, name, description, resource_type) VALUES
		('550e8400-e29b-41d4-a716-446655440010', 'ci:read', 'Read configuration items', 'ci'),
		('550e8400-e29b-41d4-a716-446655440011', 'ci:create', 'Create configuration items', 'ci'),
		('550e8400-e29b-41d4-a716-446655440012', 'ci:update', 'Update configuration items', 'ci'),
		('550e8400-e29b-41d4-a716-446655440013', 'ci:delete', 'Delete configuration items', 'ci'),
		('550e8400-e29b-41d4-a716-446655440020', 'ci_type:read', 'Read CI type definitions', 'ci_type'),
		('550e8400-e29b-41d4-a716-446655440021', 'ci_type:create', 'Create CI type definitions', 'ci_type'),
		('550e8400-e29b-41d4-a716-446655440030', 'relationship:read', 'Read relationships', 'relationship'),
		('550e8400-e29b-41d4-a716-446655440031', 'relationship:create', 'Create relationships', 'relationship'),
		('550e8400-e29b-41d4-a716-446655440040', 'user:read', 'Read users', 'user'),
		('550e8400-e29b-41d4-a716-446655440041', 'user:create', 'Create users', 'user'),
		('550e8400-e29b-41d4-a716-446655440050', 'audit:read', 'Read audit logs', 'audit')
	`)
	if err != nil {
		t.Fatalf("Could not insert test permissions: %v", err)
	}

	// Assign permissions to roles
	_, err = db.Exec(ctx, `
		INSERT INTO role_permissions (role_id, permission_id) VALUES
		('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440010'),
		('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440011'),
		('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440012'),
		('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440013'),
		('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440020'),
		('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440021'),
		('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440030'),
		('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440031'),
		('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440040'),
		('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440041'),
		('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440050'),
		('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440010'),
		('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440011'),
		('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440012'),
		('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440013'),
		('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440020'),
		('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440030'),
		('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440031'),
		('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440050'),
		('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440010'),
		('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440020'),
		('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440030'),
		('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440050')
	`)
	if err != nil {
		t.Fatalf("Could not assign permissions to roles: %v", err)
	}
}

// runMigrations runs the database schema migrations
func runMigrations(db *pgxpool.Pool) error {
	ctx := context.Background()

	// Wait for database to be ready
	for i := 0; i < 30; i++ {
		if err := db.Ping(ctx); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Enable UUID extension
	_, err := db.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	if err != nil {
		return fmt.Errorf("failed to enable UUID extension: %w", err)
	}

	// Create tables (simplified version of the migration)
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(100) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(50) UNIQUE NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS permissions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) UNIQUE NOT NULL,
			description TEXT,
			resource_type VARCHAR(50),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS role_permissions (
			role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
			permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
			PRIMARY KEY (role_id, permission_id)
		);

		CREATE TABLE IF NOT EXISTS user_roles (
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
			assigned_by UUID REFERENCES users(id),
			assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			PRIMARY KEY (user_id, role_id)
		);

		CREATE TABLE IF NOT EXISTS ci_type_definitions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) UNIQUE NOT NULL,
			description TEXT,
			required_attributes JSONB NOT NULL DEFAULT '[]',
			optional_attributes JSONB NOT NULL DEFAULT '[]',
			created_by UUID REFERENCES users(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS configuration_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			ci_type VARCHAR(100) NOT NULL REFERENCES ci_type_definitions(name),
			attributes JSONB NOT NULL DEFAULT '{}',
			tags TEXT[] DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_by UUID REFERENCES users(id),
			updated_by UUID REFERENCES users(id),
			CONSTRAINT unique_name_per_type UNIQUE (name, ci_type)
		);

		CREATE TABLE IF NOT EXISTS relationships (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			source_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
			target_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
			relationship_type VARCHAR(50) NOT NULL,
			attributes JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_by UUID REFERENCES users(id),
			CONSTRAINT no_self_relationship CHECK (source_id != target_id),
			CONSTRAINT unique_relationship UNIQUE (source_id, target_id, relationship_type)
		);

		CREATE TABLE IF NOT EXISTS audit_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			entity_type VARCHAR(50) NOT NULL,
			entity_id UUID,
			action VARCHAR(50) NOT NULL,
			performed_by UUID REFERENCES users(id),
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			details JSONB NOT NULL DEFAULT '{}',
			ip_address INET,
			user_agent TEXT
		);
	`)

	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Create indexes
	_, err = db.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_cis_type ON configuration_items(ci_type);
		CREATE INDEX IF NOT EXISTS idx_cis_name ON configuration_items(name);
		CREATE INDEX IF NOT EXISTS idx_relationships_source ON relationships(source_id);
		CREATE INDEX IF NOT EXISTS idx_relationships_target ON relationships(target_id);
	`)

	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}