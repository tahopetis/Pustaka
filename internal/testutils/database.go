package testutils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testDatabaseURLEnv names the DSN of the PostgreSQL instance used by
// database-backed tests. When it is unset those tests skip instead of starting
// a container: the previous dockertest helper required a Docker daemon and
// pulled the retired github.com/docker/docker module into the module graph.
const testDatabaseURLEnv = "TEST_DATABASE_URL"

// allowUnsafeTestDatabaseEnv opts out of the "_test database name" guard. Only
// use it for throwaway instances: the helper creates tables and deletes rows.
const allowUnsafeTestDatabaseEnv = "TEST_DATABASE_URL_ALLOW_ANY"

// testDatabaseConfig validates the configured DSN and returns a pool config.
// skip=true means no DSN is configured and the caller should skip the test.
//
// The guard exists because this helper is destructive (DDL + DELETE): the
// target database name must contain "test" unless explicitly overridden.
func testDatabaseConfig(raw string, allowUnsafe bool) (cfg *pgxpool.Config, skip bool, err error) {
	dsn := strings.TrimSpace(raw)
	if dsn == "" {
		return nil, true, nil
	}

	cfg, err = pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, false, fmt.Errorf("invalid %s: %w", testDatabaseURLEnv, err)
	}

	name := cfg.ConnConfig.Database
	if !allowUnsafe && !strings.Contains(strings.ToLower(name), "test") {
		return nil, false, fmt.Errorf(
			"%s points at database %q, but the test helper creates tables and deletes rows; "+
				"the database name must contain \"test\" (set %s=1 to override)",
			testDatabaseURLEnv, name, allowUnsafeTestDatabaseEnv,
		)
	}

	return cfg, false, nil
}

// SetupTestDB connects to the PostgreSQL instance configured by
// TEST_DATABASE_URL and applies the real schema migrations. It skips the test
// when the variable is unset, so the default suite needs no Docker daemon.
func SetupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv(testDatabaseURLEnv))
	cfg, skip, err := testDatabaseConfig(
		dsn,
		os.Getenv(allowUnsafeTestDatabaseEnv) != "",
	)
	if err != nil {
		t.Fatal(err)
	}
	if skip {
		t.Skipf("%s not set; skipping database-backed test", testDatabaseURLEnv)
	}

	ctx := context.Background()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("could not create test database pool: %v", err)
	}

	cleanup := func() { pool.Close() }

	if err := waitForDatabase(ctx, pool); err != nil {
		cleanup()
		t.Fatalf("could not connect to test database: %v", err)
	}

	if err := runMigrations(dsn); err != nil {
		cleanup()
		t.Fatalf("could not apply test schema: %v", err)
	}

	return pool, cleanup
}

// waitForDatabase retries the ping until PostgreSQL accepts connections. Each
// attempt is bounded so one hung connect cannot exceed the overall 30s budget.
func waitForDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		attemptCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		lastErr = pool.Ping(attemptCtx)
		cancel()
		if lastErr == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return lastErr
}

// CleanupDB truncates every public table (except golang-migrate's
// schema_migrations) between tests. Discovered dynamically so the helper keeps
// working as migrations add tables.
func CleanupDB(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()

	rows, err := db.Query(ctx, `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public' AND tablename <> 'schema_migrations'
		ORDER BY tablename
	`)
	if err != nil {
		t.Fatalf("could not list test tables: %v", err)
	}

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			t.Fatalf("could not scan table name: %v", err)
		}
		tables = append(tables, name)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("could not read test tables: %v", err)
	}

	if len(tables) == 0 {
		return
	}

	statement := "TRUNCATE TABLE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE"
	if _, err := db.Exec(ctx, statement); err != nil {
		t.Fatalf("could not clean test database: %v", err)
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

// runMigrations applies the real schema migrations from cmd/migrations, so the
// test database matches production instead of a hand-maintained copy.
func runMigrations(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open migration connection: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir(), "pgx", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// migrationsDir resolves cmd/migrations relative to this source file so the
// helper works regardless of the test working directory.
func migrationsDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("cmd", "migrations")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "cmd", "migrations")
}
