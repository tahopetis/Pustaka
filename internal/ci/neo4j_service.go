package ci

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

type Neo4jService struct {
	repo   *Neo4jRepository
	logger *pustakaLogger.Logger
}

func NewNeo4jService(driver neo4j.DriverWithContext, logger *pustakaLogger.Logger) *Neo4jService {
	return &Neo4jService{
		repo:   NewNeo4jRepository(driver, logger),
		logger: logger,
	}
}

func (s *Neo4jService) Initialize(ctx context.Context) error {
	return s.repo.InitializeSchema(ctx)
}

func (s *Neo4jService) Close(ctx context.Context) error {
	return s.repo.Close(ctx)
}

// CI Operations

func (s *Neo4jService) SyncCI(ctx context.Context, ci *ConfigurationItem) error {
	return s.repo.SyncCI(ctx, ci)
}

func (s *Neo4jService) UpdateCI(ctx context.Context, ci *ConfigurationItem) error {
	return s.repo.UpdateCI(ctx, ci)
}

func (s *Neo4jService) DeleteCI(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteCI(ctx, id)
}

// Relationship Operations

func (s *Neo4jService) CreateRelationship(ctx context.Context, rel *Relationship, sourceCI, targetCI *ConfigurationItem) error {
	return s.repo.CreateRelationship(ctx, rel, sourceCI, targetCI)
}

func (s *Neo4jService) UpdateRelationship(ctx context.Context, rel *Relationship) error {
	return s.repo.UpdateRelationship(ctx, rel)
}

func (s *Neo4jService) DeleteRelationship(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteRelationship(ctx, id)
}

// Query Operations

func (s *Neo4jService) GetCIRelationships(ctx context.Context, ciID uuid.UUID) ([]RelationshipGraph, error) {
	return s.repo.GetCIRelationships(ctx, ciID)
}

func (s *Neo4jService) GetGraphData(ctx context.Context, filters GraphFilters) (*GraphData, error) {
	return s.repo.GetGraphData(ctx, filters)
}

// Advanced Graph Operations

func (s *Neo4jService) GetCINetwork(ctx context.Context, ciID uuid.UUID, depth int) (*CINetwork, error) {
	session := s.repo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH (center:ConfigurationItem {id: $ci_id})
			MATCH path = (center)-[*1..$depth]-(related:ConfigurationItem)
			RETURN path, related
			LIMIT 100
		`

		params := map[string]interface{}{
			"ci_id": ciID.String(),
			"depth": depth,
		}

		cursor, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, fmt.Errorf("failed to query CI network: %w", err)
		}

		network := &CINetwork{
			Center: ciID,
			Nodes:  make([]GraphNode, 0),
			Edges:  make([]GraphEdge, 0),
		}
		nodeMap := make(map[string]bool)

		for cursor.Next(ctx) {
			record := cursor.Record()

			// Extract nodes and relationships from path
			// path := record.Values[0] // TODO: Process path to extract relationships
			relatedNode := record.Values[1]

			// Process related node
			if relatedNode != nil {
				relatedProps := relatedNode.(neo4j.Node).Props
				relatedID := relatedProps["id"].(string)

				relatedIDUUID, err := uuid.Parse(relatedID)
				if err != nil {
					return nil, fmt.Errorf("failed to parse related ID: %w", err)
				}

				if !nodeMap[relatedID] {
					node := GraphNode{
						ID:   relatedIDUUID,
						Name: relatedProps["name"].(string),
						Type: relatedProps["type"].(string),
					}
					network.Nodes = append(network.Nodes, node)
					nodeMap[relatedID] = true
				}
			}

			// TODO: Process path to extract relationships - Neo4j Path API needs to be updated
			// Note: This is commented out to avoid syntax issues
		}

		return network, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get CI network: %w", err)
	}

	return result.(*CINetwork), nil
}

func (s *Neo4jService) FindCycles(ctx context.Context) ([][]uuid.UUID, error) {
	session := s.repo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH path = (ci:ConfigurationItem)-[:RELATES_TO*]->(ci)
			WITH [node IN nodes(path) | node.id] AS cycle
			RETURN cycle
			LIMIT 50
		`

		cursor, err := tx.Run(ctx, cypher, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to query cycles: %w", err)
		}

		var cycles [][]uuid.UUID
		for cursor.Next(ctx) {
			record := cursor.Record()

			cycleInterface := record.Values[0]
			cycleSlice := cycleInterface.([]interface{})

			var cycle []uuid.UUID
			for _, idInterface := range cycleSlice {
				if idStr, ok := idInterface.(string); ok {
					id, err := uuid.Parse(idStr)
					if err == nil {
						cycle = append(cycle, id)
					}
				}
			}

			if len(cycle) > 0 {
				cycles = append(cycles, cycle)
			}
		}

		return cycles, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find cycles: %w", err)
	}

	return result.([][]uuid.UUID), nil
}

func (s *Neo4jService) GetImpactAnalysis(ctx context.Context, ciID uuid.UUID) (*ImpactAnalysis, error) {
	session := s.repo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH (ci:ConfigurationItem {id: $ci_id})

			// Downstream dependencies (what depends on this CI)
			MATCH (ci)<-[depRel:RELATES_TO*]-(dependent:ConfigurationItem)
			WITH ci, dependent, depRel, length(depRel) AS depth

			// Upstream dependencies (what this CI depends on)
			MATCH (ci)-[depRel2:RELATES_TO*]->(dependency:ConfigurationItem)
			WITH ci, dependent, dependency, depRel, depRel2, depth, length(depRel2) AS depth2

			RETURN
				COLLECT(DISTINCT {
					id: dependent.id,
					name: dependent.name,
					type: dependent.type,
					depth: depth,
					direction: 'downstream'
				}) AS downstream,
				COLLECT(DISTINCT {
					id: dependency.id,
					name: dependency.name,
					type: dependency.type,
					depth: depth2,
					direction: 'upstream'
				}) AS upstream
			LIMIT 100
		`

		params := map[string]interface{}{
			"ci_id": ciID.String(),
		}

		cursor, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, fmt.Errorf("failed to query impact analysis: %w", err)
		}

		impact := &ImpactAnalysis{
			CI:          ciID,
			Downstream:  make([]CIImpact, 0),
			Upstream:    make([]CIImpact, 0),
		}

		if cursor.Next(ctx) {
			record := cursor.Record()

			// Process downstream dependencies
			if downstreamInterface := record.Values[0]; downstreamInterface != nil {
				downstreamList := downstreamInterface.([]interface{})
				for _, depInterface := range downstreamList {
					dep := depInterface.(map[string]interface{})
					impact.Downstream = append(impact.Downstream, CIImpact{
						ID:        uuid.MustParse(dep["id"].(string)),
						Name:      dep["name"].(string),
						Type:      dep["type"].(string),
						Depth:     int(dep["depth"].(int64)),
						Direction: dep["direction"].(string),
					})
				}
			}

			// Process upstream dependencies
			if upstreamInterface := record.Values[1]; upstreamInterface != nil {
				upstreamList := upstreamInterface.([]interface{})
				for _, depInterface := range upstreamList {
					dep := depInterface.(map[string]interface{})
					impact.Upstream = append(impact.Upstream, CIImpact{
						ID:        uuid.MustParse(dep["id"].(string)),
						Name:      dep["name"].(string),
						Type:      dep["type"].(string),
						Depth:     int(dep["depth"].(int64)),
						Direction: dep["direction"].(string),
					})
				}
			}
		}

		return impact, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get impact analysis: %w", err)
	}

	return result.(*ImpactAnalysis), nil
}

func (s *Neo4jService) GetCITypesByUsage(ctx context.Context) ([]CITypeUsage, error) {
	session := s.repo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH (ci:ConfigurationItem)
			WITH ci.type AS ciType, count(*) AS count
			RETURN ciType, count
			ORDER BY count DESC
		`

		cursor, err := tx.Run(ctx, cypher, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to query CI types by usage: %w", err)
		}

		var usage []CITypeUsage
		for cursor.Next(ctx) {
			record := cursor.Record()

			usage = append(usage, CITypeUsage{
				Type:  record.Values[0].(string),
				Count: int(record.Values[1].(int64)),
			})
		}

		return usage, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get CI types by usage: %w", err)
	}

	return result.([]CITypeUsage), nil
}

func (s *Neo4jService) GetMostConnectedCIs(ctx context.Context, limit int) ([]CIConnectivity, error) {
	session := s.repo.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH (ci:ConfigurationItem)-[r:RELATES_TO]-(related:ConfigurationItem)
			WITH ci, count(DISTINCT related) AS connectionCount
			RETURN ci.id AS id, ci.name AS name, ci.type AS type, connectionCount
			ORDER BY connectionCount DESC
			LIMIT $limit
		`

		params := map[string]interface{}{
			"limit": limit,
		}

		cursor, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, fmt.Errorf("failed to query most connected CIs: %w", err)
		}

		var connectivity []CIConnectivity
		for cursor.Next(ctx) {
			record := cursor.Record()

			connectivity = append(connectivity, CIConnectivity{
				ID:             uuid.MustParse(record.Values[0].(string)),
				Name:           record.Values[1].(string),
				Type:           record.Values[2].(string),
				ConnectionCount: int(record.Values[3].(int64)),
			})
		}

		return connectivity, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get most connected CIs: %w", err)
	}

	return result.([]CIConnectivity), nil
}