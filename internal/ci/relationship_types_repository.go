package ci

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RelationshipTypeRepository handles database operations for relationship types
type RelationshipTypeRepository struct {
	db *pgxpool.Pool
}

func NewRelationshipTypeRepository(db *pgxpool.Pool) *RelationshipTypeRepository {
	return &RelationshipTypeRepository{
		db: db,
	}
}

// Create creates a new relationship type
func (r *RelationshipTypeRepository) Create(ctx context.Context, rt *RelationshipType) (*RelationshipType, error) {
	query := `
		INSERT INTO relationship_types (
			id, name, display_name, description, forward_label, reverse_label,
			color, icon, category, is_active, is_system, allowed_source_types,
			allowed_target_types, cardinality_source, cardinality_target,
			bidirectional, attributes, created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		) RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		rt.ID, rt.Name, rt.DisplayName, rt.Description, rt.ForwardLabel, rt.ReverseLabel,
		rt.Color, rt.Icon, rt.Category, rt.IsActive, rt.IsSystem, rt.AllowedSourceTypes,
		rt.AllowedTargetTypes, rt.CardinalitySource, rt.CardinalityTarget,
		rt.Bidirectional, rt.Attributes, rt.CreatedAt, rt.UpdatedAt, rt.CreatedBy, rt.UpdatedBy,
	).Scan(&rt.ID, &rt.CreatedAt, &rt.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create relationship type: %w", err)
	}

	return rt, nil
}

// GetByID retrieves a relationship type by ID
func (r *RelationshipTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*RelationshipType, error) {
	query := `
		SELECT id, name, display_name, description, forward_label, reverse_label,
			   color, icon, category, is_active, is_system, allowed_source_types,
			   allowed_target_types, cardinality_source, cardinality_target,
			   bidirectional, attributes, created_at, updated_at, created_by, updated_by
		FROM relationship_types
		WHERE id = $1
	`

	rt := &RelationshipType{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rt.ID, &rt.Name, &rt.DisplayName, &rt.Description, &rt.ForwardLabel, &rt.ReverseLabel,
		&rt.Color, &rt.Icon, &rt.Category, &rt.IsActive, &rt.IsSystem, &rt.AllowedSourceTypes,
		&rt.AllowedTargetTypes, &rt.CardinalitySource, &rt.CardinalityTarget,
		&rt.Bidirectional, &rt.Attributes, &rt.CreatedAt, &rt.UpdatedAt, &rt.CreatedBy, &rt.UpdatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("relationship type not found")
		}
		return nil, fmt.Errorf("failed to get relationship type: %w", err)
	}

	return rt, nil
}

// GetByName retrieves a relationship type by name
func (r *RelationshipTypeRepository) GetByName(ctx context.Context, name string) (*RelationshipType, error) {
	query := `
		SELECT id, name, display_name, description, forward_label, reverse_label,
			   color, icon, category, is_active, is_system, allowed_source_types,
			   allowed_target_types, cardinality_source, cardinality_target,
			   bidirectional, attributes, created_at, updated_at, created_by, updated_by
		FROM relationship_types
		WHERE name = $1
	`

	rt := &RelationshipType{}
	err := r.db.QueryRow(ctx, query, name).Scan(
		&rt.ID, &rt.Name, &rt.DisplayName, &rt.Description, &rt.ForwardLabel, &rt.ReverseLabel,
		&rt.Color, &rt.Icon, &rt.Category, &rt.IsActive, &rt.IsSystem, &rt.AllowedSourceTypes,
		&rt.AllowedTargetTypes, &rt.CardinalitySource, &rt.CardinalityTarget,
		&rt.Bidirectional, &rt.Attributes, &rt.CreatedAt, &rt.UpdatedAt, &rt.CreatedBy, &rt.UpdatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("relationship type not found")
		}
		return nil, fmt.Errorf("failed to get relationship type: %w", err)
	}

	return rt, nil
}

// List retrieves paginated relationship types with filters
func (r *RelationshipTypeRepository) List(ctx context.Context, filters ListRelationshipTypeFilters, page, limit int) (*RelationshipTypeListResponse, error) {
	offset := (page - 1) * limit

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filters.Search != "" {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR display_name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex+1, argIndex+2)
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
		argIndex += 3
	}

	if filters.Category != "" {
		whereClause += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, filters.Category)
		argIndex++
	}

	if filters.IsActive != nil {
		whereClause += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, *filters.IsActive)
		argIndex++
	}

	if filters.IsSystem != nil {
		whereClause += fmt.Sprintf(" AND is_system = $%d", argIndex)
		args = append(args, *filters.IsSystem)
		argIndex++
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if filters.Sort != "" {
		validSorts := map[string]string{
			"name":        "name",
			"category":    "category",
			"created_at":  "created_at",
			"updated_at":  "updated_at",
		}
		if sortField, ok := validSorts[filters.Sort]; ok {
			orderBy = fmt.Sprintf("ORDER BY %s", sortField)
			if filters.Order == "asc" {
				orderBy += " ASC"
			} else {
				orderBy += " DESC"
			}
		}
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM relationship_types %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count relationship types: %w", err)
	}

	// Get relationship types
	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT id, name, display_name, description, forward_label, reverse_label,
			   color, icon, category, is_active, is_system, allowed_source_types,
			   allowed_target_types, cardinality_source, cardinality_target,
			   bidirectional, attributes, created_at, updated_at, created_by, updated_by
		FROM relationship_types
		%s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationship types: %w", err)
	}
	defer rows.Close()

	var relationshipTypes []RelationshipType
	for rows.Next() {
		rt := RelationshipType{}
		err := rows.Scan(
			&rt.ID, &rt.Name, &rt.DisplayName, &rt.Description, &rt.ForwardLabel, &rt.ReverseLabel,
			&rt.Color, &rt.Icon, &rt.Category, &rt.IsActive, &rt.IsSystem, &rt.AllowedSourceTypes,
			&rt.AllowedTargetTypes, &rt.CardinalitySource, &rt.CardinalityTarget,
			&rt.Bidirectional, &rt.Attributes, &rt.CreatedAt, &rt.UpdatedAt, &rt.CreatedBy, &rt.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relationship type: %w", err)
		}
		relationshipTypes = append(relationshipTypes, rt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate relationship types: %w", err)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &RelationshipTypeListResponse{
		RelationshipTypes: relationshipTypes,
		Page:              page,
		Limit:             limit,
		Total:             total,
		TotalPages:        totalPages,
	}, nil
}

// GetActive retrieves all active relationship types
func (r *RelationshipTypeRepository) GetActive(ctx context.Context) ([]RelationshipType, error) {
	query := `
		SELECT id, name, display_name, description, forward_label, reverse_label,
			   color, icon, category, is_active, is_system, allowed_source_types,
			   allowed_target_types, cardinality_source, cardinality_target,
			   bidirectional, attributes, created_at, updated_at, created_by, updated_by
		FROM relationship_types
		WHERE is_active = true
		ORDER BY category, name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active relationship types: %w", err)
	}
	defer rows.Close()

	var relationshipTypes []RelationshipType
	for rows.Next() {
		rt := RelationshipType{}
		err := rows.Scan(
			&rt.ID, &rt.Name, &rt.DisplayName, &rt.Description, &rt.ForwardLabel, &rt.ReverseLabel,
			&rt.Color, &rt.Icon, &rt.Category, &rt.IsActive, &rt.IsSystem, &rt.AllowedSourceTypes,
			&rt.AllowedTargetTypes, &rt.CardinalitySource, &rt.CardinalityTarget,
			&rt.Bidirectional, &rt.Attributes, &rt.CreatedAt, &rt.UpdatedAt, &rt.CreatedBy, &rt.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relationship type: %w", err)
		}
		relationshipTypes = append(relationshipTypes, rt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate relationship types: %w", err)
	}

	return relationshipTypes, nil
}

// Update updates an existing relationship type
func (r *RelationshipTypeRepository) Update(ctx context.Context, rt *RelationshipType) (*RelationshipType, error) {
	query := `
		UPDATE relationship_types
		SET display_name = $2, description = $3, forward_label = $4, reverse_label = $5,
			color = $6, icon = $7, category = $8, is_active = $9,
			allowed_source_types = $10, allowed_target_types = $11,
			cardinality_source = $12, cardinality_target = $13,
			bidirectional = $14, attributes = $15, updated_at = $16, updated_by = $17
		WHERE id = $1
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		rt.ID, rt.DisplayName, rt.Description, rt.ForwardLabel, rt.ReverseLabel,
		rt.Color, rt.Icon, rt.Category, rt.IsActive,
		rt.AllowedSourceTypes, rt.AllowedTargetTypes,
		rt.CardinalitySource, rt.CardinalityTarget,
		rt.Bidirectional, rt.Attributes, time.Now(), rt.UpdatedBy,
	).Scan(&rt.ID, &rt.CreatedAt, &rt.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("relationship type not found")
		}
		return nil, fmt.Errorf("failed to update relationship type: %w", err)
	}

	return rt, nil
}

// Delete deletes a relationship type
func (r *RelationshipTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM relationship_types WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete relationship type: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("relationship type not found")
	}

	return nil
}

// GetUsageCount returns the number of relationships using this type
func (r *RelationshipTypeRepository) GetUsageCount(ctx context.Context, id uuid.UUID) (int64, error) {
	// This would typically query the relationships table in Neo4j or PostgreSQL
	// For now, we'll return 0 as a placeholder
	// In a full implementation, you'd query the actual relationships table

	// Example query if relationships were stored in PostgreSQL:
	// query := `SELECT COUNT(*) FROM relationships WHERE relationship_type_id = $1`
	// var count int64
	// err := r.db.QueryRow(ctx, query, id).Scan(&count)
	// return count, err

	return 0, nil
}

// GetUsage returns usage statistics for all relationship types
func (r *RelationshipTypeRepository) GetUsage(ctx context.Context) ([]RelationshipTypeUsage, error) {
	// This would join relationship_types with relationships table
	// For now, return empty slice as placeholder
	// In a full implementation, you'd query actual usage data

	return []RelationshipTypeUsage{}, nil
}

// GetStatistics returns comprehensive statistics about relationship types
func (r *RelationshipTypeRepository) GetStatistics(ctx context.Context) (*RelationshipTypeStatistics, error) {
	stats := &RelationshipTypeStatistics{}

	// Get total types
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM relationship_types").Scan(&stats.TotalTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to get total types: %w", err)
	}

	// Get active types
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM relationship_types WHERE is_active = true").Scan(&stats.ActiveTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to get active types: %w", err)
	}

	// Get system types
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM relationship_types WHERE is_system = true").Scan(&stats.SystemTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to get system types: %w", err)
	}

	// Get custom types
	stats.CustomTypes = stats.TotalTypes - stats.SystemTypes

	// Get total relationships (placeholder - would come from actual relationships table)
	stats.TotalRelationships = 0

	// Get categories distribution
	categoriesQuery := `
		SELECT category, COUNT(*)
		FROM relationship_types
		WHERE category IS NOT NULL
		GROUP BY category
	`
	rows, err := r.db.Query(ctx, categoriesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories distribution: %w", err)
	}
	defer rows.Close()

	stats.Categories = make(map[string]int64)
	for rows.Next() {
		var category string
		var count int64
		if err := rows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		stats.Categories[category] = count
	}

	// Get cardinality distribution
	cardinalityQuery := `
		SELECT cardinality_source || '-to-' || cardinality_source as cardinality, COUNT(*)
		FROM relationship_types
		GROUP BY cardinality_source, cardinality_target
	`
	rows, err = r.db.Query(ctx, cardinalityQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get cardinality distribution: %w", err)
	}
	defer rows.Close()

	stats.CardinalityDistribution = make(map[string]int64)
	for rows.Next() {
		var cardinality string
		var count int64
		if err := rows.Scan(&cardinality, &count); err != nil {
			return nil, fmt.Errorf("failed to scan cardinality: %w", err)
		}
		stats.CardinalityDistribution[cardinality] = count
	}

	return stats, nil
}