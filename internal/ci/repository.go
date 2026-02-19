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

// Helper function to convert *string to string
func getStringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Configuration Item operations

func (r *Repository) CreateCI(ctx context.Context, ci *ConfigurationItem) (*ConfigurationItem, error) {
	query := `
		INSERT INTO configuration_items (id, name, ci_type, attributes, tags, lifecycle_status_id, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, ci_type, attributes, tags, lifecycle_status_id, created_at, updated_at, created_by, updated_by
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
		ci.LifecycleStatusID,
		ci.CreatedBy,
		now,
		now,
	).Scan(
		&result.ID,
		&result.Name,
		&result.CIType,
		&result.Attributes,
		&result.Tags,
		&result.LifecycleStatusID,
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
		SELECT ci.id, ci.name, ci.ci_type, ci.attributes, ci.tags, ci.lifecycle_status_id,
		       ls.name, ls.display_name, ls.description, ls.color, ls.icon,
		       ci.created_at, ci.updated_at, ci.created_by, ci.updated_by
		FROM configuration_items ci
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		WHERE ci.id = $1
	`

	var ci ConfigurationItem
	var lifecycleStatusName, lifecycleStatusDisplayName, lifecycleStatusDescription, lifecycleStatusColor, lifecycleStatusIcon *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ci.ID,
		&ci.Name,
		&ci.CIType,
		&ci.Attributes,
		&ci.Tags,
		&ci.LifecycleStatusID,
		&lifecycleStatusName,
		&lifecycleStatusDisplayName,
		&lifecycleStatusDescription,
		&lifecycleStatusColor,
		&lifecycleStatusIcon,
		&ci.CreatedAt,
		&ci.UpdatedAt,
		&ci.CreatedBy,
		&ci.UpdatedBy,
	)

	// Build lifecycle status object if present
	if ci.LifecycleStatusID != nil {
		ci.LifecycleStatus = &LifecycleStatus{
			ID:          *ci.LifecycleStatusID,
			Name:        getStringOrEmpty(lifecycleStatusName),
			DisplayName: getStringOrEmpty(lifecycleStatusDisplayName),
			Description: lifecycleStatusDescription,
			Color:       lifecycleStatusColor,
			Icon:        lifecycleStatusIcon,
		}
	}

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
		whereClause += fmt.Sprintf(" AND (ci.name ILIKE $%d OR ci.attributes::text ILIKE $%d)", argIndex, argIndex+1)
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

	if filters.LifecycleStatusID != nil {
		whereClause += fmt.Sprintf(" AND lifecycle_status_id = $%d", argIndex)
		args = append(args, filters.LifecycleStatusID)
		argIndex++
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if filters.Sort != "" {
		orderField := filters.Sort
		if orderField == "name" {
			orderField = "ci.name"  // Use table alias to avoid ambiguity with lifecycle_status.name
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM configuration_items ci %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, fmt.Errorf("failed to count CIs: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT ci.id, ci.name, ci.ci_type, ci.attributes, ci.tags, ci.lifecycle_status_id,
		       ls.name, ls.display_name, ls.description, ls.color, ls.icon,
		       ci.created_at, ci.updated_at, ci.created_by, ci.updated_by
		FROM configuration_items ci
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		%s %s
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
		var lifecycleStatusName, lifecycleStatusDisplayName, lifecycleStatusDescription, lifecycleStatusColor, lifecycleStatusIcon *string
		err := rows.Scan(
			&ci.ID,
			&ci.Name,
			&ci.CIType,
			&ci.Attributes,
			&ci.Tags,
			&ci.LifecycleStatusID,
			&lifecycleStatusName,
			&lifecycleStatusDisplayName,
			&lifecycleStatusDescription,
			&lifecycleStatusColor,
			&lifecycleStatusIcon,
			&ci.CreatedAt,
			&ci.UpdatedAt,
			&ci.CreatedBy,
			&ci.UpdatedBy,
		)

		// Build lifecycle status object if present
		if ci.LifecycleStatusID != nil {
			ci.LifecycleStatus = &LifecycleStatus{
				ID:          *ci.LifecycleStatusID,
				Name:        getStringOrEmpty(lifecycleStatusName),
				DisplayName: getStringOrEmpty(lifecycleStatusDisplayName),
				Description: lifecycleStatusDescription,
				Color:       lifecycleStatusColor,
				Icon:        lifecycleStatusIcon,
			}
		}
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

	if updates.LifecycleStatusID != nil {
		setClauses = append(setClauses, fmt.Sprintf("lifecycle_status_id = $%d", argIndex))
		args = append(args, updates.LifecycleStatusID)
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

	query := fmt.Sprintf("UPDATE configuration_items %s WHERE id = $%d RETURNING id, name, ci_type, attributes, tags, lifecycle_status_id, created_at, updated_at, created_by, updated_by", setClause, argIndex)
	args = append(args, id)

	var result ConfigurationItem
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&result.ID,
		&result.Name,
		&result.CIType,
		&result.Attributes,
		&result.Tags,
		&result.LifecycleStatusID,
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
		INSERT INTO ci_type_definitions (id, name, description, is_amortizable, required_attributes, optional_attributes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, description, is_amortizable, required_attributes, optional_attributes, created_by, created_at, updated_at
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
		ciType.IsAmortizable,
		ciType.RequiredAttributes,
		ciType.OptionalAttributes,
		ciType.CreatedBy,
		now,
		now,
	).Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.IsAmortizable,
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
		SELECT id, name, description, is_amortizable, required_attributes, optional_attributes, created_by, created_at, updated_at
		FROM ci_type_definitions
		WHERE id = $1
	`

	var ciType CITypeDefinition
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ciType.ID,
		&ciType.Name,
		&ciType.Description,
		&ciType.IsAmortizable,
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
		SELECT id, name, description, is_amortizable, required_attributes, optional_attributes, created_by, created_at, updated_at
		FROM ci_type_definitions
		WHERE name = $1
	`

	var ciType CITypeDefinition
	err := r.db.QueryRow(ctx, query, name).Scan(
		&ciType.ID,
		&ciType.Name,
		&ciType.Description,
		&ciType.IsAmortizable,
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
		SELECT id, name, description, is_amortizable, required_attributes, optional_attributes, created_by, created_at, updated_at
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
			&ciType.IsAmortizable,
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

	if updates.IsAmortizable != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_amortizable = $%d", argIndex))
		args = append(args, updates.IsAmortizable)
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

	query := fmt.Sprintf("UPDATE ci_type_definitions %s WHERE id = $%d RETURNING id, name, description, is_amortizable, required_attributes, optional_attributes, created_by, created_at, updated_at", setClause, argIndex)
	args = append(args, id)

	var result CITypeDefinition
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.IsAmortizable,
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

func (r *Repository) GetCICreationsByDate(ctx context.Context, fromDate, toDate string) ([]DailyCount, error) {
	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM configuration_items
		WHERE created_at >= $1::date AND created_at < ($2::date + INTERVAL '1 day')
		GROUP BY DATE(created_at)
		ORDER BY date
	`

	rows, err := r.db.Query(ctx, query, fromDate, toDate)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, map[string]interface{}{
			"from_date": fromDate,
			"to_date":   toDate,
		})
		return nil, fmt.Errorf("failed to get CI creations by date: %w", err)
	}
	defer rows.Close()

	var results []DailyCount
	for rows.Next() {
		var date time.Time
		var count int64
		if err := rows.Scan(&date, &count); err != nil {
			r.logger.ErrorDatabase("SCAN", "configuration_items", err, nil)
			return nil, fmt.Errorf("failed to scan CI creation data: %w", err)
		}
		results = append(results, DailyCount{
			Date:  date.Format("2006-01-02"),
			Count: count,
		})
	}

	if err := rows.Err(); err != nil {
		r.logger.ErrorDatabase("ROWS", "configuration_items", err, nil)
		return nil, fmt.Errorf("error iterating CI creation data: %w", err)
	}

	return results, nil
}

// CountCIsWithStatus counts how many CIs have a specific lifecycle status
func (r *Repository) CountCIsWithStatus(ctx context.Context, statusID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM configuration_items WHERE lifecycle_status_id = $1`

	var count int64
	err := r.db.QueryRow(ctx, query, statusID).Scan(&count)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, map[string]interface{}{
			"status_id": statusID,
		})
		return 0, fmt.Errorf("failed to count CIs with lifecycle status: %w", err)
	}

	return count, nil
}

// Health score calculation functions

// GetHealthScoreMetrics retrieves metrics needed to calculate health scores
func (r *Repository) GetHealthScoreMetrics(ctx context.Context) (*HealthScoreMetrics, error) {
	// Get total CIs first
	totalCIs, err := r.CountCIs(ctx)
	if err != nil {
		return nil, err
	}

	// Get CIs updated in last 90 days
	currentQuery := `
		SELECT COUNT(*) FROM configuration_items
		WHERE updated_at > NOW() - INTERVAL '90 days'
	`
	var currentCIs int
	err = r.db.QueryRow(ctx, currentQuery).Scan(&currentCIs)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, fmt.Errorf("failed to get current CIs: %w", err)
	}

	// Get compliant CIs (following naming convention)
	compliantQuery := `
		SELECT COUNT(*) FROM configuration_items
		WHERE name ~ '^[a-z][a-z0-9-]*$'
	`
	var compliantCIs int
	err = r.db.QueryRow(ctx, compliantQuery).Scan(&compliantCIs)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, fmt.Errorf("failed to get compliant CIs: %w", err)
	}

	// For completeness, we need to check if all required attributes are present
	// This is complex, so for now we'll count CIs that have non-empty attributes
	completeQuery := `
		SELECT COUNT(*) FROM configuration_items
		WHERE attributes != '{}'::jsonb
	`
	var completeCIs int
	err = r.db.QueryRow(ctx, completeQuery).Scan(&completeCIs)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, fmt.Errorf("failed to get complete CIs: %w", err)
	}

	return &HealthScoreMetrics{
		TotalCIs:     int(totalCIs),
		CompleteCIs:  completeCIs,
		CurrentCIs:   currentCIs,
		CompliantCIs: compliantCIs,
	}, nil
}

// SaveHealthScore saves a health score snapshot to history
func (r *Repository) SaveHealthScore(ctx context.Context, score *HealthScore, metrics *HealthScoreMetrics) error {
	query := `
		INSERT INTO health_score_history (
			calculated_at, overall_score, completeness_score, correctness_score, compliance_score,
			total_cis, complete_cis, current_cis, compliant_cis
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		score.CalculatedAt,
		score.Overall,
		score.Completeness,
		score.Correctness,
		score.Compliance,
		metrics.TotalCIs,
		metrics.CompleteCIs,
		metrics.CurrentCIs,
		metrics.CompliantCIs,
	)

	if err != nil {
		r.logger.ErrorDatabase("INSERT", "health_score_history", err, nil)
		return fmt.Errorf("failed to save health score: %w", err)
	}

	r.logger.InfoDatabase("INSERT", "health_score_history", 0, map[string]interface{}{
		"overall_score": score.Overall,
	})

	return nil
}

// GetHealthScoreHistory retrieves health score history for trend calculation
func (r *Repository) GetHealthScoreHistory(ctx context.Context, days int) ([]HealthScoreHistory, error) {
	query := `
		SELECT id, calculated_at, overall_score, completeness_score, correctness_score, compliance_score,
		       total_cis, complete_cis, current_cis, compliant_cis, created_at
		FROM health_score_history
		WHERE calculated_at > NOW() - INTERVAL '1 day' * $1
		ORDER BY calculated_at DESC
	`

	rows, err := r.db.Query(ctx, query, days)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "health_score_history", err, map[string]interface{}{
			"days": days,
		})
		return nil, fmt.Errorf("failed to get health score history: %w", err)
	}
	defer rows.Close()

	var history []HealthScoreHistory
	for rows.Next() {
		var record HealthScoreHistory
		err := rows.Scan(
			&record.ID,
			&record.CalculatedAt,
			&record.OverallScore,
			&record.CompletenessScore,
			&record.CorrectnessScore,
			&record.ComplianceScore,
			&record.TotalCIs,
			&record.CompleteCIs,
			&record.CurrentCIs,
			&record.CompliantCIs,
			&record.CreatedAt,
		)
		if err != nil {
			r.logger.ErrorDatabase("SCAN", "health_score_history", err, nil)
			return nil, fmt.Errorf("failed to scan health score history: %w", err)
		}
		history = append(history, record)
	}

	if err := rows.Err(); err != nil {
		r.logger.ErrorDatabase("ROWS", "health_score_history", err, nil)
		return nil, fmt.Errorf("error iterating health score history: %w", err)
	}

	return history, nil
}

// Data Quality Metrics functions

// GetDataQualityMetrics retrieves all data quality metrics
func (r *Repository) GetDataQualityMetrics(ctx context.Context, includeDetails bool, limit int) (*DataQualityMetrics, error) {
	totalCIs, err := r.CountCIs(ctx)
	if err != nil {
		return nil, err
	}

	if totalCIs == 0 {
		return &DataQualityMetrics{}, nil
	}

	// Get metrics concurrently
	type result struct {
		name   string
		count  int
		stale30 int
		stale60 int
		stale90 int
		err    error
	}
	results := make(chan result, 7)

	// Missing attributes
	go func() {
		count, _, err := r.getCIsWithMissingAttributes(ctx, limit)
		results <- result{name: "missing_attributes", count: count, err: err}
	}()

	// Orphaned CIs
	go func() {
		count, _, err := r.getOrphanedCIs(ctx, limit)
		results <- result{name: "orphaned_cis", count: count, err: err}
	}()

	// No lifecycle status
	go func() {
		count, _, err := r.getCIsWithoutLifecycleStatus(ctx, limit)
		results <- result{name: "no_lifecycle_status", count: count, err: err}
	}()

	// No tags
	go func() {
		count, _, err := r.getCIsWithoutTags(ctx, limit)
		results <- result{name: "no_tags", count: count, err: err}
	}()

	// Stale data - get all three values
	go func() {
		stale30, stale60, stale90, err := r.getStaleCIs(ctx)
		results <- result{name: "stale", stale30: stale30, stale60: stale60, stale90: stale90, err: err}
	}()

	// Duplicates
	go func() {
		count, err := r.getDuplicateCIs(ctx)
		results <- result{name: "duplicates", count: count, err: err}
	}()

	// Collect results
	metrics := &DataQualityMetrics{}
	completed := 0

	for completed < 6 {
		select {
		case res := <-results:
			switch res.name {
			case "missing_attributes":
				if res.err != nil {
					return nil, fmt.Errorf("failed to get CIs with missing attributes: %w", res.err)
				}
				metrics.MissingAttributes.Count = res.count
				metrics.MissingAttributes.Percentage = calculatePercentage(int64(res.count), totalCIs)

			case "orphaned_cis":
				if res.err != nil {
					return nil, fmt.Errorf("failed to get orphaned CIs: %w", res.err)
				}
				metrics.OrphanedCIs.Count = res.count
				metrics.OrphanedCIs.Percentage = calculatePercentage(int64(res.count), totalCIs)

			case "no_lifecycle_status":
				if res.err != nil {
					return nil, fmt.Errorf("failed to get CIs without lifecycle status: %w", res.err)
				}
				metrics.NoLifecycleStatus.Count = res.count
				metrics.NoLifecycleStatus.Percentage = calculatePercentage(int64(res.count), totalCIs)

			case "no_tags":
				if res.err != nil {
					return nil, fmt.Errorf("failed to get CIs without tags: %w", res.err)
				}
				metrics.NoTags.Count = res.count
				metrics.NoTags.Percentage = calculatePercentage(int64(res.count), totalCIs)

			case "stale":
				if res.err != nil {
					return nil, fmt.Errorf("failed to get stale CIs: %w", res.err)
				}
				// Store all three stale values
				metrics.Stale30Days = res.stale30
				metrics.Stale60Days = res.stale60
				metrics.Stale90Days = res.stale90

			case "duplicates":
				if res.err != nil {
					return nil, fmt.Errorf("failed to get duplicate CIs: %w", res.err)
				}
				metrics.Duplicates = res.count
			}
			completed++
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// If details are requested, fetch the CI lists
	if includeDetails {
		_, missingCIs, err := r.getCIsWithMissingAttributes(ctx, limit)
		if err == nil {
			metrics.MissingAttributes.CIs = missingCIs
		}

		_, orphanedCIs, err := r.getOrphanedCIs(ctx, limit)
		if err == nil {
			metrics.OrphanedCIs.CIs = orphanedCIs
		}

		_, noStatusCIs, err := r.getCIsWithoutLifecycleStatus(ctx, limit)
		if err == nil {
			metrics.NoLifecycleStatus.CIs = noStatusCIs
		}

		_, noTagsCIs, err := r.getCIsWithoutTags(ctx, limit)
		if err == nil {
			metrics.NoTags.CIs = noTagsCIs
		}
	}

	return metrics, nil
}

// getCIsWithMissingAttributes returns CIs that are missing required attributes
func (r *Repository) getCIsWithMissingAttributes(ctx context.Context, limit int) (int, []QualityIssueCI, error) {
	// For now, we'll check if attributes field is empty or null
	// A more sophisticated check would verify against required_attributes in ci_type_definitions
	query := `
		SELECT id, name, ci_type, created_at, updated_at
		FROM configuration_items
		WHERE attributes = '{}'::jsonb OR attributes IS NULL
		ORDER BY updated_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, nil, fmt.Errorf("failed to get CIs with missing attributes: %w", err)
	}
	defer rows.Close()

	var cis []QualityIssueCI
	for rows.Next() {
		var ci QualityIssueCI
		err := rows.Scan(&ci.ID, &ci.Name, &ci.CIType, &ci.CreatedAt, &ci.UpdatedAt)
		if err != nil {
			r.logger.ErrorDatabase("SCAN", "configuration_items", err, nil)
			return 0, nil, fmt.Errorf("failed to scan CI: %w", err)
		}
		cis = append(cis, ci)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*) FROM configuration_items
		WHERE attributes = '{}'::jsonb OR attributes IS NULL
	`
	var count int
	err = r.db.QueryRow(ctx, countQuery).Scan(&count)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, nil, err
	}

	return count, cis, nil
}

// getOrphanedCIs returns CIs that have no relationships
func (r *Repository) getOrphanedCIs(ctx context.Context, limit int) (int, []QualityIssueCI, error) {
	query := `
		SELECT DISTINCT c.id, c.name, c.ci_type, c.created_at, c.updated_at
		FROM configuration_items c
		LEFT JOIN relationships r ON c.id = r.source_id OR c.id = r.target_id
		WHERE r.id IS NULL
		ORDER BY c.updated_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, nil, fmt.Errorf("failed to get orphaned CIs: %w", err)
	}
	defer rows.Close()

	var cis []QualityIssueCI
	for rows.Next() {
		var ci QualityIssueCI
		err := rows.Scan(&ci.ID, &ci.Name, &ci.CIType, &ci.CreatedAt, &ci.UpdatedAt)
		if err != nil {
			r.logger.ErrorDatabase("SCAN", "configuration_items", err, nil)
			return 0, nil, fmt.Errorf("failed to scan CI: %w", err)
		}
		cis = append(cis, ci)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(DISTINCT c.id)
		FROM configuration_items c
		LEFT JOIN relationships r ON c.id = r.source_id OR c.id = r.target_id
		WHERE r.id IS NULL
	`
	var count int
	err = r.db.QueryRow(ctx, countQuery).Scan(&count)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, nil, err
	}

	return count, cis, nil
}

// getCIsWithoutLifecycleStatus returns CIs without lifecycle status
func (r *Repository) getCIsWithoutLifecycleStatus(ctx context.Context, limit int) (int, []QualityIssueCI, error) {
	query := `
		SELECT id, name, ci_type, created_at, updated_at
		FROM configuration_items
		WHERE lifecycle_status_id IS NULL
		ORDER BY updated_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, nil, fmt.Errorf("failed to get CIs without lifecycle status: %w", err)
	}
	defer rows.Close()

	var cis []QualityIssueCI
	for rows.Next() {
		var ci QualityIssueCI
		err := rows.Scan(&ci.ID, &ci.Name, &ci.CIType, &ci.CreatedAt, &ci.UpdatedAt)
		if err != nil {
			r.logger.ErrorDatabase("SCAN", "configuration_items", err, nil)
			return 0, nil, fmt.Errorf("failed to scan CI: %w", err)
		}
		cis = append(cis, ci)
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM configuration_items WHERE lifecycle_status_id IS NULL`
	var count int
	err = r.db.QueryRow(ctx, countQuery).Scan(&count)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, nil, err
	}

	return count, cis, nil
}

// getCIsWithoutTags returns CIs that have no tags
func (r *Repository) getCIsWithoutTags(ctx context.Context, limit int) (int, []QualityIssueCI, error) {
	query := `
		SELECT id, name, ci_type, created_at, updated_at
		FROM configuration_items
		WHERE tags = '{}'::text[] OR array_length(tags, 1) IS NULL
		ORDER BY updated_at DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, nil, fmt.Errorf("failed to get CIs without tags: %w", err)
	}
	defer rows.Close()

	var cis []QualityIssueCI
	for rows.Next() {
		var ci QualityIssueCI
		err := rows.Scan(&ci.ID, &ci.Name, &ci.CIType, &ci.CreatedAt, &ci.UpdatedAt)
		if err != nil {
			r.logger.ErrorDatabase("SCAN", "configuration_items", err, nil)
			return 0, nil, fmt.Errorf("failed to scan CI: %w", err)
		}
		cis = append(cis, ci)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*) FROM configuration_items
		WHERE tags = '{}'::text[] OR array_length(tags, 1) IS NULL
	`
	var count int
	err = r.db.QueryRow(ctx, countQuery).Scan(&count)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, nil, err
	}

	return count, cis, nil
}

// getStaleCIs returns counts of CIs stale by 30, 60, and 90 days
func (r *Repository) getStaleCIs(ctx context.Context) (stale30, stale60, stale90 int, err error) {
	// Stale 30+ days
	query30 := `
		SELECT COUNT(*) FROM configuration_items
		WHERE updated_at < NOW() - INTERVAL '30 days'
	`
	err = r.db.QueryRow(ctx, query30).Scan(&stale30)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, 0, 0, fmt.Errorf("failed to get stale 30 CIs: %w", err)
	}

	// Stale 60+ days
	query60 := `
		SELECT COUNT(*) FROM configuration_items
		WHERE updated_at < NOW() - INTERVAL '60 days'
	`
	err = r.db.QueryRow(ctx, query60).Scan(&stale60)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return stale30, 0, 0, fmt.Errorf("failed to get stale 60 CIs: %w", err)
	}

	// Stale 90+ days
	query90 := `
		SELECT COUNT(*) FROM configuration_items
		WHERE updated_at < NOW() - INTERVAL '90 days'
	`
	err = r.db.QueryRow(ctx, query90).Scan(&stale90)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return stale30, stale60, 0, fmt.Errorf("failed to get stale 90 CIs: %w", err)
	}

	return stale30, stale60, stale90, nil
}

// getDuplicateCIs returns count of duplicate CIs (by name and type)
func (r *Repository) getDuplicateCIs(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) FROM (
			SELECT name, ci_type, COUNT(*)
			FROM configuration_items
			GROUP BY name, ci_type
			HAVING COUNT(*) > 1
		) duplicates
	`
	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, fmt.Errorf("failed to get duplicate CIs: %w", err)
	}

	return count, nil
}

// Helper function to calculate percentage
func calculatePercentage(count, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(count) / float64(total) * 100
}

// Asset Aging Metrics functions

// GetAssetAgingMetrics retrieves comprehensive asset aging metrics
func (r *Repository) GetAssetAgingMetrics(ctx context.Context, eolThresholdMonths int, limit int) (*AssetAgingMetrics, error) {
	// Get all metrics concurrently
	type result struct {
		metrics     *AssetAgingMetrics
		distribution AgeDistribution
		eolAssets    []ApproachingEOLAsset
		avgAge       float64
		oldest       *OldestAsset
		err          error
	}
	results := make(chan result, 1)

	go func() {
		var metrics AssetAgingMetrics

		// Get age distribution
		distribution, err := r.getAgeDistribution(ctx)
		if err != nil {
			results <- result{err: fmt.Errorf("failed to get age distribution: %w", err)}
			return
		}
		metrics.Distribution = distribution

		// Get approaching EOL assets
		eolAssets, err := r.getApproachingEOLAssets(ctx, eolThresholdMonths, limit)
		if err != nil {
			results <- result{err: fmt.Errorf("failed to get approaching EOL assets: %w", err)}
			return
		}
		metrics.ApproachingEOL = eolAssets

		// Get average age
		avgAge, err := r.getAverageAge(ctx)
		if err != nil {
			results <- result{err: fmt.Errorf("failed to get average age: %w", err)}
			return
		}
		metrics.AverageAgeMonths = avgAge

		// Get oldest asset
		oldest, err := r.getOldestAsset(ctx)
		if err != nil {
			// Don't fail on oldest asset error, just log it
			r.logger.Error().Err(err).Str("component", "asset_aging").Msg("Failed to get oldest asset")
		} else {
			metrics.OldestAsset = oldest
		}

		results <- result{metrics: &metrics}
	}()

	select {
	case res := <-results:
		if res.err != nil {
			return nil, res.err
		}
		return res.metrics, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// getAgeDistribution calculates the distribution of assets by age
func (r *Repository) getAgeDistribution(ctx context.Context) (AgeDistribution, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '1 year') AS less_than_1_year,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '3 years' AND created_at < NOW() - INTERVAL '1 year') AS one_to_3_years,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '5 years' AND created_at < NOW() - INTERVAL '3 years') AS three_to_5_years,
			COUNT(*) FILTER (WHERE created_at < NOW() - INTERVAL '5 years') AS more_than_5_years
		FROM configuration_items
	`

	var distribution AgeDistribution
	err := r.db.QueryRow(ctx, query).Scan(
		&distribution.LessThan1Year,
		&distribution.OneTo3Years,
		&distribution.ThreeTo5Years,
		&distribution.MoreThan5Years,
	)

	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return AgeDistribution{}, fmt.Errorf("failed to get age distribution: %w", err)
	}

	return distribution, nil
}

// getApproachingEOLAssets finds assets approaching end-of-life within threshold
func (r *Repository) getApproachingEOLAssets(ctx context.Context, thresholdMonths int, limit int) ([]ApproachingEOLAsset, error) {
	// Default to 6 months if not specified
	if thresholdMonths <= 0 {
		thresholdMonths = 6
	}

	query := `
		SELECT
			ci.id,
			ci.name,
			ci.ci_type,
			ci.created_at + (ci.useful_life_months * INTERVAL '1 month') AS eol_date,
			EXTRACT(DAY FROM (ci.created_at + (ci.useful_life_months * INTERVAL '1 month') - NOW())) AS days_until_eol
		FROM configuration_items ci
		WHERE ci.useful_life_months IS NOT NULL
			AND ci.created_at + (ci.useful_life_months * INTERVAL '1 month') < NOW() + INTERVAL '1 month' * $1
			AND ci.created_at + (ci.useful_life_months * INTERVAL '1 month') > NOW()
		ORDER BY days_until_eol ASC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(ctx, query, thresholdMonths)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, fmt.Errorf("failed to get approaching EOL assets: %w", err)
	}
	defer rows.Close()

	var assets []ApproachingEOLAsset
	for rows.Next() {
		var asset ApproachingEOLAsset
		var eolDate time.Time

		err := rows.Scan(
			&asset.CI.ID,
			&asset.CI.Name,
			&asset.Type,
			&eolDate,
			&asset.DaysUntilEOL,
		)
		if err != nil {
			r.logger.ErrorDatabase("SCAN", "configuration_items", err, nil)
			return nil, fmt.Errorf("failed to scan EOL asset: %w", err)
		}

		asset.EOLDate = eolDate
		assets = append(assets, asset)
	}

	return assets, nil
}

// getAverageAge calculates the average age of all assets in months
func (r *Repository) getAverageAge(ctx context.Context) (float64, error) {
	query := `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (NOW() - created_at)) / (30 * 86400)), 0) AS avg_age_months
		FROM configuration_items
		WHERE created_at IS NOT NULL
	`

	var avgAge float64
	err := r.db.QueryRow(ctx, query).Scan(&avgAge)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, fmt.Errorf("failed to get average age: %w", err)
	}

	return avgAge, nil
}

// getOldestAsset finds the oldest asset in the CMDB
func (r *Repository) getOldestAsset(ctx context.Context) (*OldestAsset, error) {
	query := `
		SELECT
			id,
			name,
			created_at,
			FLOOR(EXTRACT(EPOCH FROM (NOW() - created_at)) / (30 * 86400)) AS age_months
		FROM configuration_items
		WHERE created_at IS NOT NULL
		ORDER BY created_at ASC
		LIMIT 1
	`

	var oldest OldestAsset
	err := r.db.QueryRow(ctx, query).Scan(
		&oldest.ID,
		&oldest.Name,
		&oldest.CreatedAt,
		&oldest.AgeMonths,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // No assets yet
		}
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, fmt.Errorf("failed to get oldest asset: %w", err)
	}

	return &oldest, nil
}

// GetRiskMetrics retrieves risk assessment metrics for the dashboard
func (r *Repository) GetRiskMetrics(ctx context.Context, limit int) (*RiskMetrics, error) {
	// Get high-risk CIs (this will also calculate the overall risk metrics)
	highRiskCIs, err := r.getHighRiskCIs(ctx, limit)
	if err != nil {
		return nil, err
	}

	// Calculate metrics from high-risk CIs
	metrics := &RiskMetrics{
		HighRiskCIs:         highRiskCIs,
		NoRedundancyCount:   0,
		CriticalAssetsCount: 0,
		SPOFCount:           0,
		RiskScore:           0,
		ComplianceViolations: 0,
	}

	// Count by category
	for _, ci := range highRiskCIs {
		if ci.IsCritical {
			metrics.CriticalAssetsCount++
		}
		if !ci.HasRedundancy && ci.IsAmortizable {
			metrics.NoRedundancyCount++
		}
		if !ci.HasRedundancy && ci.IsAmortizable && ci.AgeMonths > 36 {
			metrics.SPOFCount++
		}
	}

	// Calculate overall risk score based on percentages
	totalCIs := 0
	if len(highRiskCIs) > 0 {
		// Get total count of amortizable CIs for percentage calculation
		var totalAmortizable int
		countQuery := `SELECT COUNT(*) FROM configuration_items WHERE is_amortizable = true`
		_ = r.db.QueryRow(ctx, countQuery).Scan(&totalAmortizable)
		totalCIs = totalAmortizable

		if totalCIs > 0 {
			spofRisk := (float64(metrics.SPOFCount) / float64(totalCIs)) * 40
			criticalRisk := (float64(metrics.CriticalAssetsCount) / float64(totalCIs)) * 30
			noRedundancyRisk := (float64(metrics.NoRedundancyCount) / float64(totalCIs)) * 30
			metrics.RiskScore = int(spofRisk + criticalRisk + noRedundancyRisk)
			if metrics.RiskScore > 100 {
				metrics.RiskScore = 100
			}
		}
	}

	return metrics, nil
}

// getNoRedundancyCount returns count of assets without redundancy tags
func (r *Repository) getNoRedundancyCount(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM configuration_items ci
		WHERE ci.is_amortizable = true
		AND NOT EXISTS (
			SELECT 1
			FROM unnest(ci.tags) tag
			WHERE tag IN ('backup', 'redundancy', 'ha', 'failover', 'cluster')
		)
	`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return 0, fmt.Errorf("failed to get no redundancy count: %w", err)
	}

	return count, nil
}

// getHighRiskCIs retrieves high-risk CIs with detailed risk information
func (r *Repository) getHighRiskCIs(ctx context.Context, limit int) ([]HighRiskCI, error) {
	query := `
		SELECT
			ci.id,
			ci.name,
			ci.ci_type,
			ctd.is_amortizable,
			EXISTS (
				SELECT 1
				FROM unnest(ci.tags) tag
				WHERE tag IN ('backup', 'redundancy', 'ha', 'failover', 'cluster')
			) as has_redundancy,
			EXISTS (
				SELECT 1
				FROM unnest(ci.tags)
				LIMIT 1
			) as has_tags,
			FLOOR(EXTRACT(EPOCH FROM (NOW() - ci.created_at)) / (30 * 86400))::int as age_months,
			-- Calculate risk score (0-100)
			LEAST(100,
				(CASE WHEN NOT ctd.is_amortizable THEN 10 ELSE 0 END) +
				(CASE WHEN ctd.is_amortizable AND NOT EXISTS (
					SELECT 1 FROM unnest(ci.tags) tag
					WHERE tag IN ('backup', 'redundancy', 'ha', 'failover', 'cluster')
				) THEN 30 ELSE 0 END) +
				(CASE WHEN FLOOR(EXTRACT(EPOCH FROM (NOW() - ci.created_at)) / (30 * 86400)) > 60 THEN 20 ELSE 0 END) +
				(CASE WHEN NOT EXISTS (SELECT 1 FROM unnest(ci.tags) LIMIT 1) THEN 20 ELSE 0 END) +
				(CASE WHEN ci.lifecycle_status_id IS NULL THEN 20 ELSE 0 END)
			)::int as risk_score
		FROM configuration_items ci
		INNER JOIN ci_type_definitions ctd ON ci.ci_type = ctd.name
		WHERE ctd.is_amortizable = true
		ORDER BY risk_score DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		r.logger.ErrorDatabase("SELECT", "configuration_items", err, nil)
		return nil, fmt.Errorf("failed to get high-risk CIs: %w", err)
	}
	defer rows.Close()

	var highRiskCIs []HighRiskCI
	for rows.Next() {
		var ci HighRiskCI
		err := rows.Scan(
			&ci.ID,
			&ci.Name,
			&ci.CIType,
			&ci.IsAmortizable,
			&ci.HasRedundancy,
			&ci.HasTags,
			&ci.AgeMonths,
			&ci.RiskScore,
		)
		if err != nil {
			r.logger.ErrorDatabase("SCAN", "configuration_items", err, nil)
			return nil, fmt.Errorf("failed to scan high-risk CI: %w", err)
		}

		// Mark as critical if no redundancy and amortizable
		ci.IsCritical = ci.IsAmortizable && !ci.HasRedundancy

		highRiskCIs = append(highRiskCIs, ci)
	}

	return highRiskCIs, nil
}

