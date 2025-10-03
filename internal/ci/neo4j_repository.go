package ci

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

type Neo4jRepository struct {
	driver neo4j.DriverWithContext
	logger *pustakaLogger.Logger
}

func NewNeo4jRepository(driver neo4j.DriverWithContext, logger *pustakaLogger.Logger) *Neo4jRepository {
	return &Neo4jRepository{
		driver: driver,
		logger: logger,
	}
}

func (r *Neo4jRepository) Close(ctx context.Context) error {
	return r.driver.Close(ctx)
}

// CI Operations

func (r *Neo4jRepository) SyncCI(ctx context.Context, ci *ConfigurationItem) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Create CI node with properties
		cypher := `
			MERGE (ci:ConfigurationItem {id: $id})
			SET ci.name = $name,
				ci.type = $type,
				ci.attributes = $attributes,
				ci.tags = $tags,
				ci.created_at = $created_at,
				ci.updated_at = $updated_at,
				ci.created_by = $created_by
			RETURN ci
		`

		attributesJSON, _ := json.Marshal(ci.Attributes)
		tagsJSON, _ := json.Marshal(ci.Tags)

		params := map[string]interface{}{
			"id":          ci.ID.String(),
			"name":        ci.Name,
			"type":        ci.CIType,
			"attributes":  string(attributesJSON),
			"tags":        string(tagsJSON),
			"created_at":  ci.CreatedAt.Unix(),
			"updated_at":  ci.UpdatedAt.Unix(),
			"created_by":  ci.CreatedBy.String(),
		}

		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			r.logger.Error()
			return nil, fmt.Errorf("failed to sync CI: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to sync CI: %w", err)
	}

	r.logger.Info().Interface("details", map[string]interface{}{
		"ci_id":   ci.ID,
		"ci_name": ci.Name,
	})

	return nil
}

func (r *Neo4jRepository) UpdateCI(ctx context.Context, ci *ConfigurationItem) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH (ci:ConfigurationItem {id: $id})
			SET ci.name = $name,
				ci.type = $type,
				ci.attributes = $attributes,
				ci.tags = $tags,
				ci.updated_at = $updated_at
			RETURN ci
		`

		attributesJSON, _ := json.Marshal(ci.Attributes)
		tagsJSON, _ := json.Marshal(ci.Tags)

		params := map[string]interface{}{
			"id":          ci.ID.String(),
			"name":        ci.Name,
			"type":        ci.CIType,
			"attributes":  string(attributesJSON),
			"tags":        string(tagsJSON),
			"updated_at":  ci.UpdatedAt.Unix(),
		}

		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			r.logger.Error().Interface("details", map[string]interface{}{
				"ci_id": ci.ID,
			})
			return nil, fmt.Errorf("failed to update CI: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to update CI: %w", err)
	}

	r.logger.Info().Interface("details", map[string]interface{}{
		"ci_id": ci.ID,
	})

	return nil
}

func (r *Neo4jRepository) DeleteCI(ctx context.Context, id uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Delete the CI node and all its relationships
		cypher := `
			MATCH (ci:ConfigurationItem {id: $id})
			DETACH DELETE ci
		`

		params := map[string]interface{}{
			"id": id.String(),
		}

		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			r.logger.Error().Interface("details", map[string]interface{}{
				"ci_id": id,
			})
			return nil, fmt.Errorf("failed to delete CI: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to delete CI: %w", err)
	}

	r.logger.Info().Interface("details", map[string]interface{}{
		"ci_id": id,
	})

	return nil
}

// Relationship Operations

func (r *Neo4jRepository) CreateRelationship(ctx context.Context, rel *Relationship, sourceCI, targetCI *ConfigurationItem) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH (source:ConfigurationItem {id: $source_id})
			MATCH (target:ConfigurationItem {id: $target_id})
			CREATE (source)-[r:RELATES_TO {
				id: $rel_id,
				type: $rel_type,
				attributes: $attributes,
				created_at: $created_at,
				created_by: $created_by
			}]->(target)
			RETURN r
		`

		attributesJSON, _ := json.Marshal(rel.Attributes)

		params := map[string]interface{}{
			"source_id":   sourceCI.ID.String(),
			"target_id":   targetCI.ID.String(),
			"rel_id":      rel.ID.String(),
			"rel_type":    rel.RelationshipType,
			"attributes":  string(attributesJSON),
			"created_at":  rel.CreatedAt.Unix(),
			"created_by":  rel.CreatedBy.String(),
		}

		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			r.logger.Error().Interface("details", map[string]interface{}{
				"relationship_id": rel.ID,
				"source_id":       sourceCI.ID,
				"target_id":       targetCI.ID,
			})
			return nil, fmt.Errorf("failed to create relationship: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to create relationship: %w", err)
	}

	r.logger.Info().Interface("details", map[string]interface{}{
		"relationship_id": rel.ID,
		"source_id":       sourceCI.ID,
		"target_id":       targetCI.ID,
	})

	return nil
}

func (r *Neo4jRepository) UpdateRelationship(ctx context.Context, rel *Relationship) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH ()-[r:RELATES_TO {id: $rel_id}]-()
			SET r.type = $rel_type,
				r.attributes = $attributes,
				r.updated_at = $updated_at
			RETURN r
		`

		attributesJSON, _ := json.Marshal(rel.Attributes)

		params := map[string]interface{}{
			"rel_id":     rel.ID.String(),
			"rel_type":   rel.RelationshipType,
			"attributes": string(attributesJSON),
			"updated_at": time.Now().Unix(),
		}

		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			r.logger.Error().Interface("details", map[string]interface{}{
				"relationship_id": rel.ID,
			})
			return nil, fmt.Errorf("failed to update relationship: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to update relationship: %w", err)
	}

	r.logger.Info().Interface("details", map[string]interface{}{
		"relationship_id": rel.ID,
	})

	return nil
}

func (r *Neo4jRepository) DeleteRelationship(ctx context.Context, id uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH ()-[r:RELATES_TO {id: $rel_id}]-()
			DELETE r
		`

		params := map[string]interface{}{
			"rel_id": id.String(),
		}

		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			r.logger.Error().Interface("details", map[string]interface{}{
				"relationship_id": id,
			})
			return nil, fmt.Errorf("failed to delete relationship: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	r.logger.Info().Interface("details", map[string]interface{}{
		"relationship_id": id,
	})

	return nil
}

// Query Operations

func (r *Neo4jRepository) GetCIRelationships(ctx context.Context, ciID uuid.UUID) ([]RelationshipGraph, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH (source:ConfigurationItem {id: $ci_id})-[r:RELATES_TO]->(target:ConfigurationItem)
			RETURN r, target
			UNION
			MATCH (source:ConfigurationItem)-[r:RELATES_TO]->(target:ConfigurationItem {id: $ci_id})
			RETURN r, source
		`

		params := map[string]interface{}{
			"ci_id": ciID.String(),
		}

		cursor, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, fmt.Errorf("failed to query relationships: %w", err)
		}

		var relationships []RelationshipGraph
		for cursor.Next(ctx) {
			record := cursor.Record()

			relNode := record.Values[0]
			ciNode := record.Values[1]

			relProps := relNode.(neo4j.Node).Props
			ciProps := ciNode.(neo4j.Node).Props

			// Parse relationship attributes
			var attributes map[string]interface{}
			if attrStr, ok := relProps["attributes"].(string); ok && attrStr != "" {
				json.Unmarshal([]byte(attrStr), &attributes)
			}

			// Parse CI attributes
			var ciAttributes map[string]interface{}
			if attrStr, ok := ciProps["attributes"].(string); ok && attrStr != "" {
				json.Unmarshal([]byte(attrStr), &ciAttributes)
			}

			// Parse CI tags
			var tags []string
			if tagsStr, ok := ciProps["tags"].(string); ok && tagsStr != "" {
				json.Unmarshal([]byte(tagsStr), &tags)
			}

			rel := RelationshipGraph{
				ID:                uuid.MustParse(relProps["id"].(string)),
				RelationshipType:  relProps["type"].(string),
				Attributes:        attributes,
				CreatedAt:         time.Unix(relProps["created_at"].(int64), 0),
				CreatedBy:         uuid.MustParse(relProps["created_by"].(string)),
				RelatedCI: ConfigurationItem{
					ID:         uuid.MustParse(ciProps["id"].(string)),
					Name:       ciProps["name"].(string),
					CIType:     ciProps["type"].(string),
					Attributes: ciAttributes,
					Tags:       tags,
					CreatedAt:  time.Unix(ciProps["created_at"].(int64), 0),
					UpdatedAt:  time.Unix(ciProps["updated_at"].(int64), 0),
					CreatedBy:  uuid.MustParse(ciProps["created_by"].(string)),
				},
			}

			relationships = append(relationships, rel)
		}

		return relationships, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get CI relationships: %w", err)
	}

	return result.([]RelationshipGraph), nil
}

func (r *Neo4jRepository) GetGraphData(ctx context.Context, filters GraphFilters) (*GraphData, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Build base query
		cypher := `
			MATCH (ci:ConfigurationItem)
		`

		whereClause := ""
		params := make(map[string]interface{})

		// Add filters
		if len(filters.CITypes) > 0 {
			whereClause += "WHERE ci.type IN $ci_types "
			params["ci_types"] = filters.CITypes
		}

		if filters.Search != "" {
			if whereClause != "" {
				whereClause += "AND "
			} else {
				whereClause += "WHERE "
			}
			whereClause += "toLower(ci.name) CONTAINS toLower($search) "
			params["search"] = filters.Search
		}

		cypher += whereClause

		// Get nodes and relationships
		cypher += `
			OPTIONAL MATCH (ci)-[r:RELATES_TO]->(related:ConfigurationItem)
			RETURN ci, r, related
			LIMIT $limit
		`

		params["limit"] = filters.Limit
		if params["limit"] == nil {
			params["limit"] = 100
		}

		cursor, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, fmt.Errorf("failed to query graph data: %w", err)
		}

		graphData := &GraphData{
			Nodes: make([]GraphNode, 0),
			Edges: make([]GraphEdge, 0),
		}
		nodeMap := make(map[string]bool)
		edgeMap := make(map[string]bool)

		for cursor.Next(ctx) {
			record := cursor.Record()

			ciNode := record.Values[0]
			relNode := record.Values[1]
			relatedNode := record.Values[2]

			// Process CI node
			if ciNode != nil {
				ciProps := ciNode.(neo4j.Node).Props
				ciID := ciProps["id"].(string)

				if !nodeMap[ciID] {
					// Parse CI attributes
					var attributes map[string]interface{}
					if attrStr, ok := ciProps["attributes"].(string); ok && attrStr != "" {
						json.Unmarshal([]byte(attrStr), &attributes)
					}

					// Parse CI tags
					var tags []string
					if tagsStr, ok := ciProps["tags"].(string); ok && tagsStr != "" {
						json.Unmarshal([]byte(tagsStr), &tags)
					}

					ciIDUUID, err := uuid.Parse(ciID)
					if err != nil {
						return nil, fmt.Errorf("failed to parse CI ID: %w", err)
					}

					node := GraphNode{
						ID:         ciIDUUID,
						Name:       ciProps["name"].(string),
						Type:       ciProps["type"].(string),
						Attributes: attributes,
						Tags:       tags,
					}
					graphData.Nodes = append(graphData.Nodes, node)
					nodeMap[ciID] = true
				}
			}

			// Process relationship and related node
			if relNode != nil && relatedNode != nil {
				relProps := relNode.(neo4j.Relationship).Props
				relatedProps := relatedNode.(neo4j.Node).Props

				sourceID := ciNode.(neo4j.Node).Props["id"].(string)
				targetID := relatedProps["id"].(string)

				// Add related node if not already added
				if !nodeMap[targetID] {
					// Parse related CI attributes
					var attributes map[string]interface{}
					if attrStr, ok := relatedProps["attributes"].(string); ok && attrStr != "" {
						json.Unmarshal([]byte(attrStr), &attributes)
					}

					// Parse related CI tags
					var tags []string
					if tagsStr, ok := relatedProps["tags"].(string); ok && tagsStr != "" {
						json.Unmarshal([]byte(tagsStr), &tags)
					}

					targetIDUUID, err := uuid.Parse(targetID)
					if err != nil {
						return nil, fmt.Errorf("failed to parse target ID: %w", err)
					}

					node := GraphNode{
						ID:         targetIDUUID,
						Name:       relatedProps["name"].(string),
						Type:       relatedProps["type"].(string),
						Attributes: attributes,
						Tags:       tags,
					}
					graphData.Nodes = append(graphData.Nodes, node)
					nodeMap[targetID] = true
				}

				// Parse relationship attributes
				var attributes map[string]interface{}
				if attrStr, ok := relProps["attributes"].(string); ok && attrStr != "" {
					json.Unmarshal([]byte(attrStr), &attributes)
				}

				// Parse relationship ID
				relIDUUID, err := uuid.Parse(relProps["id"].(string))
				if err != nil {
					return nil, fmt.Errorf("failed to parse relationship ID: %w", err)
				}

				// Create edge key to detect duplicates
				edgeKey := fmt.Sprintf("%s-%s-%s", sourceID, targetID, relProps["type"].(string))

				// Only add edge if we haven't added it before
				if !edgeMap[edgeKey] {
				// Add link
				link := GraphEdge{
					ID:               relIDUUID,
					Source:           sourceID,
					Target:           targetID,
					RelationshipType: relProps["type"].(string),
					Attributes:       attributes,
				}
				graphData.Edges = append(graphData.Edges, link)
					edgeMap[edgeKey] = true // Mark this edge as added
				}
			}
		}

		return graphData, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get graph data: %w", err)
	}

	return result.(*GraphData), nil
}

// Schema and Index Management

func (r *Neo4jRepository) InitializeSchema(ctx context.Context) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	queries := []string{
		// Create indexes for performance
		"CREATE INDEX configuration_item_id_index IF NOT EXISTS FOR (ci:ConfigurationItem) ON (ci.id)",
		"CREATE INDEX configuration_item_name_index IF NOT EXISTS FOR (ci:ConfigurationItem) ON (ci.name)",
		"CREATE INDEX configuration_item_type_index IF NOT EXISTS FOR (ci:ConfigurationItem) ON (ci.type)",
		"CREATE INDEX relationship_id_index IF NOT EXISTS FOR ()-[r:RELATES_TO]-() ON (r.id)",
		"CREATE INDEX relationship_type_index IF NOT EXISTS FOR ()-[r:RELATES_TO]-() ON (r.type)",

		// Create constraints
		"CREATE CONSTRAINT configuration_item_id_unique IF NOT EXISTS FOR (ci:ConfigurationItem) REQUIRE ci.id IS UNIQUE",
		"CREATE CONSTRAINT relationship_id_unique IF NOT EXISTS FOR ()-[r:RELATES_TO]-() REQUIRE r.id IS UNIQUE",
	}

	for _, query := range queries {
		_, err := session.Run(ctx, query, nil)
		if err != nil {
			r.logger.Error()
			return fmt.Errorf("failed to initialize schema: %w", err)
		}
	}

	r.logger.Info()
	return nil
}