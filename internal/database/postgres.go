package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type PostgresDB struct {
	Pool *pgxpool.Pool
}

func NewPostgresDB(databaseURL string, maxOpenConns, maxIdleConns int, maxLifetime time.Duration) (*PostgresDB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Configure connection pool
	config.MaxConns = int32(maxOpenConns)
	config.MinConns = int32(maxIdleConns)
	config.MaxConnLifetime = maxLifetime
	config.HealthCheckPeriod = 30 * time.Second

	// Connection timeout
	config.ConnConfig.ConnectTimeout = 10 * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Info().
		Str("database", databaseURL).
		Int("max_conns", maxOpenConns).
		Int("min_conns", maxIdleConns).
		Msg("Connected to PostgreSQL")

	return &PostgresDB{Pool: pool}, nil
}

func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Info().Msg("PostgreSQL connection pool closed")
	}
}

func (db *PostgresDB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Query executes a query that returns rows
func (db *PostgresDB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return db.Pool.Query(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (db *PostgresDB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.Pool.QueryRow(ctx, query, args...)
}

// Exec executes a query without returning any rows
func (db *PostgresDB) Exec(ctx context.Context, query string, args ...interface{}) (string, error) {
	result, err := db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

// Begin begins a new transaction
func (db *PostgresDB) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.Pool.Begin(ctx)
}

// BeginTx begins a transaction with the given options
func (db *PostgresDB) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return db.Pool.BeginTx(ctx, txOptions)
}

// Stats returns connection pool statistics
func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}