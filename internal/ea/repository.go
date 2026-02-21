package ea

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles EA data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new EA repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ============================================================================
// EA Teams CRUD
// ============================================================================

// CreateTeam creates a new EA team
func (r *Repository) CreateTeam(ctx context.Context, req *CreateEATeamRequest, createdBy uuid.UUID) (*EATeam, error) {
	team := &EATeam{
		ID:        uuid.New(),
		Name:      req.Name,
		Description: req.Description,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: createdBy,
	}

	query := `
		INSERT INTO ea_teams (id, name, description, created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, description, created_at, updated_at, created_by
	`

	err := r.db.QueryRow(ctx, query,
		team.ID, team.Name, team.Description, team.CreatedAt, team.UpdatedAt, team.CreatedBy,
	).Scan(
		&team.ID, &team.Name, &team.Description, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy,
	)

	if err != nil {
			return nil, fmt.Errorf("failed to create EA team: %w", err)
	}

	return team, nil
}

// GetTeamByName retrieves an EA team by name
func (r *Repository) GetTeamByName(ctx context.Context, name string) (*EATeam, error) {
	team := &EATeam{}

	query := `
		SELECT id, name, description, created_at, updated_at, created_by
		FROM ea_teams
		WHERE name = $1
	`

	err := r.db.QueryRow(ctx, query, name).Scan(
		&team.ID, &team.Name, &team.Description, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get EA team: %w", err)
	}

	return team, nil
}

// GetTeamByID retrieves an EA team by ID
func (r *Repository) GetTeamByID(ctx context.Context, id uuid.UUID) (*EATeam, error) {
	team := &EATeam{}

	query := `
		SELECT id, name, description, created_at, updated_at, created_by
		FROM ea_teams
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.Description, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get EA team: %w", err)
	}

	return team, nil
}

// ListTeams retrieves all EA teams
func (r *Repository) ListTeams(ctx context.Context) ([]*EATeam, error) {
	query := `
		SELECT id, name, description, created_at, updated_at, created_by
		FROM ea_teams
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list EA teams: %w", err)
	}
	defer rows.Close()

	var teams []*EATeam
	for rows.Next() {
		team := &EATeam{}
		err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan EA team: %w", err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// UpdateTeam updates an EA team
func (r *Repository) UpdateTeam(ctx context.Context, id uuid.UUID, req *UpdateEATeamRequest) (*EATeam, error) {
	team := &EATeam{}

	query := `
		UPDATE ea_teams
		SET name = COALESCE($2, name),
		    description = COALESCE($3, description),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at, created_by
	`

	err := r.db.QueryRow(ctx, query, id, req.Name, req.Description).Scan(
		&team.ID, &team.Name, &team.Description, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update EA team: %w", err)
	}

	return team, nil
}

// DeleteTeam deletes an EA team
func (r *Repository) DeleteTeam(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ea_teams WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete EA team: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrTeamNotFound
	}

	return nil
}

// ============================================================================
// CI Type Queries
// ============================================================================

// GetCITypeByName retrieves a CI type definition by name
func (r *Repository) GetCITypeByName(ctx context.Context, name string) (*CITypeDefinition, error) {
	ciType := &CITypeDefinition{}

	query := `
		SELECT id, name, description, required_attributes, optional_attributes,
		       created_at, updated_at, created_by
		FROM ci_type_definitions
		WHERE name = $1
	`

	var requiredAttrs, optionalAttrs []byte
	err := r.db.QueryRow(ctx, query, name).Scan(
		&ciType.ID, &ciType.Name, &ciType.Description, &requiredAttrs, &optionalAttrs,
		&ciType.CreatedAt, &ciType.UpdatedAt, &ciType.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get CI type: %w", err)
	}

	// Unmarshal JSONB attributes
	if err := json.Unmarshal(requiredAttrs, &ciType.RequiredAttributes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal required attributes: %w", err)
	}
	if err := json.Unmarshal(optionalAttrs, &ciType.OptionalAttributes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal optional attributes: %w", err)
	}

	return ciType, nil
}

// ListEACITypes retrieves all EA CI type definitions
func (r *Repository) ListEACITypes(ctx context.Context) ([]*CITypeDefinition, error) {
	query := `
		SELECT id, name, description, required_attributes, optional_attributes,
		       created_at, updated_at, created_by
		FROM ci_type_definitions
		WHERE name LIKE 'EA.%'
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list EA CI types: %w", err)
	}
	defer rows.Close()

	var types []*CITypeDefinition
	for rows.Next() {
		ciType := &CITypeDefinition{}
		var requiredAttrs, optionalAttrs []byte

		err := rows.Scan(
			&ciType.ID, &ciType.Name, &ciType.Description, &requiredAttrs, &optionalAttrs,
			&ciType.CreatedAt, &ciType.UpdatedAt, &ciType.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan CI type: %w", err)
		}

		// Unmarshal JSONB attributes
		if err := json.Unmarshal(requiredAttrs, &ciType.RequiredAttributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal required attributes: %w", err)
		}
		if err := json.Unmarshal(optionalAttrs, &ciType.OptionalAttributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal optional attributes: %w", err)
		}

		types = append(types, ciType)
	}

	return types, nil
}

// ============================================================================
// EA Entity CRUD (stored in configuration_items table)
// ============================================================================

// Create creates a new EA entity
func (r *Repository) Create(ctx context.Context, entity *EAEntity) (*EAEntity, error) {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}

	now := time.Now()
	query := `
		INSERT INTO configuration_items (
			id, name, ci_type, attributes, tags, lifecycle_status_id,
			created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, ci_type, attributes, tags, lifecycle_status_id,
		          created_at, updated_at, created_by, updated_by
	`

	var result EAEntity

	err := r.db.QueryRow(ctx, query,
		entity.ID, entity.Name, entity.CIType, entity.Attributes, entity.Tags,
		entity.LifecycleStatusID, entity.CreatedBy, now, now,
	).Scan(
		&result.ID, &result.Name, &result.CIType, &result.Attributes, &result.Tags,
		&result.LifecycleStatusID, &result.CreatedAt, &result.UpdatedAt, &result.CreatedBy, &result.UpdatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create EA entity: %w", err)
	}

	// Build lifecycle status object if present
	if result.LifecycleStatusID != nil {
		result.LifecycleStatus = &LifecycleStatus{
			ID:          *result.LifecycleStatusID,
			Name:        "", // Will be filled by service layer
			DisplayName: "",
		}
	}

	// Set data quality score
	result.DataQualityScore = entity.DataQualityScore

	return &result, nil
}

// GetByID retrieves an EA entity by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*EAEntity, error) {
	entityID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid entity ID: %w", err)
	}

	query := `
		SELECT ci.id, ci.name, ci.ci_type, ci.attributes, ci.tags, ci.lifecycle_status_id,
		       ls.name, ls.display_name, ls.description, ls.color, ls.icon,
		       ci.created_at, ci.updated_at, ci.created_by, ci.updated_by
		FROM configuration_items ci
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		WHERE ci.id = $1 AND ci.ci_type LIKE 'EA.%'
	`

	var entity EAEntity
	var lifecycleStatusName, lifecycleStatusDisplayName, lifecycleStatusDescription, lifecycleStatusColor, lifecycleStatusIcon *string

	err = r.db.QueryRow(ctx, query, entityID).Scan(
		&entity.ID, &entity.Name, &entity.CIType, &entity.Attributes, &entity.Tags,
		&entity.LifecycleStatusID, &lifecycleStatusName, &lifecycleStatusDisplayName,
		&lifecycleStatusDescription, &lifecycleStatusColor, &lifecycleStatusIcon,
		&entity.CreatedAt, &entity.UpdatedAt, &entity.CreatedBy, &entity.UpdatedBy,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("EA entity not found")
		}
		return nil, fmt.Errorf("failed to get EA entity: %w", err)
	}

	// Build lifecycle status object if present
	if entity.LifecycleStatusID != nil {
		entity.LifecycleStatus = &LifecycleStatus{
			ID:          *entity.LifecycleStatusID,
			Name:        getStringValue(lifecycleStatusName),
			DisplayName: getStringValue(lifecycleStatusDisplayName),
			Description: lifecycleStatusDescription,
			Color:       lifecycleStatusColor,
			Icon:        lifecycleStatusIcon,
		}
	}

	// Extract owner from attributes
	if owner, ok := entity.Attributes["ea_owner"].(string); ok {
		entity.Owner = owner
	}

	// Extract description from attributes
	if desc, ok := entity.Attributes["description"].(string); ok {
		entity.Description = &desc
	}

	// Extract data quality score from attributes
	if dqs, ok := entity.Attributes["data_quality_score"].(float64); ok {
		entity.DataQualityScore = dqs
	}

	return &entity, nil
}

// Update updates an EA entity
func (r *Repository) Update(ctx context.Context, entity *EAEntity) error {
	now := time.Now()

	query := `
		UPDATE configuration_items
		SET name = $2,
		    attributes = $3,
		    tags = $4,
		    lifecycle_status_id = $5,
		    updated_at = $6,
		    updated_by = $7
		WHERE id = $1 AND ci_type LIKE 'EA.%'
		RETURNING id, name, ci_type, attributes, tags, lifecycle_status_id,
		          created_at, updated_at, created_by, updated_by
	`

	var result EAEntity
	err := r.db.QueryRow(ctx, query,
		entity.ID, entity.Name, entity.Attributes, entity.Tags,
		entity.LifecycleStatusID, now, entity.UpdatedBy,
	).Scan(
		&result.ID, &result.Name, &result.CIType, &result.Attributes, &result.Tags,
		&result.LifecycleStatusID, &result.CreatedAt, &result.UpdatedAt,
		&result.CreatedBy, &result.UpdatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to update EA entity: %w", err)
	}

	return nil
}

// Delete soft deletes an EA entity
func (r *Repository) Delete(ctx context.Context, id string, forceDelete bool) error {
	entityID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid entity ID: %w", err)
	}

	// Check for relationships first
	relationshipCount, err := r.CheckRelationships(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check relationships: %w", err)
	}

	if relationshipCount > 0 && !forceDelete {
		return &ErrRelationshipsExist{Count: relationshipCount}
	}

	query := `DELETE FROM configuration_items WHERE id = $1 AND ci_type LIKE 'EA.%'`

	result, err := r.db.Exec(ctx, query, entityID)
	if err != nil {
		return fmt.Errorf("failed to delete EA entity: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("EA entity not found")
	}

	return nil
}

// List retrieves EA entities with filtering and pagination
func (r *Repository) List(ctx context.Context, filter EAFilter) ([]EAEntity, int, error) {
	offset := (filter.Page - 1) * filter.PageSize

	// Build WHERE clause
	whereClause := "WHERE ci.ci_type LIKE 'EA.%'"
	args := []interface{}{}
	argIndex := 1

	if filter.Domain != "" {
		whereClause += fmt.Sprintf(" AND ci.ci_type LIKE $%d", argIndex)
		args = append(args, "EA."+filter.Domain+"%")
		argIndex++
	}

	if filter.CIType != "" {
		whereClause += fmt.Sprintf(" AND ci.ci_type = $%d", argIndex)
		args = append(args, filter.CIType)
		argIndex++
	}

	if filter.LifecycleStatus != "" {
		whereClause += fmt.Sprintf(" AND ls.name = $%d", argIndex)
		args = append(args, filter.LifecycleStatus)
		argIndex++
	}

	if filter.Search != "" {
		whereClause += fmt.Sprintf(" AND (ci.name ILIKE $%d OR ci.attributes::text ILIKE $%d)", argIndex, argIndex+1)
		args = append(args, "%"+filter.Search+"%", "%"+filter.Search+"%")
		argIndex += 2
	}

	if len(filter.Tags) > 0 {
		whereClause += fmt.Sprintf(" AND ci.tags && $%d", argIndex)
		args = append(args, filter.Tags)
		argIndex++
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM configuration_items ci LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id %s", whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count EA entities: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT ci.id, ci.name, ci.ci_type, ci.attributes, ci.tags, ci.lifecycle_status_id,
		       ls.name, ls.display_name, ls.description, ls.color, ls.icon,
		       ci.created_at, ci.updated_at, ci.created_by, ci.updated_by
		FROM configuration_items ci
		LEFT JOIN lifecycle_statuses ls ON ci.lifecycle_status_id = ls.id
		%s
		ORDER BY ci.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list EA entities: %w", err)
	}
	defer rows.Close()

	var entities []EAEntity
	for rows.Next() {
		var entity EAEntity
		var lifecycleStatusName, lifecycleStatusDisplayName, lifecycleStatusDescription, lifecycleStatusColor, lifecycleStatusIcon *string

		err := rows.Scan(
			&entity.ID, &entity.Name, &entity.CIType, &entity.Attributes, &entity.Tags,
			&entity.LifecycleStatusID, &lifecycleStatusName, &lifecycleStatusDisplayName,
			&lifecycleStatusDescription, &lifecycleStatusColor, &lifecycleStatusIcon,
			&entity.CreatedAt, &entity.UpdatedAt, &entity.CreatedBy, &entity.UpdatedBy,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan EA entity: %w", err)
		}

		// Build lifecycle status object if present
		if entity.LifecycleStatusID != nil {
			entity.LifecycleStatus = &LifecycleStatus{
				ID:          *entity.LifecycleStatusID,
				Name:        getStringValue(lifecycleStatusName),
				DisplayName: getStringValue(lifecycleStatusDisplayName),
				Description: lifecycleStatusDescription,
				Color:       lifecycleStatusColor,
				Icon:        lifecycleStatusIcon,
			}
		}

		// Extract owner from attributes
		if owner, ok := entity.Attributes["ea_owner"].(string); ok {
			entity.Owner = owner
		}

		// Extract description from attributes
		if desc, ok := entity.Attributes["description"].(string); ok {
			entity.Description = &desc
		}

		// Extract data quality score from attributes
		if dqs, ok := entity.Attributes["data_quality_score"].(float64); ok {
			entity.DataQualityScore = dqs
		}

		entities = append(entities, entity)
	}

	return entities, total, nil
}

// CheckRelationships checks if an EA entity has any relationships
func (r *Repository) CheckRelationships(ctx context.Context, id string) (int, error) {
	entityID, err := uuid.Parse(id)
	if err != nil {
		return 0, fmt.Errorf("invalid entity ID: %w", err)
	}

	query := `
		SELECT COUNT(*)
		FROM relationships
		WHERE source_id = $1 OR target_id = $1
	`

	var count int
	err = r.db.QueryRow(ctx, query, entityID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to check relationships: %w", err)
	}

	return count, nil
}

// Helper function to handle nil strings
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// CITypeDefinition represents a CI type definition (reused from CI package)
type CITypeDefinition struct {
	ID                 uuid.UUID              `json:"id"`
	Name               string                 `json:"name"`
	Description        *string                `json:"description"`
	RequiredAttributes []AttributeDefinition  `json:"required_attributes"`
	OptionalAttributes []AttributeDefinition  `json:"optional_attributes"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	CreatedBy          uuid.UUID              `json:"created_by"`
}

// AttributeDefinition represents an attribute definition in JSONB
type AttributeDefinition struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Validation  map[string]interface{} `json:"validation,omitempty"`
}

// EAEntity represents an EA entity (stored as a CI)
type EAEntity struct {
	ID                uuid.UUID              `json:"id"`
	Name              string                 `json:"name"`
	CIType            string                 `json:"ci_type"`
	Description       *string                `json:"description,omitempty"`
	Owner             string                 `json:"owner"` // EA team name
	Attributes        map[string]interface{} `json:"attributes"`
	Tags              []string               `json:"tags"`
	LifecycleStatusID *uuid.UUID             `json:"lifecycle_status_id,omitempty"`
	LifecycleStatus   *LifecycleStatus       `json:"lifecycle_status,omitempty"`
	DataQualityScore  float64                `json:"data_quality_score"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	CreatedBy         uuid.UUID              `json:"created_by"`
	UpdatedBy         *uuid.UUID             `json:"updated_by,omitempty"`
}

// LifecycleStatus represents a lifecycle status
type LifecycleStatus struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description *string   `json:"description,omitempty"`
	Color       *string   `json:"color,omitempty"`
	Icon        *string   `json:"icon,omitempty"`
}

// EAFilter represents filtering options for listing EA entities
type EAFilter struct {
	Domain           string   `json:"domain,omitempty"`
	CIType           string   `json:"ci_type,omitempty"`
	LifecycleStatus  string   `json:"lifecycle_status,omitempty"`
	Search           string   `json:"search,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	Page             int      `json:"page"`
	PageSize         int      `json:"page_size"`
}

// ValidationResult represents the result of validating an EA entity
type ValidationResult struct {
	IsValid           bool             `json:"is_valid"`
	Errors            []ValidationError `json:"errors,omitempty"`
	DataQualityScore  float64          `json:"data_quality_score"`
	ValidAttributes   int              `json:"valid_attributes"`
	TotalRequired     int              `json:"total_required"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field       string `json:"field"`
	Message     string `json:"message"`
	Severity    string `json:"severity"` // error, warning
}
