package database

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/rs/zerolog/log"
)

type Neo4jDB struct {
	Driver neo4j.DriverWithContext
	Database string
}

func NewNeo4jDB(uri, username, password, database string, maxPoolSize int) (*Neo4jDB, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, fmt.Errorf("unable to create Neo4j driver: %w", err)
	}

	// Configure connection pool
	err = driver.VerifyConnectivity(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to verify Neo4j connectivity: %w", err)
	}

	log.Info().
		Str("uri", uri).
		Str("database", database).
		Int("max_pool_size", maxPoolSize).
		Msg("Connected to Neo4j")

	return &Neo4jDB{
		Driver: driver,
		Database: database,
	}, nil
}

func (db *Neo4jDB) Close() error {
	if db.Driver != nil {
		err := db.Driver.Close(context.Background())
		if err != nil {
			return fmt.Errorf("error closing Neo4j driver: %w", err)
		}
		log.Info().Msg("Neo4j driver closed")
	}
	return nil
}

func (db *Neo4jDB) Health(ctx context.Context) error {
	session := db.Driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
		DatabaseName: db.Database,
	})
	defer session.Close(ctx)

	_, err := session.Run(ctx, "RETURN 1", nil)
	return err
}

// NewSession creates a new Neo4j session
func (db *Neo4jDB) NewSession(ctx context.Context, config neo4j.SessionConfig) neo4j.SessionWithContext {
	if config.DatabaseName == "" {
		config.DatabaseName = db.Database
	}
	return db.Driver.NewSession(ctx, config)
}

// ExecuteRead executes a read query
func (db *Neo4jDB) ExecuteRead(ctx context.Context, query string, params map[string]interface{}, handleResult func(neo4j.ResultWithContext) error) error {
	session := db.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to execute read query: %w", err)
	}

	return handleResult(result)
}

// ExecuteWrite executes a write query
func (db *Neo4jDB) ExecuteWrite(ctx context.Context, query string, params map[string]interface{}, handleResult func(neo4j.ResultWithContext) error) error {
	session := db.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to execute write query: %w", err)
	}

	return handleResult(result)
}

// ExecuteWriteTransaction executes a write transaction
func (db *Neo4jDB) ExecuteWriteTransaction(ctx context.Context, work func(neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	session := db.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	return session.ExecuteWrite(ctx, work)
}

// InitializeIndexes creates necessary indexes for the graph database
func (db *Neo4jDB) InitializeIndexes(ctx context.Context) error {
	queries := []string{
		// Configuration item indexes
		"CREATE INDEX ci_id_idx IF NOT EXISTS FOR (n:ConfigurationItem) ON (n.id)",
		"CREATE INDEX ci_type_idx IF NOT EXISTS FOR (n:ConfigurationItem) ON (n.type)",
		"CREATE INDEX ci_name_idx IF NOT EXISTS FOR (n:ConfigurationItem) ON (n.name)",
		"CREATE INDEX ci_created_at_idx IF NOT EXISTS FOR (n:ConfigurationItem) ON (n.created_at)",

		// User indexes
		"CREATE INDEX user_id_idx IF NOT EXISTS FOR (n:User) ON (n.id)",
		"CREATE INDEX user_username_idx IF NOT EXISTS FOR (n:User) ON (n.username)",

		// Full-text search index
		"CREATE FULLTEXT INDEX ci_search_idx IF NOT EXISTS FOR (n:ConfigurationItem) ON EACH [n.name, n.type]",
	}

	for _, query := range queries {
		err := db.ExecuteWrite(ctx, query, nil, func(result neo4j.ResultWithContext) error {
			// Consume the result
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to create index with query '%s': %w", query, err)
		}
	}

	log.Info().Msg("Neo4j indexes initialized")
	return nil
}