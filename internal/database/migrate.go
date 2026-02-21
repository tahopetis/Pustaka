package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

// RunMigrations executes all pending database migrations
// databaseURL: PostgreSQL connection string
// migrationsPath: Path to migrations directory (can be relative or absolute)
// logger: Logger instance for logging migration progress
func RunMigrations(databaseURL string, migrationsPath string, logger *pustakaLogger.Logger) error {
	// Convert to absolute path if relative
	if !filepath.IsAbs(migrationsPath) {
		abs, err := filepath.Abs(migrationsPath)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path for migrations: %w", err)
		}
		migrationsPath = abs
	}

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", migrationsPath)
	}

	// Construct the full migration source URL
	migrationsSourceURL := fmt.Sprintf("file://%s", migrationsPath)

	logger.Info().
		Str("path", migrationsPath).
		Msg("Initializing database migrations")

	// Create migration instance with retry logic for database connection
	var m *migrate.Migrate
	var err error

	// Retry up to 30 times with 2 second intervals (total 1 minute timeout)
	maxRetries := 30
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		m, err = migrate.New(
			migrationsSourceURL,
			databaseURL,
		)
		if err == nil {
			break
		}

		// Check if error is connection-related
		if i < maxRetries-1 {
			logger.Warn().
				Err(err).
				Int("attempt", i+1).
				Int("max_retries", maxRetries).
				Msg("Failed to connect to database for migrations, retrying...")
			time.Sleep(retryInterval)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to create migration instance after %d attempts: %w", maxRetries, err)
	}
	defer m.Close()

	// Get current migration version
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		logger.Info().Msg("No migrations have been applied yet")
	} else {
		if dirty {
			logger.Warn().
				Uint("version", version).
				Msg("Database is in a dirty state, attempting to fix...")
			if err := m.Force(int(version)); err != nil {
				return fmt.Errorf("failed to fix dirty database state: %w", err)
			}
			logger.Info().Msg("Database dirty state fixed")
		}
		logger.Info().
			Uint("current_version", version).
			Msg("Current migration version")
	}

	// Apply all pending migrations
	logger.Info().Msg("Applying pending migrations...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		logger.Info().Msg("No new migrations to apply")
	} else {
		logger.Info().Msg("Database migrations completed successfully")
	}

	// Get final version
	finalVersion, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get final migration version: %w", err)
	}

	if err == nil {
		logger.Info().
			Uint("final_version", finalVersion).
			Msg("Database is at migration version")
	}

	return nil
}
