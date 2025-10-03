package ci

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

type Repository struct {
	db     *pgxpool.Pool
	logger *pustakaLogger.Logger
}

func NewRepository(db *pgxpool.Pool, logger *pustakaLogger.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// Configuration Item operations

func (r *Repository) CreateCI(ctx context.Context, ci *ConfigurationItem) (*ConfigurationItem, error) {
	query := `
		INSERT INTO configuration_items (id, name, ci_type, attributes, tags, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, ci_type, attributes, tags, created_at, updated_at, created_by, updated_by
	`

	if ci.ID == uuid.Nil {
		ci.ID = uuid.New()
	}

	now := time.Now()
	var result ConfigurationItem
	err := r.db.QueryRow(ctx, query,
		ci.ID,
		ci.Name,
		ci.CIType,
		ci.Attributes,
		ci.Tags,
		ci.CreatedBy,
		now,
		now,
	).Scan(
		&result.ID,
		&result.Name,
		&result.CIType,
		&result.Attributes,
		&result.Tags,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)

	if err != nil {
		r.logger.ErrorDatabase("INSERT", "configuration_items", err, map[string]interface{}{
			"ci_name": ci.Name,
			"ci_type": ci.CIType,
		})
		return nil, fmt.Errorf("failed to create CI: %w", err)
	}

	r.logger.InfoDatabase("INSERT", "configuration_items", 0, map[string]interface{}{
		"ci_id": result.ID,
		"ci_name": result.Name,
	})

	return &result, nil
}

func (r *Repository) GetCI(ctx context.Context, id uuid.UUID) (*ConfigurationItem, error) {
	query := `
		SELECT id, name, ci_type, attributes, tags, created_at, updated_at, created_by, updated_by
		FROM configuration_items
		WHERE id = $1
	`

	var ci ConfigurationItem
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ci.ID,
		&ci.Name,
		&ci.CIType,
		&ci.Attributes,
		&ci.Tags,
		&ci.CreatedAt,
		&ci.UpdatedAt,
		&ci.CreatedBy,
		&ci.UpdatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("CI not found")
		}
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, map[string]interface{}{
			"ci_id": id,
		})
		return nil, fmt.Errorf("failed to get CI: %w", err)
	}

	return &ci, nil
}

func (r *Repository) ListCIs(ctx context.Context, filters ListCIFilters, page, limit int) (*CIListResponse, error) {
	offset := (page - 1) * limit

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filters.CIType != "" {
		whereClause += fmt.Sprintf(" AND ci_type = $%d", argIndex)
		args = append(args, filters.CIType)
		argIndex++
	}

	if filters.Search != "" {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR attributes::text ILIKE $%d)", argIndex, argIndex+1)
		args = append(args, "%"+filters.Search+"%", "%"+filters.Search+"%")
		argIndex += 2
	}

	if len(filters.Tags) > 0 {
		whereClause += fmt.Sprintf(" AND tags && $%d", argIndex)
		args = append(args, filters.Tags)
		argIndex++
	}

	if filters.CreatedBy != "" {
		whereClause += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, filters.CreatedBy)
		argIndex++
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if filters.Sort != "" {
		orderField := filters.Sort
		if orderField == "name" {
			orderField = "name"
		} else if orderField == "type" {
			orderField = "ci_type"
		} else if orderField == "updated_at" {
			orderField = "updated_at"
		} else {
			orderField = "created_at"
		}

		orderDirection := "DESC"
		if filters.Order == "asc" {
			orderDirection = "ASC"
		}
		orderBy = fmt.Sprintf("ORDER BY %s %s", orderField, orderDirection)
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM configuration_items %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, fmt.Errorf("failed to count CIs: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, name, ci_type, attributes, tags, created_at, updated_at, created_by, updated_by
		FROM configuration_items %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, fmt.Errorf("failed to list CIs: %w", err)
	}
	defer rows.Close()

	var cis []ConfigurationItem
	for rows.Next() {
		var ci ConfigurationItem
		err := rows.Scan(
			&ci.ID,
			&ci.Name,
			&ci.CIType,
			&ci.Attributes,
			&ci.Tags,
			&ci.CreatedAt,
			&ci.UpdatedAt,
			&ci.CreatedBy,
			&ci.UpdatedBy,
		)
		if err != nil {
			r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
			return nil, fmt.Errorf("failed to scan CI: %w", err)
		}
		cis = append(cis, ci)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &CIListResponse{
		CIs:        cis,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (r *Repository) UpdateCI(ctx context.Context, id uuid.UUID, updates *UpdateCIRequest, updatedBy uuid.UUID) (*ConfigurationItem, error) {
	// Get current CI for audit
	current, err := r.GetCI(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build UPDATE query
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.Attributes != nil {
		setClauses = append(setClauses, fmt.Sprintf("attributes = $%d", argIndex))
		args = append(args, updates.Attributes)
		argIndex++
	}

	if updates.Tags != nil {
		setClauses = append(setClauses, fmt.Sprintf("tags = $%d", argIndex))
		args = append(args, updates.Tags)
		argIndex++
	}

	if len(setClauses) == 0 {
		return current, nil
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	setClauses = append(setClauses, fmt.Sprintf("updated_by = $%d", argIndex))
	args = append(args, updatedBy)
	argIndex++

	setClause := "SET " + setClauses[0]
	for i := 1; i < len(setClauses); i++ {
		setClause += ", " + setClauses[i]
	}

	query := fmt.Sprintf("UPDATE configuration_items %s WHERE id = $%d RETURNING id, name, ci_type, attributes, tags, created_at, updated_at, created_by, updated_by", setClause, argIndex)
	args = append(args, id)

	var result ConfigurationItem
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&result.ID,
		&result.Name,
		&result.CIType,
		&result.Attributes,
		&result.Tags,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)

	if err != nil {
		r.logger.ErrorDatabase("UPDATE", "configuration_items", err, map[string]interface{}{
			"ci_id": id,
		})
		return nil, fmt.Errorf("failed to update CI: %w", err)
	}

	r.logger.InfoDatabase("UPDATE", "configuration_items", 0, map[string]interface{}{
		"ci_id": id,
		"updated_by": updatedBy,
	})

	return &result, nil
}

func (r *Repository) DeleteCI(ctx context.Context, id uuid.UUID) error {
	// Check for existing relationships
	var relationshipCount int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM relationships WHERE source_id = $1 OR target_id = $1", id).Scan(&relationshipCount)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "relationships", err, map[string]interface{}{
			"ci_id": id,
		})
		return fmt.Errorf("failed to check relationships: %w", err)
	}

	if relationshipCount > 0 {
		return fmt.Errorf("cannot delete CI with existing relationships")
	}

	query := "DELETE FROM configuration_items WHERE id = $1"
	_, err = r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorDatabase("DELETE", "configuration_items", err, map[string]interface{}{
			"ci_id": id,
		})
		return fmt.Errorf("failed to delete CI: %w", err)
	}

	r.logger.InfoDatabase("DELETE", "configuration_items", 0, map[string]interface{}{
		"ci_id": id,
	})

	return nil
}

// CI Type operations

func (r *Repository) CreateCIType(ctx context.Context, ciType *CITypeDefinition) (*CITypeDefinition, error) {
	query := `
		INSERT INTO ci_type_definitions (id, name, description, required_attributes, optional_attributes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, description, required_attributes, optional_attributes, created_by, created_at, updated_at
	`

	if ciType.ID == uuid.Nil {
		ciType.ID = uuid.New()
	}

	now := time.Now()
	var result CITypeDefinition
	err := r.db.QueryRow(ctx, query,
		ciType.ID,
		ciType.Name,
		ciType.Description,
		ciType.RequiredAttributes,
		ciType.OptionalAttributes,
		ciType.CreatedBy,
		now,
		now,
	).Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.RequiredAttributes,
		&result.OptionalAttributes,
		&result.CreatedBy,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		r.logger.ErrorDatabase("INSERT", "ci_type_definitions", err, map[string]interface{}{
			"ci_type_name": ciType.Name,
		})
		return nil, fmt.Errorf("failed to create CI type: %w", err)
	}

	return &result, nil
}

func (r *Repository) GetCIType(ctx context.Context, id uuid.UUID) (*CITypeDefinition, error) {
	query := `
		SELECT id, name, description, required_attributes, optional_attributes, created_by, created_at, updated_at
		FROM ci_type_definitions
		WHERE id = $1
	`

	var ciType CITypeDefinition
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ciType.ID,
		&ciType.Name,
		&ciType.Description,
		&ciType.RequiredAttributes,
		&ciType.OptionalAttributes,
		&ciType.CreatedBy,
		&ciType.CreatedAt,
		&ciType.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("CI type not found")
		}
		r.logger.ErrorDatabase("SELECT", "ci_type_definitions", err, map[string]interface{}{
			"ci_type_id": id,
		})
		return nil, fmt.Errorf("failed to get CI type: %w", err)
	}

	return &ciType, nil
}

func (r *Repository) GetCITypeByName(ctx context.Context, name string) (*CITypeDefinition, error) {
	query := `
		SELECT id, name, description, required_attributes, optional_attributes, created_by, created_at, updated_at
		FROM ci_type_definitions
		WHERE name = $1
	`

	var ciType CITypeDefinition
	err := r.db.QueryRow(ctx, query, name).Scan(
		&ciType.ID,
		&ciType.Name,
		&ciType.Description,
		&ciType.RequiredAttributes,
		&ciType.OptionalAttributes,
		&ciType.CreatedBy,
		&ciType.CreatedAt,
		&ciType.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("CI type not found")
		}
		r.logger.ErrorDatabase("SELECT", "ci_type_definitions", err, map[string]interface{}{
			"ci_type_name": name,
		})
		return nil, fmt.Errorf("failed to get CI type: %w", err)
	}

	return &ciType, nil
}

func (r *Repository) ListCITypes(ctx context.Context, page, limit int, search string) (*CITypeListResponse, error) {
	offset := (page - 1) * limit

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex+1)
		args = append(args, "%"+search+"%", "%"+search+"%")
		argIndex += 2
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ci_type_definitions %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "ci_type_definitions", err, nil)
		return nil, fmt.Errorf("failed to count CI types: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, name, description, required_attributes, optional_attributes, created_by, created_at, updated_at
		FROM ci_type_definitions %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "ci_type_definitions", err, nil)
		return nil, fmt.Errorf("failed to list CI types: %w", err)
	}
	defer rows.Close()

	var ciTypes []CITypeDefinition
	for rows.Next() {
		var ciType CITypeDefinition
		err := rows.Scan(
			&ciType.ID,
			&ciType.Name,
			&ciType.Description,
			&ciType.RequiredAttributes,
			&ciType.OptionalAttributes,
			&ciType.CreatedBy,
			&ciType.CreatedAt,
			&ciType.UpdatedAt,
		)
		if err != nil {
			r.logger.ErrorDatabase("SELECT", "ci_type_definitions", err, nil)
			return nil, fmt.Errorf("failed to scan CI type: %w", err)
		}
		ciTypes = append(ciTypes, ciType)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &CITypeListResponse{
		CITypes:    ciTypes,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (r *Repository) UpdateCIType(ctx context.Context, id uuid.UUID, updates *UpdateCITypeRequest) (*CITypeDefinition, error) {
	// Build UPDATE query
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, updates.Description)
		argIndex++
	}

	if updates.RequiredAttributes != nil {
		setClauses = append(setClauses, fmt.Sprintf("required_attributes = $%d", argIndex))
		args = append(args, updates.RequiredAttributes)
		argIndex++
	}

	if updates.OptionalAttributes != nil {
		setClauses = append(setClauses, fmt.Sprintf("optional_attributes = $%d", argIndex))
		args = append(args, updates.OptionalAttributes)
		argIndex++
	}

	if len(setClauses) == 0 {
		return r.GetCIType(ctx, id)
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	setClause := "SET " + setClauses[0]
	for i := 1; i < len(setClauses); i++ {
		setClause += ", " + setClauses[i]
	}

	query := fmt.Sprintf("UPDATE ci_type_definitions %s WHERE id = $%d RETURNING id, name, description, required_attributes, optional_attributes, created_by, created_at, updated_at", setClause, argIndex)
	args = append(args, id)

	var result CITypeDefinition
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.RequiredAttributes,
		&result.OptionalAttributes,
		&result.CreatedBy,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		r.logger.ErrorDatabase("UPDATE", "ci_type_definitions", err, map[string]interface{}{
			"ci_type_id": id,
		})
		return nil, fmt.Errorf("failed to update CI type: %w", err)
	}

	return &result, nil
}

func (r *Repository) DeleteCIType(ctx context.Context, id uuid.UUID) error {
	// Check for existing CIs of this type
	var ciCount int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM configuration_items WHERE ci_type = (SELECT name FROM ci_type_definitions WHERE id = $1)", id).Scan(&ciCount)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, map[string]interface{}{
			"ci_type_id": id,
		})
		return fmt.Errorf("failed to check CIs: %w", err)
	}

	if ciCount > 0 {
		return fmt.Errorf("cannot delete CI type with existing CIs")
	}

	query := "DELETE FROM ci_type_definitions WHERE id = $1"
	_, err = r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorDatabase("DELETE", "ci_type_definitions", err, map[string]interface{}{
			"ci_type_id": id,
		})
		return fmt.Errorf("failed to delete CI type: %w", err)
	}

	return nil
}

// Relationship operations

func (r *Repository) GetCIByNameAndType(ctx context.Context, name, ciType string) (*ConfigurationItem, error) {
	query := `
		SELECT id, name, ci_type, attributes, tags, created_at, updated_at, created_by, updated_by
		FROM configuration_items
		WHERE name = $1 AND ci_type = $2
	`

	var ci ConfigurationItem
	err := r.db.QueryRow(ctx, query, name, ciType).Scan(
		&ci.ID,
		&ci.Name,
		&ci.CIType,
		&ci.Attributes,
		&ci.Tags,
		&ci.CreatedAt,
		&ci.UpdatedAt,
		&ci.CreatedBy,
		&ci.UpdatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, map[string]interface{}{
			"ci_name": name,
			"ci_type": ciType,
		})
		return nil, fmt.Errorf("failed to get CI by name and type: %w", err)
	}

	return &ci, nil
}

func (r *Repository) CreateRelationship(ctx context.Context, rel *Relationship) (*Relationship, error) {
	query := `
		INSERT INTO relationships (id, source_id, target_id, relationship_type, attributes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, source_id, target_id, relationship_type, attributes, created_at, updated_at, created_by, updated_by
	`

	if rel.ID == uuid.Nil {
		rel.ID = uuid.New()
	}

	now := time.Now()
	var result Relationship
	err := r.db.QueryRow(ctx, query,
		rel.ID,
		rel.SourceID,
		rel.TargetID,
		rel.RelationshipType,
		rel.Attributes,
		rel.CreatedBy,
		now,
		now,
	).Scan(
		&result.ID,
		&result.SourceID,
		&result.TargetID,
		&result.RelationshipType,
		&result.Attributes,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)

	if err != nil {
		r.logger.ErrorDatabase("INSERT", "relationships", err, map[string]interface{}{
			"relationship_id": rel.ID,
			"source_id":       rel.SourceID,
			"target_id":       rel.TargetID,
		})
		return nil, fmt.Errorf("failed to create relationship: %w", err)
	}

	r.logger.InfoDatabase("INSERT", "relationships", 0, map[string]interface{}{
		"relationship_id": result.ID,
	})

	return &result, nil
}

func (r *Repository) GetRelationship(ctx context.Context, id uuid.UUID) (*Relationship, error) {
	query := `
		SELECT id, source_id, target_id, relationship_type, attributes, created_at, updated_at, created_by, updated_by
		FROM relationships
		WHERE id = $1
	`

	var rel Relationship
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rel.ID,
		&rel.SourceID,
		&rel.TargetID,
		&rel.RelationshipType,
		&rel.Attributes,
		&rel.CreatedAt,
		&rel.UpdatedAt,
		&rel.CreatedBy,
		&rel.UpdatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("relationship not found")
		}
		r.logger.ErrorDatabase("SELECT", "relationships", err, map[string]interface{}{
			"relationship_id": id,
		})
		return nil, fmt.Errorf("failed to get relationship: %w", err)
	}

	return &rel, nil
}

func (r *Repository) ListRelationships(ctx context.Context, filters ListRelationshipFilters, page, limit int) (*RelationshipListResponse, error) {
	offset := (page - 1) * limit

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filters.SourceID != nil {
		whereClause += fmt.Sprintf(" AND source_id = $%d", argIndex)
		args = append(args, filters.SourceID)
		argIndex++
	}

	if filters.TargetID != nil {
		whereClause += fmt.Sprintf(" AND target_id = $%d", argIndex)
		args = append(args, filters.TargetID)
		argIndex++
	}

	if filters.RelationshipType != "" {
		whereClause += fmt.Sprintf(" AND relationship_type = $%d", argIndex)
		args = append(args, filters.RelationshipType)
		argIndex++
	}

	if filters.Search != "" {
		whereClause += fmt.Sprintf(" AND (id::text ILIKE $%d OR relationship_type ILIKE $%d OR attributes::text ILIKE $%d)", argIndex, argIndex+1, argIndex+2)
		args = append(args, "%"+filters.Search+"%", "%"+filters.Search+"%", "%"+filters.Search+"%")
		argIndex += 3
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM relationships %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "relationships", err, nil)
		return nil, fmt.Errorf("failed to count relationships: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, source_id, target_id, relationship_type, attributes, created_at, updated_at, created_by, updated_by
		FROM relationships %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "relationships", err, nil)
		return nil, fmt.Errorf("failed to list relationships: %w", err)
	}
	defer rows.Close()

	var relationships []Relationship
	for rows.Next() {
		var rel Relationship
		err := rows.Scan(
			&rel.ID,
			&rel.SourceID,
			&rel.TargetID,
			&rel.RelationshipType,
			&rel.Attributes,
			&rel.CreatedAt,
			&rel.UpdatedAt,
			&rel.CreatedBy,
			&rel.UpdatedBy,
		)
		if err != nil {
			r.logger.ErrorDatabase("SELECT", "relationships", err, nil)
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		relationships = append(relationships, rel)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &RelationshipListResponse{
		Relationships: relationships,
		Page:          page,
		Limit:         limit,
		Total:         total,
		TotalPages:    totalPages,
	}, nil
}

func (r *Repository) UpdateRelationship(ctx context.Context, id uuid.UUID, updates *UpdateRelationshipRequest, updatedBy uuid.UUID) (*Relationship, error) {
	// Get current relationship for audit
	current, err := r.GetRelationship(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build UPDATE query
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	// Relationship type cannot be updated - create new relationship instead
	// if updates.RelationshipType != nil {
	//	setClauses = append(setClauses, fmt.Sprintf("relationship_type = $%d", argIndex))
	//	args = append(args, updates.RelationshipType)
	//	argIndex++
	// }

	if updates.Attributes != nil {
		setClauses = append(setClauses, fmt.Sprintf("attributes = $%d", argIndex))
		args = append(args, updates.Attributes)
		argIndex++
	}

	if len(setClauses) == 0 {
		return current, nil
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	setClauses = append(setClauses, fmt.Sprintf("updated_by = $%d", argIndex))
	args = append(args, updatedBy)
	argIndex++

	setClause := "SET " + setClauses[0]
	for i := 1; i < len(setClauses); i++ {
		setClause += ", " + setClauses[i]
	}

	query := fmt.Sprintf("UPDATE relationships %s WHERE id = $%d RETURNING id, source_id, target_id, relationship_type, attributes, created_at, updated_at, created_by, updated_by", setClause, argIndex)
	args = append(args, id)

	var result Relationship
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&result.ID,
		&result.SourceID,
		&result.TargetID,
		&result.RelationshipType,
		&result.Attributes,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.CreatedBy,
		&result.UpdatedBy,
	)

	if err != nil {
		r.logger.ErrorDatabase("UPDATE", "relationships", err, map[string]interface{}{
			"relationship_id": id,
		})
		return nil, fmt.Errorf("failed to update relationship: %w", err)
	}

	r.logger.InfoDatabase("UPDATE", "relationships", 0, map[string]interface{}{
		"relationship_id": id,
		"updated_by":      updatedBy,
	})

	return &result, nil
}

func (r *Repository) DeleteRelationship(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM relationships WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorDatabase("DELETE", "relationships", err, map[string]interface{}{
			"relationship_id": id,
		})
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	r.logger.InfoDatabase("DELETE", "relationships", 0, map[string]interface{}{
		"relationship_id": id,
	})

	return nil
}

// Count methods for dashboard statistics

func (r *Repository) CountCIs(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM configuration_items"

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, fmt.Errorf("failed to count CIs: %w", err)
	}

	return count, nil
}

func (r *Repository) CountCITypes(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM ci_type_definitions"

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "ci_type_definitions", err, nil)
		return 0, fmt.Errorf("failed to count CI types: %w", err)
	}

	return count, nil
}

func (r *Repository) CountRelationships(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM relationships"

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "relationships", err, nil)
		return 0, fmt.Errorf("failed to count relationships: %w", err)
	}

	return count, nil
}

func (r *Repository) CountUsers(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM users"

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "users", err, nil)
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}